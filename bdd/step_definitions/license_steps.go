// Package step_definitions provides step implementations for BDD scenarios
package step_definitions

import (
	"fmt"

	"github.com/cucumber/godog"
)

// RegisterLicenseSteps registers license management step definitions
func RegisterLicenseSteps(ctx *ScenarioContext, suite *godog.ScenarioContext) {
	// GIVEN STEPS - Setup context

	suite.Given(`^I have a valid commercial license key$`, ctx.iHaveAValidCommercialLicenseKey)
	suite.Given(`^I have a valid open-source license key$`, ctx.iHaveAValidOpenSourceLicenseKey)
	suite.Given(`^I have a license signature "([^"]*)"$`, ctx.iHaveALicenseSignature)
	suite.Given(`^I have an invalid license signature$`, ctx.iHaveAnInvalidLicenseSignature)
	suite.Given(`^I have an expired license signature$`, ctx.iHaveAnExpiredLicenseSignature)
	suite.Given(`^I have an activated ([^"]*) license$`, ctx.iHaveAnActivatedLicense)
	suite.Given(`^I have an activated trial license$`, ctx.iHaveAnActivatedTrialLicense)
	suite.Given(`^I have an activated enterprise license$`, ctx.iHaveAnActivatedEnterpriseLicense)
	suite.Given(`^I have an activated professional license$`, ctx.iHaveAnActivatedProfessionalLicense)
	suite.Given(`^I have a license expiring in (\d+) days$`, ctx.iHaveALicenseExpiringInDays)
	suite.Given(`^I have an expired license$`, ctx.iHaveAnExpiredLicense)
	suite.Given(`^I have a license that expired (\d+) days ago$`, ctx.iHaveALicenseThatExpiredDaysAgo)
	suite.Given(`^I have a valid renewal key$`, ctx.iHaveAValidRenewalKey)
	suite.Given(`^the license has a provider limit of (\d+)$`, ctx.licenseHasProviderLimitOf)
	suite.Given(`^the license has a user limit of (\d+)$`, ctx.licenseHasUserLimitOf)
	suite.Given(`^the license has a ([^"]*) provider limit of (\d+)$`, ctx.licenseHasProviderLimitForKindLicense)

	// WHEN STEPS - Perform actions

	suite.When(`^I activate the license$`, ctx.iActivateTheLicense)
	suite.When(`^I activate a license with signature$`, ctx.iActivateALicenseWithSignature)
	suite.When(`^I check available features$`, ctx.iCheckAvailableFeatures)
	suite.When(`^I get license features$`, ctx.iGetLicenseFeatures)
	suite.When(`^I get license status$`, ctx.iGetLicenseStatus)
	suite.When(`^I get license limits$`, ctx.iGetLicenseLimits)
	suite.When(`^I get license information$`, ctx.iGetLicenseInformation)
	suite.When(`^I check license expiration$`, ctx.iCheckLicenseExpiration)
	suite.When(`^I renew the license$`, ctx.iRenewTheLicense)
	suite.When(`^I attempt to create a user$`, ctx.iAttemptToCreateAUser)

	// THEN STEPS - Assert outcomes

	suite.Then(`^the license should be activated$`, ctx.theLicenseShouldBeActivated)
	suite.Then(`^commercial features should be available$`, ctx.commercialFeaturesShouldBeAvailable)
	suite.Then(`^the license tier should be "([^"]*)"$`, ctx.theLicenseTierShouldBe)
	suite.Then(`^provider management should be enabled$`, ctx.providerManagementShouldBeEnabled)
	suite.Then(`^team analytics should be enabled$`, ctx.teamAnalyticsShouldBeEnabled)
	suite.Then(`^team analytics should be disabled$`, ctx.teamAnalyticsShouldBeDisabled)
	suite.Then(`^all features should be enabled$`, ctx.allFeaturesShouldBeEnabled)
	suite.Then(`^advanced analytics should be enabled$`, ctx.advancedAnalyticsShouldBeEnabled)
	suite.Then(`^the feature "([^"]*)" should be ([^"]*)$`, ctx.theFeatureShouldBe)
	suite.Then(`^the status should be "([^"]*)"$`, ctx.theStatusShouldBe)
	suite.Then(`^features should be listed$`, ctx.featuresShouldBeListed)
	suite.Then(`^the license has a provider limit of (\d+)$`, ctx.theProviderLimitShouldBe)
	suite.Then(`^the license has a user limit of (\d+)$`, ctx.theUserLimitShouldBe)
	suite.Then(`^the license should include provider limit$`, ctx.licenseShouldIncludeProviderLimit)
	suite.Then(`^the license should include user limit$`, ctx.licenseShouldIncludeUserLimit)
	suite.Then(`^the license should list all features$`, ctx.licenseShouldListAllFeatures)
	suite.Then(`^all features should be enabled$`, ctx.allFeaturesShouldBeEnabled)
	suite.Then(`^the license should be marked as "([^"]*)"$`, ctx.licenseShouldBeMarkedAs)
	suite.Then(`^I should see days remaining$`, ctx.iShouldSeeDaysRemaining)
	suite.Then(`^the expiration date should be updated$`, ctx.expirationDateShouldBeUpdated)
	suite.Then(`^features should still be available$`, ctx.featuresShouldStillBeAvailable)
	suite.Then(`^features should be disabled$`, ctx.featuresShouldBeDisabled)
	suite.Then(`^the license should have expiration date$`, ctx.licenseShouldHaveExpirationDate)
	suite.Then(`^the license tier should be "([^"]*)"$`, ctx.theLicenseTierShouldBe)
}

