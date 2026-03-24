package main

import (
	"fearfree-backend/config"
	"fearfree-backend/database"
	"fearfree-backend/models"
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("No .env file found or failed to load")
	}
	config.LoadConfig()
	database.ConnectDB()

	// Fix Admin
	var adminUser models.User
	if err := database.DB.Where("username = ?", "admin").First(&adminUser).Error; err == nil {
		adminUser.Role = "admin"
		hash, _ := bcrypt.GenerateFromPassword([]byte("admin1234"), 10)
		adminUser.PasswordHash = string(hash)
		database.DB.Save(&adminUser)
		fmt.Println("Fixed admin role.")
	}

	// Fix Doctor
	var docUser models.User
	if err := database.DB.Where("username = ?", "doctor").First(&docUser).Error; err == nil {
		docUser.Role = "doctor"
		hash, _ := bcrypt.GenerateFromPassword([]byte("doctor1234"), 10)
		docUser.PasswordHash = string(hash)
		database.DB.Save(&docUser)
		fmt.Println("Fixed doctor role.")
	}
}
