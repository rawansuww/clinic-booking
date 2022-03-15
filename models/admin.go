package models

import "gorm.io/gorm"

type Admin struct {
	gorm.Model
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}
