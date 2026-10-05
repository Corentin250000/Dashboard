package auth

import (
	"errors"
	"time"

	"dashboard-backend/internal/models"

	"github.com/gofiber/fiber/v3"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// Register creates a local account and sends an email confirmation token.
// @Summary Register
// @Description Create a user account with email and password (minimum 8 characters). The account stays unconfirmed until the emailed token is validated via GET /confirm, and login is blocked meanwhile. No authenticated session is created here; a temporary HTTP-only pending_confirmation cookie is set so the browser can call POST /confirm/resend without a body.
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body RegisterRequest true "Registration credentials (valid email + password of at least 8 characters)"
// @Success 201 {object} RegisterSuccessResponse "Account created: check the inbox for the confirmation email."
// @Failure 400 {object} RegisterBadRequestResponse "Malformed JSON body, invalid email format, or password shorter than 8 characters."
// @Failure 409 {object} RegisterConflictResponse "This email address is already registered."
// @Failure 500 {object} RegisterServerErrorResponse "Password hashing, token generation, or account creation failed."
// @Failure 503 {object} RegisterUnavailableResponse "The database is not configured."
// @Router /auth/register [post]
func Register(c fiber.Ctx) error {
	if db == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status":  "error",
			"message": "database not configured",
		})
	}
	var req RegisterRequest

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

	var existing models.User
	if err := db.Where("email = ?", req.Email).First(&existing).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"status":  "error",
			"message": "This email is already in use",
		})
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Server error",
		})
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Server error",
		})
	}

	token, err := generateSecureToken()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Server error",
		})
	}

	hashStr := string(hash)
	user := models.User{
		Email:               req.Email,
		PasswordHash:        &hashStr,
		IsConfirmed:         false,
		ConfirmationToken:   &token,
		ConfirmationExpires: ptrTime(time.Now().Add(confirmationLifetime)),
	}

	if err := db.Create(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Unable to create account",
		})
	}
	setPendingConfirmationCookie(c, token, *user.ConfirmationExpires)

	go sendConfirmationEmail(user.Email, token)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Account created, please check your email",
	})
}
