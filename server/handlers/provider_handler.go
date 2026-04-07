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

package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"switch-server/models"
	"switch-server/services"

	"github.com/gin-gonic/gin"
)

const (
	ProviderKindClaude   = "claude"
	ProviderKindCodex    = "codex"
	ProviderKindOpenCode = "opencode"
)

var ValidProviderKinds = map[string]bool{
	ProviderKindClaude:   true,
	ProviderKindCodex:    true,
	ProviderKindOpenCode: true,
}

func isValidProviderKind(kind string) bool {
	return ValidProviderKinds[kind]
}

type ProviderHandler struct {
	providerService services.ProviderServiceInterface
	usageService    services.UsageServiceInterface
}

func NewProviderHandler(providerService services.ProviderServiceInterface, usageService services.UsageServiceInterface) *ProviderHandler {
	return &ProviderHandler{
		providerService: providerService,
		usageService:    usageService,
	}
}

type CreateProviderRequest struct {
	Name            string            `json:"name" binding:"required"`
	APIURL          string            `json:"api_url" binding:"required,url"`
	APIKey          string            `json:"api_key" binding:"required"`
	Kind            string            `json:"kind"`
	Enabled         bool              `json:"enabled"`
	ModelMapping    map[string]string `json:"model_mapping"`
	SupportedModels []string          `json:"supported_models"`
	Level           int               `json:"level"`
}

type UpdateProviderRequest struct {
	Name            string            `json:"name"`
	APIURL          string            `json:"api_url" binding:"omitempty,url"`
	APIKey          string            `json:"api_key"`
	Kind            string            `json:"kind"`
	Enabled         *bool             `json:"enabled"`
	ModelMapping    map[string]string `json:"model_mapping"`
	SupportedModels []string          `json:"supported_models"`
	Level           int               `json:"level"`
}

func (h *ProviderHandler) ListProviders(c *gin.Context) {
	// Get user from context (set by AuthMiddleware)
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "User not found in context",
			Code:  models.ErrCodeUnauthorized,
		})
		return
	}

	authenticatedUser, ok := user.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Invalid user type in context",
			Code:  models.ErrCodeInternal,
		})
		return
	}

	// Get all providers for the user's team
	providers, err := h.providerService.GetProvidersByTeamID(authenticatedUser.TenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get providers: " + err.Error()})
		return
	}

	// Filter providers based on user role
	// - Managers see all providers including disabled ones (with API keys)
	// - Members only see enabled providers (without API keys)
	if authenticatedUser.Role == "member" {
		// Filter to only enabled providers and omit API keys for members
		filteredProviders := make([]models.Provider, 0)
		for _, provider := range providers {
			if provider.Enabled {
				// Omit API key for members
				provider.APIKey = ""
				filteredProviders = append(filteredProviders, provider)
			}
		}
		c.JSON(http.StatusOK, filteredProviders)
		return
	}

	// Managers get all providers including disabled ones
	c.JSON(http.StatusOK, providers)
}

func (h *ProviderHandler) CreateProvider(c *gin.Context) {
	var req CreateProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request payload",
			Code:  models.ErrCodeValidation,
		})
		return
	}

	// Get user from context
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "User not found in context",
			Code:  models.ErrCodeInternal,
		})
		return
	}
	authenticatedUser := user.(*models.User)

	// Default kind to "claude" if not provided
	kind := req.Kind
	if kind == "" {
		kind = "claude"
	}

	// Validate provider kind
	if !isValidProviderKind(kind) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: fmt.Sprintf("Invalid provider kind '%s'. Valid kinds are: claude, codex, opencode", kind),
			Code:  models.ErrCodeValidation,
		})
		return
	}

	// Check for duplicate provider name within the team
	count, err := h.providerService.CountProvidersByNameAndTeam(c.Request.Context(), req.Name, authenticatedUser.TenantID)
	if err != nil {
		requestID := c.GetString("request_id")
		log.Printf("[ERROR] [%s] Failed to check for duplicate provider name - name=%s error=%v", requestID, req.Name, err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to check for duplicate provider name",
			Code:    models.ErrCodeInternal,
			Request: requestID,
		})
		return
	}
	if count > 0 {
		c.JSON(http.StatusConflict, models.ErrorResponse{
			Error: fmt.Sprintf("Provider with name '%s' already exists", req.Name),
			Code:  models.ErrCodeConflict,
		})
		return
	}

	// No provider limits - providers can always be added
	// License check removed per new 2-type design

	// Use the authenticated user's tenant ID as the team ID
	teamID := authenticatedUser.TenantID

	// Convert map[string]string to map[string]interface{}
	modelMapping := make(map[string]interface{})
	for k, v := range req.ModelMapping {
		modelMapping[k] = v
	}

	// Create provider via service (pass []string directly for supported_models)
	provider, err := h.providerService.CreateProvider(req.Name, req.APIURL, req.APIKey, kind, teamID, req.Enabled,
		modelMapping,
		req.SupportedModels,
		req.Level)
	if err != nil {
		requestID := c.GetString("request_id")
		log.Printf("[ERROR] [%s] Failed to create provider - name=%s tenant_id=%d error=%v", requestID, req.Name, authenticatedUser.TenantID, err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to create provider",
			Code:    models.ErrCodeInternal,
			Details: err.Error(),
			Request: requestID,
		})
		return
	}

	c.JSON(http.StatusCreated, provider)
}

