package about

import (
	"net"
	"net/http"
	"time"
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
// Add a new Service here and it automatically appears in /about.json.
// A widget WITHOUT params is invalid per the subject, so every widget
// declared here must expose at least one configurable param.
func Registry() []Service {
	return []Service{
		{
			Name: "weather",
			Widgets: []Widget{
				{
					Name:        "city_temperature",
					Description: "Display temperature for a city",
					Params: []Param{
						{Name: "city", Type: "string"},
					},
				},
				{
					Name:        "precipitation_forecast",
					Description: "Display precipitation forecast for a city",
					Params: []Param{
						{Name: "city", Type: "string"},
						{Name: "days", Type: "integer"},
					},
				},
				{
					Name:        "sky_overview",
					Description: "Display sky overview for given coordinates",
					Params: []Param{
						{Name: "latitude", Type: "string"},
						{Name: "longitude", Type: "string"},
					},
				},
			},
		},
		{
			Name: "rss",
			Widgets: []Widget{
				{
					Name:        "article_list",
					Description: "Displaying the list of the last articles",
					Params: []Param{
						{Name: "link", Type: "string"},
						{Name: "number", Type: "integer"},
					},
				},
			},
		},
		{
			Name: "github",
			Widgets: []Widget{
				{
					Name:        "recent_commits",
					Description: "Display recent commits messages and authors",
					Params: []Param{
						{Name: "owner", Type: "string"},
						{Name: "repo", Type: "string"},
						{Name: "number", Type: "integer"},
					},
				},
				{
					Name:        "security_alerts",
					Description: "Display repository security alerts count",
					Params: []Param{
						{Name: "owner", Type: "string"},
						{Name: "repo", Type: "string"},
					},
				},
			},
		},
		{
			Name: "market",
			Widgets: []Widget{
				{
					Name:        "exchange_rate",
					Description: "Display exchange rate for a currency pair",
					Params: []Param{
						{Name: "base", Type: "string"},
						{Name: "target", Type: "string"},
					},
				},
				{
					Name:        "price_history",
					Description: "Display price history for an asset",
					Params: []Param{
						{Name: "symbol", Type: "string"},
						{Name: "days", Type: "integer"},
					},
				},
			},
		},
	}
}

// Build builds the full /about.json payload for the incoming request.
func Build(r *http.Request) Response {
	var out Response
	out.Client.Host = ClientIP(r)
	out.Server.CurrentTime = time.Now().Unix()
	out.Server.Services = Registry()
	return out
}

// ClientIP resolves the caller IP, proxy-aware.
// Priority: X-Forwarded-For (first entry) > X-Real-IP > RemoteAddr.
func ClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Format: "client, proxy1, proxy2"
		for i := 0; i < len(xff); i++ {
			if xff[i] == ',' {
				return trimSpace(xff[:i])
			}
		}
		return trimSpace(xff)
	}
	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return trimSpace(xrip)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func trimSpace(s string) string {
	start := 0
	for start < len(s) && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	end := len(s)
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}
