# Request Routing

When AI tools make requests, the system automatically routes them to the best available provider. If a provider fails, it automatically tries the next one.

## Purpose

Ensure AI tool requests are reliably delivered to working providers, with automatic failover for uninterrupted service.

## User Personas

- **Member** - Uses AI tools, expects requests to work without manual intervention
- **Manager** - Monitors provider health and adjusts routing priorities

## Functional Requirements (EARS Format)

### 1. Provider Selection

**Purpose:** Define how the system selects providers for routing.

#### Event-Driven Requirements (Selection Process)

- **RR-02-101:** `When an AI tool makes a request, the system shall identify which tool type is being used.`
- **RR-02-102:** `When an AI tool makes a request, the system shall filter to enabled and healthy providers for that tool.`
- **RR-02-103:** `When an AI tool makes a request, the system shall sort providers by priority (lowest number first).`
- **RR-02-104:** `When an AI tool makes a request, the system shall select the provider with the highest priority (lowest number).`
- **RR-02-105:** `When an AI tool makes a request, the system shall forward the request to the selected provider.`
- **RR-02-106:** `When a selected provider returns a successful response, the system shall return the response to the AI tool.`

#### Ubiquitous Requirements

- **RR-02-001:** `The system shall route requests to the provider with the lowest priority number.`
- **RR-02-002:** `The system shall never use disabled providers for routing.`

---

### 2. Automatic Failover

**Purpose:** Define how the system handles provider failures.

#### Event-Driven Requirements (Failover Trigger)

- **RR-02-201:** `When a selected provider fails to respond, the system shall attempt the next provider in the priority list.`
- **RR-02-202:** `When a selected provider times out (30 seconds), the system shall mark the attempt as failed.`
- **RR-02-203:** `When a selected provider returns an HTTP error (4xx or 5xx), the system shall mark the attempt as failed.`
- **RR-02-204:** `When a selected provider returns HTTP 429 (Too Many Requests), the system shall immediately mark the provider as unhealthy.`
- **RR-02-205:** `When a selected provider experiences a network error, the system shall mark the attempt as failed.`
- **RR-02-206:** `When failover occurs, the system shall notify the user with the message "Provider A unavailable, using Provider B".`
- **RR-02-207:** `When a backup provider successfully handles a request, the system shall return the response to the AI tool.`

#### State-Driven Requirements (Failover Limits)

- **RR-02-301:** `While the system has attempted 3 providers without success, the system shall stop attempting additional providers.`
- **RR-02-302:** `While the system has attempted 3 providers without success, the system shall return an error to the AI tool.`
- **RR-02-303:** `While the system has attempted 3 providers without success, the system shall display the error message "All AI providers unavailable. Please contact your manager."`
- **RR-02-304:** `While the system has attempted 3 providers without success, the system shall display which providers were tried and why they failed.`

---

### 3. Provider Health Management

**Purpose:** Define how the system tracks and manages provider health.

#### Event-Driven Requirements (Health Tracking)

- **RR-02-401:** `When a provider fails 3 consecutive times, the system shall mark the provider as unhealthy.`
- **RR-02-402:** `When a provider returns HTTP 429, the system shall mark the provider as unhealthy immediately.`
- **RR-02-403:** `When 5 minutes have passed since a provider was marked unhealthy, the system shall automatically clear the unhealthy status.`
- **RR-02-404:** `When a manager tests an unhealthy provider and the test succeeds, the system shall immediately mark the provider as healthy.`

#### State-Driven Requirements (Unhealthy Provider Behavior)

- **RR-02-501:** `While a provider is marked unhealthy, the system shall skip the provider in routing.`
- **RR-02-502:** `While a provider is marked unhealthy, the system shall not attempt the provider for routing.`
- **RR-02-503:** `While a provider is in 5-minute unhealthy cooldown, the system shall not reset the cooldown early.`

#### Complex Requirements (Health Status Recovery)

- **RR-02-601:** `When an unhealthy provider's status is cleared after 5 minutes, if the next request succeeds, then the system shall mark the provider as healthy.`
- **RR-02-602:** `When an unhealthy provider's status is cleared after 5 minutes, if the next request fails, then the system shall reset the failure count and require 3 consecutive failures to mark unhealthy again.`

---

### 4. Provider Priority Management

**Purpose:** Control how managers adjust provider routing order.

#### Event-Driven Requirements (Priority Changes)

- **RR-02-701:** `When a manager changes a provider's priority, the system shall update the provider's priority value.`
- **RR-02-702:** `When a manager changes a provider's priority, the system shall apply the change to the next request (not the current one).`

#### Event-Driven Requirements (Priority Validation)

- **RR-02-703:** `When a manager sets provider priorities, the system shall ensure priority numbers are unique within a tool type.`
- **RR-02-704:** `When a manager sets provider priorities, the system shall validate that at least one provider has priority set.`

