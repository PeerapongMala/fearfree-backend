package controllers

import (
	"fearfree-backend/database"
	"fearfree-backend/models"

	"github.com/gofiber/fiber/v2"
)

// GET /auth/v1/me
func GetProfile(c *fiber.Ctx) error {
	// 1. ดึง UserID ที่ยาม (Middleware) แปะไว้ให้
	userID := c.Locals("user_id").(uint)

	// 2. Query ข้อมูลจาก Database
	var user models.User
	// Preload Roles มาด้วย, Preload Auth มาด้วย (แต่ไม่เอารหัสผ่าน)
	if err := database.DB.Preload("Roles").First(&user, userID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "ไม่พบผู้ใช้งาน"})
	}

	// 3. ส่งข้อมูลกลับ
	return c.JSON(fiber.Map{
		"status": "success",
		"data":   user,
	})
}

// ... (ต่อจาก GetProfile)

type UpdateProfileInput struct {
	FullName       string `json:"full_name"`
	Age            int    `json:"age"`
	MostFearAnimal string `json:"most_fear_animal"`
}

// PATCH /auth/v1/me
func UpdateProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	var input UpdateProfileInput

	// 1. รับค่าจาก Body
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ข้อมูลไม่ถูกต้อง"})
	}

	// 2. อัปเดตข้อมูลลง DB
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "ไม่พบผู้ใช้"})
	}

	// อัปเดตเฉพาะฟิลด์ที่ส่งมา (ถ้าไม่ส่งมาจะเป็นค่าว่าง/0 ก็ข้ามไป หรือทับไปเลยตาม Logic)
	// ในที่นี้จะอัปเดตหมดตาม Input
	user.FullName = input.FullName
	user.Age = input.Age
	user.MostFearAnimal = input.MostFearAnimal

	database.DB.Save(&user)

	return c.JSON(fiber.Map{
		"message": "อัปเดตข้อมูลสำเร็จ",
		"data":    user,
	})
}
