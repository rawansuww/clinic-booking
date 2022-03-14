package models

type Patient struct {
	ID       uint          `gorm:"primarykey" json:"id,omitempty"`
	Name     string        `json:"name"`
	Email    string        `json:"email"`
	Password string        `json:"password,omitempty"`
	Role     string        `json:"role,omitempty"`
	History  []Appointment `gorm:"foreignKey:p_id" json:"history,omitempty"`
}
