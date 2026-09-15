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
func Registry() []Service {
	return []Service{}
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
