package widgets

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"dashboard-backend/internal/auth"
	"dashboard-backend/internal/models"
	"dashboard-backend/internal/utils"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

const googleCalendarReauthMessage = "Connect with Google to use Google Calendar."

type CalendarEventDTO struct {
	ID       string `json:"id" example:"event_123"`
	Title    string `json:"title" example:"Team meeting"`
	Start    string `json:"start" example:"2026-10-08T09:00:00+02:00"`
	End      string `json:"end" example:"2026-10-08T10:00:00+02:00"`
	AllDay   bool   `json:"allDay" example:"false"`
	Location string `json:"location,omitempty" example:"Paris office"`
}

type CalendarErrorResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"Google Calendar is temporarily unavailable."`
}

type googleCalendarResponse struct {
	Items []googleCalendarEvent `json:"items"`
}

type googleCalendarEvent struct {
	ID       string             `json:"id"`
	Summary  string             `json:"summary"`
	Location string             `json:"location"`
	Start    googleCalendarDate `json:"start"`
	End      googleCalendarDate `json:"end"`
}

type googleCalendarDate struct {
	DateTime string `json:"dateTime"`
	Date     string `json:"date"`
}

var errCalendarToken = errors.New("google calendar token refresh failed")

// GetCalendarEvents returns events from the authenticated user's primary calendar.
// @Summary Google Calendar upcoming events
// @Description Return normalized events from the authenticated user's primary Google Calendar.
// @Tags Google Calendar widgets
// @Produce json
// @Param timeMin query string false "RFC 3339 period start (defaults to now)" default(now) example(2026-10-08T00:00:00+02:00)
// @Param timeMax query string false "RFC 3339 period end (defaults to now + 24 hours)" default(now + 24h) example(2026-10-09T00:00:00+02:00)
// @Param maxResults query int false "Number of events (1-50)" minimum(1) maximum(50) default(10) example(10)
// @Param keyword query string false "Optional text filter; not a request body field" example(team meeting)
// @Success 200 {array} CalendarEventDTO
// @Failure 400 {object} CalendarErrorResponse "Invalid calendar query parameter."
// @Failure 401 {object} CalendarErrorResponse "Google must be connected or reconnected."
// @Failure 403 {object} CalendarErrorResponse "Calendar API or scope may be missing."
// @Failure 502 {object} CalendarErrorResponse "Google Calendar is temporarily unavailable."
// @Router /api/calendar/events [get]
func GetCalendarEvents(c fiber.Ctx) error {
	if db == nil {
		return calendarError(c, fiber.StatusServiceUnavailable, "The database is not configured.")
	}
	timeMin, timeMax, maxResults, keyword, err := validateCalendarQuery(c.Queries())
	if err != nil {
		log.Printf("Invalid Google Calendar query: %v", err)
		return calendarError(c, fiber.StatusBadRequest, err.Error())
	}
	user, ok := auth.CurrentUser(c)
	if !ok {
		return calendarError(c, fiber.StatusUnauthorized, googleCalendarReauthMessage)
	}

	client, err := calendarHTTPClient(user.ID, false)
	if err != nil {
		if errors.Is(err, errCalendarToken) || errors.Is(err, gorm.ErrRecordNotFound) {
			return calendarError(c, fiber.StatusUnauthorized, googleCalendarReauthMessage)
		}
		log.Printf("Google Calendar client setup failed: %v", err)
		return calendarError(c, fiber.StatusBadGateway, "Google Calendar is temporarily unavailable.")
	}

	query := url.Values{
		"timeMin":      []string{timeMin.Format(time.RFC3339)},
		"timeMax":      []string{timeMax.Format(time.RFC3339)},
		"singleEvents": []string{"true"},
		"orderBy":      []string{"startTime"},
		"maxResults":   []string{strconv.Itoa(maxResults)},
		"timeZone":     []string{"Europe/Paris"},
		"fields":       []string{"items(id,summary,start,end,location)"},
	}
	if keyword != "" {
		query.Set("q", keyword)
	}
	request, err := calendarRequest(query)
	if err != nil {
		log.Printf("Google Calendar request creation failed: %v", err)
		return calendarError(c, fiber.StatusBadGateway, "Google Calendar is temporarily unavailable.")
	}
	response, err := client.Do(request)
	if err != nil {
		if errors.Is(err, errCalendarToken) {
			return calendarError(c, fiber.StatusUnauthorized, googleCalendarReauthMessage)
		}
		log.Printf("Google Calendar request failed: %v", err)
		return calendarError(c, fiber.StatusBadGateway, "Google Calendar is temporarily unavailable.")
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusUnauthorized {
		response.Body.Close()
		// Google can reject a token before oauth2 considers it expired (for
		// example after a provider-side invalidation). Force one refresh and
		// retry before asking the user to reconnect.
		client, refreshErr := calendarHTTPClient(user.ID, true)
		if refreshErr != nil {
			return calendarError(c, fiber.StatusUnauthorized, googleCalendarReauthMessage)
		}
		retry, retryErr := calendarRequest(query)
		if retryErr != nil {
			return calendarError(c, fiber.StatusBadGateway, "Google Calendar is temporarily unavailable.")
		}
		response, err = client.Do(retry)
		if err != nil {
			return calendarError(c, fiber.StatusUnauthorized, googleCalendarReauthMessage)
		}
		defer response.Body.Close()
		if response.StatusCode == http.StatusUnauthorized {
			return calendarError(c, fiber.StatusUnauthorized, googleCalendarReauthMessage)
		}
	}
	if response.StatusCode == http.StatusForbidden {
		return calendarError(c, fiber.StatusForbidden, "The Google Calendar API or Calendar scope may be missing.")
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		log.Printf("Google Calendar returned HTTP status %d", response.StatusCode)
		return calendarError(c, fiber.StatusBadGateway, "Google Calendar is temporarily unavailable.")
	}

	var payload googleCalendarResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		log.Printf("Google Calendar response decoding failed: %v", err)
		return calendarError(c, fiber.StatusBadGateway, "Google Calendar is temporarily unavailable.")
	}
	result := make([]CalendarEventDTO, 0, len(payload.Items))
	for _, event := range payload.Items {
		result = append(result, normalizeCalendarEvent(event))
	}
	return c.JSON(result)
}

func validateCalendarQuery(query map[string]string) (time.Time, time.Time, int, string, error) {
	now := time.Now()
	timeMin := now
	timeMax := now.Add(24 * time.Hour)
	var err error
	if raw := query["timeMin"]; raw != "" {
		timeMin, err = time.Parse(time.RFC3339, raw)
		if err != nil {
			return time.Time{}, time.Time{}, 0, "", errors.New("timeMin must use RFC 3339 format with a timezone")
		}
	}
	if raw := query["timeMax"]; raw != "" {
		timeMax, err = time.Parse(time.RFC3339, raw)
		if err != nil {
			return time.Time{}, time.Time{}, 0, "", errors.New("timeMax must use RFC 3339 format with a timezone")
		}
	}
	if !timeMin.Before(timeMax) {
		return time.Time{}, time.Time{}, 0, "", errors.New("timeMin must be before timeMax")
	}
	maxResults := 10
	if raw := query["maxResults"]; raw != "" {
		maxResults, err = strconv.Atoi(raw)
		if err != nil || maxResults < 1 || maxResults > 50 {
			return time.Time{}, time.Time{}, 0, "", errors.New("maxResults must be an integer between 1 and 50")
		}
	}
	return timeMin, timeMax, maxResults, strings.TrimSpace(query["keyword"]), nil
}

func normalizeCalendarEvent(event googleCalendarEvent) CalendarEventDTO {
	if event.Start.Date != "" {
		// Google's end.date is exclusive for all-day events; the frontend adjusts display.
		return CalendarEventDTO{ID: event.ID, Title: event.Summary, Start: event.Start.Date, End: event.End.Date, AllDay: true, Location: event.Location}
	}
	return CalendarEventDTO{ID: event.ID, Title: event.Summary, Start: event.Start.DateTime, End: event.End.DateTime, Location: event.Location}
}

func calendarRequest(query url.Values) (*http.Request, error) {
	return http.NewRequestWithContext(context.Background(), http.MethodGet, "https://www.googleapis.com/calendar/v3/calendars/primary/events?"+query.Encode(), nil)
}

func calendarHTTPClient(userID uuid.UUID, forceRefresh bool) (*http.Client, error) {
	// New connections use the calendar catalog entry. Fall back to the
	// historical google entry so existing Google connections keep working.
	var connection models.UserService
	var service models.Service
	if err := db.Where("name = ?", "calendar").First(&service).Error; err == nil {
		err = db.Where("user_id = ? AND service_id = ?", userID, service.ID).First(&connection).Error
		if err == nil {
			return calendarClientFromConnection(connection, forceRefresh)
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if err := db.Where("name = ?", "google").First(&service).Error; err != nil {
		return nil, err
	}
	if err := db.Where("user_id = ? AND service_id = ?", userID, service.ID).First(&connection).Error; err != nil {
		return nil, err
	}
	return calendarClientFromConnection(connection, forceRefresh)
}

func calendarClientFromConnection(connection models.UserService, forceRefresh bool) (*http.Client, error) {
	if connection.OAuthAccessTokenEnc == nil {
		return nil, errCalendarToken
	}
	access, err := decryptServiceToken(*connection.OAuthAccessTokenEnc)
	if err != nil || access == "" {
		return nil, errCalendarToken
	}
	refresh := ""
	if connection.OAuthRefreshTokenEnc != nil {
		refresh, err = decryptServiceToken(*connection.OAuthRefreshTokenEnc)
		if err != nil {
			return nil, errCalendarToken
		}
	}
	token := &oauth2.Token{AccessToken: access, RefreshToken: refresh, TokenType: "Bearer"}
	if connection.OAuthTokenExpiresAt != nil {
		token.Expiry = *connection.OAuthTokenExpiresAt
	}
	// x/oauth2 deliberately treats a zero expiry as valid forever. That is
	// unsafe for stored Google access tokens: force a refresh whenever the
	// expiry was not persisted, or when a retry was requested after a 401.
	if calendarTokenNeedsRefresh(token.Expiry, refresh, forceRefresh) {
		token.Expiry = time.Now().Add(-time.Second)
	}
	config := &oauth2.Config{ClientID: utils.GetEnv("CLIENT_ID_GOOGLE", ""), ClientSecret: utils.GetEnv("CLIENT_SECRET_GOOGLE", ""), Endpoint: oauth2.Endpoint{AuthURL: "https://accounts.google.com/o/oauth2/auth", TokenURL: "https://oauth2.googleapis.com/token"}}
	source := &persistingCalendarTokenSource{base: config.TokenSource(context.Background(), token), connection: connection, access: access, refresh: refresh}
	if _, err := source.Token(); err != nil {
		if errors.Is(err, errCalendarToken) {
			return nil, err
		}
		return nil, err
	}
	return oauth2.NewClient(context.Background(), source), nil
}

func calendarTokenNeedsRefresh(expiry time.Time, refreshToken string, forceRefresh bool) bool {
	return refreshToken != "" && (expiry.IsZero() || forceRefresh)
}

type persistingCalendarTokenSource struct {
	base       oauth2.TokenSource
	connection models.UserService
	access     string
	refresh    string
	mu         sync.Mutex
}

func (source *persistingCalendarTokenSource) Token() (*oauth2.Token, error) {
	token, err := source.base.Token()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errCalendarToken, err)
	}
	source.mu.Lock()
	defer source.mu.Unlock()
	if token.AccessToken != source.access || (token.RefreshToken != "" && token.RefreshToken != source.refresh) {
		if err := persistCalendarToken(source.connection, token); err != nil {
			return nil, err
		}
		source.access = token.AccessToken
		if token.RefreshToken != "" {
			source.refresh = token.RefreshToken
		}
	}
	return token, nil
}

func persistCalendarToken(connection models.UserService, token *oauth2.Token) error {
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

func calendarError(c fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(CalendarErrorResponse{Status: "error", Message: message})
}
