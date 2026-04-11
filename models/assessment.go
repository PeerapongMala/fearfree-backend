package models

import "time"

type FearLevel string

const (
	FearLow    FearLevel = "low"
	FearMedium FearLevel = "medium"
	FearHigh   FearLevel = "high"
)

type Assessment struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	PatientID       uint      `json:"patient_id" gorm:"index"`
	InitialScore    int       `json:"initial_score"`
	CalculatedLevel FearLevel `json:"calculated_level" gorm:"type:fear_level"`
	CreatedAt       time.Time `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`

	Patient Patient `json:"-" gorm:"foreignKey:PatientID"`
}

func (Assessment) TableName() string { return "assessments" }
