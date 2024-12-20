package handlers

import (
	"car-rental/internal/models"
	"car-rental/internal/services"
	"car-rental/pkg/database"
	"fmt"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"net/http"
	"os"
)

type CreatePaymentRequest struct {
	Amount float64 `json:"amount" validate:"required,min=10000"`
}

func WebhookHandler(c echo.Context) error {
	fmt.Println("Webhook received")

	// Verify Xendit Callback Token
	callbackToken := c.Request().Header.Get("X-Callback-Token")
	if callbackToken != os.Getenv("XENDIT_CALLBACK_TOKEN") {
		fmt.Println("Invalid callback token received:", callbackToken)
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid callback token")
	}

	var webhookData struct {
		ExternalID string  `json:"external_id"`
		Status     string  `json:"status"`
		Amount     float64 `json:"amount"`
		ID         string  `json:"id"`
	}

	if err := c.Bind(&webhookData); err != nil {
		fmt.Println("Error binding webhook data:", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid webhook data")
	}

	fmt.Printf("Webhook data received: %+v\n", webhookData)

	tx := database.DB.Begin()

	// Get payment data
	var payment models.Payment
	if err := tx.Where("external_id = ?", webhookData.ExternalID).First(&payment).Error; err != nil {
		tx.Rollback()
		fmt.Printf("Payment not found for external_id: %s\n", webhookData.ExternalID)
		return echo.NewHTTPError(http.StatusNotFound, "Payment not found")
	}

	// Update payment status
	if err := tx.Model(&payment).Updates(map[string]interface{}{
		"status": webhookData.Status,
	}).Error; err != nil {
		tx.Rollback()
		fmt.Printf("Failed to update payment status: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update payment")
	}

	// If payment successful, update rental and car
	if webhookData.Status == "PAID" {
		fmt.Println("Payment status is PAID, processing updates...")

		// Get rental data with User preloaded
		var rental models.RentalHistory
		if err := tx.Preload("User").First(&rental, payment.RentalID).Error; err != nil {
			tx.Rollback()
			fmt.Printf("Rental not found for ID: %d\n", payment.RentalID)
			return echo.NewHTTPError(http.StatusNotFound, "Rental not found")
		}

		// Update rental status
		if err := tx.Model(&rental).Update("status", "active").Error; err != nil {
			tx.Rollback()
			fmt.Printf("Failed to update rental status: %v\n", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update rental")
		}

		// Update car stock
		if err := tx.Model(&models.Car{}).Where("id = ?", rental.CarID).
			UpdateColumn("stock_availability", gorm.Expr("stock_availability - ?", 1)).Error; err != nil {
			tx.Rollback()
			fmt.Printf("Failed to update car stock: %v\n", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update car stock")
		}

		// Send email notification
		fmt.Printf("Sending email to: %s\n", rental.User.Email)
		emailService := services.NewEmailService()
		err := emailService.SendEmail(
			rental.User.Email,
			"Payment Successful",
			fmt.Sprintf("Your payment of Rp%.2f for rental #%d has been confirmed. Thank you for using our service!",
				webhookData.Amount, rental.ID),
		)
		if err != nil {
			fmt.Printf("Failed to send email: %v\n", err)
			// Don't rollback transaction if email fails
		} else {
			fmt.Println("Email sent successfully")
		}
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		fmt.Printf("Failed to commit transaction: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to process payment")
	}

	fmt.Println("Webhook processed successfully")
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
