# Model Mapping

Model mapping allows users to refer to models by generic names (like "claude-sonnet") while the system routes to the provider-specific model (like "claude-3-5-sonnet-20241022").

## Purpose

Enable consistent model naming across providers, simplifying configuration and improving user experience.

## User Personas

- **Manager** - Team administrator who configures model mappings for the team
- **Member** - Team member who uses simple model names instead of provider-specific versions

## User Stories

- As a **manager**, I want to map model names so my team can use simple names
- As a **member**, I want to use short model names like "claude-sonnet" instead of full version strings
- As a **manager**, I want to update model mappings without changing every user's configuration

## What is Model Mapping?

Model mapping translates user-friendly model names to provider-specific model names.

**Example:**
```
User writes: "claude-sonnet"
Provider needs: "claude-3-5-sonnet-20241022"
Mapping handles the translation
```

## How It Works

### Mapping Format

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

**Example Resolution:**
```
User requests: "claude-sonnet"
Mappings: "claude-sonnet" → "claude-3-5-sonnet", "claude-*" → "claude-*"
Result: "claude-3-5-sonnet" (exact match wins)

User requests: "claude-opus"
Mappings: "claude-*" → "claude-*"
Result: "claude-opus" (wildcard match)

User requests: "gpt-4"
Mappings: (none defined)
Result: "gpt-4" (passes through unchanged)
```

### Mapping Resolution

1. User's AI tool requests a model (e.g., "claude-sonnet")
2. System checks provider's model mappings
3. If mapping exists, use mapped model name
4. If no mapping, use requested model name as-is

## Use Cases

### Simplified Model Names

**Before mapping:**
- User must remember: "claude-3-5-sonnet-20241022"

**After mapping:**
- User types: "claude-sonnet"
- System routes to: "claude-3-5-sonnet-20241022"

### Provider Updates

**Scenario:** Provider releases new model version

**Without mapping:**
- Every user must update their model references
- High coordination effort

**With mapping:**
- Manager updates one mapping
- All users automatically use new model

### Cross-Provider Consistency

**Scenario:** Using multiple providers with same base models

**Without mapping:**
- Each provider has different model names
- Complex configuration per provider

**With mapping:**
- All providers use same mapped names
- Simple, consistent configuration

## Configuration Workflow

### Setting Up Mappings

**User Story:** As a manager, I want to configure model mappings.

**Workflow:**
1. Manager opens provider configuration
2. Selects "Model Mappings" section
3. Enters mapping: simple name → provider model name
4. Saves configuration
5. Mappings sync to all team members

**Example Entry:**
- Alias: `claude-sonnet`
- Target: `claude-3-5-sonnet-20241022`

### Default Mappings

System provides sensible defaults for common providers:
- `claude-sonnet` → `claude-3-5-sonnet-20241022`
- `claude-haiku` → `claude-3-5-haiku-20241022`
- `gpt-4` → `gpt-4-turbo-preview`

Managers can override defaults as needed.

## Functional Requirements

- **FR-001:** System must apply mappings before routing to provider
- **FR-002:** Mappings must be per-provider (not global)
- **FR-003:** Unmapped model names must pass through unchanged
- **FR-004:** Mapping changes must sync to all members within 5 minutes
- **FR-005:** Exact match mappings take precedence over wildcard mappings
- **FR-006:** If no mappings are defined for a provider, model names pass through unchanged

## Business Rules

- **BR-001:** Mappings are provider-specific (each provider has its own mappings)
- **BR-002:** Mapping keys must be unique within a provider
- **BR-003:** Invalid target models should show error on provider test
- **BR-004:** Exact matches take precedence over wildcard matches
- **BR-005:** Only one wildcard (`*`) allowed per pattern

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
