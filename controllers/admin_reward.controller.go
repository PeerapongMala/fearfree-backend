package controllers

import (
	"fearfree-backend/database"
	"fearfree-backend/models"

	"github.com/gofiber/fiber/v2"
)

type RewardInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	CostCoins   int    `json:"cost_coins"`
	Stock       int    `json:"stock"`
	ImageUrl    string `json:"image_url"`
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

	return c.Status(201).JSON(fiber.Map{"success": true, "data": reward})
}

// 3. PUT /admin/rewards/:id (อัปเดต)
func AdminUpdateReward(c *fiber.Ctx) error {
	rewardID, _ := c.ParamsInt("id")
	var input RewardInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ข้อมูลไม่ถูกต้อง"})
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

	return c.JSON(fiber.Map{"success": true, "data": reward})
}

// 4. DELETE /admin/rewards/:id (ลบ)
func AdminDeleteReward(c *fiber.Ctx) error {
	rewardID, _ := c.ParamsInt("id")

	// ลบประวัติการแลกที่เชื่อมโยงก่อน (หรืออาจใช้ Cascade delete)
	database.DB.Where("reward_id = ?", rewardID).Delete(&models.RedemptionHistory{})

	if err := database.DB.Delete(&models.Reward{}, rewardID).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "ลบของรางวัลไม่สำเร็จ"})
	}

	return c.JSON(fiber.Map{"success": true})
}