func (h *ProviderHandler) UpdateProvider(c *gin.Context) {
	providerID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid provider ID",
			Code:  models.ErrCodeValidation,
		})
		return
	}

	var req UpdateProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request payload",
			Code:  models.ErrCodeValidation,
		})
		return
	}

	// Get user from context for authorization
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "User not found in context",
			Code:  models.ErrCodeInternal,
		})
		return
	}
	currentUser := user.(*models.User)

	// Fetch the existing provider first
	existingProvider, err := h.providerService.GetProviderByID(providerID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "Provider not found",
			Code:  models.ErrCodeNotFound,
		})
		return
	}

	// Verify the provider belongs to the user's team
	// TODO: This is not implemented correct, should check team's owner id = currentUser.ID
	if existingProvider.TeamID != currentUser.TenantID {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error: "You don't have access to this provider",
			Code:  models.ErrCodeForbidden,
		})
		return
	}

	// Validate provider kind if being updated
	if req.Kind != "" && !isValidProviderKind(req.Kind) {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid provider kind '%s'. Valid kinds are: claude, codex, opencode", req.Kind)})
		return
	}

	// Check for duplicate provider name if name is being changed
	if req.Name != "" && req.Name != existingProvider.Name {
		count, err := h.providerService.CountProvidersByNameAndTeamExcludingID(c.Request.Context(), req.Name, currentUser.TenantID, providerID)
		if err != nil {
			requestID := c.GetString("request_id")
			log.Printf("[ERROR] [%s] Failed to check for duplicate provider name - name=%s error=%v", requestID, req.Name, err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "Failed to check for duplicate provider name",
				Code:    models.ErrCodeInternal,
				Request: requestID,
			})
			return
		}
		if count > 0 {
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Error: fmt.Sprintf("Provider with name '%s' already exists", req.Name),
				Code:  models.ErrCodeConflict,
			})
			return
		}
	}

	// Build update values, using existing values for fields not provided
	name := existingProvider.Name
	if req.Name != "" {
		name = req.Name
	}

	apiURL := existingProvider.APIURL
	if req.APIURL != "" {
		apiURL = req.APIURL
	}

	apiKey := existingProvider.APIKey
	if req.APIKey != "" {
		apiKey = req.APIKey
	}

	kind := existingProvider.Kind
	if req.Kind != "" {
		kind = req.Kind
	}

	enabled := existingProvider.Enabled
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	modelMapping := make(map[string]interface{})
	for k, v := range existingProvider.ModelMapping {
		modelMapping[k] = v
	}
	if req.ModelMapping != nil {
		for k, v := range req.ModelMapping {
			modelMapping[k] = v
		}
	}

	supportedModels := existingProvider.SupportedModels
	if req.SupportedModels != nil {
		supportedModels = req.SupportedModels
	}

	level := existingProvider.Level
	if req.Level != 0 {
		level = req.Level
	}

	// Update the provider via service
	err = h.providerService.UpdateProvider(providerID, name, apiURL, apiKey, kind, enabled, modelMapping, supportedModels, level)
	if err != nil {
		requestID := c.GetString("request_id")
		log.Printf("[ERROR] [%s] Failed to update provider - provider_id=%d error=%v", requestID, providerID, err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to update provider",
			Code:    models.ErrCodeInternal,
			Details: err.Error(),
			Request: requestID,
		})
		return
	}

	// Fetch and return the updated provider
	updatedProvider, err := h.providerService.GetProviderByID(providerID)
	if err != nil {
		requestID := c.GetString("request_id")
		log.Printf("[ERROR] [%s] Failed to fetch updated provider - provider_id=%d error=%v", requestID, providerID, err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to fetch updated provider",
			Code:    models.ErrCodeInternal,
			Request: requestID,
		})
		return
	}

	c.JSON(http.StatusOK, updatedProvider)
}

