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

    Scenario: Create provider with duplicate name in same kind
      Given I have a unique provider name "my-provider"
      And I create a claude provider with name "my-provider"
      When I create another claude provider with name "my-provider"
      Then I should receive a 400 error
      And the error message should contain "duplicate"

    Scenario: Create provider with duplicate name across kinds
      Given I have a unique provider name "shared-provider"
      And I create a claude provider with name "shared-provider"
      When I create a codex provider with name "shared-provider"
      Then I should receive a 400 error
      And the error message should contain "duplicate"

    Scenario: Create provider with invalid kind
      Given I have a unique provider name
      When I create a "invalid-kind" provider with API key "sk-test-123"
      Then I should receive a 400 error
      And the error message should contain "invalid kind"

    Scenario: Create provider with empty API key
      Given I have a unique provider name
      When I create a claude provider with API key ""
      Then I should receive a 400 error
      And the error message should contain "api_key"

    Scenario: Create provider with empty name
      When I create a claude provider with name ""
      Then I should receive a 400 error
      And the error message should contain "name"

    Scenario: Create provider with priority
      Given I have a unique provider name
      When I create a claude provider with API key "sk-test-123" and priority 5
      Then the provider should be created successfully
      And the provider priority should be 5

    Scenario: Create provider disabled by default
      Given I have a unique provider name
      When I create a claude provider with API key "sk-test-123" and disabled
      Then the provider should be created successfully
      And the provider should not be enabled

  Rule: Provider Retrieval

    Scenario: List all providers
      When I list all providers
      Then the operation should succeed
      And I should see at least 1 provider
      And the API key should not be visible

    @wip
    Scenario: Get provider by ID
      Given I have created a provider
      When I get the provider by ID
      Then the operation should succeed
      And the provider should have a name
      And the API key should not be visible

    Scenario: Get provider statistics
      Given I have created a provider
      When I get provider statistics
      Then the operation should succeed
      And the statistics should contain total requests
      And the statistics should contain success rate

    Scenario: Get non-existent provider
      When I get provider with ID 99999
      Then I should receive a 404 error
      And the error message should contain "not found"

    Scenario: List providers filtered by kind
      Given I have created a claude provider
      And I have created a codex provider
      When I list providers with kind "claude"
      Then I should only see claude providers
      And I should not see codex providers

  Rule: Provider Updates

    Scenario: Update provider name
      Given I have created a provider
      When I update the provider name to "updated-test-name"
      Then the operation should succeed
      And the provider name should be "updated-test-name"

    Scenario: Update provider API key
      Given I have created a provider
      When I update the provider API key to "sk-new-key-456"
      Then the operation should succeed
      And the API key should not be visible in response

    Scenario: Update provider priority
      Given I have created a provider with priority 1
      When I update the provider priority to 10
      Then the operation should succeed
      And the provider priority should be 10

    Scenario: Disable provider
      Given I have created an enabled provider
      When I disable the provider
      Then the operation should succeed
      And the provider should not be enabled

    Scenario: Enable provider
      Given I have created a disabled provider
      When I enable the provider
      Then the operation should succeed
      And the provider should be enabled

    Scenario: Update non-existent provider
      When I update provider with ID 99999
      Then I should receive a 404 error

  Rule: Provider Deletion

    Scenario: Delete provider
      Given I have created a provider
      When I delete the provider
      Then the operation should succeed
      And the provider should not exist

    Scenario: Delete non-existent provider
      When I delete provider with ID 99999
      Then I should receive a 404 error

    Scenario: Delete provider by ID
      Given I have created a provider
      When I delete provider by ID
      Then the operation should succeed
      And the provider should not exist

    Scenario: Cannot delete default provider
      Given there is a default provider with ID 1
      When I attempt to delete the default provider
      Then I should receive a 403 error
      And the error message should contain "default"

  Rule: Provider Connectivity Testing

    Scenario: Test provider connectivity successfully
      Given I have created a provider with valid API key
      When I test the provider connectivity
      Then the operation should succeed
      And the connection should be successful

    Scenario: Test provider connectivity with invalid API key
      Given I have created a provider with invalid API key
      When I test the provider connectivity
      Then the operation should succeed
      And the connection should fail
      And the error should indicate authentication failure

    Scenario: Test connectivity for non-existent provider
      When I test connectivity for provider with ID 99999
      Then I should receive a 404 error

  Rule: Provider Limits

    Scenario Outline: Provider limits by tier
      Given the license has a provider limit of <limit>
      And I have created <limit> providers
      When I attempt to create a provider
      Then I should receive a 403 error
      And the error message should contain "provider limit"

      Examples:
        | limit |
        | 2     |
        | 5     |
        | 10    |

    @wip
    Scenario: Count providers towards limit
      Given the license has a provider limit of 3
      And I have created 2 providers
      When I create a provider
      Then the operation should succeed
      When I list all providers
      Then the total provider count should be 3

    Scenario: Provider limit does not affect updates
      Given the license has a provider limit of 2
      And I have created 2 providers
      When I update the first provider
      Then the operation should succeed

    @wip
    Scenario: Provider limit does not affect deletions
      Given the license has a provider limit of 2
      And I have created 2 providers
      When I delete the first provider
      Then the operation should succeed
      And I should be able to create a new provider

    Scenario: Provider kind limits
      Given the license has a claude provider limit of 1
      And I have created a claude provider
      When I attempt to create a claude provider
      Then I should receive a 403 error
      And the error message should contain "claude provider limit"

    Scenario: Different provider kinds have separate limits
      Given the license has a claude provider limit of 1
      And the license has a codex provider limit of 1
      And I have created a claude provider
      When I create a codex provider
      Then the operation should succeed

    Scenario: Unlimited providers for enterprise tier
      Given the license has enterprise tier
      And I have created 100 providers
      When I create a provider
      Then the operation should succeed
