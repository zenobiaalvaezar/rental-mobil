package handlers

import (
	"car-rental/internal/models"
	"car-rental/pkg/database"
	"fmt"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"net/http"
	"time"
)

type CreatePaymentRequest struct {
	Amount float64 `json:"amount" validate:"required,min=10000"`
}

func WebhookHandler(c echo.Context) error {
	// Log webhook data yang diterima
	fmt.Println("----------------------------------------")
	fmt.Println("Webhook received at:", time.Now())

	callbackToken := c.Request().Header.Get("X-CALLBACK-TOKEN")
	fmt.Printf("Callback token: %s\n", callbackToken)

	var webhookData struct {
		ExternalID string  `json:"external_id"`
		Status     string  `json:"status"`
		Amount     float64 `json:"amount"`
		ID         string  `json:"id"`
	}

	if err := c.Bind(&webhookData); err != nil {
		fmt.Printf("Error binding webhook data: %v\n", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid webhook data")
	}

	fmt.Printf("Webhook data: \n")
	fmt.Printf("- External ID: %s\n", webhookData.ExternalID)
	fmt.Printf("- Status: %s\n", webhookData.Status)
	fmt.Printf("- Amount: %.2f\n", webhookData.Amount)
	fmt.Printf("- ID: %s\n", webhookData.ID)

	tx := database.DB.Begin()

	// Log query yang akan dijalankan
	fmt.Printf("\nSearching for payment in database...\n")
	fmt.Printf("Query: external_id = %s\n", webhookData.ExternalID)

	var payment models.Payment
	if err := tx.Where("external_id = ?", webhookData.ExternalID).First(&payment).Error; err != nil {
		fmt.Printf("Error finding payment: %v\n", err)
		tx.Rollback()
		return echo.NewHTTPError(http.StatusNotFound, "Payment not found")
	}

	fmt.Printf("Payment found! ID: %d\n", payment.ID)

	// Log status update
	fmt.Printf("\nUpdating payment status...\n")
	if err := tx.Model(&payment).Updates(map[string]interface{}{
		"status": webhookData.Status,
	}).Error; err != nil {
		fmt.Printf("Error updating payment: %v\n", err)
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update payment")
	}

	if webhookData.Status == "PAID" {
		fmt.Printf("\nPayment is PAID, updating rental and car...\n")

		var rental models.RentalHistory
		if err := tx.First(&rental, payment.RentalID).Error; err != nil {
			fmt.Printf("Error finding rental: %v\n", err)
			tx.Rollback()
			return echo.NewHTTPError(http.StatusNotFound, "Rental not found")
		}

		fmt.Printf("Found rental ID: %d\n", rental.ID)

		// Update rental status
		if err := tx.Model(&rental).Update("status", "active").Error; err != nil {
			fmt.Printf("Error updating rental status: %v\n", err)
			tx.Rollback()
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update rental")
		}

		// Update car stock
		if err := tx.Model(&models.Car{}).Where("id = ?", rental.CarID).
			UpdateColumn("stock_availability", gorm.Expr("stock_availability - ?", 1)).Error; err != nil {
			fmt.Printf("Error updating car stock: %v\n", err)
			tx.Rollback()
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update car stock")
		}

		fmt.Printf("Successfully updated rental status and car stock\n")
	}

	if err := tx.Commit().Error; err != nil {
		fmt.Printf("Error committing transaction: %v\n", err)
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to process webhook")
	}

	fmt.Printf("\nWebhook processed successfully!\n")
	fmt.Println("----------------------------------------")

	return c.JSON(http.StatusOK, map[string]string{
		"status": "success",
	})
}

// GetPaymentHistory mengambil history pembayaran user
// GetPaymentHistory handler
// GetPaymentHistory mengambil history pembayaran user
func GetPaymentHistory(c echo.Context) error {
	userID := c.Get("userID").(uint)

	var payments []models.Payment
	// Join dengan rental untuk filter by userID
	if err := database.DB.Joins("JOIN rental_history ON rental_history.id = payments.rental_id").
		Where("rental_history.user_id = ?", userID).
		Preload("Rental").
		Preload("Rental.Car").
		Find(&payments).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch payment history")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": payments,
	})
}

// GetPaymentDetail mengambil detail pembayaran tertentu
func GetPaymentDetail(c echo.Context) error {
	userID := c.Get("userID").(uint)
	paymentID := c.Param("id")

	var payment models.Payment
	if err := database.DB.Preload("Rental").
		Preload("Rental.Car").
		Joins("JOIN rental_history ON rental_history.id = payments.rental_id").
		Where("payments.id = ? AND rental_history.user_id = ?", paymentID, userID).
		First(&payment).Error; err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Payment not found")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": payment,
	})
}
