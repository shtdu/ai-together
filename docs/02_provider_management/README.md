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
