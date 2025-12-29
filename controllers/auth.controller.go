package controllers

import (
	"fearfree-backend/database"
	"fearfree-backend/models"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Struct สำหรับรับค่าจากหน้าบ้าน
type SignupInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
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

	// 1. Hash Password
	hash, _ := bcrypt.GenerateFromPassword([]byte(input.Password), 10)

	// 2. เตรียมข้อมูลลง DB
	user := models.User{
		FullName: input.FullName,
		Email:    input.Email,
		Roles:    []models.Role{{ID: 3}}, // ให้เป็น Patient (ID 3) โดย default
	}

	// 3. สร้าง User และ Auth พร้อมกัน
	tx := database.DB.Begin() // เริ่ม Transaction
	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "สร้าง User ไม่สำเร็จ (Email อาจซ้ำ)"})
	}

	auth := models.Auth{
		Username:     input.Username,
		PasswordHash: string(hash),
		UserID:       user.ID,
	}
	if err := tx.Create(&auth).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "สร้างบัญชีไม่สำเร็จ (Username อาจซ้ำ)"})
	}
	tx.Commit() // บันทึกจริง

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
	var auth models.Auth
	if err := database.DB.Where("username = ?", input.Username).First(&auth).Error; err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "ชื่อผู้ใช้หรือรหัสผ่านผิด"})
	}

	// 2. เช็ครหัสผ่าน
	if err := bcrypt.CompareHashAndPassword([]byte(auth.PasswordHash), []byte(input.Password)); err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "ชื่อผู้ใช้หรือรหัสผ่านผิด"})
	}

	// 3. ดึงข้อมูล User เพิ่มเติม (เพื่อเอา Role)
	var user models.User
	database.DB.Preload("Roles").First(&user, auth.UserID)

	// 4. สร้าง JWT Token
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = user.ID
	claims["role"] = "patient" // default
	if len(user.Roles) > 0 {
		claims["role"] = user.Roles[0].RoleName
	}
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix() // Token อยู่ได้ 3 วัน

	t, _ := token.SignedString([]byte(os.Getenv("JWT_SECRET")))

	return c.JSON(fiber.Map{
		"message": "เข้าสู่ระบบสำเร็จ",
		"token":   t,
		"user": fiber.Map{
			"id":        user.ID,
			"full_name": user.FullName,
			"role":      claims["role"],
		},
	})
}
