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
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client wraps http.Client with provider-specific defaults
type Client struct {
	*http.Client
}

// NewClient creates a new HTTP client with a long timeout for streaming
func NewClient() *Client {
	return &Client{
		Client: &http.Client{
			Timeout: 3000 * time.Second,
		},
	}
}

// NewClientWithTimeout creates a new HTTP client with custom timeout
func NewClientWithTimeout(timeout time.Duration) *Client {
	return &Client{
		Client: &http.Client{
			Timeout: timeout,
		},
	}
}

// PostWithHeaders sends a POST request with custom headers, query parameters, and body
func (c *Client) PostWithHeaders(url string, headers map[string]string, queryParams map[string]string, body []byte) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// Set query params
	if len(queryParams) > 0 {
		q := req.URL.Query()
		for k, v := range queryParams {
			q.Set(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	return resp, nil
}

// CopyResponse copies the response body to the provided writer and returns any error
func CopyResponse(resp *http.Response, w io.Writer) error {
	return CopyResponseWithHook(resp, w, nil)
}

// CopyResponseWithHook copies the response body to the provided writer with an optional hook function
// The hook function is called for each chunk of data read from the response body.
// Hook signature: func(data []byte) (bool, []byte) - returns (shouldContinue, modifiedData)
func CopyResponseWithHook(resp *http.Response, w io.Writer, hook func([]byte) (bool, []byte)) error {
	if resp == nil {
		return fmt.Errorf("response is nil")
	}
	defer resp.Body.Close()

	// Copy headers
	if hw, ok := w.(http.ResponseWriter); ok {
		for k, v := range resp.Header {
			for _, val := range v {
				hw.Header().Add(k, val)
			}
		}
		hw.WriteHeader(resp.StatusCode)
	}

	// Copy body with optional hook
	if hook != nil {
		buf := make([]byte, 32*1024) // 32KB buffer
		for {
			n, err := resp.Body.Read(buf)
			if n > 0 {
				data := buf[:n]
				shouldContinue, modifiedData := hook(data)
				if _, writeErr := w.Write(modifiedData); writeErr != nil {
					return writeErr
				}
				if !shouldContinue {
					return nil
				}
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}
		}
		return nil
	}

	// Copy body without hook (standard path)
	_, err := io.Copy(w, resp.Body)
	return err
}

// IsSuccess returns true if status code is 2xx
func IsSuccess(statusCode int) bool {
	return statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices
}

// IsClientError returns true if status code is 4xx
func IsClientError(statusCode int) bool {
	return statusCode >= http.StatusBadRequest && statusCode < http.StatusInternalServerError
}

// IsServerError returns true if status code is 5xx
func IsServerError(statusCode int) bool {
	return statusCode >= http.StatusInternalServerError
}

// IsRedirect returns true if status code is 3xx
func IsRedirect(statusCode int) bool {
	return statusCode >= http.StatusMultipleChoices && statusCode < http.StatusBadRequest
}
