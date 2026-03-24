package main

import (
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := "host=localhost user=postgres password=1234 dbname=fearfree_db port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("DB connect failed:", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("Test1234"), 10)
	if err != nil {
		log.Fatal("Hash failed:", err)
	}

	result := db.Exec("UPDATE users SET password_hash = ? WHERE username IN ('admin','doctor','demo','user01')", string(hash))
	if result.Error != nil {
		log.Fatal("Update failed:", result.Error)
	}
	fmt.Printf("Updated %d users with password Test1234\n", result.RowsAffected)
}
