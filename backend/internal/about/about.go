package about

import (
	"time"

	"dashboard-backend/internal/utils"
)

type Param struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type Widget struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Params      []Param `json:"params"`
}

type Service struct {
	Name    string   `json:"name"`
	Widgets []Widget `json:"widgets"`
}

type Response struct {
	Client struct {
		Host string `json:"host"`
	} `json:"client"`
	Server struct {
		CurrentTime int64     `json:"current_time"`
		Services    []Service `json:"services"`
	} `json:"server"`
}

// Registry is the single source of truth for services/widgets.
func Registry() []Service {
	return []Service{
		{
			Name: "sncf",
			Widgets: []Widget{
				{
					Name:        "departures",
					Description: "Display train departures for a station",
					Params: []Param{
						{Name: "station", Type: "string"},
					},
				},
				{
					Name:        "journeys",
					Description: "Search train journeys between two stations for a given date and time",
					Params: []Param{
						{Name: "from", Type: "string"},
						{Name: "to", Type: "string"},
						{Name: "datetime", Type: "string"},
					},
				},
			},
		},
		{
			Name: "spotify",
			Widgets: []Widget{
				{
					Name:        "currently_playing",
					Description: "Display the track currently playing on Spotify; show_progress accepts 0 or 1.",
					Params:      []Param{{Name: "show_progress", Type: "integer"}},
				},
				{
					Name:        "top_items",
					Description: "Display the user's top Spotify artists or tracks; item_type accepts artists or tracks, time_range accepts short_term, medium_term or long_term (default medium_term), and limit accepts 1 to 10 (default 5).",
					Params:      []Param{{Name: "item_type", Type: "string"}, {Name: "time_range", Type: "string"}, {Name: "limit", Type: "integer"}},
				},
			},
		},
		{
			Name: "steam",
			Widgets: []Widget{
				{
					Name:        "friends",
					Description: "Display the Steam status of a friend searched by pseudo",
					Params: []Param{
						{Name: "pseudo", Type: "string"},
					},
				},
				{
					Name:        "friends_online",
					Description: "Display Steam friends currently online, optionally limited to friends in a game",
					Params: []Param{
						{Name: "limit", Type: "integer"},
						{Name: "inGameOnly", Type: "integer"},
					},
				},
			},
		},
	}
}

// Build builds the full /about.json payload.
// clientHost must be resolved by the caller (proxy-aware), e.g. via ClientIP.
func Build(clientHost string) Response {
	var out Response
	out.Client.Host = clientHost
	out.Server.CurrentTime = time.Now().Unix()
	out.Server.Services = Registry()
	return out
}

// ClientIP resolves the caller IP from raw request parts, proxy-aware.
// Priority: X-Forwarded-For (first entry) > X-Real-IP > remoteAddr.
func ClientIP(xForwardedFor, xRealIP, remoteAddr string) string {
	if xForwardedFor != "" {
		return utils.FirstCSV(xForwardedFor)
	}
	if xRealIP != "" {
		return utils.TrimSpace(xRealIP)
	}
	return remoteAddr
}
