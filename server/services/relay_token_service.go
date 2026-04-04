package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"switch-server/models"
	"switch-server/repository"
)

const (
	// TokenBytes is the number of random bytes in a relay token (32 bytes = 64 hex chars)
	TokenBytes = 32
	// TokenPrefixLength is the number of chars shown for identification
	TokenPrefixLength = 8
	// lastUsedWorkerCount is the number of background workers for last-used updates
	lastUsedWorkerCount = 4
	// lastUsedQueueSize is the max buffered updates before workers block
	lastUsedQueueSize = 1024
)

var (
	// ErrTokenNotFound is returned when a relay token does not exist (revoked or never created).
	ErrTokenNotFound = errors.New("relay token not found")
	// ErrInvalidTokenFormat is returned when the token string doesn't match expected format.
	ErrInvalidTokenFormat = errors.New("invalid token format")
)

// RateLimiter provides per-token in-memory rate limiting using a sliding window
type RateLimiter struct {
	mu       sync.Mutex
	limits   map[int64]*tokenBucket
	rate     int
	interval time.Duration
}

type tokenBucket struct {
	count  int
	expiry time.Time
}

// NewRateLimiter creates a new rate limiter with the given max requests per interval
func NewRateLimiter(rate int, interval time.Duration) *RateLimiter {
	rl := &RateLimiter{
		limits:   make(map[int64]*tokenBucket),
		rate:     rate,
		interval: interval,
	}
	go rl.cleanup()
	return rl
}

// Allow checks if a request from the given token ID is within the rate limit
func (rl *RateLimiter) Allow(tokenID int64) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	bucket, exists := rl.limits[tokenID]
	if !exists || now.After(bucket.expiry) {
		rl.limits[tokenID] = &tokenBucket{
			count:  1,
			expiry: now.Add(rl.interval),
		}
		return true
	}

	if bucket.count >= rl.rate {
		return false
	}

	bucket.count++
	return true
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for id, bucket := range rl.limits {
			if now.After(bucket.expiry) {
				delete(rl.limits, id)
			}
		}
		rl.mu.Unlock()
	}
}

// hashToken returns the SHA256 hex digest of a raw token for storage/lookup
func hashToken(rawToken string) string {
	h := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(h[:])
}

// RelayTokenServiceInterface defines the contract for relay token service operations
type RelayTokenServiceInterface interface {
	GenerateToken(ctx context.Context, userID, tenantID int64) (*models.RelayToken, string, error)
	RevokeToken(ctx context.Context, userID int64) error
	GetTokenInfo(ctx context.Context, userID int64) (*models.RelayToken, error)
	ValidateToken(ctx context.Context, rawToken string) (*models.RelayToken, error)
	Close()
}

// RelayTokenService handles relay token business logic
type RelayTokenService struct {
	repo       repository.RelayTokenRepositoryInterface
	lastUsedCh chan int64
	closeCh    chan struct{}
	wg         sync.WaitGroup
}

func NewRelayTokenService(repo repository.RelayTokenRepositoryInterface) *RelayTokenService {
	s := &RelayTokenService{
		repo:       repo,
		lastUsedCh: make(chan int64, lastUsedQueueSize),
		closeCh:    make(chan struct{}),
	}

	// Start bounded workers for last-used updates
	s.wg.Add(lastUsedWorkerCount)
	for i := 0; i < lastUsedWorkerCount; i++ {
		go s.lastUsedWorker(i)
	}

	return s
}

// Close shuts down background workers. Call this during graceful shutdown.
func (s *RelayTokenService) Close() {
	close(s.closeCh)
	s.wg.Wait()
}

// lastUsedWorker processes last-used update requests from the bounded queue.
// Each worker uses a timeout context to prevent slow DB from blocking indefinitely.
func (s *RelayTokenService) lastUsedWorker(id int) {
	defer s.wg.Done()
	for {
		select {
		case <-s.closeCh:
			return
		case tokenID := <-s.lastUsedCh:
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			if err := s.repo.UpdateRelayTokenLastUsed(ctx, tokenID); err != nil {
				slog.Warn("failed to update relay token last-used",
					"worker_id", id,
					"token_id", tokenID,
					"error", err,
				)
			}
			cancel()
		}
	}
}

// GenerateToken creates a new relay token for a user.
// Returns the token model and the raw token string (shown only once).
// If the user already has an active token, it is revoked first.
func (s *RelayTokenService) GenerateToken(ctx context.Context, userID, tenantID int64) (*models.RelayToken, string, error) {
	// Revoke existing token if any
	_ = s.repo.RevokeRelayTokenByUserID(ctx, userID)

	// Generate a secure random token
	rawToken, err := generateSecureToken(TokenBytes)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	// Hash the token for storage using SHA256 (deterministic, fast lookup)
	tokenHash := hashToken(rawToken)
	prefix := rawToken[:TokenPrefixLength]

	token, err := s.repo.CreateRelayToken(ctx, tokenHash, prefix, userID, tenantID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to store relay token: %w", err)
	}

	slog.Info("relay token generated",
		"user_id", userID,
		"tenant_id", tenantID,
		"token_prefix", prefix,
	)

	return token, rawToken, nil
}

// RevokeToken revokes the active relay token for a user
func (s *RelayTokenService) RevokeToken(ctx context.Context, userID int64) error {
	err := s.repo.RevokeRelayTokenByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to revoke relay token: %w", err)
	}
	slog.Info("relay token revoked", "user_id", userID)
	return nil
}

// GetTokenInfo returns the token info for a user (for display purposes)
func (s *RelayTokenService) GetTokenInfo(ctx context.Context, userID int64) (*models.RelayToken, error) {
	token, err := s.repo.GetRelayTokenByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get relay token info: %w", err)
	}
	return token, nil
}

// ValidateToken checks if a raw token is valid and returns the associated token record.
// Returns ErrTokenNotFound if the token doesn't exist, ErrInvalidTokenFormat if the
// format is wrong, or a wrapped DB error for infrastructure failures.
func (s *RelayTokenService) ValidateToken(ctx context.Context, rawToken string) (*models.RelayToken, error) {
	// Strict format validation: must be exactly 64 hex characters
	if len(rawToken) != TokenBytes*2 {
		return nil, ErrInvalidTokenFormat
	}
	for _, c := range rawToken {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return nil, ErrInvalidTokenFormat
		}
	}

	tokenHash := hashToken(rawToken)
	token, err := s.repo.GetRelayTokenByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTokenNotFound
		}
		// Preserve the original DB error so the caller can differentiate
		// not-found from infrastructure failures
		return nil, fmt.Errorf("failed to validate relay token: %w", err)
	}

	// Enqueue async last-used update via bounded worker pool
	select {
	case s.lastUsedCh <- token.ID:
		// enqueued successfully
	default:
		// queue full — skip this update rather than blocking the request
		slog.Warn("last-used update queue full, skipping",
			"token_id", token.ID,
		)
	}

	return token, nil
}

// generateSecureToken generates a cryptographically secure random hex token
func generateSecureToken(numBytes int) (string, error) {
	b := make([]byte, numBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
