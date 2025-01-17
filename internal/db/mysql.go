package db

import (
    "database/sql"
    "sync"
    _ "github.com/go-sql-driver/mysql"
	"math/rand"
	"strconv"
	"time"
	"log"
	"fmt"
	"os"
)
 
func Start(wg *sync.WaitGroup) { 
	defer wg.Done()
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
	
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS billdb.bills(id int AUTO_INCREMENT primary key, name text not null, employee_id int not null, value bigint not null)")
	if err != nil {
		log.Println("Ошибка создания таблицы:", err)
		return
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS billdb.employees(id int AUTO_INCREMENT primary key, secret text not null)")
	if err != nil {
		log.Println("Ошибка создания таблицы:", err)
		return
	}


	_, err = db.Exec("truncate table billdb.employees")
	if err != nil{
		log.Fatalf("Failed to reach employees db:", err)
	}
	_, err = db.Exec("truncate table billdb.bills")
	if err != nil{
		log.Fatalf("Failed to reach bills db:", err)
	}
	
	for i := 0; i < 8; i++{
		_, err := db.Exec("insert into billdb.employees (secret) values (?)", "empty")
		if err != nil{
			log.Printf("Failed to add info to emps", err)
		}
		//fmt.Println(res.LastInsertId())
	}
	
	employees := make([]int, 8, 8)
	for i := 0; i < 15; i++{
		rand.Seed(time.Now().UnixNano())
		emp := rand.Intn(8)
		val := rand.Intn(1500)
		_, err := db.Exec("insert into billdb.bills (name, employee_id, value) values (?, ?, ?)", 
        "Name" + strconv.Itoa(rand.Intn(16)), emp, val)
		if err != nil{
			log.Printf("Failed to add info to bills", err)
		}
		employees[emp] += val
		if err != nil{
			panic(err)
		}
		//fmt.Println(res.LastInsertId())
	}

	maxInd := 0
	for i := 0; i < len(employees); i++{
		if employees[i] > employees[maxInd]{
			maxInd = i
		}
	}

	_, err = db.Exec("update billdb.employees set secret = ? where id=?", flag, maxInd)
	
}

