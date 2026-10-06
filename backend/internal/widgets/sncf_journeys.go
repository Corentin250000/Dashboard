package widgets

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"dashboard-backend/internal/utils"

	"github.com/gofiber/fiber/v3"
)

const sncfAPIBaseURL = "https://api.sncf.com"

var errSncfNoSolution = errors.New("SNCF returned no solution")

type JourneyRequest struct {
	From     string `json:"from" example:"Mulhouse"`
	To       string `json:"to" example:"Belfort-Ville"`
	Datetime string `json:"datetime" example:"2026-10-03T08:00"`
}

type SncfJourneyResponse struct {
	DepartureTime string `json:"departure_time" example:"08:12 (03/10/2026)"`
	ArrivalTime   string `json:"arrival_time" example:"09:05 (03/10/2026)"`
	Transfers     int    `json:"transfers" example:"0"`
}

type SncfJourneysResponse struct {
	Journeys []SncfJourneyResponse `json:"journeys"`
}

type sncfJourneysPayload struct {
	Journeys []sncfJourney `json:"journeys"`
	Error    *sncfAPIError `json:"error"`
}

type sncfJourney struct {
	DepartureDateTime string `json:"departure_date_time"`
	ArrivalDateTime   string `json:"arrival_date_time"`
	Transfers         int    `json:"nb_transfers"`
}

type sncfAPIError struct {
	ID string `json:"id"`
}

func parseJourneyDatetime(value string) (time.Time, error) {
	return time.Parse("2006-01-02T15:04", value)
}

func formatNavitiaJourneyTime(value string) (string, error) {
	parsed, err := time.Parse("20060102T150405", value)
	if err != nil {
		return "", err
	}
	return parsed.Format("15:04 (02/01/2006)"), nil
}

func mapSncfJourneys(payload sncfJourneysPayload) ([]SncfJourneyResponse, error) {
	if payload.Error != nil && payload.Error.ID == "no_solution" {
		return []SncfJourneyResponse{}, nil
	}

	limit := len(payload.Journeys)
	if limit > 5 {
		limit = 5
	}
	journeys := make([]SncfJourneyResponse, 0, limit)
	for _, journey := range payload.Journeys[:limit] {
		departure, err := formatNavitiaJourneyTime(journey.DepartureDateTime)
		if err != nil {
			return nil, fmt.Errorf("invalid departure date from SNCF: %w", err)
		}
		arrival, err := formatNavitiaJourneyTime(journey.ArrivalDateTime)
		if err != nil {
			return nil, fmt.Errorf("invalid arrival date from SNCF: %w", err)
		}
		journeys = append(journeys, SncfJourneyResponse{
			DepartureTime: departure,
			ArrivalTime:   arrival,
			Transfers:     journey.Transfers,
		})
	}
	return journeys, nil
}

func getSncfJourneys(ctx context.Context, fromID, toID string, datetime time.Time) ([]SncfJourneyResponse, error) {
	apiKey := utils.GetEnv("API_KEY_SNCF", "")
	if apiKey == "" {
		return nil, errors.New("SNCF API is not configured")
	}

	query := url.Values{
		"from":     {fromID},
		"to":       {toID},
		"datetime": {datetime.Format("20060102T150405")},
		"count":    {"5"},
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, sncfAPIBaseURL+"/v1/coverage/sncf/journeys?"+query.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("create SNCF journeys request: %w", err)
	}
	request.SetBasicAuth(apiKey, "")

	response, err := (&http.Client{Timeout: 30 * time.Second}).Do(request)
	if err != nil {
		return nil, fmt.Errorf("request SNCF journeys: %w", err)
	}
	defer response.Body.Close()

	var payload sncfJourneysPayload
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode SNCF journeys response: %w", err)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("SNCF journeys returned HTTP %d", response.StatusCode)
	}
	if payload.Error != nil && payload.Error.ID == "no_solution" {
		return []SncfJourneyResponse{}, errSncfNoSolution
	}
	return mapSncfJourneys(payload)
}

// GetJourneys returns up to five journeys between two stations.
// @Summary SNCF journeys
// @Description Searches up to five train journeys between two SNCF stations for a local departure datetime. Station names are matched exactly against the SNCF stop_areas catalog. If Navitia has no solution, the endpoint returns 200 with an empty journeys array. Public route: no authentication required.
// @Tags SNCF
// @Accept json
// @Produce json
// @Param journey body JourneyRequest true "Journey search request. Example: {\"from\":\"Mulhouse\",\"to\":\"Belfort-Ville\",\"datetime\":\"2026-10-03T08:00\"}"
// @Success 200 {object} SncfJourneysResponse "Up to five journeys. Example: {\"journeys\":[{\"departure_time\":\"08:12 (03/10/2026)\",\"arrival_time\":\"09:05 (03/10/2026)\",\"transfers\":0}]}"
// @Failure 400 {object} SncfBadRequestResponse "Malformed JSON body, missing from/to/datetime field, or datetime not formatted as YYYY-MM-DDTHH:mm."
// @Failure 404 {object} SncfNotFoundResponse "The departure or arrival station is not present in the SNCF catalog."
// @Failure 500 {object} SncfServerErrorResponse "The station catalog could not be queried."
// @Failure 502 {object} SncfBadGatewayResponse "SNCF API unreachable, non-2xx response, or invalid payload."
// @Failure 503 {object} SncfUnavailableResponse "The database or API_KEY_SNCF is not configured."
// @Router /api/sncf/journeys [post]
func GetJourneys(c fiber.Ctx) error {
	if db == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status": "error", "message": "The database is not configured.",
		})
	}

	var request JourneyRequest
	if err := c.Bind().Body(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status": "error", "message": "The request body must contain from, to, and datetime fields.",
		})
	}
	request.From = strings.TrimSpace(request.From)
	request.To = strings.TrimSpace(request.To)
	request.Datetime = strings.TrimSpace(request.Datetime)
	if request.From == "" || request.To == "" || request.Datetime == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status": "error", "message": "The request body must contain non-empty from, to, and datetime fields.",
		})
	}
	datetime, err := parseJourneyDatetime(request.Datetime)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status": "error", "message": "datetime must use the format YYYY-MM-DDTHH:mm.",
		})
	}

	fromID, err := getSncfStationID(request.From)
	if err != nil {
		return journeyStationLookupError(c, "departure", err)
	}
	toID, err := getSncfStationID(request.To)
	if err != nil {
		return journeyStationLookupError(c, "arrival", err)
	}

	journeys, err := getSncfJourneys(c.Context(), fromID, toID, datetime)
	if err != nil {
		if errors.Is(err, errSncfNoSolution) {
			return c.JSON(SncfJourneysResponse{Journeys: []SncfJourneyResponse{}})
		}
		log.Printf("SNCF journeys request failed: %v", err)
		if strings.Contains(err.Error(), "not configured") {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status": "error", "message": "The SNCF API is not configured.",
			})
		}
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"status": "error", "message": "The SNCF API is unavailable or returned an invalid response.",
		})
	}
	return c.JSON(SncfJourneysResponse{Journeys: journeys})
}

func journeyStationLookupError(c fiber.Ctx, role string, err error) error {
	if _, ok := err.(*fiber.Error); ok {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status": "error", "message": fmt.Sprintf("The %s station was not found in the SNCF catalog.", role),
		})
	}
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"status": "error", "message": "Unable to query the station catalog.",
	})
}
