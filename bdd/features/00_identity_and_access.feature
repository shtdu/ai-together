# BDD Test Suite - Identity and Access Management
# This feature covers authentication, user management, roles, permissions, and multi-tenancy

Feature: Identity and Access Management
  As a system administrator
  I want to control who can access the system and what they can do
  So that our team's data and configurations remain secure

  Background:
    Given the test server is running

  Rule: Authentication

    Scenario: Login with valid credentials as manager
      Given a user exists with email "manager@example.com" and password "TestPassword123!"
      And the user has role "admin"
      When I login with email "manager@example.com" and password "TestPassword123!"
      Then I should receive a valid authentication token
      And my profile should contain my email
      And my profile should contain my role

    @wip
    Scenario: Login with valid credentials as member
      Given a user exists with email "member@example.com" and password "MemberPass123!"
      And the user has role "member"
      When I login with email "member@example.com" and password "MemberPass123!"
      Then I should receive a valid authentication token
      And my profile should contain my email
      And my profile should contain my role

    Scenario: Login with invalid email
      When I login with email "nonexistent@example.com" and password "TestPassword123!"
      Then I should receive an "user_not_found" error
      And the response status code should be 401

    Scenario: Login with invalid password
      Given a user exists with email "manager@example.com" and password "TestPassword123!"
      When I login with email "manager@example.com" and password "WrongPassword"
      Then I should receive an "invalid_password" error
      And the response status code should be 401

    Scenario: Login non-existent user
      When I login with email "nobody@example.com" and password "SomePassword"
      Then I should receive an "user_not_found" error
      And the response status code should be 401

    Scenario: Logout successfully
      Given I am logged in as a manager
      When I logout
      Then the response status code should be 204
      And my authentication token should be invalid

    Scenario: Refresh valid authentication token
      Given I am logged in as a manager
      When I refresh my authentication token
      Then I should receive a valid authentication token
      And the response status code should be 200

    Scenario: Refresh invalid authentication token
      When I refresh my authentication token with "invalid-token"
      Then I should receive an "invalid_token" error
      And the response status code should be 401

    Scenario: Refresh expired authentication token
      Given I have an expired authentication token
      When I refresh my authentication token
      Then I should receive an "token_expired" error
      And the response status code should be 401

    Scenario: Verify valid authentication token
      Given I am logged in as a manager
      When I verify my authentication token
      Then the response status code should be 200
      And my profile should contain my email

    Scenario: Verify invalid authentication token
      When I verify my authentication token with "invalid-token"
      Then I should receive an "invalid_token" error
      And the response status code should be 401

    Scenario: Verify expired authentication token
      Given I have an expired authentication token
      When I verify my authentication token
      Then I should receive an "token_expired" error
      And the response status code should be 401

  Rule: Account Lockout

    Scenario: Account locked after 5 failed attempts
      Given a user exists with email "locktest@example.com" and password "TestPassword123!"
      When I fail to login 5 times with email "locktest@example.com" and wrong password
      Then the account should be locked
      And I should receive an "account_locked" error
      And the response status code should be 401

    Scenario: Locked account cannot login with correct password
      Given a user exists with email "locked@example.com" and password "TestPassword123!"
      And the account is locked
      When I login with email "locked@example.com" and password "TestPassword123!"
      Then I should receive an "account_locked" error
      And the response status code should be 401

    Scenario: Account unlocks after 15 minutes
      Given a user exists with email "timed@example.com" and password "TestPassword123!"
      And the account was locked 16 minutes ago
      When I login with email "timed@example.com" and password "TestPassword123!"
      Then I should receive a valid authentication token

    Scenario: Successful login resets failed attempt counter
      Given a user exists with email "reset@example.com" and password "TestPassword123!"
      And I have failed to login 4 times
      When I login with email "reset@example.com" and password "TestPassword123!"
      Then the failed attempt counter should be reset to 0

  Rule: User Registration

    Scenario: Register new organization successfully
      Given I am not authenticated
      And I have a unique email "newuser@example.com"
      And I have a strong password "StrongPass123!"
      When I register a new account
      Then the response status code should be 201
      And a new organization should be created
      And I should receive a valid authentication token
      And I should be granted the Manager role

    Scenario: Registration creates unique organization name
      Given I am not authenticated
      And I have a unique email "orgtest@example.com"
      When I register a new account
      Then the organization name should be auto-generated

    Scenario: Registration validates password requirements
      Given I am not authenticated
      And I have a unique email "weakpass@example.com"
      And I have a weak password "short"
      When I register a new account
      Then I should receive a 400 error
      And the error message should contain "password"

    Scenario Outline: Registration password requirements
      Given I am not authenticated
      And I have a unique email "passtest@example.com"
      And I have password "<password>"
      When I register a new account
      Then I should receive a 400 error

      Examples:
        | password |
        | short |
        | nouppercase123! |
        | NOLOWERCASE123! |
        | NoNumbers! |
        | NoSpecial123 |

    Scenario: Registration rejects duplicate email
      Given a user exists with email "existing@example.com" and password "TestPassword123!"
      And I am not authenticated
      And I have a unique email "existing@example.com"
      When I register a new account
      Then the response status code should be 400
      And the error message should contain "email already exists"

    Scenario: Registration auto-logs in new user
      Given I am not authenticated
      And I have a unique email "autologin@example.com"
      And I have a strong password "StrongPass123!"
      When I register a new account
      Then I should receive a valid authentication token
      And I should be able to access my user profile

  Rule: Password Reset

    Scenario: Request password reset email
      Given a user exists with email "reset@example.com" and password "TestPassword123!"
      When I request a password reset
      Then the response status code should be 200
      And a reset token should be generated
      And the reset token should expire in 1 hour

    Scenario: Reset password with valid token
      Given a user exists with email "validreset@example.com" and password "TestPassword123!"
      And I have a valid reset token
      When I reset password to "NewPassword123!"
      Then the response status code should be 200
      And I can login with the new password

    Scenario: Reset password with expired token
      Given a user exists with email "expired@example.com" and password "TestPassword123!"
      And I have an expired reset token
      When I attempt to reset password
      Then I should receive a 400 error
      And the error message should contain "expired"

    Scenario: Reset password with invalid token
      Given I have an invalid reset token
      When I attempt to reset password
      Then I should receive a 400 error
      And the error message should contain "invalid"

  Rule: User Profiles

    Scenario: Get own profile as manager
      Given I am logged in as a manager
      When I get my user profile
      Then the response status code should be 200
      And my profile should contain my email
      And my profile should contain my role
      And my profile should contain tenant ID

    Scenario: Get own profile as member
      Given I am logged in as a member
      When I get my user profile
      Then the response status code should be 200
      And my profile should contain my email
      And my profile should contain my role
      And my profile should contain tenant ID

    Scenario: Get profile without authentication
      When I get my user profile
      Then I should receive an "unauthorized" error
      And the response status code should be 401

    Scenario: Profile contains correct fields
      Given I am logged in as a manager
      When I get my user profile
      Then the response status code should be 200
      And my profile should contain "email" field
      And my profile should contain "name" field
      And my profile should contain "role" field
      And my profile should contain "tenant_id" field

    Scenario: Profile has correct tenant ID
      Given I am logged in as a member
      And I belong to tenant with ID "1"
      When I get my user profile
      Then the response status code should be 200
      And my profile should contain tenant ID

  Rule: Role-Based Access Control

    @wip
    Scenario: Manager can create provider
      Given I am logged in as a manager
      And I have a unique provider name "test-provider"
      When I attempt to create a provider
      Then the operation should succeed
      And the response status code should be 201

    @wip
    Scenario: Member cannot create provider
      Given I am logged in as a member
      And I have a unique provider name "test-provider"
      When I attempt to create a provider
      Then I should receive a 403 error
      And the error message should contain "permission denied"

    @wip
    Scenario: Manager can view team analytics
      Given I am logged in as a manager
      When I get team analytics
      Then the operation should succeed
      And the response status code should be 200

    # Note: "Member cannot view team analytics" moved to 03_usage_insights.feature (domain-appropriate location)

    @wip
    Scenario: Manager can manage users
      Given I am logged in as a manager
      And I have a unique user "new-user@example.com"
      When I attempt to create a user
      Then the operation should succeed
      And the response status code should be 201

    @wip
    Scenario: Member cannot manage users
      Given I am logged in as a member
      And I have a unique user "new-user@example.com"
      When I attempt to create a user
      Then I should receive a 403 error
      And the error message should contain "permission denied"

    @wip
    Scenario: Manager can delete provider
      Given I am logged in as a manager
      And I have created a provider
      When I delete the first provider
      Then the operation should succeed
      And the response status code should be 200

    @wip
    Scenario: Member can view own usage
      Given I am logged in as a member
      When I get my usage statistics
      Then the operation should succeed
      And the response status code should be 200

    @wip
    Scenario: Manager can update provider
      Given I am logged in as a manager
      And I have created a provider
      When I update the first provider
      Then the operation should succeed
      And the response status code should be 200

    @wip
    Scenario: Member cannot update provider
      Given I am logged in as a member
      And I have a provider with ID "1"
      When I attempt to update the provider
      Then I should receive a 403 error
      And the error message should contain "permission denied"

  Rule: Multi-Tenant Isolation

    Scenario: Users from different tenants cannot access each other's data
      Given I am logged in as a manager in tenant "1"
      And there is a provider in tenant "2"
      When I attempt to get the provider
      Then I should receive a 404 error
      And the error message should contain "not found"

    Scenario: Manager can only see users from own tenant
      Given I am logged in as a manager in tenant "1"
      And there are users in tenant "2"
      When I list all users
      Then I should only see users from tenant "1"

    Scenario: Profile contains correct tenant ID
      Given I am logged in as a member in tenant "1"
      When I get my user profile
      Then my profile should contain tenant ID
      And the tenant ID should be "1"

    Scenario: Resources are isolated by tenant
      Given I am logged in as a manager in tenant "1"
      And I create a provider with name "tenant1-provider"
      When I login as a manager in tenant "2"
      And I list all providers
      Then I should not see "tenant1-provider"

  Rule: License Activation

    @wip
    Scenario: Activate commercial license successfully
      Given I am logged in as a manager
      And I have a valid commercial license key
      When I activate the license
      Then the license should be activated
      And commercial features should be available
      And the license tier should be "professional"

    @wip
    Scenario: Activate open-source license successfully
      Given I am logged in as a manager
      And I have a valid open-source license key
      When I activate the license
      Then the license should be activated
      And the license tier should be "trial"

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

    Scenario: Activate license with invalid signature
      Given I am logged in as a manager
      And I have an invalid license signature
      When I activate a license with signature
      Then I should receive a 400 error
      And the error message should contain "invalid signature"

    @wip
    Scenario: Activate license with expired signature
      Given I am logged in as a manager
      And I have an expired license signature
      When I activate a license with signature
      Then I should receive a 400 error
      And the error message should contain "expired"

    Scenario: Activate license without authentication
      Given I am not authenticated
      And I have a valid license key
      When I activate the license
      Then I should receive a 401 error

    @wip
    Scenario: Non-manager cannot activate license
      Given I am logged in as a member
      And I have a valid license key
      When I activate the license
      Then I should receive a 403 error
      And the error message should contain "permission denied"

  Rule: License Features

    Scenario: Commercial license enables provider management
      Given I have an activated commercial license
      When I check available features
      Then provider management should be enabled
      And team analytics should be enabled

    Scenario: Trial license has limited features
      Given I have an activated trial license
      When I check available features
      Then provider management should be enabled
      And team analytics should be disabled

    Scenario: Enterprise license enables all features
      Given I have an activated enterprise license
      When I check available features
      Then all features should be enabled
      And advanced analytics should be enabled

    Scenario: License feature flags are correct
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

    Scenario: Check license status returns correct information
      Given I have an activated professional license
      When I get license status
      Then the status should be "active"
      And the tier should be "professional"
      And features should be listed

  Rule: License Limits

    # Note: Per spec, provider and user limits are NOT enforced
    # Both Open Source and Commercial licenses have unlimited providers and seats
    # Only team limits are enforced (1 team for Open Source, unlimited for Commercial)

    @wip
    Scenario: License kind limits
      Given I have an activated professional license
      And the license has a claude provider limit of 5
      And I have created 5 claude providers
      When I attempt to create a claude provider
      Then I should receive a 403 error
      And the error message should contain "claude provider limit"

  Rule: License Expiration

    Scenario: License expires after end date
      Given I have a license expiring in 1 day
      When I check license expiration
      Then the license should be marked as "expiring_soon"
      And I should see days remaining

    # Note: Removed "Expired license cannot create providers" - per spec BR-002,
    # expired licenses "preserve data access but disable advanced features".
    # Provider creation is considered core functionality, not advanced feature.

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

    Scenario: Renew expired license
      Given I have an expired license
      And I have a valid renewal key
      When I renew the license
      Then the license should be activated
      And the expiration date should be updated

    Scenario: License grace period after expiration
      Given I have a license that expired 1 day ago
      When I check license status
      Then the license should be marked as "grace_period"
      And features should still be available

    Scenario: License suspended after grace period
      Given I have a license that expired 30 days ago
      When I check license status
      Then the license should be marked as "suspended"
      And features should be disabled

  Rule: License Status

    Scenario: Get license information
      Given I have an activated professional license
      When I get license information
      Then the license tier should be "professional"
      And the license status should be "active"
      And the license should have expiration date

    Scenario: License status includes limits
      Given I have an activated trial license
      When I get license information
      Then the license should include provider limit
      And the license should include user limit

    Scenario: License status includes features
      Given I have an activated enterprise license
      When I get license information
      Then the license should list all features
      And all features should be enabled

    Scenario: Check license without authentication
      Given I am not authenticated
      When I get license information
      Then I should receive a 401 error

    Scenario: Manager can view license information
      Given I am logged in as a manager
      When I get license information
      Then the operation should succeed

    # Note: Per spec FR-004 "License status visible to all users", members CAN view license info
    Scenario: Member can view license information
      Given I am logged in as a member
      When I get license information
      Then the operation should succeed

  Rule: Team Invitation

    Scenario: Manager sends team invitation
      Given I am logged in as a manager
      And I have a unique invitee email "invited@example.com"
      When I send a team invitation
      Then the response status code should be 201
      And an invitation record should be created
      And the invitation should have a unique token

    Scenario: Invitation expires after 7 days
      Given I am logged in as a manager
      And there is an invitation created 8 days ago
      When the invitee attempts to accept the invitation
      Then the invitation should be expired

    Scenario: Accept invitation with new user
      Given I am not authenticated
      And there is a pending invitation for "newuser@example.com"
      When I accept the invitation with password "NewUser123!"
      Then a new user should be created
      And the user should have the Member role
      And I should receive a valid authentication token

    Scenario: Accept invitation sets up password
      Given I am not authenticated
      And there is a pending invitation for "setup@example.com"
      When I accept the invitation with password "SetupPass123!"
      Then the user password should be set
      And the user should be able to login

    Scenario: Manager cancels pending invitation
      Given I am logged in as a manager
      And there is a pending invitation for "cancel@example.com"
      When I cancel the invitation
      Then the invitation should be invalidated
      And the invitee cannot accept the invitation

    Scenario: Duplicate email in organization
      Given I am logged in as a manager
      And a user exists with email "duplicate@example.com"
      When I invite "duplicate@example.com"
      Then I should receive a 400 error
      And the error message should contain "already exists"

    Scenario: Resend invitation email
      Given I am logged in as a manager
      And there is a pending invitation for "resend@example.com"
      When I resend the invitation
      Then the response status code should be 200
      And the invitation token should remain valid

    Scenario: Member cannot send invitations
      Given I am logged in as a member
      When I attempt to send a team invitation
      Then I should receive a 403 error

  Rule: Last Manager Protection

    Scenario: Last manager cannot demote themselves
      Given I am logged in as the only manager
      When I attempt to change my role to member
      Then I should receive a 403 error
      And the error message should contain "last manager"

    Scenario: Last manager cannot be deactivated by another manager
      Given I am logged in as a manager
      And there is only one other manager
      When I attempt to deactivate the other manager
      Then I should receive a 403 error
      And the error message should contain "last manager"

    Scenario: Manager can be demoted when other managers exist
      Given I am logged in as a manager
      And there are at least 2 managers in the organization
      When I demote another manager to member
      Then the operation should succeed
      And the user should have the Member role
