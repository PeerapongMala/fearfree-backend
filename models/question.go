package models

type Question struct {
	ID     uint   `json:"id" gorm:"primaryKey"`
	Prompt string `json:"prompt" gorm:"type:text;not null"`
	SortOrder int `json:"sort_order" gorm:"not null;default:0"`
}

func (Question) TableName() string { return "questions" }
