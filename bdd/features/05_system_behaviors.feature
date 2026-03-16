# BDD Test Suite - System Behaviors
# This feature covers health checks and system readiness

Feature: System Behaviors
  As a system operator
  I want to monitor system health
  So that I can ensure reliable operation

  Background:
    Given the test server is running

  Rule: Health Checks

    Scenario: Check system health endpoint
      When I check the health endpoint
      Then the system should be healthy
      And the response status code should be 200

    Scenario: Check health returns system information
      When I check the health endpoint
      Then the response should contain version
      And the response should contain uptime
      And the response should contain database status

    Scenario: Database status should be OK
      When I check the health endpoint
      Then database status should be OK
      And database connection should be active

    Scenario: Health check without database
      Given the database is not available
      When I check the health endpoint
      Then the system should be unhealthy
      And database status should be "error"

  Rule: System Readiness

    Scenario: System is ready when all services are up
      Given the test server is running
      And the database is connected
      When I check the readiness endpoint
      Then the system should be ready
      And the response status code should be 200

    Scenario: System is not ready when database is down
      Given the test server is running
      And the database is not connected
      When I check the readiness endpoint
      Then the system should not be ready
      And the response status code should be 503
