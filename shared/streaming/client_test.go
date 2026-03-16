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


package streaming

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ==================== CopyResponse Tests ====================

func TestCopyResponse(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		headers    map[string]string
	}{
		{
			name:       "simple response",
			statusCode: http.StatusOK,
			body:       "Hello, World!",
			headers:    map[string]string{"Content-Type": "text/plain"},
		},
		{
			name:       "JSON response",
			statusCode: http.StatusOK,
			body:       `{"message": "success"}`,
			headers:    map[string]string{"Content-Type": "application/json"},
		},
		{
			name:       "error response",
			statusCode: http.StatusBadRequest,
			body:       `{"error": "invalid request"}`,
			headers:    map[string]string{"Content-Type": "application/json"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				for k, v := range tt.headers {
					w.Header().Set(k, v)
				}
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.body))
			}))
			defer server.Close()

			// Make request
			client := NewClient()
			resp, err := client.PostWithHeaders(server.URL, nil, nil, nil)
			if err != nil {
				t.Fatalf("GetWithHeaders failed: %v", err)
			}

			// Copy response
			var buf strings.Builder
			err = CopyResponse(resp, &buf)
			if err != nil {
				t.Fatalf("CopyResponse failed: %v", err)
			}

			// Verify body
			if buf.String() != tt.body {
				t.Errorf("CopyResponse() body = %q, want %q", buf.String(), tt.body)
			}
		})
	}
}

// ==================== CopyResponseWithHook Tests ====================

func TestCopyResponseWithHook(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		hook       func([]byte) (bool, []byte)
		expectStop bool
	}{
		{
			name: "no hook - just copy",
			body: "Hello, World!",
			hook: nil,
		},
		{
			name: "hook that modifies data",
			body: "hello",
			hook: func(data []byte) (bool, []byte) {
				modified := strings.ToUpper(string(data))
				return true, []byte(modified)
			},
		},
		{
			name: "hook that stops early",
			body: "Hello, World!",
			hook: func(data []byte) (bool, []byte) {
				return false, data // Stop after first chunk
			},
			expectStop: true,
		},
		{
			name: "hook that counts bytes",
			body: "Hello",
			hook: func(data []byte) (bool, []byte) {
				// Just pass through
				return true, data
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/plain")
				w.Write([]byte(tt.body))
			}))
			defer server.Close()

			// Make request
			client := NewClient()
			resp, err := client.PostWithHeaders(server.URL, nil, nil, nil)
			if err != nil {
				t.Fatalf("GetWithHeaders failed: %v", err)
			}

			// Copy response with hook
			var buf strings.Builder
			err = CopyResponseWithHook(resp, &buf, tt.hook)
			if err != nil {
				t.Fatalf("CopyResponseWithHook failed: %v", err)
			}

			// Verify result based on hook type
			if tt.hook == nil {
				if buf.String() != tt.body {
					t.Errorf("CopyResponseWithHook() body = %q, want %q", buf.String(), tt.body)
				}
			} else {
				// For modifying hook, check upper case
				if strings.Contains(tt.name, "modifies") {
					expected := strings.ToUpper(tt.body)
					if buf.String() != expected {
						t.Errorf("CopyResponseWithHook() body = %q, want %q", buf.String(), expected)
					}
				}
			}
		})
	}
}

// ==================== PostWithHeaders Tests ====================

