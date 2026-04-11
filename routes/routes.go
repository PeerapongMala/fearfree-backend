package routes

import (
	"fearfree-backend/handlers/admin"
	"fearfree-backend/handlers/assessment"
	"fearfree-backend/handlers/auth"
	"fearfree-backend/handlers/doctor"
	"fearfree-backend/handlers/game"
	"fearfree-backend/handlers/hospital"
	"fearfree-backend/handlers/reward"
	"fearfree-backend/handlers/user"
	"fearfree-backend/middleware"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

func Setup(app *fiber.App) {
	api := app.Group("/api/v1")

	authGroup := api.Group("/auth", limiter.New(limiter.Config{
		Max:        10,
		Expiration: 1 * time.Minute,
	}))
	authGroup.Post("/signup", auth.Signup)
	authGroup.Post("/login", auth.Login)
	authGroup.Post("/refresh", auth.RefreshToken)
	authGroup.Post("/logout", middleware.Protect, auth.Logout)

	// Users endpoints (Moved from Auth)
	users := api.Group("/users", middleware.Protect)
	users.Get("/me", user.GetProfile)
	users.Patch("/me", user.UpdateProfile)
	users.Get("/me/redemptions", user.GetMyRedemptions) // Moved from rewards
	users.Get("/play-history", user.GetMyPlayHistory)

	stages := api.Group("/stages", middleware.Protect)
	stages.Get("/categories", game.ListAnimalCategories)                      // ดึงหมวดหมู่
	stages.Get("/categories/:categoryId/animals", game.ListAnimalsByCategory) // ดึงสัตว์
	stages.Get("/animals/:animalId/levels", game.ListStagesByAnimal)          // ดึงด่าน (ใช้คำว่า levels แทน stages)
	stages.Post("/levels/:levelId/results", game.SubmitStageResult)           // บันทึกผลการเล่น

	rewards := api.Group("/rewards", middleware.Protect)
	rewards.Get("/", reward.ListRewards)                   // ดูของรางวัล
	rewards.Post("/:rewardId/redeem", reward.RedeemReward) // แลกรางวัล

	assessments := api.Group("/assessments", middleware.Protect)
	assessments.Get("/questions", assessment.GetAssessments)
	assessments.Post("/results", assessment.SubmitAssessment)

	hospitals := api.Group("/hospitals", middleware.Protect)
	hospitals.Get("/", hospital.ListHospitals)         // ดูรายชื่อทั้งหมด
	hospitals.Post("/select", hospital.SelectHospital) // เลือกสังกัด
	hospitals.Get("/me", hospital.GetMyHospital)       // ดูสังกัดของฉัน

	doctorGroup := api.Group("/doctor", middleware.Protect, middleware.IsDoctor)
	doctorGroup.Get("/patients", doctor.GetPatients)
	doctorGroup.Post("/patients", doctor.CreatePatientDoctor)
	doctorGroup.Delete("/patients/:id", doctor.DeletePatient)
	doctorGroup.Get("/patients/:id/history", doctor.GetPatientPlayHistoryAggregated)
	doctorGroup.Get("/patients/:id/test-history", doctor.GetPatientTestHistoryNotes)
	doctorGroup.Get("/patients/:id/redemptions", doctor.GetPatientRedemptionsDoc)

	adminGroup := api.Group("/admin", middleware.Protect, middleware.IsAdmin)
	// Rewards
	adminGroup.Get("/rewards", admin.AdminGetRewards)
	adminGroup.Post("/rewards", admin.AdminCreateReward)
	adminGroup.Put("/rewards/:id", admin.AdminUpdateReward)
	adminGroup.Delete("/rewards/:id", admin.AdminDeleteReward)

	// Categories
	adminGroup.Post("/categories", admin.AdminCreateCategory)
	adminGroup.Put("/categories/:id", admin.AdminUpdateCategory)
	adminGroup.Delete("/categories/:id", admin.AdminDeleteCategory)

	// Animals
	adminGroup.Post("/animals", admin.AdminCreateAnimal)
	adminGroup.Put("/animals/:id", admin.AdminUpdateAnimal)
	adminGroup.Delete("/animals/:id", admin.AdminDeleteAnimal)

	// Stages
	adminGroup.Post("/stages", admin.AdminCreateStage)
	adminGroup.Put("/stages/:id", admin.AdminUpdateStage)
	adminGroup.Delete("/stages/:id", admin.AdminDeleteStage)

	// Audit Logs
	adminGroup.Get("/audit-logs", admin.AdminGetAuditLogs)
}
