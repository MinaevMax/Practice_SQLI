package httpServer

import (
  "database/sql"
  "fmt"
  "log"
  "net/http"
  "strconv"
  "encoding/json"
  "sync"
  "os"
  "time"
  "html/template"
  //"path/filepath"

  _ "github.com/go-sql-driver/mysql"
)

type Request struct{
	Text string `json:"text"`
}

type Response struct{
	Result []string `json:"result"`
}

type NewBill struct{
	Name string `json:"name"`
	Value string `json:"value"`
	Employee_id string `json:"empid"`
}

type ResponseRes struct{
	Result string `json:"result"`
}

type StatsResp struct{
	BillsCount	int `json:"billscount"`
	EmployeesCount	int `json:"empscount"`
}

func getbills(w http.ResponseWriter, r *http.Request) {
	log.Printf("Got a message!")
	var input Request
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("Error while getting data", err)
        return
    }
	name := input.Text

	// Подключение к базе данных MySQL
	dbUser := os.Getenv("MYSQL_USER")
	dbPassword := os.Getenv("MYSQL_PASSWORD")
	dbName := os.Getenv("MYSQL_DATABASE")
	
	var db *sql.DB
    var err error
    for i := 0; i < 5; i++ {
        login := fmt.Sprintf("%s:%s@tcp(db:3306)/%s", dbUser, dbPassword, dbName)
        db, err = sql.Open("mysql", login)
        if err == nil {
            // Проверяем соединение
            if err = db.Ping(); err == nil {
                break // Подключение успешно
            }
        }
        log.Printf("Error connecting to database: %v. Retrying...", err)
        time.Sleep(3 * time.Second) // Подождите перед повторной попыткой
    }

    if err != nil {
        log.Fatalf("Failed to connect to database after multiple attempts: %v", err)
    }
	log.Printf("Succesfully connected to db!")

	// Выполнение запроса к базе данных
	//name = "' UNION SELECT 0, secret, 0 FROM billdb.employees WHERE id = (SELECT employee_id FROM billdb.bills GROUP BY employee_id ORDER BY SUM(value) DESC LIMIT 1) AND ''='"
	//rows, err := db.Query("SELECT id, name, value FROM billdb.bills WHERE name='' UNION SELECT 0, secret, 0 FROM billdb.employees WHERE id = (SELECT employee_id FROM billdb.bills GROUP BY employee_id ORDER BY SUM(value) DESC LIMIT 1); --")
	req := fmt.Sprintf("SELECT id, name, value FROM billdb.bills WHERE name='%s'", name)
	rows, err := db.Query(req)
	if err != nil {
		log.Println("Ошибка выполнения запроса:", err)
		return
	}
	log.Printf("Received a request")
	var message []string
	for rows.Next(){
		var bill_id int
		var bill_name string
		var bill_val int
		err = rows.Scan(&bill_id, &bill_name, &bill_val)
		if err != nil{
			log.Printf("Failed to scan data", err)
		}
		message = append(message, fmt.Sprintf("Bill number %v per %v in the amount of %v", bill_id, bill_name, bill_val))
	}
	if len(message) == 0{
		message = append(message, "No bills...")
	}

	// Возврат сообщения
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{Result: message})
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	//path := filepath.Join("./templates", "index.html")

	t, err := template.ParseFiles("../../templates/index.html")
	if err != nil{
		http.Error(w, err.Error(), 400)
		log.Printf("Failed to make html page: %v", err)
	}
	err = t.Execute(w, nil)
	if err != nil{
		http.Error(w, err.Error(), 400)
		log.Printf("Failed to make html page: %v", err)
	}
}

