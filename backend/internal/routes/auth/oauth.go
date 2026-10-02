package auth

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"dashboard-backend/internal/models"
	"dashboard-backend/internal/utils"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/google"
	"gorm.io/gorm"
)

const (
	oauthStateCookiePrefix   = "oauth_state_"
	oauthSessionCookiePrefix = "oauth_session_"
)

// ConfigureOAuth registers the providers supported directly by Goth.
func ConfigureOAuth() {
	goth.UseProviders(
		google.New(
			utils.GetEnv("CLIENT_ID_GOOGLE", ""),
			utils.GetEnv("CLIENT_SECRET_GOOGLE", ""),
			"http://localhost:8080/auth/google/callback",
			"openid",
			"profile",
			"email",
		),
		github.New(
			utils.GetEnv("CLIENT_ID_GITHUB", ""),
			utils.GetEnv("CLIENT_SECRET_GITHUB", ""),
			"http://localhost:8080/auth/github/callback",
			"user:email",
		),
	)
}

// OAuthStart redirects the browser to the selected provider.
func OAuthStart(c fiber.Ctx) error {
	providerName := strings.ToLower(c.Params("provider"))
	provider, err := goth.GetProvider(providerName)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "OAuth provider not supported",
		})
	}

	if !providerConfigured(providerName) {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status":  "error",
			"message": "OAuth provider is not configured",
		})
	}

	state, err := generateSecureToken()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Unable to start OAuth login",
		})
	}

	session, err := provider.BeginAuth(state)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Unable to start OAuth login",
		})
	}
	authURL, err := session.GetAuthURL()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Unable to build OAuth URL",
		})
	}

	secure := c.Protocol() == "https"
	c.Cookie(&fiber.Cookie{
		Name:     oauthStateCookiePrefix + providerName,
		Value:    state,
		HTTPOnly: true,
		Secure:   secure,
		SameSite: "Lax",
		Expires:  time.Now().Add(10 * time.Minute),
	})
	c.Cookie(&fiber.Cookie{
		Name:     oauthSessionCookiePrefix + providerName,
		Value:    base64.RawURLEncoding.EncodeToString([]byte(session.Marshal())),
		HTTPOnly: true,
		Secure:   secure,
		SameSite: "Lax",
		Expires:  time.Now().Add(10 * time.Minute),
	})

	return c.Redirect().To(authURL)
}

// OAuthCallback completes a Goth login and creates the local session.
func OAuthCallback(c fiber.Ctx) error {
	providerName := strings.ToLower(c.Params("provider"))
	provider, err := goth.GetProvider(providerName)
	if err != nil {
		return oauthError(c, fiber.StatusNotFound, "OAuth provider not supported")
	}

	stateCookie := c.Cookies(oauthStateCookiePrefix + providerName)
	if stateCookie == "" || stateCookie != c.Query("state") {
		return oauthError(c, fiber.StatusBadRequest, "Invalid OAuth state")
	}

	sessionData := c.Cookies(oauthSessionCookiePrefix + providerName)
	if sessionData == "" {
		return oauthError(c, fiber.StatusBadRequest, "OAuth session expired")
	}
	serializedSession, err := base64.RawURLEncoding.DecodeString(sessionData)
	if err != nil {
		return oauthError(c, fiber.StatusBadRequest, "Invalid OAuth session")
	}
	session, err := provider.UnmarshalSession(string(serializedSession))
	if err != nil {
		return oauthError(c, fiber.StatusBadRequest, "Invalid OAuth session")
	}
	params := url.Values{}
	for key, value := range c.Queries() {
		params.Set(key, value)
	}
	if _, err = session.Authorize(provider, params); err != nil {
		return oauthError(c, fiber.StatusBadGateway, "OAuth authorization failed")
	}
	user, err := provider.FetchUser(session)
	if err != nil {
		return oauthError(c, fiber.StatusBadGateway, "Unable to fetch OAuth profile")
	}
	if strings.TrimSpace(user.Email) == "" {
		return oauthError(c, fiber.StatusBadRequest, "OAuth provider did not return an email")
	}

	localUser, err := findOrCreateOAuthUser(user.Email)
	if err != nil {
		log.Printf("OAuth user persistence failed for provider %s: %v", providerName, err)
		return oauthError(c, fiber.StatusInternalServerError, "Unable to create account")
	}
	if err := setSessionCookie(c, localUser.ID); err != nil {
		return oauthError(c, fiber.StatusInternalServerError, "Unable to create session")
	}

	clearOAuthCookies(c, providerName)
	return c.Redirect().To(frontendRedirect("/dashboard"))
}

func providerConfigured(providerName string) bool {
	key := strings.ToUpper(providerName)
	return utils.GetEnv("CLIENT_ID_" + key, "") != "" && utils.GetEnv("CLIENT_SECRET_" + key, "") != ""
}

func findOrCreateOAuthUser(email string) (models.User, error) {
	var user models.User
	err := db.Where("email = ?", email).First(&user).Error
	if err == nil {
		return user, nil
	}
	if err != gorm.ErrRecordNotFound {
		return user, err
	}

	user = models.User{
		Email:       email,
		IsConfirmed: true,
	}
	return user, db.Create(&user).Error
}

func setSessionCookie(c fiber.Ctx, userID uuid.UUID) error {
	sessionID, err := createSession(userID)
	if err != nil {
		return err
	}
	c.Cookie(&fiber.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		HTTPOnly: true,
		Secure:   c.Protocol() == "https",
		SameSite: "Lax",
		Expires:  time.Now().Add(7 * 24 * time.Hour),
	})
	return nil
}

func clearOAuthCookies(c fiber.Ctx, providerName string) {
	for _, name := range []string{
		oauthStateCookiePrefix + providerName,
		oauthSessionCookiePrefix + providerName,
	} {
		c.Cookie(&fiber.Cookie{Name: name, Value: "", Expires: time.Unix(0, 0), HTTPOnly: true})
	}
}

func oauthError(c fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{"status": "error", "message": message})
}

func frontendRedirect(path string) string {
	return fmt.Sprintf("%s%s", strings.TrimSuffix(utils.GetEnv("FRONTEND_URL", "http://localhost:8081"), "/"), path)
}
