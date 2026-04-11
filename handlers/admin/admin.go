package admin

import (
	"fearfree-backend/database"
	"fearfree-backend/handlers/shared"
	"fearfree-backend/models"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// === CATEGORIES ===

type CategoryInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func AdminCreateCategory(c *fiber.Ctx) error {
	var input CategoryInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ข้อมูลไม่ถูกต้อง"})
	}

	cat := models.AnimalCategory{
		Name:        input.Name,
		Description: input.Description,
	}

	if err := database.DB.Create(&cat).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "สร้างหมวดหมู่ไม่สำเร็จ"})
	}

	adminID := c.Locals("user_id").(uint)
	shared.LogAudit(c, adminID, "create_category", fmt.Sprintf("Created category: %s (ID: %d)", cat.Name, cat.ID))

	return c.Status(201).JSON(fiber.Map{"success": true, "data": cat})
}

func AdminUpdateCategory(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "id ไม่ถูกต้อง"})
	}
	var input CategoryInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ข้อมูลไม่ถูกต้อง"})
	}

	var cat models.AnimalCategory
	if err := database.DB.First(&cat, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "ไม่พบหมวดหมู่"})
	}

	cat.Name = input.Name
	cat.Description = input.Description
	if err := database.DB.Save(&cat).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "อัปเดตหมวดหมู่ไม่สำเร็จ"})
	}
	return c.JSON(fiber.Map{"success": true, "data": cat})
}

func AdminDeleteCategory(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "id ไม่ถูกต้อง"})
	}
	var cat models.AnimalCategory
	if err := database.DB.First(&cat, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "ไม่พบหมวดหมู่"})
	}

	tx := database.DB.Begin()

	// Find all animals in this category
	var animals []models.Animal
	if err := tx.Where("category_id = ?", cat.ID).Find(&animals).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "ดึงข้อมูลสัตว์ไม่สำเร็จ"})
	}

	for _, animal := range animals {
		// Delete all stages for this animal (patient progress is preserved as historical data)
		if err := tx.Where("animal_id = ?", animal.ID).Delete(&models.Stage{}).Error; err != nil {
			tx.Rollback()
			return c.Status(500).JSON(fiber.Map{"error": "ลบด่านไม่สำเร็จ"})
		}
	}

	// Delete all animals in this category
	if err := tx.Where("category_id = ?", cat.ID).Delete(&models.Animal{}).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "ลบสัตว์ไม่สำเร็จ"})
	}

	// Delete the category
	if err := tx.Delete(&cat).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "ลบหมวดหมู่ไม่สำเร็จ"})
	}

	if err := tx.Commit().Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "บันทึกข้อมูลไม่สำเร็จ"})
	}

	adminID := c.Locals("user_id").(uint)
	shared.LogAudit(c, adminID, "delete_category", fmt.Sprintf("Deleted category ID: %d (%s)", cat.ID, cat.Name))

	return c.JSON(fiber.Map{"success": true})
}

// === ANIMALS ===

