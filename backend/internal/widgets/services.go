package widgets

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"strings"
	"time"

	"dashboard-backend/internal/auth"
	"dashboard-backend/internal/models"
	"dashboard-backend/internal/utils"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/markbates/goth"
	"gorm.io/gorm"
)

const (
	serviceOAuthStateCookiePrefix   = "service_oauth_state_"
	serviceOAuthSessionCookiePrefix = "service_oauth_session_"
	serviceOAuthCookieLifetime      = 10 * time.Minute
)

type ServiceResponse struct {
	Name               string                  `json:"name" example:"google"`
	DisplayName        string                  `json:"display_name" example:"Google"`
	Description        string                  `json:"description" example:"Google services"`
	RequiresConnection bool                    `json:"requires_connection" example:"true"`
	Connected          bool                    `json:"connected" example:"false"`
	Widgets            []ServiceWidgetResponse `json:"widgets"`
}

type ServiceWidgetResponse struct {
	Type           string   `json:"type" example:"google.calendar"`
	AvailableSizes []string `json:"available_sizes" example:"small,wide,big"`
}

type ServicesResponse struct {
	Services []ServiceResponse `json:"services"`
}

type ServiceErrorResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"The requested service does not exist."`
}

type ServiceUnauthorizedResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"Authentication required."`
}

type ServiceNotFoundResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"The requested service does not exist."`
}

type ServiceNotConfiguredResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"OAuth credentials for this service are not configured."`
}

type ServiceBadRequestResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"This service does not require a connection."`
}

type ServiceOAuthStateResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"Invalid OAuth state. Restart the connection flow."`
}

type ServiceOAuthProviderResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"The service rejected OAuth authorization."`
}

type ServiceServerErrorResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"Unable to persist the service connection."`
}

type ServiceUnavailableResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"The database is not configured."`
}

type ServiceDisconnectErrorResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"Unable to disconnect the service."`
}

// SetServicesDB is kept for callers that initialize service handlers
// independently. The handlers share the database configured by SetDB.
func SetServicesDB(database *gorm.DB) {
	db = database
}

// ListUserServices returns the service catalog and the current user's
// connection status without exposing credentials or tokens.
// @Summary List services
// @Description Return every catalog service, whether it requires a third-party connection, whether the current user is connected, and its available widget sizes.
// @Tags services
// @Produce json
// @Success 200 {object} ServicesResponse "Service catalog with per-user connection status."
// @Failure 401 {object} ServiceUnauthorizedResponse "The session_id cookie is missing, expired or invalid."
// @Failure 500 {object} ServiceServerErrorResponse "The service catalog or the user's connection statuses could not be loaded."
// @Failure 503 {object} ServiceUnavailableResponse "The database is not configured."
// @Router /api/me/services [get]
func ListUserServices(c fiber.Ctx) error {
	if db == nil {
		return serviceError(c, fiber.StatusServiceUnavailable, "The database is not configured.")
	}
	user, ok := auth.CurrentUser(c)
	if !ok {
		return serviceError(c, fiber.StatusUnauthorized, "Authentication required.")
	}

	var services []models.Service
	if err := db.Preload("Widgets").Find(&services).Error; err != nil {
		return serviceError(c, fiber.StatusInternalServerError, "Unable to load services.")
	}
	var connections []models.UserService
	if err := db.Where("user_id = ?", user.ID).Find(&connections).Error; err != nil {
		return serviceError(c, fiber.StatusInternalServerError, "Unable to load service connections.")
	}
	connected := make(map[uuidKey]bool, len(connections))
	for _, connection := range connections {
		connected[uuidKey(connection.ServiceID)] = hasServiceCredentials(connection)
	}

	response := ServicesResponse{Services: make([]ServiceResponse, 0, len(services))}
	for _, service := range services {
		item := ServiceResponse{
			Name:               service.Name,
			DisplayName:        service.DisplayName,
			Description:        service.Description,
			RequiresConnection: service.AuthType != "none",
			Connected:          connected[uuidKey(service.ID)],
			Widgets:            []ServiceWidgetResponse{},
		}
		for _, widget := range service.Widgets {
			item.Widgets = append(item.Widgets, ServiceWidgetResponse{
				Type:           service.Name + "." + widget.Name,
				AvailableSizes: allowedSizeCodes(widget.ID),
			})
		}
		response.Services = append(response.Services, item)
	}
	return c.JSON(response)
}

