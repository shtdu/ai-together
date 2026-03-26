# BDD Test Suite - Team Invitations
# This feature covers team invitation creation, acceptance, cancellation, and expiration

Feature: Team Invitations
  As a team manager
  I want to invite users to join my team
  So that we can collaborate effectively on AI configuration

  Background:
    Given the test server is running

  # ============================================================================
  # TEAM INVITATION
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
