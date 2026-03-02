package controllers

import (
	"fearfree-backend/config"
	"fearfree-backend/database"
	"fearfree-backend/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Struct สำหรับรับค่าจากหน้าบ้าน (สมัครสมาชิก)
type SignupInput struct {
	Username       string `json:"username"`
	Password       string `json:"password"`
	Email          string `json:"email"`
	FullName       string `json:"full_name"`
	Age            int    `json:"age"`
	MostFearAnimal string `json:"most_fear_animal"`
	FearLevel      string `json:"fear_level"`
}

type LoginInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// ✅ Signup (สมัครสมาชิก)
func Signup(c *fiber.Ctx) error {
	var input SignupInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ส่งข้อมูลมาผิดรูปแบบ"})
	}

	// 0. Validate Password
	if len(input.Password) < 8 {
		return c.Status(400).JSON(fiber.Map{"error": "รหัสผ่านต้องมีความยาวอย่างน้อย 8 ตัวอักษร"})
	}

	// 1. Hash Password
	hash, _ := bcrypt.GenerateFromPassword([]byte(input.Password), 10)

	// 2. เตรียมข้อมูล User ลง DB
	user := models.User{
		Username:     input.Username,
		Email:        input.Email,
		PasswordHash: string(hash),
		Role:         models.RolePatient, // เริ่มต้นด้วยสิทธิ์ Patient
	}

	// 3. เริ่ม Transaction สำหรับสร้าง User และ Patient ต่อเนื่องกัน
	tx := database.DB.Begin()
	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "สร้าง User ไม่สำเร็จ (Username หรือ Email อาจซ้ำ)"})
	}

	if input.FearLevel == "" {
		input.FearLevel = "low" // Default Enum Value
	}

	patient := models.Patient{
		UserID:         user.ID,
		FullName:       input.FullName,
		Age:            input.Age,
		MostFearAnimal: input.MostFearAnimal,
		FearLevel:      input.FearLevel,
	}

	if err := tx.Create(&patient).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "สร้างโปรไฟล์ผู้ป่วยไม่สำเร็จ"})
	}

	tx.Commit() // บันทึกเสร็จสมบูรณ์

	return c.Status(201).JSON(fiber.Map{
		"message": "สมัครสมาชิกสำเร็จ!",
		"user_id": user.ID,
	})
}

// ✅ Login (เข้าสู่ระบบ)
func Login(c *fiber.Ctx) error {
	var input LoginInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ข้อมูลไม่ถูกต้อง"})
	}

	// 1. หา User จาก Username
	var user models.User
	if err := database.DB.Where("username = ?", input.Username).First(&user).Error; err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "ชื่อผู้ใช้หรือรหัสผ่านผิด"})
	}

	// 2. เช็ครหัสผ่านด้วย Bcrypt
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "ชื่อผู้ใช้หรือรหัสผ่านผิด"})
	}

	// 3. สร้าง JWT Token
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = user.ID
	claims["role"] = user.Role
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix() // Token อยู่ได้ 3 วัน

	t, _ := token.SignedString([]byte(config.Env.JWTSecret))

	return c.JSON(fiber.Map{
		"message": "เข้าสู่ระบบสำเร็จ",
		"token":   t,
		"user": fiber.Map{
			"id":       user.ID,
			"username": user.Username,
			"role":     user.Role,
		},
	})
}
