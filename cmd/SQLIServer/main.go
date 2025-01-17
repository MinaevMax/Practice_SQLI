package main

import (
	"log"
	"sql-injection-server/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatalln(err)
	}
}