// StartOAuth starts the OAuth connection flow for a service on behalf of the
// already authenticated application user.
// @Summary Connect a service
// @Description Start a third-party connection for Spotify (OAuth 2.0) or Steam (OpenID 2.0 via Goth). Google and GitHub are reserved for application login. The browser is redirected to the provider; the callback stores encrypted tokens or the Steam account identifier for the current user.
// @Tags services
// @Produce json
// @Param service path string true "Service technical name" Enums(spotify,steam)
// @Success 302 "Redirect to the Spotify OAuth 2.0 or Steam OpenID 2.0 authorization page."
// @Failure 400 {object} ServiceBadRequestResponse "The selected service has auth_type=none and does not need a connection."
// @Failure 401 {object} ServiceUnauthorizedResponse "The session_id cookie is missing, expired or invalid."
// @Failure 404 {object} ServiceNotFoundResponse "The service is not present in the catalog, or no Goth provider is registered for it."
// @Failure 500 {object} ServiceServerErrorResponse "The OAuth state, provider authorization URL or OAuth session could not be created."
// @Failure 503 {object} ServiceUnavailableResponse "The database or the service OAuth credentials are not configured."
// @Router /api/me/services/{service}/connect [get]
func StartOAuth(c fiber.Ctx) error {
	if db == nil {
		return serviceError(c, fiber.StatusServiceUnavailable, "The database is not configured.")
	}
	if _, ok := auth.CurrentUser(c); !ok {
		return serviceError(c, fiber.StatusUnauthorized, "Authentication required.")
	}
	serviceName := strings.ToLower(strings.TrimSpace(c.Params("service")))
	service, err := findService(serviceName)
	if err != nil {
		return serviceNotFound(c)
	}
	if service.AuthType == "none" {
		return serviceError(c, fiber.StatusBadRequest, "This service does not require a connection.")
	}
	if !serviceProviderConfigured(serviceName) {
		return serviceError(c, fiber.StatusServiceUnavailable, "OAuth credentials for this service are not configured.")
	}

	provider, err := goth.GetProvider(serviceProviderName(serviceName))
	if err != nil {
		return serviceError(c, fiber.StatusNotFound, "OAuth is not configured for this service.")
	}
	state, err := randomServiceToken()
	if err != nil {
		return serviceError(c, fiber.StatusInternalServerError, "Unable to create the OAuth state.")
	}
	oauthSession, err := provider.BeginAuth(state)
	if err != nil {
		return serviceError(c, fiber.StatusInternalServerError, "Unable to start the OAuth flow.")
	}
	authURL, err := oauthSession.GetAuthURL()
	if err != nil {
		return serviceError(c, fiber.StatusInternalServerError, "Unable to build the OAuth provider URL.")
	}

	secure := c.Protocol() == "https"
	c.Cookie(&fiber.Cookie{Name: serviceOAuthStateCookiePrefix + serviceName, Value: state, HTTPOnly: true, Secure: secure, SameSite: "Lax", Expires: time.Now().Add(serviceOAuthCookieLifetime)})
	c.Cookie(&fiber.Cookie{Name: serviceOAuthSessionCookiePrefix + serviceName, Value: base64.RawURLEncoding.EncodeToString([]byte(oauthSession.Marshal())), HTTPOnly: true, Secure: secure, SameSite: "Lax", Expires: time.Now().Add(serviceOAuthCookieLifetime)})
	return c.Redirect().To(authURL)
}

