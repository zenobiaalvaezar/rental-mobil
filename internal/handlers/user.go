package handlers

import (
	"car-rental/internal/models"
	"car-rental/internal/services"
	"car-rental/pkg/database"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
)

type TopUpRequest struct {
	Amount float64 `json:"amount" validate:"required,min=10000"`
}

// GetProfile handler
func GetProfile(c echo.Context) error {
	userID := c.Get("userID").(uint)

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "User not found")
	}

	formattedCreatedAt := user.CreatedAt.Format("2006-01-02 15:04:05")

	return c.JSON(http.StatusOK, map[string]interface{}{
		"id":             user.ID,
		"email":          user.Email,
		"deposit_amount": user.DepositAmount,
		"created_at":     formattedCreatedAt,
	})
}

// TopUp handler
func TopUp(c echo.Context) error {
	userID := c.Get("userID").(uint)

	var req TopUpRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Update saldo user
	result := database.DB.Model(&models.User{}).
		Where("id = ?", userID).
		UpdateColumn("deposit_amount", database.DB.Raw("deposit_amount + ?", req.Amount))

	if result.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to process top up")
	}

	// Get updated user data
	var user models.User
	database.DB.First(&user, userID)

	// Kirim email notifikasi
	emailService := services.NewEmailService()
	go emailService.SendEmail(
		user.Email,
		"Top Up Successful",
		fmt.Sprintf("Your deposit has been topped up with Rp%.2f. Current balance: Rp%.2f",
			req.Amount, user.DepositAmount),
	)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":         "Top up successful",
		"current_balance": user.DepositAmount,
	})
}
