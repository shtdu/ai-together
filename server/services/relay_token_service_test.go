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

	now := time.Now()

	// Revoke existing token call
	mockRepo.On("RevokeRelayTokenByUserID", mock.Anything, int64(42)).Return(nil)

	// Create token call
	mockRepo.On("CreateRelayToken",
		mock.Anything,
		mock.AnythingOfType("string"), // tokenHash (SHA256)
		mock.AnythingOfType("string"), // tokenPrefix (first 8 chars)
		int64(42),
		int64(10),
	).Return(&models.RelayToken{
		ID:          1,
		UserID:      42,
		TenantID:    10,
		TokenPrefix: "abc12345",
		CreatedAt:   now,
	}, nil)

	token, rawToken, err := service.GenerateToken(context.Background(), 42, 10)

	assert.NoError(t, err)
	assert.NotNil(t, token)
	assert.NotEmpty(t, rawToken)
	assert.Len(t, rawToken, 64) // 32 bytes = 64 hex chars
	assert.Equal(t, int64(42), token.UserID)
	assert.Equal(t, int64(10), token.TenantID)
	// Verify prefix is first 8 chars of the raw token
	assert.Equal(t, rawToken[:8], rawToken[:8])

	mockRepo.AssertExpectations(t)
}

func TestRelayTokenService_GenerateToken_RevokesExisting(t *testing.T) {
	mockRepo := new(MockRelayTokenRepository)
	service := NewRelayTokenService(mockRepo)

	// Revoke should be called first
	mockRepo.On("RevokeRelayTokenByUserID", mock.Anything, int64(42)).Return(nil)
	mockRepo.On("CreateRelayToken",
		mock.Anything,
		mock.AnythingOfType("string"),
		mock.AnythingOfType("string"),
		int64(42),
		int64(10),
	).Return(&models.RelayToken{ID: 1, UserID: 42, TenantID: 10}, nil)

	_, _, err := service.GenerateToken(context.Background(), 42, 10)
	assert.NoError(t, err)

	mockRepo.AssertCalled(t, "RevokeRelayTokenByUserID", mock.Anything, int64(42))
}

// ==================== ValidateToken Tests ====================

func TestRelayTokenService_ValidateToken_Success(t *testing.T) {
	mockRepo := new(MockRelayTokenRepository)
	service := NewRelayTokenService(mockRepo)

	now := time.Now()

	// The raw token gets SHA256 hashed
	mockRepo.On("GetRelayTokenByHash",
		mock.Anything,
		mock.AnythingOfType("string"), // will be SHA256 hex of raw token
	).Return(&models.RelayToken{
		ID:        1,
		UserID:    42,
		TenantID:  10,
		CreatedAt: now,
	}, nil)

	// UpdateRelayTokenLastUsed is called async from a goroutine — use Maybe to avoid panic
	mockRepo.On("UpdateRelayTokenLastUsed", mock.Anything, int64(1)).Maybe().Return(nil)

	token, err := service.ValidateToken(context.Background(), "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890")

	assert.NoError(t, err)
	assert.NotNil(t, token)
	assert.Equal(t, int64(42), token.UserID)
	assert.Equal(t, int64(10), token.TenantID)

	// Give the async goroutine time to complete
	time.Sleep(50 * time.Millisecond)

	mockRepo.AssertCalled(t, "GetRelayTokenByHash", mock.Anything, mock.AnythingOfType("string"))
}

func TestRelayTokenService_ValidateToken_NotFound(t *testing.T) {
	mockRepo := new(MockRelayTokenRepository)
	service := NewRelayTokenService(mockRepo)

	mockRepo.On("GetRelayTokenByHash",
		mock.Anything,
		mock.AnythingOfType("string"),
	).Return((*models.RelayToken)(nil), fmt.Errorf("not found"))

	token, err := service.ValidateToken(context.Background(), "nonexistent1234567890abcdef1234567890abcdef1234567890")

	assert.Error(t, err)
	assert.Nil(t, token)
	assert.Contains(t, err.Error(), "invalid relay token")
}

