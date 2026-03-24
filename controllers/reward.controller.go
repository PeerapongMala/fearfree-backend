package controllers

import (
	"fearfree-backend/database"
	"fearfree-backend/models"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// 1. ดึงรายการของรางวัลทั้งหมด
func ListRewards(c *fiber.Ctx) error {
	var rewards []models.Reward
	if rewards == nil {
		rewards = []models.Reward{}
	}
	// เรียงตามราคา
	if err := database.DB.Order("cost_coins asc").Find(&rewards).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "ดึงข้อมูลรางวัลไม่สำเร็จ"})
	}
	return c.JSON(fiber.Map{"data": rewards})
}

// 2. แลกของรางวัล (ตัดเหรียญ + ตัดสต็อก)
func RedeemReward(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	rewardID, err := c.ParamsInt("rewardId")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "rewardId ไม่ถูกต้อง"})
	}

	tx := database.DB.Begin()

	// ดึง Patient ID ภายใน transaction พร้อม FOR UPDATE lock
	var patient models.Patient
	if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("user_id = ?", userID).First(&patient).Error; err != nil {
		tx.Rollback()
		return c.Status(404).JSON(fiber.Map{"error": "ไม่พบข้อมูลผู้ป่วย"})
	}

	var reward models.Reward
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&reward, rewardID).Error; err != nil {
		tx.Rollback()
		return c.Status(404).JSON(fiber.Map{"error": "ไม่พบของรางวัล"})
	}

	if reward.Stock <= 0 {
		tx.Rollback()
		return c.Status(400).JSON(fiber.Map{"error": "ของรางวัลหมดแล้ว"})
	}

	if patient.Balance < int64(reward.CostCoins) {
		tx.Rollback()
		return c.Status(400).JSON(fiber.Map{"error": "เหรียญไม่พอ"})
	}

	// อัปเดตเหรียญ Patient ด้วย atomic expression
	if err := tx.Model(&patient).Update("balance", gorm.Expr("balance - ?", reward.CostCoins)).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "ตัดเหรียญไม่สำเร็จ"})
	}

	// ตัดสต็อก Reward ด้วย atomic expression
	if err := tx.Model(&reward).Update("stock", gorm.Expr("stock - ?", 1)).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "ตัดสต็อกไม่สำเร็จ"})
	}

	// บันทึกประวัติ
	history := models.RedemptionHistory{
		PatientID: patient.ID,
		RewardID:  uint(rewardID),
	}
	if err := tx.Create(&history).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "บันทึกประวัติไม่สำเร็จ"})
	}

	if err := tx.Commit().Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "บันทึกข้อมูลไม่สำเร็จ"})
	}

	// Re-read patient after commit for accurate response
	database.DB.Where("user_id = ?", userID).First(&patient)

	logAudit(c, userID, "redeem_reward", fmt.Sprintf("Redeemed reward ID: %d (%s), cost: %d coins", reward.ID, reward.Name, reward.CostCoins))

	return c.JSON(fiber.Map{
		"success":         true,
		"message":         "แลกรางวัลสำเร็จ!",
		"remaining_coins": patient.Balance,
	})
}

// 3. ดูประวัติการแลกของฉัน
func GetMyRedemptions(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	var patient models.Patient
	if err := database.DB.Where("user_id = ?", userID).First(&patient).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "ไม่พบข้อมูลผู้ป่วย"})
	}

	histories := []models.RedemptionHistory{}
	if err := database.DB.Preload("Reward").Where("patient_id = ?", patient.ID).Order("redeemed_at desc").Find(&histories).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "ดึงข้อมูลประวัติไม่สำเร็จ"})
	}

	return c.JSON(fiber.Map{"data": histories})
}
