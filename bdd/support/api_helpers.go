// Package support provides test utilities and context management for BDD tests
package support

import (
	"fmt"
	"net/http"
)

// handleAPIResponse processes API response and returns error if unsuccessful
// statusCode is the HTTP status code from the response
// body is the raw response body bytes
// Returns error if status code indicates failure (>= 400)
func handleAPIResponse(statusCode int, body []byte) error {
	if statusCode >= 400 {
		errMsg := string(body)
		if errMsg == "" {
			errMsg = http.StatusText(statusCode)
		}
		return fmt.Errorf("API returned error %d: %s", statusCode, errMsg)
	}
	return nil
}
