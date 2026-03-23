# ========================================================================
# Configuration Sync BDD Scenarios
# ========================================================================
# Feature: Configuration Distribution
# Domain: 03 - Configuration Sync
# Requirements: docs/ears/03_configuration_sync/
#
# These scenarios cover the canonical happy paths and key failure scenarios
# for configuration distribution from manager to member clients.
# ========================================================================

Feature: Configuration Distribution
  As a manager
  I want provider configurations to automatically sync to team members
  So that everyone uses consistent settings without manual configuration

  # ------------------------------------------------------------------------
  # Background: Manager and member setup
  # ------------------------------------------------------------------------
  Background:
    Given the test server is running
    And I am logged in as a manager
    And I have a unique provider name "sync-test-claude"

  # ------------------------------------------------------------------------
  # CS-03-001 to CS-03-005: Automatic Configuration Synchronization
  # ------------------------------------------------------------------------
  @p0 @requirement:CS-03-001
  Scenario: Manager updates provider configuration and member receives it
    When I create a Claude provider with API key "sk-test-sync-123"
    Then the provider should be created successfully
    And the provider should have ID
    And the provider kind should be "claude"
    And the provider API key should be "sk-test-sync-123"

    # When a member logs in, they should receive the provider configuration
    When I login as a member
    And I list all providers
    Then I should see at least 1 provider
    And each provider should have kind
    And the operation should succeed

  # ------------------------------------------------------------------------
  # CS-03-201 to CS-03-203: Member Pull Operations
  # ------------------------------------------------------------------------
  @p1 @requirement:CS-03-201
  Scenario: Member pulls configuration updates manually
    Given a provider exists with name "team-provider"
    When I login as a member
    And I request configuration sync
    Then the operation should succeed
    And I should receive the latest configuration
    And the configuration should match the server version

  # ------------------------------------------------------------------------
  # CS-03-301 to CS-03-305: Configuration Validation
  # ------------------------------------------------------------------------
  @p0 @requirement:CS-03-301
  Scenario: System validates required fields before accepting configuration
    When I attempt to create a provider without required fields
    Then the operation should fail
    And I should receive a 400 error
    And the error message should contain "required" or "missing"

  @p1 @requirement:CS-03-302
  Scenario: System validates API key format before accepting configuration
    When I attempt to create a provider with invalid API key format
    Then the operation should fail
    And I should receive a 400 error
    And the error message should contain "invalid" or "format"

  # ------------------------------------------------------------------------
  # CS-03-601, CS-03-602: Sync Status Display
  # ------------------------------------------------------------------------
  @p1 @requirement:CS-03-601
  Scenario: System shows sync status when configuration matches server
    Given a provider exists with name "synced-provider"
    When I login as a member
    And I check sync status
    Then the operation should succeed
    And the sync status should be "synced"

  @p1 @requirement:CS-03-602
  Scenario: System shows offline indicator when server is unreachable
    Given the server is not available
    When I login as a member
    And I check sync status
    Then the operation should succeed
    And the sync status should be "offline"

  # ------------------------------------------------------------------------
  # Offline Mode: Cached Configuration
  # ------------------------------------------------------------------------
  @p2 @requirement:CS-03-002
  Scenario: Member uses cached configuration when offline
    Given a member has previously synced configuration
    And the member has cached provider "offline-claude"
    When the server becomes unavailable
    And I list all providers
    Then I should see at least 1 provider
    And the operation should succeed
    And the cached configuration should be used

  # ------------------------------------------------------------------------
  # Offline Mode: Queued Usage Sync
  # ------------------------------------------------------------------------
  @p2 @requirement:CS-03-003
  Scenario: Member queues usage data when offline and syncs when reconnected
    Given a member has accumulated usage data while offline
    And the member has 5 queued usage records
    When the server becomes available
    And I sync usage data
    Then the operation should succeed
    And all 5 usage records should be uploaded
    And the queued records should be cleared

  # ------------------------------------------------------------------------
  # Stale Configuration Recovery
  # ------------------------------------------------------------------------
  @p1 @requirement:CS-03-701
  Scenario: Member recovers from stale configuration when server reconnects
    Given a member has stale cached configuration
    And the cached configuration is 24 hours old
    When I connect to the server
    And I request configuration sync
    Then the operation should succeed
    And I should receive the latest configuration
    And the local configuration should be updated

  @p1 @requirement:CS-03-702
  Scenario: Server wins conflict when multiple managers update simultaneously
    Given multiple managers exist
    And manager A updates provider at time T1
    And manager B updates provider at time T2
    When T2 is later than T1
    And I sync the configuration
    Then the configuration from manager B should be used
    And the operation should succeed

  # ------------------------------------------------------------------------
  # Configuration Rollback
  # ------------------------------------------------------------------------
  @p2 @requirement:CS-03-401
  Scenario: Manager views configuration history
    Given I have created and updated providers multiple times
    When I view configuration history
    Then the operation should succeed
    And I should see at least 2 previous versions
    And each version should have timestamp

  @p2 @requirement:CS-03-402
  Scenario: Manager rolls back to previous configuration version
    Given I have configuration version V2
    And a previous configuration version V1 exists
    When I rollback to version V1
    Then the operation should succeed
    And the current configuration should match V1
    And the configuration should be distributed to members
