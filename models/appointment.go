package models

import (
	"time"

	"gorm.io/gorm"
)

type Appointment struct {
	gorm.Model
	//AID       uint          `gorm:"primaryKey" json:"aID"`
	PID       uint          `json:"pID"`
	DID       uint          `json:"dID"`
	Booked    bool          `json:"booked"`
	StartTime time.Time     `json:"startTime"`
	EndTime   time.Time     `json:"dateTime"`
	Duration  time.Duration `json:"duration"`
}
