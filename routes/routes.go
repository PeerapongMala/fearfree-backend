package routes

import (
	"fearfree-backend/controllers"
	"fearfree-backend/middleware"

	"github.com/gofiber/fiber/v2"
)

func Setup(app *fiber.App) {
	// ... (Auth Routes เดิม) ...
	auth := app.Group("/auth")

	auth.Post("/signup", controllers.Signup)
	auth.Post("/login", controllers.Login)

	v1 := app.Group("/auth/v1", middleware.Protect)

	v1.Get("/me", controllers.GetProfile)
	v1.Patch("/me", controllers.UpdateProfile)

	stages := app.Group("/stages/v1", middleware.Protect)

	stages.Get("/categories", controllers.ListAnimalCategories)           // ดึงหมวดหมู่
	stages.Get("/:categoryId/animals", controllers.ListAnimalsByCategory) // ดึงสัตว์
	stages.Get("/:animalId/stage", controllers.ListStagesByAnimal)        // ดึงด่าน
	stages.Post("/:stageId/result", controllers.SubmitStageResult)

	rewards := app.Group("/rewards/v1", middleware.Protect)

	rewards.Get("/", controllers.ListRewards)                   // ดูของรางวัล
	rewards.Post("/:rewardId/redeem", controllers.RedeemReward) // แลกรางวัล
	rewards.Get("/me", controllers.GetMyRedemptions)            // ดูประวัติ

	assessment := app.Group("/assessment/v1", middleware.Protect)

	assessment.Get("/", controllers.GetAssessments)
	assessment.Post("/submit", controllers.SubmitAssessment)

	hospitals := app.Group("/hospitals/v1", middleware.Protect)

	hospitals.Get("/", controllers.ListHospitals)         // ดูรายชื่อทั้งหมด
	hospitals.Post("/select", controllers.SelectHospital) // เลือกสังกัด
	hospitals.Get("/me", controllers.GetMyHospital)       // ดูสังกัดของฉัน
}
