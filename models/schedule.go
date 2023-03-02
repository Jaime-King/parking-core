package models

import "time"

type Schedule struct {
	Username     string
	StartTime    time.Time
	EndTime      time.Time
	Area         int64
	NextParkTime time.Time
	Progress     string
	Message      string
	Sessions     int64
}
