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

    @p2 @requirement:PM-02-040
    Scenario: Enable already enabled provider
      Given I have created an enabled provider
      When I enable the provider
      Then the operation should succeed
      And the provider should be enabled

    @p2 @requirement:PM-02-041
    Scenario: Disable already disabled provider
      Given I have created a disabled provider
      When I disable the provider
      Then the operation should succeed
      And the provider should not be enabled

    @p1 @requirement:PM-02-042
    Scenario: Enable non-existent provider
      When I enable provider with ID 99999
      Then I should receive a 404 error

    @p1 @requirement:PM-02-043
    Scenario: Disable non-existent provider
      When I disable provider with ID 99999
      Then I should receive a 404 error

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

  Rule: Provider Validation Edge Cases

    @p1 @requirement:PM-02-026
    Scenario: Create provider with very long name
      Given I have a unique provider name
      When I create a claude provider with name of 300 characters
      Then I should receive a 400 error
      And the error message should contain "name" or "too long"

    @p1 @requirement:PM-02-027
    Scenario: Create provider with special characters in name
      Given I have a unique provider name "provider-with-special-chars-@#$"
      When I create a claude provider with name "provider-with-special-chars-@#$"
      Then the provider should be created successfully
      And the name should be sanitized

    @p1 @requirement:PM-02-028
    Scenario: Create provider with invalid API URL
      Given I have a unique provider name
      When I create a claude provider with API URL "invalid-url"
      Then I should receive a 400 error
      And the error message should contain "url" or "invalid"

    @p1 @requirement:PM-02-029
    Scenario: Create provider with missing optional fields
      Given I have a unique provider name
      When I create a claude provider with only required fields
      Then the provider should be created successfully
      And default values should be applied

    @p2 @requirement:PM-02-030
    Scenario: Create provider with all optional fields
      Given I have a unique provider name
      When I create a claude provider with all optional fields
      Then the provider should be created successfully
      And all fields should be set correctly

    @p1 @requirement:PM-02-031
    Scenario: Update provider with invalid priority
      Given I have created a provider
      When I update the provider priority to -1
      Then I should receive a 400 error
      And the error message should contain "priority"

    @p1 @requirement:PM-02-032
    Scenario: Update provider to duplicate another provider's name
      Given I have created a provider "first-provider"
      And I create another provider "second-provider"
      When I update "second-provider" name to "first-provider"
      Then I should receive a 400 error
      And the error message should contain "duplicate"

    @p2 @requirement:PM-02-033
    Scenario: Update provider multiple fields at once
      Given I have created a provider
      When I update provider name, API key, and priority simultaneously
      Then all fields should be updated successfully

    @p1 @requirement:PM-02-034
    Scenario: Delete non-existent provider returns 404
      When I delete provider with ID 99999
      Then I should receive a 404 error

    @p1 @requirement:PM-02-035
    Scenario: Update non-existent provider returns 404
      When I update provider with ID 99999
      Then I should receive a 404 error

    @p1 @requirement:PM-02-036
    Scenario: Get statistics for non-existent provider
      When I get provider statistics for ID 99999
      Then I should receive a 404 error

    @p2 @requirement:PM-02-037
    Scenario: List providers returns empty array when no providers exist
      Given I have deleted all providers
      When I list all providers
      Then I should see an empty list

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

  Rule: Provider Management Edge Cases

    @p2 @requirement:PM-02-044
    Scenario: Update provider name with empty string
      Given I have created a provider
      When I update the provider name to ""
      Then I should receive a 400 error
      And the error message should contain "name"

    @p2 @requirement:PM-02-045
    Scenario: Update provider API key with empty string
      Given I have created a provider
      When I update the provider API key to ""
      Then I should receive a 400 error
      And the error message should contain "api_key"

    @p2 @requirement:PM-02-046
    Scenario: Update provider with invalid priority
      Given I have created a provider
      When I update the provider priority to -1
      Then I should receive a 400 error
      And the error message should contain "priority"

    @p2 @requirement:PM-02-047
    Scenario: Create provider with duplicate name in same kind
      Given I have created a claude provider "duplicate-test"
      When I create another claude provider "duplicate-test"
      Then I should receive a 400 error
      And the error message should contain "duplicate"

    @p1 @requirement:PM-02-048
    Scenario: List providers when none exist
      Given I have deleted all providers
      When I list all providers
      Then I should see an empty list

    @p2 @requirement:PM-02-049
    Scenario: Create multiple providers in sequence
      Given I have created a claude provider
      When I create a codex provider
      And I create an opencode provider
      Then the operation should succeed
      And I should see at least 3 providers

    @p1 @requirement:PM-02-050
    Scenario: Get provider by name via list filtering
      Given I have created a claude provider "unique-test-provider"
      When I list providers with kind "claude"
      Then I should only see claude providers
      And I should see at least 1 provider

    @p1 @requirement:PM-02-051
    Scenario: Delete provider with invalid ID format
      Given I am logged in as a manager
      When I attempt to delete provider with invalid ID "abc"
      Then I should receive a 400 error

    @p1 @requirement:PM-02-052
    Scenario: Delete provider with very large ID
      Given I am logged in as a manager
      When I attempt to delete provider with ID 999999999
      Then I should receive a 404 error

  Rule: Provider Statistics

    @p1 @requirement:PM-02-053
    Scenario: Get provider stats for non-existent provider
      Given I am logged in as a manager
      When I attempt to get stats for provider ID 99999
      Then I should receive a 404 error

    @p1 @requirement:PM-02-054
    Scenario: Get provider stats without authentication
      Given I am not authenticated
      When I attempt to get stats for provider ID 1
      Then I should receive a 401 error

  Rule: Provider Enable and Disable

    @p1 @requirement:PM-02-055
    Scenario: Enable provider successfully
      Given I have created a provider
      And the provider is disabled
      When I enable the provider
      Then the operation should succeed
      And the provider should be enabled

    @p1 @requirement:PM-02-056
    Scenario: Disable provider successfully
      Given I have created a provider
      And the provider is enabled
      When I disable the provider
      Then the operation should succeed
      And the provider should be disabled

    @p1 @requirement:PM-02-057
    Scenario: Enable non-existent provider
      Given I am logged in as a manager
      When I attempt to enable provider with ID 99999
      Then I should receive a 404 error

    @p1 @requirement:PM-02-058
    Scenario: Disable non-existent provider
      Given I am logged in as a manager
      When I attempt to disable provider with ID 99999
      Then I should receive a 404 error

    @p1 @requirement:PM-02-059
    Scenario: Enable provider without authentication
      Given I am not authenticated
      When I attempt to enable provider with ID 1
      Then I should receive a 401 error

    @p1 @requirement:PM-02-060
    Scenario: Disable provider without authentication
      Given I am not authenticated
      When I attempt to disable provider with ID 1
      Then I should receive a 401 error

    @p2 @requirement:PM-02-061
    Scenario: Enable already enabled provider
      Given I have created a provider
      And the provider is enabled
      When I enable the provider
      Then the operation should succeed
      And the provider should still be enabled

    @p2 @requirement:PM-02-062
    Scenario: Disable already disabled provider
      Given I have created a provider
      And the provider is disabled
      When I disable the provider
      Then the operation should succeed
      And the provider should still be disabled

  Rule: Provider Test Connectivity Error Paths

    @p1 @requirement:PM-02-063
    Scenario: Test provider connectivity without authentication
      Given I am not authenticated
      When I attempt to test provider with ID 1
      Then I should receive a 401 error

    @p2 @requirement:PM-02-064
    Scenario: Test provider with timeout error
      Given I have created a provider with very long timeout
      When I test the provider connectivity
      Then the operation should succeed
      And the connection should timeout or fail

    @p2 @requirement:PM-02-072
    Scenario: Test provider with invalid URL format
      Given I have created a provider with invalid URL
      When I test the provider connectivity
      Then the operation should succeed
      And the connection should fail

    @p2 @requirement:PM-02-073
    Scenario: Test provider without authentication
      Given I am not authenticated
      When I attempt to test provider with ID 1
      Then I should receive a 401 error

  Rule: Provider List and Filter

    @p1 @requirement:PM-02-065
    Scenario: List providers by kind
      Given I have created a claude provider
      And I have created a codex provider
      When I list providers with kind "claude"
      Then I should only see claude providers
      And I should see at least 1 provider

    @p1 @requirement:PM-02-066
    Scenario: List providers with enabled filter
      Given I have created a provider
      And the provider is enabled
      When I list enabled providers
      Then I should see at least 1 provider
      And all providers should be enabled

    @p1 @requirement:PM-02-067
    Scenario: List providers without authentication
      Given I am not authenticated
      When I attempt to list providers
      Then I should receive a 401 error

  Rule: Provider CRUD Additional Scenarios

    @p2 @requirement:PM-02-068
    Scenario: Create provider with minimum required fields
      Given I am logged in as a manager
      When I create a claude provider with only name and API key
      Then the provider should be created successfully
      And default values should be applied

    @p2 @requirement:PM-02-069
    Scenario: Get provider after creation
      Given I am logged in as a manager
      And I have created a provider
      When I get the provider details
      Then the operation should succeed
      And I should see provider information

    @p2 @requirement:PM-02-071
    Scenario: List providers returns providers in order
      Given I am logged in as a manager
      And I have created 3 providers
      When I list all providers
      Then I should see at least 3 providers

  Rule: Provider Update Detailed Scenarios

    @p2 @requirement:PM-02-074
    Scenario: Update provider name to same value
      Given I am logged in as a manager
      And I have created a provider "same-name-test"
      When I update the provider name to "same-name-test"
      Then the operation should succeed
      And the provider name should be "same-name-test"

    @p2 @requirement:PM-02-076
    Scenario: Update provider all optional fields
      Given I am logged in as a manager
      And I have created a provider
      When I update provider with name, API key, priority, and URL
      Then all fields should be updated

    @p1 @requirement:PM-02-077
    Scenario: Update provider with zero priority
      Given I am logged in as a manager
      And I have created a provider
      When I update the provider priority to 0
      Then the operation should succeed
      And the provider priority should be 0

    @p1 @requirement:PM-02-078
    Scenario: Update provider with very high priority
      Given I am logged in as a manager
      And I have created a provider
      When I update the provider priority to 9999
      Then the operation should succeed
      And the provider priority should be 9999

  Rule: Provider List Pagination and Filtering

    @p2 @requirement:PM-02-079
    Scenario: List providers filters by enabled status correctly
      Given I am logged in as a manager
      And I have created an enabled provider
      And I have created a disabled provider
      When I list enabled providers
      Then I should see at least 1 provider
      And I should not see disabled providers

    @p2 @requirement:PM-02-080
    Scenario: List providers filters by kind correctly
      Given I am logged in as a manager
      And I have created a claude provider
      And I have created a codex provider
      And I have created an opencode provider
      When I list providers with kind "claude"
      Then I should only see claude providers

    @p1 @requirement:PM-02-081
    Scenario: List providers returns correct provider details
      Given I am logged in as a manager
      And I have created a provider "details-test"
      When I list all providers
      Then I should see provider "details-test"
      And the provider should have name
      And the provider should have kind

    @p2 @requirement:PM-02-082
    Scenario: List providers when only default exists
      Given I am logged in as a manager
      And I have deleted all custom providers
      When I list all providers
      Then I should see at least 1 provider

  Rule: Provider Connectivity Test Scenarios

    @p1 @requirement:PM-02-083
    Scenario: Test connectivity returns provider status
      Given I am logged in as a manager
      And I have created a provider
      When I test the provider connectivity
      Then the operation should succeed
      And the response should contain status information

    @p2 @requirement:PM-02-084
    Scenario: Test connectivity with disabled provider
      Given I am logged in as a manager
      And I have created a disabled provider
      When I test the provider connectivity
      Then the operation should succeed

    @p2 @requirement:PM-02-085
    Scenario: Test connectivity after updating API key
      Given I am logged in as a manager
      And I have created a provider
      And I update the provider API key to "new-test-key-789"
      When I test the provider connectivity
      Then the operation should succeed

    @p2 @requirement:PM-02-086
    Scenario: Test connectivity multiple providers
      Given I am logged in as a manager
      And I have created a claude provider
      And I have created a codex provider
      When I test the claude provider connectivity
      And I test the codex provider connectivity
      Then both tests should complete

  Rule: Provider Statistics Scenarios

    @p1 @requirement:PM-02-087
    Scenario: Get provider stats returns zero for new provider
      Given I am logged in as a manager
      And I have created a provider
      When I get provider statistics
      Then the operation should succeed
      And the statistics should show zero requests

    @p2 @requirement:PM-02-088
    Scenario: Get provider stats includes timestamp
      Given I am logged in as a manager
      And I have created a provider
      When I get provider statistics
      Then the response should contain timestamp or created time

    @p2 @requirement:PM-02-089
    Scenario: Get provider stats for default provider
      Given I am logged in as a manager
      When I get provider statistics for ID 1
      Then the operation should succeed
      And I should see statistics data

  Rule: Provider Toggle State Scenarios

    @p2 @requirement:PM-02-090
    Scenario: Toggle provider state multiple times
      Given I am logged in as a manager
      And I have created a provider
      When I disable the provider
      And I enable the provider
      And I disable the provider again
      And I enable the provider again
      Then the provider should be enabled

    @p2 @requirement:PM-02-091
    Scenario: Enable provider then update it
      Given I am logged in as a manager
      And I have created a disabled provider
      When I enable the provider
      And I update the provider name to "updated-enabled-name"
      Then the operation should succeed
      And the provider should be enabled

    @p2 @requirement:PM-02-092
    Scenario: Update provider then check enabled state
      Given I am logged in as a manager
      And I have created an enabled provider
      When I update the provider name
      Then the provider should still be enabled

  Rule: Provider List Extended Queries

    @p1 @requirement:PM-02-096
    Scenario: List providers returns provider IDs
      Given I am logged in as a manager
      And I have created a provider
      When I list all providers
      Then I should see at least 1 provider
      And each provider should have an ID

    @p2 @requirement:PM-02-097
    Scenario: List providers returns provider kinds
      Given I am logged in as a manager
      And I have created a claude provider
      When I list all providers
      Then I should see at least 1 provider
      And each provider should have a kind

  Rule: Provider Creation Extended Tests

    @p1 @requirement:PM-02-098
    Scenario: Create provider without specifying kind
      Given I am logged in as a manager
      And I have a unique provider name
      When I create a provider with API key "sk-default-kind" and no kind specified
      Then the provider should be created successfully
      And the provider kind should be "claude"

    @p1 @requirement:PM-02-099
    Scenario: Create provider with enabled false
      Given I am logged in as a manager
      And I have a unique provider name
      When I create a claude provider with API key "sk-disabled" and disabled
      Then the provider should be created successfully
      And the provider should not be enabled

    @p2 @requirement:PM-02-100
    Scenario: Create provider with all optional fields
      Given I am logged in as a manager
      And I have a unique provider name
      When I create a provider with name, API key, API URL, and level
      Then the provider should be created successfully
      And all fields should be set correctly

    @p2 @requirement:PM-02-101
    Scenario: List providers includes all provider data
      Given I am logged in as a manager
      And I have created a provider
      When I list all providers
      Then I should see at least 1 provider
      And the provider should have kind
      And the provider should have name

  Rule: Provider Enable/Disable Extended Tests

    @p1 @requirement:PM-02-104
    Scenario: Enable provider with non-numeric ID
      Given I am logged in as a manager
      When I attempt to enable provider with ID "invalid"
      Then I should receive a 400 error

    @p2 @requirement:PM-02-105
    Scenario: Disable provider with non-numeric ID
      Given I am logged in as a manager
      When I attempt to disable provider with ID "abc"
      Then I should receive a 400 error

  Rule: Provider Update Extended Tests

    @p1 @requirement:PM-02-108
    Scenario: Update provider with only name
      Given I am logged in as a manager
      And I have created a provider
      When I update the provider with only name "updated-name-only"
      Then the operation should succeed
      And the provider name should be "updated-name-only"

    @p1 @requirement:PM-02-109
    Scenario: Update provider with only API key
      Given I am logged in as a manager
      And I have created a provider
      When I update the provider with only API key "new-api-key-123"
      Then the operation should succeed

    @p1 @requirement:PM-02-110
    Scenario: Update provider with only priority level
      Given I am logged in as a manager
      And I have created a provider
      When I update the provider with only priority level 7
      Then the operation should succeed
      And the provider priority should be 7

    @p2 @requirement:PM-02-111
    Scenario: Update provider with multiple fields
      Given I am logged in as a manager
      And I have created a provider
      When I update the provider with name, API key, and level
      Then the operation should succeed
      And all fields should be updated

    @p2 @requirement:PM-02-112
    Scenario: Update non-existent provider
      Given I am logged in as a manager
      When I update provider with ID 99999
      Then I should receive a 404 error

  Rule: Provider Delete Extended Tests

    @p1 @requirement:PM-02-113
    Scenario: Delete provider twice returns error second time
      Given I am logged in as a manager
      And I have created a provider
      When I delete the provider
      And I attempt to delete the provider again
      Then the second delete should fail with 404

    @p1 @requirement:PM-02-114
    Scenario: Delete provider with invalid ID format
      Given I am logged in as a manager
      When I attempt to delete provider with invalid ID "not-a-number"
      Then I should receive a 400 error

  Rule: Provider Test Connectivity Extended

    @p1 @requirement:PM-02-115
    Scenario: Test connectivity multiple providers
      Given I am logged in as a manager
      And I have created a claude provider
      And I have created a codex provider
      When I test the claude provider connectivity
      Then the operation should succeed
      When I test the codex provider connectivity
      Then the operation should succeed

    @p1 @requirement:PM-02-116
    Scenario: Test connectivity after update
      Given I am logged in as a manager
      And I have created a provider
      When I update the provider API key to "new-test-key"
      And I test the provider connectivity
      Then the operation should succeed

    @p2 @requirement:PM-02-117
    Scenario: Test connectivity on enabled provider
      Given I am logged in as a manager
      And I have created an enabled provider
      When I test the provider connectivity
      Then the operation should succeed

    @p2 @requirement:PM-02-118
    Scenario: Test connectivity on disabled provider
      Given I am logged in as a manager
      And I have created a disabled provider
      When I test the provider connectivity
      Then the operation should succeed

  Rule: Provider Statistics Extended

    @p1 @requirement:PM-02-120
    Scenario: Get provider stats after update
      Given I am logged in as a manager
      And I have created a provider
      When I update the provider name
      And I get provider statistics
      Then the operation should succeed

    @p2 @requirement:PM-02-121
    Scenario: Get provider stats for default provider
      Given I am logged in as a manager
      When I get provider statistics for ID 1
      Then the operation should succeed
      And I should see statistics data

  Rule: Provider Enable/Disable Extended

    @p1 @requirement:PM-02-123
    Scenario: Enable provider then update
      Given I am logged in as a manager
      And I have created a disabled provider
      When I enable the provider
      And I update the provider name to "updated-after-enable"
      Then the operation should succeed
      And the provider should be enabled

    @p2 @requirement:PM-02-124
    Scenario: Update provider then check enabled state
      Given I am logged in as a manager
      And I have created an enabled provider
      When I update the provider name
      Then the provider should still be enabled

  Rule: Provider List Extended

    @p1 @requirement:PM-02-125
    Scenario: List providers returns enabled status
      Given I am logged in as a manager
      And I have created an enabled provider
      When I list all providers
      Then I should see at least 1 provider
      And all providers should be enabled

    @p1 @requirement:PM-02-126
    Scenario: List providers filter by enabled state
      Given I am logged in as a manager
      And I have created a disabled provider
      And I have created an enabled provider
      When I list all providers
      Then I should see at least 2 providers

    @p2 @requirement:PM-02-127
    Scenario: List providers returns provider types
      Given I am logged in as a manager
      And I have created a claude provider
      And I have created a codex provider
      When I list all providers
      Then I should see at least 2 providers
      And each provider should have a kind

    @p2 @requirement:PM-02-128
    Scenario: List providers returns priority levels
      Given I am logged in as a manager
      And I have created a provider with priority level 5
      When I list all providers
      Then I should see at least 1 provider

  Rule: Provider Creation Variations

    @p1 @requirement:PM-02-129
    Scenario: Create provider with minimal fields
      Given I am logged in as a manager
      And I have a unique provider name
      When I create a provider with only name and API key
      Then the provider should be created successfully
      And the provider should have a name

    @p1 @requirement:PM-02-130
    Scenario: Create provider with all optional fields set
      Given I am logged in as a manager
      And I have a unique provider name
      When I create a provider with name, API key, API URL, level, and enabled state
      Then the provider should be created successfully
      And all fields should be set correctly

    @p2 @requirement:PM-02-131
    Scenario: Create provider returns provider ID
      Given I am logged in as a manager
      And I have a unique provider name
      When I create a claude provider with API key "sk-id-test"
      Then the provider should be created successfully
      And the provider should have an ID

  Rule: Provider Error Paths

    @p1 @requirement:PM-02-132
    Scenario: Create provider with invalid kind fails
      Given I am logged in as a manager
      And I have a unique provider name
      When I create a provider with kind "invalid_kind"
      Then I should receive a validation error
      And the error should mention "invalid kind" or "valid kinds"

    @p1 @requirement:PM-02-133
    Scenario: Create provider with duplicate name fails
      Given I am logged in as a manager
      And I have created a provider named "duplicate-test"
      When I create a provider with name "duplicate-test"
      Then I should receive a conflict error
      And the error should mention "already exists"

    @p1 @requirement:PM-02-134
    Scenario: Update non-existent provider fails
      Given I am logged in as a manager
      When I update provider with ID 99999
      Then I should receive a 404 error

    @p1 @requirement:PM-02-135
    Scenario: Update provider with invalid ID fails
      Given I am logged in as a manager
      When I update provider with ID "invalid"
      Then I should receive a validation error

    @p2 @requirement:PM-02-136
    Scenario: Test non-existent provider fails
      Given I am logged in as a manager
      When I test provider with ID 99999
      Then I should receive a 404 error

  Rule: Provider List Variations

    @p1 @requirement:PM-02-139
    Scenario: List providers after creating multiple
      Given I am logged in as a manager
      And I have created a claude provider
      And I have created a codex provider
      When I list all providers
      Then I should see at least 2 providers

    @p1 @requirement:PM-02-140
    Scenario: List providers returns all required fields
      Given I am logged in as a manager
      And I have created a provider
      When I list all providers
      Then the operation should succeed
      And each provider should have an ID

    @p2 @requirement:PM-02-141
    Scenario: List providers as member
      Given I am logged in as a member
      When I list all providers
      Then the operation should succeed

    @p2 @requirement:PM-02-142
    Scenario: List providers without authentication fails
      Given I am not authenticated
      When I list all providers
      Then I should receive a 401 error

  Rule: Provider Basic Operations

    @p1 @requirement:PM-02-143
    Scenario: List providers returns provider list
      Given I am logged in as a manager
      When I list all providers
      Then the operation should succeed
      And I should see at least 1 provider

    @p1 @requirement:PM-02-144
    Scenario: Create and list providers
      Given I am logged in as a manager
      And I have created a provider
      When I list all providers
      Then I should see at least 1 provider

    @p2 @requirement:PM-02-145
    Scenario: Get provider details after creation
      Given I am logged in as a manager
      And I have created a provider
      When I get the provider by ID
      Then the operation should succeed

    @p2 @requirement:PM-02-146
    Scenario: Provider operations in sequence
      Given I am logged in as a manager
      And I have created a provider
      When I list all providers
      And I get the provider by ID
      Then both operations should succeed

  Rule: Provider Enable Disable Extended

    @p1 @requirement:PM-02-147
    Scenario: Enable provider multiple times
      Given I am logged in as a manager
      And I have created a provider
      When I enable the provider
      And I enable the provider
      And I enable the provider
      Then all operations should succeed

    @p1 @requirement:PM-02-148
    Scenario: Disable provider multiple times
      Given I am logged in as a manager
      And I have created a provider
      When I disable the provider
      And I disable the provider
      And I disable the provider
      Then all operations should succeed

    @p1 @requirement:PM-02-149
    Scenario: Enable then disable provider
      Given I am logged in as a manager
      And I have created a provider
      When I enable the provider
      And I disable the provider
      And I enable the provider
      Then all operations should succeed

    @p2 @requirement:PM-02-150
    Scenario: Provider list after enable disable
      Given I am logged in as a manager
      And I have created a provider
      When I enable the provider
      And I list all providers
      And I disable the provider
      And I list all providers
      Then all operations should succeed