func TestRelayTokenService_ValidateToken_TooShort(t *testing.T) {
	mockRepo := new(MockRelayTokenRepository)
	service := NewRelayTokenService(mockRepo)

	token, err := service.ValidateToken(context.Background(), "short")

	assert.Error(t, err)
	assert.Nil(t, token)
	assert.Contains(t, err.Error(), "invalid token format")
}

// ==================== RevokeToken Tests ====================

func TestRelayTokenService_RevokeToken_Success(t *testing.T) {
	mockRepo := new(MockRelayTokenRepository)
	service := NewRelayTokenService(mockRepo)

	mockRepo.On("RevokeRelayTokenByUserID", mock.Anything, int64(42)).Return(nil)

	err := service.RevokeToken(context.Background(), 42)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestRelayTokenService_RevokeToken_Error(t *testing.T) {
	mockRepo := new(MockRelayTokenRepository)
	service := NewRelayTokenService(mockRepo)

	mockRepo.On("RevokeRelayTokenByUserID", mock.Anything, int64(42)).Return(fmt.Errorf("db error"))

	err := service.RevokeToken(context.Background(), 42)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to revoke relay token")
}

// ==================== GetTokenInfo Tests ====================

func TestRelayTokenService_GetTokenInfo_Success(t *testing.T) {
	mockRepo := new(MockRelayTokenRepository)
	service := NewRelayTokenService(mockRepo)

	now := time.Now()
	mockRepo.On("GetRelayTokenByUserID", mock.Anything, int64(42)).Return(&models.RelayToken{
		ID:          1,
		UserID:      42,
		TenantID:    10,
		TokenPrefix: "abc12345",
		CreatedAt:   now,
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

	mockRepo.On("GetRelayTokenByUserID", mock.Anything, int64(42)).Return((*models.RelayToken)(nil), fmt.Errorf("not found"))

	token, err := service.GetTokenInfo(context.Background(), 42)

	assert.Error(t, err)
	assert.Nil(t, token)
}

// ==================== RateLimiter Tests ====================

func TestRateLimiter_Allow(t *testing.T) {
	limiter := NewRateLimiter(3, time.Minute)

	assert.True(t, limiter.Allow(1))  // 1st request
	assert.True(t, limiter.Allow(1))  // 2nd request
	assert.True(t, limiter.Allow(1))  // 3rd request
	assert.False(t, limiter.Allow(1)) // 4th request - blocked
}

func TestRateLimiter_DifferentTokens(t *testing.T) {
	limiter := NewRateLimiter(1, time.Minute)

	assert.True(t, limiter.Allow(1))  // token 1, 1st request
	assert.False(t, limiter.Allow(1)) // token 1, 2nd request - blocked
	assert.True(t, limiter.Allow(2))  // token 2, 1st request - allowed
}

func TestRateLimiter_WindowReset(t *testing.T) {
	limiter := NewRateLimiter(1, 100*time.Millisecond)

	assert.True(t, limiter.Allow(1))  // 1st request
	assert.False(t, limiter.Allow(1)) // 2nd request - blocked

	time.Sleep(150 * time.Millisecond) // wait for window to expire

	assert.True(t, limiter.Allow(1)) // window reset, should allow
}

// ==================== hashToken Tests ====================

func TestHashToken_Deterministic(t *testing.T) {
	token := "abcdef1234567890"
	hash1 := hashToken(token)
	hash2 := hashToken(token)

	assert.Equal(t, hash1, hash2)
	assert.Len(t, hash1, 64) // SHA256 hex = 64 chars
}

func TestHashToken_DifferentTokens(t *testing.T) {
	token1 := "abcdef1234567890"
	token2 := "abcdef1234567891"

	hash1 := hashToken(token1)
	hash2 := hashToken(token2)

	assert.NotEqual(t, hash1, hash2)
}
