package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"github.com/resend/resend-go/v4"
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
	apiKey := os.Getenv("API_KEY_RESEND")
	if apiKey == "" {
		apiKey = os.Getenv("RESEND_API_KEY")
	}

	if apiKey == "" {
		log.Printf("confirmation email to %s not sent: missing API key (token=%s)", email, token)
		return
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}
	confirmationLink := fmt.Sprintf("%s/confirm?token=%s", strings.TrimSuffix(frontendURL, "/"), token)

	client := resend.NewClient(apiKey)

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
				<html lang="en">
				<head>
				<meta charset="UTF-8" />
				<meta name="viewport" content="width=device-width, initial-scale=1.0" />
				<title>Confirm your email</title>
				</head>
				<body style="margin:0; padding:0; background-color:#f4f4f7; font-family:Arial, Helvetica, sans-serif;">
				<table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="background-color:#f4f4f7; padding:40px 0;">
					<tr>
					<td align="center">
						<table role="presentation" width="480" cellpadding="0" cellspacing="0" style="background-color:#ffffff; border-radius:8px; overflow:hidden; box-shadow:0 2px 6px rgba(0,0,0,0.05);">
						<!-- Header -->
						<tr>
							<td style="background-color:#4f46e5; padding:24px; text-align:center;">
							<h1 style="color:#ffffff; margin:0; font-size:20px;">Dashboard</h1>
							</td>
						</tr>

						<!-- Body -->
						<tr>
							<td style="padding:32px 24px;">
							<h2 style="color:#111827; font-size:20px; margin-top:0;">Confirm your email address</h2>
							<p style="color:#374151; font-size:15px; line-height:1.5;">
								Thanks for signing up! Please confirm your email address by clicking the button below.
								This link will expire in 24 hours.
							</p>

							<table role="presentation" cellpadding="0" cellspacing="0" style="margin:32px auto;">
								<tr>
								<td style="border-radius:6px; background-color:#4f46e5;">
									<a href="%s"
									style="display:inline-block; padding:12px 28px; font-size:15px; color:#ffffff; text-decoration:none; font-weight:bold; border-radius:6px;">
									Confirm Email
									</a>
								</td>
								</tr>
							</table>

							<p style="color:#6b7280; font-size:13px; line-height:1.5;">
								If the button doesn't work, copy and paste this link into your browser:<br />
								<a href="%s" style="color:#4f46e5; word-break:break-all;">%s</a>
							</p>

							<p style="color:#9ca3af; font-size:13px; margin-top:32px;">
								If you didn't create an account, you can safely ignore this email.
							</p>
							</td>
						</tr>

						<!-- Footer -->
						<tr>
							<td style="background-color:#f9fafb; padding:16px 24px; text-align:center;">
							<p style="color:#9ca3af; font-size:12px; margin:0;">
								&copy; 2026 Dashboard. All rights reserved.
							</p>
							</td>
						</tr>
						</table>
					</td>
					</tr>
				</table>
			</body>
		</html>`, confirmationLink, confirmationLink, confirmationLink)

    params := &resend.SendEmailRequest{
        From:    "onboarding@resend.dev",
        To:      []string{email},
        Subject: "Verify your email to get started on Dashboard",
        Html:    htmlBody,
    }

    _, err := client.Emails.Send(params)
    if err != nil {
        log.Printf("failed to send confirmation email to %s: %v", email, err)
		return
    }
	log.Printf("confirmation email sent to %s: token=%s", email, token)
}
