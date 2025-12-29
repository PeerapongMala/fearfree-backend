package models

import "time"

// คลังข้อสอบ (Assessment Store)
type AssessmentStore struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	Version  int    `json:"version"`
	Seq      int    `json:"seq"`
	Prompt   string `json:"prompt"` // คำถาม
	IsActive bool   `json:"is_active"`
}

// ผลการประเมิน (Assessment Result)
type AssessmentResult struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	AssessmentID uint      `json:"assessment_id"` // อ้างอิง version ของข้อสอบ (ในที่นี้ใช้ ID แทนกลุ่ม)
	FearLevel    string    `json:"fear_level"`    // low, medium, high
	Percent      float64   `json:"percent"`
	UserID       uint      `json:"user_id"`
	CreatedAt    time.Time `json:"created_at"`
}

// ตั้งชื่อตาราง
func (AssessmentStore) TableName() string  { return "assessment_store" }
func (AssessmentResult) TableName() string { return "assessment_result" }
