package api

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func nextDateHandler(w http.ResponseWriter, r *http.Request) {
	nowStr := r.FormValue("now")
	dateStr := r.FormValue("date")
	repeatStr := r.FormValue("repeat")

	log.Printf("Received parameters: now=%s, date=%s, repeat=%s", nowStr, dateStr, repeatStr)

	now, err := time.Parse(TimeFormat, nowStr)
	if err != nil {
		http.Error(w, "parcing err 'now'", http.StatusBadRequest)
		return
	}

	date, err := time.Parse(TimeFormat, dateStr)
	if err != nil {
		http.Error(w, "parcing err 'date'", http.StatusBadRequest)
		return
	}

	nextDate, err := NextDate(now, date.Format(TimeFormat), repeatStr)
	if err != nil {
		http.Error(w, fmt.Sprintf("err calculate next date: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Next date: %s", nextDate)
	fmt.Fprintf(w, "%s", nextDate)
}
