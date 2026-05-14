package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port            string
	MongoURI        string
	DBName          string
	JWTSecret       string
	JWTExpiry       time.Duration
	JWTRefreshExpiry time.Duration
	GinMode         string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment")
	}

	jwtExpiryMinutes, _ := strconv.Atoi(getEnv("JWT_EXPIRY_MINUTES", "60"))
	jwtRefreshDays, _ := strconv.Atoi(getEnv("JWT_REFRESH_DAYS", "7"))

	return &Config{
		Port:             getEnv("API_PORT", "8080"),
		MongoURI:         getEnv("MONGO_URI", "mongodb://localhost:27017"),
		DBName:           getEnv("MONGO_DB_NAME", "lambare"),
		JWTSecret:        getEnv("JWT_SECRET", "change-me-in-production-please"),
		JWTExpiry:        time.Duration(jwtExpiryMinutes) * time.Minute,
		JWTRefreshExpiry: time.Duration(jwtRefreshDays) * 24 * time.Hour,
		GinMode:          getEnv("GIN_MODE", "debug"),
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
