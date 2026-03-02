package main

import (
	"fearfree-backend/config"
	"fearfree-backend/database"
	"fearfree-backend/models"
	"fmt"
)

func main() {
	// 1. Load Config & Connect DB
	config.LoadConfig()
	database.ConnectDB()

	// 2. Clear existing (optional, for clean slate)
	fmt.Println("Clearing old data...")
	// database.DB.Exec("DELETE FROM patient_progress")
	// database.DB.Exec("DELETE FROM stages")
	// database.DB.Exec("DELETE FROM animals")
	// database.DB.Exec("DELETE FROM animal_categories")
	// database.DB.Exec("DELETE FROM rewards")

	fmt.Println("Seeding Data...")

	// 3. Create Categories
	fmt.Println("Seeding Categories...")
	categories := []models.AnimalCategory{
		{Name: "Reptiles", Description: "สัตว์เลื้อยคลาน"},
		{Name: "Insects", Description: "แมลง"},
	}
	for _, c := range categories {
		database.DB.FirstOrCreate(&c, models.AnimalCategory{Name: c.Name})
	}

	// 4. Create Animals
	fmt.Println("Seeding Animals...")
	var reptileCategory, insectCategory models.AnimalCategory
	database.DB.Where("name = ?", "Reptiles").First(&reptileCategory)
	database.DB.Where("name = ?", "Insects").First(&insectCategory)

	animals := []models.Animal{
		{CategoryID: reptileCategory.ID, Name: "Snake", Description: "งู", ThumbnailUrl: "https://example.com/snake.png"},
		{CategoryID: reptileCategory.ID, Name: "Lizard", Description: "จิ้งจก", ThumbnailUrl: "https://example.com/lizard.png"},
		{CategoryID: insectCategory.ID, Name: "Spider", Description: "แมงมุม", ThumbnailUrl: "https://example.com/spider.png"},
		{CategoryID: insectCategory.ID, Name: "Cockroach", Description: "แมลงสาบ", ThumbnailUrl: "https://example.com/cockroach.png"},
	}
	for _, a := range animals {
		database.DB.FirstOrCreate(&a, models.Animal{Name: a.Name})
	}

	// 5. Create Stages
	fmt.Println("Seeding Stages...")
	var snake, spider models.Animal
	database.DB.Where("name = ?", "Snake").First(&snake)
	database.DB.Where("name = ?", "Spider").First(&spider)

	stages := []models.Stage{
		// Snake Stages
		{AnimalID: snake.ID, StageNo: 1, MediaType: models.MediaImage, MediaUrl: "https://example.com/snake1.png", DisplayTimeSec: 10, RewardCoins: 10},
		{AnimalID: snake.ID, StageNo: 2, MediaType: models.MediaImage, MediaUrl: "https://example.com/snake2.png", DisplayTimeSec: 15, RewardCoins: 20},
		{AnimalID: snake.ID, StageNo: 3, MediaType: models.MediaVideo, MediaUrl: "https://example.com/snake_video.mp4", DisplayTimeSec: 30, RewardCoins: 50},

		// Spider Stages
		{AnimalID: spider.ID, StageNo: 1, MediaType: models.MediaImage, MediaUrl: "https://example.com/spider1.png", DisplayTimeSec: 10, RewardCoins: 10},
		{AnimalID: spider.ID, StageNo: 2, MediaType: models.MediaVideo, MediaUrl: "https://example.com/spider_video.mp4", DisplayTimeSec: 20, RewardCoins: 30},
	}
	for _, s := range stages {
		database.DB.FirstOrCreate(&s, models.Stage{AnimalID: s.AnimalID, StageNo: s.StageNo})
	}

	// 6. Create Rewards
	fmt.Println("Seeding Rewards...")
	rewards := []models.Reward{
		{Name: "ตุ๊กตาหมี", Description: "ตุ๊กตาหมีน่ารัก", ImageUrl: "https://example.com/bear.png", CostCoins: 100, Stock: 5},
		{Name: "ส่วนลด 10%", Description: "คูปองส่วนลด 10%", ImageUrl: "https://example.com/coupon.png", CostCoins: 50, Stock: 10},
		{Name: "บัตรสตาร์บัคส์ 100 บาท", Description: "บัตรเงินสด", ImageUrl: "https://example.com/starbucks.png", CostCoins: 500, Stock: 2},
	}
	for _, r := range rewards {
		database.DB.FirstOrCreate(&r, models.Reward{Name: r.Name})
	}

	fmt.Println("✅ Database has been successfully seeded!")
}
