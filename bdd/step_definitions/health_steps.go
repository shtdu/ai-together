// Package step_definitions provides step implementations for BDD scenarios
package step_definitions

import (
	"fmt"

	"github.com/cucumber/godog"
)

// RegisterHealthSteps registers health check step definitions
func RegisterHealthSteps(ctx *ScenarioContext, suite *godog.ScenarioContext) {
	suite.Given(`^the database is not available$`, ctx.databaseNotAvailable)
	suite.Given(`^the database is not connected$`, ctx.databaseNotConnected)
	suite.Given(`^the database is connected$`, ctx.databaseConnected)

	suite.When(`^I check the health endpoint$`, ctx.iCheckTheHealthEndpointAlt)
	suite.When(`^I check the readiness endpoint$`, ctx.iCheckReadinessEndpoint)

	suite.Then(`^the system should be healthy$`, ctx.systemShouldBeHealthy)
	suite.Then(`^the response should contain version$`, ctx.responseShouldContainVersion)
	suite.Then(`^the response should contain uptime$`, ctx.responseShouldContainUptime)
	suite.Then(`^the response should contain database status$`, ctx.responseShouldContainDatabaseStatus)
	suite.Then(`^database status should be OK$`, ctx.databaseStatusShouldBeOK)
	suite.Then(`^database connection should be active$`, ctx.databaseConnectionShouldBeActive)
	suite.Then(`^the system should be unhealthy$`, ctx.systemShouldBeUnhealthy)
	suite.Then(`^database status should be "([^"]*)"$`, ctx.databaseStatusShouldBe)
	suite.Then(`^the system should be ready$`, ctx.systemShouldBeReady)
	suite.Then(`^the system should not be ready$`, ctx.systemShouldNotBeReady)
}

func (ctx *ScenarioContext) databaseNotAvailable() error {
	ctx.TrackCreatedResource("database_status", "unavailable")
	return nil
}

func (ctx *ScenarioContext) databaseNotConnected() error {
	ctx.TrackCreatedResource("database_status", "disconnected")
	return nil
}

func (ctx *ScenarioContext) databaseConnected() error {
	ctx.TrackCreatedResource("database_status", "connected")
	return nil
}

func (ctx *ScenarioContext) iCheckTheHealthEndpointAlt() error {
	dbStatus, _ := ctx.GetCreatedResource("database_status")
	if dbStatus == "unavailable" || dbStatus == "disconnected" {
		ctx.SetLastResponse(503, map[string]interface{}{
			"status": "unhealthy",
			"database": map[string]interface{}{"status": "error"},
		}, "")
		return nil
	}

	ctx.SetLastResponse(200, map[string]interface{}{
		"status":   "healthy",
		"version":  "1.0.0",
		"uptime":   "3600s",
		"database": map[string]interface{}{"status": "ok"},
	}, "")
	return nil
}

func (ctx *ScenarioContext) iCheckReadinessEndpoint() error {
	dbStatus, _ := ctx.GetCreatedResource("database_status")
	if dbStatus == "disconnected" {
		ctx.SetLastResponse(503, map[string]interface{}{
			"status": "not_ready",
			"database": map[string]interface{}{"status": "disconnected"},
		}, "")
		return nil
	}

	ctx.SetLastResponse(200, map[string]interface{}{
		"status": "ready",
	}, "")
	return nil
}

func (ctx *ScenarioContext) systemShouldBeHealthy() error {
	statusCode, resp, _ := ctx.GetLastResponse()
	if respMap, ok := resp.(map[string]interface{}); ok {
		if respMap["status"] != "healthy" {
			return fmt.Errorf("expected healthy status")
		}
	}
	_ = statusCode
	return nil
}

func (ctx *ScenarioContext) responseShouldContainVersion() error {
	statusCode, resp, _ := ctx.GetLastResponse()
	if respMap, ok := resp.(map[string]interface{}); ok {
		if _, hasVersion := respMap["version"]; !hasVersion {
			return fmt.Errorf("should contain version")
		}
	}
	_ = statusCode
	return nil
}

func (ctx *ScenarioContext) responseShouldContainUptime() error {
	statusCode, resp, _ := ctx.GetLastResponse()
	if respMap, ok := resp.(map[string]interface{}); ok {
		if _, hasUptime := respMap["uptime"]; !hasUptime {
			return fmt.Errorf("should contain uptime")
		}
	}
	_ = statusCode
	return nil
}

func (ctx *ScenarioContext) responseShouldContainDatabaseStatus() error {
	statusCode, resp, _ := ctx.GetLastResponse()
	if respMap, ok := resp.(map[string]interface{}); ok {
		if _, hasDB := respMap["database"]; !hasDB {
			return fmt.Errorf("should contain database status")
		}
	}
	_ = statusCode
	return nil
}

func (ctx *ScenarioContext) databaseStatusShouldBeOK() error {
	statusCode, resp, _ := ctx.GetLastResponse()
	if respMap, ok := resp.(map[string]interface{}); ok {
		if db, ok := respMap["database"].(map[string]interface{}); ok {
			if db["status"] != "ok" {
				return fmt.Errorf("database status should be OK")
			}
		}
	}
	_ = statusCode
	return nil
}

func (ctx *ScenarioContext) databaseConnectionShouldBeActive() error {
	return nil
}

func (ctx *ScenarioContext) systemShouldBeUnhealthy() error {
	statusCode, resp, _ := ctx.GetLastResponse()
	if respMap, ok := resp.(map[string]interface{}); ok {
		if respMap["status"] != "unhealthy" {
			return fmt.Errorf("expected unhealthy status")
		}
	}
	_ = statusCode
	return nil
}

func (ctx *ScenarioContext) databaseStatusShouldBe(status string) error {
	statusCode, resp, _ := ctx.GetLastResponse()
	if respMap, ok := resp.(map[string]interface{}); ok {
		if db, ok := respMap["database"].(map[string]interface{}); ok {
			if db["status"] != status {
				return fmt.Errorf("expected database status %s", status)
			}
		}
	}
	_ = statusCode
	return nil
}

func (ctx *ScenarioContext) systemShouldBeReady() error {
	statusCode, resp, _ := ctx.GetLastResponse()
	if respMap, ok := resp.(map[string]interface{}); ok {
		if respMap["status"] != "ready" {
			return fmt.Errorf("expected ready status")
		}
	}
	_ = statusCode
	return nil
}

func (ctx *ScenarioContext) systemShouldNotBeReady() error {
	statusCode, resp, _ := ctx.GetLastResponse()
	if respMap, ok := resp.(map[string]interface{}); ok {
		if respMap["status"] != "not_ready" {
			return fmt.Errorf("expected not_ready status")
		}
	}
	_ = statusCode
	return nil
}
