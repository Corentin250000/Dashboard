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
	"github.com/mailjet/mailjet-apiv3-go/v4"
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
	apiKey := utils.GetEnv("API_KEY_MAILJET", "")
	secretKey := utils.GetEnv("SECRET_KEY_MAILJET", "")
	senderEmail := utils.GetEnv("MAIL_SENDER_ACCOUNT", "")

	if apiKey == "" || secretKey == "" {
		log.Printf("confirmation email to %s not sent: missing Mailjet API credentials (token=%s)", email, token)
		return
	}
	if senderEmail == "" {
		log.Printf("confirmation email to %s not sent: missing MAIL_SENDER_ACCOUNT (token=%s)", email, token)
		return
	}

	frontendURL := utils.GetEnv("FRONTEND_URL", "http://localhost:8081")
	confirmationLink := fmt.Sprintf("%s/confirm?token=%s", strings.TrimSuffix(frontendURL, "/"), token)

	client := mailjet.NewMailjetClient(apiKey, secretKey)

	htmlBody, err := mail.RenderConfirmationEmail(confirmationLink)
	if err != nil {
		log.Printf("confirmation email to %s not rendered: %v", email, err)
		return
	}

	recipients := mailjet.RecipientsV31{{Email: email}}
	messages := &mailjet.MessagesV31{
		Info: []mailjet.InfoMessagesV31{
			{
				From:     &mailjet.RecipientV31{Email: senderEmail},
				To:       &recipients,
				Subject:  "Verify your email to get started on Dashboard",
				HTMLPart: htmlBody,
			},
		},
	}

	results, err := client.SendMailV31(messages)
	if err != nil {
		log.Printf("failed to send confirmation email to %s: %v", email, err)
		return
	}
	if results == nil || len(results.ResultsV31) == 0 {
		log.Printf("failed to send confirmation email to %s: Mailjet returned no message result", email)
		return
	}
	for _, result := range results.ResultsV31 {
		if result.Status != "success" {
			log.Printf("failed to send confirmation email to %s: Mailjet returned status %q", email, result.Status)
			return
		}
	}
	log.Printf("confirmation email sent to %s: token=%s", email, token)
}
