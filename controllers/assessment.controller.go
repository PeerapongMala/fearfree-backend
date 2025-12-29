package controllers

import (
	"fearfree-backend/database"
	"fearfree-backend/models"

	"github.com/gofiber/fiber/v2"
)

// GET /assessment/v1 (ดึงคำถาม)
func GetAssessments(c *fiber.Ctx) error {
	var questions []models.AssessmentStore
	// ดึงเฉพาะที่ Active และเรียงตามลำดับ (Seq)
	if err := database.DB.Where("is_active = ?", true).Order("seq asc").Find(&questions).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "ดึงแบบประเมินไม่สำเร็จ"})
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

	// 1. คำนวณคะแนนรวม
	totalScore := 0
	maxScore := len(input.Answers) * 5 // สมมติคะแนนเต็มข้อละ 5

	for _, ans := range input.Answers {
		totalScore += ans.Score
	}

	// 2. คำนวณเปอร์เซ็นต์
	percent := (float64(totalScore) / float64(maxScore)) * 100

	// 3. ตัดเกรด (Logic ตัวอย่าง)
	fearLevel := "low"
	if percent > 70 {
		fearLevel = "high"
	} else if percent > 30 {
		fearLevel = "medium"
	}

	// ใน func SubmitAssessment ...

	// หา ID ของแบบประเมินเวอร์ชันล่าสุดมาใช้ (แทนการใส่เลข 1 ดื้อๆ)
	var assessment models.AssessmentStore
	if err := database.DB.Where("version = ?", 1).First(&assessment).Error; err != nil {
		// ถ้าหาไม่เจอ ให้ default เป็น 0 หรือ handle error
		return c.Status(500).JSON(fiber.Map{"error": "ไม่พบแบบประเมินในระบบ"})
	}

	// 4. บันทึกผลลง DB
	result := models.AssessmentResult{
		UserID:       userID,
		FearLevel:    fearLevel,
		Percent:      percent,
		AssessmentID: assessment.ID,
	}

	if err := database.DB.Create(&result).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "บันทึกผลไม่สำเร็จ"})
	}

	return c.JSON(fiber.Map{
		"message":    "ประเมินผลสำเร็จ",
		"fear_level": fearLevel,
		"percent":    percent,
	})
}
