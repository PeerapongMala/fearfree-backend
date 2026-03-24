package main

import (
	"log"
	"os"

	"fearfree-backend/config"
	"fearfree-backend/database"
	"fearfree-backend/middleware"
	"fearfree-backend/models"
	"fearfree-backend/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	config.LoadConfig()

	database.ConnectDB()

	// สร้าง ENUM types ใน PostgreSQL ก่อน
	database.DB.Exec("DO $$ BEGIN CREATE TYPE user_role AS ENUM ('patient', 'doctor', 'admin'); EXCEPTION WHEN duplicate_object THEN null; END $$;")
	database.DB.Exec("DO $$ BEGIN CREATE TYPE fear_level AS ENUM ('low', 'medium', 'high'); EXCEPTION WHEN duplicate_object THEN null; END $$;")
	database.DB.Exec("DO $$ BEGIN CREATE TYPE media_type_enum AS ENUM ('image', 'video'); EXCEPTION WHEN duplicate_object THEN null; END $$;")
	database.DB.Exec("DO $$ BEGIN CREATE TYPE progress_status AS ENUM ('locked', 'in_progress', 'completed'); EXCEPTION WHEN duplicate_object THEN null; END $$;")

	if err := database.DB.AutoMigrate(
		&models.User{},
		&models.Patient{},
		&models.Assessment{},
		&models.AnimalCategory{},
		&models.Animal{},
		&models.Stage{},
		&models.PatientProgress{},
		&models.Reward{},
		&models.RedemptionHistory{},
		&models.Hospital{},
		&models.UserHospital{},
		&models.LoginAttempt{},
		&models.AuditLog{},
	); err != nil {
		log.Fatal("AutoMigrate failed: ", err)
	}

	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.ErrorHandler,
	})

	// ตั้งค่า CORS (เพื่อให้ Frontend Next.js เรียกมาได้)
	corsOrigin := os.Getenv("CORS_ORIGIN")
	if corsOrigin == "" {
		corsOrigin = "http://localhost:3000"
	}
	app.Use(cors.New(cors.Config{
		AllowOrigins: corsOrigin,
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	// Security Headers
	app.Use(func(c *fiber.Ctx) error {
		c.Set("X-Frame-Options", "DENY")
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-XSS-Protection", "1; mode=block")
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		return c.Next()
	})

	// เรียกใช้ Routes
	routes.Setup(app)

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "FearFree Backend is Running!"})
	})

	port := config.Env.Port
	log.Fatal(app.Listen(":" + port))

}
