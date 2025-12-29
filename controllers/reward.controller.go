package controllers

import (
	"fearfree-backend/database"
	"fearfree-backend/models"

	"github.com/gofiber/fiber/v2"
)

// ✅ 1. ดึงรายการของรางวัลทั้งหมด
func ListRewards(c *fiber.Ctx) error {
	var rewards []models.Reward
	// ดึงเฉพาะของที่เปิดใช้งาน (IsActive = true) และเรียงตามราคา
	if err := database.DB.Where("is_active = ?", true).Order("cost_coins asc").Find(&rewards).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "ดึงข้อมูลรางวัลไม่สำเร็จ"})
	}
	return c.JSON(fiber.Map{"data": rewards})
}

// ✅ 2. แลกของรางวัล (ตัดเหรียญ + ตัดสต็อก)
func RedeemReward(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	rewardID, _ := c.ParamsInt("rewardId")

	// เริ่ม Transaction (สำคัญมาก! เพื่อป้องกันข้อมูลพังกลางคัน)
	tx := database.DB.Begin()

	// A. ตรวจสอบข้อมูลรางวัล และ ล็อกแถว (Locking) กันคนแย่งกดพร้อมกัน
	var reward models.Reward
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&reward, rewardID).Error; err != nil {
		tx.Rollback()
		return c.Status(404).JSON(fiber.Map{"error": "ไม่พบของรางวัล"})
	}

	// B. เช็คสต็อก
	if reward.Stock <= 0 {
		tx.Rollback()
		return c.Status(400).JSON(fiber.Map{"error": "ของรางวัลหมดแล้ว"})
	}

	// C. เช็คเงิน User
	var user models.User
	if err := tx.First(&user, userID).Error; err != nil {
		tx.Rollback()
		return c.Status(404).JSON(fiber.Map{"error": "ไม่พบผู้ใช้"})
	}

	if user.Balance < int64(reward.CostCoins) {
		tx.Rollback()
		return c.Status(400).JSON(fiber.Map{"error": "เหรียญไม่พอ"})
	}

	// D. ทำการแลก (ตัดเหรียญ - ตัดสต็อก - สร้างประวัติ)
	// 1. ตัดเหรียญ User
	if err := tx.Model(&user).Update("balance", user.Balance-int64(reward.CostCoins)).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "ตัดเหรียญไม่สำเร็จ"})
	}

	// 2. ตัดสต็อก Reward
	if err := tx.Model(&reward).Update("stock", reward.Stock-1).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "ตัดสต็อกไม่สำเร็จ"})
	}

	// 3. บันทึกประวัติ (RewardsUser)
	history := models.RewardsUser{
		UserID:   userID,
		RewardID: uint(rewardID),
		Status:   "fulfilled",
	}
	if err := tx.Create(&history).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "บันทึกประวัติไม่สำเร็จ"})
	}

	// ถ้าทุกอย่างผ่านหมด -> Commit ยืนยันการบันทึก
	tx.Commit()

	return c.JSON(fiber.Map{
		"message": "แลกรางวัลสำเร็จ!",
		"data":    history,
	})
}

// ✅ 3. ดูประวัติการแลกของฉัน
func GetMyRedemptions(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	var histories []models.RewardsUser

	// Preload Reward เพื่อให้เห็นชื่อของรางวัลด้วย
	if err := database.DB.Preload("Reward").Where("user_id = ?", userID).Order("created_at desc").Find(&histories).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "ดึงข้อมูลประวัติไม่สำเร็จ"})
	}

	return c.JSON(fiber.Map{"data": histories})
}
