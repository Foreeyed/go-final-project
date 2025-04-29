package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

var db *sql.DB

const schema = `CREATE TABLE IF NOT EXISTS scheduler (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date CHAR(8) NOT NULL DEFAULT "",
    title VARCHAR(255) NOT NULL,
    comment TEXT,
    repeat VARCHAR(128)
);
CREATE INDEX IF NOT EXISTS idx_date ON scheduler(date);`

func InitDB(dbFile string) error {

	dbFileWithPath, ok := os.LookupEnv("TODO_DBFILE")
	if ok {
		dbFile = dbFileWithPath
	}

	_, err := os.Stat(dbFile)
	if err != nil {
		fmt.Println("database not found, creating...")
	}

	db, err = sql.Open("sqlite", dbFile)
	if err != nil {
		return fmt.Errorf("can not open db with err: %v", err)
	}

	_, err = db.Exec(schema)
	if err != nil {
		return fmt.Errorf("can not create schema with err: %v", err)
	}

	return nil
}


func GetDB() *sql.DB {
	return db
}