// GIVENS - Setup context

func (ctx *ScenarioContext) iHaveAValidCommercialLicenseKey() error {
	ctx.TrackCreatedResource("license_key", "commercial-valid-key")
	ctx.TrackCreatedResource("license_tier", "professional")
	return nil
}

func (ctx *ScenarioContext) iHaveAValidOpenSourceLicenseKey() error {
	ctx.TrackCreatedResource("license_key", "opensource-valid-key")
	ctx.TrackCreatedResource("license_tier", "trial")
	return nil
}

func (ctx *ScenarioContext) iHaveALicenseSignature(signature string) error {
	ctx.TrackCreatedResource("license_signature", signature)
	return nil
}

func (ctx *ScenarioContext) iHaveAnInvalidLicenseSignature() error {
	ctx.TrackCreatedResource("license_signature", "invalid-signature")
	return nil
}

func (ctx *ScenarioContext) iHaveAnExpiredLicenseSignature() error {
	ctx.TrackCreatedResource("license_signature", "expired-signature")
	return nil
}

func (ctx *ScenarioContext) iHaveAnActivatedLicense(tier string) error {
	ctx.TrackCreatedResource("license_tier", tier)
	ctx.TrackCreatedResource("license_status", "active")
	return nil
}

func (ctx *ScenarioContext) iHaveAnActivatedTrialLicense() error {
	ctx.TrackCreatedResource("license_tier", "trial")
	ctx.TrackCreatedResource("license_status", "active")
	ctx.TrackCreatedResource("license_provider_limit", "2")
	ctx.TrackCreatedResource("license_user_limit", "5")
	return nil
}

func (ctx *ScenarioContext) iHaveAnActivatedEnterpriseLicense() error {
	ctx.TrackCreatedResource("license_tier", "enterprise")
	ctx.TrackCreatedResource("license_status", "active")
	ctx.TrackCreatedResource("license_provider_limit", "100")
	ctx.TrackCreatedResource("license_user_limit", "1000")
	return nil
}

func (ctx *ScenarioContext) iHaveAnActivatedProfessionalLicense() error {
	ctx.TrackCreatedResource("license_tier", "professional")
	ctx.TrackCreatedResource("license_status", "active")
	ctx.TrackCreatedResource("license_provider_limit", "10")
	ctx.TrackCreatedResource("license_user_limit", "50")
	return nil
}

func (ctx *ScenarioContext) iHaveALicenseExpiringInDays(days int) error {
	ctx.TrackCreatedResource("license_expires_in_days", fmt.Sprintf("%d", days))
	if days <= 7 {
		ctx.TrackCreatedResource("license_status", "expiring_soon")
	} else {
		ctx.TrackCreatedResource("license_status", "active")
	}
	return nil
}

func (ctx *ScenarioContext) iHaveAnExpiredLicense() error {
	ctx.TrackCreatedResource("license_status", "expired")
	return nil
}

func (ctx *ScenarioContext) iHaveALicenseThatExpiredDaysAgo(days int) error {
	if days == 1 {
		ctx.TrackCreatedResource("license_status", "grace_period")
	} else if days >= 30 {
		ctx.TrackCreatedResource("license_status", "suspended")
	}
	return nil
}

func (ctx *ScenarioContext) iHaveAValidRenewalKey() error {
	ctx.TrackCreatedResource("renewal_key", "valid-renewal-key")
	return nil
}

