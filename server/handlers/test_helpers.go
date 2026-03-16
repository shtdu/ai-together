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

package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
)

// Helper function to make JSON requests in tests
func MakeJSONRequest(handler http.Handler, method, url string, body interface{}, headers map[string]string) *httptest.ResponseRecorder {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			panic(err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req, _ := http.NewRequest(method, url, reqBody)
	req.Header.Set("Content-Type", "application/json")

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	return w
}

// Helper function to parse JSON response in tests
func ParseJSONResponse(w *httptest.ResponseRecorder, v interface{}) error {
	body, err := io.ReadAll(w.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}

// Helper function to set up test gin router
func SetupTestRouter(handlers ...gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/test", handlers...)
	return router
}

// Helper function to create a test context
func CreateTestContext() *gin.Context {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c
}
