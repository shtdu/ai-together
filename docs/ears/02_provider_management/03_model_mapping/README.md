# Model Mapping

Model mapping allows users to refer to models by generic names (like "claude-sonnet") while the system routes to the provider-specific model (like "claude-3-5-sonnet-20241022").

## Purpose

Enable consistent model naming across providers, simplifying configuration and improving user experience.

## User Personas

- **Manager** - Team administrator who configures model mappings for the team
- **Member** - Team member who uses simple model names instead of provider-specific versions

## Functional Requirements (EARS Format)

### 1. Model Mapping System

**Purpose:** Define the core model mapping functionality.

#### Ubiquitous Requirements

- **MM-02-001:** `The system shall support model name mappings for each provider.`
- **MM-02-002:** `The system shall apply mappings before routing requests to providers.`
- **MM-02-003:** `The system shall make mappings provider-specific (not global).`
- **MM-02-004:** `The system shall pass through unmapped model names unchanged.`
- **MM-02-005:** `The system shall give exact match mappings precedence over wildcard mappings.`

---

### 2. Mapping Configuration

**User Story:** As a manager, I want to configure model mappings.

#### Event-Driven Requirements (Creating Mappings)

- **MM-02-101:** `When a manager opens provider configuration, the system shall display the "Model Mappings" section.`
- **MM-02-102:** `When a manager enters a mapping alias, the system shall validate the alias is unique within the provider.`
- **MM-02-103:** `When a manager enters a target model name, the system shall validate the model name format.`
- **MM-02-104:** `When a manager saves a model mapping, the system shall store the mapping for the provider.`
- **MM-02-105:** `When a manager saves model mappings, the system shall sync the mappings to all team members within 5 minutes.`

#### Event-Driven Requirements (Wildcard Mappings)

- **MM-02-106:** `When a manager enters a wildcard pattern (containing *), the system shall validate the pattern contains only one wildcard.`
- **MM-02-107:** `When a manager enters a wildcard pattern, the system shall validate the wildcard is not the only character.`
- **MM-02-108:** `When a manager saves a wildcard mapping, the system shall store the pattern for matching.`

#### Unwanted Behaviour Requirements (Configuration Errors)

- **MM-02-201:** `If a manager enters a duplicate mapping alias, then the system shall display a warning and use the last mapping entered.`
- **MM-02-202:** `If a manager enters an invalid target model name, then the system shall display a validation error.`
- **MM-02-203:** `If a manager enters a wildcard pattern with multiple wildcards, then the system shall display an error indicating only one wildcard is allowed.`

---

### 3. Mapping Resolution

**Purpose:** Define how the system resolves model names using mappings.

#### Event-Driven Requirements (Exact Match Resolution)

- **MM-02-301:** `When a user requests a model name, the system shall check for an exact match in the provider's mappings.`
- **MM-02-302:** `When a user requests a model name that exactly matches a mapping alias, the system shall route the request using the target model name.`
- **MM-02-303:** `When a user requests a model name that exactly matches a mapping alias, the system shall not check wildcard patterns.`

#### Event-Driven Requirements (Wildcard Match Resolution)

- **MM-02-304:** `When a user requests a model name with no exact match, the system shall check for wildcard pattern matches.`
- **MM-02-305:** `When a user requests a model name that matches a wildcard pattern, the system shall route the request using the wildcard target pattern.`
- **MM-02-306:** `When multiple wildcard patterns match a requested model name, the system shall use the most specific pattern.`

#### Event-Driven Requirements (Passthrough Behavior)

- **MM-02-307:** `When a user requests a model name and no mappings exist for the provider, the system shall pass the model name through unchanged.`
- **MM-02-308:** `When a user requests a model name and no mapping matches (exact or wildcard), the system shall pass the model name through unchanged.`

#### Complex Requirements (Mapping Resolution Logic)

- **MM-02-401:** `When a user requests a model name, if an exact match exists, then the system shall use the exact match target without evaluating wildcard patterns.`
- **MM-02-402:** `When a user requests a model name, if no exact match exists and a wildcard pattern matches, then the system shall apply the wildcard mapping.`
- **MM-02-403:** `When a user requests a model name, if no mappings match (exact or wildcard), then the system shall pass the model name through unchanged.`

---

### 4. Default Mappings

**Purpose:** Provide sensible defaults for common providers.

