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
	ItemsOnPage  int `json:"items_on_page"`
	Departures  json.RawMessage `json:"departures"`
	Disruptions json.RawMessage `json:"disruptions"`
}

// SetDB configures the database used by widget handlers.
func SetDB(database *gorm.DB) {
	db = database
}

// getIdSncfStation returns the Navitia ID for the station name sent by the frontend.
func getIdSncfStation(stationName string) (string, error) {
	var station models.SncfStation

	if err := db.Where("name = ?", stationName).Order("id ASC").First(&station).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", fiber.ErrNotFound
		}
		return "", fmt.Errorf("find SNCF station %q: %w", stationName, err)
	}
	return station.ID, nil
}

// https://api.sncf.com/v1/coverage/sncf/stop_areas/{stop_areas.id}/departures?start_page=0&count=10
func GetDepartures(c fiber.Ctx) error {
	if db == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status":  "error",
			"message": "database not configured",
		})
	}
	var req TrainStationRequest

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}
	req.TrainStation = strings.TrimSpace(req.TrainStation)

	if err := validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid fields",
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
			"message": "Unable to find station",
		})
	}

	apiKey := utils.GetEnv("API_KEY_SNCF", "")
	if apiKey == "" {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status":  "error",
			"message": "SNCF API not configured",
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
			"message": "Unable to create SNCF request",
		})
	}
	request.SetBasicAuth(apiKey, "")

	response, err := (&http.Client{Timeout: 30 * time.Second}).Do(request)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"status":  "error",
			"message": "Unable to reach SNCF API",
		})
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"status":  "error",
			"message": fmt.Sprintf("SNCF API returned HTTP %d", response.StatusCode),
		})
	}

	var payload sncfDeparturesResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid response from SNCF API",
		})
	}

	return c.JSON(fiber.Map{
		"itemsOnPage":  payload.ItemsOnPage,
		"departures":  payload.Departures,
		"disruptions": payload.Disruptions,
	})
}
