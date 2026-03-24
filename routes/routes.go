package routes

import (
	"fearfree-backend/controllers"
	"fearfree-backend/middleware"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

func Setup(app *fiber.App) {
	api := app.Group("/api/v1")

	auth := api.Group("/auth", limiter.New(limiter.Config{
		Max:        10,
		Expiration: 1 * time.Minute,
	}))
	auth.Post("/signup", controllers.Signup)
	auth.Post("/login", controllers.Login)
	auth.Post("/refresh", controllers.RefreshToken)

	// Users endpoints (Moved from Auth)
	users := api.Group("/users", middleware.Protect)
	users.Get("/me", controllers.GetProfile)
	users.Patch("/me", controllers.UpdateProfile)
	users.Get("/me/redemptions", controllers.GetMyRedemptions) // Moved from rewards
	users.Get("/play-history", controllers.GetMyPlayHistory)

	stages := api.Group("/stages", middleware.Protect)
	stages.Get("/categories", controllers.ListAnimalCategories)                      // ดึงหมวดหมู่
	stages.Get("/categories/:categoryId/animals", controllers.ListAnimalsByCategory) // ดึงสัตว์
	stages.Get("/animals/:animalId/levels", controllers.ListStagesByAnimal)          // ดึงด่าน (ใช้คำว่า levels แทน stages)
	stages.Post("/levels/:levelId/results", controllers.SubmitStageResult)           // บันทึกผลการเล่น

	rewards := api.Group("/rewards", middleware.Protect)
	rewards.Get("/", controllers.ListRewards)                   // ดูของรางวัล
	rewards.Post("/:rewardId/redeem", controllers.RedeemReward) // แลกรางวัล

	assessments := api.Group("/assessments", middleware.Protect)
	assessments.Get("/questions", controllers.GetAssessments)
	assessments.Post("/results", controllers.SubmitAssessment)

	hospitals := api.Group("/hospitals", middleware.Protect)
	hospitals.Get("/", controllers.ListHospitals)         // ดูรายชื่อทั้งหมด
	hospitals.Post("/select", controllers.SelectHospital) // เลือกสังกัด
	hospitals.Get("/me", controllers.GetMyHospital)       // ดูสังกัดของฉัน

	doctor := api.Group("/doctor", middleware.Protect, middleware.IsDoctor)
	doctor.Get("/patients", controllers.GetPatients)
	doctor.Post("/patients", controllers.CreatePatientDoctor)
	doctor.Delete("/patients/:id", controllers.DeletePatient)
	doctor.Get("/patients/:id/history", controllers.GetPatientPlayHistoryAggregated)
	doctor.Get("/patients/:id/test-history", controllers.GetPatientTestHistoryNotes)
	doctor.Get("/patients/:id/redemptions", controllers.GetPatientRedemptionsDoc)

	admin := api.Group("/admin", middleware.Protect, middleware.IsAdmin)
	// Rewards
	admin.Get("/rewards", controllers.AdminGetRewards)
	admin.Post("/rewards", controllers.AdminCreateReward)
	admin.Put("/rewards/:id", controllers.AdminUpdateReward)
	admin.Delete("/rewards/:id", controllers.AdminDeleteReward)

	// Categories
	admin.Post("/categories", controllers.AdminCreateCategory)
	admin.Put("/categories/:id", controllers.AdminUpdateCategory)
	admin.Delete("/categories/:id", controllers.AdminDeleteCategory)

	// Animals
	admin.Post("/animals", controllers.AdminCreateAnimal)
	admin.Put("/animals/:id", controllers.AdminUpdateAnimal)
	admin.Delete("/animals/:id", controllers.AdminDeleteAnimal)

	// Stages
	admin.Post("/stages", controllers.AdminCreateStage)
	admin.Put("/stages/:id", controllers.AdminUpdateStage)
	admin.Delete("/stages/:id", controllers.AdminDeleteStage)

	// Audit Logs
	admin.Get("/audit-logs", controllers.AdminGetAuditLogs)
}
