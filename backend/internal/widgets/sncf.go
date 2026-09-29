package widgets

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"dashboard-backend/internal/models"
	"dashboard-backend/internal/utils"

	"gorm.io/gorm"
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
			(payload.Pagination.TotalResult > 0 && (page + 1) * sncfStationsPageSize >= payload.Pagination.TotalResult) ||
			(payload.Pagination.ItemsOnPage > 0 && payload.Pagination.ItemsOnPage < sncfStationsPageSize) {
			return nil
		}
		page++
	}
}
