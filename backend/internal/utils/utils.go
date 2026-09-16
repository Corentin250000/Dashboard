// Package utils gathers small dependency-free helpers shared across the backend.
package utils

import (
	"os"
	"strconv"
	"strings"
)

// TrimSpace removes leading and trailing spaces and tabs.
func TrimSpace(s string) string {
	return strings.Trim(s, " \t")
}

// FirstCSV returns the first value of a comma-separated header
// (e.g. X-Forwarded-For: "client, proxy1, proxy2" -> "client"), trimmed.
// Returns "" for an empty input.
func FirstCSV(s string) string {
	i := strings.IndexByte(s, ',')

	if i >= 0 {
		return TrimSpace(s[:i])
	}
	return TrimSpace(s)
}

// GetEnv returns the value of the environment variable key,
// or fallback when unset or empty.
func GetEnv(key, fallback string) string {
	env_variable := os.Getenv(key)

	if env_variable != "" {
		return env_variable
	}
	return fallback
}

// GetEnvInt returns the environment variable key parsed as an integer,
// or fallback when unset, empty or not a valid integer.
func GetEnvInt(key string, fallback int) int {
	env_variable := os.Getenv(key)

	if env_variable != "" {
		n, err := strconv.Atoi(env_variable)
		if err == nil {
			return n
		}
	}
	return fallback
}
