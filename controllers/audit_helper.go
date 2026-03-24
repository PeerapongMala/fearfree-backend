package controllers

import (
	"fearfree-backend/database"
	"fearfree-backend/models"
	"log"
	"strconv"
	"strings"
	"unicode"

	"github.com/gofiber/fiber/v2"
)

// logAudit saves an audit log entry asynchronously so it doesn't slow down the request.
func logAudit(c *fiber.Ctx, userID uint, action string, detail string) {
	ip := c.IP()
	go func() {
		entry := models.AuditLog{
			UserID:    userID,
			Action:    action,
			Detail:    detail,
			IPAddress: ip,
		}
		if err := database.DB.Create(&entry).Error; err != nil {
			log.Printf("audit log error: %v", err)
		}
	}()
}

// validatePasswordComplexity checks that the password contains at least 1 digit,
// 1 uppercase letter, and 1 special character.
func validatePasswordComplexity(password string) bool {
	var hasDigit, hasUpper, hasSpecial bool
	specialChars := "!@#$%^&*()_+-=[]{}|;:,.<>?"
	for _, ch := range password {
		switch {
		case unicode.IsDigit(ch):
			hasDigit = true
		case unicode.IsUpper(ch):
			hasUpper = true
		case strings.ContainsRune(specialChars, ch):
			hasSpecial = true
		}
	}
	return hasDigit && hasUpper && hasSpecial
}

// AdminGetAuditLogs returns paginated audit logs for admin review.
func AdminGetAuditLogs(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	var total int64
	database.DB.Model(&models.AuditLog{}).Count(&total)

	var logs []models.AuditLog
	if err := database.DB.Order("created_at desc").Offset(offset).Limit(limit).Find(&logs).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "ดึงข้อมูล audit log ไม่สำเร็จ"})
	}

	return c.JSON(fiber.Map{
		"data":  logs,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}
