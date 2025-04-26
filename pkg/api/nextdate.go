package api

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const TimeFormat = "20060102"

func NextDate(now time.Time, dstart string, repeat string) (string, error) {

	date, err := time.Parse(TimeFormat, dstart)
	if err != nil {
		return "", fmt.Errorf("incorrect date format: %v", err)
	}

	parts := strings.Split(repeat, " ")
	if len(parts) == 0 {
		return "", fmt.Errorf("incorrect repeat format: %v", repeat)
	}

	switch parts[0] {
	case "y":
		date = date.AddDate(1, 0, 0)
		for !afterNow(date, now) {
			date = date.AddDate(1, 0, 0)
		}
		return date.Format(TimeFormat), nil

	case "d":
		if len(parts) < 2 {
			return "", fmt.Errorf("interval is not specified")
		}

		days, err := strconv.Atoi(parts[1])
		if err != nil {
			return "", fmt.Errorf("incorrect format of day interval: %v", parts[1])
		}
		if days < 1 || days > 400 {
			return "", fmt.Errorf("days out of range")
		}

		for {
			date = date.AddDate(0, 0, days)
			if afterNow(date, now) {
				break
			}
		}
		return date.Format(TimeFormat), nil

	default:
		return "", fmt.Errorf("incorrect repeat format: %s", repeat)
	}
}

func afterNow(date, now time.Time) bool {
	date = date.Truncate(24 * time.Hour)
	now = now.Truncate(24 * time.Hour)
	return date.After(now)
}
