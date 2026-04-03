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
	"switch-server/services"
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

func setupRelayTokenRouter(handler *RelayTokenHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Middleware to set up user context for admin
	router.Use(func(c *gin.Context) {
		userEmail := c.GetHeader("X-Test-User-Email")
		if userEmail != "" {
			userID := c.GetHeader("X-Test-User-ID")
			userRole := c.GetHeader("X-Test-User-Role")
			tenantID := c.GetHeader("X-Test-User-Tenant-ID")

			var id, tenant int64
			if userID != "" {
				_, _ = fmt.Sscanf(userID, "%d", &id)
			}
			if tenantID != "" {
				_, _ = fmt.Sscanf(tenantID, "%d", &tenant)
			}

			user := &models.User{
				ID:       id,
				Email:    userEmail,
				Role:     userRole,
				TenantID: tenant,
			}
			c.Set("user", user)
		}
		c.Next()
	})

	router.POST("/api/v1/users/:id/relay-token", handler.GenerateRelayToken)
	router.DELETE("/api/v1/users/:id/relay-token", handler.RevokeRelayToken)
	router.GET("/api/v1/users/:id/relay-token", handler.GetRelayTokenInfo)

	return router
}

func setupRelayTokenHandlerWithUser(relayTokenService services.RelayTokenServiceInterface, user *models.User) (*RelayTokenHandler, *gin.Engine) {
	mockUserService := new(MockUserService)
	// Set up mock for GetUserByID
	mockUserService.On("GetUserByID", mock.AnythingOfType("int64")).Return(
		&models.User{
			ID:       2,
			Email:    "member@example.com",
			Name:     "Member",
			Role:     "member",
			TenantID: user.TenantID,
		}, nil,
	)

	handler := NewRelayTokenHandler(relayTokenService, mockUserService)
	router := setupRelayTokenRouter(handler)
	return handler, router
}

// ==================== GenerateRelayToken Tests ====================

func TestRelayTokenHandler_GenerateRelayToken_Success(t *testing.T) {
	mockRelayTokenService := new(MockRelayTokenService)

	adminUser := createTestUser(1, "admin@example.com", "Admin", "manager", 1)
	_, router := setupRelayTokenHandlerWithUser(mockRelayTokenService, adminUser)

	now := time.Now()
	mockRelayTokenService.On("GenerateToken", mock.Anything, int64(2), int64(1)).Return(
		&models.RelayToken{
			ID:          1,
			UserID:      2,
			TenantID:    1,
			TokenPrefix: "abc12345",
			CreatedAt:   now,
		},
		"abc12345def67890...",
		nil,
	)

	req, _ := http.NewRequest("POST", "/api/v1/users/2/relay-token", nil)
	req.Header.Set("X-Test-User-Email", adminUser.Email)
	req.Header.Set("X-Test-User-ID", "1")
	req.Header.Set("X-Test-User-Role", "manager")
	req.Header.Set("X-Test-User-Tenant-ID", "1")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "token")
	assert.Contains(t, response, "prefix")
	assert.Equal(t, "abc12345", response["prefix"])

	mockRelayTokenService.AssertExpectations(t)
}

