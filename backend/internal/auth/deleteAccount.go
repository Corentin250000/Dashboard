package auth

import (
	"dashboard-backend/internal/models"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

// DeleteAccount permanently deletes the authenticated user's account and all
// data owned by that account.
// @Summary Delete account
// @Description Permanently delete the authenticated user's account, sessions, service connections, and dashboard widgets. The session cookie is cleared after successful deletion.
// @Tags auth
// @Produce json
// @Success 200 {object} DeleteAccountSuccessResponse "Account deleted."
// @Failure 401 {object} SessionUnauthorizedResponse "Missing, expired, or invalid session."
// @Failure 500 {object} DeleteAccountErrorResponse "Unable to delete the account."
// @Failure 503 {object} DeleteAccountUnavailableResponse "The database is not configured."
// @Router /api/me/account [delete]
func DeleteAccount(c fiber.Ctx) error {
	if db == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status":  "error",
			"message": "database not configured",
		})
	}

	user, ok := CurrentUser(c)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Authenticated user not available",
		})
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", user.ID).Delete(&models.WidgetInstance{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", user.ID).Delete(&models.UserService{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", user.ID).Delete(&models.Session{}).Error; err != nil {
			return err
		}
		return tx.Delete(&models.User{}, "id = ?", user.ID).Error
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Unable to delete account",
		})
	}

	clearSessionCookie(c)
	clearPendingConfirmationCookie(c)
	return c.JSON(DeleteAccountSuccessResponse{Message: "Account deleted"})
}
