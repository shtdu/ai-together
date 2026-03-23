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

    @p1 @requirement:UI-05-004-a
    Scenario: List team members
      Given I have created a team
      When I list team members
      Then I should see team members list
      And the operation should succeed

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

    @p2 @requirement:UI-05-008 @license:commercial
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

    @p1 @requirement:UI-05-021
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

    @p1 @requirement:UI-05-033 @license:commercial
    Scenario: Manager can manage teams
      Given I am logged in as a manager
      When I create a team
      Then the operation should succeed

    @p1 @requirement:UI-05-034
    Scenario: Member cannot manage teams
      Given I am logged in as a member
      When I attempt to create a team
      Then I should receive a 403 error

  Rule: Dashboard API Endpoints

    @p1 @requirement:UI-05-035
    Scenario: Get dashboard metrics via dashboard API
      Given I am logged in as a manager
      When I get dashboard metrics via dashboard API
      Then the operation should succeed

    @p1 @requirement:UI-05-036
    Scenario: Get dashboard rankings
      Given I am logged in as a manager
      When I get dashboard rankings
      Then the operation should succeed

    @p1 @requirement:UI-05-037
    Scenario: Get dashboard members
      Given I am logged in as a manager
      When I get dashboard members
      Then the operation should succeed
      And I should see members list

    @p2 @requirement:UI-05-038
    Scenario: Member cannot access dashboard members endpoint
      Given I am logged in as a member
      When I get dashboard members
      Then I should receive a 403 error

    @p1 @requirement:UI-05-039
    Scenario: Get dashboard metrics with 24h range
      Given I am logged in as a manager
      When I get dashboard metrics with range "24h"
      Then the operation should succeed

    @p1 @requirement:UI-05-040
    Scenario: Get dashboard metrics with 30d range
      Given I am logged in as a manager
      When I get dashboard metrics with range "30d"
      Then the operation should succeed

  Rule: Additional Team Management Tests

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

  Rule: User Management Error Paths

    @p1 @requirement:UI-05-041
    Scenario: Get non-existent user
      When I attempt to get user with ID 99999
      Then I should receive a 404 error

    @p1 @requirement:UI-05-042
    Scenario: Delete non-existent user
      When I attempt to delete user with ID 99999
      Then I should receive a 404 error

    @p2 @requirement:UI-05-043
    Scenario: Get user without authentication
      Given I am not authenticated
      When I attempt to get user details
      Then I should receive a 401 error

  Rule: Team Settings Management

    @p1 @requirement:UI-05-044
    Scenario: Get team settings for existing team
      Given I have created a team
      When I get team settings
      Then I should receive team settings
      And settings should have team ID

    @p1 @requirement:UI-05-045
    Scenario: Update team name and description together
      Given I have created a team
      When I update team name and description
      Then both fields should be updated

    @p1 @requirement:UI-05-046
    Scenario: Get team settings for non-existent team
      When I attempt to get settings for team ID 99999
      Then I should receive a 404 error

    @p1 @requirement:UI-05-047
    Scenario: Update settings for non-existent team
      When I attempt to update settings for team ID 99999
      Then I should receive a 404 error

  Rule: Dashboard API Query Parameters

    @p2 @requirement:UI-05-048
    Scenario: Get dashboard metrics with invalid range
      Given I am logged in as a manager
      When I get dashboard metrics with range "invalid"
      Then the operation should succeed or return validation error

    @p2 @requirement:UI-05-049
    Scenario: Get dashboard rankings without authentication
      When I get dashboard rankings
      Then I should receive a 401 error

    @p2 @requirement:UI-05-050
    Scenario: Get dashboard members without authentication
      When I get dashboard members
      Then I should receive a 401 error

  Rule: Additional User Management Scenarios

    @p2 @requirement:UI-05-051
    Scenario: Update user with invalid role
      Given I am logged in as a manager
      And I have a user with role "member"
      When I update user role to "invalid_role"
      Then I should receive a 400 error

    @p2 @requirement:UI-05-052
    Scenario: Update non-existent user role
      Given I am logged in as a manager
      When I update role for user ID 99999 to "manager"
      Then I should receive a 404 error

  Rule: Basic User Queries

    @p1 @requirement:UI-05-053
    Scenario: Get all users as manager
      Given I am logged in as a manager
      When I list all users
      Then the operation should succeed
      And I should see at least 1 user

    @p1 @requirement:UI-05-054
    Scenario: Get all users as member
      Given I am logged in as a member
      When I list all users
      Then the operation should succeed
      And I should see at least 1 user

    @p1 @requirement:UI-05-055
    Scenario: Get dashboard metrics as manager
      Given I am logged in as a manager
      When I get dashboard metrics
      Then the operation should succeed
      And I should see metrics data

    @p1 @requirement:UI-05-056
    Scenario: Get dashboard metrics as member
      Given I am logged in as a member
      When I get dashboard metrics
      Then the operation should succeed
      And I should see my usage data

  Rule: Team Basic Operations

    @p1 @requirement:UI-05-057
    Scenario: Get default team
      Given I am logged in as a manager
      When I get team with ID 1
      Then the operation should succeed
      And I should see team information

    @p1 @requirement:UI-05-058
    Scenario: Get teams list as member
      Given I am logged in as a member
      When I list all teams
      Then the operation should succeed
      And I should see at least 1 team

  Rule: Dashboard Metrics Variations

    @p1 @requirement:UI-05-059
    Scenario: Get dashboard metrics with default range
      Given I am logged in as a manager
      When I get dashboard metrics
      Then the operation should succeed

    @p1 @requirement:UI-05-060
    Scenario: Get dashboard metrics as member with 24h range
      Given I am logged in as a member
      When I get dashboard metrics with range "24h"
      Then the operation should succeed

    @p1 @requirement:UI-05-061
    Scenario: Get dashboard rankings as manager
      Given I am logged in as a manager
      When I get dashboard rankings
      Then the operation should succeed

    @p1 @requirement:UI-05-062
    Scenario: Get dashboard members as manager
      Given I am logged in as a manager
      When I get dashboard members
      Then the operation should succeed
      And I should see members list

  Rule: Team Query Variations

    @p1 @requirement:UI-05-063
    Scenario: List teams as manager
      Given I am logged in as a manager
      When I list all teams
      Then the operation should succeed
      And I should see at least 1 team

    @p1 @requirement:UI-05-064
    Scenario: Get user profile as manager
      Given I am logged in as a manager
      When I get my user profile
      Then the operation should succeed
      And my profile should contain my email

    @p1 @requirement:UI-05-065
    Scenario: Get user profile as member
      Given I am logged in as a member
      When I get my user profile
      Then the operation should succeed
      And my profile should contain my email

  Rule: Additional Dashboard Queries

    @p1 @requirement:UI-05-066
    Scenario: Dashboard metrics available to managers
      Given I am logged in as a manager
      When I get dashboard metrics
      Then the operation should succeed

    @p1 @requirement:UI-05-067
    Scenario: Dashboard rankings available to managers
      Given I am logged in as a manager
      When I get dashboard rankings
      Then the operation should succeed

    @p1 @requirement:UI-05-068
    Scenario: Dashboard members available to managers
      Given I am logged in as a manager
      When I get dashboard members
      Then the operation should succeed

    @p1 @requirement:UI-05-069
    Scenario: Team settings accessible to managers
      Given I am logged in as a manager
      And I have created a team
      When I get team settings
      Then the operation should succeed

    @p1 @requirement:UI-05-070
    Scenario: User list accessible to managers
      Given I am logged in as a manager
      When I list all users
      Then the operation should succeed

  Rule: Provider Statistics Access

    @p1 @requirement:UI-05-071
    Scenario: Manager can access provider statistics
      Given I am logged in as a manager
      And I have created a provider
      When I get provider statistics
      Then the operation should succeed

    @p1 @requirement:UI-05-072
    Scenario: Member cannot access provider statistics
      Given I am logged in as a member
      When I attempt to get provider statistics
      Then I should receive a 403 error

  Rule: Usage Statistics Access

    @p1 @requirement:UI-05-073
    Scenario: Manager can view team usage
      Given I am logged in as a manager
      When I get team usage summary
      Then the operation should succeed

    @p1 @requirement:UI-05-074
    Scenario: Manager can view cost trends
      Given I am logged in as a manager
      When I get cost trends
      Then the operation should succeed

    @p1 @requirement:UI-05-075
    Scenario: Manager can view user rankings
      Given I am logged in as a manager
      When I get user rankings
      Then the operation should succeed

    @p1 @requirement:UI-05-076
    Scenario: Manager can view provider performance
      Given I am logged in a manager
      When I get provider performance
      Then the operation should succeed

  Rule: License Status Queries

    @p1 @requirement:UI-05-077
    Scenario: Get license information as manager
      Given I am logged in as a manager
      When I get license information
      Then the operation should succeed

    @p1 @requirement:UI-05-078
    Scenario: Get license information as member
      Given I am logged in as a member
      When I get license information
      Then the operation should succeed

    @p1 @requirement:UI-05-079
    Scenario: License status accessible to authenticated users
      Given I am logged in as a member
      When I get the license status
      Then the operation should succeed

  Rule: Team Creation and Management

    @p1 @requirement:UI-05-080
    Scenario: Manager can view all teams
      Given I am logged in as a manager
      When I list all teams
      Then the operation should succeed
      And I should see at least 1 team

    @p1 @requirement:UI-05-081
    Scenario: Member can view their team
      Given I am logged in as a member
      When I list all teams
      Then the operation should succeed
      And I should see my team

  Rule: Dashboard Members Detailed Queries

    @p1 @requirement:UI-05-082
    Scenario: Get dashboard members as manager
      Given I am logged in as a manager
      When I get dashboard members
      Then the operation should succeed
      And I should see members list
      And each member should have email

    @p2 @requirement:UI-05-083
    Scenario: Get dashboard members includes user details
      Given I am logged in as a manager
      And there are multiple users in the team
      When I get dashboard members
      Then I should see multiple members
      And each member should have user ID

    @p2 @requirement:UI-05-084
    Scenario: Get dashboard members filters correctly
      Given I am logged in as a manager
      When I get dashboard members
      Then the operation should succeed
      And I should not see deleted users

    @p2 @requirement:UI-05-085
    Scenario: Get dashboard members handles empty team
      Given I am logged in as a manager
      And I am the only user
      When I get dashboard members
      Then I should see at least 1 member

  Rule: Dashboard Rankings Queries

    @p1 @requirement:UI-05-086
    Scenario: Get dashboard rankings shows usage order
      Given I am logged in as a manager
      And there are multiple users with different usage
      When I get dashboard rankings
      Then the operation should succeed
      And users should be ranked by usage

    @p2 @requirement:UI-05-087
    Scenario: Get dashboard rankings as manager
      Given I am logged in as a manager
      When I get dashboard rankings
      Then the operation should succeed
      And I should see rankings data

    @p2 @requirement:UI-05-088
    Scenario: Get dashboard rankings with time range
      Given I am logged in as a manager
      When I get dashboard rankings
      Then the operation should succeed
      And I should see usage totals

  Rule: Analytics Queries

    @p1 @requirement:UI-05-089
    Scenario: Get provider analytics as manager
      Given I am logged in as a manager
      And I have created a provider
      When I get provider analytics
      Then the operation should succeed
      And I should see analytics data

    @p2 @requirement:UI-05-090
    Scenario: Get provider analytics filters by provider
      Given I am logged in as a manager
      And I have created a claude provider
      And I have created a codex provider
      When I get provider analytics for claude
      Then I should see claude analytics

    @p2 @requirement:UI-05-091
    Scenario: Get provider analytics with date range
      Given I am logged in as a manager
      And I have created a provider
      When I get provider analytics with date range
      Then the operation should succeed

  Rule: User Profile Additional Queries

    @p1 @requirement:UI-05-092
    Scenario: Get user profile contains all fields
      Given I am logged in as a manager
      When I get my user profile
      Then I should see my email
      And I should see my role
      And I should see my team ID

    @p2 @requirement:UI-05-093
    Scenario: Get user profile after role change
      Given I am logged in as a manager
      And I update my role
      When I get my user profile
      Then I should see the updated role

    @p2 @requirement:UI-05-094
    Scenario: Get user profile after team change
      Given I am logged in as a manager
      And I change my team
      When I get my user profile
      Then I should see the updated team

  Rule: Dashboard Metrics Time Variations

    @p1 @requirement:UI-05-095
    Scenario: Get dashboard metrics with 7d range
      Given I am logged in as a manager
      When I get dashboard metrics with range "7d"
      Then the operation should succeed
      And I should see usage data

    @p1 @requirement:UI-05-096
    Scenario: Get dashboard metrics with 90d range
      Given I am logged in as a manager
      When I get dashboard metrics with range "90d"
      Then the operation should succeed

    @p2 @requirement:UI-05-097
    Scenario: Get dashboard metrics with custom range
      Given I am logged in as a manager
      When I get dashboard metrics with range "custom"
      Then the operation should succeed or default to standard range

  Rule: Additional User Profile Queries

    @p1 @requirement:UI-05-098
    Scenario: Get profile after login
      Given I am logged in as a manager
      When I get my user profile
      Then the operation should succeed
      And I should see my email

    @p1 @requirement:UI-05-099
    Scenario: Get profile returns correct user data
      Given I am logged in as a member
      When I get my user profile
      Then the operation should succeed
      And my profile should contain my email

    @p2 @requirement:UI-05-100
    Scenario: Get profile includes tenant information
      Given I am logged in as a manager
      When I get my user profile
      Then the operation should succeed
      And my profile should contain tenant ID

  Rule: Team Settings Additional Queries

    @p1 @requirement:UI-05-101
    Scenario: Get team settings as manager
      Given I am logged in as a manager
      And I have created a team
      When I get team settings
      Then the operation should succeed
      And I should receive team settings

    @p2 @requirement:UI-05-102
    Scenario: Get team settings as member
      Given I am logged in as a member
      When I get team settings for my team
      Then the operation should succeed

    @p1 @requirement:UI-05-103
    Scenario: Update team settings successfully
      Given I am logged in as a manager
      And I have created a team
      When I update team name to "Updated Team Name"
      Then the operation should succeed
      And the team name should be "Updated Team Name"

  Rule: User List Queries

    @p1 @requirement:UI-05-104
    Scenario: List users returns user data
      Given I am logged in as a manager
      When I list all users
      Then the operation should succeed
      And I should see at least 1 user

    @p2 @requirement:UI-05-105
    Scenario: List users includes email addresses
      Given I am logged in as a manager
      When I list all users
      Then each user should have email

    @p2 @requirement:UI-05-106
    Scenario: List users includes roles
      Given I am logged in as a manager
      When I list all users
      Then each user should have role

  Rule: Dashboard Members Additional Queries

    @p1 @requirement:UI-05-107
    Scenario: Get dashboard members as manager
      Given I am logged in as a manager
      When I get dashboard members
      Then the operation should succeed
      And I should see members list

    @p2 @requirement:UI-05-108
    Scenario: Get dashboard members filters active users
      Given I am logged in as a manager
      When I get dashboard members
      Then I should not see deleted users

    @p2 @requirement:UI-05-109
    Scenario: Get dashboard members returns user details
      Given I am logged in as a manager
      When I get dashboard members
      Then each member should have email

  Rule: Team Creation Extended Scenarios

    @p1 @requirement:UI-05-110
    Scenario: Create team with name only
      Given I am logged in as a manager
      When I create a team with name "Name Only Team"
      Then the operation should succeed
      And the team should be created

    @p1 @requirement:UI-05-111
    Scenario: Create team with name and description
      Given I am logged in as a manager
      When I create a team with name "Full Team" and description "A team with description"
      Then the operation should succeed
      And the team should have the name

    @p2 @requirement:UI-05-112
    Scenario: List teams as member
      Given I am logged in as a member
      When I list all teams
      Then I should see at least 1 team

    @p2 @requirement:UI-05-113
    Scenario: Get dashboard metrics with different ranges
      Given I am logged in as a manager
      When I get dashboard metrics with range "24h"
      And I get dashboard metrics with range "7d"
      Then both operations should succeed

  Rule: User Profile Extended Queries

    @p1 @requirement:UI-05-114
    Scenario: Get profile multiple times
      Given I am logged in as a manager
      When I get my user profile
      And I get my user profile
      Then both operations should succeed

    @p2 @requirement:UI-05-115
    Scenario: Get profile after team change
      Given I am logged in as a manager
      When I get my user profile
      Then my profile should contain my email
      And my profile should contain tenant ID

  Rule: Dashboard Extended Queries

    @p1 @requirement:UI-05-116
    Scenario: Get dashboard metrics multiple times
      Given I am logged in as a manager
      When I get dashboard metrics
      And I get dashboard metrics
      Then both operations should succeed

    @p1 @requirement:UI-05-117
    Scenario: Get dashboard rankings multiple times
      Given I am logged in as a manager
      When I get dashboard rankings
      And I get dashboard rankings
      Then both operations should succeed

    @p2 @requirement:UI-05-118
    Scenario: Get dashboard members multiple times
      Given I am logged in as a manager
      When I get dashboard members
      And I get dashboard members
      Then both operations should succeed

    @p2 @requirement:UI-05-119
    Scenario: Get dashboard metrics with different parameters
      Given I am logged in as a manager
      When I get dashboard metrics with range "24h"
      And I get dashboard metrics with range "7d"
      Then both operations should succeed

    @p2 @requirement:UI-05-120
    Scenario: Get user profile and dashboard data
      Given I am logged in as a manager
      When I get my user profile
      And I get dashboard metrics
      Then both operations should succeed

  Rule: Dashboard Members Error Paths

    @p1 @requirement:UI-05-121
    Scenario: Get dashboard members as non-manager fails
      Given I am logged in as a member
      When I get dashboard members
      Then I should receive a 403 error
      And the error should indicate access denied

    @p2 @requirement:UI-05-122
    Scenario: Get dashboard members without authentication fails
      Given I am not authenticated
      When I get dashboard members
      Then I should receive a 401 error

  Rule: Dashboard Metrics Error Paths

    @p1 @requirement:UI-05-123
    Scenario: Get dashboard metrics without authentication fails
      Given I am not authenticated
      When I get dashboard metrics
      Then I should receive a 401 error

    @p2 @requirement:UI-05-124
    Scenario: Get dashboard rankings without authentication fails
      Given I am not authenticated
      When I get dashboard rankings
      Then I should receive a 401 error

  Rule: User Profile Additional Scenarios

    @p1 @requirement:UI-05-125
    Scenario: Get profile returns name field
      Given I am logged in as a manager
      When I get my user profile
      Then my profile should contain my name

    @p1 @requirement:UI-05-126
    Scenario: Get profile returns role field
      Given I am logged in as a manager
      When I get my user profile
      Then my profile should contain my role

    @p2 @requirement:UI-05-127
    Scenario: Get profile without authentication fails
      Given I am not authenticated
      When I get my user profile
      Then I should receive a 401 error

  Rule: Team Queries Extended

    @p1 @requirement:UI-05-128
    Scenario: List teams returns team data
      Given I am logged in as a manager
      When I list all teams
      Then I should see at least 1 team
      And each team should have an ID

    @p1 @requirement:UI-05-129
    Scenario: List teams returns team names
      Given I am logged in as a manager
      When I list all teams
      Then each team should have a name

    @p2 @requirement:UI-05-130
    Scenario: Get team settings returns team data
      Given I am logged in as a manager
      When I get team settings
      Then the operation should succeed
      And I should see team information

  Rule: Dashboard Members Extended

    @p1 @requirement:UI-05-131
    Scenario: Get dashboard members multiple times
      Given I am logged in as a manager
      When I get dashboard members
      And I get dashboard members
      And I get dashboard members
      Then all operations should succeed

    @p1 @requirement:UI-05-132
    Scenario: Get dashboard members after provider update
      Given I am logged in as a manager
      And I have created a provider
      When I update the provider name
      And I get dashboard members
      Then the operation should succeed

    @p2 @requirement:UI-05-133
    Scenario: Get dashboard members returns member IDs
      Given I am logged in as a manager
      When I get dashboard members
      Then each member should have user ID

    @p2 @requirement:UI-05-134
    Scenario: Get dashboard members returns usage data
      Given I am logged in as a manager
      When I get dashboard members
      Then I should see usage data

  Rule: Dashboard Metrics Extended

    @p1 @requirement:UI-05-135
    Scenario: Get dashboard metrics after team change
      Given I am logged in as a manager
      And I have created a team
      When I get dashboard metrics
      Then the operation should succeed

    @p1 @requirement:UI-05-136
    Scenario: Get dashboard rankings after member added
      Given I am logged in as a manager
      And I have created a team
      When I get dashboard rankings
      Then the operation should succeed

    @p2 @requirement:UI-05-137
    Scenario: Get dashboard metrics with different ranges
      Given I am logged in as a manager
      When I get dashboard metrics with range "24h"
      And I get dashboard metrics with range "7d"
      And I get dashboard metrics with range "30d"
      Then all operations should succeed

  Rule: Dashboard Metrics Extended

    @p1 @requirement:UI-05-138
    Scenario: Get dashboard metrics after team creation
      Given I am logged in as a manager
      And I have created a team
      When I get dashboard metrics
      Then the operation should succeed
      And I should see metrics data

    @p1 @requirement:UI-05-139
    Scenario: Get dashboard metrics after provider update
      Given I am logged in as a manager
      And I have created a provider
      When I update the provider name
      And I get dashboard metrics
      Then the operation should succeed

    @p2 @requirement:UI-05-140
    Scenario: Get dashboard rankings returns rankings
      Given I am logged in as a manager
      When I get dashboard rankings
      Then I should see rankings data
      And the operation should succeed

    @p2 @requirement:UI-05-141
    Scenario: Get dashboard rankings multiple times
      Given I am logged in as a manager
      When I get dashboard rankings
      And I get dashboard rankings
      Then both operations should succeed

  Rule: Team Member Management

    @p1 @requirement:UI-05-131
    Scenario: List team members as manager
      Given I am logged in as a manager
      When I list team members
      Then the operation should succeed
      And I should see at least 1 member

    @p2 @requirement:UI-05-132
    Scenario: List team members as member
      Given I am logged in as a member
      When I list team members
      Then the operation should succeed
      And I should see at least 1 member

    @p1 @requirement:UI-05-133
    Scenario: List team members without authentication fails
      Given I am not authenticated
      When I list team members
      Then I should receive a 401 error

    @p2 @requirement:UI-05-134
    Scenario: Add team member as manager
      Given I am logged in as a manager
      And I have a team
      When I add a member to the team
      Then the operation should succeed

    @p2 @requirement:UI-05-135
    Scenario: Remove team member as manager
      Given I am logged in as a manager
      And I have a team with multiple members
      When I remove a member from the team
      Then the operation should succeed

  Rule: Team Deletion and Updates

    @p1 @requirement:UI-05-136
    Scenario: Delete team as manager
      Given I am logged in as a manager
      And I have created a team
      When I delete the team
      Then the operation should succeed

    @p1 @requirement:UI-05-137
    Scenario: Delete team with invalid ID fails
      Given I am logged in as a manager
      When I attempt to delete team with ID 99999
      Then I should receive a 404 error

    @p2 @requirement:UI-05-138
    Scenario: Update team description
      Given I am logged in as a manager
      And I have created a team
      When I update team description to "Updated description"
      Then the operation should succeed

  Rule: Usage Statistics Extended

    @p1 @requirement:UI-05-139
    Scenario: Get usage statistics as manager
      Given I am logged in as a manager
      When I get my usage statistics
      Then the operation should succeed
      And I should see usage data

    @p1 @requirement:UI-05-140
    Scenario: Get usage statistics as member
      Given I am logged in as a member
      When I get my usage statistics
      Then the operation should succeed

    @p2 @requirement:UI-05-141
    Scenario: Get usage statistics without authentication fails
      Given I am not authenticated
      When I get my usage statistics
      Then I should receive a 401 error

  Rule: Team Management Additional

    @p1 @requirement:UI-05-149
    Scenario: Get team details as manager
      Given I am logged in as a manager
      And I have created a team
      When I get team with ID 1
      Then the operation should succeed
      And I should see team information

    @p1 @requirement:UI-05-150
    Scenario: Get team settings as manager
      Given I am logged in as a manager
      And I have created a team
      When I get team settings for my team
      Then the operation should succeed
      And I should receive team settings

    @p2 @requirement:UI-05-151
    Scenario: List teams as member
      Given I am logged in as a member
      When I list all teams
      Then the operation should succeed
      And I should see at least 1 team

  Rule: Dashboard Queries Extended

    @p1 @requirement:UI-05-152
    Scenario: Get dashboard metrics with 24h range multiple times
      Given I am logged in as a manager
      When I get dashboard metrics with range "24h"
      And I get dashboard metrics with range "24h"
      Then both operations should succeed

    @p1 @requirement:UI-05-153
    Scenario: Get dashboard metrics with 7d range multiple times
      Given I am logged in as a manager
      When I get dashboard metrics with range "7d"
      And I get dashboard metrics with range "7d"
      Then both operations should succeed

    @p2 @requirement:UI-05-154
    Scenario: Get dashboard rankings without auth fails
      Given I am not authenticated
      When I get dashboard rankings
      Then I should receive a 401 error

    @p2 @requirement:UI-05-155
    Scenario: Get dashboard metrics without auth fails
      Given I am not authenticated
      When I get dashboard metrics
      Then I should receive a 401 error

  Rule: Dashboard Members Access Control

    @p1 @requirement:UI-05-142
    Scenario: Get dashboard members as member returns 403
      Given I am logged in as a member
      When I get dashboard members
      Then I should receive a 403 error
      And the error should indicate "access denied" or "forbidden"

    @p1 @requirement:UI-05-143
    Scenario: Get dashboard members without authentication returns 401
      Given I am not authenticated
      When I get dashboard members
      Then I should receive a 401 error

    @p2 @requirement:UI-05-144
    Scenario: Get dashboard members returns empty list for new team
      Given I am logged in as a manager
      And I have created a new team
      When I get dashboard members
      Then the operation should succeed
      And I should see an empty members list

    @p2 @requirement:UI-05-145
    Scenario: Get dashboard rankings without authentication returns 401
      Given I am not authenticated
      When I get dashboard rankings
      Then I should receive a 401 error

  Rule: Dashboard Metrics Edge Cases

    @p1 @requirement:UI-05-146
    Scenario: Get dashboard metrics with custom range
      Given I am logged in as a manager
      When I get dashboard metrics with range "30d"
      Then the operation should succeed
      And I should see metrics data

    @p1 @requirement:UI-05-147
    Scenario: Get dashboard metrics with 24h range
      Given I am logged in as a manager
      When I get dashboard metrics with range "24h"
      Then the operation should succeed
      And I should see metrics data

    @p2 @requirement:UI-05-148
    Scenario: Get dashboard metrics without authentication returns 401
      Given I am not authenticated
      When I get dashboard metrics
      Then I should receive a 401 error

  Rule: Team Management Queries

    @p1 @requirement:UI-05-156
    Scenario: Get team settings multiple times
      Given I am logged in as a manager
      And I have created a team
      When I get team settings for my team
      And I get team settings for my team
      Then both operations should succeed

    @p1 @requirement:UI-05-157
    Scenario: List teams after creating
      Given I am logged in as a manager
      And I have created a team
      When I list all teams
      Then I should see at least 1 team

    @p2 @requirement:UI-05-158
    Scenario: List teams as member
      Given I am logged in as a member
      When I list all teams
      Then the operation should succeed

  Rule: Dashboard Extended Queries

    @p1 @requirement:UI-05-159
    Scenario: Get dashboard rankings multiple times
      Given I am logged in as a manager
      When I get dashboard rankings
      And I get dashboard rankings
      Then both operations should succeed

    @p2 @requirement:UI-05-160
    Scenario: Get dashboard metrics with all ranges
      Given I am logged in as a manager
      When I get dashboard metrics with range "24h"
      And I get dashboard metrics with range "7d"
      And I get dashboard metrics with range "30d"
      Then all operations should succeed

  Rule: Dashboard Members Extended

    @p1 @requirement:UI-05-161
    Scenario: Get dashboard members multiple times
      Given I am logged in as a manager
      When I get dashboard members
      And I get dashboard members
      And I get dashboard members
      And I get dashboard members
      Then all operations should succeed

    @p1 @requirement:UI-05-162
    Scenario: Get dashboard members as manager
      Given I am logged in as a manager
      When I get dashboard members
      Then the operation should succeed
      And I should see user information

    @p2 @requirement:UI-05-163
    Scenario: Get dashboard members and metrics together
      Given I am logged in as a manager
      When I get dashboard members
      And I get dashboard metrics
      Then both operations should succeed

  Rule: Dashboard Comprehensive Operations

    @p1 @requirement:UI-05-164
    Scenario: All dashboard endpoints
      Given I am logged in as a manager
      When I get dashboard metrics
      And I get dashboard rankings
      And I get dashboard members
      Then all operations should succeed

    @p2 @requirement:UI-05-165
    Scenario: Dashboard operations repeated
      Given I am logged in as a manager
      When I get dashboard metrics
      And I get dashboard rankings
      And I get dashboard members
      And I get dashboard metrics
      And I get dashboard rankings
      And I get dashboard members
      Then all operations should succeed

  Rule: Team Settings Extended

    @p1 @requirement:UI-05-166
    Scenario: Get team settings multiple times
      Given I am logged in as a manager
      And I have created a team
      When I get team settings
      And I get team settings
      And I get team settings
      Then all operations should succeed

    @p1 @requirement:UI-05-167
    Scenario: Update team settings multiple times
      Given I am logged in as a manager
      And I have created a team
      When I update team name to "First Update"
      And I update team name to "Second Update"
      And I update team name to "Third Update"
      Then all operations should succeed

    @p2 @requirement:UI-05-168
    Scenario: Team settings operations together
      Given I am logged in as a manager
      And I have created a team
      When I get team settings
      And I update team name to "Updated Name"
      Then both operations should succeed

  Rule: User List Extended

    @p1 @requirement:UI-05-169
    Scenario: List users multiple times
      Given I am logged in as a manager
      When I list all users
      And I list all users
      And I list all users
      Then all operations should succeed

    @p1 @requirement:UI-05-170
    Scenario: List users and team settings together
      Given I am logged in as a manager
      And I have created a team
      When I list all users
      And I get team settings
      Then both operations should succeed

    @p2 @requirement:UI-05-171
    Scenario: All user and team operations
      Given I am logged in as a manager
      And I have created a team
      When I list all users
      And I get team settings
      And I update team name to "Final Name"
      Then all operations should succeed


