package models

import "time"

type User struct {
	Username		string
	Name 			string
	AtEmail			string
	AtPassword 		string
	Plate 			string
	CycleLength		time.Duration
}