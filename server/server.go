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

package main

import (
	"fmt"
	"log"
	"time"

	"switch-server/config"
	"switch-server/handlers"
	"switch-server/internal/db"
	"switch-server/middleware"
	"switch-server/repository"
	"switch-server/services"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// startServer initializes and starts the server.
// Extracted to enable testing with coverage collection.
func startServer() {
	// Load configuration
	cfg := config.LoadConfig()

	if err := db.RunMigrations(cfg.DatabaseURL); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Initialize database connection
	database, err := db.ConnectDBWithDebug(cfg.DatabaseURL, cfg.Debug)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer database.Close()

	// Initialize RBAC enforcer
	rbacEnforcer, err := middleware.NewEnforcer()
	if err != nil {
		log.Fatal("Failed to initialize RBAC enforcer:", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(database)
	teamRepo := repository.NewTeamRepository(database)
	providerRepo := repository.NewProviderRepository(database)
	usageRepo := repository.NewUsageRepository(database)
	licenseRepo := repository.NewLicenseRepository(database)
	relayTokenRepo := repository.NewRelayTokenRepository(database)

	// Initialize services
	userService := services.NewUserService(userRepo)
	teamService := services.NewTeamService(teamRepo, userRepo, userService)
	providerService := services.NewProviderService(providerRepo)
	usageService := services.NewUsageService(usageRepo)
	licenseService, err := services.NewLicenseService(licenseRepo, cfg.LicensePublicKey)
	if err != nil {
		log.Fatal("Failed to initialize license service:", err)
	}
	relayTokenService := services.NewRelayTokenService(relayTokenRepo)
	relayRateLimiter := services.NewRateLimiter(60, time.Minute) // 60 req/min per token

	// Initialize Gin router
	gin.SetMode(gin.ReleaseMode)
	if cfg.Debug {
		gin.SetMode(gin.DebugMode)
	}

	router := gin.New()
	// Custom middleware for structured logging and error handling
	router.Use(middleware.RequestIDMiddleware())    // Add request IDs for tracking
	router.Use(middleware.ErrorLoggingMiddleware()) // Custom error logging with context
	router.Use(gin.Recovery())                      // Recovery from panics

	// Initialize handlers
	authHandlers := handlers.NewAuthHandler(userService, teamService)
	teamHandlers := handlers.NewTeamHandler(teamService, userService, licenseService)
	providerHandlers := handlers.NewProviderHandler(providerService, usageService)
	usageHandlers := handlers.NewUsageHandler(usageService)
	relayHandlers := handlers.NewRelayHandler(providerService, usageService)
	healthHandlers := handlers.NewHealthHandler(database.Pool(), ServerVersion)
	dashboardHandlers := handlers.NewDashboardHandler(usageService, userService)
	analyticsHandlers := handlers.NewAnalyticsHandler(usageService, userService)
	userHandlers := handlers.NewUserHandler(userService)
	licenseHandlers := handlers.NewLicenseHandler(licenseService)
	relayTokenHandlers := handlers.NewRelayTokenHandler(relayTokenService)

	// Health check endpoint (public)
	router.GET("/health", healthHandlers.HealthCheck)

	// Setup endpoints (public)
	setupHandlers := handlers.NewSetupHandler(database.Pool())
	setup := router.Group("/api/v1/setup")
	{
		setup.GET("/status", setupHandlers.GetSetupStatus)
		setup.POST("/admin", setupHandlers.CreateInitialAdmin)
	}

	// Public routes
	public := router.Group("/auth")
	{
		public.POST("/register", authHandlers.Register)
		public.POST("/login", authHandlers.Login)
		public.POST("/refresh", authHandlers.Refresh)
		public.POST("/verify", authHandlers.Verify)
	}

	// Relay proxy endpoints — separate group with custom auth chain:
	// 1. RelayTokenAuthMiddleware (validates relay tokens, falls through for JWT)
	// 2. AuthMiddleware (JWT session auth — catches JWT tokens and rejects invalid relay tokens)
	// 3. LicenseMiddleware
	relayGroup := router.Group("/api/v1/relay")
	relayGroup.Use(middleware.RelayTokenAuthMiddleware(relayTokenService, relayRateLimiter))
	relayGroup.Use(middleware.AuthMiddleware())
	relayGroup.Use(middleware.LicenseMiddleware(licenseService))
	relayGroup.POST("/:tool/v1/messages", relayHandlers.RelayMessages)
	relayGroup.POST("/:tool/v1/chat/completions", relayHandlers.RelayChatCompletions)

	// Protected routes
	protected := router.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware())
	protected.Use(middleware.LicenseMiddleware(licenseService))
	{
		// Auth
		protected.POST("/logout", authHandlers.Logout)

		// User profile
		protected.GET("/user/profile", authHandlers.GetProfile)
		protected.PUT("/user/profile", authHandlers.UpdateProfile)
		protected.PUT("/user/password", authHandlers.ChangePassword)

		// Relay token management — member-facing endpoints (any authenticated user)
		protected.POST("/user/relay-token", relayTokenHandlers.MyRelayToken)
		protected.GET("/user/relay-token", relayTokenHandlers.GetMyRelayTokenInfo)
		protected.DELETE("/user/relay-token", relayTokenHandlers.RevokeMyRelayToken)

		// Team management
		protected.GET("/teams", teamHandlers.ListTeams)
		protected.POST("/teams", teamHandlers.CreateTeam)
		protected.PUT("/teams/:id", teamHandlers.UpdateTeam)
		protected.DELETE("/teams/:id", teamHandlers.DeleteTeam)

		protected.GET("/teams/:id/members", teamHandlers.ListTeamMembers)
		protected.POST("/teams/:id/members", teamHandlers.AddTeamMember)
		protected.DELETE("/teams/:id/members/:memberId", teamHandlers.RemoveTeamMember)

		protected.GET("/teams/:id/settings", teamHandlers.GetTeamSettings)
		protected.PUT("/teams/:id/settings", teamHandlers.UpdateTeamSettings)

		// Provider management - read operations (members and managers)
		protected.GET("/providers", middleware.RequirePermission(rbacEnforcer, "providers", "read"), providerHandlers.ListProviders)
		protected.GET("/providers/:id/stats", middleware.RequirePermission(rbacEnforcer, "providers", "read"), providerHandlers.GetProviderStats)

		// Provider management - write operations (managers only)
		protected.POST("/providers", middleware.RequirePermission(rbacEnforcer, "providers", "write"), providerHandlers.CreateProvider)
		protected.PUT("/providers/:id", middleware.RequirePermission(rbacEnforcer, "providers", "write"), providerHandlers.UpdateProvider)
		protected.DELETE("/providers/:id", middleware.RequirePermission(rbacEnforcer, "providers", "write"), providerHandlers.DeleteProvider)
		protected.POST("/providers/:id/test", middleware.RequirePermission(rbacEnforcer, "providers", "write"), providerHandlers.TestProvider)
		protected.POST("/providers/:id/enable", middleware.RequirePermission(rbacEnforcer, "providers", "write"), providerHandlers.EnableProvider)
		protected.DELETE("/providers/:id/disable", middleware.RequirePermission(rbacEnforcer, "providers", "write"), providerHandlers.DisableProvider)

		// Usage tracking
		protected.GET("/usage/current", usageHandlers.GetCurrentUsage)
		protected.GET("/usage/stats", usageHandlers.GetUsageStats)
		protected.POST("/usage/batch", usageHandlers.CreateBatchUsageRecords)

		// Dashboard
		protected.GET("/dashboard/metrics", dashboardHandlers.GetMetrics)
		protected.GET("/dashboard/rankings", dashboardHandlers.GetRankings)
		protected.GET("/dashboard/members", dashboardHandlers.GetMembers)

		// Analytics (managers only)
		protected.GET("/analytics/providers", middleware.RequirePermission(rbacEnforcer, "analytics", "read"), analyticsHandlers.GetProviderAnalytics)
		protected.GET("/analytics/users", middleware.RequirePermission(rbacEnforcer, "analytics", "read"), analyticsHandlers.GetUserAnalytics)
		protected.GET("/analytics/history", middleware.RequirePermission(rbacEnforcer, "analytics", "read"), analyticsHandlers.GetHistory)
		protected.GET("/analytics/filters", analyticsHandlers.GetFilterOptions)
		// Personal analytics (any authenticated user)
		protected.GET("/analytics/personal", analyticsHandlers.GetPersonalAnalytics)
		protected.GET("/analytics/personal/history", analyticsHandlers.GetPersonalHistory)
		protected.GET("/analytics/personal/filters", analyticsHandlers.GetPersonalFilterOptions)

		// User management (managers only)
		protected.GET("/users", middleware.RequirePermission(rbacEnforcer, "users", "read"), userHandlers.ListUsers)
		protected.POST("/users", middleware.RequirePermission(rbacEnforcer, "users", "write"), userHandlers.CreateUser)
		protected.PUT("/users/:id", middleware.RequirePermission(rbacEnforcer, "users", "write"), userHandlers.UpdateUser)
		protected.DELETE("/users/:id", middleware.RequirePermission(rbacEnforcer, "users", "write"), userHandlers.DeleteUser)

		// License management - members can read, managers can write
		protected.GET("/license", middleware.RequirePermission(rbacEnforcer, "license", "read"), licenseHandlers.GetLicense)
		protected.POST("/license/activate", middleware.RequirePermission(rbacEnforcer, "license", "write"), licenseHandlers.ActivateLicense)
		// Public tier info
		router.GET("/tiers", licenseHandlers.GetTiers)

	}

	// Start the server
	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("Server starting on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func main() {
	startServer()
}
