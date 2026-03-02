package main

import (
	"log"

	"fearfree-backend/config"
	"fearfree-backend/database"
	"fearfree-backend/middleware"
	"fearfree-backend/models"
	"fearfree-backend/routes" // ✅ เพิ่มบรรทัดนี้

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

	database.DB.AutoMigrate(
		&models.User{},
		&models.Patient{},
		&models.Assessment{},
		&models.AnimalCategory{},
		&models.Animal{},
		&models.Stage{},
		&models.PatientProgress{},
		&models.Reward{},
		&models.RedemptionHistory{},
	)

	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.ErrorHandler,
	})

	// ✅ ตั้งค่า CORS (เพื่อให้ Frontend Next.js เรียกมาได้)
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:3000",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	// ✅ เรียกใช้ Routes
	routes.Setup(app)

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "FearFree Backend is Running! 🚀"})
	})

	port := config.Env.Port
	log.Fatal(app.Listen(":" + port))

}
