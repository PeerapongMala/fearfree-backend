package database

import (
	"fmt"
	"log"
	"time"

	"fearfree-backend/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDB() {
	var dsn string
	if config.Env.DatabaseURL != "" {
		// Render / production: use DATABASE_URL directly
		dsn = config.Env.DatabaseURL
	} else {
		// Local development: build DSN from individual env vars
		dsn = fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=require TimeZone=Asia/Bangkok",
			config.Env.DBHost,
			config.Env.DBUser,
			config.Env.DBPassword,
			config.Env.DBName,
			config.Env.DBPort,
		)
	}

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})

	if err != nil {
		log.Fatal("Failed to connect to database: \n", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatal("Failed to get underlying sql.DB: ", err)
	}
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	sqlDB.SetConnMaxIdleTime(2 * time.Minute)

	log.Println("Database Connected Successfully!")
}
