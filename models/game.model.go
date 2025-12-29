package models

import "time"

// --- หมวดหมู่สัตว์ & สัตว์ ---
type AnimalCategory struct {
	ID    uint   `json:"id" gorm:"primaryKey"`
	Cname string `json:"cname"`
}

type Animal struct {
	ID         uint   `json:"id" gorm:"primaryKey"`
	Aname      string `json:"name" gorm:"column:aname"` // ชื่อใน DB คือ aname
	CategoryID uint   `json:"category_id"`
}

// --- ไฟล์สื่อ (รูป/วิดีโอ) ---
type MediaStore struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	MediaType string    `json:"media_type"`
	Mime      string    `json:"mime"`
	Checksum  string    `json:"checksum"`
	CreatedAt time.Time `json:"created_at"`
}

// --- ระบบด่าน (Gameplay) ---
type Stage struct {
	ID       uint `json:"id" gorm:"primaryKey"`
	StageNo  int  `json:"stage_no"`
	AnimalID uint `json:"animal_id"`
	MediaID  uint `json:"media_id"`

	// Preload (จอยตาราง)
	Media  *MediaStore `json:"media" gorm:"foreignKey:MediaID"`
	Animal *Animal     `json:"animal" gorm:"foreignKey:AnimalID"`
}

type StageResult struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id"`
	StageID   uint      `json:"stage_id"`
	Answer    string    `json:"answer"`
	UpdatedAt time.Time `json:"updated_at"`
}

type GameRules struct {
	ID                   uint `json:"id" gorm:"primaryKey"`
	StageDurationSeconds int  `json:"stage_duration_seconds"`
	CoinPerStage         int  `json:"coin_per_stage"`
}

// --- ของรางวัล (Rewards) ---
type Reward struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CostCoins   int       `json:"cost_coins"`
	Stock       int       `json:"stock"`
	ImageURL    string    `json:"image_url"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}

type RewardsUser struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id"`
	RewardID  uint      `json:"reward_id"`
	Status    string    `json:"status"` // 'reserved', 'fulfilled'
	CreatedAt time.Time `json:"created_at"`

	Reward *Reward `json:"reward" gorm:"foreignKey:RewardID"`
}

// ตั้งชื่อตารางให้ตรงกับ DB
func (AnimalCategory) TableName() string { return "animal_categories" }
func (Animal) TableName() string         { return "animal" }
func (MediaStore) TableName() string     { return "media_store" }
func (Stage) TableName() string          { return "stages" }
func (StageResult) TableName() string    { return "stage_result" }
func (GameRules) TableName() string      { return "game_rules" }
func (Reward) TableName() string         { return "rewards" }
func (RewardsUser) TableName() string    { return "rewards_users" }
