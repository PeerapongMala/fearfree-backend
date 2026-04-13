package game

import (
	"fearfree-backend/database"
	"fearfree-backend/models"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// 1. ดึงหมวดหมู่สัตว์ (Reptiles, Insects)
func ListAnimalCategories(c *fiber.Ctx) error {
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
	database.DB.Model(&models.AnimalCategory{}).Count(&total)

	categories := []models.AnimalCategory{}
	if err := database.DB.Offset(offset).Limit(limit).Find(&categories).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "ดึงข้อมูลหมวดหมู่ไม่สำเร็จ"})
	}
	return c.JSON(fiber.Map{
		"success": true,
		"data":    categories,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}

// 2. ดึงสัตว์ในหมวดนั้นๆ (Snake, Spider)
func ListAnimalsByCategory(c *fiber.Ctx) error {
	categoryId := c.Params("categoryId")

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
	database.DB.Model(&models.Animal{}).Where("category_id = ?", categoryId).Count(&total)

	animals := []models.Animal{}
	if err := database.DB.Where("category_id = ?", categoryId).Offset(offset).Limit(limit).Find(&animals).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "ดึงข้อมูลสัตว์ไม่สำเร็จ"})
	}
	return c.JSON(fiber.Map{
		"success": true,
		"data":    animals,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}

// 3. ดึงด่านของสัตว์ตัวนั้น (Level 1, 2, 3)
func ListStagesByAnimal(c *fiber.Ctx) error {
	animalId := c.Params("animalId")
	stages := []models.Stage{}

	if err := database.DB.Where("animal_id = ?", animalId).Order("stage_no asc").Find(&stages).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "ดึงข้อมูลด่านไม่สำเร็จ"})
	}
	return c.JSON(fiber.Map{"data": stages})
}

// Struct สำหรับรับค่าจากหน้าบ้าน (Body)
type SubmitStageInput struct {
	Answer      string `json:"answer"`       // ส่งมาว่า "pass" หรือ "fail"
	SymptomNote string `json:"symptom_note"` // บันทึกอาการกลัว
}

// 4. ส่งผลการเล่น (จบด่าน) อัปเดตตาราง PatientProgress
func SubmitStageResult(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	levelID, err := c.ParamsInt("levelId")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "levelId ไม่ถูกต้อง"})
	}
	var input SubmitStageInput

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ข้อมูลไม่ถูกต้อง"})
	}

	if len(input.SymptomNote) > 2000 {
		return c.Status(400).JSON(fiber.Map{"error": "บันทึกอาการต้องมีความยาวไม่เกิน 2000 ตัวอักษร"})
	}

	var patient models.Patient
	if err := database.DB.Where("user_id = ?", userID).First(&patient).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "ไม่พบผู้ป่วย"})
	}

	var stage models.Stage
	if err := database.DB.First(&stage, levelID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "ไม่พบด่านนี้"})
	}

	if input.Answer != "pass" {
		return c.JSON(fiber.Map{"message": "บันทึกผล: ยังไม่ผ่านด่าน"})
	}

	// ค้นหาด่านถัดไปของสัตว์ตัวเดิม
	var nextStage models.Stage
	hasNext := false
	if err := database.DB.Where("animal_id = ? AND stage_no = ?", stage.AnimalID, stage.StageNo+1).First(&nextStage).Error; err == nil {
		hasNext = true
	}

	tx := database.DB.Begin()

	var progress models.PatientProgress
	result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("patient_id = ? AND stage_id = ?", patient.ID, levelID).First(&progress)

	now := time.Now()

	if result.RowsAffected == 0 {
		// First Clear
		newProgress := models.PatientProgress{
			PatientID:   patient.ID,
			StageID:     uint(levelID),
			Status:      models.StatusCompleted,
			SymptomNote: input.SymptomNote,
			CompletedAt: &now,
			UnlockDate:  &now,
		}
		if err := tx.Create(&newProgress).Error; err != nil {
			tx.Rollback()
			return c.Status(500).JSON(fiber.Map{"error": "บันทึกผลไม่สำเร็จ"})
		}

		// Add Coins to Patient
		if err := tx.Model(&patient).Update("balance", gorm.Expr("balance + ?", stage.RewardCoins)).Error; err != nil {
			tx.Rollback()
			return c.Status(500).JSON(fiber.Map{"error": "อัปเดตเหรียญไม่สำเร็จ"})
		}

		if err := tx.Commit().Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "บันทึกข้อมูลไม่สำเร็จ"})
		}
		return c.JSON(fiber.Map{
			"success":      true,
			"message":      "ผ่านด่านสำเร็จ! (ได้รับเหรียญรางวัล)",
			"earned_coins": stage.RewardCoins,
			"next_stage": fiber.Map{
				"has_next": hasNext,
				"stage_id": nextStage.ID,
				"stage_no": nextStage.StageNo,
			},
		})

	} else {
		// Replay
		progress.Status = models.StatusCompleted
		progress.SymptomNote = input.SymptomNote
		progress.CompletedAt = &now
		if err := tx.Save(&progress).Error; err != nil {
			tx.Rollback()
			return c.Status(500).JSON(fiber.Map{"error": "บันทึกผลไม่สำเร็จ"})
		}

		if err := tx.Commit().Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "บันทึกข้อมูลไม่สำเร็จ"})
		}
		return c.JSON(fiber.Map{
			"success":      true,
			"message":      "ผ่านด่านสำเร็จ! (เคยได้รับรางวัลไปแล้ว)",
			"earned_coins": 0,
			"next_stage": fiber.Map{
				"has_next": hasNext,
				"stage_id": nextStage.ID,
				"stage_no": nextStage.StageNo,
			},
		})
	}
}
