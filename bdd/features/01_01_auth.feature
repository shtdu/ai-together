# BDD Test Suite - Authentication Operations
# This feature covers login, logout, token refresh, token verification, password reset, and account lockout

Feature: Authentication Operations
  As a system administrator
  I want to authenticate users securely and manage their sessions
  So that our team's data and configurations remain secure

  Background:
    Given the test server is running

  # ============================================================================
  # AUTH OPERATIONS - Login
  # ============================================================================

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

  # ============================================================================
  # AUTH OPERATIONS - Logout
  # ============================================================================

  Rule: Authentication - Logout

    @p1 @requirement:IA-01-006
    Scenario: Logout successfully
      Given I am logged in as a manager
      When I logout
      Then the response status code should be 204
      And my authentication token should be invalid

  # ============================================================================
  # AUTH OPERATIONS - Token Refresh
  # ============================================================================

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

  # ============================================================================
  # AUTH OPERATIONS - Token Verify
  # ============================================================================

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
  # PASSWORD RESET
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
  # ACCOUNT LOCKOUT
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