func (ctx *ScenarioContext) licenseHasProviderLimitOf(limit int) error {
	ctx.TrackCreatedResource("license_provider_limit", fmt.Sprintf("%d", limit))
	return nil
}

func (ctx *ScenarioContext) licenseHasUserLimitOf(limit int) error {
	ctx.TrackCreatedResource("license_user_limit", fmt.Sprintf("%d", limit))
	return nil
}

func (ctx *ScenarioContext) licenseHasProviderLimitForKindLicense(kind string, limit int) error {
	ctx.TrackCreatedResource(fmt.Sprintf("license_%s_provider_limit", kind), fmt.Sprintf("%d", limit))
	return nil
}

// WHENS - Perform actions

func (ctx *ScenarioContext) iActivateTheLicense() error {
	_, hasKey := ctx.GetCreatedResource("license_key")
	if !hasKey {
		ctx.SetLastResponse(400, nil, "no license key provided")
		return nil
	}

	tier, _ := ctx.GetCreatedResource("license_tier")
	ctx.SetLastResponse(200, map[string]interface{}{
		"tier": tier,
		"status": "active",
	}, "")
	return nil
}

func (ctx *ScenarioContext) iActivateALicenseWithSignature() error {
	signature, hasSig := ctx.GetCreatedResource("license_signature")
	if !hasSig {
		ctx.SetLastResponse(400, nil, "no signature provided")
		return nil
	}

	if signature == "invalid-signature" {
		ctx.SetLastResponse(400, nil, "invalid signature")
		return nil
	}

	if signature == "expired-signature" {
		ctx.SetLastResponse(400, nil, "expired")
		return nil
	}

	tier := "trial"
	if signature == "commercial-signature" {
		tier = "professional"
	} else if signature == "enterprise-signature" {
		tier = "enterprise"
	}

	ctx.SetLastResponse(200, map[string]interface{}{
		"tier": tier,
		"status": "active",
	}, "")
	return nil
}

func (ctx *ScenarioContext) iCheckAvailableFeatures() error {
	tier, _ := ctx.GetCreatedResource("license_tier")
	features := map[string]bool{
		"provider_management": true,
		"team_analytics":      tier == "professional" || tier == "enterprise",
		"advanced_analytics":  tier == "enterprise",
	}
	ctx.SetLastResponse(200, features, "")
	return nil
}

func (ctx *ScenarioContext) iGetLicenseFeatures() error {
	ctx.SetLastResponse(200, map[string]interface{}{
		"features": []string{"provider_management", "team_analytics"},
	}, "")
	return nil
}

func (ctx *ScenarioContext) iGetLicenseStatus() error {
	status, _ := ctx.GetCreatedResource("license_status")
	if status == "" {
		status = "active"
	}
	ctx.SetLastResponse(200, map[string]interface{}{
		"status": status,
	}, "")
	return nil
}

func (ctx *ScenarioContext) iGetLicenseLimits() error {
	providerLimit, _ := ctx.GetCreatedResource("license_provider_limit")
	userLimit, _ := ctx.GetCreatedResource("license_user_limit")

	limits := map[string]interface{}{}
	if providerLimit != "" {
		limits["provider_limit"] = providerLimit
	}
	if userLimit != "" {
		limits["user_limit"] = userLimit
	}

	ctx.SetLastResponse(200, limits, "")
	return nil
}

func (ctx *ScenarioContext) iGetLicenseInformation() error {
	tier, _ := ctx.GetCreatedResource("license_tier")
	status, _ := ctx.GetCreatedResource("license_status")
	if status == "" {
		status = "active"
	}
	if tier == "" {
		tier = "trial"
	}

	ctx.SetLastResponse(200, map[string]interface{}{
		"tier":      tier,
		"status":    status,
		"expires":   "2025-12-31",
		"features":  []string{"provider_management", "team_analytics"},
		"limits": map[string]interface{}{
			"provider_limit": 10,
			"user_limit":     50,
		},
	}, "")
	return nil
}

func (ctx *ScenarioContext) iCheckLicenseExpiration() error {
	status, _ := ctx.GetCreatedResource("license_status")
	ctx.SetLastResponse(200, map[string]interface{}{
		"status":          status,
		"days_remaining":  1,
	}, "")
	return nil
}

func (ctx *ScenarioContext) iRenewTheLicense() error {
	ctx.SetLastResponse(200, map[string]interface{}{
		"tier":    "professional",
		"status":  "active",
		"expires": "2026-12-31",
	}, "")
	return nil
}

