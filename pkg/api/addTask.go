package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"go1f/pkg/db"
)

func checkDate(task *db.Task) error {
	now := time.Now().In(time.UTC)

	if task.Date == "" {
		task.Date = now.Format(TimeFormat)
		log.Printf("Дата не указана, установлена текущая дата: %v", task.Date)
	}

	t, err := time.Parse(TimeFormat, task.Date)
	if err != nil {
		return fmt.Errorf("некорректный формат даты: %v", err)
	}

	log.Printf("Парсинг даты задачи: %v", t)

	if !afterNow(now, t) {
		log.Printf("Дата задачи (%v) не может быть меньше сегодняшней (%v)", t, now)

		if task.Repeat == "" {
			task.Date = now.Format(TimeFormat)
			log.Printf("Дата установлена на текущую: %v", task.Date)
		} else {

			next, err := NextDate(now, task.Date, task.Repeat)
			if err != nil {
				return fmt.Errorf("ошибка в правиле повторения: %v", err)
			}

			nextParsed, err := time.Parse(TimeFormat, next)
			if err != nil {
				return fmt.Errorf("ошибка разбора следующей даты: %v", err)
			}

			log.Printf("Вычислена следующая дата: %v", nextParsed)

			if task.Repeat == "d 1" {
				task.Date = now.Format(TimeFormat)
				log.Printf("Дата установлена на сегодняшнюю, так как повторение 'd 1': %v", task.Date)
			} else if task.Repeat == "y" {
				if !afterNow(now, nextParsed) {
					task.Date = nextParsed.AddDate(1, 0, 0).Format(TimeFormat)
					log.Printf("Дата установлена на следующий год, так как повторение 'y': %v", task.Date)
				}
			} else {
				task.Date = next
			}
		}
	}

	return nil
}

func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task db.Task

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJson(w, map[string]string{"error": fmt.Sprintf("Ошибка чтения тела запроса: %v", err)}, http.StatusBadRequest)
		return
	}
	log.Printf("Полученные данные: %s", body)

	decoder := json.NewDecoder(bytes.NewReader(body))
	if err := decoder.Decode(&task); err != nil {
		writeJson(w, map[string]string{"error": fmt.Sprintf("Ошибка декодирования JSON: %v", err)}, http.StatusBadRequest)
		return
	}

	if task.Title == "" {
		writeJson(w, map[string]string{"error": "Поле 'title' обязательно"}, http.StatusBadRequest)
		return
	}

	if err := checkDate(&task); err != nil {
		writeJson(w, map[string]string{"error": err.Error()}, http.StatusBadRequest)
		return
	}

	id, err := db.AddTask(&task)
	if err != nil {
		writeJson(w, map[string]string{"error": fmt.Sprintf("Ошибка добавления задачи в базу данных: %v", err)}, http.StatusInternalServerError)
		return
	}

	writeJson(w, map[string]interface{}{"id": id}, http.StatusOK)
}


func writeJson(w http.ResponseWriter, data any, statusCode int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(statusCode)

	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		log.Printf("ошибка сериализации JSON: %v", err)
		http.Error(w, "ошибка сериализации ответа", http.StatusInternalServerError)
	}
}
