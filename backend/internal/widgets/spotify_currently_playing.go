package widgets

import (
	"context"
	"errors"
	"net/http"
	"sync"

	"dashboard-backend/internal/auth"
	"dashboard-backend/internal/models"
	"dashboard-backend/internal/utils"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

type CurrentlyPlayingResponse struct {
	Playing    bool     `json:"playing" example:"true"`
	TrackName  string   `json:"track_name,omitempty" example:"Blinding Lights"`
	Artists    []string `json:"artists,omitempty" example:"The Weeknd"`
	AlbumName  string   `json:"album_name,omitempty" example:"After Hours"`
	ProgressMS int      `json:"progress_ms,omitempty" example:"42100"`
	DurationMS int      `json:"duration_ms,omitempty" example:"200040"`
	SpotifyURL string   `json:"spotify_url,omitempty" example:"https://open.spotify.com/track/0VjIjW4GlUZAMYd2vXMi3b"`
}

type spotifyErrorResponse struct {
	Status  string `json:"status" example:"error"`
	Code    string `json:"code" example:"spotify_reauth_required"`
	Message string `json:"message" example:"Reconnect Spotify to use this widget."`
}

var spotifyUserLocks sync.Map // map[uuid.UUID]*sync.Mutex

func spotifyLock(userID uuid.UUID) *sync.Mutex {
	lock, _ := spotifyUserLocks.LoadOrStore(userID, &sync.Mutex{})
	return lock.(*sync.Mutex)
}

// GetSpotifyCurrentlyPlaying returns only the fields needed by the widget.
// @Summary Spotify currently playing
// @Description Returns the track currently playing for the authenticated Spotify account. When playback is stopped, when no track is active, or when Spotify returns a podcast episode, the response is {"playing":false}.
// @Tags Spotify widgets
// @Produce json
// @Success 200 {object} CurrentlyPlayingResponse "Current track, or playing=false when no supported track is active."
// @Failure 401 {object} spotifyErrorResponse "Spotify must be reconnected because the token is invalid, revoked, or lacks the required scope."
// @Failure 403 {object} spotifyErrorResponse "The current user is not subscribed to Spotify."
// @Failure 429 {object} spotifyErrorResponse "Spotify rate limit reached; Retry-After is returned when provided by Spotify."
// @Failure 502 {object} spotifyErrorResponse "Spotify is unavailable or returned an invalid response."
// @Router /api/spotify/currently-playing [get]
func GetSpotifyCurrentlyPlaying(c fiber.Ctx) error {
	user, ok := auth.CurrentUser(c)
	if !ok {
		return spotifyError(c, fiber.StatusUnauthorized, "spotify_reauth_required", "Spotify must be connected again.")
	}
	lock := spotifyLock(user.ID)
	lock.Lock()
	defer lock.Unlock()
	client, transport, err := spotifyClient(user.ID)
	if err != nil {
		return spotifyServiceError(c, err)
	}
	current, err := client.PlayerCurrentlyPlaying(c.Context())
	if err != nil {
		return spotifyAPIError(c, transport, err)
	}
	return c.JSON(mapCurrentlyPlaying(current))
}

func mapCurrentlyPlaying(current *spotify.CurrentlyPlaying) CurrentlyPlayingResponse {
	if current == nil || current.Item == nil {
		return CurrentlyPlayingResponse{Playing: false}
	}
	track := current.Item
	return CurrentlyPlayingResponse{Playing: current.Playing, TrackName: track.Name, Artists: artistNames(track.Artists), AlbumName: track.Album.Name, ProgressMS: int(current.Progress), DurationMS: int(track.Duration), SpotifyURL: track.ExternalURLs["spotify"]}
}

func artistNames(artists []spotify.SimpleArtist) []string {
	names := make([]string, 0, len(artists))
	for _, artist := range artists {
		names = append(names, artist.Name)
	}
	return names
}

type retryAfterTransport struct {
	base       http.RoundTripper
	mu         sync.Mutex
	retryAfter string
	status     int
}

func (t *retryAfterTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	response, err := t.base.RoundTrip(request)
	if response != nil && response.StatusCode == http.StatusTooManyRequests {
		t.mu.Lock()
		t.retryAfter = response.Header.Get("Retry-After")
		t.status = response.StatusCode
		t.mu.Unlock()
	}
	return response, err
}
func (t *retryAfterTransport) header() string      { t.mu.Lock(); defer t.mu.Unlock(); return t.retryAfter }
func (t *retryAfterTransport) responseStatus() int { t.mu.Lock(); defer t.mu.Unlock(); return t.status }

func spotifyClient(userID uuid.UUID) (*spotify.Client, *retryAfterTransport, error) {
	var connection models.UserService
	var service models.Service
	if err := db.Where("name = ?", "spotify").First(&service).Error; err != nil {
		return nil, nil, err
	}
	if err := db.Where("user_id = ? AND service_id = ?", userID, service.ID).First(&connection).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, errNotSubscribed
		}
		return nil, nil, err
	}
	if connection.OAuthAccessTokenEnc == nil {
		return nil, nil, errReauthRequired
	}
	access, err := decryptServiceToken(*connection.OAuthAccessTokenEnc)
	if err != nil {
		return nil, nil, errReauthRequired
	}
	refresh := ""
	if connection.OAuthRefreshTokenEnc != nil {
		refresh, err = decryptServiceToken(*connection.OAuthRefreshTokenEnc)
		if err != nil {
			return nil, nil, errReauthRequired
		}
	}
	token := &oauth2.Token{AccessToken: access, RefreshToken: refresh, TokenType: "Bearer"}
	if connection.OAuthTokenExpiresAt != nil {
		token.Expiry = *connection.OAuthTokenExpiresAt
	}
	config := &oauth2.Config{ClientID: utils.GetEnv("CLIENT_ID_SPOTIFY", ""), ClientSecret: utils.GetEnv("CLIENT_SECRET_SPOTIFY", ""), Endpoint: oauth2.Endpoint{AuthURL: "https://accounts.spotify.com/authorize", TokenURL: "https://accounts.spotify.com/api/token"}}
	validToken, err := config.TokenSource(context.Background(), token).Token()
	if err != nil {
		return nil, nil, errReauthRequired
	}
	oauthClient := config.Client(context.Background(), validToken)
	transport := &retryAfterTransport{base: oauthClient.Transport}
	if validToken.AccessToken != token.AccessToken || (validToken.RefreshToken != "" && validToken.RefreshToken != refresh) {
		if err := persistSpotifyToken(connection, validToken); err != nil {
			return nil, transport, err
		}
	}
	oauthClient.Transport = transport
	return spotify.New(oauthClient), transport, nil
}