// OAuthCallback completes a service OAuth connection and stores encrypted
// access and refresh tokens in user_services.
// @Summary Complete service connection
// @Description Complete the Spotify OAuth 2.0 or Steam OpenID 2.0 flow, encrypt the received Spotify tokens (or store the Steam account identifier), upsert the current user's service connection, then redirect to the frontend subscription page.
// @Tags services
// @Produce json
// @Param service path string true "Service technical name" Enums(spotify,steam)
// @Param state query string false "OAuth state returned by Spotify; Steam OpenID does not return this parameter"
// @Param code query string false "Authorization code returned by the provider"
// @Success 302 "Connection stored; redirect to the frontend /subscription page."
// @Failure 400 {object} ServiceOAuthStateResponse "The state does not match the service OAuth cookie, or the temporary OAuth session expired/cannot be decoded."
// @Failure 401 {object} ServiceUnauthorizedResponse "The current Dashboard session is missing, expired or invalid."
// @Failure 404 {object} ServiceNotFoundResponse "The service is not present in the catalog, or no Goth service provider is registered."
// @Failure 502 {object} ServiceOAuthProviderResponse "Spotify/Steam rejected authorization, or the provider profile could not be retrieved."
// @Failure 500 {object} ServiceServerErrorResponse "The service tokens could not be encrypted, loaded or persisted."
// @Router /api/me/services/{service}/callback [get]
func OAuthCallback(c fiber.Ctx) error {
	if db == nil {
		return serviceError(c, fiber.StatusServiceUnavailable, "The database is not configured.")
	}
	user, ok := auth.CurrentUser(c)
	if !ok {
		return serviceError(c, fiber.StatusUnauthorized, "Authentication required.")
	}
	serviceName := strings.ToLower(strings.TrimSpace(c.Params("service")))
	service, err := findService(serviceName)
	if err != nil {
		return serviceNotFound(c)
	}
	provider, err := goth.GetProvider(serviceProviderName(serviceName))
	if err != nil {
		return serviceError(c, fiber.StatusNotFound, "OAuth is not configured for this service.")
	}

	state := c.Cookies(serviceOAuthStateCookiePrefix + serviceName)
	// Steam uses OpenID 2.0 and does not return the OAuth state query
	// parameter. The authenticated application session and Steam's signed
	// OpenID response provide the callback binding for that provider.
	if serviceName != "steam" && (state == "" || state != c.Query("state")) {
		return serviceError(c, fiber.StatusBadRequest, "Invalid OAuth state. Restart the connection flow.")
	}
	encoded := c.Cookies(serviceOAuthSessionCookiePrefix + serviceName)
	if encoded == "" {
		return serviceError(c, fiber.StatusBadRequest, "OAuth session expired. Restart the connection flow.")
	}
	serialized, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return serviceError(c, fiber.StatusBadRequest, "Invalid OAuth session.")
	}
	oauthSession, err := provider.UnmarshalSession(string(serialized))
	if err != nil {
		return serviceError(c, fiber.StatusBadRequest, "Invalid OAuth session.")
	}
	params := url.Values{}
	for key, value := range c.Queries() {
		params.Set(key, value)
	}
	if _, err = oauthSession.Authorize(provider, params); err != nil {
		return serviceError(c, fiber.StatusBadGateway, "The service rejected OAuth authorization.")
	}
	providerUser, err := provider.FetchUser(oauthSession)
	if err != nil {
		return serviceError(c, fiber.StatusBadGateway, "Unable to fetch the service profile.")
	}
	accessToken := ""
	refreshToken := ""
	if providerUser.AccessToken != "" {
		accessToken, err = encryptServiceToken(providerUser.AccessToken)
		if err != nil {
			return serviceError(c, fiber.StatusInternalServerError, "Unable to encrypt the service access token.")
		}
	}
	if providerUser.RefreshToken != "" {
		refreshToken, err = encryptServiceToken(providerUser.RefreshToken)
		if err != nil {
			return serviceError(c, fiber.StatusInternalServerError, "Unable to encrypt the service refresh token.")
		}
	}

	connection := models.UserService{}
	err = db.Where("user_id = ? AND service_id = ?", user.ID, service.ID).First(&connection).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		connection = models.UserService{UserID: user.ID, ServiceID: service.ID}
	} else if err != nil {
		return serviceError(c, fiber.StatusInternalServerError, "Unable to load the service connection.")
	}
	connection.OAuthAccessTokenEnc = &accessToken
	connection.OAuthRefreshTokenEnc = &refreshToken
	connection.OAuthTokenExpiresAt = timePointer(providerUser.ExpiresAt)
	if providerUser.UserID != "" {
		externalID := providerUser.UserID
		connection.ExternalUsername = &externalID
	}
	if err := db.Save(&connection).Error; err != nil {
		return serviceError(c, fiber.StatusInternalServerError, "Unable to persist the service connection.")
	}
	if serviceName == "steam" && providerUser.UserID != "" {
		go func(ownerSteamID string) {
			if err := SyncFriends(context.Background(), ownerSteamID); err != nil {
				log.Printf("Steam friends synchronization failed: %v", err)
			}
		}(providerUser.UserID)
	}
	clearServiceOAuthCookies(c, serviceName)
	return c.Redirect().To(frontendURL("/subscription"))
}

