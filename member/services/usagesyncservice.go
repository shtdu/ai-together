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

package services

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"codeswitch/internal/db"
	"codeswitch/internal/models"
	"github.com/code-together/shared/integration"
)

const (
	batchSize     = 50 // Maximum number of records to sync in one batch
	retentionDays = 7  // Number of days to keep synced records
)

// UsageSyncService handles synchronization of usage statistics to the server
type UsageSyncService struct {
	authService *AuthService
	apiClient   integration.ClientWithResponsesInterface
}

// wrapAPIError converts an HTTP response into an error, parsing the response body
// for structured error information from the server
func (us *UsageSyncService) wrapAPIError(httpResp *http.Response, action string) error {
	if httpResp == nil {
		return fmt.Errorf("nil HTTP response from server")
	}

	// Successful response
	if httpResp.StatusCode >= 200 && httpResp.StatusCode < 300 {
		return nil
	}

	// Parse the error response
	apiErr := integration.UnwrapJSONResponse(httpResp)
	if apiErr != nil {
		// Add context about what action was being performed
		return fmt.Errorf("%s: %w", action, apiErr)
	}

	// Fallback to status code if parsing failed
	return fmt.Errorf("%s: server returned status %d", action, httpResp.StatusCode)
}

// NewUsageSyncService creates a new usage sync service
func NewUsageSyncService(authService *AuthService, apiClient integration.ClientWithResponsesInterface) *UsageSyncService {
	return &UsageSyncService{
		authService: authService,
		apiClient:   apiClient,
	}
}

// SyncUsageStats synchronizes pending usage records to the server
func (us *UsageSyncService) SyncUsageStats() (int, error) {
	// Check if user is authenticated
	if !us.authService.IsAuthenticated() {
		return 0, fmt.Errorf("user not authenticated")
	}

	// Get current user info
	user, err := us.authService.GetCurrentUser()
	if err != nil {
		return 0, fmt.Errorf("failed to get current user: %w", err)
	}

	// Get pending records
	pendingRecords, err := us.getPendingRecords()
	if err != nil {
		return 0, fmt.Errorf("failed to get pending records: %w", err)
	}

	if len(pendingRecords) == 0 {
		return 0, nil // Nothing to sync
	}

	// Transform and batch upload
	syncedCount := 0
	for i := 0; i < len(pendingRecords); i += batchSize {
		end := i + batchSize
		if end > len(pendingRecords) {
			end = len(pendingRecords)
		}

		batch := pendingRecords[i:end]
		syncedIDs, err := us.syncBatch(batch, user.ID, user.TenantID)
		if len(syncedIDs) > 0 {
			if err := us.MarkRecordsAsSynced(syncedIDs); err != nil {
				return syncedCount, fmt.Errorf("failed to mark records as synced: %w", err)
			}
			syncedCount += len(syncedIDs)
		}
		if err != nil {
			return syncedCount, fmt.Errorf("failed to sync batch %d-%d: %w", i, end, err)
		}
	}

	// Clean up old synced records after successful sync
	if _, err := us.DeleteOldRecords(); err != nil {
		// Log error but don't fail the sync operation
		// The cleanup is best-effort to keep database size manageable
	}

	return syncedCount, nil
}

// DeleteAllLocalUsageData deletes ALL records from request_log table
// This is used during logout to clean up all local usage data
func (us *UsageSyncService) DeleteAllLocalUsageData() (int, error) {
	slog.Info("deleting all local usage data")

	query := `DELETE FROM request_log`

	result, err := db.DB.Exec(query)
	if err != nil {
		return 0, fmt.Errorf("failed to delete all usage data: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get affected rows: %w", err)
	}

	slog.Info("deleted all local usage data", "deleted_count", rowsAffected)
	return int(rowsAffected), nil
}

// DeleteOldRecords removes synced records older than retentionDays
// This is called automatically after each sync to keep the database size manageable
func (us *UsageSyncService) DeleteOldRecords() (int, error) {
	// Delete records that are:
	// 1. Already synced to server (synced_to_server = 1)
	// 2. Older than retentionDays (7 days by default)
	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)
	const timeLayout = "2006-01-02 15:04:05"

	query := `DELETE FROM request_log
	          WHERE synced_to_server = 1
	          AND created_at < ?`

	result, err := db.DB.Exec(query, cutoffDate.Format(timeLayout))
	if err != nil {
		return 0, fmt.Errorf("failed to delete old records: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected > 0 {
		slog.Info("cleaned up old request_log entries",
			"deleted_count", rowsAffected,
			"cutoff_date", cutoffDate.Format(timeLayout),
		)
	}

	return int(rowsAffected), nil
}

// GetPendingRecordCount returns the number of pending (unsynced) records
func (us *UsageSyncService) GetPendingRecordCount() (int, error) {
	var count int
	err := db.DB.Get(&count, "SELECT COUNT(*) FROM request_log WHERE synced_to_server = 0")
	if err != nil {
		return 0, fmt.Errorf("failed to count pending records: %w", err)
	}
	return count, nil
}

