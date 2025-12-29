package main

import (
	"fearfree-backend/database"
	"fearfree-backend/models"
	"fearfree-backend/routes" // ✅ เพิ่มบรรทัดนี้
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Note: .env file not found")
	}

	database.ConnectDB()

	database.DB.AutoMigrate(
		&models.User{},
		&models.Auth{},
		&models.Role{},
		&models.AnimalCategory{},
		&models.Animal{},
		&models.MediaStore{},
		&models.Stage{},
		&models.StageResult{},
		&models.GameRules{},
		&models.Reward{},
		&models.RewardsUser{},
		&models.AssessmentStore{},
		&models.AssessmentResult{},
		&models.Hospital{},
		&models.UserHospital{},
	)

	app := fiber.New()

	// ✅ ตั้งค่า CORS (เพื่อให้ Frontend Next.js เรียกมาได้)
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:3000",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	// ✅ เรียกใช้ Routes
	routes.Setup(app)

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "FearFree Backend is Running! 🚀"})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(app.Listen(":" + port))

}
