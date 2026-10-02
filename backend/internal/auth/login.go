package auth

import (
	"time"

	"dashboard-backend/internal/models"

	"github.com/gofiber/fiber/v3"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// Login authenticates a confirmed local account and sets its session cookie.
// @Summary Log in
// @Description Authenticate a user with an email and password.
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "Login credentials"
// @Success 200 {object} LoginSuccessResponse
// @Failure 400 {object} LoginBadRequestResponse
// @Failure 401 {object} LoginUnauthorizedResponse
// @Failure 403 {object} LoginUnconfirmedResponse
// @Failure 500 {object} LoginServerErrorResponse
// @Failure 503 {object} LoginUnavailableResponse
// @Router /auth/login [post]
func Login(c fiber.Ctx) error {
	if db == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status":  "error",
			"message": "database not configured",
		})
	}
	var req LoginRequest

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	if err := validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid fields",
		})
	}

	var user models.User
	if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid credentials",
		})
	}

	// Compte OAuth-only : pas de mot de passe local possible
	if user.PasswordHash == nil || *user.PasswordHash == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "This account uses an external login",
		})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(req.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid credentials",
		})
	}

	if !user.IsConfirmed {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"status":  "error",
			"message": "Account not confirmed, please check your email",
		})
	}

	sessionID, err := createSession(user.ID) // insert en DB/store, retourne l'ID de session
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Server error",
		})
	}

	c.Cookie(&fiber.Cookie{
		Name:     sessionCookieName,
		Value:    sessionID,
		HTTPOnly: true,
		Secure:   c.Protocol() == "https",
		SameSite: "Lax",
		Expires:  time.Now().Add(sessionLifetime),
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Logged in",
	})
}
