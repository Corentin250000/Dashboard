package widgets

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"dashboard-backend/internal/models"
	"dashboard-backend/internal/utils"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

var (
	db       *gorm.DB
	validate = validator.New()
)

const sncfStationsPageSize = 1000

type sncfStationsResponse struct {
	Pagination struct {
		TotalResult int `json:"total_result"`
		ItemsOnPage int `json:"items_on_page"`
	} `json:"pagination"`
	StopAreas []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"stop_areas"`
}

// GetAllStations inserts the Navitia stop_areas catalog into PostgreSQL.
func GetAllStations(database *gorm.DB) error {
	apiKey := utils.GetEnv("API_KEY_SNCF", "")

	if apiKey == "" {
		return errors.New("API_KEY_SNCF is not set")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	page := 0
	for {
		query := url.Values{
			"start_page": {strconv.Itoa(page)},
			"count":      {strconv.Itoa(sncfStationsPageSize)},
			"depth":      {"0"},
		}
		requestURL := "https://api.sncf.com/v1/coverage/sncf/stop_areas?" + query.Encode()
		request, err := http.NewRequest(http.MethodGet, requestURL, nil)
		if err != nil {
			return fmt.Errorf("create SNCF request: %w", err)
		}
		request.SetBasicAuth(apiKey, "")

		response, err := client.Do(request)
		if err != nil {
			return fmt.Errorf("request SNCF stations page %d: %w", page, err)
		}
		if response.Body == nil {
			return fmt.Errorf("SNCF stations page %d returned an empty body", page)
		}

		var payload sncfStationsResponse
		decodeErr := json.NewDecoder(response.Body).Decode(&payload)
		closeErr := response.Body.Close()
		if decodeErr != nil {
			return fmt.Errorf("decode SNCF stations page %d: %w", page, decodeErr)
		}
		if closeErr != nil {
			return fmt.Errorf("close SNCF stations page %d: %w", page, closeErr)
		}
		if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
			return fmt.Errorf("SNCF stations page %d returned HTTP %d", page, response.StatusCode)
		}

		stations := make([]models.SncfStation, 0, len(payload.StopAreas))
		for _, stopArea := range payload.StopAreas {
			stations = append(stations, models.SncfStation{
				ID:   stopArea.ID,
				Name: stopArea.Name,
			})
		}
		if len(stations) > 0 {
			err = database.Create(&stations).Error
			if err != nil {
				return fmt.Errorf("insert SNCF stations page %d: %w", page, err)
			}
		}

		if len(payload.StopAreas) == 0 ||
			(payload.Pagination.TotalResult > 0 && (page+1)*sncfStationsPageSize >= payload.Pagination.TotalResult) ||
			(payload.Pagination.ItemsOnPage > 0 && payload.Pagination.ItemsOnPage < sncfStationsPageSize) {
			return nil
		}
		page++
	}
}

type TrainStationRequest struct {
	TrainStation string `json:"text" validate:"required"`
}

type sncfDeparturesResponse struct {
	Departures  []sncfDeparture  `json:"departures"`
	Disruptions []sncfDisruption `json:"disruptions"`
}

type sncfDeparture struct {
	StopDateTime struct {
		DepartureDateTime     string `json:"departure_date_time"`
		BaseDepartureDateTime string `json:"base_departure_date_time"`
	} `json:"stop_date_time"`
	DisplayInformations struct {
		Direction string `json:"direction"`
	} `json:"display_informations"`
	Links []struct {
		Rel string `json:"rel"`
		ID  string `json:"id"`
	} `json:"links"`
}

type sncfDisruption struct {
	ID              string `json:"id"`
	ImpactedObjects []struct {
		ImpactedStops []struct {
			StopPoint struct {
				ID string `json:"id"`
			} `json:"stop_point"`
			BaseDepartureTime    string `json:"base_departure_time"`
			AmendedDepartureTime string `json:"amended_departure_time"`
			StopTimeEffect       string `json:"stop_time_effect"`
			DepartureStatus      string `json:"departure_status"`
		} `json:"impacted_stops"`
	} `json:"impacted_objects"`
}

type SncfDepartureResponse struct {
	ScheduledDeparture string `json:"scheduled_departure"`
	Departure          string `json:"departure"`
	Destination        string `json:"destination"`
	DelayMinutes       int    `json:"delay_minutes"`
	Cancelled          bool   `json:"cancelled"`
}

type SncfDeparturesResponse struct {
	Departures []SncfDepartureResponse `json:"departures"`
}

type SncfErrorResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"Invalid request body"`
}

type SncfBadRequestResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"The request body must contain a non-empty text field."`
}

type SncfNotFoundResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"The requested station was not found in the SNCF catalog."`
}

type SncfServerErrorResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"Unable to create the SNCF request."`
}

type SncfBadGatewayResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"The SNCF API is unavailable or returned an invalid response."`
}

type SncfUnavailableResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"The SNCF API is not configured."`
}

// SetDB configures the database used by widget handlers.
func SetDB(database *gorm.DB) {
	db = database
}

// getIdSncfStation returns the Navitia ID for the station name sent by the frontend.
func getIdSncfStation(stationName string) (string, error) {
	return getSncfStationID(stationName)
}

// getSncfStationID returns the Navitia ID for an exact station name.
func getSncfStationID(stationName string) (string, error) {
	var station models.SncfStation

	if err := db.Where("name = ?", stationName).Order("id ASC").First(&station).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", fiber.NewError(fiber.StatusNotFound, "The requested station was not found in the SNCF catalog")
		}
		return "", fmt.Errorf("find SNCF station %q: %w", stationName, err)
	}
	return station.ID, nil
}

// GetDepartures returns the next departures for a station.
// @Summary SNCF departures
// @Description Returns the next 10 departures (scheduled time, real time, destination, delay in minutes, cancellation flag) for a station. The station name is matched exactly against the SNCF stop_areas catalog. Public route: no authentication required.
// @Tags SNCF
// @Accept json
// @Produce json
// @Param station body TrainStationRequest true "Station name exactly as listed in the SNCF catalog (e.g. Paris Gare du Nord)"
// @Success 200 {object} SncfDeparturesResponse "Up to 10 upcoming departures, earliest first."
// @Failure 400 {object} SncfBadRequestResponse "Malformed JSON body or missing/empty text field."
// @Failure 404 {object} SncfNotFoundResponse "No station with this exact name in the SNCF catalog."
// @Failure 500 {object} SncfServerErrorResponse "Station lookup or SNCF request construction failed."
// @Failure 502 {object} SncfBadGatewayResponse "SNCF API unreachable, non-2xx response, or invalid payload."
// @Failure 503 {object} SncfUnavailableResponse "API_KEY_SNCF is not set on the server."
// @Router /api/sncf/departures [post]
func GetDepartures(c fiber.Ctx) error {
	if db == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status":  "error",
			"message": "The database is not configured.",
		})
	}
	var req TrainStationRequest

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "The request body must contain a non-empty text field.",
		})
	}
	req.TrainStation = strings.TrimSpace(req.TrainStation)

	if err := validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "The request body must contain a non-empty text field.",
		})
	}

	idStation, err := getIdSncfStation(req.TrainStation)
	if err != nil {
		if fiberError, ok := err.(*fiber.Error); ok {
			return c.Status(fiberError.Code).JSON(fiber.Map{
				"status":  "error",
				"message": fiberError.Message,
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Unable to query the station catalog.",
		})
	}

	apiKey := utils.GetEnv("API_KEY_SNCF", "")
	if apiKey == "" {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status":  "error",
			"message": "The SNCF API is not configured.",
		})
	}

	query := url.Values{
		"start_page": {"0"},
		"count":      {"10"},
	}
	requestURL := "https://api.sncf.com/v1/coverage/sncf/stop_areas/" + url.PathEscape(idStation) + "/departures?" + query.Encode()
	request, err := http.NewRequestWithContext(c.Context(), http.MethodGet, requestURL, nil)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Unable to create the SNCF request.",
		})
	}
	request.SetBasicAuth(apiKey, "")

	response, err := (&http.Client{Timeout: 30 * time.Second}).Do(request)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"status":  "error",
			"message": "Unable to reach the SNCF API.",
		})
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"status":  "error",
			"message": fmt.Sprintf("The SNCF API returned HTTP %d.", response.StatusCode),
		})
	}

	var payload sncfDeparturesResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"status":  "error",
			"message": "The SNCF API returned an invalid response.",
		})
	}

	stationCode := strings.TrimPrefix(idStation, "stop_area:SNCF:")
	results := make([]SncfDepartureResponse, 0, len(payload.Departures))
	for _, departure := range payload.Departures {
		result := SncfDepartureResponse{
			ScheduledDeparture: departure.StopDateTime.BaseDepartureDateTime,
			Departure:          departure.StopDateTime.DepartureDateTime,
			Destination:        departure.DisplayInformations.Direction,
		}

		for _, link := range departure.Links {
			if link.Rel != "disruptions" {
				continue
			}
			for _, disruption := range payload.Disruptions {
				if disruption.ID != link.ID {
					continue
				}
				for _, impactedObject := range disruption.ImpactedObjects {
					for _, impactedStop := range impactedObject.ImpactedStops {
						if !strings.Contains(impactedStop.StopPoint.ID, ":"+stationCode+":") {
							continue
						}
						if impactedStop.StopTimeEffect == "deleted" || impactedStop.DepartureStatus == "deleted" {
							result.Cancelled = true
						}
						if impactedStop.AmendedDepartureTime != "" {
							result.Departure = impactedStop.AmendedDepartureTime
						}
						if delay := departureDelayMinutes(impactedStop.BaseDepartureTime, impactedStop.AmendedDepartureTime); delay > result.DelayMinutes {
							result.DelayMinutes = delay
						}
					}
				}
			}
		}

		if delay := departureDelayMinutes(result.ScheduledDeparture, result.Departure); delay > result.DelayMinutes {
			result.DelayMinutes = delay
		}
		result.ScheduledDeparture = formatDepartureTime(result.ScheduledDeparture)
		result.Departure = formatDepartureTime(result.Departure)
		results = append(results, result)
	}

	return c.JSON(SncfDeparturesResponse{Departures: results})
}

func departureDelayMinutes(scheduled, actual string) int {
	if scheduled == "" || actual == "" {
		return 0
	}
	scheduledTime, err := time.Parse("20060102T150405", scheduled)
	if err != nil {
		return 0
	}
	actualTime, err := time.Parse("20060102T150405", actual)
	if err != nil || actualTime.Before(scheduledTime) {
		return 0
	}
	return int(actualTime.Sub(scheduledTime).Minutes())
}

func formatDepartureTime(value string) string {
	parsedTime, err := time.Parse("20060102T150405", value)
	if err != nil {
		return ""
	}
	return parsedTime.Format("15:04")
}