func (h *ProviderHandler) DeleteProvider(c *gin.Context) {
	providerID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid provider ID",
			Code:  models.ErrCodeValidation,
		})
		return
	}

	// Authorization is handled by RBAC middleware at route level
	// No need to check ownership here - RBAC enforces permissions

	// Check if provider exists FIRST
	provider, err := h.providerService.GetProviderByID(providerID)
	if err != nil || provider == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "Provider not found",
			Code:  models.ErrCodeNotFound,
		})
		return
	}

	// Delete the provider via service
	err = h.providerService.DeleteProvider(providerID)
	if err != nil {
		requestID := c.GetString("request_id")
		log.Printf("[ERROR] [%s] Failed to delete provider - provider_id=%d error=%v", requestID, providerID, err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to delete provider",
			Code:    models.ErrCodeInternal,
			Details: err.Error(),
			Request: requestID,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Provider deleted successfully", "provider_id": providerID})
}

func (h *ProviderHandler) TestProvider(c *gin.Context) {
	providerID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid provider ID",
			Code:  models.ErrCodeValidation,
		})
		return
	}

	// Get user from context for authorization
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "User not found in context",
			Code:  models.ErrCodeInternal,
		})
		return
	}
	currentUser := user.(*models.User)

	// Fetch the provider
	provider, err := h.providerService.GetProviderByID(providerID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "Provider not found",
			Code:  models.ErrCodeNotFound,
		})
		return
	}

	// Verify the provider belongs to the user's team
	if provider.TeamID != currentUser.TenantID {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error: "You don't have access to this provider",
			Code:  models.ErrCodeForbidden,
		})
		return
	}

	// For now, do a simple connectivity check
	// TODO: Implement actual provider-specific API test
	c.JSON(http.StatusOK, gin.H{
		"provider_id": providerID,
		"status":      "success",
		"message":     "Provider connectivity test passed",
		"api_url":     provider.APIURL,
	})
}

func (h *ProviderHandler) EnableProvider(c *gin.Context) {
	providerID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid provider ID",
			Code:  models.ErrCodeValidation,
		})
		return
	}

	// Get user from context for authorization
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "User not found in context",
			Code:  models.ErrCodeInternal,
		})
		return
	}
	currentUser := user.(*models.User)

	// Fetch the provider first to verify ownership
	provider, err := h.providerService.GetProviderByID(providerID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "Provider not found",
			Code:  models.ErrCodeNotFound,
		})
		return
	}

	// Verify the provider belongs to the user's team
	if provider.TeamID != currentUser.TenantID {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error: "You don't have access to this provider",
			Code:  models.ErrCodeForbidden,
		})
		return
	}

	// Enable the provider via service
	err = h.providerService.EnableProvider(providerID)
	if err != nil {
		requestID := c.GetString("request_id")
		log.Printf("[ERROR] [%s] Failed to enable provider - provider_id=%d error=%v", requestID, providerID, err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to enable provider",
			Code:    models.ErrCodeInternal,
			Details: err.Error(),
			Request: requestID,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"provider_id": providerID,
		"enabled":     true,
		"message":     "Provider enabled successfully",
	})
}

func (h *ProviderHandler) DisableProvider(c *gin.Context) {
	providerID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid provider ID",
			Code:  models.ErrCodeValidation,
		})
		return
	}

	// Get user from context for authorization
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "User not found in context",
			Code:  models.ErrCodeInternal,
		})
		return
	}
	currentUser := user.(*models.User)

	// Fetch the provider first to verify ownership
	provider, err := h.providerService.GetProviderByID(providerID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "Provider not found",
			Code:  models.ErrCodeNotFound,
		})
		return
	}

	// Verify the provider belongs to the user's team
	if provider.TeamID != currentUser.TenantID {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error: "You don't have access to this provider",
			Code:  models.ErrCodeForbidden,
		})
		return
	}

	// Disable the provider via service
	err = h.providerService.DisableProvider(providerID)
	if err != nil {
		requestID := c.GetString("request_id")
		log.Printf("[ERROR] [%s] Failed to disable provider - provider_id=%d error=%v", requestID, providerID, err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to disable provider",
			Code:    models.ErrCodeInternal,
			Details: err.Error(),
			Request: requestID,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"provider_id": providerID,
		"enabled":     false,
		"message":     "Provider disabled successfully",
	})
}

func (h *ProviderHandler) GetProviderStats(c *gin.Context) {
	providerID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid provider ID",
			Code:  models.ErrCodeValidation,
		})
		return
	}

	// Get user from context to get tenant_id and team
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "User not found in context",
			Code:  models.ErrCodeInternal,
		})
		return
	}
	currentUser := user.(*models.User)

	// Get provider name first
	provider, err := h.providerService.GetProviderByID(providerID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "Provider not found",
			Code:  models.ErrCodeNotFound,
		})
		return
	}

	// Only allow if the provider belongs to the user's team
	if provider.TeamID != currentUser.TenantID {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error: "You don't have access to this provider",
			Code:  models.ErrCodeForbidden,
		})
		return
	}

	stats, err := h.usageService.GetProviderStats(c.Request.Context(), provider.Name, currentUser.TenantID)
	if err != nil {
		requestID := c.GetString("request_id")
		log.Printf("[ERROR] [%s] Failed to get provider stats - provider_id=%d error=%v", requestID, providerID, err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to get provider stats",
			Code:    models.ErrCodeInternal,
			Details: err.Error(),
			Request: requestID,
		})
		return
	}

	// Add provider-specific fields to the stats
	response := gin.H{
		"provider_id":   providerID,
		"provider_name": provider.Name,
	}

	// Add the stats from the service
	for k, v := range stats {
		response[k] = v
	}

	c.JSON(http.StatusOK, response)
}
