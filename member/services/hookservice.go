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

package services

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"

	"codeswitch/internal/hookdb"

	"github.com/code-together/shared/hook-common/storage"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

// writeRequest represents a database write request
type writeRequest struct {
	sessionID string
	toolName  string
	eventJSON []byte
	result    chan error
}

// HookService handles hook event collection and management
type HookService struct {
	initialized bool
	writeQueue  chan writeRequest
	wg          sync.WaitGroup
	stopCtx     context.Context
	cancel      context.CancelFunc
}

// NewHookService creates a new hook service and initializes the database
func NewHookService() *HookService {
	hs := &HookService{
		writeQueue: make(chan writeRequest, 100), // Bounded channel for backpressure
	}
	hs.stopCtx, hs.cancel = context.WithCancel(context.Background())

	// Initialize database (non-fatal)
	if err := hs.Init(); err != nil {
		slog.Warn("failed to initialize hook service database", "error", err)
		// Service continues but will return 503 for requests
	}

	// Start background writer goroutine
	hs.wg.Add(1)
	go hs.writeLoop()

	return hs
}

// writeLoop runs in a background goroutine, processing writes sequentially
func (hs *HookService) writeLoop() {
	defer hs.wg.Done()

	for {
		select {
		case <-hs.stopCtx.Done():
			// Drain remaining requests before shutting down
			slog.Info("hook service writer: draining queue before shutdown")
			for {
				select {
				case req := <-hs.writeQueue:
					_ = storage.StoreEvent(hookdb.HookDB, req.sessionID, req.toolName, req.eventJSON)
					// Try to notify sender, but don't block if they've already timed out
					select {
					case req.result <- nil:
						// Sender still waiting, notified successfully
					default:
						// No receiver - sender already timed out or abandoned request
					}
				default:
					// Queue is empty
					return
				}
			}

		case req := <-hs.writeQueue:
			// Process write request
			err := storage.StoreEvent(hookdb.HookDB, req.sessionID, req.toolName, req.eventJSON)
			req.result <- err
		}
	}
}

// Init initializes the hook events database
func (hs *HookService) Init() error {
	if err := hookdb.Init(); err != nil {
		hs.initialized = false
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	hs.initialized = true
	return nil
}

// Close closes the database connection and stops the writer goroutine
func (hs *HookService) Close() error {
	slog.Info("hook service: shutting down writer goroutine")

	// Stop accepting new requests
	hs.cancel()

	// Wait for writer goroutine to finish draining queue
	hs.wg.Wait()

	hs.initialized = false
	return hookdb.Close()
}

// IsInitialized returns true if the hook database is ready
func (hs *HookService) IsInitialized() bool {
	return hs.initialized && hookdb.HookDB != nil
}

// CollectEvent handles hook event collection via HTTP
func (hs *HookService) CollectEvent(c *gin.Context) {
	toolName := c.Param("tool_name")

	// Validate tool_name
	if toolName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tool_name parameter is required"})
		return
	}

	// Check database initialized
	if !hs.IsInitialized() {
		slog.Warn("hook database not initialized", "tool", toolName)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "hook events database not initialized",
		})
		return
	}

	// Read and validate event
	eventJSON, err := ValidateAndReadEvent(c.Request.Body)
	if err != nil {
		slog.Warn("invalid event data", "tool", toolName, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate required fields (session_id and hook_event_name)
	sessionID := gjson.GetBytes(eventJSON, "session_id").String()
	if sessionID == "" {
		slog.Warn("missing session_id in event", "tool", toolName)
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id is required"})
		return
	}

	eventName := gjson.GetBytes(eventJSON, "hook_event_name").String()
	if eventName == "" {
		slog.Warn("missing hook_event_name in event", "tool", toolName, "session", sessionID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "hook_event_name is required"})
		return
	}

	// Add timestamp
	eventWithTS, timestampErr := AddTimestamp(eventJSON)
	if timestampErr != nil {
		slog.Error("failed to add timestamp", "tool", toolName, "session", sessionID, "error", timestampErr)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process event"})
		return
	}

	// // Directly store the event without using the queue
	// err = storage.StoreEvent(hookdb.HookDB, sessionID, toolName, eventWithTS)
	// if err != nil {
	// 	slog.Error("failed to store event", "tool", toolName, "session", sessionID, "error", err)
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store event"})
	// 	return
	// }

	// Check queue depth for monitoring
	queueDepth := len(hs.writeQueue)
	if queueDepth > 80 { // 80% of capacity (100)
		slog.Warn("hook event queue is nearly full", "depth", queueDepth, "capacity", cap(hs.writeQueue))
	}

	// Queue the write request
	resultChan := make(chan error, 1)
	writeReq := writeRequest{
		sessionID: sessionID,
		toolName:  toolName,
		eventJSON: eventWithTS,
		result:    resultChan,
	}

	select {
	case hs.writeQueue <- writeReq:
		// Successfully queued
	default:
		// Queue is full! Return error immediately
		slog.Error("hook event queue is full", "tool", toolName, "session", sessionID)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "event queue is full, please retry",
		})
		return
	}

	// Wait for write to complete
	err = <-resultChan
	if err != nil {
		slog.Error("failed to store event", "tool", toolName, "session", sessionID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store event"})
		return
	}
	slog.Debug("event stored successfully", "tool", toolName, "session", sessionID, "event", eventName)
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "event stored"})
}

// StoreEvent stores an event (can be used programmatically, not just HTTP)
// Note: This method still writes directly to DB without going through the queue
// Use sparingly and only for non-hot-path code
func (hs *HookService) StoreEvent(sessionID string, toolName string, eventJSON []byte) error {
	if !hs.IsInitialized() {
		return fmt.Errorf("hook database not initialized")
	}

	// Validate required fields
	sessionIDFromEvent := gjson.GetBytes(eventJSON, "session_id").String()
	if sessionIDFromEvent == "" {
		return fmt.Errorf("missing session_id in event data")
	}

	eventName := gjson.GetBytes(eventJSON, "hook_event_name").String()
	if eventName == "" {
		return fmt.Errorf("missing hook_event_name in event data")
	}

	// Add timestamp if not present
	eventWithTS, err := AddTimestamp(eventJSON)
	if err != nil {
		return fmt.Errorf("failed to add timestamp: %w", err)
	}

	// Store event directly (bypasses queue for programmatic access)
	return storage.StoreEvent(hookdb.HookDB, sessionID, toolName, eventWithTS)
}

// GetHealth returns the health status of the hook service
func (hs *HookService) GetHealth() map[string]interface{} {
	status := map[string]interface{}{
		"enabled": hs.IsInitialized(),
	}

	if hs.IsInitialized() {
		status["database"] = "~/.code-together/hook-events.db"
		status["queue_depth"] = len(hs.writeQueue)
		status["queue_capacity"] = cap(hs.writeQueue)
	}

	return status
}
