package models

import "time"

type User struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	Age            int       `json:"age"`
	FullName       string    `json:"full_name"`
	Email          string    `json:"email" gorm:"unique"`
	MostFearAnimal string    `json:"most_fear_animal"`
	Balance        int64     `json:"balance" gorm:"default:0"`
	CodePatient    *string   `json:"code_patient" gorm:"unique"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	// Relations
	Auth  *Auth  `json:"-" gorm:"foreignKey:UserID"`
	Roles []Role `json:"roles" gorm:"many2many:users_role;"`
}

type Auth struct {
	ID           uint   `json:"id" gorm:"primaryKey"`
	Username     string `json:"username" gorm:"unique"`
	PasswordHash string `json:"-"`
	UserID       uint   `json:"user_id"`
}

type Role struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	RoleName string `json:"role_name"`
}

// ตั้งชื่อตารางให้ตรงกับ DB
func (User) TableName() string { return "users" }
func (Auth) TableName() string { return "auth" }
func (Role) TableName() string { return "role" }
