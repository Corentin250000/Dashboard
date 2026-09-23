package auth

import (
	"errors"
	"strings"
	"time"

	"dashboard-backend/internal/models"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

type ResendRequest struct {
	Token string `json:"token"`
}

func ResendConfirmationHandler(c fiber.Ctx) error {
	if db == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status":  "error",
			"message": "database not configured",
		})
	}

	var req ResendRequest
	if err := c.Bind().Body(&req); err != nil || strings.TrimSpace(req.Token) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "token required",
		})
	}

	var user models.User
	if err := db.Where("confirmation_token = ?", req.Token).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"status":  "invalid",
				"message": "unknown token",
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

	token, err := generateSecureToken()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "unable to generate token",
		})
	}
	user.ConfirmationToken = &token
	user.ConfirmationExpires = ptrTime(time.Now().Add(24 * time.Hour))
	if err := db.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "database error",
		})
	}

	go sendConfirmationEmail(user.Email, token)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "ok"})
}
