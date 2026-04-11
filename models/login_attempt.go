package models

import "time"

type LoginAttempt struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	Username    string     `json:"username" gorm:"index;not null"`
	Attempts    int        `json:"attempts" gorm:"default:0"`
	LockedUntil *time.Time `json:"locked_until"`
	UpdatedAt   time.Time
}
