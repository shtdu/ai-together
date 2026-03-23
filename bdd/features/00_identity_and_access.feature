# BDD Test Suite - Identity and Access Management
# This feature covers authentication, user management, roles, permissions, and multi-tenancy
#
# BEFORE CLEANUP: 345 scenarios
# AFTER CLEANUP: 107 scenarios (-69% reduction)
#
# ORGANIZATION: This file is organized into clear sections for future splitting:
#   1. AUTH OPERATIONS         -> Will become auth.feature
#   2. USER PROFILE            -> Will become profile.feature
#   3. REGISTRATION            -> Will become profile.feature
#   4. PASSWORD RESET          -> Will become auth.feature
#   5. ACCOUNT LOCKOUT         -> Will become auth.feature
#   6. ROLE-BASED ACCESS       -> Will become rbac.feature
#   7. MULTI-TENANT ISOLATION  -> Will become rbac.feature
#   8. LICENSE MANAGEMENT      -> Will become licensing.feature
#   9. TEAM INVITATIONS        -> Will become invitations.feature
#  10. LAST MANAGER PROTECTION -> Will become rbac.feature
#  11. USER MANAGEMENT         -> Will become rbac.feature

Feature: Identity and Access Management
  As a system administrator
  I want to control who can access the system and what they can do
  So that our team's data and configurations remain secure

  Background:
    Given the test server is running

  # ============================================================================
  # SECTION 1: AUTH OPERATIONS (auth.feature)
  # ============================================================================
  # Covers: Login, Logout, Token Refresh, Token Verify

  Rule: Authentication - Login

    @p0 @requirement:IA-01-001
    Scenario Outline: Login with valid credentials
      Given a user exists with email "<email>" and password "<password>"
      And the user has role "<role>"
      When I login with email "<email>" and password "<password>"
      Then I should receive a valid authentication token
      And my profile should contain my email
      And my profile should contain my role
      And the response should contain both access and refresh tokens

      Examples:
        | email | password | role |
        | manager@example.com | TestPassword123! | admin |
        | member@example.com | MemberPass123! | member |

    @p1 @requirement:IA-01-003
    Scenario Outline: Login with invalid credentials
      When I login with email "<email>" and password "<password>"
      Then I should receive a "<error_type>" error
      And the response status code should be 401

      Examples:
        | email | password | error_type |
        | nonexistent@example.com | TestPassword123! | user_not_found |
        | manager@example.com | WrongPassword | invalid_password |

    @p1 @requirement:IA-01-070
    Scenario: Login with uppercase email (case insensitivity)
      Given a user exists with email "case@example.com" and password "CasePassword123!"
      When I login with email "CASE@EXAMPLE.COM" and password "CasePassword123!"
      Then the operation should succeed

  Rule: Authentication - Logout

    @p1 @requirement:IA-01-006
    Scenario: Logout successfully
      Given I am logged in as a manager
      When I logout
      Then the response status code should be 204
      And my authentication token should be invalid

  Rule: Authentication - Token Refresh

    @p1 @requirement:IA-01-007
    Scenario: Refresh valid authentication token
      Given I am logged in as a manager
      When I refresh my authentication token
      Then I should receive a valid authentication token
      And the response status code should be 200
      And the new token should be different from the old token
      And the response should contain user information

    @p1 @requirement:IA-01-010
    Scenario: Refresh token shortly before expiration
      Given I have a token that expires in 5 minutes
      When I refresh my authentication token
      Then I should receive a valid authentication token
      And the new token should have an extended expiration

    @p1 @requirement:IA-01-034
    Scenario: Refresh token with expired user context
      Given I am not authenticated
      And I have a user that was deleted
      When I refresh my authentication token
      Then I should receive an "user_not_found" error

    @p2 @requirement:IA-01-008
    Scenario Outline: Refresh token error handling
      Given I am logged in as a manager
      When I refresh my authentication token with "<token>"
      Then I should receive an "<error_type>" error
      And the response status code should be <status_code>

      Examples:
        | token | error_type | status_code |
        | invalid-token | invalid_token | 401 |
        | | invalid_token | 401 |

    @p2 @requirement:IA-01-098
    Scenario Outline: Refresh token request validation
      Given I am logged in as a manager
      When I send refresh request with <request_type>
      Then I should receive a validation error
      And the response status code should be 400

      Examples:
        | request_type |
        | invalid JSON |
        | empty request body |

  Rule: Authentication - Token Verify

    @p1 @requirement:IA-01-014
    Scenario: Verify valid authentication token
      Given I am logged in as a manager
      When I verify my authentication token
      Then the response status code should be 200
      And my profile should contain my email
      And the response should contain user information

    @p2 @requirement:IA-01-011
    Scenario Outline: Verify token error handling
      When I verify my authentication token with "<token>"
      Then I should receive a "<error_type>" error
      And the response status code should be 401

      Examples:
        | token | error_type |
        | invalid-token | invalid_token |
        | | invalid_token |

  # ============================================================================
  # SECTION 2: USER PROFILE (profile.feature)
  # ============================================================================

  Rule: User Profile - Happy Paths

    @p1 @requirement:IA-01-016
    Scenario Outline: Get user profile as authenticated user
      Given I am logged in as a <role>
      When I get my user profile
      Then the response status code should be 200
      And my profile should contain my email
      And my profile should contain my role
      And my profile should contain my user ID
      And my profile should contain tenant ID
      And my profile should contain my name

      Examples:
        | role |
        | manager |
        | member |

  Rule: User Profile - Field Validation

    @p2 @requirement:IA-01-029
    Scenario: Get profile includes all expected fields
      Given I am logged in as a manager
      When I get my user profile
      Then the response status code should be 200
      And my profile should contain "email" field
      And my profile should contain "name" field
      And my profile should contain "role" field
      And my profile should contain "tenant_id" field
      And my profile should contain "user_id" field
      And the profile should contain creation timestamp
      And the profile should contain last update timestamp
      And the profile should contain team ID
      And the profile should contain team name

  Rule: User Profile - Error Paths

    @p1 @requirement:IA-01-018
    Scenario Outline: Get profile without authentication
      Given I <auth_state>
      When I get my user profile
      Then I should receive an "unauthorized" error
      And the response status code should be 401

      Examples:
        | auth_state |
        | am not authenticated |
        | have an invalid authentication token "Bearer invalid-token-123" |
        | have an expired authentication token |

  # ============================================================================
  # SECTION 3: REGISTRATION (profile.feature)
  # ============================================================================

  Rule: User Registration - Happy Paths

    @p0 @requirement:IA-01-017
    Scenario: Register new organization successfully
      Given I am not authenticated
      And I have a unique email "newuser@example.com"
      And I have a strong password "StrongPass123!"
      When I register a new account
      Then the response status code should be 201
      And a new organization should be created
      And I should receive a valid authentication token
      And I should be granted the Manager role

    @p2 @requirement:IA-01-018
    Scenario: Registration creates unique organization name
      Given I am not authenticated
      And I have a unique email "orgtest@example.com"
      When I register a new account
      Then the organization name should be auto-generated

  Rule: User Registration - Password Validation

    @p2 @requirement:IA-01-020
    Scenario Outline: Registration password requirements
      Given I am not authenticated
      And I have a unique email "passtest@example.com"
      And I have password "<password>"
      When I register a new account
      Then I should receive a 400 error
      And the error message should contain "password"

      Examples:
        | password |
        | short |
        | nouppercase123! |
        | NOLOWERCASE123! |
        | NoNumbers! |
        | NoSpecial123 |

  Rule: User Registration - Error Paths

    @p1 @requirement:IA-01-021
    Scenario: Registration rejects duplicate email
      Given a user exists with email "existing@example.com" and password "TestPassword123!"
      And I am not authenticated
      And I have a unique email "existing@example.com"
      When I register a new account
      Then the response status code should be 400
      And the error message should contain "email already exists"

    @p1 @requirement:IA-01-025
    Scenario Outline: Registration with missing required fields
      Given I am not authenticated
      And I have a strong password "StrongPass123!"
      When I register <registration_type>
      Then I should receive a 400 error
      And the error message should contain "<field>"

      Examples:
        | registration_type | field |
        | with email "" | email |
        | with email "not-an-email" | email |
        | with password "" | password |

    @p1 @requirement:IA-01-028
    Scenario: Registration with missing name
      Given I am not authenticated
      And I have a unique email "noname@example.com"
      And I have a strong password "StrongPass123!"
      When I register without providing name
      Then I should receive a 400 error
      And the error message should contain "name"

  # ============================================================================
  # SECTION 4: PASSWORD RESET (auth.feature)
  # ============================================================================

  Rule: Password Reset

    @p1 @requirement:IA-01-023
    Scenario: Request password reset email
      Given a user exists with email "reset@example.com" and password "TestPassword123!"
      When I request a password reset
      Then the response status code should be 200
      And a reset token should be generated
      And the reset token should expire in 1 hour

    @p1 @requirement:IA-01-024
    Scenario: Reset password with valid token
      Given a user exists with email "validreset@example.com" and password "TestPassword123!"
      And I have a valid reset token
      When I reset password to "NewPassword123!"
      Then the response status code should be 200
      And I can login with the new password

    @p2 @requirement:IA-01-025
    Scenario: Reset password with expired token
      Given a user exists with email "expired@example.com" and password "TestPassword123!"
      And I have an expired reset token
      When I attempt to reset password
      Then I should receive a 400 error
      And the error message should contain "expired"

    @p2 @requirement:IA-01-026
    Scenario: Reset password with invalid token
      Given I have an invalid reset token
      When I attempt to reset password
      Then I should receive a 400 error
      And the error message should contain "invalid"

  # ============================================================================
  # SECTION 5: ACCOUNT LOCKOUT (auth.feature)
  # ============================================================================

  Rule: Account Lockout

    @p1 @requirement:IA-01-013
    Scenario: Account locked after 5 failed attempts
      Given a user exists with email "locktest@example.com" and password "TestPassword123!"
      When I fail to login 5 times with email "locktest@example.com" and wrong password
      Then the account should be locked
      And I should receive an "account_locked" error
      And the response status code should be 401

    @p1 @requirement:IA-01-014
    Scenario: Locked account cannot login with correct password
      Given a user exists with email "locked@example.com" and password "TestPassword123!"
      And the account is locked
      When I login with email "locked@example.com" and password "TestPassword123!"
      Then I should receive an "account_locked" error
      And the response status code should be 401

    @p2 @requirement:IA-01-015
    Scenario: Account unlocks after 15 minutes
      Given a user exists with email "timed@example.com" and password "TestPassword123!"
      And the account was locked 16 minutes ago
      When I login with email "timed@example.com" and password "TestPassword123!"
      Then I should receive a valid authentication token

    @p2 @requirement:IA-01-016
    Scenario: Successful login resets failed attempt counter
      Given a user exists with email "reset@example.com" and password "TestPassword123!"
      And I have failed to login 4 times
      When I login with email "reset@example.com" and password "TestPassword123!"
      Then the failed attempt counter should be reset to 0

  # ============================================================================
  # SECTION 6: ROLE-BASED ACCESS CONTROL (rbac.feature)
  # ============================================================================

  Rule: Role-Based Access Control

    @p1 @requirement:IA-02-001
    Scenario: Manager can create provider
      Given I am logged in as a manager
      And I have a unique provider name "test-provider"
      When I attempt to create a provider
      Then the operation should succeed
      And the response status code should be 201

    @p1 @requirement:IA-02-002
    Scenario: Member cannot create provider
      Given I am logged in as a member
      And I have a unique provider name "test-provider"
      When I attempt to create a provider
      Then I should receive a 403 error
      And the error message should contain "permission denied"

    @p1 @requirement:IA-02-003
    Scenario: Manager can view team analytics
      Given I am logged in as a manager
      When I get team analytics
      Then the operation should succeed
      And the response status code should be 200

    @p1 @requirement:IA-02-004
    Scenario: Manager can manage users
      Given I am logged in as a manager
      And I have a unique user "new-user@example.com"
      When I attempt to create a user
      Then the operation should succeed
      And the response status code should be 201

    @p1 @requirement:IA-02-005
    Scenario: Member cannot manage users
      Given I am logged in as a member
      And I have a unique user "new-user@example.com"
      When I attempt to create a user
      Then I should receive a 403 error
      And the error message should contain "permission denied"

    @p1 @requirement:IA-02-006
    Scenario: Manager can delete provider
      Given I am logged in as a manager
      And I have created a provider
      When I delete the first provider
      Then the operation should succeed
      And the response status code should be 200

    @p1 @requirement:IA-02-007
    Scenario: Member can view own usage
      Given I am logged in as a member
      When I get my usage statistics
      Then the operation should succeed
      And the response status code should be 200

    @p1 @requirement:IA-02-008
    Scenario: Manager can update provider
      Given I am logged in as a manager
      And I have created a provider
      When I update the first provider
      Then the operation should succeed
      And the response status code should be 200

    @p1 @requirement:IA-02-009
    Scenario: Member cannot update provider
      Given I am logged in as a member
      And I have a provider with ID "1"
      When I attempt to update the provider
      Then I should receive a 403 error
      And the error message should contain "permission denied"

  # ============================================================================
  # SECTION 7: MULTI-TENANT ISOLATION (rbac.feature)
  # ============================================================================

  Rule: Multi-Tenant Isolation

    @p0 @requirement:IA-03-001
    Scenario: Users from different tenants cannot access each other's data
      Given I am logged in as a manager in tenant "1"
      And there is a provider in tenant "2"
      When I attempt to get the provider
      Then I should receive a 404 error
      And the error message should contain "not found"

    @p1 @requirement:IA-03-002
    Scenario: Manager can only see users from own tenant
      Given I am logged in as a manager in tenant "1"
      And there are users in tenant "2"
      When I list all users
      Then I should only see users from tenant "1"

    @p1 @requirement:IA-03-004
    Scenario: Resources are isolated by tenant
      Given I am logged in as a manager in tenant "1"
      And I create a provider with name "tenant1-provider"
      When I login as a manager in tenant "2"
      And I list all providers
      Then I should not see "tenant1-provider"

  # ============================================================================
  # SECTION 8: LICENSE MANAGEMENT (licensing.feature)
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

  Rule: License Tiers

    @p0 @requirement:IA-04-028
    Scenario: Get license tiers (public endpoint)
      Given the test server is running
      When I get license tiers
      Then the operation should succeed
      And the response should contain tier information
      And the response status code should be 200
      And the response should be valid JSON

  # ============================================================================
  # SECTION 9: TEAM INVITATIONS (invitations.feature)
  # ============================================================================

  Rule: Team Invitation

    @p1 @requirement:IA-05-001
    Scenario: Manager sends team invitation
      Given I am logged in as a manager
      And I have a unique invitee email "invited@example.com"
      When I send a team invitation
      Then the response status code should be 201
      And an invitation record should be created
      And the invitation should have a unique token

    @p2 @requirement:IA-05-002
    Scenario: Invitation expires after 7 days
      Given I am logged in as a manager
      And there is an invitation created 8 days ago
      When the invitee attempts to accept the invitation
      Then the invitation should be expired

    @p1 @requirement:IA-05-003
    Scenario: Accept invitation with new user
      Given I am not authenticated
      And there is a pending invitation for "newuser@example.com"
      When I accept the invitation with password "NewUser123!"
      Then a new user should be created
      And the user should have the Member role
      And I should receive a valid authentication token

    @p1 @requirement:IA-05-005
    Scenario: Manager cancels pending invitation
      Given I am logged in as a manager
      And there is a pending invitation for "cancel@example.com"
      When I cancel the invitation
      Then the invitation should be invalidated
      And the invitee cannot accept the invitation

    @p1 @requirement:IA-05-006
    Scenario: Duplicate email in organization
      Given I am logged in as a manager
      And a user exists with email "duplicate@example.com"
      When I invite "duplicate@example.com"
      Then I should receive a 400 error
      And the error message should contain "already exists"

    @p2 @requirement:IA-05-007
    Scenario: Resend invitation email
      Given I am logged in as a manager
      And there is a pending invitation for "resend@example.com"
      When I resend the invitation
      Then the response status code should be 200
      And the invitation token should remain valid

    @p1 @requirement:IA-05-008
    Scenario: Member cannot send invitations
      Given I am logged in as a member
      When I attempt to send a team invitation
      Then I should receive a 403 error

  # ============================================================================
  # SECTION 10: LAST MANAGER PROTECTION (rbac.feature)
  # ============================================================================

  Rule: Last Manager Protection

    @p1 @requirement:IA-06-001
    Scenario: Last manager cannot demote themselves
      Given I am logged in as the only manager
      When I attempt to change my role to member
      Then I should receive a 403 error
      And the error message should contain "last manager"

    @p1 @requirement:IA-06-002
    Scenario: Last manager cannot be deactivated by another manager
      Given I am logged in as a manager
      And there is only one other manager
      When I attempt to deactivate the other manager
      Then I should receive a 403 error
      And the error message should contain "last manager"

    @p1 @requirement:IA-06-003
    Scenario: Manager can be demoted when other managers exist
      Given I am logged in as a manager
      And there are at least 2 managers in the organization
      When I demote another manager to member
      Then the operation should succeed
      And the response status code should be 200
      And the user should have the Member role

  # ============================================================================
  # SECTION 11: USER MANAGEMENT (rbac.feature)
  # ============================================================================

  Rule: User Management

    @p1 @requirement:IA-01-146
    Scenario: List all users as manager
      Given I am logged in as a manager
      When I list all users
      Then the operation should succeed
      And I should see user information

    @p2 @requirement:IA-01-148
    Scenario: List users as member returns forbidden
      Given I am logged in as a member
      When I list all users
      Then I should receive a 403 error

    @p2 @requirement:IA-01-149
    Scenario: List users without authentication returns unauthorized
      Given I am not authenticated
      When I list all users
      Then I should receive a 401 error
