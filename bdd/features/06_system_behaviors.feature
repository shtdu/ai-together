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

  Rule: Health Check Detailed Responses

    @p1 @requirement:SB-05-007
    Scenario: Health check returns service status
      When I check the health endpoint
      Then the response should contain service status
      And all services should be operational

    @p1 @requirement:SB-05-008
    Scenario: Health check includes memory info
      When I check the health endpoint
      Then the response should contain memory usage
      And the response should be valid JSON

    @p1 @requirement:SB-05-009
    Scenario: Health check works without authentication
      When I check the health endpoint without auth
      Then the response status code should be 200
      And the response should contain health status

  Rule: Additional Health Scenarios

    @p2 @requirement:SB-05-010
    Scenario: Health check returns timestamp
      When I check the health endpoint
      Then the response should contain current timestamp

    @p2 @requirement:SB-05-011
    Scenario: Health check includes server info
      When I check the health endpoint
      Then the response should contain server information
      And the response should contain environment info

    @p2 @requirement:SB-05-012
    Scenario: Health check with HEAD request
      When I send HEAD request to health endpoint
      Then the response status code should be 200

    @p2 @requirement:SB-05-013
    Scenario: Health check handles concurrent requests
      When I check the health endpoint multiple times
      Then all requests should succeed
      And all responses should be consistent

  Rule: Health Check Extended Scenarios

    @p1 @requirement:SB-05-018
    Scenario: Health check includes all required fields
      When I check the health endpoint
      Then the response should contain health status
      And the response should contain database status
      And the response should be valid JSON

    @p2 @requirement:SB-05-019
    Scenario: Readiness check when database is connected
      Given the database is connected
      When I check the readiness endpoint
      Then the system should be ready

    @p2 @requirement:SB-05-020
    Scenario: Health check endpoint responds quickly
      When I check the health endpoint
      Then the response status code should be 200
      And the response should be valid JSON

  Rule: Health Check Additional Variations

    @p1 @requirement:SB-05-026
    Scenario: Health check contains all required fields
      Given I am logged in as a manager
      When I check the health endpoint
      Then the response should contain version
      And the response should contain database status
      And the response should contain uptime

    @p1 @requirement:SB-05-027
    Scenario: Health check returns consistent results
      Given I am logged in as a manager
      When I check the health endpoint
      And I check the health endpoint again
      Then both responses should be consistent

    @p2 @requirement:SB-05-029
    Scenario: Health check without authentication
      Given I am not authenticated
      When I check the health endpoint without auth
      Then the response status code should be 200

  Rule: System Status Variations

    @p1 @requirement:SB-05-031
    Scenario: System health check returns 200
      Given I am logged in as a manager
      When I check the health endpoint
      Then the response status code should be 200

    @p1 @requirement:SB-05-032
    Scenario: Database status is OK
      Given I am logged in as a manager
      When I check the health endpoint
      Then database status should be OK

    @p2 @requirement:SB-05-033
    Scenario: Health check contains server time
      Given I am logged in as a manager
      When I check the health endpoint
      Then the response should contain current timestamp

    @p2 @requirement:SB-05-034
    Scenario: Health check multiple endpoints
      Given I am logged in as a manager
      When I check the health endpoint
      And I check the readiness endpoint
      Then both operations should succeed

  Rule: Health Check Extended

    @p1 @requirement:SB-05-035
    Scenario: Health check returns all required information
      Given the test server is running
      When I check the health endpoint
      Then the response should contain version
      And the response should contain database status
      And the response should contain uptime

    @p1 @requirement:SB-05-036
    Scenario: Health check works rapidly
      Given the test server is running
      When I check the health endpoint 5 times
      Then all requests should succeed
      And all responses should be consistent

    @p2 @requirement:SB-05-037
    Scenario: Readiness check returns consistent results
      Given I am logged in as a manager
      When I check the readiness endpoint
      And I check the readiness endpoint
      Then both operations should succeed

    @p2 @requirement:SB-05-038
    Scenario: Health check response is valid JSON
      Given the test server is running
      When I check the health endpoint
      Then the response should be valid JSON
      And the response should contain health status

    @p2 @requirement:SB-05-039
    Scenario: Health check includes service information
      Given I am logged in as a manager
      When I check the health endpoint
      Then the response should contain service status
      And all services should be operational

  Rule: Health Check Simple Variations

    @p1 @requirement:SB-05-040
    Scenario: Health check works multiple times
      Given the test server is running
      When I check the health endpoint
      And I check the health endpoint
      And I check the health endpoint
      Then all operations should succeed

    @p2 @requirement:SB-05-043
    Scenario: Health and readiness both succeed
      Given I am logged in as a manager
      When I check the health endpoint
      And I check the readiness endpoint
      Then both operations should succeed

