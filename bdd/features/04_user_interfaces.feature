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
    Scenario: Create team with minimal data
      Given I have a unique team name
      When I create a team with name "Minimal Team" and no description
      Then the team should be created
      And the team description should be empty

    @p1 @requirement:UI-05-003
    Scenario: Create team with name and description
      Given I have a unique team name
      When I create a team with name "Full Team" and description "A team with description"
      Then the operation should succeed
      And the team should have the name

    @p1 @requirement:UI-05-004
    Scenario: Update team name
      Given I have created a team
      When I update the team name to "Updated Team"
      Then the team name should be "Updated Team"

    @p1 @requirement:UI-05-005
    Scenario: Update team description
      Given I have created a team
      When I update the team description
      Then the description should be updated

    @p1 @requirement:UI-05-006
    Scenario: Update team name and description together
      Given I have created a team
      When I update team name and description
      Then both fields should be updated

    @p1 @requirement:UI-05-007
    Scenario: Get team settings
      Given I have created a team
      When I get team settings
      Then I should receive team settings
      And settings should have team ID

    @p1 @requirement:UI-05-008
    Scenario: List all teams
      When I list all teams
      Then I should see at least 1 team
      And each team should have name

    @p1 @requirement:UI-05-009
    Scenario: List teams as member
      Given I am logged in as a member
      When I list all teams
      Then I should see at least 1 team

    @p1 @requirement:UI-05-010
    Scenario: Delete team
      Given I have created a team
      When I delete the team
      Then the team should not exist

    @p1 @requirement:UI-05-011
    Scenario: Cannot delete default team
      When I attempt to delete default team
      Then I should receive a 403 error

    @p1 @requirement:UI-05-012
    Scenario: Add team member
      Given I have created a team
      And I have a user "user@example.com"
      When I add team member "user@example.com"
      Then the user should be added to team
      And team member count should increase

    @p1 @requirement:UI-05-013
    Scenario: Remove team member
      Given I have a team with members
      When I remove team member
      Then member should be removed
      And team member count should decrease

    @p1 @requirement:UI-05-014
    Scenario: List team members
      Given I have created a team
      When I list team members
      Then I should see team members list
      And the operation should succeed

    @p2 @requirement:UI-05-015
    Scenario: Remove non-existent team member
      Given I have created a team
      When I attempt to remove member with ID 99999
      Then the operation should succeed

  Rule: User Management

    @p1 @requirement:UI-05-016
    Scenario: Create a user
      Given I have a unique user "newuser@example.com"
      When I create a user
      Then the user should be created
      And the user should have ID

    @p1 @requirement:UI-05-017
    Scenario: Update user password
      Given I have a user "user@example.com"
      When I update user password
      Then password should be updated
      And old password should not work

    @p1 @requirement:UI-05-018
    Scenario: Deactivate user
      Given I have an active user
      When I deactivate user
      Then user should be deactivated
      And user cannot login

    @p1 @requirement:UI-05-019
    Scenario: Reactivate user
      Given I have a deactivated user
      When I reactivate user
      Then user should be activated
      And user can login

    @p1 @requirement:UI-05-020
    Scenario: List all users
      When I list all users
      Then I should see all users
      And each user should have email

    @p1 @requirement:UI-05-021
    Scenario: List users includes email addresses
      When I list all users
      Then each user should have email

    @p1 @requirement:UI-05-022
    Scenario: List users includes roles
      When I list all users
      Then each user should have role

    @p1 @requirement:UI-05-023
    Scenario: Get user details
      Given I have a user "user@example.com"
      When I get user details
      Then I should see user information
      And information should be accurate

    @p1 @requirement:UI-05-024
    Scenario: Update user role to manager
      Given I have a user with role "member"
      When I update user role to "manager"
      Then the user should have manager role

    @p1 @requirement:UI-05-025
    Scenario: Update user role to member
      Given I have a user with role "manager"
      And there are at least 2 managers
      When I update user role to "member"
      Then the user should have member role

    @p1 @requirement:UI-05-026
    Scenario: Manager cannot delete themselves
      Given I am the only manager
      When I attempt to delete myself
      Then I should receive a 403 error

    @p1 @requirement:UI-05-027
    Scenario: Member cannot create users
      Given I am logged in as a member
      When I attempt to create a user
      Then I should receive a 403 error

    @p2 @requirement:UI-05-028
    Scenario: Update user with invalid role
      Given I have a user with role "member"
      When I update user role to "invalid_role"
      Then I should receive a 400 error

  Rule: Dashboard Metrics

    @p1 @requirement:UI-05-029
    Scenario: Get dashboard metrics
      When I get dashboard metrics
      Then I should see total usage
      And I should see total cost
      And I should see active users

    @p1 @requirement:UI-05-030
    Scenario Outline: Get dashboard metrics with different ranges
      When I get dashboard metrics with range "<range>"
      Then the operation should succeed

      Examples:
        | range |
        | 24h   |
        | 7d    |
        | 30d   |
        | 90d   |

    @p1 @requirement:UI-05-031
    Scenario: Get dashboard metrics as member
      Given I am logged in as a member
      When I get dashboard metrics
      Then the operation should succeed
      And I should see my usage data

    @p1 @requirement:UI-05-032
    Scenario: Get dashboard rankings
      When I get dashboard rankings
      Then I should see users ranked by usage
      And top user should be listed first

    @p1 @requirement:UI-05-033
    Scenario: Get dashboard members
      When I get dashboard members
      Then the operation should succeed
      And I should see members list
      And each member should have email

    @p1 @requirement:UI-05-034
    Scenario: Get team usage summary
      Given I have team usage data
      When I get team usage summary
      Then I should see team totals
      And I should see per-user breakdown

    @p1 @requirement:UI-05-035
    Scenario: Get cost trends
      When I get cost trends
      Then I should see daily costs
      And I should see trend line

    @p1 @requirement:UI-05-036
    Scenario: Get user rankings
      When I get user rankings
      Then I should see users ranked by usage

    @p1 @requirement:UI-05-037
    Scenario: Get provider performance
      When I get provider performance
      Then I should see provider statistics
      And I should see success rates

  Rule: Dashboard Permissions

    @p1 @requirement:UI-05-038
    Scenario: Manager can access dashboard
      Given I am logged in as a manager
      When I get dashboard metrics
      Then the operation should succeed

    @p1 @requirement:UI-05-039
    Scenario: Member can access their own usage statistics
      Given I am logged in as a member
      When I get dashboard metrics
      Then the operation should succeed
      And I should see total usage

    @p1 @requirement:UI-05-040
    Scenario: Manager can manage teams
      Given I am logged in as a manager
      When I create a team
      Then the operation should succeed

    @p1 @requirement:UI-05-041
    Scenario: Member cannot manage teams
      Given I am logged in as a member
      When I attempt to create a team
      Then I should receive a 403 error

    @p1 @requirement:UI-05-042
    Scenario: Member cannot access dashboard members endpoint
      Given I am logged in as a member
      When I get dashboard members
      Then I should receive a 403 error

    @p1 @requirement:UI-05-043
    Scenario: Manager can access provider statistics
      Given I have created a provider
      When I get provider statistics
      Then the operation should succeed

    @p1 @requirement:UI-05-044
    Scenario: Member cannot access provider statistics
      Given I am logged in as a member
      When I attempt to get provider statistics
      Then I should receive a 403 error

  Rule: User Profile

    @p1 @requirement:UI-05-045
    Scenario: Get user profile as manager
      When I get my user profile
      Then the operation should succeed
      And my profile should contain my email
      And my profile should contain my role
      And my profile should contain my team ID

    @p1 @requirement:UI-05-046
    Scenario: Get user profile as member
      Given I am logged in as a member
      When I get my user profile
      Then the operation should succeed
      And my profile should contain my email

    @p2 @requirement:UI-05-047
    Scenario: Get profile after role change
      Given I update my role
      When I get my user profile
      Then I should see the updated role

    @p2 @requirement:UI-05-048
    Scenario: Get profile after team change
      Given I change my team
      When I get my user profile
      Then I should see the updated team

  Rule: License Status

    @p1 @requirement:UI-05-049
    Scenario: Get license information as manager
      When I get license information
      Then the operation should succeed

    @p1 @requirement:UI-05-050
    Scenario: Get license information as member
      Given I am logged in as a member
      When I get license information
      Then the operation should succeed

    @p1 @requirement:UI-05-051
    Scenario: License status accessible to authenticated users
      Given I am logged in as a member
      When I get the license status
      Then the operation should succeed

  Rule: Provider Analytics

    @p1 @requirement:UI-05-052
    Scenario: Get provider analytics as manager
      Given I have created a provider
      When I get provider analytics
      Then the operation should succeed
      And I should see analytics data

    @p2 @requirement:UI-05-053
    Scenario: Get provider analytics filters by provider
      Given I have created a claude provider
      And I have created a codex provider
      When I get provider analytics for claude
      Then I should see claude analytics

    @p2 @requirement:UI-05-054
    Scenario: Get provider analytics with date range
      Given I have created a provider
      When I get provider analytics with date range
      Then the operation should succeed

  Rule: Dashboard Members Detail

    @p1 @requirement:UI-05-055
    Scenario: Get dashboard members includes user details
      Given there are multiple users in the team
      When I get dashboard members
      Then I should see multiple members
      And each member should have user ID

    @p2 @requirement:UI-05-056
    Scenario: Get dashboard members handles empty team
      Given I am the only user
      When I get dashboard members
      Then I should see at least 1 member

    @p2 @requirement:UI-05-057
    Scenario: Get dashboard members filters active users
      When I get dashboard members
      Then I should not see deleted users

    @p2 @requirement:UI-05-058
    Scenario: Get dashboard members returns usage data
      When I get dashboard members
      Then I should see usage data

  Rule: Dashboard Rankings Detail

    @p1 @requirement:UI-05-059
    Scenario: Get dashboard rankings shows usage order
      Given there are multiple users with different usage
      When I get dashboard rankings
      Then the operation should succeed
      And users should be ranked by usage

    @p2 @requirement:UI-05-060
    Scenario: Get dashboard rankings with time range
      When I get dashboard rankings
      Then the operation should succeed
      And I should see usage totals

  Rule: Authentication Error Paths

    @p1 @requirement:UI-05-061
    Scenario: Get user without authentication
      Given I am not authenticated
      When I attempt to get user details
      Then I should receive a 401 error

    @p1 @requirement:UI-05-062
    Scenario: Get dashboard metrics without authentication fails
      Given I am not authenticated
      When I get dashboard metrics
      Then I should receive a 401 error

    @p1 @requirement:UI-05-063
    Scenario: Get dashboard rankings without authentication fails
      Given I am not authenticated
      When I get dashboard rankings
      Then I should receive a 401 error

    @p1 @requirement:UI-05-064
    Scenario: Get dashboard members without authentication fails
      Given I am not authenticated
      When I get dashboard members
      Then I should receive a 401 error

    @p1 @requirement:UI-05-065
    Scenario: Get profile without authentication fails
      Given I am not authenticated
      When I get my user profile
      Then I should receive a 401 error

    @p1 @requirement:UI-05-066
    Scenario: Get usage statistics without authentication fails
      Given I am not authenticated
      When I get my usage statistics
      Then I should receive a 401 error

    @p1 @requirement:UI-05-067
    Scenario: List team members without authentication fails
      Given I am not authenticated
      When I list team members
      Then I should receive a 401 error

  Rule: Team Error Paths

    @p1 @requirement:UI-05-068
    Scenario: Get non-existent team
      When I attempt to get team with ID 99999
      Then I should receive a 404 error

    @p1 @requirement:UI-05-069
    Scenario: Delete non-existent team
      When I attempt to delete team with ID 99999
      Then I should receive a 404 error

    @p1 @requirement:UI-05-070
    Scenario: Get team settings for non-existent team
      When I attempt to get settings for team ID 99999
      Then I should receive a 404 error

    @p1 @requirement:UI-05-071
    Scenario: Update settings for non-existent team
      When I attempt to update settings for team ID 99999
      Then I should receive a 404 error

    @p1 @requirement:UI-05-072
    Scenario: Delete team with invalid ID fails
      When I attempt to delete team with ID 99999
      Then I should receive a 404 error

  Rule: User Error Paths

    @p1 @requirement:UI-05-073
    Scenario: Get non-existent user
      When I attempt to get user with ID 99999
      Then I should receive a 404 error

    @p1 @requirement:UI-05-074
    Scenario: Delete non-existent user
      When I attempt to delete user with ID 99999
      Then I should receive a 404 error

    @p2 @requirement:UI-05-075
    Scenario: Update non-existent user role
      When I update role for user ID 99999 to "manager"
      Then I should receive a 404 error

  Rule: Dashboard Extended

    @p1 @requirement:UI-05-076
    Scenario: All dashboard endpoints
      When I get dashboard metrics
      And I get dashboard rankings
      And I get dashboard members
      Then all operations should succeed

    @p1 @requirement:UI-05-077
    Scenario: Dashboard metrics with different ranges
      When I get dashboard metrics with range "24h"
      And I get dashboard metrics with range "7d"
      Then both operations should succeed

    @p2 @requirement:UI-05-078
    Scenario: Get dashboard metrics with invalid range
      When I get dashboard metrics with range "invalid"
      Then the operation should succeed or return validation error

    @p2 @requirement:UI-05-079
    Scenario: Get user profile and dashboard data
      When I get my user profile
      And I get dashboard metrics
      Then both operations should succeed

    @p2 @requirement:UI-05-080
    Scenario: Get dashboard members and metrics together
      When I get dashboard members
      And I get dashboard metrics
      Then both operations should succeed

  Rule: Team Extended

    @p1 @requirement:UI-05-081
    Scenario: Get default team
      When I get team with ID 1
      Then the operation should succeed
      And I should see team information

    @p1 @requirement:UI-05-082
    Scenario: Team settings operations together
      Given I have created a team
      When I get team settings
      And I update team name to "Updated Name"
      Then both operations should succeed

    @p2 @requirement:UI-05-083
    Scenario: Get team settings as member
      Given I am logged in as a member
      When I get team settings for my team
      Then the operation should succeed

    @p2 @requirement:UI-05-084
    Scenario: Add team member as manager
      Given I have a team
      When I add a member to the team
      Then the operation should succeed

    @p2 @requirement:UI-05-085
    Scenario: Remove team member as manager
      Given I have a team with multiple members
      When I remove a member from the team
      Then the operation should succeed

    @p2 @requirement:UI-05-086
    Scenario: List teams after creating
      Given I have created a team
      When I list all teams
      Then I should see at least 1 team

    @p2 @requirement:UI-05-087
    Scenario: List users and team settings together
      Given I have created a team
      When I list all users
      And I get team settings
      Then both operations should succeed

    @p2 @requirement:UI-05-088
    Scenario: All user and team operations
      Given I have created a team
      When I list all users
      And I get team settings
      And I update team name to "Final Name"
      Then all operations should succeed

    @p1 @requirement:UI-05-089
    Scenario: Get team details as manager
      Given I have created a team
      When I get team with ID 1
      Then the operation should succeed
      And I should see team information

    @p1 @requirement:UI-05-090
    Scenario: Get team settings as manager
      Given I have created a team
      When I get team settings for my team
      Then the operation should succeed
      And I should receive team settings

    @p2 @requirement:UI-05-091
    Scenario: Update team description
      Given I have created a team
      When I update team description to "Updated description"
      Then the operation should succeed

    @p1 @requirement:UI-05-092
    Scenario: Get usage statistics as manager
      When I get my usage statistics
      Then the operation should succeed
      And I should see usage data

    @p1 @requirement:UI-05-093
    Scenario: Get usage statistics as member
      Given I am logged in as a member
      When I get my usage statistics
      Then the operation should succeed

    @p1 @requirement:UI-05-094
    Scenario: List teams returns team data
      When I list all teams
      Then I should see at least 1 team
      And each team should have an ID

    @p1 @requirement:UI-05-095
    Scenario: List teams returns team names
      When I list all teams
      Then each team should have a name

    @p2 @requirement:UI-05-096
    Scenario: Get team settings returns team data
      When I get team settings
      Then the operation should succeed
      And I should see team information
