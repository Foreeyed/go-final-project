package api

import "net/http"

func taskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
    case http.MethodPost:
		addTaskHandler(w, r)
	}
}
