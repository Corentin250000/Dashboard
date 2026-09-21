package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"time"

	"dashboard-backend/internal/mail"
	"dashboard-backend/internal/utils"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/resend/resend-go/v4"
	"gorm.io/gorm"
)

var (
	db       *gorm.DB
	validate = validator.New()
)

// SetDB injects the GORM handle used by the auth handlers.
// It must be called once at startup before registering routes.
func SetDB(database *gorm.DB) {
	db = database
}

// generateSecureToken returns a random 256-bit hex-encoded token
// used for email confirmation (and session IDs).
func generateSecureToken() (string, error) {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

// ptrTime returns a pointer to t (convenience for nullable timestamps).
func ptrTime(t time.Time) *time.Time {
	return &t
}

// createSession creates a new session for userID and returns its ID.
// TODO: persist the session (DB/store) once the sessions table exists;
// for now it returns a signed random token only.
func createSession(userID uuid.UUID) (string, error) {
	token, err := generateSecureToken()
	if err != nil {
		return "", err
	}
	_ = userID
	return token, nil
}

// sendConfirmationEmail queues the confirmation email.
func sendConfirmationEmail(email, token string) {
	apiKey := utils.GetEnv("API_KEY_RESEND", "")

	if apiKey == "" {
		log.Printf("confirmation email to %s not sent: missing API key (token=%s)", email, token)
		return
	}

	frontendURL := utils.GetEnv("FRONTEND_URL", "http://localhost:8081")
	confirmationLink := fmt.Sprintf("%s/confirm?token=%s", strings.TrimSuffix(frontendURL, "/"), token)

	client := resend.NewClient(apiKey)

	htmlBody, err := mail.RenderConfirmationEmail(confirmationLink)
	if err != nil {
		log.Printf("confirmation email to %s not rendered: %v", email, err)
		return
	}

	params := &resend.SendEmailRequest{
		From:    "onboarding@resend.dev",
		To:      []string{email},
		Subject: "Verify your email to get started on Dashboard",
		Html:    htmlBody,
	}

	_, err = client.Emails.Send(params)
	if err != nil {
		log.Printf("failed to send confirmation email to %s: %v", email, err)
		return
	}
	log.Printf("confirmation email sent to %s: token=%s", email, token)
}
