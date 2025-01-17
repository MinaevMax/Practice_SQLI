package httpServer

import (
  "database/sql"
  "fmt"
  "log"
  "net/http"
  "strings"
  "sync"
  "os"
  "time"

  _ "github.com/go-sql-driver/mysql"
)

func getbills(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")

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
		message = append(message, fmt.Sprintf("Bill number %v per %v in the amount of %v", bill_id, bill_name, bill_val))
	}

	// Возврат сообщения
	fmt.Fprintf(w, strings.Join(message, "\n"))
}

func Start(wg *sync.WaitGroup) {
	defer wg.Done()
	port := os.Getenv("PORT")
	http.HandleFunc("/bills/check", getbills)
	log.Println("Starting server on 8080...")
	log.Fatal(http.ListenAndServe(port, nil))
}