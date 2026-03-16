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

package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"switch-server/models"
)

const (
	requestIDKey     = "request_id"
	userIDKey        = "user_id"
	tenantIDKey      = "tenant_id"
	contextUserIDKey = "userID"
	contextTenantKey = "tenantID"
)

// RequestIDMiddleware generates a unique request ID for each request
// and injects it into the context for logging and error tracking
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate unique request ID
		requestID := uuid.New().String()

		// Store in context
		c.Set(requestIDKey, requestID)
		c.Set("request_id", requestID) // For backward compatibility

		// Also set as header for client tracking
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

// ErrorLoggingMiddleware provides structured error logging with context
// It handles errors consistently and provides detailed logging for debugging
func ErrorLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		startTime := time.Now()

		// Create response writer wrapper to capture status code
		w := &responseWriter{ResponseWriter: c.Writer, status: http.StatusOK}

		// Replace writer
		c.Writer = w

		// Get request context info
		requestID, _ := c.Get(requestIDKey)
		userID, _ := c.Get(contextUserIDKey)
		tenantID, _ := c.Get(contextTenantKey)

		// Log request with context
		logRequest(c, requestID, userID, tenantID, startTime)

		// Process request
		c.Next()

		// Log any errors that occurred during request processing
		if len(c.Errors) > 0 {
			logErrors(c, requestID, userID, tenantID)
		}

		// Log response completion
		logResponse(c, requestID, userID, tenantID, startTime, w.status)
	}
}

// responseWriter wraps gin.ResponseWriter to capture status code
type responseWriter struct {
	gin.ResponseWriter
	status int
}

func (w *responseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *responseWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.ResponseWriter.Write(b)
}

// logRequest logs incoming request details with context using slog
func logRequest(c *gin.Context, requestID, userID, tenantID interface{}, startTime time.Time) {
	args := []slog.Attr{
		slog.String("event", "request"),
		slog.String("request_id", fmt.Sprintf("%v", requestID)),
		slog.String("method", c.Request.Method),
		slog.String("path", c.Request.URL.Path),
		slog.String("remote_addr", c.ClientIP()),
		slog.String("user_agent", c.Request.UserAgent()),
	}

	if userID != nil {
		args = append(args, slog.Int64("user_id", userID.(int64)))
	}
	if tenantID != nil {
		args = append(args, slog.Int64("tenant_id", tenantID.(int64)))
	} else {
		args = append(args, slog.String("auth", "public"))
	}

	slog.LogAttrs(c.Request.Context(), slog.LevelInfo, "incoming request", args...)
}

// logResponse logs response completion with context using slog
func logResponse(c *gin.Context, requestID, userID, tenantID interface{}, startTime time.Time, statusCode int) {
	duration := time.Since(startTime)

	args := []slog.Attr{
		slog.String("event", "response"),
		slog.String("request_id", fmt.Sprintf("%v", requestID)),
		slog.String("method", c.Request.Method),
		slog.String("path", c.Request.URL.Path),
		slog.Int("status", statusCode),
		slog.Duration("duration", duration),
	}

	if userID != nil {
		args = append(args, slog.Int64("user_id", userID.(int64)))
	}
	if tenantID != nil {
		args = append(args, slog.Int64("tenant_id", tenantID.(int64)))
	}

	// Determine log level based on status code
	level := slog.LevelInfo
	if statusCode >= 400 && statusCode < 500 {
		level = slog.LevelWarn
	} else if statusCode >= 500 {
		level = slog.LevelError
	}

	slog.LogAttrs(c.Request.Context(), level, "request completed", args...)
}

// logErrors logs errors with full context using slog
func logErrors(c *gin.Context, requestID, userID, tenantID interface{}) {
	for _, err := range c.Errors {
		errType := "error"
		level := slog.LevelError
		errTypeStr := "unknown"

		switch err.Type {
		case gin.ErrorTypeBind:
			errType = "validation"
			errTypeStr = "gin.ErrorTypeBind"
			level = slog.LevelWarn
		case gin.ErrorTypePublic:
			errType = "public"
			errTypeStr = "gin.ErrorTypePublic"
			level = slog.LevelInfo
		case gin.ErrorTypePrivate:
			errType = "internal"
			errTypeStr = "gin.ErrorTypePrivate"
			level = slog.LevelError
		}

		args := []slog.Attr{
			slog.String("event", errType),
			slog.String("request_id", fmt.Sprintf("%v", requestID)),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.String("error", err.Error()),
			slog.String("error_type", errTypeStr),
		}

		if userID != nil {
			args = append(args, slog.Int64("user_id", userID.(int64)))
		}
		if tenantID != nil {
			args = append(args, slog.Int64("tenant_id", tenantID.(int64)))
		}

		slog.LogAttrs(c.Request.Context(), level, fmt.Sprintf("%s error", errType), args...)

		// Log stack trace for internal errors in debug mode
		if gin.Mode() == gin.DebugMode && err.Type == gin.ErrorTypePrivate {
			slog.LogAttrs(c.Request.Context(), slog.LevelDebug, "stack trace",
				slog.String("request_id", fmt.Sprintf("%v", requestID)),
				slog.String("stack", string(debug.Stack())),
			)
		}
	}
}