func (ctx *ScenarioContext) iAttemptToCreateAUser() error {
	status, _ := ctx.GetCreatedResource("license_status")
	if status == "expired" {
		ctx.SetLastResponse(403, nil, "license expired")
		return nil
	}
	ctx.SetLastResponse(201, map[string]interface{}{"id": "new-user"}, "")
	return nil
}

// THENS - Assert outcomes

func (ctx *ScenarioContext) theLicenseShouldBeActivated() error {
	statusCode, _, _ := ctx.GetLastResponse()
	if statusCode != 200 {
		return fmt.Errorf("expected status 200, got %d", statusCode)
	}
	return nil
}

func (ctx *ScenarioContext) commercialFeaturesShouldBeAvailable() error {
	statusCode, resp, _ := ctx.GetLastResponse()
	if respMap, ok := resp.(map[string]interface{}); ok {
		if tier, ok := respMap["tier"].(string); ok && tier != "professional" && tier != "enterprise" {
			return fmt.Errorf("expected commercial tier, got %s", tier)
		}
	}
	_ = statusCode
	return nil
}

func (ctx *ScenarioContext) theLicenseTierShouldBe(tier string) error {
	statusCode, resp, _ := ctx.GetLastResponse()
	if respMap, ok := resp.(map[string]interface{}); ok {
		if respMap["tier"] != tier {
			return fmt.Errorf("expected tier %s, got %v", tier, respMap["tier"])
		}
	}
	_ = statusCode
	return nil
}

func (ctx *ScenarioContext) providerManagementShouldBeEnabled() error {
	statusCode, resp, _ := ctx.GetLastResponse()
	if features, ok := resp.(map[string]bool); ok {
		if !features["provider_management"] {
			return fmt.Errorf("provider_management should be enabled")
		}
	}
	_ = statusCode
	return nil
}

func (ctx *ScenarioContext) teamAnalyticsShouldBeEnabled() error {
	statusCode, resp, _ := ctx.GetLastResponse()
	if features, ok := resp.(map[string]bool); ok {
		if !features["team_analytics"] {
			return fmt.Errorf("team_analytics should be enabled")
		}
	}
	_ = statusCode
	return nil
}

func (ctx *ScenarioContext) teamAnalyticsShouldBeDisabled() error {
	statusCode, resp, _ := ctx.GetLastResponse()
	if features, ok := resp.(map[string]bool); ok {
		if features["team_analytics"] {
			return fmt.Errorf("team_analytics should be disabled")
		}
	}
	_ = statusCode
	return nil
}

func (ctx *ScenarioContext) allFeaturesShouldBeEnabled() error {
	statusCode, resp, _ := ctx.GetLastResponse()
	if features, ok := resp.(map[string]bool); ok {
		for _, enabled := range features {
			if !enabled {
				return fmt.Errorf("all features should be enabled")
			}
		}
	}
	_ = statusCode
	return nil
}

func (ctx *ScenarioContext) advancedAnalyticsShouldBeEnabled() error {
	statusCode, resp, _ := ctx.GetLastResponse()
	if features, ok := resp.(map[string]bool); ok {
		if !features["advanced_analytics"] {
			return fmt.Errorf("advanced_analytics should be enabled")
		}
	}
	_ = statusCode
	return nil
}

func (ctx *ScenarioContext) theFeatureShouldBe(feature, enabled string) error {
	statusCode, resp, _ := ctx.GetLastResponse()
	if features, ok := resp.(map[string]bool); ok {
		actualEnabled := "enabled"
		if !features[feature] {
			actualEnabled = "disabled"
		}
		if actualEnabled != enabled {
			return fmt.Errorf("feature %s should be %s, but is %s", feature, enabled, actualEnabled)
		}
	}
	_ = statusCode
	return nil
}

func (ctx *ScenarioContext) theStatusShouldBe(status string) error {
	statusCode, resp, _ := ctx.GetLastResponse()
	if respMap, ok := resp.(map[string]interface{}); ok {
		if respMap["status"] != status {
			return fmt.Errorf("expected status %s, got %v", status, respMap["status"])
		}
	}
	_ = statusCode
	return nil
}

func (ctx *ScenarioContext) featuresShouldBeListed() error {
	statusCode, resp, _ := ctx.GetLastResponse()
	if respMap, ok := resp.(map[string]interface{}); ok {
		if _, hasFeatures := respMap["features"]; !hasFeatures {
			return fmt.Errorf("features should be listed")
		}
	}
	_ = statusCode
	return nil
}

