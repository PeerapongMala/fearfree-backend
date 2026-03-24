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

	var user models.User
	if err := database.DB.Where("username = ?", "admin").First(&user).Error; err != nil {
		hash, _ := bcrypt.GenerateFromPassword([]byte("admin1234"), 10)
		user = models.User{
			Username:     "admin",
			PasswordHash: string(hash),
			Email:        "admin@fearfree.test",
			Role:         "admin",
		}
		database.DB.Create(&user)
		fmt.Println("Created admin account: admin / admin1234")
	} else {
		// Force update role to admin just in case it was a patient
		user.Role = "admin"
		// also force password to admin1234
		hash, _ := bcrypt.GenerateFromPassword([]byte("admin1234"), 10)
		user.PasswordHash = string(hash)
		database.DB.Save(&user)
		fmt.Println("Admin exists. Updated role to admin and reset pass to: admin1234")
	}

	// Also make sure we have a Doctor
	var doc models.User
	if err := database.DB.Where("username = ?", "doctor").First(&doc).Error; err != nil {
		hash, _ := bcrypt.GenerateFromPassword([]byte("doctor1234"), 10)
		doc = models.User{
			Username:     "doctor",
			PasswordHash: string(hash),
			Email:        "doctor@fearfree.test",
			Role:         "doctor",
		}
		database.DB.Create(&doc)
		fmt.Println("Created doctor account: doctor / doctor1234")
	} else {
		doc.Role = "doctor"
		hash, _ := bcrypt.GenerateFromPassword([]byte("doctor1234"), 10)
		doc.PasswordHash = string(hash)
		database.DB.Save(&doc)
		fmt.Println("Doctor exists. Updated role to doctor and reset pass to: doctor1234")
	}
}
