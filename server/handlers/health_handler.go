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
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// HealthChecker defines the interface for checking database health
type HealthChecker interface {
	Ping(ctx context.Context) error
}

// pgxPoolWrapper wraps pgxpool.Pool to implement HealthChecker
type pgxPoolWrapper struct {
	pool *pgxpool.Pool
}

func (w *pgxPoolWrapper) Ping(ctx context.Context) error {
	return w.pool.Ping(ctx)
}

type HealthHandler struct {
	db      HealthChecker
	version string
}

func NewHealthHandler(db *pgxpool.Pool, version string) *HealthHandler {
	var checker HealthChecker
	if db != nil {
		checker = &pgxPoolWrapper{pool: db}
	}
	return &HealthHandler{
		db:      checker,
		version: version,
	}
}

// newHealthHandlerWithChecker creates a HealthHandler with a custom HealthChecker (for testing)
func newHealthHandlerWithChecker(checker HealthChecker, version string) *HealthHandler {
	return &HealthHandler{
		db:      checker,
		version: version,
	}
}

type HealthResponse struct {
	Version   string         `json:"version"`
	Status    string         `json:"status"`
	Timestamp string         `json:"timestamp"`
	Database  DatabaseStatus `json:"database"`
}

type DatabaseStatus struct {
	Status  string `json:"status"`
	Latency string `json:"latency,omitempty"`
	Error   string `json:"error,omitempty"`
}

func (h *HealthHandler) HealthCheck(c *gin.Context) {
	response := HealthResponse{
		Version:   h.version,
		Status:    "ok",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	start := time.Now()
	err := h.db.Ping(ctx)
	latency := time.Since(start)

	if err != nil {
		response.Status = "degraded"
		response.Database = DatabaseStatus{
			Status: "down",
			Error:  err.Error(),
		}
		c.JSON(http.StatusServiceUnavailable, response)
		return
	}

	response.Database = DatabaseStatus{
		Status:  "up",
		Latency: latency.String(),
	}

	c.JSON(http.StatusOK, response)
}
