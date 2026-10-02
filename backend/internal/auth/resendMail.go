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

// ResendConfirmationHandler rotates a confirmation token and sends it again.
// @Summary Resend confirmation email
// @Description Rotate the confirmation token of an unconfirmed account and send a new confirmation email. Rate-limited to 3 requests per 15 minutes. Already-confirmed accounts return "already_confirmed".
// @Tags auth
// @Accept json
// @Produce json
// @Param token body ResendRequest true "Current (expired or lost) confirmation token"
// @Success 200 {object} ResendSuccessResponse "New confirmation email queued (or account already confirmed)."
// @Failure 400 {object} ResendBadRequestResponse "Malformed JSON body or missing token."
// @Failure 404 {object} ResendNotFoundResponse "No account matches this token (unknown or already consumed)."
// @Failure 500 {object} ResendErrorResponse "Database error while rotating the token."
// @Failure 503 {object} ResendUnavailableResponse "The database is not configured."
// @Router /confirm/resend [post]
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