func TestRelayTokenHandler_GenerateRelayToken_InvalidUserID(t *testing.T) {
	mockRelayTokenService := new(MockRelayTokenService)
	mockUserService := new(MockUserService)

	handler := NewRelayTokenHandler(mockRelayTokenService, mockUserService)
	router := setupRelayTokenRouter(handler)

	adminUser := createTestUser(1, "admin@example.com", "Admin", "manager", 1)

	req, _ := http.NewRequest("POST", "/api/v1/users/invalid/relay-token", nil)
	req.Header.Set("X-Test-User-Email", adminUser.Email)
	req.Header.Set("X-Test-User-ID", "1")
	req.Header.Set("X-Test-User-Role", "manager")
	req.Header.Set("X-Test-User-Tenant-ID", "1")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRelayTokenHandler_GenerateRelayToken_UserNotFound(t *testing.T) {
	mockRelayTokenService := new(MockRelayTokenService)
	mockUserService := new(MockUserService)

	mockUserService.On("GetUserByID", int64(999)).Return((*models.User)(nil), fmt.Errorf("not found"))

	handler := NewRelayTokenHandler(mockRelayTokenService, mockUserService)
	router := setupRelayTokenRouter(handler)

	adminUser := createTestUser(1, "admin@example.com", "Admin", "manager", 1)

	req, _ := http.NewRequest("POST", "/api/v1/users/999/relay-token", nil)
	req.Header.Set("X-Test-User-Email", adminUser.Email)
	req.Header.Set("X-Test-User-ID", "1")
	req.Header.Set("X-Test-User-Role", "manager")
	req.Header.Set("X-Test-User-Tenant-ID", "1")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockUserService.AssertExpectations(t)
}

func TestRelayTokenHandler_GenerateRelayToken_DifferentTenant(t *testing.T) {
	mockRelayTokenService := new(MockRelayTokenService)
	mockUserService := new(MockUserService)

	// Target user in tenant 2
	mockUserService.On("GetUserByID", int64(2)).Return(
		&models.User{ID: 2, Email: "user@other.com", TenantID: 2}, nil,
	)

	handler := NewRelayTokenHandler(mockRelayTokenService, mockUserService)
	router := setupRelayTokenRouter(handler)

	// Admin in tenant 1
	adminUser := createTestUser(1, "admin@example.com", "Admin", "manager", 1)

	req, _ := http.NewRequest("POST", "/api/v1/users/2/relay-token", nil)
	req.Header.Set("X-Test-User-Email", adminUser.Email)
	req.Header.Set("X-Test-User-ID", "1")
	req.Header.Set("X-Test-User-Role", "manager")
	req.Header.Set("X-Test-User-Tenant-ID", "1")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRelayTokenHandler_GenerateRelayToken_NotManager(t *testing.T) {
	mockRelayTokenService := new(MockRelayTokenService)
	mockUserService := new(MockUserService)

	mockUserService.On("GetUserByID", int64(2)).Return(
		&models.User{ID: 2, Email: "member@example.com", TenantID: 1}, nil,
	)

	handler := NewRelayTokenHandler(mockRelayTokenService, mockUserService)
	router := setupRelayTokenRouter(handler)

	// Member in tenant 1 (not a manager)
	memberUser := createTestUser(3, "member@example.com", "Member", "member", 1)

	req, _ := http.NewRequest("POST", "/api/v1/users/2/relay-token", nil)
	req.Header.Set("X-Test-User-Email", memberUser.Email)
	req.Header.Set("X-Test-User-ID", "3")
	req.Header.Set("X-Test-User-Role", "member")
	req.Header.Set("X-Test-User-Tenant-ID", "1")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

// ==================== RevokeRelayToken Tests ====================

func TestRelayTokenHandler_RevokeRelayToken_Success(t *testing.T) {
	mockRelayTokenService := new(MockRelayTokenService)

	adminUser := createTestUser(1, "admin@example.com", "Admin", "manager", 1)
	_, router := setupRelayTokenHandlerWithUser(mockRelayTokenService, adminUser)

	mockRelayTokenService.On("RevokeToken", mock.Anything, int64(2)).Return(nil)

	req, _ := http.NewRequest("DELETE", "/api/v1/users/2/relay-token", nil)
	req.Header.Set("X-Test-User-Email", adminUser.Email)
	req.Header.Set("X-Test-User-ID", "1")
	req.Header.Set("X-Test-User-Role", "manager")
	req.Header.Set("X-Test-User-Tenant-ID", "1")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockRelayTokenService.AssertExpectations(t)
}

func TestRelayTokenHandler_RevokeRelayToken_ServiceError(t *testing.T) {
	mockRelayTokenService := new(MockRelayTokenService)

	adminUser := createTestUser(1, "admin@example.com", "Admin", "manager", 1)
	_, router := setupRelayTokenHandlerWithUser(mockRelayTokenService, adminUser)

	mockRelayTokenService.On("RevokeToken", mock.Anything, int64(2)).Return(fmt.Errorf("db error"))

	req, _ := http.NewRequest("DELETE", "/api/v1/users/2/relay-token", nil)
	req.Header.Set("X-Test-User-Email", adminUser.Email)
	req.Header.Set("X-Test-User-ID", "1")
	req.Header.Set("X-Test-User-Role", "manager")
	req.Header.Set("X-Test-User-Tenant-ID", "1")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockRelayTokenService.AssertExpectations(t)
}

// ==================== GetRelayTokenInfo Tests ====================

func TestRelayTokenHandler_GetRelayTokenInfo_Success(t *testing.T) {
	mockRelayTokenService := new(MockRelayTokenService)

	adminUser := createTestUser(1, "admin@example.com", "Admin", "manager", 1)
	_, router := setupRelayTokenHandlerWithUser(mockRelayTokenService, adminUser)

	now := time.Now()
	mockRelayTokenService.On("GetTokenInfo", mock.Anything, int64(2)).Return(
		&models.RelayToken{
			ID:          1,
			UserID:      2,
			TenantID:    1,
			TokenPrefix: "abc12345",
			CreatedAt:   now,
		}, nil,
	)

	req, _ := http.NewRequest("GET", "/api/v1/users/2/relay-token", nil)
	req.Header.Set("X-Test-User-Email", adminUser.Email)
	req.Header.Set("X-Test-User-ID", "1")
	req.Header.Set("X-Test-User-Role", "manager")
	req.Header.Set("X-Test-User-Tenant-ID", "1")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "prefix")
	assert.Equal(t, "abc12345", response["prefix"])

	mockRelayTokenService.AssertExpectations(t)
}

func TestRelayTokenHandler_GetRelayTokenInfo_NoToken(t *testing.T) {
	mockRelayTokenService := new(MockRelayTokenService)

	adminUser := createTestUser(1, "admin@example.com", "Admin", "manager", 1)
	_, router := setupRelayTokenHandlerWithUser(mockRelayTokenService, adminUser)

	mockRelayTokenService.On("GetTokenInfo", mock.Anything, int64(2)).Return(
		(*models.RelayToken)(nil), fmt.Errorf("not found"),
	)

	req, _ := http.NewRequest("GET", "/api/v1/users/2/relay-token", nil)
	req.Header.Set("X-Test-User-Email", adminUser.Email)
	req.Header.Set("X-Test-User-ID", "1")
	req.Header.Set("X-Test-User-Role", "manager")
	req.Header.Set("X-Test-User-Tenant-ID", "1")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockRelayTokenService.AssertExpectations(t)
}

// ==================== GenerateRelayToken - No User Context ====================

func TestRelayTokenHandler_GenerateRelayToken_NoUserContext(t *testing.T) {
	mockRelayTokenService := new(MockRelayTokenService)
	mockUserService := new(MockUserService)

	mockUserService.On("GetUserByID", int64(2)).Return(
		&models.User{ID: 2, Email: "member@example.com", TenantID: 1}, nil,
	)

	handler := NewRelayTokenHandler(mockRelayTokenService, mockUserService)
	router := setupRelayTokenRouter(handler)

	req, _ := http.NewRequest("POST", "/api/v1/users/2/relay-token", nil)
	// No user context headers set

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
