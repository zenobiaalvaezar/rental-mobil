package database

import (
	"fmt"
	"gorm.io/gorm"
	"log"
	"os"

	"gorm.io/driver/postgres"
)

var DB *gorm.DB

func InitDB() {
	// Build DSN string from environment variables
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=require",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	// Log the DSN for debugging (you can remove this in production)
	log.Println("Connecting to database with DSN:", dsn)

	// Open database connection
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Assign the connection to the global variable
	DB = db
	log.Println("Database connected successfully")
}