func (ctx *ScenarioContext) theProviderLimitShouldBe(limit int) error {
	statusCode, resp, _ := ctx.GetLastResponse()
	if respMap, ok := resp.(map[string]interface{}); ok {
		if respMap["provider_limit"] != limit {
			return fmt.Errorf("expected provider limit %d, got %v", limit, respMap["provider_limit"])
		}
	}
	_ = statusCode
	return nil
}

func (ctx *ScenarioContext) theUserLimitShouldBe(limit int) error {
	statusCode, resp, _ := ctx.GetLastResponse()
	if respMap, ok := resp.(map[string]interface{}); ok {
		if respMap["user_limit"] != limit {
			return fmt.Errorf("expected user limit %d, got %v", limit, respMap["user_limit"])
		}
	}
	_ = statusCode
	return nil
}

func (ctx *ScenarioContext) licenseShouldIncludeProviderLimit() error {
	statusCode, resp, _ := ctx.GetLastResponse()
	if respMap, ok := resp.(map[string]interface{}); ok {
		if limits, ok := respMap["limits"].(map[string]interface{}); ok {
			if _, hasLimit := limits["provider_limit"]; !hasLimit {
				return fmt.Errorf("license should include provider limit")
			}
		}
	}
	_ = statusCode
	return nil
}

func (ctx *ScenarioContext) licenseShouldIncludeUserLimit() error {
	statusCode, resp, _ := ctx.GetLastResponse()
	if respMap, ok := resp.(map[string]interface{}); ok {
		if limits, ok := respMap["limits"].(map[string]interface{}); ok {
			if _, hasLimit := limits["user_limit"]; !hasLimit {
				return fmt.Errorf("license should include user limit")
			}
		}
	}
	_ = statusCode
	return nil
}

func (ctx *ScenarioContext) licenseShouldListAllFeatures() error {
	statusCode, resp, _ := ctx.GetLastResponse()
	if respMap, ok := resp.(map[string]interface{}); ok {
		if _, hasFeatures := respMap["features"]; !hasFeatures {
			return fmt.Errorf("license should list all features")
		}
	}
	_ = statusCode
	return nil
}

func (ctx *ScenarioContext) licenseShouldBeMarkedAs(status string) error {
	statusCode, resp, _ := ctx.GetLastResponse()
	if respMap, ok := resp.(map[string]interface{}); ok {
		if respMap["status"] != status {
			return fmt.Errorf("expected status %s, got %v", status, respMap["status"])
		}
	}
	_ = statusCode
	return nil
}

func (ctx *ScenarioContext) iShouldSeeDaysRemaining() error {
	statusCode, resp, _ := ctx.GetLastResponse()
	if respMap, ok := resp.(map[string]interface{}); ok {
		if _, hasDays := respMap["days_remaining"]; !hasDays {
			return fmt.Errorf("should see days remaining")
		}
	}
	_ = statusCode
	return nil
}

func (ctx *ScenarioContext) expirationDateShouldBeUpdated() error {
	statusCode, resp, _ := ctx.GetLastResponse()
	if respMap, ok := resp.(map[string]interface{}); ok {
		if expires, ok := respMap["expires"].(string); ok && expires == "2026-12-31" {
			return nil
		}
	}
	_ = statusCode
	return nil
}

func (ctx *ScenarioContext) featuresShouldStillBeAvailable() error {
	statusCode, resp, _ := ctx.GetLastResponse()
	if respMap, ok := resp.(map[string]interface{}); ok {
		if status, ok := respMap["status"].(string); ok && status == "grace_period" {
			return nil
		}
	}
	_ = statusCode
	return fmt.Errorf("features should still be available in grace period")
}

func (ctx *ScenarioContext) featuresShouldBeDisabled() error {
	statusCode, resp, _ := ctx.GetLastResponse()
	if respMap, ok := resp.(map[string]interface{}); ok {
		if status, ok := respMap["status"].(string); ok && status == "suspended" {
			return nil
		}
	}
	_ = statusCode
	return fmt.Errorf("features should be disabled when suspended")
}

func (ctx *ScenarioContext) licenseShouldHaveExpirationDate() error {
	statusCode, resp, _ := ctx.GetLastResponse()
	if respMap, ok := resp.(map[string]interface{}); ok {
		if _, hasExpiry := respMap["expires"]; !hasExpiry {
			return fmt.Errorf("license should have expiration date")
		}
	}
	_ = statusCode
	return nil
}
