package env

import (
	"os"
	"strconv"
	"strings"
)

// GetString returns the environment variable, or defaultValue when unset.
func GetString(key, defaultValue string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}
	return val
}

// GetInt returns the environment variable as an int, or defaultValue when unset/invalid.
func GetInt(key string, defaultValue int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}
	intVal, err := strconv.Atoi(val)
	if err != nil {
		return defaultValue
	}
	return intVal
}

// OverlayString returns the env value when set, otherwise current (typically from YAML).
func OverlayString(current, envKey string) string {
	if v := os.Getenv(envKey); v != "" {
		return v
	}
	return current
}

// OverlayInt returns the env value when set, otherwise current.
func OverlayInt(current int, envKey string) int {
	if v := os.Getenv(envKey); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return current
}

// CSV returns a comma-separated env value as a trimmed slice.
func CSV(key, defaultValue string) []string {
	raw := GetString(key, defaultValue)
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			out = append(out, s)
		}
	}
	return out
}
