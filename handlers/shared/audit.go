package shared

import (
	"fearfree-backend/database"
	"fearfree-backend/models"
	"log"

	"github.com/gofiber/fiber/v2"
)

// LogAudit saves an audit log entry asynchronously so it doesn't slow down the request.
func LogAudit(c *fiber.Ctx, userID uint, action string, detail string) {
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
