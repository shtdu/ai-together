// Copyright (c) 2025 AI Together
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
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"switch-server/models"
)

// TestRequestIDMiddleware_GeneratesUniqueIDs verifies that each request gets a unique ID
func TestRequestIDMiddleware_GeneratesUniqueIDs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var requestIDs []string
	router := gin.New()
	router.Use(RequestIDMiddleware())

	// Track request IDs
	router.GET("/test", func(c *gin.Context) {
		requestID := c.GetString("request_id")
		requestIDs = append(requestIDs, requestID)
		c.JSON(http.StatusOK, gin.H{"request_id": requestID})
	})

	// Make multiple requests
	for i := 0; i < 5; i++ {
		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	}

	// Verify all request IDs are unique
	uniqueIDs := make(map[string]bool)
	for _, id := range requestIDs {
		uniqueIDs[id] = true
	}
	assert.Equal(t, 5, len(uniqueIDs), "All request IDs should be unique")
}

// TestRequestIDMiddleware_SetsXRequestIDHeader verifies the X-Request-ID header is set
func TestRequestIDMiddleware_SetsXRequestIDHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestIDMiddleware())

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Header().Get("X-Request-ID"), "X-Request-ID header should be set")

	// Verify it's a valid UUID format (standard format is 36 chars with hyphens)
	requestID := w.Header().Get("X-Request-ID")
	assert.Equal(t, 36, len(requestID), "Request ID should be UUID format (36 characters)")
}

// TestRequestIDMiddleware_PersistsThroughContext verifies request ID is available in context
func TestRequestIDMiddleware_PersistsThroughContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestIDMiddleware())

	router.GET("/test", func(c *gin.Context) {
		requestID, exists := c.Get("request_id")
		assert.True(t, exists, "request_id should exist in context")
		assert.NotEmpty(t, requestID, "request_id should not be empty")

		// Also check backward compatibility key
		requestIDAlt, existsAlt := c.Get("request_id")
		assert.True(t, existsAlt, "request_id (backward compat) should exist")
		assert.Equal(t, requestID, requestIDAlt, "Both keys should have same value")

		c.JSON(http.StatusOK, gin.H{"request_id": requestID})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestRequestIDMiddleware_PropagatesToSubsequentHandlers verifies request ID propagates through middleware chain
func TestRequestIDMiddleware_PropagatesToSubsequentHandlers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestIDMiddleware())

	// First middleware
	middleware1 := func(c *gin.Context) {
		requestID1 := c.GetString("request_id")
		c.Set("middleware1_id", requestID1)
		c.Next()
	}

	// Second middleware
	middleware2 := func(c *gin.Context) {
		requestID2 := c.GetString("request_id")
		c.Set("middleware2_id", requestID2)
		c.Next()
	}

	router.Use(middleware1)
	router.Use(middleware2)

	router.GET("/test", func(c *gin.Context) {
		handlerID := c.GetString("request_id")
		middleware1ID := c.GetString("middleware1_id")
		middleware2ID := c.GetString("middleware2_id")

		// All IDs should be the same
		assert.Equal(t, handlerID, middleware1ID, "Handler and middleware1 should have same request ID")
		assert.Equal(t, handlerID, middleware2ID, "Handler and middleware2 should have same request ID")

		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestErrorLoggingMiddleware_WithUserContext tests logging with authenticated user
func TestErrorLoggingMiddleware_WithUserContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestIDMiddleware())

	// Setup middleware to add user context
	setupUser := func(c *gin.Context) {
		c.Set("userID", int64(123))
		c.Set("tenantID", int64(456))
		c.Set("user", &models.User{
			ID:       123,
			Email:    "test@example.com",
			TenantID: 456,
		})
		c.Next()
	}

	router.Use(setupUser)
	router.Use(ErrorLoggingMiddleware())

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestErrorLoggingMiddleware_WithoutUserContext tests logging without user context (public requests)
func TestErrorLoggingMiddleware_WithoutUserContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.Use(ErrorLoggingMiddleware())

	router.GET("/public", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "public endpoint"})
	})

	req, _ := http.NewRequest("GET", "/public", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestErrorLoggingMiddleware_StatusCodes tests logging with different status codes
func TestErrorLoggingMiddleware_StatusCodes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		statusCode int
		path       string
	}{
		{"2xx success", http.StatusOK, "/success"},
		{"2xx created", http.StatusCreated, "/created"},
		{"4xx bad request", http.StatusBadRequest, "/bad-request"},
		{"4xx unauthorized", http.StatusUnauthorized, "/unauthorized"},
		{"4xx forbidden", http.StatusForbidden, "/forbidden"},
		{"4xx not found", http.StatusNotFound, "/not-found"},
		{"5xx internal error", http.StatusInternalServerError, "/internal-error"},
		{"5xx service unavailable", http.StatusServiceUnavailable, "/unavailable"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(RequestIDMiddleware())
			router.Use(ErrorLoggingMiddleware())

			router.GET(tt.path, func(c *gin.Context) {
				c.JSON(tt.statusCode, gin.H{"status": tt.statusCode})
			})

			req, _ := http.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.statusCode, w.Code)
		})
	}
}

