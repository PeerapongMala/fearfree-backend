package controllers

import (
	"fearfree-backend/database"
	"fearfree-backend/models"

	"github.com/gofiber/fiber/v2"
)

// GET /auth/v1/me
func GetProfile(c *fiber.Ctx) error {
	// 1. ดึง UserID ที่ Middleware แปะไว้ให้
	userID := c.Locals("user_id").(uint)

	// 2. Query ข้อมูลจาก Database โดย Preload Patient มาด้วย
	var user models.User
	if err := database.DB.Preload("Patient").First(&user, userID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "ไม่พบผู้ใช้งาน"})
	}

	// 3. ส่งข้อมูลกลับ
	return c.JSON(fiber.Map{
		"status": "success",
		"data":   user,
	})
}

type UpdateProfileInput struct {
	FullName       string `json:"full_name"`
	Age            int    `json:"age"`
	MostFearAnimal string `json:"most_fear_animal"`
}

// PATCH /auth/v1/me
func UpdateProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	var input UpdateProfileInput

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ข้อมูลไม่ถูกต้อง"})
	}

	var patient models.Patient
	if err := database.DB.Where("user_id = ?", userID).First(&patient).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "ไม่พบข้อมูลโปรไฟล์ผู้ป่วย"})
	}

	// อัปเดตข้อมูล
	patient.FullName = input.FullName
	patient.Age = input.Age
	patient.MostFearAnimal = input.MostFearAnimal

	database.DB.Save(&patient)

	return c.JSON(fiber.Map{
		"message": "อัปเดตข้อมูลสำเร็จ",
		"data":    patient,
	})
}

// GET /users/play-history
func GetMyPlayHistory(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	var patient models.Patient
	if err := database.DB.Where("user_id = ?", userID).First(&patient).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "ไม่พบข้อมูลโปรไฟล์ผู้ป่วย"})
	}

	var history []models.PatientProgress
	if err := database.DB.Preload("Stage").Preload("Stage.Animal").Where("patient_id = ?", patient.ID).Order("completed_at desc").Find(&history).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "ดึงข้อมูลประวัติการเล่นไม่สำเร็จ"})
	}

	// Transform to match frontend MyPlayHistoryItem
	var result []fiber.Map
	for _, h := range history {
		animalName := ""
		stageNo := 0
		if h.Stage.ID != 0 {
			stageNo = h.Stage.StageNo
			if h.Stage.Animal.ID != 0 {
				animalName = h.Stage.Animal.Name
			}
		}

		result = append(result, fiber.Map{
			"id":           h.ID,
			"animal_name":  animalName,
			"stage_no":     stageNo,
			"earned_coins": h.Stage.RewardCoins,
			"completed_at": h.CompletedAt,
			"symptom_note": h.SymptomNote,
		})
	}

	return c.JSON(fiber.Map{
		"data": result,
	})
}
