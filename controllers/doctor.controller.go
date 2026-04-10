package controllers

import (
	"crypto/rand"
	"fearfree-backend/database"
	"fearfree-backend/models"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

// 1. GET /doctor/patients
func GetPatients(c *fiber.Ctx) error {
	doctorUserID := c.Locals("user_id").(uint)

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	var total int64
	database.DB.Model(&models.Patient{}).Where("created_by_doctor_id = ?", doctorUserID).Count(&total)

	var patients []models.Patient
	if err := database.DB.Where("created_by_doctor_id = ?", doctorUserID).Order("created_at desc").Offset(offset).Limit(limit).Find(&patients).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "ไม่สามารถดึงข้อมูลผู้ป่วยได้"})
	}

	result := []fiber.Map{}
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

	return c.JSON(fiber.Map{
		"success": true,
		"data":    result,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}

const codeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // no 0/O/1/I to avoid confusion

func generateCodePatient() (string, error) {
	b := make([]byte, 8)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(codeAlphabet))))
		if err != nil {
			return "", err
		}
		b[i] = codeAlphabet[n.Int64()]
	}
	return string(b), nil
}

type CreatePatientInput struct {
	FullName       string `json:"full_name"`
	MostFearAnimal string `json:"most_fear_animal"`
}

// 2. POST /doctor/patients
func CreatePatientDoctor(c *fiber.Ctx) error {
	doctorUserID := c.Locals("user_id").(uint)

	var input CreatePatientInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ข้อมูลไม่ถูกต้อง"})
	}

	// Validate required fields
	if strings.TrimSpace(input.FullName) == "" {
		return c.Status(400).JSON(fiber.Map{"error": "กรุณาระบุชื่อผู้ป่วย (full_name)"})
	}
	if strings.TrimSpace(input.MostFearAnimal) == "" {
		return c.Status(400).JSON(fiber.Map{"error": "กรุณาระบุสัตว์ที่กลัวที่สุด (most_fear_animal)"})
	}

	const maxRetries = 5
	for attempt := 0; attempt < maxRetries; attempt++ {
		codePatient, err := generateCodePatient()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "สร้างรหัสผู้ป่วยไม่สำเร็จ"})
		}

		// Generate a separate random password (8 chars)
		password, err := generateCodePatient()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "สร้างรหัสผ่านไม่สำเร็จ"})
		}

		passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "เข้ารหัสรหัสผ่านไม่สำเร็จ"})
		}

		tx := database.DB.Begin()

		user := models.User{
			Username:     codePatient,
			Email:        fmt.Sprintf("%s@fearfree.local", codePatient),
			PasswordHash: string(passwordHash),
			Role:         models.RolePatient,
		}

		if err := tx.Create(&user).Error; err != nil {
			tx.Rollback()
			// Retry on unique constraint violation (unlikely with random codes)
			if attempt < maxRetries-1 {
				continue
			}
			return c.Status(500).JSON(fiber.Map{"error": "สร้าง User ไม่สำเร็จ"})
		}

		patient := models.Patient{
			UserID:            user.ID,
			CreatedByDoctorID: doctorUserID,
			FullName:          input.FullName,
			MostFearAnimal:    input.MostFearAnimal,
			FearLevel:         "medium", // Default
			CodePatient:       &codePatient,
		}

		if err := tx.Create(&patient).Error; err != nil {
			tx.Rollback()
			if attempt < maxRetries-1 {
				continue
			}
			return c.Status(500).JSON(fiber.Map{"error": "สร้าง Patient ไม่สำเร็จ"})
		}

		if err := tx.Commit().Error; err != nil {
			if attempt < maxRetries-1 {
				continue
			}
			return c.Status(500).JSON(fiber.Map{"error": "บันทึกข้อมูลไม่สำเร็จ"})
		}

		logAudit(c, doctorUserID, "create_patient", fmt.Sprintf("Created patient: %s (code: %s)", input.FullName, codePatient))

		return c.Status(201).JSON(fiber.Map{
			"success": true,
			"data": fiber.Map{
				"id":               patient.ID,
				"full_name":        patient.FullName,
				"fear_level":       patient.FearLevel,
				"most_fear_animal": patient.MostFearAnimal,
				"code_patient":     codePatient,
				"password":         password,
				"created_at":       patient.CreatedAt.Format("02 Jan 2006"),
			},
		})
	}

	return c.Status(500).JSON(fiber.Map{"error": "สร้าง Patient ไม่สำเร็จหลังจากลองหลายครั้ง"})
}

