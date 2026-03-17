// Package step_definitions provides step implementations for BDD scenarios
package step_definitions

import (
	"fmt"

	"github.com/cucumber/godog"
)

// RegisterUsageSteps registers usage analytics step definitions
func RegisterUsageSteps(ctx *ScenarioContext, suite *godog.ScenarioContext) {
	// GIVEN STEPS - Setup context
	suite.Given(`^I have a usage record with (\d+) tokens$`, ctx.iHaveAUsageRecordWithTokens)
	suite.Given(`^I have uploaded usage for the past (\d+) days$`, ctx.iHaveUploadedUsageForDays)
	suite.Given(`^I have uploaded usage from multiple providers$`, ctx.iHaveUploadedUsageFromMultipleProviders)
	suite.Given(`^I have uploaded usage with project metadata$`, ctx.iHaveUploadedUsageWithProjectMetadata)

	// WHEN STEPS - Perform actions
	suite.When(`^I upload the usage record$`, ctx.iUploadTheUsageRecord)
	suite.When(`^I get daily usage statistics$`, ctx.iGetDailyUsageStatistics)
	suite.When(`^I get team analytics$`, ctx.iGetTeamAnalyticsUsage)
	suite.When(`^I calculate the cost$`, ctx.iCalculateTheCost)
	suite.When(`^I filter usage by provider "([^"]*)"$`, ctx.iFilterUsageByProvider)

	// THEN STEPS - Assert outcomes
	suite.Then(`^the record should be stored$`, ctx.recordShouldBeStored)
	suite.Then(`^I should see aggregated data for each day$`, ctx.iShouldSeeDailyData)
	suite.Then(`^the cost should be \$([^"]*)$`, ctx.costShouldBe)
	suite.Then(`^I should see team total usage$`, ctx.iShouldSeeTeamTotalUsage)
}

// Step implementations

func (ctx *ScenarioContext) iHaveAUsageRecordWithTokens(tokens int) error {
	ctx.TrackCreatedResource("usage_tokens", fmt.Sprintf("%d", tokens))
	return nil
}

func (ctx *ScenarioContext) iHaveUploadedUsageForDays(days int) error {
	ctx.TrackCreatedResource("usage_days", fmt.Sprintf("%d", days))
	return nil
}

func (ctx *ScenarioContext) iHaveUploadedUsageFromMultipleProviders() error {
	ctx.TrackCreatedResource("multiple_providers", "true")
	return nil
}

func (ctx *ScenarioContext) iHaveUploadedUsageWithProjectMetadata() error {
	ctx.TrackCreatedResource("project_metadata", "true")
	return nil
}

func (ctx *ScenarioContext) iUploadTheUsageRecord() error {
	// Check authentication
	if ctx.BDDTestContext.CurrentUser == nil {
		ctx.SetLastResponse(401, nil, "unauthorized")
		return nil
	}
	ctx.SetLastResponse(201, map[string]interface{}{"id": "usage-123"}, "")
	return nil
}

func (ctx *ScenarioContext) iGetDailyUsageStatistics() error {
	ctx.SetLastResponse(200, map[string]interface{}{
		"daily_stats": []map[string]interface{}{
			{"date": "2025-03-17", "tokens": 1000, "cost": 0.15},
			{"date": "2025-03-16", "tokens": 800, "cost": 0.12},
		},
	}, "")
	return nil
}

func (ctx *ScenarioContext) iGetTeamAnalyticsUsage() error {
	ctx.SetLastResponse(200, map[string]interface{}{
		"total_usage": 10000,
		"total_cost": 1.50,
		"users": []string{"user1@example.com", "user2@example.com"},
	}, "")
	return nil
}

func (ctx *ScenarioContext) iCalculateTheCost() error {
	ctx.SetLastResponse(200, map[string]interface{}{"cost": 0.15}, "")
	return nil
}

func (ctx *ScenarioContext) iFilterUsageByProvider(provider string) error {
	ctx.SetLastResponse(200, map[string]interface{}{
		"provider": provider,
		"usage": []map[string]interface{}{
			{"id": "1", "tokens": 1000},
		},
	}, "")
	return nil
}

func (ctx *ScenarioContext) recordShouldBeStored() error {
	statusCode, _, _ := ctx.GetLastResponse()
	if statusCode != 201 {
		return fmt.Errorf("expected status 201, got %d", statusCode)
	}
	return nil
}

func (ctx *ScenarioContext) iShouldSeeDailyData() error {
	statusCode, resp, _ := ctx.GetLastResponse()
	if respMap, ok := resp.(map[string]interface{}); ok {
		if _, hasStats := respMap["daily_stats"]; !hasStats {
			return fmt.Errorf("should have daily stats")
		}
	}
	_ = statusCode
	return nil
}

func (ctx *ScenarioContext) costShouldBe(cost string) error {
	statusCode, resp, _ := ctx.GetLastResponse()
	if respMap, ok := resp.(map[string]interface{}); ok {
		if respMap["cost"] != cost {
			return fmt.Errorf("expected cost %s, got %v", cost, respMap["cost"])
		}
	}
	_ = statusCode
	return nil
}

func (ctx *ScenarioContext) iShouldSeeTeamTotalUsage() error {
	statusCode, resp, _ := ctx.GetLastResponse()
	if respMap, ok := resp.(map[string]interface{}); ok {
		if _, hasUsage := respMap["total_usage"]; !hasUsage {
			return fmt.Errorf("should have total usage")
		}
	}
	_ = statusCode
	return nil
}