// ErrorResponse sends a standardized error response with logging
// This is a helper function for handlers to send consistent error responses
func ErrorResponse(c *gin.Context, statusCode int, errResp models.ErrorResponse) {
	// Add request ID if available
	if requestID, exists := c.Get("request_id"); exists {
		if reqID, ok := requestID.(string); ok {
			errResp.Request = reqID
		}
	}

	// Add details only in debug mode
	if gin.Mode() != gin.DebugMode {
		errResp.Details = ""
	}

	// Log the error response
	logErrorResponse(c, statusCode, errResp)

	c.JSON(statusCode, errResp)
}

// logErrorResponse logs error response details using slog
func logErrorResponse(c *gin.Context, statusCode int, errResp models.ErrorResponse) {
	requestID, _ := c.Get("request_id")
	userID, _ := c.Get("user_id")
	tenantID, _ := c.Get("tenant_id")

	args := []slog.Attr{
		slog.String("event", "error_response"),
		slog.String("request_id", fmt.Sprintf("%v", requestID)),
		slog.String("error_code", errResp.Code),
		slog.Int("status", statusCode),
		slog.String("message", errResp.Error),
		slog.String("method", c.Request.Method),
		slog.String("path", c.Request.URL.Path),
	}

	if userID != nil {
		args = append(args, slog.Int64("user_id", userID.(int64)))
	}
	if tenantID != nil {
		args = append(args, slog.Int64("tenant_id", tenantID.(int64)))
	}
	if errResp.Details != "" {
		args = append(args, slog.String("details", errResp.Details))
	}

	level := slog.LevelWarn
	if statusCode >= 500 {
		level = slog.LevelError
	}

	slog.LogAttrs(c.Request.Context(), level, "error response sent", args...)
}

// ValidationErrorResponse sends a validation error response with logging
func ValidationErrorResponse(c *gin.Context, validationErrors []models.ValidationError) {
	requestID, _ := c.Get("request_id")

	resp := models.ValidationErrorResponse{
		Error:  "Validation failed",
		Code:   models.ErrCodeValidation,
		Errors: validationErrors,
	}

	// Add request ID if available
	if reqID, ok := requestID.(string); ok {
		c.Header("X-Request-ID", reqID)
	}

	// Log the validation error
	logValidationError(c, validationErrors)

	c.JSON(http.StatusBadRequest, resp)
}

// logValidationError logs validation error details using slog
func logValidationError(c *gin.Context, validationErrors []models.ValidationError) {
	requestID, _ := c.Get("request_id")
	userID, _ := c.Get("user_id")
	tenantID, _ := c.Get("tenant_id")

	// Read request body for logging (if available)
	var bodyStr string
	if c.Request.Body != nil && c.Request.Method != "GET" {
		bodyBytes, _ := io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		bodyStr = string(bodyBytes)
	}

	args := []slog.Attr{
		slog.String("event", "validation_error"),
		slog.String("request_id", fmt.Sprintf("%v", requestID)),
		slog.Int("field_count", len(validationErrors)),
		slog.String("method", c.Request.Method),
		slog.String("path", c.Request.URL.Path),
	}

	if userID != nil {
		args = append(args, slog.Int64("user_id", userID.(int64)))
	}
	if tenantID != nil {
		args = append(args, slog.Int64("tenant_id", tenantID.(int64)))
	}
	if bodyStr != "" {
		args = append(args, slog.String("body", bodyStr))
	}

	slog.LogAttrs(c.Request.Context(), slog.LevelWarn, "validation failed", args...)

	// Log each validation error as a separate log entry
	for _, ve := range validationErrors {
		slog.LogAttrs(c.Request.Context(), slog.LevelWarn, "field validation failed",
			slog.String("request_id", fmt.Sprintf("%v", requestID)),
			slog.String("field", ve.Field),
			slog.String("message", ve.Message),
		)
	}
}

// BindJSONWithValidation binds JSON request body and provides detailed validation errors
// This replaces c.ShouldBindJSON for better error handling
func BindJSONWithValidation(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindJSON(obj); err != nil {
		// Parse validation errors
		var validationErrors []models.ValidationError

		// Handle JSON syntax errors
		if err.Error() == "unexpected EOF" || err.Error() == "invalid character" {
			validationErrors = append(validationErrors, models.ValidationError{
				Field:   "body",
				Message: "Invalid JSON format",
			})
		} else {
			// Try to extract field-specific validation errors
			// Gin's validator provides field-level errors
			if jsonErr, ok := err.(*json.UnmarshalTypeError); ok {
				validationErrors = append(validationErrors, models.ValidationError{
					Field:   jsonErr.Field,
					Message: fmt.Sprintf("Invalid type for field, expected %s", jsonErr.Type),
				})
			} else {
				// Generic validation error
				validationErrors = append(validationErrors, models.ValidationError{
					Field:   "request",
					Message: err.Error(),
				})
			}
		}

		ValidationErrorResponse(c, validationErrors)
		return err
	}
	return nil
}
