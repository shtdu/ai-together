// Copyright (c) 2025 Code Together
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.


package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ValidationError represents a single field validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// UnwrapJSONResponse attempts to parse an error response from the server
// It handles both standard ErrorResponse and ValidationErrorResponse formats
func UnwrapJSONResponse(resp *http.Response) error {
	if resp == nil {
		return fmt.Errorf("nil response")
	}

	// Don't wrap successful responses
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read error response body: %w", err)
	}

	// Restore the body for potential re-reading
	resp.Body = io.NopCloser(bytes.NewReader(body))

	// If body is empty, return a simple status error
	if len(body) == 0 {
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("HTTP %d: %s", resp.StatusCode, http.StatusText(resp.StatusCode)),
		}
	}

	// Try to parse as JSON
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		// Not JSON, return the raw body as the message
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	// Extract common fields
	apiErr := &APIError{
		StatusCode: resp.StatusCode,
		Message:    extractString(data, "error"),
		ErrorCode:  extractString(data, "code"),
		RequestID:  extractString(data, "request"),
		Details:    make(map[string]interface{}),
	}

	// Extract details if present
	if detailsVal, ok := data["details"]; ok && detailsVal != nil {
		if detailsStr, ok := detailsVal.(string); ok {
			apiErr.Details["message"] = detailsStr
		}
	}

	// Check for validation errors
	if errorsArray, ok := data["errors"].([]interface{}); ok {
		for _, e := range errorsArray {
			if errMap, ok := e.(map[string]interface{}); ok {
				apiErr.ValidationErrors = append(apiErr.ValidationErrors, ValidationError{
					Field:   extractString(errMap, "field"),
					Message: extractString(errMap, "message"),
				})
			}
		}
	}

	// If no message found, use status text
	if apiErr.Message == "" {
		apiErr.Message = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	return apiErr
}

// extractString safely extracts a string value from a map
func extractString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// IsAPIError checks if an error is an APIError
func IsAPIError(err error) bool {
	_, ok := err.(*APIError)
	return ok
}

// GetAPIError extracts an APIError from an error, or nil if not an APIError
func GetAPIError(err error) *APIError {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr
	}
	return nil
}
