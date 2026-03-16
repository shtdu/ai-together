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


package integration

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUnwrapJSONResponse_StandardError(t *testing.T) {
	// Create a test server that returns a standard error response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{
			"error": "Invalid request payload",
			"code": "VALIDATION_ERROR",
			"details": "Field 'name' is required",
			"request": "req-123"
		}`))
	}))
	defer server.Close()

	// Make a request
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Unwrap the error
	apiErr := UnwrapJSONResponse(resp)
	if apiErr == nil {
		t.Fatal("Expected error, got nil")
	}

	// Check error type
	errTyped, ok := apiErr.(*APIError)
	if !ok {
		t.Fatalf("Expected *APIError, got %T", apiErr)
	}

	// Verify fields
	if errTyped.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", errTyped.StatusCode)
	}
	if errTyped.ErrorCode != "VALIDATION_ERROR" {
		t.Errorf("Expected code 'VALIDATION_ERROR', got '%s'", errTyped.ErrorCode)
	}
	if errTyped.Message != "Invalid request payload" {
		t.Errorf("Expected message 'Invalid request payload', got '%s'", errTyped.Message)
	}
	if errTyped.RequestID != "req-123" {
		t.Errorf("Expected request ID 'req-123', got '%s'", errTyped.RequestID)
	}
	if errTyped.Details["message"] != "Field 'name' is required" {
		t.Errorf("Expected details 'Field 'name' is required', got '%v'", errTyped.Details)
	}
}

func TestUnwrapJSONResponse_ValidationError(t *testing.T) {
	// Create a test server that returns a validation error response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{
			"error": "Validation failed",
			"code": "VALIDATION_ERROR",
			"errors": [
				{"field": "email", "message": "Invalid email format"},
				{"field": "password", "message": "Password too short"}
			]
		}`))
	}))
	defer server.Close()

	// Make a request
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Unwrap the error
	apiErr := UnwrapJSONResponse(resp)
	if apiErr == nil {
		t.Fatal("Expected error, got nil")
	}

	// Check error type
	errTyped, ok := apiErr.(*APIError)
	if !ok {
		t.Fatalf("Expected *APIError, got %T", apiErr)
	}

	// Verify fields
	if errTyped.Message != "Validation failed" {
		t.Errorf("Expected message 'Validation failed', got '%s'", errTyped.Message)
	}
	if len(errTyped.ValidationErrors) != 2 {
		t.Fatalf("Expected 2 validation errors, got %d", len(errTyped.ValidationErrors))
	}

	// Check first validation error
	if errTyped.ValidationErrors[0].Field != "email" {
		t.Errorf("Expected field 'email', got '%s'", errTyped.ValidationErrors[0].Field)
	}
	if errTyped.ValidationErrors[0].Message != "Invalid email format" {
		t.Errorf("Expected message 'Invalid email format', got '%s'", errTyped.ValidationErrors[0].Message)
	}

	// Check second validation error
	if errTyped.ValidationErrors[1].Field != "password" {
		t.Errorf("Expected field 'password', got '%s'", errTyped.ValidationErrors[1].Field)
	}
}

func TestUnwrapJSONResponse_EmptyBody(t *testing.T) {
	// Create a test server that returns an error with no body
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// Make a request
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Unwrap the error
	apiErr := UnwrapJSONResponse(resp)
	if apiErr == nil {
		t.Fatal("Expected error, got nil")
	}

	// Check error type
	errTyped, ok := apiErr.(*APIError)
	if !ok {
		t.Fatalf("Expected *APIError, got %T", apiErr)
	}

	// Verify fields
	if errTyped.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", errTyped.StatusCode)
	}
	if errTyped.Message == "" {
		t.Error("Expected non-empty message")
	}
}

func TestUnwrapJSONResponse_NonJSON(t *testing.T) {
	// Create a test server that returns plain text
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Resource not found"))
	}))
	defer server.Close()

	// Make a request
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Unwrap the error
	apiErr := UnwrapJSONResponse(resp)
	if apiErr == nil {
		t.Fatal("Expected error, got nil")
	}

	// Check error type
	errTyped, ok := apiErr.(*APIError)
	if !ok {
		t.Fatalf("Expected *APIError, got %T", apiErr)
	}

	// Verify the raw body is used as message
	if errTyped.Message != "Resource not found" {
		t.Errorf("Expected message 'Resource not found', got '%s'", errTyped.Message)
	}
}

func TestUnwrapJSONResponse_Success(t *testing.T) {
	// Create a test server that returns success
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	// Make a request
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Unwrap should return nil for successful responses
	apiErr := UnwrapJSONResponse(resp)
	if apiErr != nil {
		t.Errorf("Expected nil error for successful response, got %v", apiErr)
	}
}

func TestUnwrapJSONResponse_NilResponse(t *testing.T) {
	// Unwrap should handle nil response
	apiErr := UnwrapJSONResponse(nil)
	if apiErr == nil {
		t.Fatal("Expected error for nil response, got nil")
	}
}

func TestIsAPIError(t *testing.T) {
	apiErr := &APIError{StatusCode: 400, Message: "Bad request"}
	stdErr := fmt.Errorf("standard error")

	if !IsAPIError(apiErr) {
		t.Error("Expected IsAPIError to return true for APIError")
	}
	if IsAPIError(stdErr) {
		t.Error("Expected IsAPIError to return false for standard error")
	}
}

func TestGetAPIError(t *testing.T) {
	apiErr := &APIError{StatusCode: 400, Message: "Bad request"}
	stdErr := fmt.Errorf("standard error")

	if GetAPIError(apiErr) != apiErr {
		t.Error("Expected GetAPIError to return the same APIError")
	}
	if GetAPIError(stdErr) != nil {
		t.Error("Expected GetAPIError to return nil for standard error")
	}
}
