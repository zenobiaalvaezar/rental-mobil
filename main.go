package main

import (
	"car-rental/internal/handlers"
	customMiddleware "car-rental/internal/middleware"
	"car-rental/pkg/database"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log"
)

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Initialize database
	database.InitDB()

	// Create Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Public routes
	e.GET("/", handlers.GetCars)
	e.POST("/api/v1/register", handlers.Register)
	e.POST("/api/v1/login", handlers.Login)

	// Protected routes
	api := e.Group("/api/v1")
	api.Use(customMiddleware.JWT)

	// User routes
	api.GET("/profile", handlers.GetProfile)
	api.POST("/topup", handlers.TopUp)

	// Car routes
	api.GET("/cars", handlers.GetCars)
	api.GET("/cars/:id", handlers.GetCarDetail)

	// Rental routes
	api.POST("/rentals", handlers.CreateRental)
	api.GET("/rentals", handlers.GetUserRentals)
	api.POST("/rentals/:id/return", handlers.ReturnCar)

	// Payment routes
	api.GET("/payments", handlers.GetPaymentHistory)
	api.GET("/payments/:id", handlers.GetPaymentDetail)

	// Webhook route (public)
	e.POST("/api/v1/payments/webhook", handlers.WebhookHandler)

	// Start server
	e.Logger.Fatal(e.Start(":8080"))
}
