package server

import (
	"net/http"
	"os"

	"github.com/Foreeyed/go-final-project.git/pkg/api"
)

const (
	defaultPort string = "7540"
	webDir      string = "./web"
)

func Run() error {
	api.Init()
	
	port, ok := os.LookupEnv("TODO_PORT")
	if !ok {
		port = defaultPort
	}

	http.Handle("/", http.FileServer(http.Dir(webDir)))
	err := http.ListenAndServe(":"+port, nil)

	if err != nil {
		return err
	}

	return nil
}
