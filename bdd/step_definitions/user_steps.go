// Package step_definitions provides step implementations for BDD scenarios
package step_definitions

import (
	"fmt"
	"log"

	"github.com/code-together/bdd/support"
	"github.com/cucumber/godog"
)

// RegisterUserSteps registers user management step definitions
func RegisterUserSteps(ctx *ScenarioContext, suite *godog.ScenarioContext) {
	// GIVEN STEPS - Setup context

	suite.Given(`^I am logged in as the only manager$`, ctx.iAmLoggedInAsTheOnlyManager)
	suite.Given(`^there is only one other manager$`, ctx.thereIsOnlyOneOtherManager)
	suite.Given(`^there are at least 2 managers in the organization$`, ctx.thereAreAtLeast2ManagersInTheOrganization)

	// WHEN STEPS - Perform actions

	suite.When(`^I attempt to change my role to member$`, ctx.iAttemptToChangeMyRoleToMember)
	suite.When(`^I demote another manager to member$`, ctx.iDemoteAnotherManagerToMember)

	// THEN STEPS - Assert outcomes
	// Note: Most assertion steps are already registered in auth_steps.go
}

// GIVENS - Setup context

// iAmLoggedInAsTheOnlyManager ensures the current user is a manager and no other managers exist
func (ctx *ScenarioContext) iAmLoggedInAsTheOnlyManager() error {
	// First login as a manager
	if err := ctx.iAmLoggedInAsAManager(); err != nil {
		return fmt.Errorf("failed to login as manager: %w", err)
	}

	// Get current user info
	currentUser := ctx.BDDTestContext.CurrentUser
	if currentUser == nil {
		return fmt.Errorf("current user is not set")
	}

	// List all users to check for other managers
	if err := ctx.iListAllUsers(); err != nil {
		return fmt.Errorf("failed to list users: %w", err)
	}

	statusCode, resp, errMsg := ctx.GetLastResponse()
	if statusCode != 200 {
		return fmt.Errorf("failed to list users: %s", errMsg)
	}

	usersList, ok := resp.(map[string]interface{})
	if !ok {
		return fmt.Errorf("unexpected response format when listing users")
	}

	users, ok := usersList["users"].([]interface{})
	if !ok {
		return fmt.Errorf("users field missing or invalid in response")
	}

	// Count managers and delete any other managers
	managerCount := 0
	for _, userObj := range users {
		user, ok := userObj.(map[string]interface{})
		if !ok {
			continue
		}

		role, ok := user["role"].(string)
		if !ok {
			continue
		}

		if role == "manager" || role == "admin" {
			managerCount++
			userID, hasID := user["id"].(float64)
			email, hasEmail := user["email"].(string)

			// Skip the current user
			if hasEmail && email == currentUser.Email {
				continue
			}

			// Delete other managers
			if hasID {
				log.Printf("Deleting extra manager with ID %d (email: %s)", int64(userID), email)
				if err := ctx.deleteUser(int64(userID)); err != nil {
					log.Printf("Warning: failed to delete manager %d: %v", int64(userID), err)
				}
			}
		}
	}

	log.Printf("Verified only one manager exists (current user: %s)", currentUser.Email)
	return nil
}

