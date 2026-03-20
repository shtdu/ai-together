# BDD Test Suite - Provider Management
# This feature covers provider CRUD operations, limits, validation, and connectivity testing

Feature: Provider Management
  As a manager
  I want to configure and manage AI service providers
  So that my team can use AI tools with automatic failover

  Background:
    Given the test server is running
    And I am logged in as a manager
    And the standard fixtures are loaded

  Rule: Provider Creation

    @p0 @requirement:PM-02-001
    Scenario Outline: Create provider with valid data
      Given I have a unique provider name
      When I create a <kind> provider with API key "sk-test-123"
      Then the provider should be created successfully
      And the provider kind should be "<kind>"
      And the provider should be enabled

      Examples:
        | kind    |
        | claude  |
        | codex   |
        | opencode |

    @p1 @requirement:PM-02-002
    Scenario: Create provider with duplicate name in same kind
      Given I have a unique provider name "my-provider"
      And I create a claude provider with name "my-provider"
      When I create another claude provider with name "my-provider"
      Then I should receive a 400 error
      And the error message should contain "duplicate"

    @p1 @requirement:PM-02-003
    Scenario: Create provider with duplicate name across kinds
      Given I have a unique provider name "shared-provider"
      And I create a claude provider with name "shared-provider"
      When I create a codex provider with name "shared-provider"
      Then I should receive a 400 error
      And the error message should contain "duplicate"

    @p1 @requirement:PM-02-004
    Scenario: Create provider with invalid kind
      Given I have a unique provider name
      When I create a "invalid-kind" provider with API key "sk-test-123"
      Then I should receive a 400 error
      And the error message should contain "invalid kind"

    @p1 @requirement:PM-02-005
    Scenario: Create provider with empty API key
      Given I have a unique provider name
      When I create a claude provider with API key ""
      Then I should receive a 400 error
      And the error message should contain "api_key"

    @p1 @requirement:PM-02-006
    Scenario: Create provider with empty name
      When I create a claude provider with name ""
      Then I should receive a 400 error
      And the error message should contain "name"

    @p2 @requirement:PM-02-007
    Scenario: Create provider with priority
      Given I have a unique provider name
      When I create a claude provider with API key "sk-test-123" and priority 5
      Then the provider should be created successfully
      And the provider priority should be 5

    @p2 @requirement:PM-02-008
    Scenario: Create provider disabled by default
      Given I have a unique provider name
      When I create a claude provider with API key "sk-test-123" and disabled
      Then the provider should be created successfully
      And the provider should not be enabled

  Rule: Provider Retrieval

    @p1 @requirement:PM-02-009
    Scenario: List all providers
      When I list all providers
      Then the operation should succeed
      And I should see at least 1 provider
      And the API key should not be visible

    # Note: "Get provider by ID" removed - design uses list endpoint with filtering instead

    @p2 @requirement:PM-02-010
    Scenario: Get provider statistics
      Given I have created a provider
      When I get provider statistics
      Then the operation should succeed
      And the statistics should contain total requests
      And the statistics should contain success rate

    @p1 @requirement:PM-02-011
    Scenario: Get non-existent provider
      When I get provider with ID 99999
      Then I should receive a 404 error
      And the error message should contain "not found"

    @p1 @requirement:PM-02-012
    Scenario: List providers filtered by kind
      Given I have created a claude provider
      And I have created a codex provider
      When I list providers with kind "claude"
      Then I should only see claude providers
      And I should not see codex providers

  Rule: Provider Updates

    @p1 @requirement:PM-02-013
    Scenario: Update provider name
      Given I have created a provider
      When I update the provider name to "updated-test-name"
      Then the operation should succeed
      And the provider name should be "updated-test-name"

    @p1 @requirement:PM-02-014
    Scenario: Update provider API key
      Given I have created a provider
      When I update the provider API key to "sk-new-key-456"
      Then the operation should succeed
      And the API key should not be visible in response

    @p2 @requirement:PM-02-015
    Scenario: Update provider priority
      Given I have created a provider with priority 1
      When I update the provider priority to 10
      Then the operation should succeed
      And the provider priority should be 10

    @p2 @requirement:PM-02-016
    Scenario: Disable provider
      Given I have created an enabled provider
      When I disable the provider
      Then the operation should succeed
      And the provider should not be enabled

    @p2 @requirement:PM-02-017
    Scenario: Enable provider
      Given I have created a disabled provider
      When I enable the provider
      Then the operation should succeed
      And the provider should be enabled

    @p1 @requirement:PM-02-018
    Scenario: Update non-existent provider
      When I update provider with ID 99999
      Then I should receive a 404 error

  Rule: Provider Deletion

    @p1 @requirement:PM-02-019
    Scenario: Delete provider
      Given I have created a provider
      When I delete the provider
      Then the operation should succeed
      And the provider should not exist

    @p1 @requirement:PM-02-020
    Scenario: Delete non-existent provider
      When I delete provider with ID 99999
      Then I should receive a 404 error

    @p1 @requirement:PM-02-021
    Scenario: Delete provider by ID
      Given I have created a provider
      When I delete provider by ID
      Then the operation should succeed
      And the provider should not exist

    @p1 @requirement:PM-02-022
    Scenario: Cannot delete default provider
      Given there is a default provider with ID 1
      When I attempt to delete the default provider
      Then I should receive a 403 error
      And the error message should contain "default"

  Rule: Provider Connectivity Testing

    @p2 @requirement:PM-02-023
    Scenario: Test provider connectivity successfully
      Given I have created a provider with valid API key
      When I test the provider connectivity
      Then the operation should succeed
      And the connection should be successful

    @p2 @requirement:PM-02-024
    Scenario: Test provider connectivity with invalid API key
      Given I have created a provider with invalid API key
      When I test the provider connectivity
      Then the operation should succeed
      And the connection should fail
      And the error should indicate authentication failure

    @p1 @requirement:PM-02-025
    Scenario: Test connectivity for non-existent provider
      When I test connectivity for provider with ID 99999
      Then I should receive a 404 error

  Rule: Provider Management
    # Note: Per spec, provider limits are NOT enforced - both license types allow unlimited providers
    # Provider limits by tier scenarios removed as they don't match product requirements
