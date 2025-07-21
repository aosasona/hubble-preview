package lib

import (
	"net/http"

	"github.com/rs/zerolog/log"
)

// IsHttpReachable checks if a URL is reachable via HTTP.
func IsHttpReachable(domain string) bool {
	resp, err := http.Get(domain)
	if err != nil {
		log.Error().Err(err).Msgf("Error checking URL: %s", domain)
		return false
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Error().Msgf("URL is not reachable: %s, status code: %d", domain, resp.StatusCode)
		return false
	}

	return true
}
