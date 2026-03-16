# Provider Management

Managers configure AI service providers. The system routes AI tool requests through these providers automatically, with failover if a provider fails.

## Purpose

Enable teams to configure multiple AI service providers with automatic routing and failover, ensuring reliable access to AI capabilities.

## Subdomains

| Subdomain | Status | Description |
|-----------|--------|-------------|
| **Provider Configuration** | ✅ Implemented | Adding, testing, and managing providers |
| **Request Routing** | ✅ Implemented | How requests are routed to providers |
| **Model Mapping** | ✅ Implemented | Mapping model names to provider-specific models |
| **Provider Health** | 🚧 Planned | Real-time health monitoring and performance metrics |

## Quick Facts

- **Supported AI Tools:** Claude Code, Codex, OpenCode (OpenAI-compatible)
- **Provider Types:** Claude (Anthropic), Codex (custom), OpenCode (OpenAI-compatible)
- **Failover:** Automatic switch to next provider on failure
- **Priority:** Managers control which provider is used first

## Functional Requirements (EARS Format)

### 1. Provider Configuration

**Purpose:** Allow managers to set up AI providers for each tool.

#### Ubiquitous Requirements

- **PM-02-001:** `The system shall support provider configuration for three tool types: Claude, Codex, and OpenCode.`
- **PM-02-002:** `The system shall require at least one enabled provider per tool type.`
- **PM-02-003:** `The system shall validate API credentials before saving provider configuration.`
- **PM-02-004:** `The system shall never display API keys after initial entry.`
- **PM-02-005:** `The system shall prevent duplicate provider names within an organization.`

---

### 2. Request Routing

**Purpose:** Automatically route requests to the best available provider.

#### Ubiquitous Requirements

- **PM-02-101:** `The system shall route requests to the provider with the lowest priority number.`
- **PM-02-102:** `The system shall automatically failover to the next provider on failure.`
- **PM-02-103:** `The system shall attempt maximum 3 providers before returning an error.`
- **PM-02-104:** `The system shall notify users when failover occurs.`
- **PM-02-105:** `The system shall never use disabled providers for routing.`

---

### 3. Model Mapping

**Purpose:** Enable consistent model naming across providers.

#### Ubiquitous Requirements

- **PM-02-201:** `The system shall support model name mappings for each provider.`
- **PM-02-202:** `The system shall apply mappings before routing requests to providers.`
- **PM-02-203:** `The system shall make mappings provider-specific (not global).`
- **PM-02-204:** `The system shall pass through unmapped model names unchanged.`
- **PM-02-205:** `The system shall give exact match mappings precedence over wildcard mappings.`

---

### 4. Provider Health Monitoring

**Purpose:** Track provider availability and performance.

#### Event-Driven Requirements

- **PM-02-301:** `When a provider fails 3 consecutive times, the system shall mark the provider as unhealthy.`
- **PM-02-302:** `When a provider returns HTTP 429 (rate limit), the system shall immediately mark the provider as unhealthy.`
- **PM-02-303:** `When 5 minutes have passed since a provider was marked unhealthy, the system shall automatically clear the unhealthy status.`
- **PM-02-304:** `When a provider recovers (successful request), the system shall mark the provider as healthy.`

#### State-Driven Requirements

- **PM-02-401:** `While a provider is marked unhealthy, the system shall skip the provider in routing.`
- **PM-02-402:** `While a provider is in 5-minute unhealthy cooldown, the system shall not attempt the provider for routing.`

---

### 5. Failover & Reliability

**Purpose:** Ensure uninterrupted service through automatic failover.

#### Event-Driven Requirements

- **PM-02-501:** `When a selected provider fails to respond, the system shall attempt the next provider in the priority list.`
- **PM-02-502:** `When a selected provider times out (30 seconds), the system shall mark the attempt as failed and try the next provider.`
- **PM-02-503:** `When failover occurs, the system shall notify the user with the provider names involved.`
- **PM-02-504:** `When all providers fail after 3 attempts, the system shall return an error indicating all providers are unavailable.`

---

## User Stories

- As a **manager**, I want to configure multiple providers so that we have backup options
- As a **manager**, I want to test provider connectivity so I know they work
- As a **developer**, I want requests to automatically use backup providers if the primary fails
- As a **manager**, I want to set provider priority so I control which is used first

## Related Documentation

- **Provider Configuration:** [`02.01 Provider Configuration`](01_provider_configuration/) - How to set up providers
- **Request Routing:** [`02.02 Request Routing`](02_request_routing/) - How requests are routed
- **Model Mapping:** [`02.03 Model Mapping`](03_model_mapping/) - How model names are mapped
- **Provider Health:** [`02.04 Provider Health`](04_provider_health/) - Health monitoring and alerts

---

**Next:** [Configuration Sync](../03_configuration_sync/)
