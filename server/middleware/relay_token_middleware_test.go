package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"switch-server/models"
	"switch-server/services"
)

type MockRelayTokenServiceForMiddleware struct {
	mock.Mock
}

func (m *MockRelayTokenServiceForMiddleware) Close() {}

func (m *MockRelayTokenServiceForMiddleware) GenerateToken(ctx context.Context, userID, tenantID int64) (*models.RelayToken, string, error) {
	args := m.Called(ctx, userID, tenantID)
	if args.Get(0) == nil {
		return nil, "", args.Error(2)
	}
	return args.Get(0).(*models.RelayToken), args.String(1), args.Error(2)
}

func (m *MockRelayTokenServiceForMiddleware) RevokeToken(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockRelayTokenServiceForMiddleware) GetTokenInfo(ctx context.Context, userID int64) (*models.RelayToken, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RelayToken), args.Error(1)
}

func (m *MockRelayTokenServiceForMiddleware) ValidateToken(ctx context.Context, rawToken string) (*models.RelayToken, error) {
	args := m.Called(ctx, rawToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RelayToken), args.Error(1)
}

func setupMiddlewareTestRouter(relayTokenService services.RelayTokenServiceInterface, rateLimiter *services.RateLimiter) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RelayTokenAuthMiddleware(relayTokenService, rateLimiter))
	router.Use(func(c *gin.Context) {
		// Simulate session auth fallback — only set if not already authenticated
		if _, exists := c.Get("user"); !exists {
			c.Set("user", &models.User{ID: 99, TenantID: 99})
			c.Set("auth_method", "session")
		}
		c.Next()
	})
	router.POST("/test", func(c *gin.Context) {
		user, _ := c.Get("user")
		authMethod, _ := c.Get("auth_method")
		c.JSON(200, gin.H{
			"user_id":     user.(*models.User).ID,
			"tenant_id":   user.(*models.User).TenantID,
			"auth_method": authMethod,
		})
	})
	return router
}

func TestRelayTokenAuthMiddleware_NoAuthHeader(t *testing.T) {
	mockService := new(MockRelayTokenServiceForMiddleware)
	router := setupMiddlewareTestRouter(mockService, nil)

	req, _ := http.NewRequest("POST", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertNotCalled(t, "ValidateToken")
}

func TestRelayTokenAuthMiddleware_JWTToken_FallsThrough(t *testing.T) {
	mockService := new(MockRelayTokenServiceForMiddleware)
	router := setupMiddlewareTestRouter(mockService, nil)

	req, _ := http.NewRequest("POST", "/test", nil)
	req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertNotCalled(t, "ValidateToken")
}

func TestRelayTokenAuthMiddleware_ValidRelayToken(t *testing.T) {
	mockService := new(MockRelayTokenServiceForMiddleware)
	router := setupMiddlewareTestRouter(mockService, nil)

	now := time.Now()
	mockService.On("ValidateToken", mock.Anything, "abc12345def67890abcdef1234567890abcdef1234567890abcdef1234567890").
		Return(&models.RelayToken{
			ID:        1,
			UserID:    42,
			TenantID:  10,
			CreatedAt: now,
		}, nil)

	req, _ := http.NewRequest("POST", "/test", nil)
	req.Header.Set("Authorization", "Bearer abc12345def67890abcdef1234567890abcdef1234567890abcdef1234567890")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(42), response["user_id"])
	assert.Equal(t, float64(10), response["tenant_id"])
	assert.Equal(t, "relay_token", response["auth_method"])

	mockService.AssertExpectations(t)
}

func TestRelayTokenAuthMiddleware_InvalidRelayToken_FallsThrough(t *testing.T) {
	mockService := new(MockRelayTokenServiceForMiddleware)
	router := setupMiddlewareTestRouter(mockService, nil)

	mockService.On("ValidateToken", mock.Anything, "invalidhex").
		Return((*models.RelayToken)(nil), services.ErrTokenNotFound)

	req, _ := http.NewRequest("POST", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalidhex")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(99), response["user_id"])
	assert.Equal(t, "session", response["auth_method"])

	mockService.AssertExpectations(t)
}

func TestRelayTokenAuthMiddleware_RateLimited(t *testing.T) {
	mockService := new(MockRelayTokenServiceForMiddleware)
	limiter := services.NewRateLimiter(1, time.Minute)
	router := setupMiddlewareTestRouter(mockService, limiter)

	now := time.Now()
	token := &models.RelayToken{
		ID:        1,
		UserID:    42,
		TenantID:  10,
		CreatedAt: now,
	}

	mockService.On("ValidateToken", mock.Anything, "abc12345def67890").Return(token, nil).Twice()

	// First request should succeed
	req1, _ := http.NewRequest("POST", "/test", nil)
	req1.Header.Set("Authorization", "Bearer abc12345def67890")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Second request should be rate limited
	req2, _ := http.NewRequest("POST", "/test", nil)
	req2.Header.Set("Authorization", "Bearer abc12345def67890")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)

	mockService.AssertExpectations(t)
}

func TestRelayTokenAuthMiddleware_DBError_Returns500(t *testing.T) {
	mockService := new(MockRelayTokenServiceForMiddleware)
	router := setupMiddlewareTestRouter(mockService, nil)

	// DB error (not ErrTokenNotFound or ErrInvalidTokenFormat) should return 500
	mockService.On("ValidateToken", mock.Anything, mock.AnythingOfType("string")).
		Return((*models.RelayToken)(nil), fmt.Errorf("connection refused"))

	req, _ := http.NewRequest("POST", "/test", nil)
	req.Header.Set("Authorization", "Bearer abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	mockService.AssertExpectations(t)
}
