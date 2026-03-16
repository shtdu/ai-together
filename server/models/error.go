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

package models

import "net/http"

// ErrorResponse is the standard error response format for all API errors
type ErrorResponse struct {
	Error   string `json:"error"`             // User-friendly error message
	Code    string `json:"code,omitempty"`    // Machine-readable error code
	Details string `json:"details,omitempty"` // Additional context (debug mode only)
	Request string `json:"request,omitempty"` // Request ID for support tracking
}

// ValidationError represents a single field validation error
type ValidationError struct {
	Field   string `json:"field"`   // Field name that failed validation
	Message string `json:"message"` // Validation error message
}

// ValidationErrorResponse is the error response format for validation errors
type ValidationErrorResponse struct {
	Error  string            `json:"error"`  // User-friendly error message
	Code   string            `json:"code"`   // Machine-readable error code = "VALIDATION_ERROR"
	Errors []ValidationError `json:"errors"` // List of validation errors
}

// HTTPStatus returns the appropriate HTTP status code for an error code
func HTTPStatus(code string) int {
	switch code {
	case "VALIDATION_ERROR":
		return http.StatusBadRequest
	case "UNAUTHORIZED":
		return http.StatusUnauthorized
	case "FORBIDDEN":
		return http.StatusForbidden
	case "NOT_FOUND":
		return http.StatusNotFound
	case "CONFLICT":
		return http.StatusConflict
	case "BUSINESS_RULE_VIOLATION":
		return http.StatusUnprocessableEntity
	case "SERVICE_UNAVAILABLE":
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

// Error codes constants
const (
	ErrCodeValidation         = "VALIDATION_ERROR"
	ErrCodeUnauthorized       = "UNAUTHORIZED"
	ErrCodeForbidden          = "FORBIDDEN"
	ErrCodeNotFound           = "NOT_FOUND"
	ErrCodeConflict           = "CONFLICT"
	ErrCodeBusinessRule       = "BUSINESS_RULE_VIOLATION"
	ErrCodeInternal           = "INTERNAL_ERROR"
	ErrCodeServiceUnavailable = "SERVICE_UNAVAILABLE"
)

// Common error responses
var (
	ErrInvalidRequest = ErrorResponse{
		Error: "Invalid request payload",
		Code:  ErrCodeValidation,
	}

	ErrInvalidCredentials = ErrorResponse{
		Error: "Invalid credentials",
		Code:  ErrCodeUnauthorized,
	}

	ErrUnauthorized = ErrorResponse{
		Error: "Authentication required",
		Code:  ErrCodeUnauthorized,
	}

	ErrInvalidToken = ErrorResponse{
		Error: "Invalid or expired token",
		Code:  ErrCodeUnauthorized,
	}

	ErrTokenExpired = ErrorResponse{
		Error: "Token has expired",
		Code:  ErrCodeUnauthorized,
	}

	ErrForbidden = ErrorResponse{
		Error: "Insufficient permissions to perform this action",
		Code:  ErrCodeForbidden,
	}

	ErrNotFound = ErrorResponse{
		Error: "Resource not found",
		Code:  ErrCodeNotFound,
	}

	ErrInternal = ErrorResponse{
		Error: "An internal server error occurred",
		Code:  ErrCodeInternal,
	}
)
