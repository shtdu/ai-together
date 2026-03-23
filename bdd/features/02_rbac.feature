# BDD Test Suite - Role-Based Access Control
# This feature covers RBAC, multi-tenant isolation, last manager protection, and user management

Feature: Role-Based Access Control
  As a system administrator
  I want to control user permissions and ensure data isolation
  So that users can only access appropriate resources and actions

  Background:
    Given the test server is running

  # ============================================================================
  # ROLE-BASED ACCESS CONTROL
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
  # MULTI-TENANT ISOLATION
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
  # LAST MANAGER PROTECTION
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
  # USER MANAGEMENT
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