#### Ubiquitous Requirements

- **MM-02-501:** `The system shall provide default mappings for common Claude models.`
- **MM-02-502:** `The system shall provide default mappings for common GPT models.`
- **MM-02-503:** `The system shall allow managers to override default mappings.`

#### Event-Driven Requirements (Default Mapping Application)

- **MM-02-504:** `When a provider is created with no custom mappings, the system shall apply default mappings if available.`
- **MM-02-505:** `When a manager adds a custom mapping that conflicts with a default, then the system shall use the custom mapping.`

---

### 5. Mapping Synchronization

**Purpose:** Ensure mapping changes propagate to all team members.

#### Event-Driven Requirements (Sync Workflow)

- **MM-02-601:** `When a manager saves mapping changes, the system shall push the updated mappings to all team members.`
- **MM-02-602:** `When a member receives updated mappings, the system shall apply the new mappings to subsequent requests.`

#### State-Driven Requirements (Sync Timing)

- **MM-02-603:** `While mapping synchronization is in progress, the system shall complete the sync within 5 minutes.`

---

### 6. Provider Testing with Mappings

**Purpose:** Validate mappings work correctly with provider endpoints.

#### Event-Driven Requirements (Testing Validation)

- **MM-02-701:** `When a manager tests a provider with custom mappings, the system shall validate the target model names are supported.`
- **MM-02-702:** `When a provider test indicates a mapped model is not supported, then the system shall display an error message indicating the model is not supported by the provider.`
- **MM-02-703:** `When a provider test succeeds, the system shall mark the provider and its mappings as valid.`

#### Unwanted Behaviour Requirements (Test Failures)

- **MM-02-801:** `If a mapped model is not supported by the provider, then the system shall display the error "Model [name] not supported by provider".`

## What is Model Mapping?

Model mapping translates user-friendly model names to provider-specific model names.

**Example:**
```
User writes: "claude-sonnet"
Provider needs: "claude-3-5-sonnet-20241022"
Mapping handles the translation
```

## Mapping Format

Managers define mappings as simple name → provider model name pairs:

| Simple Name (Alias) | Provider Model Name |
|-------------------|---------------------|
| claude-sonnet | claude-3-5-sonnet-20241022 |
| claude-haiku | claude-3-5-haiku-20241022 |
| gpt-4 | gpt-4-turbo-preview |

### Wildcard Mappings

Managers can define wildcard mappings using `*` to match multiple model names:

| Alias Pattern | Provider Model Pattern | Description |
|--------------|----------------------|-------------|
| claude-* | claude-* | Match all Claude models |
| gpt-* | gpt-*-preview | Match all GPT models with preview suffix |

**Wildcard Rules:**
- Exact matches take precedence over wildcards
- More specific patterns match before generic ones
- Only one `*` supported per pattern
- If no mappings defined at all, model names pass through unchanged

## Business Rules

- **BR-02-001:** Mappings are provider-specific (each provider has its own mappings)
- **BR-02-002:** Mapping keys must be unique within a provider
- **BR-02-003:** Invalid target models should show error on provider test
- **BR-02-004:** Exact matches take precedence over wildcard matches
- **BR-02-005:** Only one wildcard (`*`) allowed per pattern

## Edge Cases & Error Handling

| Scenario | System Behavior |
|----------|----------------|
| No mappings defined for provider | Model names pass through unchanged |
| Unmapped model name | Pass through to provider unchanged |
| Invalid target model | Provider returns error (model not found) |
| Duplicate mapping keys | Warning shown to manager, last mapping wins |
| Provider doesn't support mapped model | Error: "Model [name] not supported by provider" |
| Wildcard and exact match both exist | Exact match takes precedence |
| Multiple wildcards match | Most specific pattern wins (e.g., `claude-*` before `gpt-*`) |

## Success Criteria

- Mappings correctly translate model names
- Wildcard mappings match patterns correctly
- Exact matches take precedence over wildcards
- Unmapped names pass through correctly
- Mapping updates sync within 5 minutes
- Default mappings work for common providers
- Invalid mappings are caught during provider testing

---

**Related:** [02.01 Provider Configuration](../01_provider_configuration/) | [02.02 Request Routing](../02_request_routing/) | [03.01 Configuration Distribution](../../03_configuration_sync/01_distribution/)