---

### 5. Error Handling & User Notifications

**Purpose:** Define how the system communicates routing issues to users.

#### Event-Driven Requirements (Member Notifications)

- **RR-02-801:** `When failover occurs, the system shall display a notification to the member: "Switched to backup provider (Primary unavailable)".`
- **RR-02-802:** `When all providers fail, the system shall display an error to the member: "All AI providers unavailable. Please contact your manager."`
- **RR-02-803:** `When all providers fail, the system shall show which providers were tried and why they failed.`

#### Event-Driven Requirements (Manager Notifications)

- **RR-02-804:** `When a provider becomes unhealthy, the system shall update the provider status indicator in the manager dashboard.`
- **RR-02-805:** `When a provider's health status changes, the system shall display the current status (Green: Healthy, Red: Unhealthy).`

#### Unwanted Behaviour Requirements (No Providers Available)

- **RR-02-901:** `If an AI tool makes a request and all providers are disabled, then the system shall display the error "No enabled providers. Contact your manager."`
- **RR-02-902:** `If an AI tool makes a request and all providers are unhealthy, then the system shall attempt the providers in priority order.`

---

### 6. Performance & Timeouts

**Purpose:** Define timeout and performance behavior.

#### State-Driven Requirements (Timeout Behavior)

- **RR-02-1001:** `While a provider request is in progress, if 30 seconds elapse without a response, then the system shall mark the request as timed out.`
- **RR-02-1002:** `While a provider request times out, the system shall count the timeout as 1 failure toward the unhealthy threshold.`

#### Complex Requirements (Slow Response Handling)

- **RR-02-1101:** `When a provider responds slowly (approaching 30 seconds), if the response completes within the timeout, then the system shall return the response to the AI tool.`

## How Routing Works

### Selection Process

1. AI tool makes request (e.g., Claude Code makes a request)
2. System identifies which tool is being used (Claude, Codex, or OpenCode)
3. System filters to enabled and healthy providers for that tool
4. System sorts providers by priority (1 = highest)
5. System selects first provider
6. System forwards request to selected provider
7. If successful: return response to AI tool
8. If failed: mark unhealthy (if applicable), try next provider (failover)

### What is "Failure"?

A provider is considered to have failed if:
- Connection times out (30 seconds)
- Provider returns HTTP error (4xx or 5xx), including 429 (Too Many Requests)
- Network error occurs
- Provider takes too long to respond

**Rate Limiting (HTTP 429):**
- A single 429 response immediately marks provider as unhealthy
- Unhealthy status persists for 5 minutes (cooldown period)
- This prevents overwhelming a rate-limited provider with repeated requests

## Business Rules

- **BR-02-001:** Provider priority 1 is used first (lower number = higher priority)
- **BR-02-002:** Disabled providers are never used for routing
- **BR-02-003:** At least one provider must be enabled per tool
- **BR-02-004:** Failover attempts stop after 3 providers
- **BR-02-005:** Provider becomes unhealthy after 3 consecutive failures (status clears after 5 minutes)
- **BR-02-006:** Single HTTP 429 response immediately marks provider as unhealthy (5 minute cooldown)

## Edge Cases & Error Handling

| Scenario | System Behavior |
|----------|----------------|
| All providers disabled | Error: "No enabled providers. Contact your manager." |
| Single provider fails | Failover to next provider (if available) |
| All 3 attempted providers fail | Error: "All providers unavailable. Tried: [Provider A, B, C]" |
| Provider returns 429 (rate limited) | Immediately marked unhealthy, 5-minute cooldown |
| Provider becomes unhealthy (3+ other failures) | Marked unhealthy, status clears after 5 minutes |
| Unhealthy status clears (after 5 min) | Provider available for routing on next request |
| Next request succeeds after status clear | Provider remains healthy |
| Next request fails after status clear | Failure count resets, needs 3 consecutive to mark unhealthy again |
| Manager tests unhealthy provider | If successful, immediately marked healthy |
| Provider responds slowly/times out | Counts as 1 failure toward unhealthy threshold (3 consecutive = unhealthy) |
| Priority changed mid-request | Change applies to next request, not current one |

## Success Criteria

- Requests use correct provider based on priority
- Failover happens within 5 seconds of failure
- User notifications are clear and actionable
- Unhealthy status automatically clears after 5 minutes (provider available for routing)
- At least one provider is always available for configured tools

---

**Related:** [02.01 Provider Configuration](../01_provider_configuration/) | [02.03 Model Mapping](../03_model_mapping/) | [02.04 Server Relay](../04_server_relay/) | [06.03 Performance](../../06_system_behaviors/03_performance/) | [Domain 02 Overview](../README.md)
