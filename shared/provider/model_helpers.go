// Copyright (c) 2025 AI Together
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.


package provider

import (
	"fmt"
	"strings"
)

// IsModelSupported checks if provider supports specified model
// Support conditions: 1) Model is in SupportedModels (exact or wildcard match)
//  2. Model is in ModelMapping keys (exact or wildcard match)
func IsModelSupported(p Provider, modelName string) bool {
	// Backward compatible: If whitelist and mapping are not configured, assume all models are supported
	if len(p.GetSupportedModels()) == 0 &&
		len(p.GetModelMapping()) == 0 {
		return true
	}

	// Scenario A: Provider natively supports this model (exact match)
	for _, supportedModel := range p.GetSupportedModels() {
		if supportedModel == modelName {
			return true
		}
	}

	// Scenario A+: Provider natively supports this model (wildcard match)
	for _, supportedModel := range p.GetSupportedModels() {
		if matchWildcard(supportedModel, modelName) {
			return true
		}
	}

	// Scenario B: Provider supports this model through mapping (exact match)
	modelMapping := p.GetModelMapping()
	if modelMapping != nil {
		if _, exists := modelMapping[modelName]; exists {
			return true
		}

		// Scenario B+: Support through wildcard mapping
		for pattern := range modelMapping {
			if matchWildcard(pattern, modelName) {
				return true
			}
		}
	}

	// Scenario C: Not supported
	return false
}

// GetEffectiveModel gets the actual model name that should be used
// If mapping exists (exact or wildcard), returns mapped model name; otherwise returns original model name
func GetEffectiveModel(p Provider, requestedModel string) string {
	modelMapping := p.GetModelMapping()
	if len(modelMapping) == 0 {
		return requestedModel
	}

	// Look for exact mapping first
	if mappedModel, exists := modelMapping[requestedModel]; exists {
		return mappedModel
	}

	// Look for wildcard mapping
	for pattern, replacement := range modelMapping {
		if matchWildcard(pattern, requestedModel) {
			return applyWildcardMapping(pattern, replacement, requestedModel)
		}
	}

	// No mapping, return original model name
	return requestedModel
}

// ValidateConfiguration validates provider's model configuration
// Returns list of validation errors (empty means validation passed)
func ValidateConfiguration(p Provider) []string {
	errors := make([]string, 0)

	// Rule 1: ModelMapping value must be in SupportedModels
	modelMapping := p.GetModelMapping()
	supportedModels := p.GetSupportedModels()
	if modelMapping != nil && supportedModels != nil {
		for externalModel, internalModel := range modelMapping {
			// Check if it's a wildcard mapping
			if strings.Contains(internalModel, "*") {
				// Wildcard mapping is not validated for now (needs specific request to expand)
				continue
			}

			// Exact mapping needs validation
			supported := false
			for _, supportedModel := range supportedModels {
				if supportedModel == internalModel {
					supported = true
					break
				}
			}

			if !supported {
				// Check wildcard whitelist
				for _, supportedPattern := range supportedModels {
					if matchWildcard(supportedPattern, internalModel) {
						supported = true
						break
					}
				}
			}

			if !supported {
				errors = append(errors, fmt.Sprintf(
					"Invalid model mapping: '%s' -> '%s', target model '%s' not in supportedModels",
					externalModel, internalModel, internalModel,
				))
			}
		}
	}

	// Rule 2: If ModelMapping is configured but SupportedModels is not, give a warning
	if len(modelMapping) > 0 &&
		len(supportedModels) == 0 {
		errors = append(errors,
			"Warning: modelMapping is configured but supportedModels is not, mapped target models cannot be validated",
		)
	}

	// Rule 3: Detect self-mapping (usually meaningless, but not an error)
	if modelMapping != nil {
		for external, internal := range modelMapping {
			if external == internal {
				errors = append(errors, fmt.Sprintf(
					"Warning: model '%s' maps to itself, this is usually meaningless",
					external,
				))
			}
		}
	}

	return errors
}

// matchWildcard wildcard matching function
// Supports * wildcard, e.g., "claude-*" matches "claude-sonnet-4"
func matchWildcard(pattern, text string) bool {
	// If no wildcard, use exact match
	if !strings.Contains(pattern, "*") {
		return pattern == text
	}

	// Simplified implementation: only support single * wildcard
	parts := strings.Split(pattern, "*")
	if len(parts) == 2 {
		// Prefix + * or * + suffix
		prefix, suffix := parts[0], parts[1]
		return strings.HasPrefix(text, prefix) && strings.HasSuffix(text, suffix)
	}

	// Multiple * case (more complex, not supported for now)
	return false
}

// applyWildcardMapping applies wildcard mapping
// Replaces the * matched part from pattern to the * position in replacement
// Example: pattern="claude-*", replacement="anthropic/claude-*", input="claude-sonnet-4"
//
//	Output: "anthropic/claude-sonnet-4"
func applyWildcardMapping(pattern, replacement, input string) string {
	// If pattern or replacement has no wildcard, return replacement directly
	if !strings.Contains(pattern, "*") || !strings.Contains(replacement, "*") {
		return replacement
	}

	// Extract wildcard matched part
	parts := strings.Split(pattern, "*")
	if len(parts) != 2 {
		return replacement // Multiple wildcards not supported
	}

	prefix, suffix := parts[0], parts[1]

	// Verify input actually matches pattern
	if !strings.HasPrefix(input, prefix) || !strings.HasSuffix(input, suffix) {
		return replacement
	}

	// Extract middle part
	wildcardPart := input[len(prefix) : len(input)-len(suffix)]

	// Replace * in replacement
	return strings.Replace(replacement, "*", wildcardPart, 1)
}
