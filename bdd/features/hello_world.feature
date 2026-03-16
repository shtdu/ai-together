# BDD Test Suite - Hello World Feature
# This is the first BDD scenario to validate the test infrastructure is working correctly

Feature: Hello World
  As a developer
  I want to verify the BDD test infrastructure is working
  So that I can proceed with implementing comprehensive BDD scenarios

  Scenario: First BDD test - Server health check
    Given the test server is running
    When I check the health endpoint
    Then I should receive a 200 status
    And the system should be healthy

  Scenario: Generate unique provider name
    Given the test server is running
    And I have a unique provider name "test-provider"
    Then the response should be successful