// MarkRecordsAsSynced marks specific records as synced
func (us *UsageSyncService) MarkRecordsAsSynced(ids []int64) error {
	if len(ids) == 0 {
		return nil
	}

	query, args, err := buildINQuery("UPDATE request_log SET synced_to_server = 1 WHERE id", ids)
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = db.DB.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to mark records as synced: %w", err)
	}

	return nil
}

// getPendingRecords retrieves all unsynced records from the database
func (us *UsageSyncService) getPendingRecords() ([]models.RequestLog, error) {
	var records []models.RequestLog

	// Use sqlx for better struct mapping
	err := db.DB.Select(&records,
		"SELECT * FROM request_log WHERE synced_to_server = 0 ORDER BY created_at ASC",
	)
	if err != nil {
		return nil, err
	}

	return records, nil
}

// syncBatch sends a batch of records to the server
func (us *UsageSyncService) syncBatch(records []models.RequestLog, userID, tenantID int64) ([]int64, error) {
	if len(records) == 0 {
		return nil, nil
	}

	// Transform records to server format
	serverRecords := make([]integration.UsageRecord, len(records))
	recordIDs := make([]int64, len(records))
	for i, record := range records {
		serverRecords[i] = us.transformToServerFormat(record, userID, tenantID)
		recordIDs[i] = record.ID
	}

	// Send request using API client
	// Note: PostApiV1UsageBatchJSONRequestBody is an alias for []UsageRecord
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	resp, err := us.apiClient.PostApiV1UsageBatchWithResponse(ctx, serverRecords)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Check response status
	if err := us.wrapAPIError(resp.HTTPResponse, "sync usage batch"); err != nil {
		return nil, err
	}

	batchResp := resp.JSON200
	errors := batchResp.Errors

	// Parse errors if any
	failedIDs := make(map[int64]struct{})
	if errors != nil && len(*errors) > 0 {
		var err error
		failedIDs, err = parseBatchErrors(*errors)
		if err != nil {
			return nil, err
		}
	}

	syncedIDs := make([]int64, 0, len(records)-len(failedIDs))
	for _, record := range records {
		if _, failed := failedIDs[record.ID]; failed {
			continue
		}
		syncedIDs = append(syncedIDs, record.ID)
	}

	if errors != nil && len(*errors) > 0 {
		return syncedIDs, fmt.Errorf("server returned errors: %v", *errors)
	}

	return syncedIDs, nil
}

func parseBatchErrors(errors []string) (map[int64]struct{}, error) {
	failedIDs := make(map[int64]struct{}, len(errors))
	if len(errors) == 0 {
		return failedIDs, nil
	}

	re := regexp.MustCompile(`^record (\d+):`)
	for _, errMsg := range errors {
		matches := re.FindStringSubmatch(errMsg)
		if len(matches) != 2 {
			return nil, fmt.Errorf("failed to parse error record id: %q", errMsg)
		}
		id, err := strconv.ParseInt(matches[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse error record id: %q", errMsg)
		}
		failedIDs[id] = struct{}{}
	}

	return failedIDs, nil
}

// transformToServerFormat converts a local RequestLog to server UsageRecord format
func (us *UsageSyncService) transformToServerFormat(log models.RequestLog, userID, tenantID int64) integration.UsageRecord {
	record := integration.UsageRecord{
		Platform:  log.Platform,
		Model:     log.Model,
		Provider:  log.Provider,
		HttpCode:  log.HttpCode,
		CreatedAt: log.CreatedAt,
		TenantId:  &tenantID,
		UserId:    &userID,
	}

	// Convert integer fields to pointers if non-zero
	if log.InputTokens > 0 {
		record.InputTokens = &log.InputTokens
	}
	if log.OutputTokens > 0 {
		record.OutputTokens = &log.OutputTokens
	}
	if log.CacheCreateTokens > 0 {
		record.CacheCreateTokens = &log.CacheCreateTokens
	}
	if log.CacheReadTokens > 0 {
		record.CacheReadTokens = &log.CacheReadTokens
	}
	if log.ReasoningTokens > 0 {
		record.ReasoningTokens = &log.ReasoningTokens
	}

	// Convert other fields to pointers
	if log.DurationSec > 0 {
		durationSec := float32(log.DurationSec)
		record.DurationSec = &durationSec
	}
	record.IsStream = &log.IsStream

	return record
}

// buildINQuery builds a SQL IN query with the given IDs
func buildINQuery(prefix string, ids []int64) (string, []interface{}, error) {
	if len(ids) == 0 {
		return "", nil, fmt.Errorf("no IDs provided")
	}

	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf("%s IN (%s)", prefix, joinStrings(placeholders, ","))
	return query, args, nil
}

// joinStrings joins a slice of strings with a separator
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
