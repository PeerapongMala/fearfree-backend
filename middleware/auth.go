package middleware

import (
	"fearfree-backend/config"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func Protect(c *fiber.Ctx) error {
	// 1. ดึง Token จาก Header (Authorization: Bearer <token>)
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(401).JSON(fiber.Map{"error": "กรุณาเข้าสู่ระบบ (ไม่พบ Token)"})
	}

	// ตัดคำว่า "Bearer " ออก ให้เหลือแต่ตัว Token เพียวๆ
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return c.Status(401).JSON(fiber.Map{"error": "รูปแบบ Token ไม่ถูกต้อง"})
	}
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// 2. ตรวจสอบความถูกต้องของ Token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// เช็คว่าวิธีเข้ารหัสตรงกันไหม (ป้องกันการปลอมแปลง)
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fiber.ErrUnauthorized
		}
		return []byte(config.Env.JWTSecret), nil
	})

	if err != nil || !token.Valid {
		return c.Status(401).JSON(fiber.Map{"error": "Token หมดอายุหรือข้อมูไม่ถูกต้อง"})
	}

	// 3. แกะข้อมูลใน Token (Claims) ออกมาใช้งาน
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return c.Status(401).JSON(fiber.Map{"error": "Token ไม่ถูกต้อง"})
	}

	// ใช้ comma-ok pattern เพื่อป้องกัน panic
	userIDRaw, ok := claims["user_id"]
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Token ไม่มีข้อมูล user_id"})
	}
	userIDFloat, ok := userIDRaw.(float64)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Token user_id ไม่ถูกต้อง"})
	}

	roleRaw, ok := claims["role"]
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Token ไม่มีข้อมูล role"})
	}

	c.Locals("user_id", uint(userIDFloat))
	c.Locals("role", roleRaw)

	// 4. ผ่านด่านได้! ไปทำฟังก์ชันถัดไป
	return c.Next()
}

func IsAdmin(c *fiber.Ctx) error {
	role := c.Locals("role")
	if role != "admin" {
		return c.Status(403).JSON(fiber.Map{"error": "ไม่มีสิทธิ์เข้าถึง (Admin Only)"})
	}
	return c.Next()
}

func IsDoctor(c *fiber.Ctx) error {
	role := c.Locals("role")
	if role != "doctor" {
		return c.Status(403).JSON(fiber.Map{"error": "ไม่มีสิทธิ์เข้าถึง (Doctor Only)"})
	}
	return c.Next()
}
