package user

import (
	"fearfree-backend/database"
	"fearfree-backend/models"

	"github.com/gofiber/fiber/v2"
)

// ProfileResponse is an explicit DTO that exposes only safe fields.
type ProfileResponse struct {
	ID              uint   `json:"id"`
	Username        string `json:"username"`
	Email           string `json:"email"`
	Role            string `json:"role"`
	FullName        string `json:"full_name"`
	Age             int    `json:"age"`
	MostFearAnimal  string `json:"most_fear_animal"`
	FearLevel       string `json:"fear_level"`
	Balance         int64  `json:"coins"`
	FearPercentage  float64 `json:"fear_percentage"`
	FearLevelText   string `json:"fear_level_text"`
}

// fearLevelToPercentage maps fear level enum to a representative percentage.
func fearLevelToPercentage(level string) float64 {
	switch level {
	case "high":
		return 85.0
	case "medium":
		return 50.0
	default:
		return 15.0
	}
}

// fearLevelToText maps fear level enum to a human-readable Thai description.
func fearLevelToText(level string) string {
	switch level {
	case "high":
		return "สูง"
	case "medium":
		return "ปานกลาง"
	default:
		return "ต่ำ"
	}
}

// GET /auth/v1/me
func GetProfile(c *fiber.Ctx) error {
	// 1. ดึง UserID ที่ Middleware แปะไว้ให้
	userID := c.Locals("user_id").(uint)

	// 2. Query ข้อมูลจาก Database โดย Preload Patient มาด้วย
	var user models.User
	if err := database.DB.Preload("Patient").First(&user, userID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "ไม่พบผู้ใช้งาน"})
	}

	// 3. Build safe response DTO — never expose internal IDs or raw model
	resp := ProfileResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     string(user.Role),
	}

	if user.Patient != nil {
		resp.FullName = user.Patient.FullName
		resp.Age = user.Patient.Age
		resp.MostFearAnimal = user.Patient.MostFearAnimal
		resp.FearLevel = user.Patient.FearLevel
		resp.Balance = user.Patient.Balance
		resp.FearPercentage = fearLevelToPercentage(user.Patient.FearLevel)
		resp.FearLevelText = fearLevelToText(user.Patient.FearLevel)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
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

	// Validate input fields
	if len(input.FullName) > 100 {
		return c.Status(400).JSON(fiber.Map{"error": "ชื่อต้องมีความยาวไม่เกิน 100 ตัวอักษร"})
	}
	if input.Age != 0 && (input.Age < 1 || input.Age > 120) {
		return c.Status(400).JSON(fiber.Map{"error": "อายุต้องอยู่ระหว่าง 1-120 ปี"})
	}
	if len(input.MostFearAnimal) > 100 {
		return c.Status(400).JSON(fiber.Map{"error": "ชื่อสัตว์ต้องมีความยาวไม่เกิน 100 ตัวอักษร"})
	}

	var patient models.Patient
	if err := database.DB.Where("user_id = ?", userID).First(&patient).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "ไม่พบข้อมูลโปรไฟล์ผู้ป่วย"})
	}

	// อัปเดตข้อมูล
	patient.FullName = input.FullName
	patient.Age = input.Age
	patient.MostFearAnimal = input.MostFearAnimal

	if err := database.DB.Save(&patient).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "อัปเดตข้อมูลไม่สำเร็จ"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "อัปเดตข้อมูลสำเร็จ",
		"data": fiber.Map{
			"full_name":        patient.FullName,
			"age":              patient.Age,
			"most_fear_animal": patient.MostFearAnimal,
			"fear_level":       patient.FearLevel,
			"coins":            patient.Balance,
		},
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

	// Transform to match frontend MyPlayHistoryItem - initialize to empty slice
	result := []fiber.Map{}
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

// GET /users/me/redemptions - ดูประวัติการแลกของฉัน
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
