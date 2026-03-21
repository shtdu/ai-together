# Provider Configuration

Managers add and configure AI service providers for each tool that their team will use.

## Purpose

Allow managers to set up AI providers for each tool (Claude, Codex, OpenCode), with validation and testing to ensure they work correctly.

## User Personas

- **Manager** - Team administrator who configures and manages AI providers
- **Member** - Team member who uses configured providers (read-only access to provider list)

## Functional Requirements (EARS Format)

### 1. Provider Management System

**Purpose:** Define the fundamental provider configuration system.

#### Ubiquitous Requirements

- **PC-02-001:** `The system shall support provider configuration for three tool types: Claude, Codex, and OpenCode.`
- **PC-02-002:** `The system shall require at least one enabled provider per tool type.`
- **PC-02-003:** `The system shall assign priority levels to providers (lower number = higher priority).`
- **PC-02-004:** `The system shall prevent duplicate provider names within an organization.`
- **PC-02-005:** `The system shall never display API keys after initial entry.`

---

### 2. Adding a Provider

**User Story:** As a manager, I want to add a new provider.

#### Event-Driven Requirements (Provider Creation Workflow)

- **PC-02-101:** `When a manager selects a tool type, the system shall display the provider configuration form for that tool.`
- **PC-02-102:** `When a manager clicks "Add Provider", the system shall display the provider creation form.`
- **PC-02-103:** `When a manager enters a provider name, the system shall validate the name is unique within the organization.`
- **PC-02-104:** `When a manager enters an API endpoint URL, the system shall validate the URL format.`
- **PC-02-105:** `When a manager enters an API key, the system shall validate the API key format for the tool type.`
- **PC-02-106:** `When a manager saves a new provider, the system shall store the provider configuration.`
- **PC-02-107:** `When a manager saves a new provider, the system shall set the provider state to Enabled by default.`
- **PC-02-108:** `When a manager configures model mappings, the system shall associate the mappings with the provider.`

#### Unwanted Behaviour Requirements (Validation Errors)

- **PC-02-201:** `If a manager enters a duplicate provider name, then the system shall display an error message indicating the provider name already exists.`
- **PC-02-202:** `If a manager enters an invalid API endpoint URL, then the system shall display an error message indicating the URL format is invalid.`
- **PC-02-203:** `If a manager enters an invalid API key format, then the system shall display an error message indicating the API key format is invalid.`
- **PC-02-204:** `If a manager omits required fields, then the system shall prevent saving the provider and display validation errors.`

---

### 3. Testing Provider Connectivity

**User Story:** As a manager, I want to verify a provider works before using it.

#### Event-Driven Requirements (Connection Testing)

- **PC-02-301:** `When a manager clicks "Test" on a provider card, the system shall attempt to connect to the provider endpoint.`
- **PC-02-302:** `When a system tests a provider connection, the system shall validate the API key format.`
- **PC-02-303:** `When a system tests a provider connection, the system shall verify network connectivity to the endpoint.`
- **PC-02-304:** `When a system tests a provider connection, the system shall verify the endpoint is reachable.`
- **PC-02-305:** `When a connection test succeeds, the system shall mark the provider as healthy.`
- **PC-02-306:** `When a connection test fails, the system shall display the failure reason to the manager.`
- **PC-02-307:** `When a system executes a connection test, the system shall not make actual API requests that incur costs.`

#### State-Driven Requirements (Test Results)

- **PC-02-401:** `While a connection test is in progress, the system shall display a loading indicator.`
- **PC-02-402:** `While a connection test completes, the system shall display the test result within 10 seconds.`

---

### 4. Provider State Management

**Purpose:** Define provider states and transitions.

#### Event-Driven Requirements (State Changes)

- **PC-02-501:** `When a manager enables a provider, the system shall set the provider state to Enabled.`
- **PC-02-502:** `When a manager disables a provider, the system shall set the provider state to Disabled.`
- **PC-02-503:** `When a provider experiences 3 consecutive failures within 5 minutes, the system shall mark the provider as Unhealthy.`
- **PC-02-504:** `When an unhealthy provider successfully connects, the system shall mark the provider as healthy.`

#### State-Driven Requirements (Provider Behavior by State)

- **PC-02-601:** `While a provider is in Enabled state, the system shall use the provider for request routing.`
- **PC-02-602:** `While a provider is in Disabled state, the system shall skip the provider in request routing.`
- **PC-02-603:** `While a provider is in Unhealthy state, the system shall temporarily skip the provider in request routing.`
- **PC-02-604:** `While a provider is in Unhealthy state, the system shall automatically retry the provider after 5 minutes.`

#### Unwanted Behaviour Requirements (State Change Restrictions)

- **PC-02-701:** `If a manager attempts to disable the last enabled provider for a tool, then the system shall display an error message indicating at least one provider must be enabled.`

---

### 5. Provider Viewing

**Purpose:** Control who can view provider configurations.

#### Event-Driven Requirements (Manager Access)

- **PC-02-801:** `When a manager requests to view the provider list, the system shall display all configured providers.`
- **PC-02-802:** `When a manager requests to view a provider's configuration, the system shall display the provider details excluding the API key.`

#### Event-Driven Requirements (Member Access)

- **PC-02-803:** `When a member requests to view the provider list, the system shall display the list of provider names and statuses only.`
- **PC-02-804:** `When a member requests to view provider configuration details, then the system shall deny access to sensitive information.`

---

### 6. Provider Editing and Deletion

**Purpose:** Allow managers to modify and remove providers.

#### Event-Driven Requirements (Editing)

- **PC-02-901:** `When a manager edits a provider, the system shall display the provider configuration form with current values.`
- **PC-02-902:** `When a manager updates a provider's API key, the system shall replace the existing API key with the new value.`
- **PC-02-903:** `When a manager saves provider changes, the system shall validate all fields before updating.`

#### Event-Driven Requirements (Deletion)

- **PC-02-904:** `When a manager deletes a provider, the system shall remove the provider configuration.`
- **PC-02-905:** `When a manager deletes a provider, the system shall confirm the deletion action before proceeding.`

#### Unwanted Behaviour Requirements (Deletion Restrictions)

- **PC-02-1001:** `If a manager attempts to delete the only configured provider for a tool, then the system shall display an error message indicating at least one provider must exist.`

## Provider States

| State | Description | Behavior |
|-------|-------------|----------|
| **Enabled** | Active and available | Used for routing |
| **Disabled** | Inactive but configured | Skipped in routing |
| **Unhealthy** | 3+ consecutive failures in past 5 minutes | Temporarily skipped, automatically retried after 5 minutes |

## Tools and Providers

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

## Business Rules

- **BR-02-001:** At least one enabled provider required per tool (Claude, Codex, OpenCode)
- **BR-02-002:** Provider priority 1 is used first (lower number = higher priority)
- **BR-02-003:** Duplicate provider names not allowed within organization
- **BR-02-004:** API keys cannot be retrieved once saved (only replaced)
- **BR-02-005:** Provider becomes unhealthy after 3 consecutive failures

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

**Related:** [02.02 Request Routing](../02_request_routing/) | [02.03 Model Mapping](../03_model_mapping/) | [02.04 Server Relay](../04_server_relay/) | [01.04 Licensing](../../01_identity_and_access/04_licensing/) | [Domain 02 Overview](../README.md)
