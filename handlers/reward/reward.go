package reward

import (
	"fearfree-backend/database"
	"fearfree-backend/handlers/shared"
	"fearfree-backend/models"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// 1. ดึงรายการของรางวัลทั้งหมด
func ListRewards(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	var total int64
	database.DB.Model(&models.Reward{}).Count(&total)

	rewards := []models.Reward{}
	if err := database.DB.Order("cost_coins asc").Offset(offset).Limit(limit).Find(&rewards).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "ดึงข้อมูลรางวัลไม่สำเร็จ"})
	}
	return c.JSON(fiber.Map{
		"success": true,
		"data":    rewards,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
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

	var rwd models.Reward
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&rwd, rewardID).Error; err != nil {
		tx.Rollback()
		return c.Status(404).JSON(fiber.Map{"error": "ไม่พบของรางวัล"})
	}

	if rwd.Stock <= 0 {
		tx.Rollback()
		return c.Status(400).JSON(fiber.Map{"error": "ของรางวัลหมดแล้ว"})
	}

	if patient.Balance < int64(rwd.CostCoins) {
		tx.Rollback()
		return c.Status(400).JSON(fiber.Map{"error": "เหรียญไม่พอ"})
	}

	// อัปเดตเหรียญ Patient ด้วย atomic expression
	if err := tx.Model(&patient).Update("balance", gorm.Expr("balance - ?", rwd.CostCoins)).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "ตัดเหรียญไม่สำเร็จ"})
	}

	// ตัดสต็อก Reward ด้วย atomic expression
	if err := tx.Model(&rwd).Update("stock", gorm.Expr("stock - ?", 1)).Error; err != nil {
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

	shared.LogAudit(c, userID, "redeem_reward", fmt.Sprintf("Redeemed reward ID: %d (%s), cost: %d coins", rwd.ID, rwd.Name, rwd.CostCoins))

	return c.JSON(fiber.Map{
		"success":         true,
		"message":         "แลกรางวัลสำเร็จ!",
		"remaining_coins": patient.Balance,
	})
}