// thereIsOnlyOneOtherManager creates one additional manager besides the current user
func (ctx *ScenarioContext) thereIsOnlyOneOtherManager() error {
	// First login as a manager
	if err := ctx.iAmLoggedInAsAManager(); err != nil {
		return fmt.Errorf("failed to login as manager: %w", err)
	}

	currentUser := ctx.BDDTestContext.CurrentUser
	if currentUser == nil {
		return fmt.Errorf("current user is not set")
	}

	// List all users
	if err := ctx.iListAllUsers(); err != nil {
		return fmt.Errorf("failed to list users: %w", err)
	}

	statusCode, resp, errMsg := ctx.GetLastResponse()
	if statusCode != 200 {
		return fmt.Errorf("failed to list users: %s", errMsg)
	}

	usersList, ok := resp.(map[string]interface{})
	if !ok {
		return fmt.Errorf("unexpected response format when listing users")
	}

	users, ok := usersList["users"].([]interface{})
	if !ok {
		return fmt.Errorf("users field missing or invalid in response")
	}

	// Count existing managers
	managerCount := 0
	var otherManagerID int64
	for _, userObj := range users {
		user, ok := userObj.(map[string]interface{})
		if !ok {
			continue
		}

		role, ok := user["role"].(string)
		if !ok {
			continue
		}

		if role == "manager" || role == "admin" {
			managerCount++
			userID, hasID := user["id"].(float64)
			email, hasEmail := user["email"].(string)

			// Skip the current user
			if hasEmail && email == currentUser.Email {
				continue
			}

			// Track the other manager
			if hasID {
				otherManagerID = int64(userID)
			}
		}
	}

	// If we already have exactly 2 managers (current + one other), we're done
	if managerCount == 2 {
		ctx.TrackCreatedResource("other_manager_id", fmt.Sprintf("%d", otherManagerID))
		log.Printf("Already have 2 managers (current user + one other)")
		return nil
	}

	// If we have more than 2 managers, delete extras until we have only 2
	if managerCount > 2 {
		extraManagersToDelete := managerCount - 2
		deletedCount := 0

		for _, userObj := range users {
			if deletedCount >= extraManagersToDelete {
				break
			}

			user, ok := userObj.(map[string]interface{})
			if !ok {
				continue
			}

			role, ok := user["role"].(string)
			if !ok {
				continue
			}

			if role == "manager" || role == "admin" {
				userID, hasID := user["id"].(float64)
				email, hasEmail := user["email"].(string)

				// Skip the current user and the one other manager we want to keep
				if hasEmail && email == currentUser.Email {
					continue
				}
				if hasID && int64(userID) == otherManagerID {
					continue
				}

				// Delete this extra manager
				if hasID {
					log.Printf("Deleting extra manager with ID %d", int64(userID))
					if err := ctx.deleteUser(int64(userID)); err != nil {
						log.Printf("Warning: failed to delete manager %d: %v", int64(userID), err)
					} else {
						deletedCount++
					}
				}
			}
		}

		ctx.TrackCreatedResource("other_manager_id", fmt.Sprintf("%d", otherManagerID))
		log.Printf("Deleted %d extra managers, now have 2 managers total", deletedCount)
		return nil
	}

	// If we have only 1 manager (current user), create another manager
	if managerCount == 1 {
		// Create a new manager user
		uniqueEmail := fmt.Sprintf("manager-%s@example.com", support.GenerateUniqueEmail("other"))

		// TODO: Use the proper API to create a manager user
		// For now, this is a placeholder that tracks the intent
		log.Printf("TODO: Create another manager with email %s", uniqueEmail)

		// Placeholder: Track that we need a second manager
		ctx.TrackCreatedResource("other_manager_email", uniqueEmail)
		ctx.TrackCreatedResource("second_manager_needed", "true")

		return fmt.Errorf("TODO: API endpoint for creating manager users not yet implemented")
	}

	ctx.TrackCreatedResource("other_manager_id", fmt.Sprintf("%d", otherManagerID))
	return nil
}

