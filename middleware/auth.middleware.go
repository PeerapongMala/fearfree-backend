package middleware

import (
	"os"
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
	tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

	// 2. ตรวจสอบความถูกต้องของ Token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// เช็คว่าวิธีเข้ารหัสตรงกันไหม (ป้องกันการปลอมแปลง)
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fiber.ErrUnauthorized
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil || !token.Valid {
		return c.Status(401).JSON(fiber.Map{"error": "Token หมดอายุหรือข้อมูไม่ถูกต้อง"})
	}

	// 3. แกะข้อมูลใน Token (Claims) ออกมาใช้งาน
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		// เก็บ UserID และ Role ไว้ใน Context (เพื่อให้ Controller ตัวถัดไปเอาไปใช้ได้)
		// หมายเหตุ: JWT เก็บตัวเลขเป็น float64 ต้องแปลงเป็น uint
		userID := uint(claims["user_id"].(float64))
		c.Locals("user_id", userID)
		c.Locals("role", claims["role"])
	}

	// 4. ผ่านด่านได้! ไปทำฟังก์ชันถัดไป
	return c.Next()
}
