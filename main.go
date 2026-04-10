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
		&models.RefreshToken{},
		&models.Question{},
	); err != nil {
		log.Fatal("AutoMigrate failed: ", err)
	}

	// Seed assessment questions if table is empty
	var questionCount int64
	database.DB.Model(&models.Question{}).Count(&questionCount)
	if questionCount == 0 {
		questions := []models.Question{
			{SortOrder: 1, Prompt: "คุณรู้สึกกลัวเมื่อเห็นรูปภาพของสัตว์ที่คุณกลัวมากน้อยเพียงใด? หลังจากนี้จะได้ทำความคุ้นเคยในการใช้แอปพลิเคชัน"},
			{SortOrder: 2, Prompt: "คุณรู้สึกใจสั่นเมื่ออยู่ในที่ที่คิดว่าสัตว์ที่คุณกลัวอาจจะอยู่หรือไม่? คุณเคยลองเข้าใกล้สัตว์ที่กลัวแล้วรู้สึกอย่างไรบ้าง?"},
			{SortOrder: 3, Prompt: "คุณเคยมีอาการทางกายเมื่อเจอสัตว์หรือนึกถึงสัตว์ที่กลัว เช่น ใจสั่น เหงื่อออก หายใจลำบาก หรือไม่?"},
			{SortOrder: 4, Prompt: "คุณสามารถดูวิดีโอหรือภาพเคลื่อนไหวของสัตว์ที่คุณกลัวได้โดยไม่รู้สึกตื่นตระหนกหรือไม่?"},
			{SortOrder: 5, Prompt: "คุณหลีกเลี่ยงการไปสถานที่บางแห่งเนื่องจากกลัวว่าจะเจอสัตว์นั้นมากน้อยเพียงใด?"},
			{SortOrder: 6, Prompt: "คุณรู้สึกปลอดภัยเมื่อพูดคุยเกี่ยวกับสัตว์ที่คุณกลัวกับคนรอบข้างหรือไม่? ความรู้สึกที่ได้เป็นอย่างไรบ้าง?"},
			{SortOrder: 7, Prompt: "คุณเคยได้รับความรู้เกี่ยวกับสัตว์ที่กลัว และรู้สึกว่าความรู้เหล่านั้นช่วยลดความกลัวได้บ้างหรือไม่?"},
			{SortOrder: 8, Prompt: "คุณรู้สึกว่าความกลัวสัตว์ส่งผลกระทบต่อการใช้ชีวิตประจำวันของคุณมากน้อยเพียงใด?"},
			{SortOrder: 9, Prompt: "คุณรับมือกับสถานการณ์ที่ต้องเผชิญหน้ากับสัตว์ที่กลัว เช่น การเดินผ่านสุนัขข้างทางได้ดีเพียงใด?"},
			{SortOrder: 10, Prompt: "คุณสังเกตว่าความกลัวสัตว์ของตัวเองมีการเปลี่ยนแปลงไปจากเดิมหรือไม่?"},
		}
		database.DB.Create(&questions)
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
