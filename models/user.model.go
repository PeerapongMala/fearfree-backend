package models

import "time"

type UserRole string

const (
	RolePatient UserRole = "patient"
	RoleDoctor  UserRole = "doctor"
	RoleAdmin   UserRole = "admin"
)

type User struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Username     string    `json:"username" gorm:"unique;not null"`
	Email        string    `json:"email" gorm:"unique;not null"`
	PasswordHash string    `json:"-" gorm:"not null"`
	Role         UserRole  `json:"role" gorm:"type:user_role;default:'patient'"`
	CreatedAt    time.Time `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`

	Patient *Patient `json:"patient,omitempty" gorm:"foreignKey:UserID"`
}

type Patient struct {
	ID                uint      `json:"id" gorm:"primaryKey"`
	UserID            uint      `json:"user_id" gorm:"uniqueIndex"`
	CreatedByDoctorID uint      `json:"created_by_doctor_id" gorm:"index"`
	FullName          string    `json:"full_name"`
	Age               int       `json:"age"`
	MostFearAnimal    string    `json:"most_fear_animal"`
	FearLevel         string    `json:"fear_level" gorm:"type:fear_level"`
	Balance           int64     `json:"balance" gorm:"default:0"`
	CodePatient       *string   `json:"code_patient" gorm:"unique"`
	CreatedAt         time.Time `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt         time.Time `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`
}

func (User) TableName() string    { return "users" }
func (Patient) TableName() string { return "patients" }