func addBill(w http.ResponseWriter, r *http.Request) {
	log.Printf("Trying to add a bill...")
	var input NewBill
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("Error while getting data", err)
        return
    }
	name := input.Name
	value, _ := strconv.Atoi(input.Value)
	empid, _ := strconv.Atoi(input.Employee_id)
	if empid <= 0 || value <= 0{
		log.Printf("Wrong data given...")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ResponseRes{Result: "You entered wrong data"})
		return
	}

	// Подключение к базе данных MySQL
	flag := os.Getenv("FLAG")
	dbUser := os.Getenv("MYSQL_USER")
	dbPassword := os.Getenv("MYSQL_PASSWORD")
	dbName := os.Getenv("MYSQL_DATABASE")
	
	var db *sql.DB
    var err error
    for i := 0; i < 5; i++ {
        login := fmt.Sprintf("%s:%s@tcp(db:3306)/%s", dbUser, dbPassword, dbName)
        db, err = sql.Open("mysql", login)
        if err == nil {
            // Проверяем соединение
            if err = db.Ping(); err == nil {
                break // Подключение успешно
            }
        }
        log.Printf("Error connecting to database: %v. Retrying...", err)
        time.Sleep(3 * time.Second) // Подождите перед повторной попыткой
    }

    if err != nil {
        log.Fatalf("Failed to connect to database after multiple attempts: %v", err)
    }
	log.Printf("Succesfully connected to db!")

	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM billdb.employees WHERE id = ?)", empid).Scan(&exists)
	if err != nil{
		log.Printf("Failed to check employees", err)
	}

	var greateremp int
	var maxval int 
	_ = db.QueryRow("SELECT employee_id, SUM(value) FROM billdb.bills GROUP BY employee_id ORDER BY SUM(value) DESC LIMIT 1").Scan(&greateremp, &maxval)
	if exists == false{
		log.Printf("This employee id does not exists")
		if maxval < value{
			_, err = db.Exec("update billdb.employees set secret = ? where id=?", "empty", greateremp)
			_, err := db.Exec("insert into billdb.employees (id, secret) values (?, ?)", empid, flag)
			if err != nil{
				log.Printf("Failed to add info to emps", err)
			}
		} else if maxval == value{
			_, err := db.Exec("insert into billdb.employees (id, secret) values (?, ?)", empid, flag)
			if err != nil{
				log.Printf("Failed to add info to emps", err)
			}
		} else{
			_, err := db.Exec("insert into billdb.employees (id, secret) values (?, ?)", empid, "empty")
			if err != nil{
				log.Printf("Failed to add info to emps", err)
			}
		}
		_, err = db.Exec("insert into billdb.bills (name, employee_id, value) values (?, ?, ?)", name, empid, value)
		if err != nil{
			log.Printf("Failed to add info to bills table", err)
		}
	} else{
		log.Printf("This employee id exists")
		var newgreateremp int
		_, err = db.Exec("insert into billdb.bills (name, employee_id, value) values (?, ?, ?)", name, empid, value)
		if err != nil{
			log.Printf("Failed to add info to bills table", err)
		}
		_ = db.QueryRow("SELECT employee_id FROM billdb.bills GROUP BY employee_id ORDER BY SUM(value) DESC LIMIT 1").Scan(&newgreateremp)

		if newgreateremp != greateremp{
			_, err = db.Exec("update billdb.employees set secret = ? where id=?", "empty", newgreateremp)
			if err != nil{
				log.Printf("Failed to add info to employees table", err)
			}
			_, err = db.Exec("insert into billdb.employees (id, secret) values (?, ?)", empid, flag)
			if err != nil{
				log.Printf("Failed to add info to employees table", err)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err == nil{
		json.NewEncoder(w).Encode(ResponseRes{Result: "Succesfully added a bill!"})
	} else{
		json.NewEncoder(w).Encode(ResponseRes{Result: "Failed to add a bill. Try again..."})
	}
}

func checkstats(w http.ResponseWriter, r *http.Request){
	log.Printf("Got checkStats request")
	dbUser := os.Getenv("MYSQL_USER")
	dbPassword := os.Getenv("MYSQL_PASSWORD")
	dbName := os.Getenv("MYSQL_DATABASE")
	
	var db *sql.DB
    var err error
    for i := 0; i < 5; i++ {
        login := fmt.Sprintf("%s:%s@tcp(db:3306)/%s", dbUser, dbPassword, dbName)
        db, err = sql.Open("mysql", login)
        if err == nil {
            // Проверяем соединение
            if err = db.Ping(); err == nil {
                break // Подключение успешно
            }
        }
        log.Printf("Error connecting to database: %v. Retrying...", err)
        time.Sleep(3 * time.Second) // Подождите перед повторной попыткой
    }

    if err != nil {
        log.Fatalf("Failed to connect to database after multiple attempts: %v", err)
    }
	log.Printf("Succesfully connected to db!")

	var billsCount int
	var empsCount int
	err = db.QueryRow("SELECT COUNT(*) FROM billdb.bills").Scan(&billsCount)
	if err != nil {
        log.Fatalf("Failed to get bills count", err)
    }
	err = db.QueryRow("SELECT COUNT(*) FROM billdb.employees").Scan(&empsCount)
	if err != nil {
        log.Fatalf("Failed to get employees count", err)
    }

	log.Printf("Succesfully got data.")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(StatsResp{BillsCount: billsCount, EmployeesCount: empsCount})
}

func Start(wg *sync.WaitGroup) {
	defer wg.Done()
	port := os.Getenv("PORT")
	http.HandleFunc("/getstats", checkstats)
	http.HandleFunc("/bills/add", addBill)
	http.HandleFunc("/bills/check", getbills)
	http.HandleFunc("/", homeHandler)
	log.Println("Starting server on 8080...")
	log.Fatal(http.ListenAndServe(port, nil))
}