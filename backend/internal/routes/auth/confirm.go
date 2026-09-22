package auth

import (
	"errors"
	"time"

	"dashboard-backend/internal/models"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

func ConfirmHandler(c fiber.Ctx) error {
	if db == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status":  "error",
			"message": "database not configured",
		})
	}

	token := c.Query("token")
	if token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "missing token",
		})
	}

	var user models.User
	if err := db.Where("confirmation_token = ?", token).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"status":  "invalid",
				"message": "invalid or already used token",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "database error",
		})
	}

	if user.IsConfirmed {
		return c.JSON(fiber.Map{"status": "already_confirmed"})
	}
	if user.ConfirmationExpires == nil || time.Now().After(*user.ConfirmationExpires) {
		return c.Status(fiber.StatusGone).JSON(fiber.Map{"status": "expired"})
	}

	user.IsConfirmed = true
	user.ConfirmationToken = nil
	if err := db.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "database error",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "confirmed"})
}
