# Request Routing

When AI tools make requests, the system automatically routes them to the best available provider. If a provider fails, it automatically tries the next one.

## Purpose

Ensure AI tool requests are reliably delivered to working providers, with automatic failover for uninterrupted service.

## User Personas

- **Member** - Uses AI tools, expects requests to work without manual intervention
- **Manager** - Monitors provider health and adjusts routing priorities

## User Stories

- As a **member**, I want my requests to use our primary provider when it's available
- As a **member**, I want automatic failover if the primary provider fails
- As a **member**, I want to be notified when failover happens so I'm aware of issues
- As a **manager**, I want to control which provider is used first (priority)

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

### Failover Behavior

**Automatic Failover:**
- System tries next provider in priority list
- Maximum 3 providers attempted
- User sees notification: "Provider A unavailable, using Provider B"
- Most errors: provider marked as unhealthy after 3 consecutive failures
- HTTP 429 (rate limit): immediately marked as unhealthy (single failure)
- After 5 minutes, unhealthy status is automatically cleared (provider available for routing again)

**All Providers Failed:**
- User sees error: "All AI providers unavailable. Please contact your manager."
- Error shows which providers were tried and why they failed
- Suggestion: "Check your provider configurations or contact support"

## User Experience

### Member Experience

**Normal Operation:**
- Member uses AI tool normally
- Requests complete successfully
- Member unaware of which provider was used

**Failover Occurs:**
- Notification appears: "Switched to backup provider (Primary unavailable)"
- Request still completes successfully
- Member can continue working

**All Providers Down:**
- AI tool shows error
- Desktop app shows provider status (all red/unhealthy)
- Member notified to contact manager

### Manager Experience

**Monitoring Provider Health:**
- Manager dashboard shows provider status indicators
- Green: Healthy (available for routing)
- Red: Unhealthy (3+ consecutive failures OR single 429 response, skipped until status clears)

**Responding to Issues:**
- Manager can test provider connectivity (immediately marks healthy if successful)
- Manager can disable problematic provider
- Manager can adjust provider priority
- Manager can add backup provider

**Automatic Status Recovery:**
- After 5 minutes, unhealthy status is automatically cleared
- Provider becomes available for routing on next client request
- If request succeeds, provider stays healthy; if fails, marked unhealthy again

## Functional Requirements

- **FR-001:** System must route requests to provider with lowest priority number
- **FR-002:** System must automatically failover to next provider on failure
- **FR-003:** System must attempt maximum 3 providers before giving up
- **FR-004:** System must notify user when failover occurs
- **FR-005:** System must mark failed providers as unhealthy for 5 minutes (after 3 consecutive failures)
- **FR-006:** System must immediately mark provider as unhealthy on single HTTP 429 response
- **FR-007:** System must automatically clear unhealthy status after 5 minutes (making provider available for routing)

## Business Rules

- **BR-001:** Provider priority 1 is used first (lower number = higher priority)
- **BR-002:** Disabled providers are never used for routing
- **BR-003:** At least one provider must be enabled per tool
- **BR-004:** Failover attempts stop after 3 providers
- **BR-005:** Provider becomes unhealthy after 3 consecutive failures (status clears after 5 minutes)
- **BR-006:** Single HTTP 429 response immediately marks provider as unhealthy (5 minute cooldown)

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

**Related:** [02.01 Provider Configuration](../01_provider_configuration/) | [02.03 Model Mapping](../03_model_mapping/) | [06.03 Performance](../../06_system_behaviors/03_performance/) | [Domain 02 Overview](../README.md)
