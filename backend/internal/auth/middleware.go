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

// Me returns the user attached to the current session.
// @Summary Current user
// @Description Return the id and email of the account associated with the current browser session. Used by the frontend on page load to decide whether to show "Log in" or "Log out". Requires the session_id cookie.
// @Tags auth
// @Produce json
// @Success 200 {object} CurrentUserResponse "The id and email of the authenticated user."
// @Failure 401 {object} SessionUnauthorizedResponse "Missing session cookie, or expired/invalid session (cookie is cleared)."
// @Failure 500 {object} SessionErrorResponse "Unable to validate session."
// @Router /auth/me [get]
func Me(c fiber.Ctx) error {
	user, ok := CurrentUser(c)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Authenticated user not available",
		})
	}
	return c.JSON(CurrentUserResponse{ID: user.ID.String(), Email: user.Email})
}

// Logout invalidates the current session and clears the browser cookie.
// @Summary Log out
// @Description Delete the current session from the database and clear the HTTP-only session_id cookie. Third-party service tokens (user_services) are kept: disconnecting a service is a separate operation.
// @Tags auth
// @Produce json
// @Success 200 {object} LogoutSuccessResponse "The session was revoked and the cookie cleared."
// @Failure 401 {object} SessionUnauthorizedResponse "Missing session cookie, or expired/invalid session (cookie is cleared)."
// @Failure 500 {object} SessionErrorResponse "Unable to log out (database error)."
// @Router /auth/logout [post]
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