// helper: verify patient belongs to the requesting doctor
func getPatientOwnedByDoctor(c *fiber.Ctx) (*models.Patient, error) {
	doctorUserID := c.Locals("user_id").(uint)
	patientID, err := c.ParamsInt("id")
	if err != nil {
		return nil, c.Status(400).JSON(fiber.Map{"error": "id ไม่ถูกต้อง"})
	}

	var patient models.Patient
	if err := database.DB.First(&patient, patientID).Error; err != nil {
		return nil, c.Status(404).JSON(fiber.Map{"error": "ไม่พบผู้ป่วย"})
	}

	if patient.CreatedByDoctorID != doctorUserID {
		return nil, c.Status(403).JSON(fiber.Map{"error": "ไม่มีสิทธิ์เข้าถึงข้อมูลผู้ป่วยรายนี้"})
	}

	return &patient, nil
}

// 3. DELETE /doctor/patients/:id
func DeletePatient(c *fiber.Ctx) error {
	patient, err := getPatientOwnedByDoctor(c)
	if err != nil {
		return err
	}

	tx := database.DB.Begin()

	// Delete related data first
	if err := tx.Where("patient_id = ?", patient.ID).Delete(&models.PatientProgress{}).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "ลบข้อมูลความก้าวหน้าไม่สำเร็จ"})
	}
	if err := tx.Where("patient_id = ?", patient.ID).Delete(&models.Assessment{}).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "ลบข้อมูลการประเมินไม่สำเร็จ"})
	}
	if err := tx.Where("patient_id = ?", patient.ID).Delete(&models.RedemptionHistory{}).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "ลบข้อมูลประวัติการแลกรางวัลไม่สำเร็จ"})
	}

	// Delete UserHospital records for this user
	if err := tx.Where("user_id = ?", patient.UserID).Delete(&models.UserHospital{}).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "ลบข้อมูลโรงพยาบาลไม่สำเร็จ"})
	}

	// Delete patient and user
	if err := tx.Delete(&patient).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "ลบผู้ป่วยไม่สำเร็จ"})
	}
	if err := tx.Delete(&models.User{}, patient.UserID).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "ลบ User ไม่สำเร็จ"})
	}

	if err := tx.Commit().Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "บันทึกข้อมูลไม่สำเร็จ"})
	}

	doctorUserID := c.Locals("user_id").(uint)
	logAudit(c, doctorUserID, "delete_patient", fmt.Sprintf("Deleted patient ID: %d", patient.ID))

	return c.JSON(fiber.Map{"success": true})
}

// 4. GET /doctor/patients/:id/history (Play History % per animal)
func GetPatientPlayHistoryAggregated(c *fiber.Ctx) error {
	patient, err := getPatientOwnedByDoctor(c)
	if err != nil {
		return err
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
	historyList := []fiber.Map{}
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
	patient, err := getPatientOwnedByDoctor(c)
	if err != nil {
		return err
	}

	var progress []models.PatientProgress
	if err := database.DB.Preload("Stage").Preload("Stage.Animal").Where("patient_id = ?", patient.ID).Where("symptom_note != ?", "").Order("completed_at desc").Find(&progress).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "ดึงข้อมูลประวัติไม่สำเร็จ"})
	}

	testHistory := []fiber.Map{}
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
	patient, err := getPatientOwnedByDoctor(c)
	if err != nil {
		return err
	}

	var histories []models.RedemptionHistory
	if err := database.DB.Preload("Reward").Where("patient_id = ?", patient.ID).Order("redeemed_at desc").Find(&histories).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "ดึงข้อมูลประวัติการแลกรางวัลไม่สำเร็จ"})
	}

	redemptions := []fiber.Map{}
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