var errNotSubscribed = errors.New("spotify service not subscribed")
var errReauthRequired = errors.New("spotify reauthorization required")

func persistSpotifyToken(connection models.UserService, token *oauth2.Token) error {
	access, err := encryptServiceToken(token.AccessToken)
	if err != nil {
		return err
	}
	connection.OAuthAccessTokenEnc = &access
	if token.RefreshToken != "" {
		refresh, err := encryptServiceToken(token.RefreshToken)
		if err != nil {
			return err
		}
		connection.OAuthRefreshTokenEnc = &refresh
	}
	connection.OAuthTokenExpiresAt = &token.Expiry
	return db.Save(&connection).Error
}

func spotifyServiceError(c fiber.Ctx, err error) error {
	if errors.Is(err, errNotSubscribed) {
		return spotifyError(c, fiber.StatusForbidden, "service_not_subscribed", "Spotify is not connected.")
	}
	if errors.Is(err, errReauthRequired) {
		return spotifyError(c, fiber.StatusUnauthorized, "spotify_reauth_required", "Reconnect Spotify to use this widget.")
	}
	return spotifyError(c, fiber.StatusBadGateway, "spotify_unavailable", "Spotify is temporarily unavailable.")
}
func spotifyAPIError(c fiber.Ctx, transport *retryAfterTransport, err error) error {
	var apiErr spotify.Error
	if (errors.As(err, &apiErr) && apiErr.Status == http.StatusTooManyRequests) || transport.responseStatus() == http.StatusTooManyRequests {
		if retry := transport.header(); retry != "" {
			c.Set(fiber.HeaderRetryAfter, retry)
		}
		return spotifyError(c, fiber.StatusTooManyRequests, "spotify_rate_limited", "Spotify rate limit reached.")
	}
	if errors.As(err, &apiErr) && (apiErr.Status == http.StatusUnauthorized || apiErr.Status == http.StatusForbidden) {
		return spotifyError(c, fiber.StatusUnauthorized, "spotify_reauth_required", "Reconnect Spotify to use this widget.")
	}
	return spotifyError(c, fiber.StatusBadGateway, "spotify_unavailable", "Spotify is temporarily unavailable.")
}
func spotifyError(c fiber.Ctx, status int, code, message string) error {
	return c.Status(status).JSON(spotifyErrorResponse{Status: "error", Code: code, Message: message})
}
