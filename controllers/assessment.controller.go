package controllers

import (
	"fearfree-backend/database"
	"fearfree-backend/models"

	"github.com/gofiber/fiber/v2"
)

// GET /assessment/v1 (ดึงคำถาม)
func GetAssessments(c *fiber.Ctx) error {
	// จำลองชุดคำถาม 5 ข้อ เนื่องจากไม่ได้เก็บไว้ในฐานข้อมูลตาม Schema ล่าสุด
	questions := []fiber.Map{
		{"id": 1, "prompt": "คุณรู้สึกกลัวเมื่อเห็นรูปภาพของสัตว์ที่คุณกลัวหรือไม่?"},
		{"id": 2, "prompt": "คุณรู้สึกใจสั่นเมื่ออยู่ในที่ที่คิดว่าสัตว์ที่คุณกลัวอาจจะอยู่หรือไม่?"},
		{"id": 3, "prompt": "คุณหลีกเลี่ยงการไปสถานที่บางแห่งเนื่องจากกลัวว่าจะเจอสัตว์นั้นหรือไม่?"},
		{"id": 4, "prompt": "ภาพเคลื่อนไหวของสัตว์เหล่านั้นทำให้คุณรู้สึกตื่นตระหนกหรือไม่?"},
		{"id": 5, "prompt": "คุณมีความกังวลอยู่ตลอดเวลาว่าจะบังเอิญเจอสัตว์นั้นในชีวิตประจำวันหรือไม่?"},
	}
	return c.JSON(fiber.Map{"data": questions})
}

// Input สำหรับส่งคำตอบ
type AnswerItem struct {
	QuestionID uint `json:"question_id"`
	Score      int  `json:"score"` // 1-5
}

type SubmitAssessmentInput struct {
	Answers []AnswerItem `json:"answers"`
}

// POST /assessment/v1/submit (ส่งคำตอบ + คำนวณผล)
func SubmitAssessment(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	var input SubmitAssessmentInput

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ข้อมูลไม่ถูกต้อง"})
	}

	totalScore := 0
	maxScore := len(input.Answers) * 5

	if maxScore == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "ไม่พบคำตอบ"})
	}

	for _, ans := range input.Answers {
		totalScore += ans.Score
	}

	percent := (float64(totalScore) / float64(maxScore)) * 100

	fearLevel := models.FearLow
	description := "คุณมีความกลัวในระดับต่ำ สามารถใช้ชีวิตประจำวันได้ปกติ"
	if percent > 70 {
		fearLevel = models.FearHigh
		description = "ความกลัวของคุณอยู่ในระดับสูง แนะนำให้ค่อยๆ เปิดใจและทดสอบกับแอปพลิเคชันของเราอย่างต่อเนื่องเพื่อปรับตัว"
	} else if percent > 30 {
		fearLevel = models.FearMedium
		description = "คุณมีความกลัวในระดับปานกลาง ลองเรียนรู้และค่อยๆ ก้าวผ่านไปกับด่านในเกมของเรานะครับ"
	}

	var patient models.Patient
	if err := database.DB.Where("user_id = ?", userID).First(&patient).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "ไม่พบข้อมูลผู้ป่วย"})
	}

	// บันทึกผลลง DB
	result := models.Assessment{
		PatientID:       patient.ID,
		InitialScore:    totalScore,
		CalculatedLevel: fearLevel,
	}

	if err := database.DB.Create(&result).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "บันทึกผลไม่สำเร็จ"})
	}

	// อัปเดต fear_level ใน Patient ด้วย
	database.DB.Model(&patient).Update("fear_level", fearLevel)

	return c.JSON(fiber.Map{
		"message":     "ประเมินผลสำเร็จ",
		"fear_level":  fearLevel,
		"percent":     percent,
		"description": description,
	})
}