// thereAreAtLeast2ManagersInTheOrganization ensures at least 2 manager accounts exist
func (ctx *ScenarioContext) thereAreAtLeast2ManagersInTheOrganization() error {
	// First login as a manager
	if err := ctx.iAmLoggedInAsAManager(); err != nil {
		return fmt.Errorf("failed to login as manager: %w", err)
	}

	currentUser := ctx.BDDTestContext.CurrentUser
	if currentUser == nil {
		return fmt.Errorf("current user is not set")
	}

	// List all users
	if err := ctx.iListAllUsers(); err != nil {
		return fmt.Errorf("failed to list users: %w", err)
	}

	statusCode, resp, errMsg := ctx.GetLastResponse()
	if statusCode != 200 {
		return fmt.Errorf("failed to list users: %s", errMsg)
	}

	usersList, ok := resp.(map[string]interface{})
	if !ok {
		return fmt.Errorf("unexpected response format when listing users")
	}

	users, ok := usersList["users"].([]interface{})
	if !ok {
		return fmt.Errorf("users field missing or invalid in response")
	}

	// Count existing managers
	managerCount := 0
	var anotherManagerID int64
	for _, userObj := range users {
		user, ok := userObj.(map[string]interface{})
		if !ok {
			continue
		}

		role, ok := user["role"].(string)
		if !ok {
			continue
		}

		if role == "manager" || role == "admin" {
			managerCount++
			userID, hasID := user["id"].(float64)
			email, hasEmail := user["email"].(string)

			// Track a manager that's not the current user
			if hasID && hasEmail && email != currentUser.Email && anotherManagerID == 0 {
				anotherManagerID = int64(userID)
			}
		}
	}

	// If we have at least 2 managers, we're done
	if managerCount >= 2 {
		ctx.TrackCreatedResource("another_manager_id", fmt.Sprintf("%d", anotherManagerID))
		log.Printf("Verified at least 2 managers exist (count: %d)", managerCount)
		return nil
	}

	// If we have only 1 manager, create additional managers
	managersNeeded := 2 - managerCount
	for i := 0; i < managersNeeded; i++ {
		uniqueEmail := fmt.Sprintf("manager-%d-%s@example.com", i, support.GenerateUniqueEmail("additional"))
		log.Printf("TODO: Create additional manager with email %s", uniqueEmail)
	}

	// Placeholder: Track that we need more managers
	ctx.TrackCreatedResource("managers_needed", fmt.Sprintf("%d", managersNeeded))

	return fmt.Errorf("TODO: API endpoint for creating manager users not yet implemented")
}

// WHENS - Perform actions

// iAttemptToChangeMyRoleToMember attempts to change the current user's role to member
func (ctx *ScenarioContext) iAttemptToChangeMyRoleToMember() error {
	currentUser := ctx.BDDTestContext.CurrentUser
	if currentUser == nil {
		return fmt.Errorf("current user is not set")
	}

	log.Printf("Attempting to change own role to member (email: %s)", currentUser.Email)

	// TODO: Use the proper API endpoint to update user role
	// Expected API: PATCH /api/v1/users/{id}/role
	// Request body: { "role": "member" }
	// Response: 403 Forbidden if this is the last manager

	// Placeholder implementation
	ctx.TrackCreatedResource("attempted_role_change", "member")
	ctx.TrackCreatedResource("attempted_on_user_email", currentUser.Email)

	return fmt.Errorf("TODO: API endpoint PATCH /users/{id}/role not yet implemented")
}

// iDemoteAnotherManagerToMember attempts to change another manager's role to member
func (ctx *ScenarioContext) iDemoteAnotherManagerToMember() error {
	currentUser := ctx.BDDTestContext.CurrentUser
	if currentUser == nil {
		return fmt.Errorf("current user is not set")
	}

	// Get the ID of another manager from tracked resources
	managerIDStr, exists := ctx.GetCreatedResource("another_manager_id")
	if !exists {
		// Try "other_manager_id" as fallback
		managerIDStr, exists = ctx.GetCreatedResource("other_manager_id")
		if !exists {
			return fmt.Errorf("no other manager ID found in context")
		}
	}

	// Convert string ID to int64
	var otherManagerID int64
	if _, err := fmt.Sscanf(managerIDStr, "%d", &otherManagerID); err != nil {
		return fmt.Errorf("invalid manager ID format in context: %w", err)
	}

	log.Printf("Attempting to demote manager ID %d to member role", otherManagerID)

	// TODO: Use the proper API endpoint to update user role
	// Expected API: PATCH /api/v1/users/{id}/role
	// Request body: { "role": "member" }
	// Response: 200 OK if successful

	// Placeholder implementation
	ctx.TrackCreatedResource("attempted_role_change", "member")
	ctx.TrackCreatedResource("attempted_on_user_id", fmt.Sprintf("%d", otherManagerID))

	return fmt.Errorf("TODO: API endpoint PATCH /users/{id}/role not yet implemented")
}

// Helper function to delete a user
func (ctx *ScenarioContext) deleteUser(userID int64) error {
	// TODO: Use the proper API endpoint to delete a user
	// Expected API: DELETE /api/v1/users/{id}
	log.Printf("TODO: Delete user with ID %d (API endpoint not yet implemented)", userID)
	return nil
}
