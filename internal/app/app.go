package app

import (
	"sync"
	"sql-injection-server/internal/httpServer"
	"sql-injection-server/internal/db"
)

func Run() error {
	//db.Add()
	var wg sync.WaitGroup

	wg.Add(1)
	db.Start(&wg)
	wg.Add(1)
	httpServer.Start(&wg)
	return nil
}