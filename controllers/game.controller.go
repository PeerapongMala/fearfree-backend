package controllers

import (
	"fearfree-backend/database"
	"fearfree-backend/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// 1. ดึงหมวดหมู่สัตว์ (Reptiles, Insects)
func ListAnimalCategories(c *fiber.Ctx) error {
	var categories []models.AnimalCategory
	if err := database.DB.Find(&categories).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "ดึงข้อมูลหมวดหมู่ไม่สำเร็จ"})
	}
	return c.JSON(fiber.Map{"data": categories})
}

// 2. ดึงสัตว์ในหมวดนั้นๆ (Snake, Spider)
func ListAnimalsByCategory(c *fiber.Ctx) error {
	categoryId := c.Params("categoryId")
	var animals []models.Animal
	if err := database.DB.Where("category_id = ?", categoryId).Find(&animals).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "ดึงข้อมูลสัตว์ไม่สำเร็จ"})
	}
	return c.JSON(fiber.Map{"data": animals})
}

// 3. ดึงด่านของสัตว์ตัวนั้น (Level 1, 2, 3)
func ListStagesByAnimal(c *fiber.Ctx) error {
	animalId := c.Params("animalId")
	var stages []models.Stage

	// Preload Media เพื่อเอารูป/วิดีโอของด่านนั้นมาด้วย
	if err := database.DB.Preload("Media").Where("animal_id = ?", animalId).Order("stage_no asc").Find(&stages).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "ดึงข้อมูลด่านไม่สำเร็จ"})
	}
	return c.JSON(fiber.Map{"data": stages})

}

// ... (ต่อจาก func เดิม)

// Struct สำหรับรับค่าจากหน้าบ้าน (Body)
type SubmitStageInput struct {
	Answer string `json:"answer"` // ส่งมาว่า "pass" หรือ "fail"
}

// ✅ 4. ส่งผลการเล่น (จบด่าน)
func SubmitStageResult(c *fiber.Ctx) error {
	// 1. เตรียมตัวแปร
	userID := c.Locals("user_id").(uint)
	stageID, _ := c.ParamsInt("stageId")
	var input SubmitStageInput

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ข้อมูลไม่ถูกต้อง"})
	}

	// 2. ถ้าเล่นไม่ผ่าน ก็ไม่ต้องทำอะไร
	if input.Answer != "pass" {
		return c.JSON(fiber.Map{"message": "บันทึกผล: ยังไม่ผ่านด่าน"})
	}

	// 3. เริ่ม Transaction (เพราะมีการแก้เงิน + บันทึกผล ต้องทำพร้อมกัน)
	tx := database.DB.Begin()

	// 3.1 ดึงกติกาเกม (เพื่อดูว่าด่านละกี่เหรียญ)
	var rules models.GameRules
	if err := tx.First(&rules, 1).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "ไม่พบการตั้งค่าเกม"})
	}

	// 3.2 เช็คว่าเคยเล่นผ่านด่านนี้ไปหรือยัง? (ถ้าเคยแล้ว ไม่แจกเหรียญซ้ำ)
	var existingResult models.StageResult
	result := tx.Where("user_id = ? AND stage_id = ?", userID, stageID).First(&existingResult)

	if result.RowsAffected == 0 {
		// --- กรณี: เพิ่งผ่านครั้งแรก (First Clear) ---

		// A. สร้างประวัติการเล่นใหม่
		newResult := models.StageResult{
			UserID:  userID,
			StageID: uint(stageID),
			Answer:  input.Answer,
		}
		if err := tx.Create(&newResult).Error; err != nil {
			tx.Rollback()
			return c.Status(500).JSON(fiber.Map{"error": "บันทึกผลไม่สำเร็จ"})
		}

		// B. เพิ่มเหรียญให้ User
		if err := tx.Model(&models.User{}).Where("id = ?", userID).
			Update("balance", gorm.Expr("balance + ?", rules.CoinPerStage)).Error; err != nil {
			tx.Rollback()
			return c.Status(500).JSON(fiber.Map{"error": "อัปเดตเหรียญไม่สำเร็จ"})
		}

		tx.Commit()
		return c.JSON(fiber.Map{
			"message":      "ผ่านด่านสำเร็จ! (ได้รับเหรียญรางวัล)",
			"coins_earned": rules.CoinPerStage,
		})

	} else {
		// --- กรณี: เคยผ่านไปแล้ว (Replay) ---
		// แค่อัปเดตเวลาเล่นล่าสุด ไม่แจกเหรียญเพิ่ม
		existingResult.Answer = input.Answer
		tx.Save(&existingResult)

		tx.Commit()
		return c.JSON(fiber.Map{
			"message":      "ผ่านด่านสำเร็จ! (เคยได้รับรางวัลไปแล้ว)",
			"coins_earned": 0,
		})
	}
}
