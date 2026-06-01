package db

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	  // Load .env file
    errRes := godotenv.Load()
    if errRes != nil {
        fmt.Println("Error loading .env file:", errRes)
    }

	dsn := os.Getenv("DATABASE_URL")

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("Database connection established successfully")
}

func GetDB() *gorm.DB {
	return DB
}

func AutoMigrate(models ...interface{}) error {
	return DB.AutoMigrate(models...)
}