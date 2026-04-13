package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fearfree-backend/config"
	"fearfree-backend/database"
	"fearfree-backend/handlers/shared"
	"fearfree-backend/models"
	"fmt"
	"math"
	"regexp"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
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

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// hashToken returns a hex-encoded SHA-256 hash of the token string.
func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// generateTokenID creates a random 32-byte hex string used as a unique jti claim.
func generateTokenID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// storeRefreshToken persists a hashed refresh token for server-side validation.
func storeRefreshToken(userID uint, tokenString string, expiresAt time.Time) error {
	record := models.RefreshToken{
		UserID:    userID,
		TokenHash: hashToken(tokenString),
		ExpiresAt: expiresAt,
	}
	return database.DB.Create(&record).Error
}

// Signup (สมัครสมาชิก)
func Signup(c *fiber.Ctx) error {
	var input SignupInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ส่งข้อมูลมาผิดรูปแบบ"})
	}

	// 0. Validate Username & Password
	if len(input.Username) < 6 {
		return c.Status(400).JSON(fiber.Map{"error": "ชื่อผู้ใช้ต้องมีความยาวอย่างน้อย 6 ตัวอักษร"})
	}
	if len(input.Username) > 50 {
		return c.Status(400).JSON(fiber.Map{"error": "ชื่อผู้ใช้ต้องมีความยาวไม่เกิน 50 ตัวอักษร"})
	}
	if len(input.Email) > 254 {
		return c.Status(400).JSON(fiber.Map{"error": "อีเมลต้องมีความยาวไม่เกิน 254 ตัวอักษร"})
	}
	if len(input.Password) < 8 {
		return c.Status(400).JSON(fiber.Map{"error": "รหัสผ่านต้องมีความยาวอย่างน้อย 8 ตัวอักษร"})
	}
	if len(input.Password) > 72 {
		return c.Status(400).JSON(fiber.Map{"error": "รหัสผ่านต้องมีความยาวไม่เกิน 72 ตัวอักษร"})
	}
	if !shared.ValidatePasswordComplexity(input.Password) {
		return c.Status(400).JSON(fiber.Map{"error": "รหัสผ่านต้องมีตัวเลข ตัวอักษรใหญ่ และอักขระพิเศษอย่างน้อยอย่างละ 1 ตัว"})
	}

	// Validate age if provided
	if input.Age != 0 && (input.Age < 1 || input.Age > 120) {
		return c.Status(400).JSON(fiber.Map{"error": "อายุต้องอยู่ระหว่าง 1-120 ปี"})
	}

	// Validate email format
	if !emailRegex.MatchString(input.Email) {
		return c.Status(400).JSON(fiber.Map{"error": "รูปแบบอีเมลไม่ถูกต้อง"})
	}

	// Validate FearLevel enum
	if input.FearLevel != "" && input.FearLevel != "low" && input.FearLevel != "medium" && input.FearLevel != "high" {
		return c.Status(400).JSON(fiber.Map{"error": "fear_level ต้องเป็น low, medium หรือ high"})
	}

	// 1. Hash Password
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), 10)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "เข้ารหัสรหัสผ่านไม่สำเร็จ"})
	}

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

	if err := tx.Commit().Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "บันทึกข้อมูลไม่สำเร็จ"})
	}

	shared.LogAudit(c, user.ID, "signup", "User signed up: "+input.Username)

	return c.Status(201).JSON(fiber.Map{
		"message": "สมัครสมาชิกสำเร็จ!",
		"user_id": user.ID,
	})
}

