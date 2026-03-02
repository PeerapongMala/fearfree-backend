package controllers

import (
	"fearfree-backend/database"
	"fearfree-backend/models"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

// 1. GET /doctor/patients
func GetPatients(c *fiber.Ctx) error {
	var patients []models.Patient
	if err := database.DB.Order("created_at desc").Find(&patients).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "ไม่สามารถดึงข้อมูลผู้ป่วยได้"})
	}

	var result []fiber.Map
	for _, p := range patients {
		code := ""
		if p.CodePatient != nil {
			code = *p.CodePatient
		}
		result = append(result, fiber.Map{
			"id":               p.ID,
			"full_name":        p.FullName,
			"fear_level":       p.FearLevel,
			"most_fear_animal": p.MostFearAnimal,
			"code_patient":     code,
			"created_at":       p.CreatedAt.Format("02 Jan 2006"),
		})
	}

	return c.JSON(fiber.Map{"data": result})
}

func generateCodePatient() string {
	var count int64
	database.DB.Model(&models.Patient{}).Count(&count)
	return fmt.Sprintf("CHBCD%04d", count+1)
}

type CreatePatientInput struct {
	FullName       string `json:"full_name"`
	MostFearAnimal string `json:"most_fear_animal"`
}

// 2. POST /doctor/patients
func CreatePatientDoctor(c *fiber.Ctx) error {
	var input CreatePatientInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ข้อมูลไม่ถูกต้อง"})
	}

	tx := database.DB.Begin()

	codePatient := generateCodePatient()
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte(codePatient), 10)

	user := models.User{
		Username:     codePatient,
		Email:        fmt.Sprintf("%s@fearfree.local", codePatient),
		PasswordHash: string(passwordHash),
		Role:         models.RolePatient,
	}

	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "สร้าง User ไม่สำเร็จ"})
	}

	patient := models.Patient{
		UserID:         user.ID,
		FullName:       input.FullName,
		MostFearAnimal: input.MostFearAnimal,
		FearLevel:      "ปานกลาง", // Default
		CodePatient:    &codePatient,
	}

	if err := tx.Create(&patient).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "สร้าง Patient ไม่สำเร็จ"})
	}

	tx.Commit()

	return c.Status(201).JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"id":               patient.ID,
			"full_name":        patient.FullName,
			"fear_level":       patient.FearLevel,
			"most_fear_animal": patient.MostFearAnimal,
			"code_patient":     codePatient,
			"created_at":       patient.CreatedAt.Format("02 Jan 2006"),
		},
	})
}

// 3. DELETE /doctor/patients/:id
func DeletePatient(c *fiber.Ctx) error {
	patientID, _ := c.ParamsInt("id")

	var patient models.Patient
	if err := database.DB.First(&patient, patientID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "ไม่พบผู้ป่วย"})
	}

	// Delete associated user
	database.DB.Delete(&models.User{}, patient.UserID)
	database.DB.Delete(&patient)

	return c.JSON(fiber.Map{"success": true})
}

// 4. GET /doctor/patients/:id/history (Play History % per animal)
func GetPatientPlayHistoryAggregated(c *fiber.Ctx) error {
	patientID, _ := c.ParamsInt("id")

	var patient models.Patient
	if err := database.DB.First(&patient, patientID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "ไม่พบผู้ป่วย"})
	}

	// 1. Fetch all animals and their total stages
	var animals []models.Animal
	if err := database.DB.Preload("Stages").Find(&animals).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "ไม่สามารถดึงข้อมูลสัตว์ได้"})
	}

	// 2. Fetch completed stages for this patient
	var progress []models.PatientProgress
	database.DB.Where("patient_id = ? AND status = ?", patient.ID, models.StatusCompleted).Preload("Stage").Find(&progress)

	// Map completed stages by animal
	completedByAnimal := make(map[uint]int)
	for _, p := range progress {
		if p.Stage.ID != 0 {
			completedByAnimal[p.Stage.AnimalID]++
		}
	}

	// Calculate %
	var historyList []fiber.Map
	for _, a := range animals {
		total := len(a.Stages)
		completed := completedByAnimal[a.ID]
		percent := 0
		if total > 0 {
			percent = (completed * 100) / total
		}

		historyList = append(historyList, fiber.Map{
			"animal_name":      a.Name,
			"progress_percent": percent,
		})
	}

	code := ""
	if patient.CodePatient != nil {
		code = *patient.CodePatient
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"patient": fiber.Map{
				"id":           patient.ID,
				"full_name":    patient.FullName,
				"code_patient": code,
			},
			"history": historyList,
		},
	})
}

// 5. GET /doctor/patients/:id/test-history (Symptom notes per stage)
func GetPatientTestHistoryNotes(c *fiber.Ctx) error {
	patientID, _ := c.ParamsInt("id")

	var patient models.Patient
	if err := database.DB.First(&patient, patientID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "ไม่พบผู้ป่วย"})
	}

	var progress []models.PatientProgress
	if err := database.DB.Preload("Stage").Preload("Stage.Animal").Where("patient_id = ?", patientID).Where("symptom_note != ?", "").Order("completed_at desc").Find(&progress).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "ดึงข้อมูลประวัติไม่สำเร็จ"})
	}

	var testHistory []fiber.Map
	for _, p := range progress {
		animalName := ""
		stageNo := 0
		if p.Stage.ID != 0 {
			stageNo = p.Stage.StageNo
			if p.Stage.Animal.ID != 0 {
				animalName = p.Stage.Animal.Name
			}
		}

		// Ensure time is not nil before formatting
		createdAtStr := ""
		if p.CompletedAt != nil {
			createdAtStr = p.CompletedAt.Format("02 Jan 2006")
		} else if p.UnlockDate != nil {
			createdAtStr = p.UnlockDate.Format("02 Jan 2006")
		}

		testHistory = append(testHistory, fiber.Map{
			"id":           p.ID,
			"animal_name":  animalName,
			"stage_no":     stageNo,
			"symptom_note": p.SymptomNote,
			"created_at":   createdAtStr,
		})
	}

	code := ""
	if patient.CodePatient != nil {
		code = *patient.CodePatient
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"patient": fiber.Map{
				"id":           patient.ID,
				"full_name":    patient.FullName,
				"code_patient": code,
			},
			"test_history": testHistory,
		},
	})
}

// 6. GET /doctor/patients/:id/redemptions (Reward Redemptions)
func GetPatientRedemptionsDoc(c *fiber.Ctx) error {
	patientID, _ := c.ParamsInt("id")

	var patient models.Patient
	if err := database.DB.First(&patient, patientID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "ไม่พบผู้ป่วย"})
	}

	var histories []models.RedemptionHistory
	if err := database.DB.Preload("Reward").Where("patient_id = ?", patient.ID).Order("redeemed_at desc").Find(&histories).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "ดึงข้อมูลประวัติการแลกรางวัลไม่สำเร็จ"})
	}

	var redemptions []fiber.Map
	for _, r := range histories {
		redemptions = append(redemptions, fiber.Map{
			"id":          r.ID,
			"date":        r.RedeemedAt.Format("02 Jan 2006"),
			"reward_name": r.Reward.Name,
			"coins_used":  r.Reward.CostCoins,
			"status":      "success",
		})
	}

	code := ""
	if patient.CodePatient != nil {
		code = *patient.CodePatient
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"patient": fiber.Map{
				"id":           patient.ID,
				"full_name":    patient.FullName,
				"code_patient": code,
			},
			"redemptions": redemptions,
		},
	})
}