type AnimalInput struct {
	CategoryID   uint   `json:"category_id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	ThumbnailUrl string `json:"thumbnail_url"`
}

func AdminCreateAnimal(c *fiber.Ctx) error {
	var input AnimalInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ข้อมูลไม่ถูกต้อง"})
	}

	// Validate category_id exists
	var cat models.AnimalCategory
	if err := database.DB.First(&cat, input.CategoryID).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ไม่พบหมวดหมู่ที่ระบุ (category_id ไม่ถูกต้อง)"})
	}

	animal := models.Animal{
		CategoryID:   input.CategoryID,
		Name:         input.Name,
		Description:  input.Description,
		ThumbnailUrl: input.ThumbnailUrl,
	}

	if err := database.DB.Create(&animal).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "สร้างสัตว์ไม่สำเร็จ"})
	}

	adminID := c.Locals("user_id").(uint)
	shared.LogAudit(c, adminID, "create_animal", fmt.Sprintf("Created animal: %s (ID: %d)", animal.Name, animal.ID))

	return c.Status(201).JSON(fiber.Map{"success": true, "data": animal})
}

func AdminUpdateAnimal(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "id ไม่ถูกต้อง"})
	}
	var input AnimalInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ข้อมูลไม่ถูกต้อง"})
	}

	var animal models.Animal
	if err := database.DB.First(&animal, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "ไม่พบสัตว์"})
	}

	animal.CategoryID = input.CategoryID
	animal.Name = input.Name
	animal.Description = input.Description
	animal.ThumbnailUrl = input.ThumbnailUrl
	if err := database.DB.Save(&animal).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "อัปเดตสัตว์ไม่สำเร็จ"})
	}
	return c.JSON(fiber.Map{"success": true, "data": animal})
}

func AdminDeleteAnimal(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "id ไม่ถูกต้อง"})
	}
	var animal models.Animal
	if err := database.DB.First(&animal, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "ไม่พบสัตว์"})
	}

	tx := database.DB.Begin()

	// Delete all stages for this animal (patient progress is preserved as historical data)
	if err := tx.Where("animal_id = ?", animal.ID).Delete(&models.Stage{}).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "ลบด่านไม่สำเร็จ"})
	}

	// Delete the animal
	if err := tx.Delete(&animal).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "ลบสัตว์ไม่สำเร็จ"})
	}

	if err := tx.Commit().Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "บันทึกข้อมูลไม่สำเร็จ"})
	}

	adminID := c.Locals("user_id").(uint)
	shared.LogAudit(c, adminID, "delete_animal", fmt.Sprintf("Deleted animal ID: %d (%s)", animal.ID, animal.Name))

	return c.JSON(fiber.Map{"success": true})
}

// === STAGES ===

type StageInput struct {
	AnimalID       uint             `json:"animal_id"`
	StageNo        int              `json:"stage_no"`
	MediaType      models.MediaType `json:"media_type"`
	MediaUrl       string           `json:"media_url"`
	DisplayTimeSec int              `json:"display_time_sec"`
	RewardCoins    int              `json:"reward_coins"`
}

func AdminCreateStage(c *fiber.Ctx) error {
	var input StageInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ข้อมูลไม่ถูกต้อง"})
	}

	// Validate stage fields
	if input.StageNo <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "stage_no ต้องมากกว่า 0"})
	}
	if input.RewardCoins < 0 {
		return c.Status(400).JSON(fiber.Map{"error": "reward_coins ต้องมากกว่าหรือเท่ากับ 0"})
	}
	if input.DisplayTimeSec <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "display_time_sec ต้องมากกว่า 0"})
	}
	if input.MediaType != models.MediaImage && input.MediaType != models.MediaVideo {
		return c.Status(400).JSON(fiber.Map{"error": "media_type ต้องเป็น 'image' หรือ 'video'"})
	}

	// Validate animal_id exists
	var animal models.Animal
	if err := database.DB.First(&animal, input.AnimalID).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ไม่พบสัตว์ที่ระบุ (animal_id ไม่ถูกต้อง)"})
	}

	stage := models.Stage{
		AnimalID:       input.AnimalID,
		StageNo:        input.StageNo,
		MediaType:      input.MediaType,
		MediaUrl:       input.MediaUrl,
		DisplayTimeSec: input.DisplayTimeSec,
		RewardCoins:    input.RewardCoins,
	}

	if err := database.DB.Create(&stage).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "สร้างด่านไม่สำเร็จ"})
	}

	adminID := c.Locals("user_id").(uint)
	shared.LogAudit(c, adminID, "create_stage", fmt.Sprintf("Created stage %d for animal ID: %d (ID: %d)", stage.StageNo, stage.AnimalID, stage.ID))

	return c.Status(201).JSON(fiber.Map{"success": true, "data": stage})
}

func AdminUpdateStage(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "id ไม่ถูกต้อง"})
	}
	var input StageInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ข้อมูลไม่ถูกต้อง"})
	}

	var stage models.Stage
	if err := database.DB.First(&stage, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "ไม่พบด่าน"})
	}

	stage.AnimalID = input.AnimalID
	stage.StageNo = input.StageNo
	stage.MediaType = input.MediaType
	stage.MediaUrl = input.MediaUrl
	stage.DisplayTimeSec = input.DisplayTimeSec
	stage.RewardCoins = input.RewardCoins
	if err := database.DB.Save(&stage).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "อัปเดตด่านไม่สำเร็จ"})
	}
	return c.JSON(fiber.Map{"success": true, "data": stage})
}

func AdminDeleteStage(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "id ไม่ถูกต้อง"})
	}
	var stage models.Stage
	if err := database.DB.First(&stage, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "ไม่พบด่าน"})
	}

	// Patient progress is preserved as historical data; only delete the stage
	if err := database.DB.Delete(&stage).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "ลบด่านไม่สำเร็จ"})
	}

	adminID := c.Locals("user_id").(uint)
	shared.LogAudit(c, adminID, "delete_stage", fmt.Sprintf("Deleted stage ID: %d (stage %d, animal ID: %d)", stage.ID, stage.StageNo, stage.AnimalID))

	return c.JSON(fiber.Map{"success": true})
}

// === REWARDS ===

type RewardInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	CostCoins   int    `json:"cost_coins"`
	Stock       int    `json:"stock"`
	ImageUrl    string `json:"image_url"`
}

func validateRewardInput(input RewardInput) string {
	if input.CostCoins <= 0 {
		return "cost_coins ต้องมากกว่า 0"
	}
	if input.Stock < 0 {
		return "stock ต้องมากกว่าหรือเท่ากับ 0"
	}
	return ""
}

// 1. GET /admin/rewards (ดึงทั้งหมด รวมถึงอันที่สต๊อกหมด)
func AdminGetRewards(c *fiber.Ctx) error {
	var rewards []models.Reward
	if err := database.DB.Order("id asc").Find(&rewards).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "ไม่สามารถดึงข้อมูลของรางวัลได้"})
	}
	return c.JSON(fiber.Map{"data": rewards})
}

// 2. POST /admin/rewards (สร้างใหม่)
func AdminCreateReward(c *fiber.Ctx) error {
	var input RewardInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ข้อมูลไม่ถูกต้อง"})
	}

	if msg := validateRewardInput(input); msg != "" {
		return c.Status(400).JSON(fiber.Map{"error": msg})
	}

	reward := models.Reward{
		Name:        input.Name,
		Description: input.Description,
		CostCoins:   input.CostCoins,
		Stock:       input.Stock,
		ImageUrl:    input.ImageUrl,
	}

	if err := database.DB.Create(&reward).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "สร้างของรางวัลไม่สำเร็จ"})
	}

	adminID := c.Locals("user_id").(uint)
	shared.LogAudit(c, adminID, "create_reward", fmt.Sprintf("Created reward: %s (ID: %d)", reward.Name, reward.ID))

	return c.Status(201).JSON(fiber.Map{"success": true, "data": reward})
}

// 3. PUT /admin/rewards/:id (อัปเดต)
func AdminUpdateReward(c *fiber.Ctx) error {
	rewardID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "id ไม่ถูกต้อง"})
	}
	var input RewardInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ข้อมูลไม่ถูกต้อง"})
	}

	if msg := validateRewardInput(input); msg != "" {
		return c.Status(400).JSON(fiber.Map{"error": msg})
	}

	var reward models.Reward
	if err := database.DB.First(&reward, rewardID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "ไม่พบของรางวัล"})
	}

	reward.Name = input.Name
	reward.Description = input.Description
	reward.CostCoins = input.CostCoins
	reward.Stock = input.Stock
	reward.ImageUrl = input.ImageUrl

	if err := database.DB.Save(&reward).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "อัปเดตของรางวัลไม่สำเร็จ"})
	}

	adminID := c.Locals("user_id").(uint)
	shared.LogAudit(c, adminID, "update_reward", fmt.Sprintf("Updated reward ID: %d (%s)", reward.ID, reward.Name))

	return c.JSON(fiber.Map{"success": true, "data": reward})
}

// 4. DELETE /admin/rewards/:id (ลบ)
func AdminDeleteReward(c *fiber.Ctx) error {
	rewardID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "id ไม่ถูกต้อง"})
	}

	var reward models.Reward
	if err := database.DB.First(&reward, rewardID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "ไม่พบของรางวัล"})
	}

	// Redemption history is preserved as historical data; only delete the reward
	if err := database.DB.Delete(&reward).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "ลบของรางวัลไม่สำเร็จ"})
	}

	adminID := c.Locals("user_id").(uint)
	shared.LogAudit(c, adminID, "delete_reward", fmt.Sprintf("Deleted reward ID: %d (%s)", reward.ID, reward.Name))

	return c.JSON(fiber.Map{"success": true})
}

// === AUDIT LOGS ===

// AdminGetAuditLogs returns paginated audit logs for admin review.
func AdminGetAuditLogs(c *fiber.Ctx) error {
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
	database.DB.Model(&models.AuditLog{}).Count(&total)

	var logs []models.AuditLog
	if err := database.DB.Order("created_at desc").Offset(offset).Limit(limit).Find(&logs).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "ดึงข้อมูล audit log ไม่สำเร็จ"})
	}

	return c.JSON(fiber.Map{
		"data":  logs,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}
