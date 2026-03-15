# Provider Configuration

Managers add and configure AI service providers for each tool that their team will use.

## Purpose

Allow managers to set up AI providers for each tool (Claude, Codex, OpenCode), with validation and testing to ensure they work correctly.

## User Personas

- **Manager** - Team administrator who configures and manages AI providers
- **Member** - Team member who uses configured providers (read-only access to provider list)

## User Stories

- As a **manager**, I want to add a provider by entering API credentials so my team can use it
- As a **manager**, I want to test provider connectivity so I know the credentials are correct
- As a **manager**, I want to enable/disable providers without deleting them
- As a **manager**, I want to see provider status so I know if they're healthy

## Tools and Providers

Each tool has its own list of configured providers. All tools work identically - the distinction is only organizational.

| Tool | Description |
|------|-------------|
| **Claude** | AI providers for Claude Code |
| **Codex** | AI providers for Codex |
| **OpenCode** | AI providers for OpenCode and OpenAI-compatible tools |

**Provider Configuration (same for all tools):**
- Provider name (e.g., "Anthropic", "Team Custom", "Backup Provider")
- API endpoint URL (e.g., https://api.anthropic.com)
- API key (e.g., sk-ant-xxx, sk-xxx)
- Optional: Model mappings (see [02.03 Model Mapping](../03_model_mapping/))

## Configuration Workflow

### Adding a Provider

**User Story:** As a manager, I want to add a new provider.

**Workflow:**
1. Manager selects a tool (Claude, Codex, or OpenCode)
2. Manager clicks "Add Provider"
3. Enters provider name
4. Enters API endpoint URL
5. Enters API key
6. Optionally configures model mappings
7. Sets priority level
8. Clicks "Test" to verify connectivity
9. Saves provider

**Functional Requirements:**
- **FR-001:** System must validate required fields before saving
- **FR-002:** Test connection must not make actual API calls that incur costs
- **FR-003:** System must show test result (success/failure) within 10 seconds
- **FR-004:** API keys must never be displayed after initial entry

### Testing a Provider

**User Story:** As a manager, I want to verify a provider works before using it.

**Workflow:**
1. Manager clicks "Test" on provider card
2. System attempts connection to provider
3. System displays result: success or failure with reason
4. If successful, provider marked as healthy

**Test Behavior:**
- Validates API key format
- Checks network connectivity
- Verifies endpoint is reachable
- Does NOT make actual API requests (no cost incurred)

### Provider States

| State | Description | Behavior |
|-------|-------------|----------|
| **Enabled** | Active and available | Used for routing |
| **Disabled** | Inactive but configured | Skipped in routing |
| **Unhealthy** | 3+ consecutive failures in past 5 minutes | Temporarily skipped, automatically retried after 5 minutes |

## Business Rules

- **BR-001:** At least one enabled provider required per tool (Claude, Codex, OpenCode)
- **BR-002:** Provider priority 1 is used first (lower number = higher priority)
- **BR-003:** Duplicate provider names not allowed within organization
- **BR-004:** API keys cannot be retrieved once saved (only replaced)
- **BR-005:** Provider becomes unhealthy after 3 consecutive failures

## Edge Cases & Error Handling

| Scenario | System Behavior |
|----------|----------------|
| Test connection fails | Show error: "Connection failed: [reason]" |
| Invalid API key format | Show error: "Invalid API key format" |
| Trying to disable last enabled provider for a tool | Show error: "At least one provider must be enabled for this tool" |
| Duplicate provider name | Show error: "Provider with this name already exists" |
| Provider becomes unhealthy (3+ failures) | Automatically marked unhealthy, retried after 5 minutes |
| Provider recovers (successful connection) | Automatically marked healthy again |

## Success Criteria

- Managers can add a provider in under 2 minutes
- Test connection accurately reflects provider availability
- API keys are never exposed after entry
- Provider status updates in real-time
- At least one provider is always available for each tool

---

**Related:** [02.02 Request Routing](../02_request_routing/) | [02.03 Model Mapping](../03_model_mapping/) | [01.04 Licensing](../../01_identity_and_access/04_licensing/) | [Domain 02 Overview](../README.md)

