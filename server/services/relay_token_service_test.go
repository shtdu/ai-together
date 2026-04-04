package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"switch-server/models"
)

// ==================== Mock Repository ====================

type MockRelayTokenRepository struct {
	mock.Mock
}

func (m *MockRelayTokenRepository) CreateRelayToken(ctx context.Context, tokenHash, tokenPrefix string, userID, tenantID int64) (*models.RelayToken, error) {
	args := m.Called(ctx, tokenHash, tokenPrefix, userID, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RelayToken), args.Error(1)
}

func (m *MockRelayTokenRepository) GetRelayTokenByHash(ctx context.Context, tokenHash string) (*models.RelayToken, error) {
	args := m.Called(ctx, tokenHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RelayToken), args.Error(1)
}

func (m *MockRelayTokenRepository) GetRelayTokenByUserID(ctx context.Context, userID int64) (*models.RelayToken, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RelayToken), args.Error(1)
}

func (m *MockRelayTokenRepository) RevokeRelayTokenByUserID(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockRelayTokenRepository) UpdateRelayTokenLastUsed(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// ==================== GenerateToken Tests ====================

func TestRelayTokenService_GenerateToken_Success(t *testing.T) {
	mockRepo := new(MockRelayTokenRepository)
	service := NewRelayTokenService(mockRepo)
	defer service.Close()

	now := time.Now()

	mockRepo.On("RevokeRelayTokenByUserID", mock.Anything, int64(42)).Return(nil)
	mockRepo.On("CreateRelayToken",
		mock.Anything,
		mock.AnythingOfType("string"),
		mock.AnythingOfType("string"),
		int64(42),
		int64(10),
	).Return(&models.RelayToken{
		ID: 1, UserID: 42, TenantID: 10, TokenPrefix: "abc12345", CreatedAt: now,
	}, nil)

	token, rawToken, err := service.GenerateToken(context.Background(), 42, 10)

	assert.NoError(t, err)
	assert.NotNil(t, token)
	assert.NotEmpty(t, rawToken)
	assert.Len(t, rawToken, 64)
	assert.Equal(t, int64(42), token.UserID)
	assert.Equal(t, int64(10), token.TenantID)

	mockRepo.AssertExpectations(t)
}

func TestRelayTokenService_GenerateToken_RevokesExisting(t *testing.T) {
	mockRepo := new(MockRelayTokenRepository)
	service := NewRelayTokenService(mockRepo)
	defer service.Close()

	mockRepo.On("RevokeRelayTokenByUserID", mock.Anything, int64(42)).Return(nil)
	mockRepo.On("CreateRelayToken",
		mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"),
		int64(42), int64(10),
	).Return(&models.RelayToken{ID: 1, UserID: 42, TenantID: 10}, nil)

	_, _, err := service.GenerateToken(context.Background(), 42, 10)
	assert.NoError(t, err)
	mockRepo.AssertCalled(t, "RevokeRelayTokenByUserID", mock.Anything, int64(42))
}

// ==================== ValidateToken Tests ====================

func TestRelayTokenService_ValidateToken_Success(t *testing.T) {
	mockRepo := new(MockRelayTokenRepository)
	service := NewRelayTokenService(mockRepo)
	defer service.Close()

	now := time.Now()

	mockRepo.On("GetRelayTokenByHash",
		mock.Anything,
		mock.AnythingOfType("string"),
	).Return(&models.RelayToken{
		ID: 1, UserID: 42, TenantID: 10, CreatedAt: now,
	}, nil)

	mockRepo.On("UpdateRelayTokenLastUsed", mock.Anything, int64(1)).Return(nil).Maybe()

	// Valid 64-char hex token
	token, err := service.ValidateToken(context.Background(), "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890")

	assert.NoError(t, err)
	assert.NotNil(t, token)
	assert.Equal(t, int64(42), token.UserID)

	// Give worker time to process the last-used update
	time.Sleep(100 * time.Millisecond)
}

func TestRelayTokenService_ValidateToken_NotFound(t *testing.T) {
	mockRepo := new(MockRelayTokenRepository)
	service := NewRelayTokenService(mockRepo)
	defer service.Close()

	mockRepo.On("GetRelayTokenByHash",
		mock.Anything,
		mock.AnythingOfType("string"),
	).Return((*models.RelayToken)(nil), fmt.Errorf("not found"))

	_, err := service.ValidateToken(context.Background(), "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890")

	assert.Error(t, err)
	// Error should be wrapped, not the raw DB error
	assert.Contains(t, err.Error(), "failed to validate relay token")
	// But should NOT be ErrTokenNotFound sentinel — the DB error is preserved
	assert.NotErrorIs(t, err, ErrTokenNotFound)
}

func TestRelayTokenService_ValidateToken_DBError(t *testing.T) {
	mockRepo := new(MockRelayTokenRepository)
	service := NewRelayTokenService(mockRepo)
	defer service.Close()

	mockRepo.On("GetRelayTokenByHash",
		mock.Anything,
		mock.AnythingOfType("string"),
	).Return((*models.RelayToken)(nil), fmt.Errorf("connection refused"))

	_, err := service.ValidateToken(context.Background(), "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to validate relay token")
	// DB errors should be distinguishable from not-found
	assert.NotErrorIs(t, err, ErrTokenNotFound)
	assert.NotErrorIs(t, err, ErrInvalidTokenFormat)
}

func TestRelayTokenService_ValidateToken_TooShort(t *testing.T) {
	mockRepo := new(MockRelayTokenRepository)
	service := NewRelayTokenService(mockRepo)
	defer service.Close()

	_, err := service.ValidateToken(context.Background(), "short")

	assert.ErrorIs(t, err, ErrInvalidTokenFormat)
}

func TestRelayTokenService_ValidateToken_WrongLength(t *testing.T) {
	mockRepo := new(MockRelayTokenRepository)
	service := NewRelayTokenService(mockRepo)
	defer service.Close()

	// 63 chars instead of 64
	_, err := service.ValidateToken(context.Background(), "abcdef1234567890abcdef1234567890abcdef1234567890abcdef123456789")

	assert.ErrorIs(t, err, ErrInvalidTokenFormat)
}

func TestRelayTokenService_ValidateToken_NonHexChars(t *testing.T) {
	mockRepo := new(MockRelayTokenRepository)
	service := NewRelayTokenService(mockRepo)
	defer service.Close()

	// 64 chars but contains non-hex characters
	_, err := service.ValidateToken(context.Background(), "ghijklmnopqrstuvwxyzghijklmnopqrstuvwxyzghijklmnopqrstuvwxyzghij")

	assert.ErrorIs(t, err, ErrInvalidTokenFormat)
}

func TestRelayTokenService_ValidateToken_ContainsDots(t *testing.T) {
	mockRepo := new(MockRelayTokenRepository)
	service := NewRelayTokenService(mockRepo)
	defer service.Close()

	// 64 chars but has dots — would be caught by middleware's JWT check first,
	// but the service itself should also reject non-hex chars (dots are not hex)
	_, err := service.ValidateToken(context.Background(), "abc.def1234567890abcdef1234567890abcdef1234567890abcdef1234567890a")

	assert.ErrorIs(t, err, ErrInvalidTokenFormat)
}

// ==================== RevokeToken Tests ====================

func TestRelayTokenService_RevokeToken_Success(t *testing.T) {
	mockRepo := new(MockRelayTokenRepository)
	service := NewRelayTokenService(mockRepo)
	defer service.Close()

	mockRepo.On("RevokeRelayTokenByUserID", mock.Anything, int64(42)).Return(nil)

	err := service.RevokeToken(context.Background(), 42)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestRelayTokenService_RevokeToken_Error(t *testing.T) {
	mockRepo := new(MockRelayTokenRepository)
	service := NewRelayTokenService(mockRepo)
	defer service.Close()

	mockRepo.On("RevokeRelayTokenByUserID", mock.Anything, int64(42)).Return(fmt.Errorf("db error"))

	err := service.RevokeToken(context.Background(), 42)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to revoke relay token")
}

// ==================== GetTokenInfo Tests ====================

func TestRelayTokenService_GetTokenInfo_Success(t *testing.T) {
	mockRepo := new(MockRelayTokenRepository)
	service := NewRelayTokenService(mockRepo)
	defer service.Close()

	now := time.Now()
	mockRepo.On("GetRelayTokenByUserID", mock.Anything, int64(42)).Return(&models.RelayToken{
		ID: 1, UserID: 42, TenantID: 10, TokenPrefix: "abc12345", CreatedAt: now,
	}, nil)

	token, err := service.GetTokenInfo(context.Background(), 42)

	assert.NoError(t, err)
	assert.NotNil(t, token)
	assert.Equal(t, "abc12345", token.TokenPrefix)
	mockRepo.AssertExpectations(t)
}

func TestRelayTokenService_GetTokenInfo_NotFound(t *testing.T) {
	mockRepo := new(MockRelayTokenRepository)
	service := NewRelayTokenService(mockRepo)
	defer service.Close()

	mockRepo.On("GetRelayTokenByUserID", mock.Anything, int64(42)).Return((*models.RelayToken)(nil), fmt.Errorf("not found"))

	token, err := service.GetTokenInfo(context.Background(), 42)
	assert.Error(t, err)
	assert.Nil(t, token)
}

// ==================== RateLimiter Tests ====================

func TestRateLimiter_Allow(t *testing.T) {
	limiter := NewRateLimiter(3, time.Minute)

	assert.True(t, limiter.Allow(1))
	assert.True(t, limiter.Allow(1))
	assert.True(t, limiter.Allow(1))
	assert.False(t, limiter.Allow(1))
}

func TestRateLimiter_DifferentTokens(t *testing.T) {
	limiter := NewRateLimiter(1, time.Minute)

	assert.True(t, limiter.Allow(1))
	assert.False(t, limiter.Allow(1))
	assert.True(t, limiter.Allow(2))
}

func TestRateLimiter_WindowReset(t *testing.T) {
	limiter := NewRateLimiter(1, 100*time.Millisecond)

	assert.True(t, limiter.Allow(1))
	assert.False(t, limiter.Allow(1))

	time.Sleep(150 * time.Millisecond)
	assert.True(t, limiter.Allow(1))
}

// ==================== hashToken Tests ====================

func TestHashToken_Deterministic(t *testing.T) {
	hash1 := hashToken("abcdef1234567890")
	hash2 := hashToken("abcdef1234567890")
	assert.Equal(t, hash1, hash2)
	assert.Len(t, hash1, 64)
}

func TestHashToken_DifferentTokens(t *testing.T) {
	hash1 := hashToken("abcdef1234567890")
	hash2 := hashToken("abcdef1234567891")
	assert.NotEqual(t, hash1, hash2)
}