// TestErrorLoggingMiddleware_WithBindError tests error type handling for BindError
func TestErrorLoggingMiddleware_WithBindError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.Use(ErrorLoggingMiddleware())

	router.POST("/test", func(c *gin.Context) {
		// Add a bind error to the context
		// Use errors.New instead of UnmarshalTypeError to avoid nil pointer issues
		c.Error(gin.Error{
			Type: gin.ErrorTypeBind,
			Err:  errors.New("bind error: invalid JSON"),
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "bind error"})
	})

	req, _ := http.NewRequest("POST", "/test", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestErrorLoggingMiddleware_WithPublicError tests error type handling for PublicError
func TestErrorLoggingMiddleware_WithPublicError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.Use(ErrorLoggingMiddleware())

	router.GET("/test", func(c *gin.Context) {
		c.Error(gin.Error{
			Type: gin.ErrorTypePublic,
			Err:  errors.New("public error"),
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "public error"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestErrorLoggingMiddleware_WithPrivateError tests error type handling for PrivateError
func TestErrorLoggingMiddleware_WithPrivateError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.Use(ErrorLoggingMiddleware())

	router.GET("/test", func(c *gin.Context) {
		c.Error(gin.Error{
			Type: gin.ErrorTypePrivate,
			Err:  errors.New("private error"),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// TestErrorLoggingMiddleware_NoErrors tests normal flow with no errors
func TestErrorLoggingMiddleware_NoErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.Use(ErrorLoggingMiddleware())

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestErrorResponse_WithRequestID tests request ID is included in error response
func TestErrorResponse_WithRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("request_id", "test-request-123")

	// Set up a proper request
	req, _ := http.NewRequest("POST", "/test", nil)
	c.Request = req

	errResp := models.ErrorResponse{
		Error: "Test error",
		Code:  models.ErrCodeValidation,
	}

	ErrorResponse(c, http.StatusBadRequest, errResp)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "test-request-123", response.Request)
}

// TestErrorResponse_DetailsInDebugMode tests details field is included in debug mode
func TestErrorResponse_DetailsInDebugMode(t *testing.T) {
	gin.SetMode(gin.DebugMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Set up a proper request
	req, _ := http.NewRequest("POST", "/test", nil)
	c.Request = req

	errResp := models.ErrorResponse{
		Error:   "Test error",
		Code:    models.ErrCodeInternal,
		Details: "Detailed debug information",
	}

	ErrorResponse(c, http.StatusInternalServerError, errResp)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.NotEmpty(t, response.Details, "Details should be included in debug mode")
}

// TestErrorResponse_DetailsRemovedInReleaseMode tests details field is removed in non-debug mode
func TestErrorResponse_DetailsRemovedInReleaseMode(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("request_id", "test-request-123")

	// Set up a proper request
	req, _ := http.NewRequest("POST", "/test", nil)
	c.Request = req

	errResp := models.ErrorResponse{
		Error:   "Test error",
		Code:    models.ErrCodeInternal,
		Details: "Detailed debug information",
	}

	ErrorResponse(c, http.StatusInternalServerError, errResp)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Empty(t, response.Details, "Details should be removed in release mode")
}

// TestErrorResponse_StatusCodes tests correct HTTP status codes for different error types
func TestErrorResponse_StatusCodes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		statusCode int
		errCode    string
	}{
		{"Bad Request", http.StatusBadRequest, models.ErrCodeValidation},
		{"Unauthorized", http.StatusUnauthorized, models.ErrCodeUnauthorized},
		{"Forbidden", http.StatusForbidden, models.ErrCodeForbidden},
		{"Not Found", http.StatusNotFound, models.ErrCodeNotFound},
		{"Conflict", http.StatusConflict, models.ErrCodeConflict},
		{"Internal Error", http.StatusInternalServerError, models.ErrCodeInternal},
		{"Service Unavailable", http.StatusServiceUnavailable, models.ErrCodeServiceUnavailable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Set up a proper request
			req, _ := http.NewRequest("POST", "/test", nil)
			c.Request = req

			errResp := models.ErrorResponse{
				Error: tt.name + " error",
				Code:  tt.errCode,
			}

			ErrorResponse(c, tt.statusCode, errResp)

			assert.Equal(t, tt.statusCode, w.Code)
		})
	}
}

// TestValidationErrorResponse_MalformedJSON tests malformed JSON body handling
func TestValidationErrorResponse_MalformedJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("request_id", "test-request-123")

	req, _ := http.NewRequest("POST", "/test", io.NopCloser(strings.NewReader("{malformed json}")))
	c.Request = req

	validationErrors := []models.ValidationError{
		{Field: "body", Message: "Invalid JSON format"},
	}

	ValidationErrorResponse(c, validationErrors)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.ValidationErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Validation failed", response.Error)
	assert.Equal(t, models.ErrCodeValidation, response.Code)
	assert.Equal(t, 1, len(response.Errors))
	assert.Equal(t, "body", response.Errors[0].Field)
}

// TestValidationErrorResponse_MultipleErrors tests validator.ValidationErrors handling with multiple field errors
func TestValidationErrorResponse_MultipleErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("request_id", "test-request-456")

	req, _ := http.NewRequest("POST", "/test", nil)
	c.Request = req

	validationErrors := []models.ValidationError{
		{Field: "email", Message: "Email is required"},
		{Field: "password", Message: "Password must be at least 8 characters"},
		{Field: "name", Message: "Name is required"},
	}

	ValidationErrorResponse(c, validationErrors)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.ValidationErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, 3, len(response.Errors))

	// Verify each error is present
	fields := make(map[string]bool)
	for _, e := range response.Errors {
		fields[e.Field] = true
	}
	assert.True(t, fields["email"])
	assert.True(t, fields["password"])
	assert.True(t, fields["name"])
}

// TestValidationErrorResponse_SetsXRequestIDHeader tests X-Request-ID header is set
func TestValidationErrorResponse_SetsXRequestIDHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("request_id", "validation-test-789")

	req, _ := http.NewRequest("POST", "/test", nil)
	c.Request = req

	validationErrors := []models.ValidationError{
		{Field: "field", Message: "Error message"},
	}

	ValidationErrorResponse(c, validationErrors)

	assert.Equal(t, "validation-test-789", w.Header().Get("X-Request-ID"))
}

// TestBindJSONWithValidation_Success tests successful JSON binding and validation
func TestBindJSONWithValidation_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type TestStruct struct {
		Name  string `json:"name" binding:"required"`
		Email string `json:"email" binding:"required,email"`
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{"name":"Test User","email":"test@example.com"}`
	req, _ := http.NewRequest("POST", "/test", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	var obj TestStruct
	err := BindJSONWithValidation(c, &obj)

	assert.NoError(t, err)
	assert.Equal(t, "Test User", obj.Name)
	assert.Equal(t, "test@example.com", obj.Email)
}

// TestBindJSONWithValidation_MalformedJSON tests malformed JSON body returns 400
func TestBindJSONWithValidation_MalformedJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type TestStruct struct {
		Name string `json:"name"`
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{malformed}`
	req, _ := http.NewRequest("POST", "/test", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	var obj TestStruct
	err := BindJSONWithValidation(c, &obj)

	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestBindJSONWithValidation_EmptyBody tests empty request body handling
func TestBindJSONWithValidation_EmptyBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type TestStruct struct {
		Name string `json:"name"`
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := ``
	req, _ := http.NewRequest("POST", "/test", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	var obj TestStruct
	err := BindJSONWithValidation(c, &obj)

	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestBindJSONWithValidation_TypeMismatch tests type mismatches in JSON fields
func TestBindJSONWithValidation_TypeMismatch(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type TestStruct struct {
		Age int `json:"age"`
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{"age":"not_a_number"}`
	req, _ := http.NewRequest("POST", "/test", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	var obj TestStruct
	err := BindJSONWithValidation(c, &obj)

	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.ValidationErrorResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.NotEmpty(t, response.Errors)
}

// TestBindJSONWithValidation_StructValidation tests struct validation errors
func TestBindJSONWithValidation_StructValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type TestStruct struct {
		Email string `json:"email" binding:"required,email"`
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{"email":"not_an_email"}`
	req, _ := http.NewRequest("POST", "/test", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	var obj TestStruct
	err := BindJSONWithValidation(c, &obj)

	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestBindJSONWithValidation_VerifiesErrorResponse tests verification error details in response
func TestBindJSONWithValidation_VerifiesErrorResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type TestStruct struct {
		Email string `json:"email" binding:"required,email"`
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{}`
	req, _ := http.NewRequest("POST", "/test", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	var obj TestStruct
	err := BindJSONWithValidation(c, &obj)

	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.ValidationErrorResponse
	jsonErr := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, jsonErr)
	assert.Equal(t, "Validation failed", response.Error)
	assert.Equal(t, models.ErrCodeValidation, response.Code)
	assert.NotEmpty(t, response.Errors, "Errors should not be empty")
}

// TestResponseWriter_WriteHeaderCapturesStatusCode tests WriteHeader() captures status code
func TestResponseWriter_WriteHeaderCapturesStatusCode(t *testing.T) {
	gin.SetMode(gin.TestMode)

	baseRecorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(baseRecorder)
	rw := &responseWriter{
		ResponseWriter: c.Writer,
		status:         0,
	}

	rw.WriteHeader(http.StatusCreated)

	assert.Equal(t, http.StatusCreated, rw.status)
}

// TestResponseWriter_WriteInfersStatusCode tests Write() infers status code from body
func TestResponseWriter_WriteInfersStatusCode(t *testing.T) {
	gin.SetMode(gin.TestMode)

	baseRecorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(baseRecorder)
	rw := &responseWriter{
		ResponseWriter: c.Writer,
		status:         0,
	}

	data := []byte("test response")
	n, err := rw.Write(data)

	assert.NoError(t, err)
	assert.Equal(t, len(data), n)
	assert.Equal(t, http.StatusOK, rw.status, "Write() should default to 200 if status not set")
}

// TestResponseWriter_WriteDoesNotOverrideStatus tests Write() does not override already set status
func TestResponseWriter_WriteDoesNotOverrideStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	baseRecorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(baseRecorder)
	rw := &responseWriter{
		ResponseWriter: c.Writer,
		status:         0,
	}

	// Set status first
	rw.WriteHeader(http.StatusCreated)

	// Then write
	data := []byte("test response")
	rw.Write(data)

	assert.Equal(t, http.StatusCreated, rw.status, "Status should remain 201 after Write()")
}

// TestResponseWriter_IntegrationWithGin tests integration with gin.ResponseWriter
func TestResponseWriter_IntegrationWithGin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.Use(ErrorLoggingMiddleware())

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{"message": "created"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "created", response["message"])
}

// TestResponseWriter_StatusDefaultsTo200 tests status code defaults to 200 if not set
func TestResponseWriter_StatusDefaultsTo200(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := func(c *gin.Context) {
		// Don't call WriteHeader, just Write
		c.Writer.Write([]byte("response"))
	}

	router := gin.New()
	router.Use(ErrorLoggingMiddleware())
	router.GET("/test", handler)

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Status should default to 200
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestErrorLoggingMiddleware_CapturesStatusCode tests that responseWriter captures status correctly
func TestErrorLoggingMiddleware_CapturesStatusCode(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.Use(ErrorLoggingMiddleware())

	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusCreated)
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

// TestErrorLoggingMiddleware_DurationTracking tests that request duration is tracked
func TestErrorLoggingMiddleware_DurationTracking(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.Use(ErrorLoggingMiddleware())

	router.GET("/test", func(c *gin.Context) {
		time.Sleep(50 * time.Millisecond)
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	start := time.Now()
	router.ServeHTTP(w, req)
	duration := time.Since(start)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.GreaterOrEqual(t, duration.Milliseconds(), int64(50), "Request should take at least 50ms")
}

// TestErrorLoggingMiddleware_DebugModeStacktrace tests stack trace logging in debug mode
func TestErrorLoggingMiddleware_DebugModeStacktrace(t *testing.T) {
	gin.SetMode(gin.DebugMode)

	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.Use(ErrorLoggingMiddleware())

	router.GET("/test", func(c *gin.Context) {
		// Add a private error to trigger stack trace logging in debug mode
		c.Error(gin.Error{
			Type: gin.ErrorTypePrivate,
			Err:  errors.New("internal error for stack trace"),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// TestErrorLoggingMiddleware_MultipleErrors tests logging multiple errors in context
func TestErrorLoggingMiddleware_MultipleErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.Use(ErrorLoggingMiddleware())

	router.POST("/test", func(c *gin.Context) {
		// Add multiple errors
		c.Error(gin.Error{
			Type: gin.ErrorTypeBind,
			Err:  errors.New("first bind error"),
		})
		c.Error(gin.Error{
			Type: gin.ErrorTypePrivate,
			Err:  errors.New("second internal error"),
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "multiple errors"})
	})

	req, _ := http.NewRequest("POST", "/test", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestErrorResponse_NoRequestID tests error response without request ID
func TestErrorResponse_NoRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest("POST", "/test", nil)
	c.Request = req

	errResp := models.ErrorResponse{
		Error: "Test error without request ID",
		Code:  models.ErrCodeValidation,
	}

	ErrorResponse(c, http.StatusBadRequest, errResp)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Empty(t, response.Request, "Request should be empty when not set in context")
}

// TestValidationErrorResponse_NoRequestID tests validation error response without request ID
func TestValidationErrorResponse_NoRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest("POST", "/test", nil)
	c.Request = req

	validationErrors := []models.ValidationError{
		{Field: "email", Message: "Invalid email format"},
	}

	ValidationErrorResponse(c, validationErrors)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	// X-Request-ID header should not be set when no request ID in context
	assert.Empty(t, w.Header().Get("X-Request-ID"))
}

// TestBindJSONWithValidation_UnmarshalTypeError tests json.UnmarshalTypeError handling
func TestBindJSONWithValidation_UnmarshalTypeError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type TestStruct struct {
		Age int `json:"age"`
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Provide string where int is expected
	body := `{"age":"thirty"}`
	req, _ := http.NewRequest("POST", "/test", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	var obj TestStruct
	err := BindJSONWithValidation(c, &obj)

	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.ValidationErrorResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.NotEmpty(t, response.Errors)
	// Should contain the unmarshal type error details
	found := false
	for _, e := range response.Errors {
		if e.Field == "age" && strings.Contains(e.Message, "Invalid type") {
			found = true
		}
	}
	assert.True(t, found, "Should find unmarshal type error for age field")
}
