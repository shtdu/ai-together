# BDD Test Suite - Manager Dashboard
# This feature covers dashboard operations, team management, and user management

Feature: Manager Dashboard
  As a manager
  I want to view team analytics and manage users
  So that I can oversee my team's AI usage

  Background:
    Given the test server is running
    And I am logged in as a manager

  Rule: Team Management

    @p1 @requirement:UI-05-001
    Scenario: Create a team
      Given I have a unique team name
      When I create a team with name "My Team"
      Then the team should be created
      And the team should have ID

    @p1 @requirement:UI-05-002
    Scenario: Update team settings
      Given I have created a team
      When I update team settings
      Then the team should be updated
      And settings should be saved

    @p1 @requirement:UI-05-003
    Scenario: Add team member
      Given I have created a team
      And I have a user "user@example.com"
      When I add team member "user@example.com"
      Then the user should be added to team
      And team member count should increase

    @p1 @requirement:UI-05-004
    Scenario: Remove team member
      Given I have a team with members
      When I remove team member
      Then member should be removed
      And team member count should decrease

    @p1 @requirement:UI-05-005
    Scenario: List all teams
      When I list all teams
      Then I should see at least 1 team
      And each team should have name

    @p1 @requirement:UI-05-006
    Scenario: Delete team
      Given I have created a team
      When I delete the team
      Then the team should not exist

    @p1 @requirement:UI-05-007
    Scenario: Cannot delete default team
      When I attempt to delete default team
      Then I should receive a 403 error

    @wip @p2 @requirement:UI-05-008 @license:commercial
    Scenario: Create team with minimal data
      # Requires commercial license - open source tier allows only 1 team (default team)
      Given I have a unique team name
      When I create a team with name "Minimal Team" and no description
      Then the team should be created
      And the team description should be empty

    @p1 @requirement:UI-05-009
    Scenario: Get team settings
      Given I have created a team
      When I get team settings
      Then I should receive team settings
      And settings should have team ID

    @p1 @requirement:UI-05-010
    Scenario: Get non-existent team
      When I attempt to get team with ID 99999
      Then I should receive a 404 error

    @p1 @requirement:UI-05-011
    Scenario: Delete non-existent team
      When I attempt to delete team with ID 99999
      Then I should receive a 404 error

    @p2 @requirement:UI-05-012
    Scenario: Remove non-existent team member
      Given I have created a team
      When I attempt to remove member with ID 99999
      Then the operation should succeed

    @p2 @requirement:UI-05-013
    Scenario: Update team name
      Given I have created a team
      When I update the team name to "Updated Team"
      Then the team name should be "Updated Team"

    @p2 @requirement:UI-05-014
    Scenario: Update team description
      Given I have created a team
      When I update the team description
      Then the description should be updated

  Rule: User Management

    @p1 @requirement:UI-05-015
    Scenario: Create a user
      Given I have a unique user "newuser@example.com"
      When I create a user
      Then the user should be created
      And the user should have ID

    @p1 @requirement:UI-05-016
    Scenario: Update user password
      Given I have a user "user@example.com"
      When I update user password
      Then password should be updated
      And old password should not work

    @p1 @requirement:UI-05-017
    Scenario: Deactivate user
      Given I have an active user
      When I deactivate user
      Then user should be deactivated
      And user cannot login

    @p1 @requirement:UI-05-018
    Scenario: Reactivate user
      Given I have a deactivated user
      When I reactivate user
      Then user should be activated
      And user can login

    @p1 @requirement:UI-05-019
    Scenario: List all users
      When I list all users
      Then I should see all users
      And each user should have email

    @p1 @requirement:UI-05-020
    Scenario: Get user details
      Given I have a user "user@example.com"
      When I get user details
      Then I should see user information
      And information should be accurate

    @wip @p1 @requirement:UI-05-021
    Scenario: Member cannot create users
      Given I am logged in as a member
      When I attempt to create a user
      Then I should receive a 403 error

    @p2 @requirement:UI-05-022
    Scenario: Update user role to manager
      Given I have a user with role "member"
      When I update user role to "manager"
      Then the user should have manager role

    @p2 @requirement:UI-05-023
    Scenario: Update user role to member
      Given I have a user with role "manager"
      And there are at least 2 managers
      When I update user role to "member"
      Then the user should have member role

    @p1 @requirement:UI-05-024
    Scenario: Manager cannot delete themselves
      Given I am logged in as a manager
      And I am the only manager
      When I attempt to delete myself
      Then I should receive a 403 error

    @p2 @requirement:UI-05-025
    Scenario: Get user without authentication
      Given I am not authenticated
      When I attempt to get user details
      Then I should receive a 401 error

  Rule: Dashboard Metrics

    @p1 @requirement:UI-05-026
    Scenario: Get dashboard metrics
      When I get dashboard metrics
      Then I should see total usage
      And I should see total cost
      And I should see active users

    @p1 @requirement:UI-05-027
    Scenario: Get team usage summary
      Given I have team usage data
      When I get team usage summary
      Then I should see team totals
      And I should see per-user breakdown

    @p1 @requirement:UI-05-028
    Scenario: Get cost trends
      When I get cost trends
      Then I should see daily costs
      And I should see trend line

    @p1 @requirement:UI-05-029
    Scenario: Get user rankings
      When I get user rankings
      Then I should see users ranked by usage
      And top user should be listed first

    @p1 @requirement:UI-05-030
    Scenario: Get provider performance
      When I get provider performance
      Then I should see provider statistics
      And I should see success rates

  Rule: Dashboard Permissions

    @p1 @requirement:UI-05-031
    Scenario: Manager can access dashboard
      Given I am logged in as a manager
      When I get dashboard metrics
      Then the operation should succeed

    @p1 @requirement:UI-05-032
    Scenario: Member can access their own usage statistics
      Given I am logged in as a member
      When I get dashboard metrics
      Then the operation should succeed
      And I should see total usage

    @wip @p1 @requirement:UI-05-033 @license:commercial
    Scenario: Manager can manage teams
      Given I am logged in as a manager
      When I create a team
      Then the operation should succeed

    @wip @p1 @requirement:UI-05-034
    Scenario: Member cannot manage teams
      Given I am logged in as a member
      When I attempt to create a team
      Then I should receive a 403 error
