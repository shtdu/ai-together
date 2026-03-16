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
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestLoggingTransport(t *testing.T) {
	// Create a test server that returns a simple response
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer testServer.Close()

	// Create logging transport
	logger := slog.Default()
	transport := NewLoggingTransport(http.DefaultTransport, logger)

	client := &http.Client{
		Transport: transport,
		Timeout:   5 * time.Second,
	}

	// Make a request
	resp, err := client.Get(testServer.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestAuthTransport(t *testing.T) {
	// Create a test server that checks for the Authorization header
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if !strings.HasPrefix(auth, "Bearer ") {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"authorized"}`))
	}))
	defer testServer.Close()

	// Create auth transport with token getter
	getToken := func() (string, error) {
		return "test-token-123", nil
	}

	transport := NewAuthTransport(http.DefaultTransport, getToken)

	client := &http.Client{
		Transport: transport,
		Timeout:   5 * time.Second,
	}

	// Make a request
	resp, err := client.Get(testServer.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestAuthTransport_NoToken(t *testing.T) {
	// Create a test server that doesn't require auth
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer testServer.Close()

	// Create auth transport with nil token getter (should skip auth)
	transport := NewAuthTransport(http.DefaultTransport, nil)

	client := &http.Client{
		Transport: transport,
		Timeout:   5 * time.Second,
	}

	// Make a request
	resp, err := client.Get(testServer.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestRetryTransport_Success(t *testing.T) {
	// Create a test server that always returns 200
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer testServer.Close()

	logger := slog.Default()
	transport := NewRetryTransport(http.DefaultTransport, logger)

	client := &http.Client{
		Transport: transport,
		Timeout:   5 * time.Second,
	}

	// Make a request
	resp, err := client.Get(testServer.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestRetryTransport_RetryOn500(t *testing.T) {
	// Create a test server that returns 500 twice, then 200
	attempts := 0
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts <= 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer testServer.Close()

	logger := slog.Default()
	retryTransport := NewRetryTransport(http.DefaultTransport, logger)
	retryTransport.maxRetries = 3 // Allow up to 3 retries

	client := &http.Client{
		Transport: retryTransport,
		Timeout:   5 * time.Second,
	}

	// Make a request
	resp, err := client.Get(testServer.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200 after retries, got %d", resp.StatusCode)
	}

	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestTransportChain(t *testing.T) {
	// Create a test server that checks for auth and logs
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer testServer.Close()

	logger := slog.Default()
	getToken := func() (string, error) {
		return "test-token", nil
	}

	// Create full transport chain
	transport := NewTransportChain(getToken, logger, false) // Minimal logging for tests

	client := &http.Client{
		Transport: transport,
		Timeout:   5 * time.Second,
	}

	// Make a request
	resp, err := client.Get(testServer.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestNewAnonymousClient(t *testing.T) {
	// Create a test server
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer testServer.Close()

	logger := slog.Default()

	// Create anonymous client
	client, err := NewAnonymousClient(testServer.URL, logger, false) // Minimal logging for tests
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Make a request
	ctx := context.Background()
	resp, err := client.GetHealthWithResponse(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StatusCode() != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode())
	}
}

func TestNewAuthenticatedClient(t *testing.T) {
	// Create a test server that checks auth
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer testServer.Close()

	logger := slog.Default()
	getToken := func() (string, error) {
		return "test-token", nil
	}

	// Create authenticated client
	client, err := NewAuthenticatedClient(testServer.URL, getToken, logger, false) // Minimal logging for tests
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Make a request
	ctx := context.Background()
	resp, err := client.GetHealthWithResponse(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StatusCode() != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode())
	}
}

func TestPrettifyJSON_ValidJSONObject(t *testing.T) {
	transport := NewLoggingTransport(nil, slog.Default())

	input := []byte(`{"name":"test","value":123}`)
	output := transport.prettifyJSON(input)

	expected := `{
  "name": "test",
  "value": 123
}`
	if output != expected {
		t.Errorf("expected prettified JSON:\n%s\ngot:\n%s", expected, output)
	}
}

func TestPrettifyJSON_ValidJSONArray(t *testing.T) {
	transport := NewLoggingTransport(nil, slog.Default())

	input := []byte(`[{"id":1},{"id":2}]`)
	output := transport.prettifyJSON(input)

	expected := `[
  {
    "id": 1
  },
  {
    "id": 2
  }
]`
	if output != expected {
		t.Errorf("expected prettified JSON:\n%s\ngot:\n%s", expected, output)
	}
}

func TestPrettifyJSON_NonJSON(t *testing.T) {
	transport := NewLoggingTransport(nil, slog.Default())

	input := []byte(`plain text response`)
	output := transport.prettifyJSON(input)

	expected := `plain text response`
	if output != expected {
		t.Errorf("expected original string, got: %s", output)
	}
}

func TestPrettifyJSON_Empty(t *testing.T) {
	transport := NewLoggingTransport(nil, slog.Default())

	input := []byte{}
	output := transport.prettifyJSON(input)

	if output != "" {
		t.Errorf("expected empty string, got: %s", output)
	}
}

func TestPrettifyJSON_InvalidJSON(t *testing.T) {
	transport := NewLoggingTransport(nil, slog.Default())

	input := []byte(`{invalid json}`)
	output := transport.prettifyJSON(input)

	expected := `{invalid json}`
	if output != expected {
		t.Errorf("expected original string for invalid JSON, got: %s", output)
	}
}
