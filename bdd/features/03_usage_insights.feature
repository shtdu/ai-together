# BDD Test Suite - Usage Insights and Analytics
# This feature covers usage data upload, aggregation, filtering, cost calculations, and analytics

Feature: Usage Insights and Analytics
  As a team member
  I want to track and analyze AI usage
  So that we can understand costs and optimize resource usage

  Background:
    Given the test server is running
    And I am logged in as a manager

  Rule: Usage Data Upload

    Scenario: Upload single usage record successfully
      Given I have a usage record with 1000 tokens
      When I upload the usage record
      Then the record should be stored
      And the record should have an ID

    Scenario: Upload batch of usage records
      Given I have 10 usage records
      When I upload the usage records as a batch
      Then all records should be stored
      And I should receive a 200 status

    Scenario Outline: Upload usage record with different token counts
      Given I have a usage record with <tokens> tokens
      When I upload the usage record
      Then the record should be stored
      And the token count should be <tokens>

      Examples:
        | tokens |
        | 100 |
        | 1000 |
        | 10000 |
        | 100000 |

    Scenario: Upload usage record with provider information
      Given I have a usage record from provider "claude"
      And the record has 1000 tokens
      When I upload the usage record
      Then the record should be stored
      And the provider should be recorded

    Scenario: Upload usage record with user information
      Given I have a usage record for user "user@example.com"
      And the record has 500 tokens
      When I upload the usage record
      Then the record should be stored
      And the user should be recorded

    Scenario: Upload usage record with timestamp
      Given I have a usage record with timestamp "2025-03-17T10:00:00Z"
      When I upload the usage record
      Then the record should be stored
      And the timestamp should be recorded

    Scenario: Upload usage record with model information
      Given I have a usage record for model "claude-3-opus"
      When I upload the usage record
      Then the record should be stored
      And the model should be recorded

    Scenario: Reject usage record with invalid token count
      Given I have a usage record with -1 tokens
      When I upload the usage record
      Then I should receive a 400 error
      And the error message should contain "invalid token count"

    Scenario: Reject usage record without required fields
      Given I have a usage record without provider
      When I upload the usage record
      Then I should receive a 400 error

    Scenario: Upload usage record without authentication
      Given I am not authenticated
      And I have a usage record with 1000 tokens
      When I upload the usage record
      Then I should receive a 401 error

    Scenario: Upload usage record with cost information
      Given I have a usage record with 1000 tokens and cost $0.15
      When I upload the usage record
      Then the record should be stored
      And the cost should be recorded

    Scenario: Upload usage record with metadata
      Given I have a usage record with metadata
      And the metadata contains "project": "ai-assistant"
      When I upload the usage record
      Then the record should be stored
      And the metadata should be recorded

  Rule: Usage Aggregation

    @wip
    Scenario: Get daily usage statistics
      Given I have uploaded usage for the past 7 days
      When I get daily usage statistics
      Then I should see aggregated data for each day
      And each day should have total tokens
      And each day should have total cost

    Scenario: Aggregate usage by provider
      Given I have uploaded usage from multiple providers
      When I get usage aggregated by provider
      Then I should see data for each provider
      And each provider should have total tokens
      And each provider should have total requests

    Scenario: Aggregate usage by user
      Given I have uploaded usage from multiple users
      When I get usage aggregated by user
      Then I should see data for each user
      And each user should have total tokens
      And each user should have total cost

    Scenario: Aggregate usage by model
      Given I have uploaded usage from multiple models
      When I get usage aggregated by model
      Then I should see data for each model
      And each model should have total tokens
      And each model should have average cost per token

    Scenario Outline: Aggregate usage for different time periods
      Given I have uploaded usage for the past <days> days
      When I get usage statistics for the period
      Then total tokens should be calculated correctly
      And total cost should be calculated correctly

      Examples:
        | days |
        | 1 |
        | 7 |
        | 30 |
        | 90 |

    Scenario: Get usage summary for date range
      Given I have uploaded usage between "2025-03-01" and "2025-03-15"
      When I get usage summary for the date range
      Then I should see summary statistics
      And the summary should include total tokens
      And the summary should include total cost
      And the summary should include total requests

    Scenario: Aggregate usage with hourly granularity
      Given I have uploaded usage for the past 24 hours
      When I get hourly usage statistics
      Then I should see 24 data points
      And each hour should have token count

    Scenario: Aggregate usage with weekly granularity
      Given I have uploaded usage for the past 4 weeks
      When I get weekly usage statistics
      Then I should see 4 data points
      And each week should have total tokens

    Scenario: Handle empty aggregation results
      Given I have not uploaded any usage data
      When I get daily usage statistics
      Then I should see empty results
      And the response should be valid

    Scenario: Aggregate usage for specific team
      Given I have uploaded usage for team "engineering"
      When I get team usage statistics
      Then I should see only engineering team usage
      And the total should be accurate

    Scenario: Calculate success rate from usage
      Given I have uploaded 100 usage records
      And 95 records succeeded
      And 5 records failed
      When I get usage statistics
      Then the success rate should be 95%

    Scenario: Aggregate usage by project metadata
      Given I have uploaded usage with project metadata
      When I get usage aggregated by project
      Then I should see data for each project
      And each project should have total tokens

    Scenario: Calculate average tokens per request
      Given I have uploaded 100 usage records
      And total tokens is 100000
      When I get usage statistics
      Then the average tokens per request should be 1000

  Rule: Usage Filtering

    Scenario: Filter usage by provider
      Given I have uploaded usage from multiple providers
      When I filter usage by provider "claude"
      Then I should only see claude usage
      And I should not see other provider usage

    Scenario: Filter usage by user
      Given I have uploaded usage from multiple users
      When I filter usage by user "user@example.com"
      Then I should only see usage for that user

    Scenario Outline: Filter usage by date range
      Given I have uploaded usage for the past 30 days
      When I query usage for the last <days> days
      Then I should only see data from the last <days> days

      Examples:
        | days |
        | 1 |
        | 7 |
        | 14 |
        | 30 |

    Scenario: Filter usage by model
      Given I have uploaded usage from multiple models
      When I filter usage by model "claude-3-opus"
      Then I should only see usage for that model

    Scenario: Filter usage by cost range
      Given I have uploaded usage with varying costs
      When I filter usage by cost range "$0.10 to $1.00"
      Then I should only see usage within that range

    Scenario: Filter usage by status
      Given I have uploaded usage with different statuses
      When I filter usage by status "success"
      Then I should only see successful usage

    Scenario: Combine multiple filters
      Given I have uploaded usage from multiple providers and users
      When I filter by provider "claude" and user "user@example.com"
      Then I should see only matching results

    Scenario: Filter usage by team
      Given I have uploaded usage for multiple teams
      When I filter usage by team "engineering"
      Then I should only see engineering team usage

    Scenario: Filter usage with pagination
      Given I have uploaded 1000 usage records
      When I get first page of results with page size 100
      Then I should see 100 records
      And I should see pagination information

    Scenario: Sort filtered usage by date
      Given I have uploaded usage for multiple dates
      When I filter usage and sort by date descending
      Then results should be ordered by date descending

    Scenario: Sort filtered usage by cost
      Given I have uploaded usage with varying costs
      When I filter usage and sort by cost ascending
      Then results should be ordered by cost ascending

    Scenario: Filter usage without results
      Given I have uploaded usage for provider "claude"
      When I filter usage by provider "codex"
      Then I should see empty results
      And the response should be valid

  Rule: Cost Calculation

    Scenario: Calculate cost for claude usage
      Given I have a claude usage record with 1000 tokens
      And claude costs $0.15 per 1000 tokens
      When I calculate the cost
      Then the cost should be $0.15

    Scenario Outline: Calculate cost for different providers
      Given I have a <provider> usage record with 1000 tokens
      And <provider> costs $<cost> per 1000 tokens
      When I calculate the cost
      Then the cost should be $<cost>

      Examples:
        | provider | cost |
        | claude | 0.15 |
        | codex | 0.10 |
        | opencode | 0.05 |

    Scenario: Calculate total cost for multiple records
      Given I have 3 usage records
      And each record costs $0.10
      When I calculate the total cost
      Then the total cost should be $0.30

    Scenario: Calculate cost with custom pricing tier
      Given I have custom pricing tier "enterprise"
      And the tier costs $0.10 per 1000 tokens
      When I calculate cost for 10000 tokens
      Then the cost should be $1.00

    Scenario: Handle free tier usage
      Given I have a free tier license
      And I upload 1000 tokens
      When I calculate the cost
      Then the cost should be $0.00

    Scenario: Calculate cost with overage charges
      Given I have a license with 10000 included tokens
      And I have used 15000 tokens
      And overage costs $0.20 per 1000 tokens
      When I calculate the total cost
      Then the base cost should be $0.00
      And the overage cost should be $1.00
      And the total cost should be $1.00

    Scenario: Calculate cost by model tier
      Given I have usage from "claude-3-opus"
      And "claude-3-opus" costs $0.30 per 1000 tokens
      When I calculate cost for 1000 tokens
      Then the cost should be $0.30

    Scenario: Aggregate costs by team
      Given I have uploaded usage for multiple teams
      When I get costs aggregated by team
      Then I should see cost for each team
      And the total should be the sum of all teams

    Scenario: Calculate cost savings from optimization
      Given I have usage before optimization costing $100
      And I have usage after optimization costing $80
      When I calculate cost savings
      Then the savings should be $20
      And the savings percentage should be 20%

    Scenario: Forecast cost based on usage trend
      Given I have uploaded usage for the past 30 days
      And the daily average is $10
      When I forecast cost for next 30 days
      Then the forecast should be $300

  Rule: Analytics Endpoints

    Scenario: Get personal usage dashboard
      Given I am logged in as a member
      And I have uploaded my usage data
      When I get my usage dashboard
      Then I should see my total usage
      And I should see my total cost
      And I should see my usage trend

    Scenario: Get team analytics as manager
      Given I am logged in as a manager
      And my team has uploaded usage data
      When I get team analytics
      Then I should see team total usage
      And I should see team total cost
      And I should see per-user breakdown

    Scenario: Get top users by usage
      Given I have uploaded usage for multiple users
      When I get top users by usage
      Then I should see users ranked by usage
      And the top user should have highest usage

    Scenario: Get top providers by usage
      Given I have uploaded usage for multiple providers
      When I get top providers by usage
      Then I should see providers ranked by usage

    Scenario: Get usage trends over time
      Given I have uploaded usage for the past 30 days
      When I get usage trends
      Then I should see trend data points
      And I should see moving average

    Scenario: Get cost breakdown by provider
      Given I have uploaded usage for multiple providers
      When I get cost breakdown
      Then I should see cost for each provider
      And I should see percentage of total

    Scenario: Get usage comparison between periods
      Given I have uploaded usage for this month
      And I have uploaded usage for last month
      When I get period comparison
      Then I should see percentage change
      And I should see absolute change

    Scenario: Member cannot see team analytics
      Given I am logged in as a member
      When I get team analytics
      Then I should receive a 403 error
      And the error message should contain "permission denied"

    Scenario: Export usage data as CSV
      Given I have uploaded usage data
      When I export usage data as CSV
      Then I should receive a CSV file
      And the file should contain usage records

    Scenario: Get real-time usage statistics
      Given I have uploaded usage data
      When I get real-time statistics
      Then I should see current usage
      And I should see today's cost

    Scenario: Get usage by project
      Given I have uploaded usage with project metadata
      When I get usage by project
      Then I should see data for each project
      And each project should have total cost

    Scenario: Get model performance metrics
      Given I have uploaded usage with success/failure data
      When I get model performance metrics
      Then I should see success rate by model
      And I should see average response time by model

    Scenario: Get cost optimization suggestions
      Given I have uploaded usage data
      When I get optimization suggestions
      Then I should see cost saving opportunities
      And I should see recommendations

    Scenario: Get usage alerts configuration
      Given I have configured usage alerts
      When I get alert configuration
      Then I should see my alert thresholds
      And I should see alert recipients

    Scenario: Anomaly detection in usage patterns
      Given I have uploaded usage for the past 30 days
      When I check for usage anomalies
      Then unusual patterns should be flagged
      And anomalies should be explained
