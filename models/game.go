package models

import "time"

type MediaType string

const (
	MediaImage MediaType = "image"
	MediaVideo MediaType = "video"
)

type ProgressStatus string

const (
	StatusLocked     ProgressStatus = "locked"
	StatusInProgress ProgressStatus = "in_progress"
	StatusCompleted  ProgressStatus = "completed"
)

type AnimalCategory struct {
	ID          uint     `json:"id" gorm:"primaryKey"`
	Name        string   `json:"name" gorm:"unique"`
	Description string   `json:"description"`
	Animals     []Animal `json:"animals" gorm:"foreignKey:CategoryID"`
}

type Animal struct {
	ID           uint    `json:"id" gorm:"primaryKey"`
	CategoryID   uint    `json:"category_id" gorm:"index"`
	Name         string  `json:"name" gorm:"unique"`
	Description  string  `json:"description"`
	ThumbnailUrl string  `json:"thumbnail_url"`
	Stages       []Stage `json:"stages" gorm:"foreignKey:AnimalID"`
}

type Stage struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	AnimalID       uint      `json:"animal_id" gorm:"index"`
	StageNo        int       `json:"stage_no"`
	MediaType      MediaType `json:"media_type" gorm:"type:media_type_enum"`
	MediaUrl       string    `json:"media_url"`
	DisplayTimeSec int       `json:"display_time_sec"`
	RewardCoins    int       `json:"reward_coins"`

	Animal Animal `json:"animal" gorm:"foreignKey:AnimalID"`
}

type PatientProgress struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	PatientID   uint           `json:"patient_id" gorm:"uniqueIndex:idx_patient_stage"`
	StageID     uint           `json:"stage_id" gorm:"uniqueIndex:idx_patient_stage"`
	Status      ProgressStatus `json:"status" gorm:"type:progress_status;default:'locked'"`
	SymptomNote string         `json:"symptom_note"`
	UnlockDate  *time.Time     `json:"unlock_date"`
	CompletedAt *time.Time     `json:"completed_at"`

	Stage Stage `json:"stage" gorm:"foreignKey:StageID"`
}

func (AnimalCategory) TableName() string  { return "animal_categories" }
func (Animal) TableName() string          { return "animals" }
func (Stage) TableName() string           { return "stages" }
func (PatientProgress) TableName() string { return "patient_progress" }
