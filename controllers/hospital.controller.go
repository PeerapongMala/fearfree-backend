package controllers

import (
	"fearfree-backend/database"
	"fearfree-backend/models"

	"github.com/gofiber/fiber/v2"
)

// ✅ 1. ดึงรายชื่อโรงพยาบาลทั้งหมด (สำหรับทำ Dropdown ให้เลือก)
func ListHospitals(c *fiber.Ctx) error {
	var hospitals []models.Hospital
	if err := database.DB.Find(&hospitals).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "ดึงข้อมูลโรงพยาบาลไม่สำเร็จ"})
	}
	return c.JSON(fiber.Map{"data": hospitals})
}

// Struct สำหรับรับค่า input
type SelectHospitalInput struct {
	HospitalID uint `json:"hospital_id"`
}

// ✅ 2. เลือกสังกัดโรงพยาบาล (User กดเลือก)
func SelectHospital(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	var input SelectHospitalInput

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ข้อมูลไม่ถูกต้อง"})
	}

	// เช็คว่าโรงพยาบาลนี้มีอยู่จริงไหม?
	var hospital models.Hospital
	if err := database.DB.First(&hospital, input.HospitalID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "ไม่พบโรงพยาบาลที่เลือก"})
	}

	// เช็คว่า User เคยเลือกไปแล้วหรือยัง? (ถ้าเคยแล้วให้ Update, ถ้ายังให้ Create)
	var userHospital models.UserHospital
	result := database.DB.Where("user_id = ?", userID).First(&userHospital)

	if result.RowsAffected > 0 {
		// กรณี: เคยเลือกแล้ว -> เปลี่ยนโรงพยาบาลใหม่
		userHospital.HospitalID = input.HospitalID
		database.DB.Save(&userHospital)
	} else {
		// กรณี: ยังไม่เคยเลือก -> สร้างใหม่
		newUserHospital := models.UserHospital{
			UserID:     userID,
			HospitalID: input.HospitalID,
		}
		database.DB.Create(&newUserHospital)
	}

	return c.JSON(fiber.Map{
		"message": "เลือกโรงพยาบาลสำเร็จ",
		"data":    hospital,
	})
}

// ✅ 3. ดูโรงพยาบาลของฉัน (My Hospital)
func GetMyHospital(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	var userHospital models.UserHospital

	// Preload Hospital เพื่อดึงชื่อโรงพยาบาลมาด้วย
	if err := database.DB.Preload("Hospital").Where("user_id = ?", userID).First(&userHospital).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"message": "คุณยังไม่ได้สังกัดโรงพยาบาลใด"})
	}

	return c.JSON(fiber.Map{"data": userHospital.Hospital})
}
