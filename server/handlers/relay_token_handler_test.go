package handlers

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
	"github.com/stretchr/testify/require"
	"switch-server/models"
)

// ==================== Mock Relay Token Service ====================

type MockRelayTokenService struct {
	mock.Mock
}

func (m *MockRelayTokenService) GenerateToken(ctx context.Context, userID, tenantID int64) (*models.RelayToken, string, error) {
	args := m.Called(ctx, userID, tenantID)
	if args.Get(0) == nil {
		return nil, "", args.Error(2)
	}
	return args.Get(0).(*models.RelayToken), args.String(1), args.Error(2)
}

func (m *MockRelayTokenService) RevokeToken(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockRelayTokenService) GetTokenInfo(ctx context.Context, userID int64) (*models.RelayToken, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RelayToken), args.Error(1)
}

func (m *MockRelayTokenService) ValidateToken(ctx context.Context, rawToken string) (*models.RelayToken, error) {
	args := m.Called(ctx, rawToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RelayToken), args.Error(1)
}

// ==================== Helper ====================

func setupMemberRelayTokenRouter(handler *RelayTokenHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(func(c *gin.Context) {
		userID := c.GetHeader("X-Test-User-ID")
		tenantID := c.GetHeader("X-Test-User-Tenant-ID")
		role := c.GetHeader("X-Test-User-Role")

		if userID != "" {
			var id, tenant int64
			_, _ = fmt.Sscanf(userID, "%d", &id)
			_, _ = fmt.Sscanf(tenantID, "%d", &tenant)

			user := &models.User{
				ID:       id,
				Role:     role,
				TenantID: tenant,
			}
			c.Set("user", user)
		}
		c.Next()
	})

	router.POST("/api/v1/user/relay-token", handler.MyRelayToken)
	router.GET("/api/v1/user/relay-token", handler.GetMyRelayTokenInfo)
	router.DELETE("/api/v1/user/relay-token", handler.RevokeMyRelayToken)

	return router
}

func newTestHandler() (*RelayTokenHandler, *MockRelayTokenService, *gin.Engine) {
	mockService := new(MockRelayTokenService)
	handler := NewRelayTokenHandler(mockService)
	router := setupMemberRelayTokenRouter(handler)
	return handler, mockService, router
}

func memberHeaders(id, tenant int64, role string) map[string]string {
	return map[string]string{
		"X-Test-User-ID":        fmt.Sprintf("%d", id),
		"X-Test-User-Tenant-ID": fmt.Sprintf("%d", tenant),
		"X-Test-User-Role":      role,
	}
}

// ==================== MyRelayToken Tests ====================

func TestMyRelayToken_Success(t *testing.T) {
	_, mockService, router := newTestHandler()

	now := time.Now()
	mockService.On("GenerateToken", mock.Anything, int64(42), int64(1)).Return(
		&models.RelayToken{
			ID:          1,
			UserID:      42,
			TenantID:    1,
			TokenPrefix: "abc12345",
			CreatedAt:   now,
		},
		"abc12345def67890abcdef1234567890abcdef1234567890abcdef1234567890",
		nil,
	)

	req, _ := http.NewRequest("POST", "/api/v1/user/relay-token", nil)
	for k, v := range memberHeaders(42, 1, "member") {
		req.Header.Set(k, v)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "token")
	assert.Contains(t, response, "prefix")
	assert.Contains(t, response, "message")
	assert.Equal(t, "abc12345", response["prefix"])

	mockService.AssertExpectations(t)
}

func TestMyRelayToken_NoUserContext(t *testing.T) {
	_, _, router := newTestHandler()

	req, _ := http.NewRequest("POST", "/api/v1/user/relay-token", nil)
	// No user context headers

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMyRelayToken_ServiceError(t *testing.T) {
	_, mockService, router := newTestHandler()

	mockService.On("GenerateToken", mock.Anything, int64(42), int64(1)).
		Return((*models.RelayToken)(nil), "", fmt.Errorf("db error"))

	req, _ := http.NewRequest("POST", "/api/v1/user/relay-token", nil)
	for k, v := range memberHeaders(42, 1, "member") {
		req.Header.Set(k, v)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

// ==================== GetMyRelayTokenInfo Tests ====================

func TestGetMyRelayTokenInfo_Success(t *testing.T) {
	_, mockService, router := newTestHandler()

	now := time.Now()
	lastUsed := time.Now().Add(1 * time.Hour)
	mockService.On("GetTokenInfo", mock.Anything, int64(42)).Return(
		&models.RelayToken{
			ID:          1,
			UserID:      42,
			TenantID:    1,
			TokenPrefix: "abc12345",
			CreatedAt:   now,
			LastUsedAt:  &lastUsed,
		}, nil,
	)

	req, _ := http.NewRequest("GET", "/api/v1/user/relay-token", nil)
	for k, v := range memberHeaders(42, 1, "member") {
		req.Header.Set(k, v)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "abc12345", response["prefix"])
	assert.Contains(t, response, "created_at")
	assert.Contains(t, response, "last_used_at")
	// Raw token should NOT be in info response
	_, hasToken := response["token"]
	assert.False(t, hasToken)

	mockService.AssertExpectations(t)
}

func TestGetMyRelayTokenInfo_NoToken(t *testing.T) {
	_, mockService, router := newTestHandler()

	mockService.On("GetTokenInfo", mock.Anything, int64(42)).
		Return((*models.RelayToken)(nil), fmt.Errorf("not found"))

	req, _ := http.NewRequest("GET", "/api/v1/user/relay-token", nil)
	for k, v := range memberHeaders(42, 1, "member") {
		req.Header.Set(k, v)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestGetMyRelayTokenInfo_NoUserContext(t *testing.T) {
	_, _, router := newTestHandler()

	req, _ := http.NewRequest("GET", "/api/v1/user/relay-token", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ==================== RevokeMyRelayToken Tests ====================

func TestRevokeMyRelayToken_Success(t *testing.T) {
	_, mockService, router := newTestHandler()

	mockService.On("RevokeToken", mock.Anything, int64(42)).Return(nil)

	req, _ := http.NewRequest("DELETE", "/api/v1/user/relay-token", nil)
	for k, v := range memberHeaders(42, 1, "member") {
		req.Header.Set(k, v)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "message")

	mockService.AssertExpectations(t)
}

func TestRevokeMyRelayToken_ServiceError(t *testing.T) {
	_, mockService, router := newTestHandler()

	mockService.On("RevokeToken", mock.Anything, int64(42)).Return(fmt.Errorf("db error"))

	req, _ := http.NewRequest("DELETE", "/api/v1/user/relay-token", nil)
	for k, v := range memberHeaders(42, 1, "member") {
		req.Header.Set(k, v)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestRevokeMyRelayToken_NoUserContext(t *testing.T) {
	_, _, router := newTestHandler()

	req, _ := http.NewRequest("DELETE", "/api/v1/user/relay-token", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
