package main

import (
	"go1f/pkg/server"

	"github.com/Foreeyed/go-final-project.git/pkg/db"
	"github.com/labstack/gommon/log"
)

func main() {

	err := db.InitDB("scheduler.db")
	if err != nil {
		log.Fatalf("database not started with err: %s", err.Error())
	}

	err = server.Run()
	if err != nil {
		log.Fatalf("server not started with err: %s", err.Error())
	}

}
