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
	"github.com/markbates/goth/providers/spotify"
	"github.com/markbates/goth/providers/steam"
	"gorm.io/gorm"
)

const (
	oauthStateCookiePrefix   = "oauth_state_"
	oauthSessionCookiePrefix = "oauth_session_"
)

// ConfigureOAuth registers the providers supported directly by Goth.
func ConfigureOAuth() {
	googleProvider := google.New(
		utils.GetEnv("CLIENT_ID_GOOGLE", ""),
		utils.GetEnv("CLIENT_SECRET_GOOGLE", ""),
		"http://127.0.0.1:8080/auth/google/callback",
		"openid",
		"profile",
		"email",
	)
	githubProvider := github.New(
		utils.GetEnv("CLIENT_ID_GITHUB", ""),
		utils.GetEnv("CLIENT_SECRET_GITHUB", ""),
		"http://127.0.0.1:8080/auth/github/callback",
		"user:email",
	)
	serviceSpotifyProvider := spotify.New(
		utils.GetEnv("CLIENT_ID_SPOTIFY", ""),
		utils.GetEnv("CLIENT_SECRET_SPOTIFY", ""),
		"http://127.0.0.1:8080/api/me/services/spotify/callback",
		spotify.ScopeUserReadPrivate,
		spotify.ScopeUserReadEmail,
	)
	serviceSpotifyProvider.SetName("service_spotify")
	serviceSteamProvider := steam.New(
		utils.GetEnv("API_KEY_STEAM", ""),
		"http://127.0.0.1:8080/api/me/services/steam/callback",
	)
	serviceSteamProvider.SetName("service_steam")

	goth.UseProviders(
		googleProvider,
		githubProvider,
		serviceSpotifyProvider,
		serviceSteamProvider,
	)
}

// OAuthStart redirects the browser to the selected provider.
// @Summary Start OAuth authentication
// @Description Start an OAuth authentication flow for a supported provider. Sets two short-lived HTTP-only cookies (OAuth state and session, 10 minutes) then redirects the browser (302) to the provider's consent page. The frontend must navigate to this URL directly (full page load), not via fetch, so cookies and redirects work.
// @Tags auth
// @Produce json
// @Param provider path string true "OAuth provider" Enums(google,github)
// @Success 302 "Redirect to the provider consent page."
// @Failure 404 {object} OAuthProviderNotFoundResponse "Unknown provider name (only google and github are supported)."
// @Failure 503 {object} OAuthNotConfiguredResponse "Provider credentials (CLIENT_ID_<PROVIDER> / CLIENT_SECRET_<PROVIDER>) are missing."
// @Failure 500 {object} OAuthStartErrorResponse "Unable to start OAuth login or to build the provider URL."
// @Router /auth/{provider} [get]
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
// @Summary Complete OAuth authentication
// @Description Handle the provider callback: verify the state cookie, exchange the authorization code, fetch the provider profile, then create or retrieve the local user (OAuth accounts are auto-confirmed, no email verification needed). Persists a 7-day session in the database, sets the HTTP-only session_id cookie, and redirects (302) to the frontend dashboard. On failure returns JSON instead of redirecting.
// @Tags auth
// @Produce json
// @Param provider path string true "OAuth provider" Enums(google,github)
// @Param state query string true "OAuth state returned by the provider (must match the oauth_state_<provider> cookie)"
// @Param code query string false "Authorization code returned by the provider"
// @Success 302 "Authenticated: redirect to the frontend dashboard with the session_id cookie set."
// @Failure 400 {object} OAuthInvalidStateResponse "State mismatch, expired OAuth cookies, or provider returned no email address."
// @Failure 404 {object} OAuthProviderNotFoundResponse "Unknown provider name (only google and github are supported)."
// @Failure 502 {object} OAuthAuthorizationFailedResponse "Code exchange or profile fetch failed at the provider."
// @Failure 500 {object} OAuthAccountErrorResponse "Local account creation, session persistence, or redirect target failed."
// @Router /auth/{provider}/callback [get]
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
	return utils.GetEnv("CLIENT_ID_"+key, "") != "" && utils.GetEnv("CLIENT_SECRET_"+key, "") != ""
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
		Name:     sessionCookieName,
		Value:    sessionID,
		HTTPOnly: true,
		Secure:   c.Protocol() == "https",
		SameSite: "Lax",
		Path:     "/",
		Expires:  time.Now().Add(sessionLifetime),
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
	return fmt.Sprintf("%s%s", strings.TrimSuffix(utils.GetEnv("FRONTEND_URL", "http://127.0.0.1:8081"), "/"), path)
}
