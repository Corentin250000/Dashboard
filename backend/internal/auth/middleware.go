package auth

import (
	"errors"

	"dashboard-backend/internal/models"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

const currentUserLocalKey = "current_user"

// RequireAuth authenticates the session_id HTTP-only cookie and exposes the
// corresponding user to handlers through CurrentUser.
func RequireAuth(c fiber.Ctx) error {
	sessionID := c.Cookies(sessionCookieName)
	if sessionID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "Authentication required",
		})
	}

	user, err := authenticateSession(sessionID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Unable to validate session",
			})
		}
		clearSessionCookie(c)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "Session expired or invalid",
		})
	}

	c.Locals(currentUserLocalKey, user)
	c.Locals(sessionCookieName, sessionID)
	return c.Next()
}

// CurrentUser returns the user authenticated by RequireAuth.
func CurrentUser(c fiber.Ctx) (*models.User, bool) {
	user, ok := c.Locals(currentUserLocalKey).(*models.User)
	return user, ok
}

// Logout invalidates the current session and clears the browser cookie.
func Logout(c fiber.Ctx) error {
	if sessionID := c.Locals(sessionCookieName); sessionID != nil {
		if err := deleteSession(sessionID.(string)); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Unable to log out",
			})
		}
	}
	clearSessionCookie(c)
	return c.JSON(fiber.Map{"message": "Logged out"})
}

func clearSessionCookie(c fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		HTTPOnly: true,
		Secure:   c.Protocol() == "https",
		SameSite: "Lax",
		MaxAge:   -1,
	})
}
