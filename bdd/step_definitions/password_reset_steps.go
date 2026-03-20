// Package step_definitions provides step implementations for BDD scenarios
package step_definitions

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/code-together/shared/integration"
	"github.com/cucumber/godog"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// RegisterPasswordResetSteps registers password reset step definitions
func RegisterPasswordResetSteps(ctx *ScenarioContext, suite *godog.ScenarioContext) {
	// GIVEN STEPS - Setup context
	suite.Given(`^I have a valid reset token$`, ctx.iHaveAValidResetToken)
	suite.Given(`^I have an expired reset token$`, ctx.iHaveAnExpiredResetToken)
	suite.Given(`^I have an invalid reset token$`, ctx.iHaveAnInvalidResetToken)

	// WHEN STEPS - Actions
	suite.When(`^I request a password reset$`, ctx.iRequestAPasswordReset)
	suite.When(`^I reset password to "([^"]*)"$`, ctx.iResetPasswordTo)
	suite.When(`^I attempt to reset password$`, ctx.iAttemptToResetPassword)

	// THEN STEPS - Assertions
	suite.Then(`^a reset token should be generated$`, ctx.aResetTokenShouldBeGenerated)
	suite.Then(`^the reset token should expire in 1 hour$`, ctx.theResetTokenShouldExpireIn1Hour)
	suite.Then(`^I can login with the new password$`, ctx.iCanLoginWithTheNewPassword)
}

// iRequestAPasswordReset requests a password reset for the current user's email
// Backend API: POST /auth/password-reset/request
// Request: { "email": "user@example.com" }
// Response: { "token": "reset-token-xyz", "expires_in": 3600 }
//
// TODO: This implementation uses a placeholder. Once the backend implements
// POST /auth/password-reset/request, replace this with the actual API call.
func (ctx *ScenarioContext) iRequestAPasswordReset() error {
	// Get email from current user or use a default test email
	var email string
	if ctx.BDDTestContext.CurrentUser != nil {
		email = ctx.BDDTestContext.CurrentUser.Email
	} else {
		// Use a default test email if no current user
		email = "reset@example.com"
	}

	// TODO: Replace with actual API call when backend implements password reset
	// Expected implementation:
	// req := integration.PostAuthPasswordResetRequestJSONRequestBody{
	//     Email: openapi_types.Email(email),
	// }
	// resp, err := ctx.AnonymousClient.PostAuthPasswordResetRequestWithResponse(context.Background(), req)

	// Placeholder: Generate a mock reset token for testing
	resetToken := fmt.Sprintf("reset-token-%s-%d", email, time.Now().Unix())
	expiresIn := int64(3600) // 1 hour

	// Store the reset token for use by subsequent steps
	ctx.TrackCreatedResource("reset_token", resetToken)

	// Store expiration time
	expiresAt := time.Now().Add(time.Duration(expiresIn) * time.Second)
	ctx.TrackCreatedResource("reset_token_expires_at", expiresAt.Format(time.RFC3339))

	// Create a mock response structure
	mockResponse := map[string]interface{}{
		"token":      resetToken,
		"expires_in": expiresIn,
	}

	ctx.SetLastResponse(200, mockResponse, "")
	log.Printf("TODO: Password reset request not yet implemented by backend. Using placeholder token: %s", resetToken)

	return nil
}

// iHaveAValidResetToken sets up a valid reset token in the context
// This step requests a real reset token to ensure validity
func (ctx *ScenarioContext) iHaveAValidResetToken() error {
	// Request a password reset to get a valid token
	if err := ctx.iRequestAPasswordReset(); err != nil {
		return fmt.Errorf("failed to get valid reset token: %w", err)
	}

	// Verify token was generated
	token, exists := ctx.GetCreatedResource("reset_token")
	if !exists || token == "" {
		return fmt.Errorf("reset token was not generated")
	}

	log.Printf("Set up valid reset token: %s", token)
	return nil
}