// Login (เข้าสู่ระบบ)
func Login(c *fiber.Ctx) error {
	var input LoginInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ข้อมูลไม่ถูกต้อง"})
	}

	// ตรวจสอบ Account Lockout
	var attempt models.LoginAttempt
	database.DB.Where("username = ?", input.Username).FirstOrCreate(&attempt, models.LoginAttempt{Username: input.Username})

	if attempt.LockedUntil != nil && attempt.LockedUntil.After(time.Now()) {
		remaining := math.Ceil(time.Until(*attempt.LockedUntil).Minutes())
		return c.Status(429).JSON(fiber.Map{
			"error": fmt.Sprintf("บัญชีถูกล็อค กรุณารอ %.0f นาที", remaining),
		})
	}

	// 1. หา User จาก Username
	var user models.User
	if err := database.DB.Where("username = ?", input.Username).First(&user).Error; err != nil {
		// HIGH-2: Atomic increment to avoid read-modify-write race
		database.DB.Model(&models.LoginAttempt{}).Where("username = ?", input.Username).UpdateColumn("attempts", gorm.Expr("attempts + 1"))
		// Re-read to check lockout threshold
		database.DB.Where("username = ?", input.Username).First(&attempt)
		if attempt.Attempts >= 5 {
			lockUntil := time.Now().Add(15 * time.Minute)
			database.DB.Model(&attempt).Update("locked_until", &lockUntil)
		}
		return c.Status(401).JSON(fiber.Map{"error": "ชื่อผู้ใช้หรือรหัสผ่านผิด"})
	}

	// 2. เช็ครหัสผ่านด้วย Bcrypt
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		// HIGH-2: Atomic increment to avoid read-modify-write race
		database.DB.Model(&models.LoginAttempt{}).Where("username = ?", input.Username).UpdateColumn("attempts", gorm.Expr("attempts + 1"))
		// Re-read to check lockout threshold
		database.DB.Where("username = ?", input.Username).First(&attempt)
		if attempt.Attempts >= 5 {
			lockUntil := time.Now().Add(15 * time.Minute)
			database.DB.Model(&attempt).Update("locked_until", &lockUntil)
		}
		return c.Status(401).JSON(fiber.Map{"error": "ชื่อผู้ใช้หรือรหัสผ่านผิด"})
	}

	// Login สำเร็จ → reset attempts
	attempt.Attempts = 0
	attempt.LockedUntil = nil
	database.DB.Save(&attempt)

	shared.LogAudit(c, user.ID, "login", "User logged in: "+user.Username)

	// 3. สร้าง Access Token (1 ชั่วโมง)
	accessToken := jwt.New(jwt.SigningMethodHS256)
	accessClaims := accessToken.Claims.(jwt.MapClaims)
	accessClaims["user_id"] = user.ID
	accessClaims["role"] = user.Role
	accessClaims["exp"] = time.Now().Add(time.Hour * 1).Unix()

	accessTokenString, err := accessToken.SignedString([]byte(config.Env.JWTSecret))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "สร้าง Access Token ไม่สำเร็จ"})
	}

	// 4. สร้าง Refresh Token (7 วัน)
	refreshExpiry := time.Now().Add(time.Hour * 24 * 7)
	jti, err := generateTokenID()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "สร้าง Refresh Token ไม่สำเร็จ"})
	}

	refreshToken := jwt.New(jwt.SigningMethodHS256)
	refreshClaims := refreshToken.Claims.(jwt.MapClaims)
	refreshClaims["user_id"] = user.ID
	refreshClaims["role"] = user.Role
	refreshClaims["type"] = "refresh"
	refreshClaims["jti"] = jti
	refreshClaims["exp"] = refreshExpiry.Unix()

	refreshTokenString, err := refreshToken.SignedString([]byte(config.Env.JWTSecret))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "สร้าง Refresh Token ไม่สำเร็จ"})
	}

	// Store refresh token server-side
	if err := storeRefreshToken(user.ID, refreshTokenString, refreshExpiry); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "บันทึก Refresh Token ไม่สำเร็จ"})
	}

	return c.JSON(fiber.Map{
		"message":       "เข้าสู่ระบบสำเร็จ",
		"token":         accessTokenString,
		"refresh_token": refreshTokenString,
		"user": fiber.Map{
			"id":       user.ID,
			"username": user.Username,
			"role":     user.Role,
		},
	})
}

