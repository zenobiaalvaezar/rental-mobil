package handlers

import (
	"car-rental/internal/models"
	"car-rental/internal/services"
	"car-rental/pkg/database"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"time"
)

type CreateRentalRequest struct {
	CarID       uint   `json:"car_id" validate:"required"`
	RentalStart string `json:"rental_start" validate:"required"`
	RentalEnd   string `json:"rental_end" validate:"required"`
}

// CreateRental handler
func CreateRental(c echo.Context) error {
	userID := c.Get("userID").(uint)

	var req CreateRentalRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Get user data
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "User not found")
	}

	// Parse rental dates
	rentalStart, err := time.Parse("2006-01-02", req.RentalStart)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid rental start date format. Use YYYY-MM-DD")
	}

	rentalEnd, err := time.Parse("2006-01-02", req.RentalEnd)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid rental end date format. Use YYYY-MM-DD")
	}

	// Get car data
	var car models.Car
	if err := database.DB.First(&car, req.CarID).Error; err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Car not found")
	}

	// Calculate total cost
	days := int(rentalEnd.Sub(rentalStart).Hours() / 24)
	if days == 0 {
		days = 1
	}
	totalCost := car.RentalCosts * float64(days)

	// Create rental record
	rental := models.RentalHistory{
		UserID:      userID,
		CarID:       req.CarID,
		RentalStart: rentalStart,
		RentalEnd:   rentalEnd,
		TotalCost:   totalCost,
		Status:      "pending",
	}

	if err := database.DB.Create(&rental).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create rental")
	}

	// Preload User and Car for response
	if err := database.DB.Preload("User").Preload("Car").First(&rental, rental.ID).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to load rental data")
	}

	// Create payment invoice
	paymentService := services.NewPaymentService()
	invoice, err := paymentService.CreatePayment(user.Email, totalCost, rental.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create payment invoice")
	}

	// Save payment record
	payment := models.Payment{
		RentalID:   rental.ID,
		InvoiceID:  invoice.ID,
		Amount:     invoice.Amount,
		Status:     invoice.Status,
		PaymentURL: invoice.InvoiceURL,
		ExternalID: invoice.ExternalID,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := database.DB.Create(&payment).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to save payment data")
	}

	// Response
	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "Rental created, waiting for payment",
		"rental":  rental,
		"payment": map[string]interface{}{
			"payment_url": invoice.InvoiceURL,
			"amount":      invoice.Amount,
			"status":      invoice.Status,
		},
	})
}

// GetUserRentals handler
func GetUserRentals(c echo.Context) error {
	userID := c.Get("userID").(uint)

	var rentals []models.RentalHistory
	if err := database.DB.
		Preload("Car").
		Preload("User").
		Where("user_id = ?", userID).
		Find(&rentals).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch rentals")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": rentals,
	})
}

// ReturnCar handler
func ReturnCar(c echo.Context) error {
	userID := c.Get("userID").(uint)
	rentalID := c.Param("id")

	var rental models.RentalHistory
	if err := database.DB.Preload("Car").Preload("User").First(&rental, rentalID).Error; err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Rental not found")
	}

	// Validate ownership
	if rental.UserID != userID {
		return echo.NewHTTPError(http.StatusForbidden, "Not authorized")
	}

	// Validate status
	if rental.Status != "active" {
		return echo.NewHTTPError(http.StatusBadRequest, "Rental is not active")
	}

	// Begin transaction
	tx := database.DB.Begin()

	// Update rental status
	if err := tx.Model(&rental).Update("status", "completed").Error; err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update rental")
	}

	// Update car availability
	if err := tx.Model(&rental.Car).Update("stock_availability", rental.Car.StockAvailability+1).Error; err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update car availability")
	}

	// Commit transaction
	tx.Commit()

	// Send email notification
	emailService := services.NewEmailService()
	go emailService.SendEmail(
		rental.User.Email,
		"Car Return Confirmation",
		fmt.Sprintf("You have successfully returned %s on %s",
			rental.Car.Name,
			time.Now().Format("2006-01-02")),
	)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Car returned successfully",
	})
}
