# BDD Test Suite - User Profile Operations
# This feature covers user profile retrieval, updates, and user registration

Feature: User Profile Operations
  As a system user
  I want to manage my profile information
  So that I can maintain accurate account details and join new organizations

  Background:
    Given the test server is running

  # ============================================================================
  # USER PROFILE - Happy Paths
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

  # ============================================================================
  # USER PROFILE - Field Validation
  # ============================================================================

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

  # ============================================================================
  # USER PROFILE - Error Paths
  # ============================================================================

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
  # USER REGISTRATION - Happy Paths
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

    @p2 @requirement:IA-01-018-B
    Scenario: Registration creates unique organization name
      Given I am not authenticated
      And I have a unique email "orgtest@example.com"
      When I register a new account
      Then the organization name should be auto-generated

  # ============================================================================
  # USER REGISTRATION - Password Validation
  # ============================================================================

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

  # ============================================================================
  # USER REGISTRATION - Error Paths
  # ============================================================================

  Rule: User Registration - Error Paths

    @p1 @requirement:IA-01-021
    Scenario: Registration rejects duplicate email
      Given a user exists with email "existing@example.com" and password "TestPassword123!"
      And I am not authenticated
      And I have a unique email "existing@example.com"
      When I register a new account
      Then the response status code should be 400
      And the error message should contain "email already exists"

    @p1 @requirement:IA-01-025-C
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
