package httpServer

import (
  "database/sql"
  "fmt"
  "log"
  "net/http"
  //"strings"
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

func getbills(w http.ResponseWriter, r *http.Request) {
	log.Printf("Got a message!")
	var input Request
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("Error while getting data", err)
        return
    }
	name := input.Text
	log.Printf(name)

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
		err := rows.Scan(&bill_id, &bill_name, &bill_val)
		if err != nil{
			log.Printf("Error occured %v", err)
			continue
		}
		log.Printf("Bill number %v per %v in the amount of %v", bill_id, bill_name, bill_val)
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

func Start(wg *sync.WaitGroup) {
	defer wg.Done()
	port := os.Getenv("PORT")
	http.HandleFunc("/bills/check", getbills)
	http.HandleFunc("/", homeHandler)
	log.Println("Starting server on 8080...")
	log.Fatal(http.ListenAndServe(port, nil))
}