func TestPostWithHeaders(t *testing.T) {
	tests := []struct {
		name       string
		headers    map[string]string
		query      map[string]string
		body       string
		expectBody string
		expectQuery string
	}{
		{
			name: "simple POST",
			headers: map[string]string{
				"Content-Type": "application/json",
			},
			query:       nil,
			body:        `{"test": "data"}`,
			expectBody:  `{"test": "data"}`,
			expectQuery: "",
		},
		{
			name: "POST with custom headers",
			headers: map[string]string{
				"Content-Type":  "application/json",
				"Authorization": "Bearer test-token",
			},
			query:       nil,
			body:        `{"test": "data"}`,
			expectBody:  `{"test": "data"}`,
			expectQuery: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server that echoes back
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Echo headers
				for k, v := range r.Header {
					w.Header().Set(k, strings.Join(v, ","))
				}

				// Check method
				if r.Method != "POST" {
					w.WriteHeader(http.StatusMethodNotAllowed)
					return
				}

				// Check headers
				for k, v := range tt.headers {
					if r.Header.Get(k) != v {
						w.WriteHeader(http.StatusBadRequest)
						w.Write([]byte("Header mismatch: " + k))
						return
					}
				}

				// Check query
				if tt.query != nil {
					if r.URL.RawQuery != tt.expectBody {
						w.Write([]byte(r.URL.RawQuery))
						return
					}
				}

				// Echo body
				w.Write([]byte(tt.expectBody))
			}))
			defer server.Close()

			// Make request
			client := NewClient()
			resp, err := client.PostWithHeaders(server.URL, tt.headers, tt.query, []byte(tt.body))
			if err != nil {
				t.Fatalf("PostWithHeaders failed: %v", err)
			}
			defer resp.Body.Close()

			// Read response
			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("ReadAll failed: %v", err)
			}

			// Verify body matches expected
			if string(bodyBytes) != tt.expectBody {
				t.Errorf("PostWithHeaders() body = %q, want %q", string(bodyBytes), tt.expectBody)
			}
		})
	}
}

// ==================== Status Code Helpers Tests ====================

func TestIsSuccess(t *testing.T) {
	tests := []struct {
		code   int
		expect bool
	}{
		{200, true},
		{201, true},
		{204, true},
		{299, true},
		{300, false},
		{400, false},
		{500, false},
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.code)), func(t *testing.T) {
			result := IsSuccess(tt.code)
			if result != tt.expect {
				t.Errorf("IsSuccess(%d) = %v, want %v", tt.code, result, tt.expect)
			}
		})
	}
}

func TestIsClientError(t *testing.T) {
	tests := []struct {
		code   int
		expect bool
	}{
		{200, false},
		{299, false},
		{400, true},
		{401, true},
		{404, true},
		{499, true},
		{500, false},
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.code)), func(t *testing.T) {
			result := IsClientError(tt.code)
			if result != tt.expect {
				t.Errorf("IsClientError(%d) = %v, want %v", tt.code, result, tt.expect)
			}
		})
	}
}

func TestIsServerError(t *testing.T) {
	tests := []struct {
		code   int
		expect bool
	}{
		{400, false},
		{499, false},
		{500, true},
		{502, true},
		{503, true},
		{599, true},
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.code)), func(t *testing.T) {
			result := IsServerError(tt.code)
			if result != tt.expect {
				t.Errorf("IsServerError(%d) = %v, want %v", tt.code, result, tt.expect)
			}
		})
	}
}

func TestIsRedirect(t *testing.T) {
	tests := []struct {
		code   int
		expect bool
	}{
		{200, false},
		{299, false},
		{300, true},
		{301, true},
		{302, true},
		{399, true},
		{400, false},
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.code)), func(t *testing.T) {
			result := IsRedirect(tt.code)
			if result != tt.expect {
				t.Errorf("IsRedirect(%d) = %v, want %v", tt.code, result, tt.expect)
			}
		})
	}
}

// ==================== NewClient Tests ====================

func TestNewClient(t *testing.T) {
	client := NewClient()
	if client == nil {
		t.Fatal("NewClient() returned nil")
	}
	if client.Client == nil {
		t.Fatal("NewClient() Client field is nil")
	}
}

func TestNewClientWithTimeout(t *testing.T) {
	client := NewClientWithTimeout(10)
	if client == nil {
		t.Fatal("NewClientWithTimeout() returned nil")
	}
	if client.Client == nil {
		t.Fatal("NewClientWithTimeout() Client field is nil")
	}
	_ = client.Timeout
}
