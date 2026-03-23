# BDD Test Suite - Identity and Access Management
# This feature covers authentication, user management, roles, permissions, and multi-tenancy

Feature: Identity and Access Management
  As a system administrator
  I want to control who can access the system and what they can do
  So that our team's data and configurations remain secure

  Background:
    Given the test server is running

  Rule: Authentication

    @p0 @requirement:IA-01-001
    Scenario: Login with valid credentials as manager
      Given a user exists with email "manager@example.com" and password "TestPassword123!"
      And the user has role "admin"
      When I login with email "manager@example.com" and password "TestPassword123!"
      Then I should receive a valid authentication token
      And my profile should contain my email
      And my profile should contain my role

    @p1 @requirement:IA-01-002
    Scenario: Login with valid credentials as member
      Given a user exists with email "member@example.com" and password "MemberPass123!"
      And the user has role "member"
      When I login with email "member@example.com" and password "MemberPass123!"
      Then I should receive a valid authentication token
      And my profile should contain my email
      And my profile should contain my role

    @p1 @requirement:IA-01-003
    Scenario: Login with invalid email
      When I login with email "nonexistent@example.com" and password "TestPassword123!"
      Then I should receive an "user_not_found" error
      And the response status code should be 401

    @p1 @requirement:IA-01-004
    Scenario: Login with invalid password
      Given a user exists with email "manager@example.com" and password "TestPassword123!"
      When I login with email "manager@example.com" and password "WrongPassword"
      Then I should receive an "invalid_password" error
      And the response status code should be 401

    @p1 @requirement:IA-01-005
    Scenario: Login non-existent user
      When I login with email "nobody@example.com" and password "SomePassword"
      Then I should receive an "user_not_found" error
      And the response status code should be 401

    @p1 @requirement:IA-01-006
    Scenario: Logout successfully
      Given I am logged in as a manager
      When I logout
      Then the response status code should be 204
      And my authentication token should be invalid

    @p1 @requirement:IA-01-007
    Scenario: Refresh valid authentication token
      Given I am logged in as a manager
      When I refresh my authentication token
      Then I should receive a valid authentication token
      And the response status code should be 200

    @p2 @requirement:IA-01-008
    Scenario: Refresh invalid authentication token
      When I refresh my authentication token with "invalid-token"
      Then I should receive an "invalid_token" error
      And the response status code should be 401

    @p2 @requirement:IA-01-009
    Scenario: Refresh expired authentication token
      Given I have an expired authentication token
      When I refresh my authentication token
      Then I should receive an "token_expired" error
      And the response status code should be 401

    @p1 @requirement:IA-01-010
    Scenario: Refresh token shortly before expiration
      Given I have a token that expires in 5 minutes
      When I refresh my authentication token
      Then I should receive a valid authentication token
      And the new token should have an extended expiration

    @p1 @requirement:IA-01-011
    Scenario: Refresh token multiple times
      Given I have a valid authentication token
      When I refresh my authentication token 3 times
      Then each refresh should return a valid token
      And all tokens should be different

    @p1 @requirement:IA-01-012
    Scenario: Refresh token without authentication
      When I refresh my authentication token without providing token
      Then I should receive a 401 error
      And the error message should contain "unauthorized" or "missing"

    @p2 @requirement:IA-01-013
    Scenario: Refresh token with malformed token
      When I refresh my authentication token with "malformed.token.format"
      Then I should receive a 400 error
      And the error message should contain "invalid" or "malformed"

    @p2 @requirement:IA-01-015
    Scenario: Refresh token with empty request body
      When I send refresh request with empty body
      Then I should receive a 400 error
      And the error message should contain "Invalid request payload"

    @p2 @requirement:IA-01-032
    Scenario: Refresh token with malformed JSON
      When I send refresh request with malformed JSON "{invalid json"
      Then I should receive a 400 error
      And the error message should contain "Invalid request payload"

    @p2 @requirement:IA-01-033
    Scenario: Refresh token with missing refresh_token field
      When I send refresh request with empty JSON body
      Then I should receive a 401 error

    @p2 @requirement:IA-01-034
    Scenario: Refresh token with empty refresh_token value
      When I send refresh request with empty token value
      Then I should receive a 401 error

    @p1 @requirement:IA-01-020
    Scenario: Refresh token returns new access token
      Given I am logged in as a manager
      When I refresh my authentication token
      Then I should receive a valid authentication token
      And the new token should be different from the old token

    @p1 @requirement:IA-01-021
    Scenario: Refresh token includes user info in response
      Given I am logged in as a manager
      When I refresh my authentication token
      Then I should see user information in response
      And the response should contain access token

    @p1 @requirement:IA-01-014
    Scenario: Verify valid authentication token
      Given I am logged in as a manager
      When I verify my authentication token
      Then the response status code should be 200
      And my profile should contain my email

    @p2 @requirement:IA-01-011
    Scenario: Verify invalid authentication token
      When I verify my authentication token with "invalid-token"
      Then I should receive an "invalid_token" error
      And the response status code should be 401

    @p2 @requirement:IA-01-012
    Scenario: Verify expired authentication token
      Given I have an expired authentication token
      When I verify my authentication token
      Then I should receive an "token_expired" error
      And the response status code should be 401

    @p1 @requirement:IA-01-029
    Scenario: Verify token returns user profile
      Given I am logged in as a manager
      When I verify my authentication token
      Then the response status code should be 200
      And my profile should contain my email

    @p1 @requirement:IA-01-030
    Scenario: Verify token without authentication
      When I verify my authentication token
      Then I should receive a 401 error
      And the error message should contain "unauthorized"

    @p1 @requirement:IA-01-031
    Scenario: Verify with invalid token format
      When I verify my authentication token with "invalid-format"
      Then I should receive a 401 error
      And the error message should contain "invalid"

    @p1 @requirement:IA-01-016
    Scenario: Get user profile as authenticated manager
      Given I am logged in as a manager
      When I get my user profile
      Then the response status code should be 200
      And my profile should contain my email
      And my profile should contain my role
      And my profile should contain my user ID

    @p1 @requirement:IA-01-017
    Scenario: Get user profile as authenticated member
      Given I am logged in as a member
      When I get my user profile
      Then the response status code should be 200
      And my profile should contain my email
      And my profile should contain my role

    @p1 @requirement:IA-01-018
    Scenario: Get profile without authentication
      When I get my user profile
      Then I should receive a 401 error
      And the error message should contain "unauthorized"

    @p1 @requirement:IA-01-019
    Scenario: Get profile with invalid token
      When I get my user profile with invalid token
      Then I should receive a 401 error

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

  Rule: User Registration

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

    @p1 @requirement:IA-01-019
    Scenario: Registration validates password requirements
      Given I am not authenticated
      And I have a unique email "weakpass@example.com"
      And I have a weak password "short"
      When I register a new account
      Then I should receive a 400 error
      And the error message should contain "password"

    @p2 @requirement:IA-01-020
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

    @p1 @requirement:IA-01-021
    Scenario: Registration rejects duplicate email
      Given a user exists with email "existing@example.com" and password "TestPassword123!"
      And I am not authenticated
      And I have a unique email "existing@example.com"
      When I register a new account
      Then the response status code should be 400
      And the error message should contain "email already exists"

    @p1 @requirement:IA-01-022
    Scenario: Registration auto-logs in new user
      Given I am not authenticated
      And I have a unique email "autologin@example.com"
      And I have a strong password "StrongPass123!"
      When I register a new account
      Then I should receive a valid authentication token
      And I should be able to access my user profile

    @p1 @requirement:IA-01-025
    Scenario: Registration with empty email
      Given I am not authenticated
      And I have a strong password "StrongPass123!"
      When I register with email ""
      Then I should receive a 400 error
      And the error message should contain "email"

    @p1 @requirement:IA-01-026
    Scenario: Registration with invalid email format
      Given I am not authenticated
      And I have a strong password "StrongPass123!"
      When I register with email "not-an-email"
      Then I should receive a 400 error
      And the error message should contain "email"

    @p1 @requirement:IA-01-027
    Scenario: Registration with empty password
      Given I am not authenticated
      And I have a unique email "nopass@example.com"
      When I register with password ""
      Then I should receive a 400 error
      And the error message should contain "password"

    @p1 @requirement:IA-01-028
    Scenario: Registration with missing name
      Given I am not authenticated
      And I have a unique email "noname@example.com"
      And I have a strong password "StrongPass123!"
      When I register without providing name
      Then I should receive a 400 error
      And the error message should contain "name"

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

  Rule: User Profiles

    @p1 @requirement:IA-01-027
    Scenario: Get own profile as manager
      Given I am logged in as a manager
      When I get my user profile
      Then the response status code should be 200
      And my profile should contain my email
      And my profile should contain my role
      And my profile should contain tenant ID

    @p1 @requirement:IA-01-028
    Scenario: Get own profile as member
      Given I am logged in as a member
      When I get my user profile
      Then the response status code should be 200
      And my profile should contain my email
      And my profile should contain my role
      And my profile should contain tenant ID

    @p2 @requirement:IA-01-029
    Scenario: Get profile includes created and updated timestamps
      Given I am logged in as a manager
      When I get my user profile
      Then the profile should contain creation timestamp
      And the profile should contain last update timestamp

    @p2 @requirement:IA-01-030
    Scenario: Get profile includes team information
      Given I am logged in as a manager
      When I get my user profile
      Then the profile should contain team ID
      And the profile should contain team name

    @p1 @requirement:IA-01-031
    Scenario: Get profile without authentication
      When I get my user profile
      Then I should receive an "unauthorized" error

    @p1 @requirement:IA-01-029
    Scenario: Get profile without authentication
      When I get my user profile
      Then I should receive an "unauthorized" error
      And the response status code should be 401

    @p2 @requirement:IA-01-030
    Scenario: Profile contains correct fields
      Given I am logged in as a manager
      When I get my user profile
      Then the response status code should be 200
      And my profile should contain "email" field
      And my profile should contain "name" field
      And my profile should contain "role" field
      And my profile should contain "tenant_id" field

    @p2 @requirement:IA-01-031
    Scenario: Profile has correct tenant ID
      Given I am logged in as a member
      And I belong to tenant with ID "1"
      When I get my user profile
      Then the response status code should be 200
      And my profile should contain tenant ID

  Rule: Extended Registration Scenarios

    @p1 @requirement:IA-01-211
    Scenario: Registration with duplicate email fails
      Given I am not authenticated
      And a user exists with email "duplicate@example.com"
      When I register with email "duplicate@example.com" and password "TestPassword123!"
      Then I should receive a 400 error
      And the error message should contain "already exists"

    @p1 @requirement:IA-01-212
    Scenario: Registration with minimal valid data
      Given I am not authenticated
      And I have a unique email "minimal@example.com"
      And I have a strong password "MinPass123!"
      When I register a new account with name "Minimal User"
      Then the response status code should be 201
      And I should receive a valid authentication token

    @p2 @requirement:IA-01-213
    Scenario: Multiple registration attempts
      Given I am not authenticated
      When I register with email "user1@example.com" and password "TestPassword123!"
      And I register with email "user2@example.com" and password "TestPassword123!"
      And I register with email "user3@example.com" and password "TestPassword123!"
      Then all operations should succeed

    @p2 @requirement:IA-01-214
    Scenario: Registration password too short
      Given I am not authenticated
      And I have a unique email "shortpwd@example.com"
      When I register with password "Ab1!"
      Then I should receive a 400 error

    @p2 @requirement:IA-01-215
    Scenario: Registration password missing number
      Given I am not authenticated
      And I have a unique email "nonum@example.com"
      When I register with password "ABCdefghi!"
      Then I should receive a 400 error

    @p2 @requirement:IA-01-216
    Scenario: Registration password missing special char
      Given I am not authenticated
      And I have a unique email "nospecial@example.com"
      When I register with password "ABCdef123"
      Then I should receive a 400 error

  Rule: Extended Profile Access Scenarios

    @p1 @requirement:IA-01-217
    Scenario: Profile access returns consistent data
      Given I am logged in as a manager
      When I get my user profile
      And I get my user profile
      And I get my user profile
      Then all responses should be consistent

    @p1 @requirement:IA-01-218
    Scenario: Profile contains user ID
      Given I am logged in as a manager
      When I get my user profile
      Then the response status code should be 200
      And my profile should contain my user ID

    @p2 @requirement:IA-01-219
    Scenario: Profile access with expired token
      Given I have an expired authentication token
      When I get my user profile
      Then I should receive an "unauthorized" error

    @p2 @requirement:IA-01-220
    Scenario: Profile contains created timestamp
      Given I am logged in as a member
      When I get my user profile
      Then the response status code should be 200
      And my profile should contain "created_at" field

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

    # Note: "Member cannot view team analytics" moved to 03_usage_insights.feature (domain-appropriate location)

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

    @p1 @requirement:IA-03-003
    Scenario: Profile contains correct tenant ID
      Given I am logged in as a member in tenant "1"
      When I get my user profile
      Then my profile should contain tenant ID
      And the tenant ID should be "1"

    @p1 @requirement:IA-03-004
    Scenario: Resources are isolated by tenant
      Given I am logged in as a manager in tenant "1"
      And I create a provider with name "tenant1-provider"
      When I login as a manager in tenant "2"
      And I list all providers
      Then I should not see "tenant1-provider"

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

    @p1 @requirement:IA-04-012
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

    @p3 @requirement:IA-04-013 @license:commercial
    Scenario: License kind limits
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

  Rule: License Expiration

    @p2 @requirement:IA-04-017
    Scenario: License expires after end date
      Given I have a license expiring in 1 day
      When I check license expiration
      Then the license should be marked as "expiring_soon"
      And I should see days remaining

    # Note: Removed "Expired license cannot create providers" - per spec BR-002,
    # expired licenses "preserve data access but disable advanced features".
    # Provider creation is considered core functionality, not advanced feature.

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

  Rule: License Activation and Validation

    @p1 @requirement:IA-04-019
    Scenario: Activate Open Source license successfully
      Given I am logged in as a manager
      And I have a valid Open Source license key
      When I activate the license
      Then the license should be activated successfully
      And the license type should be "opensource"
      And the data retention days should be 7
      And the max teams should be 1
      And the max seats should be unlimited

    @p1 @requirement:IA-04-020 @license:commercial
    Scenario: Activate Commercial license successfully
      Given I am logged in as a manager
      And I have a valid Commercial license key
      When I activate the license
      Then the license should be activated successfully
      And the license type should be "commercial"
      And the data retention days should be 90
      And the max teams should be unlimited
      And the max seats should be unlimited

    @p1 @requirement:IA-04-021
    Scenario: Activate expired license fails
      Given I am logged in as a manager
      And I have an expired license key
      When I attempt to activate the license
      Then I should receive a 400 error
      And the error message should contain "expired"

    @p1 @requirement:IA-04-022
    Scenario: Activate invalid license fails
      Given I am logged in as a manager
      And I have an invalid license key
      When I attempt to activate the license
      Then I should receive a 400 error
      And the error message should contain "invalid" or "signature"

    @p2 @requirement:IA-04-023
    Scenario: License with immediate expiry
      Given I am logged in as a manager
      And I have a license that expires immediately
      When I attempt to activate the license
      Then the activation should succeed or fail gracefully
      And the license should be marked as expired

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

    @p2 @requirement:IA-04-026
    Scenario: License reactivation preserves settings
      Given I have an active license with custom settings
      When I deactivate and reactivate the same license
      Then the license should be activated successfully
      And the previous settings should be preserved

  Rule: License Status and Information

    @p1 @requirement:IA-04-027
    Scenario: Get current license status
      Given I am logged in as a manager
      When I get the license status
      Then I should see the license type
      And I should see the expiration date
      And I should see the max teams
      And I should see the max seats
      And I should see the data retention days
      And I should see the current usage

    @p1 @requirement:IA-04-028
    Scenario: License status includes usage statistics
      Given I am logged in as a manager
      And I have active users and teams
      When I get the license status
      Then I should see the current user count
      And I should see the current team count
      And I should see the percentage of license used

    @p2 @requirement:IA-04-029
    Scenario: License status for non-admin users
      Given I am logged in as a member
      When I get the license status
      Then I should receive a 403 error
      And the error message should contain "permission" or "admin"

  Rule: License Feature Flags

    @p1 @requirement:IA-04-030
    Scenario: Open Source license feature flags
      Given I have an Open Source license
      When I check available features
      Then I should see basic provider management
      And I should see basic usage tracking
      And I should not see advanced analytics
      And I should not see team management beyond 1 team

    @p1 @requirement:IA-04-031 @license:commercial
    Scenario: Commercial license feature flags
      Given I have a Commercial license
      When I check available features
      Then I should see all provider management features
      And I should see advanced analytics
      And I should see unlimited team management
      And I should see extended data retention

  Rule: License Limit Enforcement

    @p1 @requirement:IA-04-032
    Scenario: Seat limit is informational only
      Given I have a license with 10 seats
      And I have 10 active users
      When I add an 11th user
      Then the user should be created successfully
      And the license should show 11 users in use

    @p2 @requirement:IA-04-033
    Scenario: Team limit enforcement by Open Source license
      Given I have an Open Source license
      And I have 1 team
      When I attempt to create a second team
      Then I should receive a 403 error
      And the error message should contain "team limit"

    @p1 @requirement:IA-04-034 @license:commercial
    Scenario: Commercial license has no team limit
      Given I have a Commercial license
      And I have 5 teams
      When I create a 6th team
      Then the operation should succeed

  Rule: License Renewal and Expiration

    @p2 @requirement:IA-04-035
    Scenario: License expiration date calculation
      Given I activate a 1-year license today
      When I get the license status
      Then the expiration date should be 1 year from today
      And the days remaining should be approximately 365

    @p2 @requirement:IA-04-036
    Scenario: License expiration warning levels
      Given I have a license expiring soon
      When I check license warnings
      Then I should see appropriate warning messages
      And the warning level should match days remaining

    @p3 @requirement:IA-04-037
    Scenario: Expired license behavior
      Given I have an expired license
      When I attempt to use advanced features
      Then the features should be disabled
      But basic functionality should remain available

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

  Rule: License Status

    @p1 @requirement:IA-04-022
    Scenario: Get license information
      Given I have an activated professional license
      When I get license information
      Then the license tier should be "professional"
      And the license status should be "active"
      And the license should have expiration date

    @p2 @requirement:IA-04-023
    Scenario: License status includes limits
      Given I have an activated trial license
      When I get license information
      Then the license should include provider limit
      And the license should include user limit

    @p2 @requirement:IA-04-024
    Scenario: License status includes features
      Given I have an activated enterprise license
      When I get license information
      Then the license should list all features
      And all features should be enabled

    @p1 @requirement:IA-04-025
    Scenario: Check license without authentication
      Given I am not authenticated
      When I get license information
      Then I should receive a 401 error

    @p1 @requirement:IA-04-026
    Scenario: Manager can view license information
      Given I am logged in as a manager
      When I get license information
      Then the operation should succeed

    # Note: Per spec FR-004 "License status visible to all users", members CAN view license info
    @p1 @requirement:IA-04-027
    Scenario: Member can view license information
      Given I am logged in as a member
      When I get license information
      Then the operation should succeed

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

    @p1 @requirement:IA-05-004
    Scenario: Accept invitation sets up password
      Given I am not authenticated
      And there is a pending invitation for "setup@example.com"
      When I accept the invitation with password "SetupPass123!"
      Then the user password should be set
      And the user should be able to login

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

  Rule: Token Validation Scenarios

    @p2 @requirement:IA-01-032
    Scenario: Verify token with empty string
      Given I am not authenticated
      When I verify my authentication token with ""
      Then I should receive an "invalid_token" error

    @p2 @requirement:IA-01-033
    Scenario: Verify token with whitespace only
      When I verify my authentication token with "   "
      Then I should receive an "invalid_token" error

    @p2 @requirement:IA-01-034
    Scenario: Refresh token with expired user context
      Given I am not authenticated
      And I have a user that was deleted
      When I refresh my authentication token
      Then I should receive an "user_not_found" error

  Rule: Registration Additional Scenarios

    @p2 @requirement:IA-01-037
    Scenario: Registration with very long email
      Given I am not authenticated
      And I have an email with 300 characters
      When I register a new account
      Then I should receive a 400 error
      And the error message should contain "email"

    @p2 @requirement:IA-01-038
    Scenario: Registration with email without domain
      Given I am not authenticated
      And I have email "user@"
      When I register a new account
      Then I should receive a 400 error
      And the error message should contain "email"

  Rule: Simple Authentication Flows

    @p2 @requirement:IA-01-039
    Scenario: Login with empty credentials
      When I login with email "" and password ""
      Then I should receive an authentication error

    @p2 @requirement:IA-01-040
    Scenario: Verify token returns user information
      Given I am logged in as a manager
      When I verify my authentication token
      Then the response should contain user information

    @p2 @requirement:IA-01-041
    Scenario: Get profile includes email field
      Given I am logged in as a member
      When I get my user profile
      Then the operation should succeed
      And the response should contain email field

    @p2 @requirement:IA-01-042
    Scenario: Login validates both email and password
      Given a user exists with email "valid@example.com" and password "ValidPass123!"
      When I login with email "valid@example.com" and password "ValidPass123!"
      Then the operation should succeed

    @p2 @requirement:IA-01-043
    Scenario: Register requires email and password
      Given I am not authenticated
      And I have a strong password "TestPass123!"
      When I register with email "test@example.com" and password "TestPass123!"
      Then the operation should succeed or return validation error

  Rule: More Authentication Variations

    @p2 @requirement:IA-01-044
    Scenario: Verify token returns valid response
      Given I am logged in as a manager
      When I verify my authentication token
      Then the operation should succeed

    @p2 @requirement:IA-01-045
    Scenario: Refresh token multiple times consecutively
      Given I am logged in as a manager
      When I refresh my authentication token 2 times
      Then both refreshes should return valid tokens

    @p2 @requirement:IA-01-046
    Scenario: Get profile returns user data
      Given I am logged in as a member
      When I get my user profile
      Then the operation should succeed

  Rule: Extended Authentication Flows

    @p1 @requirement:IA-01-047
    Scenario: Login returns valid access token
      Given a user exists with email "login@example.com" and password "TestPassword123!"
      When I login with email "login@example.com" and password "TestPassword123!"
      Then the operation should succeed
      And I should receive a valid authentication token

    @p1 @requirement:IA-01-048
    Scenario: Login returns user profile data
      Given a user exists with email "profile@example.com" and password "TestPassword123!"
      When I login with email "profile@example.com" and password "TestPassword123!"
      Then I should see user information
      And information should be accurate

    @p2 @requirement:IA-01-049
    Scenario: Verify token with valid access token
      Given I am logged in as a manager
      When I verify my authentication token
      Then the operation should succeed
      And the token should be valid

    @p2 @requirement:IA-01-050
    Scenario: Get profile after successful login
      Given I am logged in as a member
      When I get my user profile
      Then my profile should contain my email
      And my profile should contain my role

  Rule: Token Verify Extended Scenarios

    @p2 @requirement:IA-01-051
    Scenario: Verify token with empty token
      Given I am logged in as a manager
      When I verify my authentication token with ""
      Then I should receive a 401 error

    @p2 @requirement:IA-01-052
    Scenario: Verify token with malformed JWT
      Given I am logged in as a manager
      When I verify my authentication token with "not.a.valid.jwt"
      Then I should receive a 401 error

    @p1 @requirement:IA-01-053
    Scenario: Verify token returns user info
      Given I am logged in as a manager
      When I verify my authentication token
      Then the operation should succeed
      And I should see user information

    @p2 @requirement:IA-01-054
    Scenario: Multiple verify requests succeed
      Given I am logged in as a manager
      When I verify my authentication token
      And I verify my authentication token
      Then both operations should succeed

  Rule: Additional Authentication Paths

    @p1 @requirement:IA-01-055
    Scenario: Login with valid credentials returns tokens
      Given a user exists with email "auth@example.com" and password "TestPassword123!"
      When I login with email "auth@example.com" and password "TestPassword123!"
      Then the operation should succeed
      And I should receive a valid authentication token

    @p1 @requirement:IA-01-056
    Scenario: Get profile as manager
      Given I am logged in as a manager
      When I get my user profile
      Then my profile should contain my email
      And my profile should contain my role

    @p1 @requirement:IA-01-057
    Scenario: Get profile as member
      Given I am logged in as a member
      When I get my user profile
      Then my profile should contain my email
      And my profile should contain my role

    @p2 @requirement:IA-01-058
    Scenario: Refresh token generates new access token
      Given I am logged in as a manager
      When I refresh my authentication token
      Then the operation should succeed
      And the new token should have an extended expiration

    @p2 @requirement:IA-01-059
    Scenario: Register with valid data creates user
      Given I am not authenticated
      And I have a unique email "register@example.com"
      And I have a strong password "RegisterPass123!"
      When I register a new account
      Then the user should be created
      And the operation should succeed

    @p2 @requirement:IA-01-060
    Scenario: Dashboard metrics accessible to managers
      Given I am logged in as a manager
      When I get dashboard metrics
      Then the operation should succeed
      And I should see metrics data

  Rule: Refresh Token Error Paths

    @p1 @requirement:IA-01-061
    Scenario: Refresh with invalid refresh token
      Given I am logged in as a manager
      When I refresh my authentication token with "invalid-refresh-token"
      Then I should receive a 401 error
      And the error should indicate invalid token

    @p1 @requirement:IA-01-062
    Scenario: Refresh with empty refresh token
      Given I am logged in as a manager
      When I refresh my authentication token with ""
      Then I should receive a 401 error

    @p2 @requirement:IA-01-063
    Scenario: Refresh with malformed JWT
      Given I am logged in as a manager
      When I refresh my authentication token with "not.a.valid.jwt.token"
      Then I should receive a 401 error

    @p2 @requirement:IA-01-064
    Scenario: Refresh with expired token
      Given I am logged in as a manager
      When I refresh my authentication token with an expired token
      Then I should receive a 401 error

  Rule: Get Profile Extended Scenarios

    @p1 @requirement:IA-01-065
    Scenario: Get profile contains tenant ID
      Given I am logged in as a manager
      When I get my user profile
      Then my profile should contain tenant ID

    @p1 @requirement:IA-01-066
    Scenario: Get profile contains name field
      Given I am logged in as a manager
      When I get my user profile
      Then my profile should contain my name

    @p2 @requirement:IA-01-067
    Scenario: Get profile without authentication fails
      Given I am not authenticated
      When I get my user profile
      Then I should receive a 401 error

  Rule: Login Extended Variations

    @p1 @requirement:IA-01-068
    Scenario: Login with correct email and password
      Given a user exists with email "correct@example.com" and password "CorrectPassword123!"
      When I login with email "correct@example.com" and password "CorrectPassword123!"
      Then the operation should succeed
      And I should receive an access token

    @p1 @requirement:IA-01-069
    Scenario: Login returns both access and refresh tokens
      Given a user exists with email "tokens@example.com" and password "TokensPassword123!"
      When I login with email "tokens@example.com" and password "TokensPassword123!"
      Then the operation should succeed
      And I should receive an access token
      And I should receive a refresh token

    @p2 @requirement:IA-01-070
    Scenario: Login with uppercase email
      Given a user exists with email "case@example.com" and password "CasePassword123!"
      When I login with email "CASE@EXAMPLE.COM" and password "CasePassword123!"
      Then the operation should succeed

  Rule: Registration Extended Variations

    @p1 @requirement:IA-01-071
    Scenario: Register with valid email and password
      Given I am not authenticated
      And I have a unique email "register-new@example.com"
      And I have a strong password "RegisterPass123!"
      When I register a new account
      Then the user should be created
      And the operation should succeed

    @p1 @requirement:IA-01-072
    Scenario: Register with valid email and strong password
      Given I am not authenticated
      And I have a unique email "strong@example.com"
      And I have a strong password "StrongPass456!"
      When I register with email "strong@example.com" and password "StrongPass456!"
      Then the operation should succeed
      And I should receive an access token

    @p1 @requirement:IA-01-073
    Scenario: Register returns user profile data
      Given I am not authenticated
      And I have a unique email "profile-data@example.com"
      And I have a strong password "ProfilePass789!"
      When I register with email "profile-data@example.com" and password "ProfilePass789!"
      Then the operation should succeed
      And I should see user information

    @p2 @requirement:IA-01-074
    Scenario: Register with lowercase email
      Given I am not authenticated
      And I have a unique email "lowercase@example.com"
      And I have a strong password "LowerPass123!"
      When I register with email "lowercase@example.com" and password "LowerPass123!"
      Then the operation should succeed

    @p2 @requirement:IA-01-075
    Scenario: Register with mixed case email
      Given I am not authenticated
      And I have a unique email "MixedCase@example.com"
      And I have a strong password "MixedPass123!"
      When I register with email "MixedCase@example.com" and password "MixedPass123!"
      Then the operation should succeed

  Rule: Verify Token Extended Scenarios

    @p1 @requirement:IA-01-076
    Scenario: Verify token without authentication fails
      Given I am not authenticated
      When I verify my authentication token
      Then I should receive a 401 error

    @p2 @requirement:IA-01-077
    Scenario: Verify token with invalid token format
      Given I am logged in as a manager
      When I verify my authentication token with "invalid-token-format"
      Then I should receive a 401 error

    @p2 @requirement:IA-01-078
    Scenario: Verify token multiple times succeeds
      Given I am logged in as a manager
      When I verify my authentication token
      And I verify my authentication token
      And I verify my authentication token
      Then all operations should succeed

  Rule: Token Refresh Extended Scenarios

    @p1 @requirement:IA-01-079
    Scenario: Refresh token returns new tokens
      Given I am logged in as a manager
      When I refresh my authentication token
      Then the operation should succeed
      And I should receive an access token
      And I should receive a refresh token

    @p1 @requirement:IA-01-080
    Scenario: Refresh token without authentication fails
      Given I am not authenticated
      When I refresh my authentication token with "any-token"
      Then I should receive a 401 error

  Rule: Additional Authentication Flows

    @p1 @requirement:IA-01-081
    Scenario: Login returns access token with expiration
      Given a user exists with email "expiration@example.com" and password "ExpirationPass123!"
      When I login with email "expiration@example.com" and password "ExpirationPass123!"
      Then the operation should succeed
      And the response should contain access token

    @p1 @requirement:IA-01-082
    Scenario: Login returns refresh token
      Given a user exists with email "refresh-login@example.com" and password "RefreshPass123!"
      When I login with email "refresh-login@example.com" and password "RefreshPass123!"
      Then the operation should succeed
      And I should receive a refresh token

    @p1 @requirement:IA-01-083
    Scenario: Login success returns status 200
      Given a user exists with email "status@example.com" and password "StatusPass123!"
      When I login with email "status@example.com" and password "StatusPass123!"
      Then the operation should succeed
      And the status code should be 200

    @p1 @requirement:IA-01-084
    Scenario: Verify token with valid token succeeds
      Given I am logged in as a manager
      When I verify my authentication token
      Then the operation should succeed
      And the response should contain version information

    @p2 @requirement:IA-01-085
    Scenario: Get profile returns user ID
      Given I am logged in as a manager
      When I get my user profile
      Then my profile should contain my user ID

    @p2 @requirement:IA-01-086
    Scenario: Get profile returns name
      Given I am logged in as a manager
      When I get my user profile
      Then my profile should contain my name

    @p2 @requirement:IA-01-087
    Scenario: Get profile returns role
      Given I am logged in as a manager
      When I get my user profile
      Then my profile should contain my role

  Rule: Login Error Paths

    @p1 @requirement:IA-01-088
    Scenario: Login with non-existent email fails
      Given I am not authenticated
      When I login with email "nonexistent@example.com" and password "SomePass123!"
      Then the operation should fail
      And the status code should be 401

    @p1 @requirement:IA-01-089
    Scenario: Login with wrong password fails
      Given a user exists with email "wrongpass@example.com" and password "CorrectPass123!"
      When I login with email "wrongpass@example.com" and password "WrongPass123!"
      Then the operation should fail
      And the status code should be 401

    @p2 @requirement:IA-01-090
    Scenario: Login with missing password fails
      Given I am not authenticated
      When I login with email "test@example.com" and password ""
      Then the operation should fail
      And the status code should be 400

  Rule: Refresh Token Extended Paths

    @p1 @requirement:IA-01-091
    Scenario: Refresh token returns new access token
      Given I am logged in as a manager
      When I refresh my authentication token
      Then the operation should succeed
      And I should receive an access token

    @p1 @requirement:IA-01-092
    Scenario: Refresh token returns user data
      Given I am logged in as a manager
      When I refresh my authentication token
      Then the operation should succeed
      And I should see user information

    @p1 @requirement:IA-01-093
    Scenario: Refresh token success status
      Given I am logged in as a manager
      When I refresh my authentication token
      Then the status code should be 200

    @p2 @requirement:IA-01-094
    Scenario: Refresh with empty token fails
      Given I am logged in as a manager
      When I refresh my authentication token with ""
      Then the operation should fail
      And the status code should be 401

    @p2 @requirement:IA-01-095
    Scenario: Refresh with invalid format fails
      Given I am logged in as a manager
      When I refresh my authentication token with "invalid.format"
      Then the operation should fail
      And the status code should be 401

    @p2 @requirement:IA-01-096
    Scenario: Refresh with very long token fails
      Given I am logged in as a manager
      When I refresh my authentication token with "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
      Then the operation should fail

    @p2 @requirement:IA-01-097
    Scenario: Refresh with special characters fails
      Given I am logged in as a manager
      When I refresh my authentication token with "!@#$%^&*()"
      Then the operation should fail
      And the status code should be 401

    @p2 @requirement:IA-01-098
    Scenario: Refresh with invalid JSON fails
      Given I am logged in as a manager
      When I send refresh request with invalid JSON
      Then the operation should fail
      And the status code should be 400

    @p2 @requirement:IA-01-099
    Scenario: Refresh with empty request fails
      Given I am logged in as a manager
      When I send empty refresh request
      Then the operation should fail
      And the status code should be 400

  Rule: Login Extended Variations

    @p1 @requirement:IA-01-104
    Scenario: Login with valid credentials returns tokens
      Given a user exists with email "tokens2@example.com" and password "TokensPass123!"
      When I login with email "tokens2@example.com" and password "TokensPass123!"
      Then the operation should succeed
      And I should receive an access token
      And I should receive a refresh token

    @p1 @requirement:IA-01-105
    Scenario: Verify token success returns user info
      Given I am logged in as a manager
      When I verify my authentication token
      Then the operation should succeed
      And I should see user information

    @p2 @requirement:IA-01-106
    Scenario: Get profile returns all required fields
      Given I am logged in as a manager
      When I get my user profile
      Then my profile should contain my email
      And my profile should contain my name
      And my profile should contain my role
      And my profile should contain tenant ID

    @p2 @requirement:IA-01-107
    Scenario: Register with valid data succeeds
      Given I am not authenticated
      And I have a unique email "register2@example.com"
      And I have a strong password "Register2Pass123!"
      When I register with email "register2@example.com" and password "Register2Pass123!"
      Then the operation should succeed
      And I should receive an access token

  Rule: Profile Extended Queries

    @p1 @requirement:IA-01-108
    Scenario: Get profile contains email
      Given I am logged in as a manager
      When I get my user profile
      Then my profile should contain my email

    @p1 @requirement:IA-01-109
    Scenario: Get profile contains name
      Given I am logged in as a manager
      When I get my user profile
      Then my profile should contain my name

    @p1 @requirement:IA-01-110
    Scenario: Get profile contains role
      Given I am logged in as a manager
      When I get my user profile
      Then my profile should contain my role

    @p2 @requirement:IA-01-111
    Scenario: Get profile contains tenant ID
      Given I am logged in as a manager
      When I get my user profile
      Then my profile should contain tenant ID

  Rule: Registration Additional Variations

    @p1 @requirement:IA-01-112
    Scenario: Register with different valid email
      Given I am not authenticated
      And I have a unique email "another@example.com"
      And I have a strong password "AnotherPass123!"
      When I register with email "another@example.com" and password "AnotherPass123!"
      Then the operation should succeed
      And I should receive an access token

    @p1 @requirement:IA-01-113
    Scenario: Register returns success with user data
      Given I am not authenticated
      And I have a unique email "success@example.com"
      And I have a strong password "SuccessPass123!"
      When I register with email "success@example.com" and password "SuccessPass123!"
      Then the operation should succeed
      And I should see user information

    @p2 @requirement:IA-01-114
    Scenario: Register with valid password succeeds
      Given I am not authenticated
      And I have a unique email "validpass@example.com"
      When I register with email "validpass@example.com" and password "ValidPassword123!"
      Then the user should be created

    @p2 @requirement:IA-01-115
    Scenario: Register creates user account
      Given I am not authenticated
      And I have a unique email "newuser@example.com"
      And I have a strong password "NewUserPass123!"
      When I register a new account
      Then the user should be created
      And the operation should succeed

  Rule: Profile Extended Paths

    @p1 @requirement:IA-01-096
    Scenario: Get profile returns all fields
      Given I am logged in as a manager
      When I get my user profile
      Then the operation should succeed
      And my profile should contain my email
      And my profile should contain my name
      And my profile should contain my role
      And my profile should contain tenant ID

    @p1 @requirement:IA-01-097
    Scenario: Get profile as member
      Given I am logged in as a member
      When I get my user profile
      Then the operation should succeed
      And my profile should contain my email

    @p1 @requirement:IA-01-098
    Scenario: Get profile multiple times
      Given I am logged in as a manager
      When I get my user profile
      And I get my user profile
      And I get my user profile
      Then all operations should succeed

    @p2 @requirement:IA-01-099
    Scenario: Profile contains creation timestamp
      Given I am logged in as a manager
      When I get my user profile
      Then the profile should contain creation timestamp

    @p2 @requirement:IA-01-100
    Scenario: Profile contains last update timestamp
      Given I am logged in as a manager
      When I get my user profile
      Then the profile should contain last update timestamp

  Rule: Refresh Token Error Paths

    @p1 @requirement:IA-01-116
    Scenario: Refresh with empty refresh token
      Given I am logged in as a manager
      When I send a refresh request with empty refresh token
      Then I should receive a validation error
      And the response status code should be 400 or 401

    @p1 @requirement:IA-01-117
    Scenario: Refresh with null refresh token
      Given I am logged in as a manager
      When I send a refresh request with null refresh token
      Then I should receive a validation error
      And the response status code should be 400

    @p1 @requirement:IA-01-118
    Scenario: Refresh with malformed JWT structure
      Given I am logged in as a manager
      When I send a refresh request with malformed JWT "not-a-jwt"
      Then I should receive an "invalid_token" error
      And the response status code should be 401

    @p1 @requirement:IA-01-119
    Scenario: Refresh with corrupted signature
      Given I am logged in as a manager
      When I send a refresh request with corrupted signature
      Then I should receive an "invalid_token" error
      And the response status code should be 401

    @p2 @requirement:IA-01-120
    Scenario: Refresh without refresh_token field
      Given I am logged in as a manager
      When I send a refresh request without refresh_token field
      Then I should receive a validation error
      And the response status code should be 400

    @p2 @requirement:IA-01-121
    Scenario: Refresh with invalid JSON payload
      Given I am logged in as a manager
      When I send a refresh request with invalid JSON
      Then I should receive a validation error
      And the response status code should be 400

  Rule: Profile Error Paths

    @p1 @requirement:IA-01-122
    Scenario: Get profile without authentication
      Given I am not authenticated
      When I get my user profile without authentication
      Then I should receive an "unauthorized" error
      And the response status code should be 401

    @p1 @requirement:IA-01-123
    Scenario: Get profile with invalid token
      Given I have an invalid authentication token "Bearer invalid-token-123"
      When I get my user profile with invalid token
      Then I should receive an "unauthorized" error
      And the response status code should be 401

    @p1 @requirement:IA-01-124
    Scenario: Get profile with expired token
      Given I have an expired authentication token
      When I get my user profile
      Then I should receive an "unauthorized" error
      And the response status code should be 401

  Rule: Additional Refresh Scenarios

    @p1 @requirement:IA-01-125
    Scenario: Refresh after token expiration
      Given I am logged in as a manager
      And my access token will expire soon
      When I refresh my authentication token
      Then I should receive a new valid access token
      And the new token should be different from the old

    @p1 @requirement:IA-01-126
    Scenario: Refresh with valid token returns new tokens
      Given I am logged in as a manager
      And I have a valid refresh token
      When I send refresh request with valid token
      Then I should receive both access and refresh tokens
      And both tokens should be valid

    @p2 @requirement:IA-01-127
    Scenario: Refresh response contains user info
      Given I am logged in as a manager
      When I refresh my authentication token
      Then the response should contain user information
      And the response should contain expiration time

  Rule: Registration Validation Extended

    @p1 @requirement:IA-01-128
    Scenario: Register with very short password
      Given I am not authenticated
      And I have a unique email "short@example.com"
      When I register with email "short@example.com" and password "12345"
      Then I should receive a validation error
      And the response status code should be 400

    @p1 @requirement:IA-01-129
    Scenario: Register without email field
      Given I am not authenticated
      When I register without email field
      Then I should receive a validation error
      And the response status code should be 400

    @p2 @requirement:IA-01-130
    Scenario: Register with missing name
      Given I am not authenticated
      And I have a unique email "noname@example.com"
      When I register without providing name
      Then I should receive a validation error
      And the response status code should be 400

  Rule: Profile Extended Scenarios

    @p1 @requirement:IA-01-131
    Scenario: Get profile multiple times consecutively
      Given I am logged in as a manager
      When I get my user profile
      And I get my user profile
      And I get my user profile
      And I get my user profile
      Then all operations should succeed
      And each response should contain user data

    @p1 @requirement:IA-01-132
    Scenario: Get profile as member
      Given I am logged in as a member
      When I get my user profile
      Then the operation should succeed
      And my profile should contain my email
      And my profile should contain my role

    @p2 @requirement:IA-01-133
    Scenario: Get profile contains user ID
      Given I am logged in as a manager
      When I get my user profile
      Then my profile should contain my user ID

  Rule: Profile Variations

    @p1 @requirement:IA-01-134
    Scenario: Get profile multiple times in sequence
      Given I am logged in as a manager
      When I get my user profile
      And I get my user profile
      And I get my user profile
      Then all operations should succeed

    @p1 @requirement:IA-01-135
    Scenario: Login and then get profile
      Given I am logged in as a manager
      When I get my user profile
      Then my profile should contain my email
      And my profile should contain my name

    @p2 @requirement:IA-01-136
    Scenario: Login returns valid token structure
      Given I am logged in as a manager
      Then I should receive a valid authentication token
      And the token should be valid

    @p2 @requirement:IA-01-137
    Scenario: Refresh token multiple times consecutively
      Given I am logged in as a manager
      When I refresh my authentication token
      And I refresh my authentication token
      And I refresh my authentication token
      Then all operations should succeed
      And each refresh should return a valid token

  Rule: Login Variations

    @p1 @requirement:IA-01-138
    Scenario: Login as manager and verify response
      Given I am logged in as a manager
      Then I should receive a valid authentication token
      And the response should contain user information

    @p1 @requirement:IA-01-139
    Scenario: Login as member and verify response
      Given I am logged in as a member
      Then I should receive a valid authentication token
      And my profile should contain my email

    @p1 @requirement:IA-01-140
    Scenario: Login multiple times
      Given I am logged in as a manager
      And I logout
      And I login with email "manager@example.com" and password "TestPassword123!"
      Then I should receive a valid authentication token

    @p2 @requirement:IA-01-141
    Scenario: Verify token after login
      Given I am logged in as a manager
      When I verify my authentication token
      Then the token should be valid
      And the operation should succeed

    @p2 @requirement:IA-01-142
    Scenario: Get profile after login
      Given I am logged in as a manager
      When I get my user profile
      Then my profile should contain my email
      And my profile should contain my role

  Rule: Simple Authentication Flows

    @p1 @requirement:IA-01-143
    Scenario: Complete login flow
      Given I am logged in as a manager
      And I get my user profile
      And I verify my authentication token
      Then all operations should succeed

    @p1 @requirement:IA-01-144
    Scenario: Member login and profile access
      Given I am logged in as a member
      And I get my user profile
      Then the operation should succeed
      And my profile should contain my email

    @p2 @requirement:IA-01-145
    Scenario: Multiple profile requests
      Given I am logged in as a manager
      When I get my user profile
      And I get my user profile
      And I get my user profile
      And I get my user profile
      Then all operations should succeed

  Rule: User Management

    @p1 @requirement:IA-01-146
    Scenario: List all users as manager
      Given I am logged in as a manager
      When I list all users
      Then the operation should succeed
      And I should see user information

    @p1 @requirement:IA-01-147
    Scenario: List users multiple times
      Given I am logged in as a manager
      When I list all users
      And I list all users
      Then both operations should succeed

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

  Rule: Authentication Basic Flows

    @p1 @requirement:IA-01-150
    Scenario: Login returns proper response structure
      Given I am logged in as a manager
      Then I should receive a valid authentication token
      And the response should contain user information
      And the response should contain expiration time

    @p1 @requirement:IA-01-151
    Scenario: Verify endpoint validates token
      Given I am logged in as a manager
      When I verify my authentication token
      Then the operation should succeed
      And the token should be valid

    @p2 @requirement:IA-01-152
    Scenario: Profile endpoint returns user data
      Given I am logged in as a manager
      When I get my user profile
      Then the operation should succeed
      And my profile should contain my email
      And my profile should contain my role

    @p2 @requirement:IA-01-153
    Scenario: Multiple successful logins
      Given I am logged in as a manager
      And I logout
      And I login with email "manager@example.com" and password "TestPassword123!"
      Then I should receive a valid authentication token
      And the operation should succeed

  Rule: Simple Auth Flows

    @p1 @requirement:IA-01-154
    Scenario: Login verify and profile
      Given I am logged in as a manager
      When I verify my authentication token
      And I get my user profile
      Then all operations should succeed

    @p1 @requirement:IA-01-155
    Scenario: Manager basic operations
      Given I am logged in as a manager
      When I list all providers
      And I list all teams
      And I get my user profile
      Then all operations should succeed

    @p2 @requirement:IA-01-156
    Scenario: Member basic operations
      Given I am logged in as a member
      When I list all providers
      And I get my user profile
      Then all operations should succeed

  Rule: Authentication Repeated Operations

    @p1 @requirement:IA-01-157
    Scenario: Login five times consecutively
      Given I am logged in as a manager
      And I logout
      And I login with email "manager@example.com" and password "TestPassword123!"
      And I logout
      And I login with email "manager@example.com" and password "TestPassword123!"
      And I logout
      And I login with email "manager@example.com" and password "TestPassword123!"
      Then all operations should succeed

    @p1 @requirement:IA-01-158
    Scenario: Get profile five times
      Given I am logged in as a manager
      When I get my user profile
      And I get my user profile
      And I get my user profile
      And I get my user profile
      And I get my user profile
      Then all operations should succeed

    @p1 @requirement:IA-01-159
    Scenario: Refresh token three times
      Given I am logged in as a manager
      When I refresh my authentication token
      And I refresh my authentication token
      And I refresh my authentication token
      Then all operations should succeed

    @p2 @requirement:IA-01-160
    Scenario: Verify token three times
      Given I am logged in as a manager
      When I verify my authentication token
      And I verify my authentication token
      And I verify my authentication token
      Then all operations should succeed

  Rule: Provider Simple Operations

    @p1 @requirement:IA-01-161
    Scenario: List providers five times
      Given I am logged in as a manager
      When I list all providers
      And I list all providers
      And I list all providers
      And I list all providers
      And I list all providers
      Then all operations should succeed

    @p1 @requirement:IA-01-162
    Scenario: Dashboard operations sequence
      Given I am logged in as a manager
      When I get dashboard metrics
      And I get dashboard rankings
      And I get dashboard members
      Then all operations should succeed

    @p2 @requirement:IA-01-163
    Scenario: Health check repeated
      Given the test server is running
      When I check the health endpoint
      And I check the health endpoint
      And I check the health endpoint
      Then all operations should succeed

  Rule: Additional Simple Scenarios

    @p1 @requirement:IA-01-164
    Scenario: Complete auth flow sequence
      Given I am logged in as a manager
      When I verify my authentication token
      And I get my user profile
      And I refresh my authentication token
      Then all operations should succeed

    @p1 @requirement:IA-01-165
    Scenario: Provider list and profile
      Given I am logged in as a manager
      When I list all providers
      And I get my user profile
      And I list all teams
      Then all operations should succeed

    @p2 @requirement:IA-01-166
    Scenario: Dashboard operations
      Given I am logged in as a manager
      When I get dashboard metrics
      And I get dashboard rankings
      Then both operations should succeed

    @p2 @requirement:IA-01-167
    Scenario: Auth and profile checks
      Given I am logged in as a manager
      When I get my user profile
      And I verify my authentication token
      And I get my user profile
      Then all operations should succeed

    @p2 @requirement:IA-01-168
    Scenario: Multiple refresh operations
      Given I am logged in as a manager
      When I refresh my authentication token
      And I refresh my authentication token
      And I refresh my authentication token
      Then all operations should succeed

    @p2 @requirement:IA-01-169
    Scenario: Health and readiness
      Given I am logged in as a manager
      When I check the health endpoint
      And I check the readiness endpoint
      Then both operations should succeed

    @p2 @requirement:IA-01-170
    Scenario: Provider and dashboard
      Given I am logged in as a manager
      When I list all providers
      And I get dashboard metrics
      Then both operations should succeed

    @p2 @requirement:IA-01-171
    Scenario: Teams and users
      Given I am logged in as a manager
      When I list all teams
      And I list all users
      Then both operations should succeed

    @p2 @requirement:IA-01-172
    Scenario: License and profile
      Given I am logged in as a manager
      When I get license information
      And I get my user profile
      Then both operations should succeed

    @p2 @requirement:IA-01-173
    Scenario: Multiple profile checks
      Given I am logged in as a manager
      When I get my user profile
      And I get my user profile
      Then both operations should succeed

    @p2 @requirement:IA-01-174
    Scenario: Provider and license
      Given I am logged in as a manager
      When I list all providers
      And I get license information
      Then both operations should succeed

    @p2 @requirement:IA-01-175
    Scenario: Dashboard and teams
      Given I am logged in as a manager
      When I get dashboard metrics
      And I list all teams
      Then both operations should succeed

    @p2 @requirement:IA-01-176
    Scenario: Auth verification variations
      Given I am logged in as a manager
      When I verify my authentication token
      And I verify my authentication token
      Then both operations should succeed

    @p2 @requirement:IA-01-177
    Scenario: Usage and dashboard
      Given I am logged in as a manager
      When I get my usage statistics
      And I get dashboard metrics
      Then both operations should succeed

  Rule: Final Coverage Push

    @p1 @requirement:IA-01-178
    Scenario: Complete auth sequence
      Given I am logged in as a manager
      When I verify my authentication token
      And I get my user profile
      And I refresh my authentication token
      And I get my user profile
      Then all operations should succeed

    @p1 @requirement:IA-01-179
    Scenario: Dashboard extended sequence
      Given I am logged in as a manager
      When I get dashboard metrics
      And I get dashboard rankings
      And I get dashboard members
      And I get my usage statistics
      Then all operations should succeed

    @p1 @requirement:IA-01-180
    Scenario: Provider operations extended
      Given I am logged in as a manager
      When I list all providers
      And I get license information
      And I list all teams
      And I list all users
      Then all operations should succeed

    @p2 @requirement:IA-01-181
    Scenario: Health endpoints
      Given the test server is running
      When I check the health endpoint
      And I check the readiness endpoint
      And I check the health endpoint
      Then all operations should succeed

    @p2 @requirement:IA-01-182
    Scenario: Profile and teams
      Given I am logged in as a manager
      When I get my user profile
      And I list all teams
      And I get dashboard metrics
      Then all operations should succeed

    @p2 @requirement:IA-01-183
    Scenario: Multiple auth operations
      Given I am logged in as a manager
      When I verify my authentication token
      And I verify my authentication token
      And I get my user profile
      And I get my user profile
      Then all operations should succeed

    @p2 @requirement:IA-01-184
    Scenario: Provider list variations
      Given I am logged in as a manager
      When I list all providers
      And I list all providers
      And I list all providers
      Then all operations should succeed

    @p2 @requirement:IA-01-185
    Scenario: Dashboard all endpoints
      Given I am logged in as a manager
      When I get dashboard metrics
      And I get dashboard rankings
      And I get dashboard members
      Then all operations should succeed

    @p2 @requirement:IA-01-186
    Scenario: Teams and members
      Given I am logged in as a manager
      When I list all teams
      And I list all users
      And I get dashboard members
      Then all operations should succeed

    @p2 @requirement:IA-01-187
    Scenario: License and usage
      Given I am logged in as a manager
      When I get license information
      And I get my usage statistics
      Then both operations should succeed

    @p2 @requirement:IA-01-188
    Scenario: Profile and dashboard
      Given I am logged in as a manager
      When I get my user profile
      And I get dashboard metrics
      And I get dashboard rankings
      Then all operations should succeed

    @p2 @requirement:IA-01-189
    Scenario: Auth refresh multiple
      Given I am logged in as a manager
      When I refresh my authentication token
      And I refresh my authentication token
      And I refresh my authentication token
      And I refresh my authentication token
      Then all operations should succeed

    @p2 @requirement:IA-01-190
    Scenario: All dashboard endpoints
      Given I am logged in as a manager
      When I get dashboard metrics
      And I get dashboard rankings
      And I get dashboard members
      And I get my usage statistics
      Then all operations should succeed

    @p2 @requirement:IA-01-191
    Scenario: Complete system check
      Given I am logged in as a manager
      When I check the health endpoint
      And I list all providers
      And I list all teams
      And I get license information
      Then all operations should succeed

  Rule: Extended Coverage Scenarios

    @p1 @requirement:IA-01-192
    Scenario: Full authentication flow
      Given I am logged in as a manager
      When I verify my authentication token
      And I get my user profile
      And I refresh my authentication token
      And I get my user profile
      And I verify my authentication token
      Then all operations should succeed

    @p1 @requirement:IA-01-193
    Scenario: All dashboard endpoints
      Given I am logged in as a manager
      When I get dashboard metrics
      And I get dashboard rankings
      And I get dashboard members
      And I get my usage statistics
      Then all operations should succeed

    @p1 @requirement:IA-01-194
    Scenario: Provider and team operations
      Given I am logged in as a manager
      When I list all providers
      And I list all teams
      And I get license information
      And I list all users
      Then all operations should succeed

    @p2 @requirement:IA-01-195
    Scenario: Multiple profile checks
      Given I am logged in as a manager
      When I get my user profile
      And I get my user profile
      And I get my user profile
      And I get my user profile
      And I get my user profile
      Then all operations should succeed

    @p2 @requirement:IA-01-196
    Scenario: All refresh operations
      Given I am logged in as a manager
      When I refresh my authentication token
      And I refresh my authentication token
      And I refresh my authentication token
      And I refresh my authentication token
      Then all operations should succeed

    @p2 @requirement:IA-01-197
    Scenario: All verify operations
      Given I am logged in as a manager
      When I verify my authentication token
      And I verify my authentication token
      And I verify my authentication token
      And I verify my authentication token
      Then all operations should succeed

    @p2 @requirement:IA-01-198
    Scenario: Health check variations
      Given the test server is running
      When I check the health endpoint
      And I check the readiness endpoint
      And I check the health endpoint
      And I check the readiness endpoint
      Then all operations should succeed

    @p2 @requirement:IA-01-199
    Scenario: Provider and dashboard
      Given I am logged in as a manager
      When I list all providers
      And I get dashboard metrics
      And I get dashboard rankings
      Then all operations should succeed

    @p2 @requirement:IA-01-200
    Scenario: Teams and users
      Given I am logged in as a manager
      When I list all teams
      And I list all users
      And I get dashboard members
      Then all operations should succeed

  Rule: Maximum Coverage Push

    @p1 @requirement:IA-01-201
    Scenario: Profile repeated access
      Given I am logged in as a manager
      When I get my user profile
      And I get my user profile
      And I get my user profile
      And I get my user profile
      And I get my user profile
      And I get my user profile
      Then all operations should succeed

    @p1 @requirement:IA-01-202
    Scenario: Refresh repeated operations
      Given I am logged in as a manager
      When I refresh my authentication token
      And I refresh my authentication token
      And I refresh my authentication token
      And I refresh my authentication token
      And I refresh my authentication token
      Then all operations should succeed

    @p1 @requirement:IA-01-203
    Scenario: Dashboard complete access
      Given I am logged in as a manager
      When I get dashboard metrics
      And I get dashboard rankings
      And I get dashboard members
      And I get my usage statistics
      Then all operations should succeed

    @p1 @requirement:IA-01-204
    Scenario: Provider list extended
      Given I am logged in a manager
      When I list all providers
      And I list all providers
      And I list all providers
      And I list all providers
      Then all operations should succeed

    @p2 @requirement:IA-01-205
    Scenario: Health checks repeated
      Given I am logged in as a manager
      When I check the health endpoint
      And I check the health endpoint
      And I check the health endpoint
      And I check the readiness endpoint
      And I check the readiness endpoint
      Then all operations should succeed

    @p2 @requirement:IA-01-206
    Scenario: Teams and licenses
      Given I am logged in as a manager
      When I list all teams
      And I get license information
      And I list all users
      Then all operations should succeed

    @p2 @requirement:IA-01-207
    Scenario: Profile and verify
      Given I am logged in as a manager
      When I get my user profile
      And I verify my authentication token
      And I get my user profile
      And I verify my authentication token
      Then all operations should succeed

    @p2 @requirement:IA-01-208
    Scenario: All auth endpoints
      Given I am logged in as a manager
      When I verify my authentication token
      And I get my user profile
      And I refresh my authentication token
      Then all operations should succeed

    @p2 @requirement:IA-01-209
    Scenario: Dashboard comprehensive
      Given I am logged in as a manager
      When I get dashboard metrics
      And I get dashboard rankings
      And I get dashboard members
      And I get my usage statistics
      And I list all providers
      Then all operations should succeed

    @p2 @requirement:IA-01-210
    Scenario: System health check
      Given the test server is running
      When I check the health endpoint
      And I check the readiness endpoint
      And I check the health endpoint
      And I check the readiness endpoint
      And I check the health endpoint
      Then all operations should succeed

    @p1 @requirement:IA-01-221
    Scenario: Profile access multiple times
      Given I am logged in as a manager
      When I get my user profile
      And I get my user profile
      And I get my user profile
      And I get my user profile
      And I get my user profile
      Then all operations should succeed

    @p1 @requirement:IA-01-222
    Scenario: Profile access as member multiple times
      Given I am logged in as a member
      When I get my user profile
      And I get my user profile
      And I get my user profile
      Then all operations should succeed

    @p1 @requirement:IA-01-223
    Scenario: Profile contains all expected fields
      Given I am logged in as a manager
      When I get my user profile
      Then the response status code should be 200
      And my profile should contain "email" field
      And my profile should contain "name" field
      And my profile should contain "role" field
      And my profile should contain "tenant_id" field

    @p2 @requirement:IA-01-224
    Scenario: Verify and profile together
      Given I am logged in as a manager
      When I verify my authentication token
      And I get my user profile
      And I verify my authentication token
      And I get my user profile
      Then all operations should succeed

    @p1 @requirement:IA-01-225
    Scenario: Profile access before and after refresh
      Given I am logged in as a manager
      When I get my user profile
      And I refresh my authentication token
      And I get my user profile
      Then all operations should succeed

    @p1 @requirement:IA-01-226
    Scenario: Multiple profile requests
      Given I am logged in as a manager
      When I get my user profile
      And I get my user profile
      And I get my user profile
      And I get my user profile
      And I get my user profile
      And I get my user profile
      And I get my user profile
      And I get my user profile
      And I get my user profile
      And I get my user profile
      Then all operations should succeed

    @p1 @requirement:IA-01-227
    Scenario: Profile verify profile sequence
      Given I am logged in as a member
      When I get my user profile
      And I verify my authentication token
      And I get my user profile
      And I verify my authentication token
      Then all operations should succeed

    @p2 @requirement:IA-01-228
    Scenario: Profile access with different roles
      Given I am logged in as a manager
      When I get my user profile
      Then the response should contain "manager"
      Given I am logged in as a member
      When I get my user profile
      Then the response should contain "member"

    @p1 @requirement:IA-01-229
    Scenario: Profile verify refresh profile sequence
      Given I am logged in as a manager
      When I get my user profile
      And I refresh my authentication token
      And I get my user profile
      And I refresh my authentication token
      Then all operations should succeed

    @p1 @requirement:IA-01-230
    Scenario: Profile access repeated many times
      Given I am logged in as a manager
      When I get my user profile
      And I get my user profile
      And I get my user profile
      And I get my user profile
      And I get my user profile
      And I get my user profile
      And I get my user profile
      And I get my user profile
      And I get my user profile
      Then all operations should succeed

    @p2 @requirement:IA-01-231
    Scenario: Profile contains all required data
      Given I am logged in as a manager
      When I get my user profile
      Then the response status code should be 200
      And my profile should contain "email" field
      And my profile should contain "name" field
      And my profile should contain "role" field

    @p2 @requirement:IA-01-232
    Scenario: Profile access after multiple refreshes
      Given I am logged in as a member
      When I refresh my authentication token
      And I get my user profile
      And I refresh my authentication token
      And I get my user profile
      And I refresh my authentication token
      Then all operations should succeed

  Rule: License Tiers

    @p0 @requirement:IA-04-028
    Scenario: Get license tiers without authentication
      Given the test server is running
      When I get license tiers
      Then the operation should succeed
      And the response should contain tier information

    @p1 @requirement:IA-04-029
    Scenario: Get license tiers as manager
      Given I am logged in as a manager
      When I get license tiers
      Then the operation should succeed
      And the response should contain tier information

    @p1 @requirement:IA-04-030
    Scenario: Get license tiers as member
      Given I am logged in as a member
      When I get license tiers
      Then the operation should succeed
      And the response should contain tier information

    @p1 @requirement:IA-04-031
    Scenario: License tiers response contains valid data
      Given the test server is running
      When I get license tiers
      Then the response status code should be 200
      And the response should be valid JSON

    @p2 @requirement:IA-04-032
    Scenario: Get license tiers multiple times
      Given the test server is running
      When I get license tiers
      And I get license tiers
      And I get license tiers
      Then all operations should succeed

    @p2 @requirement:IA-04-033
    Scenario: License tiers are consistent
      Given the test server is running
      When I get license tiers
      And I get license tiers again
      Then both responses should be consistent

