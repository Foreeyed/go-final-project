package api

import (
	"fmt"
	"net/http"

	"go1f/pkg/db"
)

type TasksResp struct {
	Tasks []*db.Task `json:"tasks"`
}

func tasksHandler(w http.ResponseWriter, r *http.Request) {
	tasks, err := db.Tasks(50) // получаем до 50 задач
	if err != nil {
		writeJson(w, map[string]string{
			"error": "ошибка получения задач: " + err.Error(),
		}, http.StatusInternalServerError)
		return
	}

	if tasks == nil {
		tasks = []*db.Task{}
	}

	result := make([]map[string]string, 0, len(tasks))
	for _, t := range tasks {
		result = append(result, map[string]string{
			"id":      fmt.Sprint(t.ID),
			"date":    t.Date,
			"title":   t.Title,
			"comment": t.Comment,
			"repeat":  t.Repeat,
		})
	}

	writeJson(w, map[string]any{
		"tasks": result,
	}, http.StatusOK)
}
