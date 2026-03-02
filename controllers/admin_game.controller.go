package controllers

import (
	"fearfree-backend/database"
	"fearfree-backend/models"

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
	return c.Status(201).JSON(fiber.Map{"success": true, "data": cat})
}

func AdminUpdateCategory(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")
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
	database.DB.Save(&cat)
	return c.JSON(fiber.Map{"success": true, "data": cat})
}

func AdminDeleteCategory(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")
	database.DB.Delete(&models.AnimalCategory{}, id)
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

	animal := models.Animal{
		CategoryID:   input.CategoryID,
		Name:         input.Name,
		Description:  input.Description,
		ThumbnailUrl: input.ThumbnailUrl,
	}

	if err := database.DB.Create(&animal).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "สร้างสัตว์ไม่สำเร็จ"})
	}
	return c.Status(201).JSON(fiber.Map{"success": true, "data": animal})
}

func AdminUpdateAnimal(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")
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
	database.DB.Save(&animal)
	return c.JSON(fiber.Map{"success": true, "data": animal})
}

func AdminDeleteAnimal(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")
	database.DB.Delete(&models.Animal{}, id)
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
	return c.Status(201).JSON(fiber.Map{"success": true, "data": stage})
}

func AdminUpdateStage(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")
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
	database.DB.Save(&stage)
	return c.JSON(fiber.Map{"success": true, "data": stage})
}

func AdminDeleteStage(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")
	database.DB.Delete(&models.Stage{}, id)
	return c.JSON(fiber.Map{"success": true})
}