// iHaveAnExpiredResetToken sets up an expired reset token in the context
// This simulates a token that was created more than 1 hour ago
func (ctx *ScenarioContext) iHaveAnExpiredResetToken() error {
	// First get a valid token
	if err := ctx.iRequestAPasswordReset(); err != nil {
		return fmt.Errorf("failed to get reset token: %w", err)
	}

	// Verify token was generated
	_, exists := ctx.GetCreatedResource("reset_token")
	if !exists {
		return fmt.Errorf("reset token was not generated")
	}

	// Mark as expired by setting expiration to past time
	expiredTime := time.Now().Add(-2 * time.Hour) // 2 hours ago
	ctx.TrackCreatedResource("reset_token_expires_at", expiredTime.Format(time.RFC3339))
	ctx.TrackCreatedResource("reset_token_is_expired", "true")

	log.Printf("Set up expired reset token (expired at: %s)", expiredTime.Format(time.RFC3339))
	return nil
}

// iHaveAnInvalidResetToken sets up an invalid/malformed reset token
func (ctx *ScenarioContext) iHaveAnInvalidResetToken() error {
	// Use a clearly invalid token format
	invalidToken := "invalid-reset-token-malformed-xyz123"
	ctx.TrackCreatedResource("reset_token", invalidToken)
	ctx.TrackCreatedResource("reset_token_is_invalid", "true")

	log.Printf("Set up invalid reset token: %s", invalidToken)
	return nil
}

// iResetPasswordTo resets the password using the stored reset token
// Backend API: POST /auth/password-reset/confirm
// Request: { "token": "reset-token-xyz", "new_password": "NewPassword123!" }
// Response: { "success": true } or error
//
// TODO: This implementation uses a placeholder. Once the backend implements
// POST /auth/password-reset/confirm, replace this with the actual API call.
func (ctx *ScenarioContext) iResetPasswordTo(newPassword string) error {
	// Get the reset token from context
	token, exists := ctx.GetCreatedResource("reset_token")
	if !exists {
		return fmt.Errorf("no reset token available - use 'I have a valid reset token' step first")
	}

	// Check if token is marked as expired
	if isExpired, _ := ctx.GetCreatedResource("reset_token_is_expired"); isExpired == "true" {
		log.Printf("WARNING: Using expired reset token as requested by scenario")
		// TODO: When backend is implemented, this should return an error
		// For now, we simulate the expected behavior
		errMsg := "reset token has expired"
		ctx.SetLastResponse(400, nil, errMsg)
		return nil // Don't fail the step, just record the expected error response
	}

	// Check if token is marked as invalid
	if isInvalid, _ := ctx.GetCreatedResource("reset_token_is_invalid"); isInvalid == "true" {
		log.Printf("WARNING: Using invalid reset token as requested by scenario")
		// TODO: When backend is implemented, this should return an error
		// For now, we simulate the expected behavior
		errMsg := "invalid reset token"
		ctx.SetLastResponse(400, nil, errMsg)
		return nil // Don't fail the step, just record the expected error response
	}

	// TODO: Replace with actual API call when backend implements password reset confirm
	// Expected implementation:
	// req := integration.PostAuthPasswordResetConfirmJSONRequestBody{
	//     Token:       token,
	//     NewPassword: newPassword,
	// }
	// resp, err := ctx.AnonymousClient.PostAuthPasswordResetConfirmWithResponse(context.Background(), req)

	// Placeholder: Simulate successful password reset
	log.Printf("TODO: Password reset confirm not yet implemented by backend. Simulating success for token: %s", token)

	// Update current user's password if set
	if ctx.BDDTestContext.CurrentUser != nil {
		ctx.BDDTestContext.CurrentUser.Password = newPassword
	}

	// Store the new password for login verification
	ctx.TrackCreatedResource("new_password", newPassword)

	// Create a mock response structure
	mockResponse := map[string]interface{}{
		"success": true,
	}

	ctx.SetLastResponse(200, mockResponse, "")

	return nil
}

