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
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestLoggingTransport_VerboseFlag verifies that LoggingTransport can be
// configured with a verbose flag to control request/response body logging
func TestLoggingTransport_VerboseFlag(t *testing.T) {
	tests := []struct {
		name    string
		verbose bool
	}{
		{"verbose=false", false},
		{"verbose=true", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test server
			testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"status":"ok"}`))
			}))
			defer testServer.Close()

			// Create transport with verbose flag - this should NOT fail
			logger := slog.Default()
			base := http.DefaultTransport
			transport := NewLoggingTransportVerbose(base, logger, tt.verbose)

			client := &http.Client{
				Transport: transport,
			}

			// Make request - should succeed without errors
			resp, err := client.Get(testServer.URL)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			defer resp.Body.Close()

			// Verify response is readable
			_, err = io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("failed to read response body: %v", err)
			}
		})
	}
}
