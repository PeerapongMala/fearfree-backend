package models

import "time"

type Reward struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CostCoins   int    `json:"cost_coins" gorm:"not null;check:cost_coins > 0"`
	Stock       int    `json:"stock" gorm:"not null;check:stock >= 0"`
	ImageUrl    string `json:"image_url"`
}

type RedemptionHistory struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	PatientID  uint      `json:"patient_id" gorm:"index"`
	RewardID   uint      `json:"reward_id" gorm:"index"`
	RedeemedAt time.Time `json:"redeemed_at" gorm:"default:CURRENT_TIMESTAMP"`

	Reward Reward `json:"reward" gorm:"foreignKey:RewardID"`
}

func (Reward) TableName() string            { return "rewards" }
func (RedemptionHistory) TableName() string { return "redemption_history" }
