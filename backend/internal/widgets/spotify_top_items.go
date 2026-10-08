package widgets

import (
	"errors"
	"strconv"

	"dashboard-backend/internal/auth"

	"github.com/gofiber/fiber/v3"
	"github.com/zmb3/spotify/v2"
)

type SpotifyTopItem struct {
	Name       string   `json:"name" example:"Blinding Lights"`
	Artists    []string `json:"artists,omitempty" example:"The Weeknd"`
	SpotifyURL string   `json:"spotify_url" example:"https://open.spotify.com/track/0VjIjW4GlUZAMYd2vXMi3b"`
	Rank       int      `json:"rank" example:"1"`
}

type SpotifyTopResponse struct {
	Items []SpotifyTopItem `json:"items"`
}

// GetSpotifyTopItems returns the authenticated user's top artists or tracks.
// @Summary Spotify top items
// @Description Returns the authenticated user's top Spotify artists or tracks. Example: /api/spotify/top?item_type=tracks&time_range=long_term&limit=5. If time_range is omitted, medium_term is used.
// @Tags Spotify widgets
// @Produce json
// @Param item_type query string true "Top item type" Enums(artists,tracks) example(tracks)
// @Param time_range query string false "Time range used by Spotify" Enums(short_term,medium_term,long_term) default(medium_term) example(medium_term)
// @Param limit query int false "Number of items" minimum(1) maximum(10) default(5) example(5)
// @Success 200 {object} SpotifyTopResponse "Ranked top artists or tracks."
// @Failure 400 {object} spotifyErrorResponse "A query parameter is missing or invalid."
// @Failure 401 {object} spotifyErrorResponse "Spotify must be reconnected because the token is invalid, revoked, or lacks the required scope."
// @Failure 403 {object} spotifyErrorResponse "The current user is not subscribed to Spotify."
// @Failure 429 {object} spotifyErrorResponse "Spotify rate limit reached; Retry-After is returned when provided by Spotify."
// @Failure 502 {object} spotifyErrorResponse "Spotify is unavailable or returned an invalid response."
// @Router /api/spotify/top [get]
func GetSpotifyTopItems(c fiber.Ctx) error {
	itemType, timeRange, limit, err := validateSpotifyTopQuery(c.Queries())
	if err != nil {
		return spotifyError(c, fiber.StatusBadRequest, "invalid_parameters", err.Error())
	}
	user, ok := auth.CurrentUser(c)
	if !ok {
		return spotifyError(c, fiber.StatusUnauthorized, "spotify_reauth_required", "Spotify must be connected again.")
	}
	client, transport, err := spotifyClient(user.ID)
	if err != nil {
		return spotifyServiceError(c, err)
	}
	var response SpotifyTopResponse
	if itemType == "artists" {
		page, callErr := client.CurrentUsersTopArtists(c.Context(), spotify.Limit(limit), spotify.Timerange(toSpotifyRange(timeRange)))
		if callErr != nil {
			return spotifyAPIError(c, transport, callErr)
		}
		response.Items = mapTopArtists(page)
	} else {
		page, callErr := client.CurrentUsersTopTracks(c.Context(), spotify.Limit(limit), spotify.Timerange(toSpotifyRange(timeRange)), spotify.Offset(0))
		if callErr != nil {
			return spotifyAPIError(c, transport, callErr)
		}
		response.Items = mapTopTracks(page)
	}
	return c.JSON(response)
}

func validateSpotifyTopQuery(query map[string]string) (string, string, int, error) {
	itemType := query["item_type"]
	if itemType != "artists" && itemType != "tracks" {
		return "", "", 0, errors.New("item_type must be artists or tracks")
	}
	timeRange := query["time_range"]
	if timeRange == "" {
		timeRange = "medium_term"
	}
	if timeRange != "short_term" && timeRange != "medium_term" && timeRange != "long_term" {
		return "", "", 0, errors.New("time_range must be short_term, medium_term or long_term")
	}
	limit := 5
	if raw := query["limit"]; raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 || parsed > 10 {
			return "", "", 0, errors.New("limit must be an integer between 1 and 10")
		}
		limit = parsed
	}
	return itemType, timeRange, limit, nil
}

func toSpotifyRange(value string) spotify.Range {
	switch value {
	case "medium_term":
		return spotify.MediumTermRange
	case "long_term":
		return spotify.LongTermRange
	default:
		return spotify.ShortTermRange
	}
}

func mapTopArtists(page *spotify.FullArtistPage) []SpotifyTopItem {
	items := make([]SpotifyTopItem, 0)
	if page == nil {
		return items
	}
	for rank, artist := range page.Artists {
		items = append(items, SpotifyTopItem{Name: artist.Name, SpotifyURL: artist.ExternalURLs["spotify"], Rank: rank + 1})
	}
	return items
}

func mapTopTracks(page *spotify.FullTrackPage) []SpotifyTopItem {
	items := make([]SpotifyTopItem, 0)
	if page == nil {
		return items
	}
	for rank, track := range page.Tracks {
		items = append(items, SpotifyTopItem{Name: track.Name, Artists: artistNames(track.Artists), SpotifyURL: track.ExternalURLs["spotify"], Rank: rank + 1})
	}
	return items
}
