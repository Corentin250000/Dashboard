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
	return []Service{}
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
