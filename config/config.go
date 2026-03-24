package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration variables
type Config struct {
	Port         string
	DBHost       string
	DBUser       string
	DBPassword   string
	DBName       string
	DBPort       string
	DatabaseURL  string
	JWTSecret    string
}

// Env is the globally accessible configuration object
var Env *Config

// LoadConfig loads environment variables and sets defaults
func LoadConfig() {
	if err := godotenv.Load(); err != nil {
		log.Println("Note: .env file not found, depending on system environment variables")
	}

	Env = &Config{
		Port:         getEnvOrDefault("PORT", "8080"),
		DBHost:       getEnvOrDefault("DB_HOST", "localhost"),
		DBUser:       getEnvOrDefault("DB_USER", "postgres"),
		DBPassword:   getEnvOrDefault("DB_PASSWORD", ""),
		DBName:       getEnvOrDefault("DB_NAME", "fearfree"),
		DBPort:       getEnvOrDefault("DB_PORT", "5432"),
		DatabaseURL:  getEnvOrDefault("DATABASE_URL", ""),
		JWTSecret:    getEnvOrDefault("JWT_SECRET", ""),
	}

	if Env.JWTSecret == "" || Env.JWTSecret == "supersecretkey" {
		log.Fatal("FATAL: JWT_SECRET environment variable must be set to a secure value (not empty or 'supersecretkey')")
	}
}

// getEnvOrDefault gets an environment variable or returns a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
