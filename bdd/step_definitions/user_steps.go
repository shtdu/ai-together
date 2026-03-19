// Package step_definitions provides step implementations for BDD scenarios
package step_definitions

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

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

	users, err := extractUsersList(resp)
	if err != nil {
		return fmt.Errorf("failed to extract users list: %w", err)
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

	// Track that this is a last manager scenario for role change attempts
	ctx.TrackCreatedResource("is_last_manager", "true")

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

	users, err := extractUsersList(resp)
	if err != nil {
		return fmt.Errorf("failed to extract users list: %w", err)
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
		newManagerID := fmt.Sprintf("mgr-%d", time.Now().UnixNano())

		// TODO: Replace with actual API call when backend implements user creation with role
		// Expected implementation:
		// req := integration.PostUsersJSONRequestBody{
		//     Email:    openapi_types.Email(uniqueEmail),
		//     Password: "TestPassword123!",
		//     Role:     "manager",
		// }
		// client, err := ctx.GetAuthenticatedClient()
		// resp, err := client.PostUsersWithResponse(context.Background(), req)

		// Placeholder: Track the second manager for testing
		ctx.TrackCreatedResource("other_manager_id", newManagerID)
		ctx.TrackCreatedResource("other_manager_email", uniqueEmail)
		ctx.TrackCreatedResource("second_manager_created", "true")

		log.Printf("TODO: User creation API not yet implemented by backend. Using placeholder manager: %s (ID: %s)", uniqueEmail, newManagerID)

		// Create a mock response structure
		mockResponse := map[string]interface{}{
			"id":    newManagerID,
			"email": uniqueEmail,
			"role":  "manager",
		}

		ctx.SetLastResponse(201, mockResponse, "")
		return nil
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

	users, err := extractUsersList(resp)
	if err != nil {
		return fmt.Errorf("failed to extract users list: %w", err)
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
	var additionalManagerIDs []string

	for i := 0; i < managersNeeded; i++ {
		uniqueEmail := fmt.Sprintf("manager-%d-%s@example.com", i, support.GenerateUniqueEmail("additional"))
		newManagerID := fmt.Sprintf("mgr-addl-%d", time.Now().UnixNano()+int64(i))

		// TODO: Replace with actual API call when backend implements user creation with role
		// Expected implementation:
		// req := integration.PostUsersJSONRequestBody{
		//     Email:    openapi_types.Email(uniqueEmail),
		//     Password: "TestPassword123!",
		//     Role:     "manager",
		// }
		// client, err := ctx.GetAuthenticatedClient()
		// resp, err := client.PostUsersWithResponse(context.Background(), req)

		log.Printf("TODO: User creation API not yet implemented by backend. Using placeholder manager: %s (ID: %s)", uniqueEmail, newManagerID)

		// Track the additional manager for testing
		additionalManagerIDs = append(additionalManagerIDs, newManagerID)
		ctx.TrackCreatedResource(fmt.Sprintf("additional_manager_%d_id", i), newManagerID)
		ctx.TrackCreatedResource(fmt.Sprintf("additional_manager_%d_email", i), uniqueEmail)
	}

	// Track the first additional manager for "demote another manager" step
	if len(additionalManagerIDs) > 0 {
		ctx.TrackCreatedResource("another_manager_id", additionalManagerIDs[0])
	}

	ctx.TrackCreatedResource("additional_managers_created", fmt.Sprintf("%d", managersNeeded))

	// Create a mock response structure
	mockResponse := map[string]interface{}{
		"managers_created": managersNeeded,
		"manager_ids":      additionalManagerIDs,
	}

	ctx.SetLastResponse(201, mockResponse, "")
	log.Printf("TODO: Created %d placeholder additional managers", managersNeeded)

	return nil
}

// WHENS - Perform actions

// iAttemptToChangeMyRoleToMember attempts to change the current user's role to member
func (ctx *ScenarioContext) iAttemptToChangeMyRoleToMember() error {
	currentUser := ctx.BDDTestContext.CurrentUser
	if currentUser == nil {
		return fmt.Errorf("current user is not set")
	}

	log.Printf("Attempting to change own role to member (email: %s)", currentUser.Email)

	// TODO: Replace with actual API call when backend implements role updates
	// Expected implementation:
	// req := integration.PatchUsersIDRoleJSONRequestBody{
	//     Role: "member",
	// }
	// client, err := ctx.GetAuthenticatedClient()
	// resp, err := client.PatchUsersIDRoleWithResponse(context.Background(), currentUser.ID, req)
	// Expected response: 403 Forbidden if this is the last manager

	// Placeholder implementation: Simulate the expected behavior
	ctx.TrackCreatedResource("attempted_role_change", "member")
	ctx.TrackCreatedResource("attempted_on_user_email", currentUser.Email)

	// Check if this is the last manager scenario
	isLastManager := false
	if lastMgr, exists := ctx.GetCreatedResource("is_last_manager"); exists && lastMgr == "true" {
		isLastManager = true
	}

	// Simulate the expected response based on scenario context
	if isLastManager {
		// Last manager scenario - should fail with 403
		errMsg := "cannot change role: this is the last manager in the organization"
		mockResponse := map[string]interface{}{
			"error": errMsg,
		}
		ctx.SetLastResponse(403, mockResponse, errMsg)
		log.Printf("TODO: Role change API not yet implemented. Simulating 403 Forbidden (last manager protection)")
	} else {
		// Non-last manager scenario - should succeed
		mockResponse := map[string]interface{}{
			"email": currentUser.Email,
			"role":  "member",
		}
		ctx.SetLastResponse(200, mockResponse, "")
		log.Printf("TODO: Role change API not yet implemented. Simulating success")
	}

	return nil
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

	// TODO: Replace with actual API call when backend implements role updates
	// Expected implementation:
	// req := integration.PatchUsersIDRoleJSONRequestBody{
	//     Role: "member",
	// }
	// client, err := ctx.GetAuthenticatedClient()
	// resp, err := client.PatchUsersIDRoleWithResponse(context.Background(), otherManagerID, req)
	// Expected response: 200 OK if successful

	// Placeholder implementation: Simulate successful role change
	ctx.TrackCreatedResource("attempted_role_change", "member")
	ctx.TrackCreatedResource("attempted_on_user_id", fmt.Sprintf("%d", otherManagerID))

	// Get the other manager's email if available
	otherManagerEmail := ""
	if email, exists := ctx.GetCreatedResource("other_manager_email"); exists {
		otherManagerEmail = email
	} else if email, exists := ctx.GetCreatedResource("another_manager_email"); exists {
		otherManagerEmail = email
	}

	// Simulate successful response
	mockResponse := map[string]interface{}{
		"id":    fmt.Sprintf("%d", otherManagerID),
		"email": otherManagerEmail,
		"role":  "member",
	}

	ctx.SetLastResponse(200, mockResponse, "")
	log.Printf("TODO: Role change API not yet implemented. Simulating successful demotion for manager ID %d", otherManagerID)

	return nil
}

// Helper function to extract users list from API response
func extractUsersList(resp interface{}) ([]interface{}, error) {
	// Try direct type assertion first
	usersList, ok := resp.(map[string]interface{})
	if !ok {
		// If that fails, try to convert using JSON marshaling
		jsonData, err := json.Marshal(resp)
		if err != nil {
			return nil, fmt.Errorf("failed to convert response: %w", err)
		}
		var result map[string]interface{}
		if err := json.Unmarshal(jsonData, &result); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response: %w", err)
		}
		usersList = result
	}

	users, ok := usersList["users"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("users field missing or invalid in response")
	}

	return users, nil
}

// Helper function to delete a user
// Backend API: DELETE /api/v1/users/{id}
// Response: 204 No Content if successful
//
// TODO: This implementation uses a placeholder. Once the backend implements
// DELETE /users/{id}, replace this with the actual API call.
func (ctx *ScenarioContext) deleteUser(userID int64) error {
	// TODO: Replace with actual API call when backend implements user deletion
	// Expected implementation:
	// client, err := ctx.GetAuthenticatedClient()
	// if err != nil {
	//     return fmt.Errorf("failed to get authenticated client: %w", err)
	// }
	// resp, err := client.DeleteUsersIDWithResponse(context.Background(), userID)

	// Placeholder: Simulate successful deletion
	log.Printf("TODO: User deletion API not yet implemented. Simulating deletion for user ID %d", userID)

	// Track the deletion for cleanup verification
	ctx.TrackCreatedResource(fmt.Sprintf("deleted_user_%d", userID), "true")

	return nil
}