// Disconnect removes the local service connection, encrypted tokens and
// widgets belonging to that service. Provider-side revocation depends on the
// provider and is not attempted generically.
// @Summary Disconnect a service
// @Description Delete the current user's local service connection and its encrypted tokens. Also removes that service's dashboard widget instances. Provider-side token revocation is not generic and may require a provider-specific implementation.
// @Tags services
// @Param service path string true "Service technical name (spotify or steam)" Enums(spotify,steam)
// @Success 204 "Connection, encrypted tokens and the user's widgets for this service deleted; empty response body."
// @Failure 401 {object} ServiceUnauthorizedResponse "The session_id cookie is missing, expired or invalid."
// @Failure 404 {object} ServiceNotFoundResponse "The service is not present in the catalog."
// @Failure 500 {object} ServiceDisconnectErrorResponse "The service connection or associated widget instances could not be deleted."
// @Failure 503 {object} ServiceUnavailableResponse "The database is not configured."
// @Router /api/me/services/{service}/connection [delete]
func Disconnect(c fiber.Ctx) error {
	if db == nil {
		return serviceError(c, fiber.StatusServiceUnavailable, "The database is not configured.")
	}
	user, ok := auth.CurrentUser(c)
	if !ok {
		return serviceError(c, fiber.StatusUnauthorized, "Authentication required.")
	}
	service, err := findService(strings.ToLower(strings.TrimSpace(c.Params("service"))))
	if err != nil {
		return serviceNotFound(c)
	}
	if err := db.Transaction(func(tx *gorm.DB) error {
		var widgetIDs []uuid.UUID
		if err := tx.Model(&models.Widget{}).Where("service_id = ?", service.ID).Pluck("id", &widgetIDs).Error; err != nil {
			return err
		}
		if len(widgetIDs) > 0 {
			if err := tx.Where("user_id = ? AND widget_id IN ?", user.ID, widgetIDs).Delete(&models.WidgetInstance{}).Error; err != nil {
				return err
			}
		}
		return tx.Where("user_id = ? AND service_id = ?", user.ID, service.ID).Delete(&models.UserService{}).Error
	}); err != nil {
		return serviceError(c, fiber.StatusInternalServerError, "Unable to disconnect the service.")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

type uuidKey = uuid.UUID

func findService(name string) (models.Service, error) {
	var service models.Service
	err := db.Where("name = ?", name).First(&service).Error
	return service, err
}

func serviceProviderName(serviceName string) string {
	return "service_" + serviceName
}

func serviceProviderConfigured(serviceName string) bool {
	switch serviceName {
	case "spotify":
		return utils.GetEnv("CLIENT_ID_SPOTIFY", "") != "" && utils.GetEnv("CLIENT_SECRET_SPOTIFY", "") != ""
	case "steam":
		return utils.GetEnv("API_KEY_STEAM", "") != ""
	default:
		return false
	}
}

func allowedSizeCodes(widgetID uuid.UUID) []string {
	var allowed []models.WidgetAllowedSize
	db.Where("widget_id = ?", widgetID).Order("size_code ASC").Find(&allowed)
	codes := make([]string, 0, len(allowed))
	for _, item := range allowed {
		codes = append(codes, item.SizeCode)
	}
	return codes
}

func hasServiceCredentials(connection models.UserService) bool {
	return connection.ExternalUsername != nil || connection.ExternalPasswordEnc != nil || connection.OAuthAccessTokenEnc != nil || connection.OAuthRefreshTokenEnc != nil
}

func randomServiceToken() (string, error) {
	var token [32]byte
	if _, err := io.ReadFull(rand.Reader, token[:]); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(token[:]), nil
}

func serviceEncryptionKey() ([]byte, error) {
	secret := utils.GetEnv("JWT_SECRET", "")
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("service token encryption secret is not configured")
	}
	hash := sha256.Sum256([]byte(secret))
	return hash[:], nil
}

func encryptServiceToken(value string) (string, error) {
	if value == "" {
		return "", nil
	}
	key, err := serviceEncryptionKey()
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(value), nil)
	return base64.RawStdEncoding.EncodeToString(ciphertext), nil
}

func timePointer(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	return &value
}

func clearServiceOAuthCookies(c fiber.Ctx, service string) {
	for _, name := range []string{serviceOAuthStateCookiePrefix + service, serviceOAuthSessionCookiePrefix + service} {
		c.Cookie(&fiber.Cookie{Name: name, Value: "", HTTPOnly: true, Expires: time.Unix(0, 0), MaxAge: -1})
	}
}

func frontendURL(path string) string {
	return fmt.Sprintf("%s%s", strings.TrimSuffix(utils.GetEnv("FRONTEND_URL", "http://127.0.0.1:8081"), "/"), path)
}

func serviceNotFound(c fiber.Ctx) error {
	return serviceError(c, fiber.StatusNotFound, "The requested service does not exist.")
}

func serviceError(c fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{"status": "error", "message": message})
}
