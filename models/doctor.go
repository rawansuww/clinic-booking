package models

import "gorm.io/gorm"

type Doctor struct {
	gorm.Model
	//	ID           uint          `gorm:"primarykey" json:"ID,omitempty"`
	Name         string        `json:"name"`
	Schedule     []Appointment `gorm:"foreignKey:d_id" json:"schedule,omitempty"`
	Email        string        `json:"email"`
	Password     string        `json:"password,omitempty"`
	Availability []string      `gorm:"type:text" json:"availability,omitempty"`
	Role         string        `json:"role,omitempty"`
}
