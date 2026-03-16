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

    Scenario: Create a team
      Given I have a unique team name
      When I create a team with name "My Team"
      Then the team should be created
      And the team should have ID

    Scenario: Update team settings
      Given I have created a team
      When I update team settings
      Then the team should be updated
      And settings should be saved

    Scenario: Add team member
      Given I have created a team
      And I have a user "user@example.com"
      When I add team member "user@example.com"
      Then the user should be added to team
      And team member count should increase

    Scenario: Remove team member
      Given I have a team with members
      When I remove team member
      Then member should be removed
      And team member count should decrease

    Scenario: List all teams
      When I list all teams
      Then I should see at least 1 team
      And each team should have name

    Scenario: Delete team
      Given I have created a team
      When I delete the team
      Then the team should not exist

    Scenario: Cannot delete default team
      When I attempt to delete default team
      Then I should receive a 403 error

  Rule: User Management

    Scenario: Create a user
      Given I have a unique user "newuser@example.com"
      When I create a user
      Then the user should be created
      And the user should have ID

    Scenario: Update user password
      Given I have a user "user@example.com"
      When I update user password
      Then password should be updated
      And old password should not work

    Scenario: Deactivate user
      Given I have an active user
      When I deactivate user
      Then user should be deactivated
      And user cannot login

    Scenario: Reactivate user
      Given I have a deactivated user
      When I reactivate user
      Then user should be activated
      And user can login

    Scenario: List all users
      When I list all users
      Then I should see all users
      And each user should have email

    Scenario: Get user details
      Given I have a user "user@example.com"
      When I get user details
      Then I should see user information
      And information should be accurate

    Scenario: Member cannot create users
      Given I am logged in as a member
      When I attempt to create a user
      Then I should receive a 403 error

  Rule: Dashboard Metrics

    Scenario: Get dashboard metrics
      When I get dashboard metrics
      Then I should see total usage
      And I should see total cost
      And I should see active users

    Scenario: Get team usage summary
      Given I have team usage data
      When I get team usage summary
      Then I should see team totals
      And I should see per-user breakdown

    Scenario: Get cost trends
      When I get cost trends
      Then I should see daily costs
      And I should see trend line

    Scenario: Get user rankings
      When I get user rankings
      Then I should see users ranked by usage
      And top user should be listed first

    Scenario: Get provider performance
      When I get provider performance
      Then I should see provider statistics
      And I should see success rates

  Rule: Dashboard Permissions

    Scenario: Manager can access dashboard
      Given I am logged in as a manager
      When I get dashboard metrics
      Then the operation should succeed

    Scenario: Member cannot access dashboard
      Given I am logged in as a member
      When I get dashboard metrics
      Then I should receive a 403 error

    Scenario: Manager can manage teams
      Given I am logged in as a manager
      When I create a team
      Then the operation should succeed

    Scenario: Member cannot manage teams
      Given I am logged in as a member
      When I attempt to create a team
      Then I should receive a 403 error
