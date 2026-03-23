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

    @p1 @requirement:UI-03-001
    Scenario: Upload single usage record successfully
      Given I have a usage record with 1000 tokens
      When I upload the usage record
      Then the record should be stored
      And the record should have an ID

    @p1 @requirement:UI-03-002
    Scenario: Upload batch of usage records
      Given I have 10 usage records
      When I upload the usage records as a batch
      Then all records should be stored
      And I should receive a 200 status

    @p1 @requirement:UI-03-003
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

    @p1 @requirement:UI-03-004
    Scenario: Upload usage record with provider information
      Given I have a usage record from provider "claude"
      And the record has 1000 tokens
      When I upload the usage record
      Then the record should be stored
      And the provider should be recorded

    @p1 @requirement:UI-03-005
    Scenario: Upload usage record with user information
      Given I have a usage record for user "user@example.com"
      And the record has 500 tokens
      When I upload the usage record
      Then the record should be stored
      And the user should be recorded

    @p1 @requirement:UI-03-006
    Scenario: Upload usage record with timestamp
      Given I have a usage record with timestamp "2025-03-17T10:00:00Z"
      When I upload the usage record
      Then the record should be stored
      And the timestamp should be recorded

    @p1 @requirement:UI-03-007
    Scenario: Upload usage record with model information
      Given I have a usage record for model "claude-3-opus"
      When I upload the usage record
      Then the record should be stored
      And the model should be recorded

    @p2 @requirement:UI-03-008
    Scenario: Reject usage record with invalid token count
      Given I have a usage record with -1 tokens
      When I upload the usage record
      Then I should receive a 400 error
      And the error message should contain "invalid token count"

    @p1 @requirement:UI-03-009
    Scenario: Reject usage record without required fields
      Given I have a usage record without provider
      When I upload the usage record
      Then I should receive a 400 error

    @p1 @requirement:UI-03-010
    Scenario: Upload usage record without authentication
      Given I am not authenticated
      And I have a usage record with 1000 tokens
      When I upload the usage record
      Then I should receive a 401 error

    @p1 @requirement:UI-03-011
    Scenario: Upload usage record with cost information
      Given I have a usage record with 1000 tokens and cost $0.15
      When I upload the usage record
      Then the record should be stored
      And the cost should be recorded

    @p2 @requirement:UI-03-012
    Scenario: Upload usage record with metadata
      Given I have a usage record with metadata
      And the metadata contains "project": "ai-assistant"
      When I upload the usage record
      Then the record should be stored
      And the metadata should be recorded

  Rule: Usage Aggregation

    @p1 @requirement:UI-03-013
    Scenario: Get daily usage statistics
      Given I have uploaded usage for the past 7 days
      When I get daily usage statistics
      Then I should see aggregated data for each day
      And each day should have total tokens
      And each day should have total cost

    @p1 @requirement:UI-03-014
    Scenario: Aggregate usage by provider
      Given I have uploaded usage from multiple providers
      When I get usage aggregated by provider
      Then I should see data for each provider
      And each provider should have total tokens
      And each provider should have total requests

    @p1 @requirement:UI-03-015
    Scenario: Aggregate usage by user
      Given I have uploaded usage from multiple users
      When I get usage aggregated by user
      Then I should see data for each user
      And each user should have total tokens
      And each user should have total cost

    @p1 @requirement:UI-03-016
    Scenario: Aggregate usage by model
      Given I have uploaded usage from multiple models
      When I get usage aggregated by model
      Then I should see data for each model
      And each model should have total tokens
      And each model should have average cost per token

    @p1 @requirement:UI-03-017
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

    @p1 @requirement:UI-03-018
    Scenario: Get usage summary for date range
      Given I have uploaded usage between "2025-03-01" and "2025-03-15"
      When I get usage summary for the date range
      Then I should see summary statistics
      And the summary should include total tokens
      And the summary should include total cost
      And the summary should include total requests

    @p2 @requirement:UI-03-019
    Scenario: Aggregate usage with hourly granularity
      Given I have uploaded usage for the past 24 hours
      When I get hourly usage statistics
      Then I should see 24 data points
      And each hour should have token count

    @p2 @requirement:UI-03-020
    Scenario: Aggregate usage with weekly granularity
      Given I have uploaded usage for the past 4 weeks
      When I get weekly usage statistics
      Then I should see 4 data points
      And each week should have total tokens

    @p2 @requirement:UI-03-021
    Scenario: Handle empty aggregation results
      Given I have not uploaded any usage data
      When I get daily usage statistics
      Then I should see empty results
      And the response should be valid

    @p1 @requirement:UI-03-022
    Scenario: Aggregate usage for specific team
      Given I have uploaded usage for team "engineering"
      When I get team usage statistics
      Then I should see only engineering team usage
      And the total should be accurate

    @p2 @requirement:UI-03-023
    Scenario: Calculate success rate from usage
      Given I have uploaded 100 usage records
      And 95 records succeeded
      And 5 records failed
      When I get usage statistics
      Then the success rate should be 95%

    @p3 @requirement:UI-03-024
    Scenario: Aggregate usage by project metadata
      Given I have uploaded usage with project metadata
      When I get usage aggregated by project
      Then I should see data for each project
      And each project should have total tokens

    @p2 @requirement:UI-03-025
    Scenario: Calculate average tokens per request
      Given I have uploaded 100 usage records
      And total tokens is 100000
      When I get usage statistics
      Then the average tokens per request should be 1000

  Rule: Usage Filtering

    @p1 @requirement:UI-03-026
    Scenario: Filter usage by provider
      Given I have uploaded usage from multiple providers
      When I filter usage by provider "claude"
      Then I should only see claude usage
      And I should not see other provider usage

    @p1 @requirement:UI-03-027
    Scenario: Filter usage by user
      Given I have uploaded usage from multiple users
      When I filter usage by user "user@example.com"
      Then I should only see usage for that user

    @p1 @requirement:UI-03-028
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

    @p1 @requirement:UI-03-029
    Scenario: Filter usage by model
      Given I have uploaded usage from multiple models
      When I filter usage by model "claude-3-opus"
      Then I should only see usage for that model

    @p2 @requirement:UI-03-030
    Scenario: Filter usage by cost range
      Given I have uploaded usage with varying costs
      When I filter usage by cost range "$0.10 to $1.00"
      Then I should only see usage within that range

    @p2 @requirement:UI-03-031
    Scenario: Filter usage by status
      Given I have uploaded usage with different statuses
      When I filter usage by status "success"
      Then I should only see successful usage

    @p1 @requirement:UI-03-032
    Scenario: Combine multiple filters
      Given I have uploaded usage from multiple providers and users
      When I filter by provider "claude" and user "user@example.com"
      Then I should see only matching results

    @p1 @requirement:UI-03-033
    Scenario: Filter usage by team
      Given I have uploaded usage for multiple teams
      When I filter usage by team "engineering"
      Then I should only see engineering team usage

    @p1 @requirement:UI-03-034
    Scenario: Filter usage with pagination
      Given I have uploaded 1000 usage records
      When I get first page of results with page size 100
      Then I should see 100 records
      And I should see pagination information

    @p2 @requirement:UI-03-035
    Scenario: Sort filtered usage by date
      Given I have uploaded usage for multiple dates
      When I filter usage and sort by date descending
      Then results should be ordered by date descending

    @p2 @requirement:UI-03-036
    Scenario: Sort filtered usage by cost
      Given I have uploaded usage with varying costs
      When I filter usage and sort by cost ascending
      Then results should be ordered by cost ascending

    @p2 @requirement:UI-03-037
    Scenario: Filter usage without results
      Given I have uploaded usage for provider "claude"
      When I filter usage by provider "codex"
      Then I should see empty results
      And the response should be valid

  Rule: Cost Calculation

    @p1 @requirement:UI-03-038
    Scenario: Calculate cost for claude usage
      Given I have a claude usage record with 1000 tokens
      And claude costs $0.15 per 1000 tokens
      When I calculate the cost
      Then the cost should be $0.15

    @p1 @requirement:UI-03-039
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

    @p1 @requirement:UI-03-040
    Scenario: Calculate total cost for multiple records
      Given I have 3 usage records
      And each record costs $0.10
      When I calculate the total cost
      Then the total cost should be $0.30

    @p2 @requirement:UI-03-041
    Scenario: Calculate cost with custom pricing tier
      Given I have custom pricing tier "enterprise"
      And the tier costs $0.10 per 1000 tokens
      When I calculate cost for 10000 tokens
      Then the cost should be $1.00

    @p2 @requirement:UI-03-042
    Scenario: Handle free tier usage
      Given I have a free tier license
      And I upload 1000 tokens
      When I calculate the cost
      Then the cost should be $0.00

    @p2 @requirement:UI-03-043
    Scenario: Calculate cost with overage charges
      Given I have a license with 10000 included tokens
      And I have used 15000 tokens
      And overage costs $0.20 per 1000 tokens
      When I calculate the total cost
      Then the base cost should be $0.00
      And the overage cost should be $1.00
      And the total cost should be $1.00

    @p1 @requirement:UI-03-044
    Scenario: Calculate cost by model tier
      Given I have usage from "claude-3-opus"
      And "claude-3-opus" costs $0.30 per 1000 tokens
      When I calculate cost for 1000 tokens
      Then the cost should be $0.30

    @p1 @requirement:UI-03-045
    Scenario: Aggregate costs by team
      Given I have uploaded usage for multiple teams
      When I get costs aggregated by team
      Then I should see cost for each team
      And the total should be the sum of all teams

    @p3 @requirement:UI-03-046
    Scenario: Calculate cost savings from optimization
      Given I have usage before optimization costing $100
      And I have usage after optimization costing $80
      When I calculate cost savings
      Then the savings should be $20
      And the savings percentage should be 20%

    @p3 @requirement:UI-03-047
    Scenario: Forecast cost based on usage trend
      Given I have uploaded usage for the past 30 days
      And the daily average is $10
      When I forecast cost for next 30 days
      Then the forecast should be $300

  Rule: Analytics Endpoints

    @p1 @requirement:UI-03-048
    Scenario: Get personal usage dashboard
      Given I am logged in as a member
      And I have uploaded my usage data
      When I get my usage dashboard
      Then I should see my total usage
      And I should see my total cost
      And I should see my usage trend

    @p1 @requirement:UI-03-049
    Scenario: Get team analytics as manager
      Given I am logged in as a manager
      And my team has uploaded usage data
      When I get team analytics
      Then I should see team total usage
      And I should see team total cost
      And I should see per-user breakdown

    @p1 @requirement:UI-03-050
    Scenario: Get top users by usage
      Given I have uploaded usage for multiple users
      When I get top users by usage
      Then I should see users ranked by usage
      And the top user should have highest usage

    @p1 @requirement:UI-03-051
    Scenario: Get top providers by usage
      Given I have uploaded usage for multiple providers
      When I get top providers by usage
      Then I should see providers ranked by usage

    @p1 @requirement:UI-03-052
    Scenario: Get usage trends over time
      Given I have uploaded usage for the past 30 days
      When I get usage trends
      Then I should see trend data points
      And I should see moving average

    @p1 @requirement:UI-03-053
    Scenario: Get cost breakdown by provider
      Given I have uploaded usage for multiple providers
      When I get cost breakdown
      Then I should see cost for each provider
      And I should see percentage of total

    @p2 @requirement:UI-03-054
    Scenario: Get usage comparison between periods
      Given I have uploaded usage for this month
      And I have uploaded usage for last month
      When I get period comparison
      Then I should see percentage change
      And I should see absolute change

    @p1 @requirement:UI-03-055
    Scenario: Member cannot see team analytics
      Given I am logged in as a member
      When I get team analytics
      Then I should receive a 403 error
      And the error message should contain "permission denied"

    @p2 @requirement:UI-03-056
    Scenario: Export usage data as CSV
      Given I have uploaded usage data
      When I export usage data as CSV
      Then I should receive a CSV file
      And the file should contain usage records

    @p1 @requirement:UI-03-057
    Scenario: Get real-time usage statistics
      Given I have uploaded usage data
      When I get real-time statistics
      Then I should see current usage
      And I should see today's cost

    @p3 @requirement:UI-03-058
    Scenario: Get usage by project
      Given I have uploaded usage with project metadata
      When I get usage by project
      Then I should see data for each project
      And each project should have total cost

    @p2 @requirement:UI-03-059
    Scenario: Get model performance metrics
      Given I have uploaded usage with success/failure data
      When I get model performance metrics
      Then I should see success rate by model
      And I should see average response time by model

    @p3 @requirement:UI-03-060
    Scenario: Get cost optimization suggestions
      Given I have uploaded usage data
      When I get optimization suggestions
      Then I should see cost saving opportunities
      And I should see recommendations

    @p3 @requirement:UI-03-061
    Scenario: Get usage alerts configuration
      Given I have configured usage alerts
      When I get alert configuration
      Then I should see my alert thresholds
      And I should see alert recipients

    @p3 @requirement:UI-03-062
    Scenario: Anomaly detection in usage patterns
      Given I have uploaded usage for the past 30 days
      When I check for usage anomalies
      Then unusual patterns should be flagged
      And anomalies should be explained

  Rule: Data Retention by License

    @p1 @requirement:UI-03-063
    Scenario: Open Source license retains data for 7 days
      Given I have an Open Source license
      And I have usage data from 8 days ago
      When I query usage statistics
      Then I should not see data older than 7 days

    @p1 @requirement:UI-03-064
    Scenario: Commercial license retains data for 90 days
      Given I have a Commercial license
      And I have usage data from 60 days ago
      When I query usage statistics
      Then I should see data from 60 days ago

    @p1 @requirement:UI-03-065
    Scenario: Data is automatically deleted after retention period
      Given I have an Open Source license
      And I have usage data from 10 days ago
      When the retention cleanup job runs
      Then the old data should be deleted

    @p2 @requirement:UI-03-066
    Scenario: License upgrade extends retention period
      Given I have usage data from 30 days ago
      And I upgrade from Open Source to Commercial license
      When I query usage statistics
      Then I should see data from 30 days ago

    @p2 @requirement:UI-03-067
    Scenario: License downgrade does not delete existing data immediately
      Given I have a Commercial license
      And I have usage data from 60 days ago
      When I downgrade to Open Source license
      Then the existing data should be retained
      But new data should follow 7-day retention

  Rule: Aggregation Accuracy

    @p1 @requirement:UI-03-068
    Scenario: Token sum aggregation is accurate
      Given I have created a claude provider "agg-test-provider"
      And I upload usage records with exact token counts
      When I get provider analytics for "agg-test-provider"
      Then the total input tokens should be 4500
      And the total output tokens should be 2250
      And the total tokens should be 6750
      And there should be no rounding errors

    @p1 @requirement:UI-03-069
    Scenario: Average calculation is accurate
      Given I have created a claude provider "avg-test-provider"
      And I upload 5 usage records with durations
      When I get provider analytics for "avg-test-provider"
      Then the average duration should be 3.0 seconds
      And the minimum duration should be 1.0 seconds
      And the maximum duration should be 5.0 seconds

    @p1 @requirement:UI-03-070
    Scenario: Model-level aggregation is accurate
      Given I have created a claude provider "model-agg-provider"
      And I upload usage across multiple models
      When I get provider analytics for "model-agg-provider"
      Then I should see data for all 3 models
      And the total tokens should be 18000
      And claude-3-opus should have 5 requests
      And claude-3-sonnet should have 3 requests
      And claude-3-haiku should have 2 requests

    @p1 @requirement:UI-03-071
    Scenario: Daily aggregation groups correctly by day
      Given I have created a claude provider "daily-agg-provider"
      And I upload usage over 3 days
      When I get daily usage statistics
      Then I should see 3 days of data
      And day 1 should have 5000 tokens
      And day 2 should have 3000 tokens
      And day 3 should have 7000 tokens
      And the total should be 15000 tokens

    @p1 @requirement:UI-03-072
    Scenario: Percentile calculation is accurate
      Given I have created a claude provider "percentile-provider"
      And I upload 100 usage records with varying token counts
      When I get provider analytics for "percentile-provider"
      Then the p50 percentile should be the median value
      And the p95 percentile should represent the 95th percentile
      And the p99 percentile should represent the 99th percentile

  Rule: Edge Cases and Anomalies

    @p1 @requirement:UI-03-073
    Scenario: Handle zero token values
      Given I have created a claude provider "zero-token-provider"
      And I upload usage records with zero tokens
      When I get provider analytics for "zero-token-provider"
      Then the aggregation should include zero-token records
      And the average should account for zeros
      And the request count should be accurate

    @p1 @requirement:UI-03-074
    Scenario: Handle negative token values as errors
      Given I have created a claude provider "negative-token-provider"
      And I upload a usage record with -100 tokens
      Then the record should be rejected
      And I should receive a 400 error
      And the error message should contain "negative"

    @p1 @requirement:UI-03-075
    Scenario: Cache tokens are included in totals
      Given I have created a claude provider "cache-provider"
      And I upload usage records with cache tokens:
        | record | input_tokens | output_tokens | cache_read_tokens |
        | 1      | 1000         | 500           | 200               |
        | 2      | 2000         | 1000          | 400               |
      When I get provider analytics for "cache-provider"
      Then the cache tokens should be included in totals
      And the total input tokens should be 3000
      And the total cache tokens should be 600

    @p1 @requirement:UI-03-076
    Scenario: Handle single record aggregation
      Given I have created a claude provider "single-record-provider"
      And I upload a single usage record with 1000 tokens
      When I get provider analytics for "single-record-provider"
      Then the total tokens should be 1000
      And the average should be 1000
      And the minimum should be 1000
      And the maximum should be 1000

    @p2 @requirement:UI-03-077
    Scenario: Handle malformed timestamps gracefully
      Given I have created a claude provider "timestamp-provider"
      And I upload usage records with invalid timestamps
      Then the records should be rejected
      And I should receive a 400 error

  Rule: Time-Based Filtering Accuracy

    @p1 @requirement:UI-03-078
    Scenario: Hourly aggregation returns 24 data points
      Given I have created a claude provider "hourly-provider"
      And I upload usage for the past 24 hours
      When I get hourly usage statistics
      Then I should see 24 data points
      And each hour should have correct token totals

    @p1 @requirement:UI-03-079
    Scenario: Weekly aggregation returns correct number of weeks
      Given I have created a claude provider "weekly-provider"
      And I upload usage for the past 4 weeks
      When I get weekly usage statistics
      Then I should see 4 data points
      And each week should have correct token totals

    @p1 @requirement:UI-03-080
    Scenario: Monthly aggregation handles partial months correctly
      Given I have created a claude provider "monthly-provider"
      And I upload usage from mid-month to mid-month
      When I get monthly usage statistics
      Then partial months should be handled correctly
      And the totals should be accurate

    @p1 @requirement:UI-03-081
    Scenario: Get current usage statistics
      When I get current usage
      Then I should see current usage data
      And the operation should succeed

  Rule: Usage Error Paths

    @p1 @requirement:UI-03-082
    Scenario: Get current usage without authentication
      Given I am not authenticated
      When I get current usage
      Then I should receive a 401 error

    @p1 @requirement:UI-03-083
    Scenario: Get usage statistics without authentication
      Given I am not authenticated
      When I get usage statistics
      Then I should receive a 401 error

    @p1 @requirement:UI-03-084
    Scenario: Upload usage record with invalid data
      When I upload usage record with invalid payload
      Then I should receive a 400 error
      And the error should indicate invalid data

    @p1 @requirement:UI-03-085
    Scenario: Upload usage batch with empty array
      When I upload empty usage batch
      Then the operation should succeed
      And no records should be stored

  Rule: Usage API Error Paths

    @p2 @requirement:UI-03-086
    Scenario: Get provider stats without authentication
      Given I am not authenticated
      When I get provider statistics for ID 1
      Then I should receive a 401 error

    @p2 @requirement:UI-03-087
    Scenario: List usage records with invalid date range
      Given I am logged in as a member
      When I get usage statistics with invalid date range
      Then the operation should succeed or return validation error

    @p2 @requirement:UI-03-088
    Scenario: Get personal analytics without authentication
      Given I am not authenticated
      When I get my usage statistics
      Then I should receive a 401 error

    @p2 @requirement:UI-03-089
    Scenario: Get team analytics without authentication
      Given I am not authenticated
      When I get team analytics
      Then I should receive a 401 error

  Rule: Usage Current Statistics

    @p1 @requirement:UI-03-090
    Scenario: Get current usage as manager
      Given I am logged in as a manager
      When I get current usage
      Then the operation should succeed
      And I should see usage data

    @p1 @requirement:UI-03-091
    Scenario: Get current usage as member
      Given I am logged in as a member
      When I get current usage
      Then the operation should succeed
      And I should see my usage data

    @p2 @requirement:UI-03-092
    Scenario: Get current usage after uploading records
      Given I am logged in as a manager
      And I have uploaded 5 usage records
      When I get current usage
      Then I should see usage data
      And the total should be greater than 0

  Rule: Usage Statistics Queries

    @p1 @requirement:UI-03-093
    Scenario: Get usage statistics as manager
      Given I am logged in as a manager
      When I get usage statistics
      Then the operation should succeed

    @p2 @requirement:UI-03-094
    Scenario: Get usage statistics with date filter
      Given I am logged in as a manager
      When I get usage statistics for today
      Then the operation should succeed

    @p1 @requirement:UI-03-095
    Scenario: Get usage statistics returns summary data
      Given I am logged in as a manager
      And I have uploaded usage records
      When I get usage statistics
      Then I should see total usage
      And I should see request count

  Rule: Usage Extended Queries

    @p1 @requirement:UI-03-096
    Scenario: Upload multiple usage records in sequence
      Given I am logged in as a manager
      When I upload a single usage record with 1000 tokens
      And I upload a single usage record with 2000 tokens
      Then both operations should succeed
      And the total usage should reflect both records

    @p1 @requirement:UI-03-097
    Scenario: Get provider statistics after usage upload
      Given I am logged in as a manager
      And I have created a provider
      And I upload usage records for that provider
      When I get provider statistics
      Then the statistics should show total requests
      And the operation should succeed

    @p2 @requirement:UI-03-098
    Scenario: Get current usage returns current period data
      Given I am logged in as a manager
      When I get current usage
      Then the operation should succeed
      And I should see usage data

    @p2 @requirement:UI-03-099
    Scenario: List usage records returns array
      Given I am logged in as a manager
      And I have uploaded usage records
      When I list all usage records
      Then I should see at least 1 usage record
      And the operation should succeed

    @p2 @requirement:UI-03-100
    Scenario: Get provider analytics as manager
      Given I am logged in as a manager
      And I have created a provider
      When I get provider analytics
      Then the operation should succeed
      And I should see analytics data

  Rule: Usage Statistics Extended

    @p1 @requirement:UI-03-101
    Scenario: Get usage statistics multiple times
      Given I am logged in as a manager
      When I get usage statistics
      And I get usage statistics
      And I get usage statistics
      Then all operations should succeed

    @p1 @requirement:UI-03-102
    Scenario: Get current usage after statistics
      Given I am logged in as a manager
      When I get usage statistics
      And I get current usage
      Then both operations should succeed

    @p2 @requirement:UI-03-103
    Scenario: List usage records returns all records
      Given I am logged in as a manager
      And I have uploaded 10 usage records
      When I list all usage records
      Then I should see at least 10 usage records

    @p2 @requirement:UI-03-104
    Scenario: Get provider analytics multiple times
      Given I am logged in as a manager
      And I have created a provider
      When I get provider analytics
      And I get provider analytics
      Then both operations should succeed

    @p2 @requirement:UI-03-105
    Scenario: Upload usage record then verify statistics
      Given I am logged in as a manager
      And I have created a provider
      When I upload a single usage record with 5000 tokens
      And I get provider statistics
      Then the statistics should show increased total

  Rule: Personal Usage Extended

    @p1 @requirement:UI-03-106
    Scenario: Get personal usage as member
      Given I am logged in as a member
      When I get my usage statistics
      Then the operation should succeed
      And I should see my usage data

    @p1 @requirement:UI-03-107
    Scenario: Get personal usage with date range
      Given I am logged in as a member
      When I get usage statistics for this week
      Then the operation should succeed

    @p2 @requirement:UI-03-108
    Scenario: Get personal usage without authentication fails
      Given I am not authenticated
      When I get my usage statistics
      Then I should receive a 401 error

    @p2 @requirement:UI-03-109
    Scenario: Upload usage without provider returns error
      Given I am logged in as a manager
      When I upload usage with invalid provider ID
      Then the operation should fail

  Rule: Analytics Extended

    @p1 @requirement:UI-03-110
    Scenario: Get analytics with different providers
      Given I am logged in as a manager
      And I have created a claude provider
      And I have created a codex provider
      When I get provider analytics for claude
      Then the operation should succeed
      When I get provider analytics for codex
      Then the operation should succeed

    @p2 @requirement:UI-03-111
    Scenario: Get analytics requires authentication
      Given I am not authenticated
      When I get provider analytics
      Then I should receive a 401 error

    @p2 @requirement:UI-03-112
    Scenario: Upload usage with large token count
      Given I am logged in as a manager
      And I have created a provider
      When I upload a single usage record with 100000 tokens
      Then the operation should succeed

  Rule: Usage Queries Extended

    @p1 @requirement:UI-03-113
    Scenario: List all usage records as manager
      Given I am logged in as a manager
      When I list all usage records
      Then the operation should succeed

    @p1 @requirement:UI-03-114
    Scenario: Get current usage as manager
      Given I am logged in as a manager
      When I get current usage statistics
      Then the operation should succeed

    @p2 @requirement:UI-03-115
    Scenario: List all usage records as member
      Given I am logged in as a member
      When I list all usage records
      Then the operation should succeed

    @p2 @requirement:UI-03-116
    Scenario: Get current usage without authentication fails
      Given I am not authenticated
      When I get current usage statistics
      Then I should receive a 401 error

  Rule: Usage Statistics Extended

    @p1 @requirement:UI-03-117
    Scenario: Get usage statistics multiple times
      Given I am logged in as a manager
      When I get usage statistics
      And I get usage statistics
      And I get usage statistics
      Then all operations should succeed

    @p1 @requirement:UI-03-118
    Scenario: Get current usage multiple times
      Given I am logged in as a manager
      When I get current usage statistics
      And I get current usage statistics
      And I get current usage statistics
      Then all operations should succeed

    @p1 @requirement:UI-03-119
    Scenario: Usage endpoints together
      Given I am logged in as a manager
      When I get usage statistics
      And I get current usage statistics
      Then both operations should succeed

    @p2 @requirement:UI-03-120
    Scenario: Get usage as member multiple times
      Given I am logged in as a member
      When I get current usage statistics
      And I get current usage statistics
      Then both operations should succeed

    @p2 @requirement:UI-03-121
    Scenario: Upload and get usage in sequence
      Given I am logged in as a manager
      When I upload a single usage record with 1000 tokens
      And I get usage statistics
      And I get current usage statistics
      Then all operations should succeed

  Rule: Usage Comprehensive

    @p1 @requirement:UI-03-122
    Scenario: All usage endpoints
      Given I am logged in as a manager
      When I get usage statistics
      And I get current usage statistics
      And I list all usage records
      Then all operations should succeed

    @p2 @requirement:UI-03-123
    Scenario: Usage operations repeated
      Given I am logged in as a manager
      When I get usage statistics
      And I get current usage statistics
      And I get usage statistics
      And I get current usage statistics
      Then all operations should succeed

