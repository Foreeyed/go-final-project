package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"go1f/pkg/db"
)

type TasksResp struct {
	Tasks []*db.Task `json:"tasks"`
}

func isValidDateFormat(date string) bool {
	_, err := time.Parse("02.01.2006", date)
	return err == nil
}

func formatDateForSearch(date string) string {
	parsedDate, err := time.Parse("02.01.2006", date)
	if err != nil {
		return ""
	}
	return parsedDate.Format("20060102")
}

func tasksHandler(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")

	var dateSearch string
	if isValidDateFormat(search) {
		dateSearch = formatDateForSearch(search)
		search = ""
	}

	tasks, err := db.Tasks(50, search, dateSearch)
	if err != nil {
		writeJson(w, map[string]string{"error": "ошибка получения задач: " + err.Error()}, http.StatusInternalServerError)
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

func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJson(w, map[string]string{"error": "Не указан идентификатор"}, http.StatusBadRequest)
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		writeJson(w, map[string]string{"error": "Задача не найдена"}, http.StatusNotFound)
	}

	writeJson(w, task, http.StatusOK)

}

func updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID      interface{} `json:"id"` // Меняем на interface{}
		Date    string      `json:"date"`
		Title   string      `json:"title"`
		Comment string      `json:"comment"`
		Repeat  string      `json:"repeat"`
	}

	// Чтение тела запроса в буфер
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		writeJson(w, map[string]string{"error": "Error reading request body"}, http.StatusBadRequest)
		return
	}

	// Логирование тела запроса
	log.Printf("Received raw request body: %s", buf.String())

	// Декодирование тела запроса с использованием json.Unmarshal
	if err := json.Unmarshal(buf.Bytes(), &req); err != nil {
		log.Printf("Error decoding request body: %v", err) // Логируем ошибку декодирования
		errEncode := json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		if errEncode != nil {
			log.Printf("Error encoding error response: %v", errEncode)
		}
		return
	}

	// Логируем полученные данные для отладки
	log.Printf("Полученные данные: %+v", req)

	// Преобразование ID в строку, если оно числовое
	idStr := ""
	switch v := req.ID.(type) {
	case string:
		idStr = v
	case float64: // Когда ID передан как число
		idStr = strconv.FormatFloat(v, 'f', 0, 64)
	default:
		log.Printf("Invalid ID type: %T", v)
		writeJson(w, map[string]string{"error": "Invalid ID type"}, http.StatusBadRequest)
		return
	}

	// Логирование ID для отладки
	log.Printf("Парсинг ID: %s", idStr)

	// Парсинг ID как целого числа
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		log.Printf("Invalid ID: %v", idStr) // Логируем ошибку парсинга ID
		writeJson(w, map[string]string{"error": "Invalid ID"}, http.StatusBadRequest)
		return
	}

	// Проверка на обязательные поля
	if req.Date == "" || req.Title == "" {
		log.Printf("Missing fields: Date=%s, Title=%s", req.Date, req.Title) // Логируем недостающие поля
		writeJson(w, map[string]string{"error": "Missing required fields"}, http.StatusBadRequest)
		return
	}

	// Парсинг даты в формате "20060102"
	dateParsed, err := time.Parse("20060102", req.Date)
	if err != nil {
		log.Printf("Invalid date format: %s", req.Date) // Логируем ошибку парсинга даты
		writeJson(w, map[string]string{"error": "Invalid date format"}, http.StatusBadRequest)
		return
	}

	// Получаем текущую дату без времени
	today := time.Now().Truncate(24 * time.Hour)

	// Сравнение только по дате (без учёта времени)
	if dateParsed.Before(today) {
		log.Printf("Date cannot be in the past: %s", req.Date) // Логируем ошибку, если дата в прошлом
		writeJson(w, map[string]string{"error": "Date cannot be in the past"}, http.StatusBadRequest)
		return
	}

	// Проверка на корректность правила повторений
	if req.Repeat != "" {
		_, err := NextDate(time.Now(), req.Date, req.Repeat)
		if err != nil {
			log.Printf("Invalid repeat rule: %s", req.Repeat) // Логируем ошибку повторения
			writeJson(w, map[string]string{"error": "Invalid repeat rule"}, http.StatusBadRequest)
			return
		}
	}

	query := `UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`
	_, err = db.GetDB().Exec(query, req.Date, req.Title, req.Comment, req.Repeat, id)
	if err != nil {
		log.Printf("Failed to update task: %v", err)
		writeJson(w, map[string]string{"error": fmt.Sprintf("Failed to update task: %v", err)}, http.StatusInternalServerError)
		return
	}

	writeJson(w, map[string]string{"status": "success", "id": idStr}, http.StatusOK)
}

func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	if id == "" {
		writeJson(w, map[string]string{"error": "Missing task ID"}, http.StatusBadRequest)
		return
	}

	err := db.DeleteTask(id)
	if err != nil {
		writeJson(w, map[string]string{"error": "Task not found"}, http.StatusNotFound)
		return
	}

	writeJson(w, map[string]string{}, http.StatusOK)
}



func doneTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if id == "" {
		writeJson(w, map[string]string{"error": "missing id"}, http.StatusBadRequest)
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		writeJson(w, map[string]string{"error": "task not found"}, http.StatusNotFound)
		return
	}

	if task.Repeat == "" {
		if err := db.DeleteTask(id); err != nil {
			writeJson(w, map[string]string{"error": "delete failed"}, http.StatusInternalServerError)
			return
		}
		writeJson(w, map[string]any{}, http.StatusOK)
		return
	}

	dateParsed, err := time.Parse("20060102", task.Date)
	if err != nil {
		writeJson(w, map[string]string{"error": "invalid task date"}, http.StatusInternalServerError)
		return
	}

	next, err := NextDate(dateParsed, task.Date, task.Repeat)
	if err != nil {
		writeJson(w, map[string]string{"error": err.Error()}, http.StatusBadRequest)
		return
	}

	if err := db.UpdateDate(next, id); err != nil {
		writeJson(w, map[string]string{"error": "update failed"}, http.StatusInternalServerError)
		return
	}
	writeJson(w, map[string]string{}, http.StatusOK)
}