// RefreshToken - ใช้ refresh token เพื่อออก access token ใหม่
func RefreshToken(c *fiber.Ctx) error {
	type RefreshInput struct {
		RefreshToken string `json:"refresh_token"`
	}

	var input RefreshInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ข้อมูลไม่ถูกต้อง"})
	}

	if input.RefreshToken == "" {
		return c.Status(400).JSON(fiber.Map{"error": "กรุณาส่ง refresh_token"})
	}

	// ตรวจสอบ refresh token
	token, err := jwt.Parse(input.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fiber.ErrUnauthorized
		}
		return []byte(config.Env.JWTSecret), nil
	})

	if err != nil || !token.Valid {
		return c.Status(401).JSON(fiber.Map{"error": "Refresh Token หมดอายุหรือไม่ถูกต้อง"})
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Refresh Token ไม่ถูกต้อง"})
	}

	// ตรวจสอบว่าเป็น refresh token จริง
	tokenType, _ := claims["type"].(string)
	if tokenType != "refresh" {
		return c.Status(401).JSON(fiber.Map{"error": "Token ประเภทไม่ถูกต้อง"})
	}

	userID, ok := claims["user_id"].(float64)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Token ไม่มีข้อมูล user_id"})
	}
	role, ok := claims["role"].(string)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Token ไม่มีข้อมูล role"})
	}

	// Validate refresh token exists server-side and revoke it
	oldHash := hashToken(input.RefreshToken)
	result := database.DB.Where("token_hash = ? AND user_id = ?", oldHash, uint(userID)).Delete(&models.RefreshToken{})
	if result.RowsAffected == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Refresh Token ถูกเพิกถอนแล้วหรือไม่ถูกต้อง"})
	}

	// สร้าง Access Token ใหม่ (1 ชั่วโมง)
	newAccessToken := jwt.New(jwt.SigningMethodHS256)
	newAccessClaims := newAccessToken.Claims.(jwt.MapClaims)
	newAccessClaims["user_id"] = userID
	newAccessClaims["role"] = role
	newAccessClaims["exp"] = time.Now().Add(time.Hour * 1).Unix()

	newAccessTokenString, err := newAccessToken.SignedString([]byte(config.Env.JWTSecret))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "สร้าง Access Token ไม่สำเร็จ"})
	}

	// สร้าง Refresh Token ใหม่ (7 วัน)
	newRefreshExpiry := time.Now().Add(time.Hour * 24 * 7)
	newJti, err := generateTokenID()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "สร้าง Refresh Token ไม่สำเร็จ"})
	}

	newRefreshToken := jwt.New(jwt.SigningMethodHS256)
	newRefreshClaims := newRefreshToken.Claims.(jwt.MapClaims)
	newRefreshClaims["user_id"] = userID
	newRefreshClaims["role"] = role
	newRefreshClaims["type"] = "refresh"
	newRefreshClaims["jti"] = newJti
	newRefreshClaims["exp"] = newRefreshExpiry.Unix()

	newRefreshTokenString, err := newRefreshToken.SignedString([]byte(config.Env.JWTSecret))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "สร้าง Refresh Token ไม่สำเร็จ"})
	}

	// Store new refresh token server-side
	if err := storeRefreshToken(uint(userID), newRefreshTokenString, newRefreshExpiry); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "บันทึก Refresh Token ไม่สำเร็จ"})
	}

	return c.JSON(fiber.Map{
		"token":         newAccessTokenString,
		"refresh_token": newRefreshTokenString,
	})
}

// Logout - revoke all refresh tokens for the authenticated user
func Logout(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	if err := database.DB.Where("user_id = ?", userID).Delete(&models.RefreshToken{}).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "ออกจากระบบไม่สำเร็จ"})
	}

	shared.LogAudit(c, userID, "logout", "User logged out")

	return c.JSON(fiber.Map{"message": "ออกจากระบบสำเร็จ"})
}
