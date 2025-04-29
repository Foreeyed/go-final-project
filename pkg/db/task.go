package db

import (
	"database/sql"
	"fmt"
)

func Init(database *sql.DB) {
	db = database
}

type Task struct {
	ID      int64  `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

func AddTask(task *Task) (int64, error) {
	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`
	res, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	return id, nil
}

func UpdateTask(task *Task) error {
	query := `UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`
	res, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if count == 0 {
		return fmt.Errorf("Задача не найдена")
	}
	return nil
}

func GetTask(id string) (*Task, error) {
	query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`

	var task Task

	err := db.QueryRow(query, id).Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("task not found")
		}
		return nil, err
	}

	return &task, nil
}


func Tasks(limit int, search string, dateSearch string) ([]*Task, error) {
	query := `SELECT id, date, title, comment, repeat FROM scheduler`

	if search != "" {
		query += ` WHERE title LIKE ? OR comment LIKE ?`
	}

	if dateSearch != "" {
		if search != "" {
			query += ` AND date = ?`
		} else {
			query += ` WHERE date = ?`
		}
	}

	query += ` ORDER BY date LIMIT ?`

	var params []interface{}
	if search != "" {
		params = append(params, "%"+search+"%", "%"+search+"%")
	}
	if dateSearch != "" {
		params = append(params, dateSearch)
	}
	params = append(params, limit)

	rows, err := db.Query(query, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, &task)
	}

	if tasks == nil {
		tasks = []*Task{}
	}

	return tasks, nil
}

func DeleteTask(id string) error {
	result, err := db.Exec("DELETE FROM scheduler WHERE id=?", id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}

func UpdateDate(newDate string, id string) error {
	query := `UPDATE scheduler SET date = ? WHERE id = ?`
	_, err := db.Exec(query, newDate, id)
	if err != nil {
		return fmt.Errorf("ошибка обновления даты задачи: %w", err)
	}
	return nil
}
