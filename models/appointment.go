package models

import (
	"time"

	"gorm.io/gorm"
)

type Appointment struct {
	gorm.Model
	PID       uint          `json:"pID"`
	DID       uint          `json:"dID"`
	StartTime time.Time     `json:"startTime"`
	EndTime   time.Time     `json:"dateTime"`
	Duration  time.Duration `json:"duration"`
}
