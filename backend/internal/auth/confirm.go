package auth

import (
	"errors"
	"time"

	"dashboard-backend/internal/models"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

// ConfirmHandler validates an email confirmation token.
// @Summary Confirm email
// @Description Confirm a user account with the single-use token sent by email (valid 24h). The token is consumed on success and a 7-day session is created automatically; already-confirmed accounts return "already_confirmed".
// @Tags auth
// @Produce json
// @Param token query string true "Email confirmation token from the confirmation link"
// @Success 200 {object} ConfirmedResponse "Account confirmed (or already confirmed)."
// @Failure 400 {object} ConfirmBadRequestResponse "Missing token query parameter."
// @Failure 404 {object} InvalidTokenResponse "Unknown or already-consumed token."
// @Failure 410 {object} ExpiredResponse "Token older than 24h (request a new one via POST /confirm/resend)."
// @Failure 500 {object} ConfirmErrorResponse "Database error while confirming."
// @Router /confirm [get]
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
		clearPendingConfirmationCookie(c)
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
	clearPendingConfirmationCookie(c)

	sessionID, err := createSession(user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "unable to create session",
		})
	}
	writeSessionCookie(c, sessionID)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "confirmed"})
}
