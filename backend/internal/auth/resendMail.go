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
// @Description Send a new email confirmation token for an unconfirmed account. Browser clients use the temporary HTTP-only pending_confirmation cookie; non-browser clients may send the legacy token body.
// @Tags auth
// @Accept json
// @Produce json
// @Param token body ResendRequest false "Optional current confirmation token for non-browser clients"
// @Success 200 {object} ResendSuccessResponse
// @Failure 400 {object} ResendBadRequestResponse
// @Failure 404 {object} ResendNotFoundResponse
// @Failure 500 {object} ResendErrorResponse
// @Failure 503 {object} ResendUnavailableResponse
// @Router /confirm/resend [post]
func ResendConfirmationHandler(c fiber.Ctx) error {
	if db == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status":  "error",
			"message": "database not configured",
		})
	}

	token := strings.TrimSpace(c.Cookies(pendingConfirmationCookieName))
	if token == "" {
		// Keep accepting the token body for non-browser clients.
		var req ResendRequest
		if err := c.Bind().Body(&req); err == nil {
			token = strings.TrimSpace(req.Token)
		}
	}
	if token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "No pending confirmation was found. Register an account first.",
		})
	}

	var user models.User
	if err := db.Where("confirmation_token = ?", token).First(&user).Error; err != nil {
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
		clearPendingConfirmationCookie(c)
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
	user.ConfirmationExpires = ptrTime(time.Now().Add(confirmationLifetime))
	if err := db.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "database error",
		})
	}
	setPendingConfirmationCookie(c, token, *user.ConfirmationExpires)

	go sendConfirmationEmail(user.Email, token)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "ok"})
}
