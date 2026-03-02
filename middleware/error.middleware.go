package middleware

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

// ErrorHandler intercepts errors returning from routes and formats them into JSON
func ErrorHandler(c *fiber.Ctx, err error) error {
	// Default status code
	code := fiber.StatusInternalServerError

	// If it's a fiber.Error, use its status code
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	// Log the error internally
	log.Printf("🔥 Error: %v", err)

	// Send a structured JSON response back to the client
	return c.Status(code).JSON(fiber.Map{
		"error":   true,
		"message": err.Error(),
	})
}