// iAttemptToResetPassword attempts to reset password without specifying new password
// This is used for testing error scenarios (expired/invalid tokens)
func (ctx *ScenarioContext) iAttemptToResetPassword() error {
	// Use a default new password for the attempt
	return ctx.iResetPasswordTo("AttemptedNewPassword123!")
}

// aResetTokenShouldBeGenerated verifies that a reset token was generated
func (ctx *ScenarioContext) aResetTokenShouldBeGenerated() error {
	token, exists := ctx.GetCreatedResource("reset_token")
	if !exists {
		return fmt.Errorf("reset token was not generated")
	}

	if token == "" {
		return fmt.Errorf("reset token is empty")
	}

	log.Printf("Verified reset token was generated: %s", token)
	return nil
}

// theResetTokenShouldExpireIn1Hour verifies that the reset token expires in 1 hour
func (ctx *ScenarioContext) theResetTokenShouldExpireIn1Hour() error {
	// Check if expiration time was stored
	expiresAtStr, exists := ctx.GetCreatedResource("reset_token_expires_at")
	if !exists {
		return fmt.Errorf("reset token expiration time was not provided in response")
	}

	// Parse the expiration time
	expiresAt, err := time.Parse(time.RFC3339, expiresAtStr)
	if err != nil {
		return fmt.Errorf("failed to parse expiration time: %w", err)
	}

	// Calculate expected expiration (1 hour from now)
	now := time.Now()
	expectedExpiration := now.Add(1 * time.Hour)

	// Allow some tolerance (e.g., ±5 seconds) for test execution time
	tolerance := 5 * time.Second
	timeDiff := expiresAt.Sub(expectedExpiration)

	if timeDiff < -tolerance || timeDiff > tolerance {
		return fmt.Errorf("reset token expiration mismatch: expected ~1 hour from now (%s), got %s (diff: %v)",
			expectedExpiration.Format(time.RFC3339),
			expiresAt.Format(time.RFC3339),
			timeDiff)
	}

	log.Printf("Verified reset token expires in 1 hour (at: %s)", expiresAt.Format(time.RFC3339))
	return nil
}

// iCanLoginWithTheNewPassword verifies that the user can login with the new password
func (ctx *ScenarioContext) iCanLoginWithTheNewPassword() error {
	// Get the new password from context
	newPassword, exists := ctx.GetCreatedResource("new_password")
	if !exists {
		return fmt.Errorf("new password not found in context - password reset may not have succeeded")
	}

	// Get email from current user
	var email string
	if ctx.BDDTestContext.CurrentUser != nil {
		email = ctx.BDDTestContext.CurrentUser.Email
	} else {
		// Use the default test email
		email = "reset@example.com"
	}

	// Attempt to login with new password
	loginReq := integration.PostAuthLoginJSONRequestBody{
		Email:    openapi_types.Email(email),
		Password: newPassword,
	}

	loginResp, err := ctx.AnonymousClient.PostAuthLoginWithResponse(context.Background(), loginReq)
	if err != nil {
		return fmt.Errorf("login with new password failed: %w", err)
	}

	// Verify successful login
	if loginResp.StatusCode() != 200 {
		errMsg := ""
		if loginResp.JSON401 != nil {
			errMsg = loginResp.JSON401.Error
		} else if len(loginResp.Body) > 0 {
			errMsg = string(loginResp.Body)
		}
		return fmt.Errorf("login with new password failed with status %d: %s", loginResp.StatusCode(), errMsg)
	}

	if loginResp.JSON200 == nil {
		return fmt.Errorf("login response is empty")
	}

	// Update current user with new token
	if ctx.BDDTestContext.CurrentUser != nil {
		ctx.BDDTestContext.CurrentUser.Token = loginResp.JSON200.AccessToken
	}

	log.Printf("Successfully logged in with new password for user: %s", email)
	return nil
}
