# BDD Test Suite - License Management
# This feature covers license activation, features, status, limits, expiration, upgrades, and data retention

Feature: License Management
  As a system administrator
  I want to activate and manage licenses
  So that my organization can access appropriate features based on our subscription tier

  Background:
    Given the test server is running

  # ============================================================================
  # LICENSE ACTIVATION
  # ============================================================================

  Rule: License Activation

    @p1 @requirement:IA-04-001 @license:commercial
    Scenario: Activate commercial license successfully
      Given I am logged in as a manager
      And I have a valid commercial license key
      When I activate the license
      Then the license should be activated
      And commercial features should be available
      And the license tier should be "professional"

    @p1 @requirement:IA-04-002
    Scenario: Activate open-source license successfully
      Given I am logged in as a manager
      And I have a valid open-source license key
      When I activate the license
      Then the license should be activated
      And the license tier should be "trial"

    @p1 @requirement:IA-04-003
    Scenario Outline: Activate license with signature
      Given I am logged in as a manager
      And I have a license signature "<signature>"
      When I activate a license with signature
      Then the license should be activated
      And the license tier should be "<tier>"

      Examples:
        | signature | tier |
        | commercial-signature | professional |
        | enterprise-signature | enterprise |
        | trial-signature | trial |

    @p2 @requirement:IA-04-004
    Scenario: Activate license with invalid signature
      Given I am logged in as a manager
      And I have an invalid license signature
      When I activate a license with signature
      Then I should receive a 400 error
      And the error message should contain "invalid signature"

    @p2 @requirement:IA-04-005 @license:commercial
    Scenario: Activate license with expired signature
      Given I am logged in as a manager
      And I have an expired license signature
      When I activate a license with signature
      Then I should receive a 400 error
      And the error message should contain "expired"

    @p1 @requirement:IA-04-006
    Scenario: Activate license without authentication
      Given I am not authenticated
      And I have a valid license key
      When I activate the license
      Then I should receive a 401 error

    @p1 @requirement:IA-04-007 @license:commercial
    Scenario: Non-manager cannot activate license
      Given I am logged in as a member
      And I have a valid license key
      When I activate the license
      Then I should receive a 403 error
      And the error message should contain "permission denied"

    @p1 @requirement:IA-04-021
    Scenario: Activate expired license fails
      Given I am logged in as a manager
      And I have an expired license key
      When I attempt to activate the license
      Then I should receive a 400 error
      And the error message should contain "expired"

  # ============================================================================
  # LICENSE FEATURES
  # ============================================================================

  Rule: License Features

    @p1 @requirement:IA-04-008
    Scenario: Commercial license enables provider management
      Given I have an activated commercial license
      When I check available features
      Then provider management should be enabled
      And team analytics should be enabled

    @p1 @requirement:IA-04-009
    Scenario: Trial license has limited features
      Given I have an activated trial license
      When I check available features
      Then provider management should be enabled
      And team analytics should be disabled

    @p1 @requirement:IA-04-010
    Scenario: Enterprise license enables all features
      Given I have an activated enterprise license
      When I check available features
      Then all features should be enabled
      And advanced analytics should be enabled

    @p2 @requirement:IA-04-011
    Scenario Outline: License feature flags are correct
      Given I have an activated <tier> license
      When I get license features
      Then the feature "<feature>" should be <enabled>

      Examples:
        | tier | feature | enabled |
        | trial | provider_management | enabled |
        | trial | team_analytics | disabled |
        | professional | provider_management | enabled |
        | professional | team_analytics | enabled |
        | enterprise | advanced_analytics | enabled |

  # ============================================================================
  # LICENSE STATUS
  # ============================================================================

  Rule: License Status

    @p1 @requirement:IA-04-022
    Scenario: Get license information
      Given I have an activated professional license
      When I get license information
      Then the license tier should be "professional"
      And the license status should be "active"
      And the license should have expiration date
      And the license should include provider limit
      And the license should include user limit

    @p1 @requirement:IA-04-025
    Scenario: Check license without authentication
      Given I am not authenticated
      When I get license information
      Then I should receive a 401 error

    # Note: Per spec FR-004 "License status visible to all users", members CAN view license info
    @p1 @requirement:IA-04-027
    Scenario: Member can view license information
      Given I am logged in as a member
      When I get license information
      Then the operation should succeed

  # ============================================================================
  # LICENSE LIMITS
  # ============================================================================

  Rule: License Limits

    # Note: Per spec, provider and user limits are NOT enforced
    # Both Open Source and Commercial licenses have unlimited providers and seats
    # Only team limits are enforced (1 team for Open Source, unlimited for Commercial)

    @p3 @requirement:IA-04-013 @license:commercial
    Scenario: License kind limits (provider limits - NOT enforced per spec)
      Given I have an activated professional license
      And the license has a claude provider limit of 5
      And I have created 5 claude providers
      When I attempt to create a claude provider
      Then I should receive a 403 error
      And the error message should contain "claude provider limit"

  # ============================================================================
  # TEAM LIMIT ENFORCEMENT
  # ============================================================================

  Rule: Team Limit Enforcement

    @p1 @requirement:IA-04-014 @license:commercial
    Scenario: Open Source license allows only 1 team
      Given I have an Open Source license
      And I have created 1 team
      When I attempt to create another team
      Then I should receive a 403 error
      And the error message should contain "team limit"

    @p1 @requirement:IA-04-015 @license:commercial
    Scenario: Commercial license allows unlimited teams
      Given I have a Commercial license
      And I have created 5 teams
      When I create another team
      Then the operation should succeed

    @p2 @requirement:IA-04-016 @license:commercial
    Scenario: Expired Commercial license reverts to 1 team limit
      Given I have an expired Commercial license
      And I have 3 existing teams
      When I attempt to create a new team
      Then I should receive a 403 error
      And the error message should contain "team limit"

    @p1 @requirement:IA-04-032
    Scenario: Seat limit is informational only
      Given I have a license with 10 seats
      And I have 10 active users
      When I add an 11th user
      Then the user should be created successfully
      And the license should show 11 users in use

  # ============================================================================
  # LICENSE EXPIRATION
  # ============================================================================

  Rule: License Expiration

    @p2 @requirement:IA-04-017
    Scenario: License expires after end date
      Given I have a license expiring in 1 day
      When I check license expiration
      Then the license should be marked as "expiring_soon"
      And I should see days remaining

    @p2 @requirement:IA-04-018
    Scenario Outline: License expiration warnings
      Given I have a license expiring in <days> days
      When I get license status
      Then the license should be marked as "<status>"

      Examples:
        | days | status |
        | 1 | expiring_soon |
        | 7 | expiring_soon |
        | 30 | active |
        | 90 | active |

    @p2 @requirement:IA-04-019
    Scenario: Renew expired license
      Given I have an expired license
      And I have a valid renewal key
      When I renew the license
      Then the license should be activated
      And the expiration date should be updated

    @p3 @requirement:IA-04-020
    Scenario: License grace period after expiration
      Given I have a license that expired 1 day ago
      When I check license status
      Then the license should be marked as "grace_period"
      And features should still be available

    @p3 @requirement:IA-04-021
    Scenario: License suspended after grace period
      Given I have a license that expired 30 days ago
      When I check license status
      Then the license should be marked as "suspended"
      And features should be disabled

  # ============================================================================
  # LICENSE UPGRADES AND DOWNGRADES
  # ============================================================================

  Rule: License Upgrades and Downgrades

    @p1 @requirement:IA-04-024
    Scenario: Upgrade from Open Source to Commercial license
      Given I have an active Open Source license
      And I have a valid Commercial license key
      When I activate the Commercial license
      Then the license type should be "commercial"
      And the data retention days should change to 90
      And the max teams should change to unlimited
      And existing data should be preserved

    @p2 @requirement:IA-04-025
    Scenario: Downgrade from Commercial to Open Source license
      Given I have an active Commercial license
      And I have data older than 7 days
      When I activate an Open Source license
      Then the license type should be "opensource"
      And the data retention days should change to 7
      But data older than 7 days should not be immediately deleted

  # ============================================================================
  # LICENSE DATA RETENTION
  # ============================================================================

  Rule: License Data Retention

    @p1 @requirement:IA-04-038
    Scenario: Open Source license 7-day retention
      Given I have an Open Source license
      And I have usage data from 10 days ago
      When I query usage statistics
      Then I should not see data older than 7 days

    @p1 @requirement:IA-04-039 @license:commercial
    Scenario: Commercial license 90-day retention
      Given I have a Commercial license
      And I have usage data from 60 days ago
      When I query usage statistics
      Then I should see data from 60 days ago

  # ============================================================================
  # LICENSE TIERS
  # ============================================================================

  Rule: License Tiers

    @p0 @requirement:IA-04-028
    Scenario: Get license tiers (public endpoint)
      Given the test server is running
      When I get license tiers
      Then the operation should succeed
      And the response should contain tier information
      And the response status code should be 200
      And the response should be valid JSON
