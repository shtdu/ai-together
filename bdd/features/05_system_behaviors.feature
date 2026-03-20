# BDD Test Suite - System Behaviors
# This feature covers health checks and system readiness
#
# Tag Format decisions:
# - Requirement ID: @requirement:ID (e.g., @requirement:IA-01-001)
# - Priority: @p0 / @p1 / @p2 / @p3 (e.g., @p0 for critical)
# - Work in Progress: @wip
# - License required: @license:commercial
#
# Priority Levels:
# - @p0 - Critical/Blocker: Must pass for any release (smoke tests)
# - @p1 - High: Core functionality
# - @p2 - Medium: Standard features
# - @p3 - Low: Edge cases, nice-to-have

Feature: System Behaviors
  As a system operator
  I want to monitor system health
  So that I can ensure reliable operation

  Background:
    Given the test server is running

  Rule: Health Checks

    @p0 @requirement:SB-05-001
    Scenario: Check system health endpoint
      When I check the health endpoint
      Then the system should be healthy
      And the response status code should be 200

    @p1 @requirement:SB-05-002
    Scenario: Check health returns system information
      When I check the health endpoint
      Then the response should contain version
      And the response should contain uptime
      And the response should contain database status

    @p1 @requirement:SB-05-003
    Scenario: Database status should be OK
      When I check the health endpoint
      Then database status should be OK
      And database connection should be active

    @p3 @requirement:SB-05-004
    Scenario: Health check without database
      Given the database is not available
      When I check the health endpoint
      Then the system should be unhealthy
      And database status should be "error"

  Rule: System Readiness

    @p0 @requirement:SB-05-005
    Scenario: System is ready when all services are up
      Given the test server is running
      And the database is connected
      When I check the readiness endpoint
      Then the system should be ready
      And the response status code should be 200

    @p3 @requirement:SB-05-006
    Scenario: System is not ready when database is down
      Given the test server is running
      And the database is not connected
      When I check the readiness endpoint
      Then the system should not be ready
      And the response status code should be 503
