# Server Relay

The server relay is a centralized HTTP proxy that forwards AI tool requests to configured providers with automatic failover, usage tracking, and multi-tenant isolation.

## Purpose

Enable AI tools to send requests through a central server that handles provider routing, failover, model mapping, and usage tracking for team analytics and cost management.

## Relay vs Proxy Distinction

| Aspect | Member Proxy | Server Relay |
|--------|--------------|--------------|
| **Location** | Runs locally on member machine | Runs on central server |
| **Port** | 18100 (localhost) | 9080 (server port) |
| **Network** | Local-only | Requires network connection |
| **Offline Mode** | Works offline with cached config | Requires server connection |
| **Usage Tracking** | Batches uploads to server | Tracks in real-time |
| **Multi-tenancy** | Single user | Isolates by tenant_id |
| **Configuration** | Receives from server | Source of truth |

## User Personas

- **Member** - Team member whose AI tool requests are relayed through the server
- **Manager** - Team administrator who views relayed usage analytics and provider performance

## Functional Requirements (EARS Format)

### 1. Relay API Endpoints

**Purpose:** Define the HTTP endpoints that accept AI tool requests for relaying.

#### Ubiquitous Requirements

- **SR-02-001:** `The system shall provide relay endpoints for three tool types: claude, codex, and opencode.`
- **SR-02-002:** `The system shall require JWT authentication for all relay endpoints.`
- **SR-02-003:** `The system shall route requests based on the tool parameter in the URL path.`
- **SR-02-004:** `The system shall extract the user's tenant_id from the JWT token for multi-tenant isolation.`

#### Event-Driven Requirements (Anthropic-Style Messages)

- **SR-02-101:** `When a client sends POST to /api/v1/relay/:tool/v1/messages, the system shall validate the tool parameter.`
- **SR-02-102:** `When a client sends POST to /api/v1/relay/:tool/v1/messages with an invalid tool, the system shall return HTTP 400 with error message "invalid tool kind".`
- **SR-02-103:** `When a client sends POST to /api/v1/relay/:tool/v1/messages with a valid tool, the system shall forward the request to configured providers.`
- **SR-02-104:** `When a client sends POST to /api/v1/relay/claude/v1/messages, the system shall use Anthropic-style message format.`
- **SR-02-105:** `When a client sends POST to /api/v1/relay/codex/v1/messages, the system shall use Codex message format.`
- **SR-02-106:** `When a client sends POST to /api/v1/relay/opencode/v1/messages, the system shall use OpenCode message format.`

#### Event-Driven Requirements (OpenAI-Style Chat Completions)

- **SR-02-107:** `When a client sends POST to /api/v1/relay/:tool/v1/chat/completions, the system shall validate the tool parameter.`
- **SR-02-108:** `When a client sends POST to /api/v1/relay/:tool/v1/chat/completions with an invalid tool, the system shall return HTTP 400 with error message "invalid tool kind".`
- **SR-02-109:** `When a client sends POST to /api/v1/relay/:tool/v1/chat/completions with a valid tool, the system shall forward the request to configured providers.`
- **SR-02-110:** `When a client sends POST to /api/v1/relay/codex/v1/chat/completions, the system shall use OpenAI-style chat completions format.`
- **SR-02-111:** `When a client sends POST to /api/v1/relay/opencode/v1/chat/completions, the system shall use OpenAI-style chat completions format.`

#### Unwanted Behaviour Requirements (Invalid Requests)

- **SR-02-201:** `If a client sends a request without JWT authentication, then the system shall return HTTP 401 Unauthorized.`
- **SR-02-202:** `If a client sends a request with an expired JWT token, then the system shall return HTTP 401 Unauthorized.`
- **SR-02-203:** `If a client sends a request for a tool with no configured providers, then the system shall return HTTP 404 with error "no providers available".`
- **SR-02-204:** `If a client sends a request for a model that no provider supports, then the system shall return HTTP 404 with error "no available provider supports model '<model>'".`

---

### 2. Request Forwarding with Failover

**Purpose:** Define how the relay forwards requests to providers with automatic failover.

#### Event-Driven Requirements (Provider Selection)

- **SR-02-301:** `When a relay request arrives, the system shall load all providers for the user's tenant_id.`
- **SR-02-302:** `When a relay request arrives, the system shall filter providers by tool kind (claude, codex, opencode).`
- **SR-02-303:** `When a relay request arrives, the system shall filter to only enabled providers.`
- **SR-02-304:** `When a relay request arrives, the system shall validate provider configuration (API URL and API key present).`
- **SR-02-305:** `When a relay request arrives, the system shall validate provider configuration using provider-specific validation rules.`
- **SR-02-306:** `When a relay request specifies a model, the system shall filter to providers that support the requested model.`
- **SR-02-307:** `When a relay request arrives, the system shall sort filtered providers by level (lower number = higher priority).`

#### Event-Driven Requirements (Request Forwarding)

- **SR-02-308:** `When a relay request arrives, the system shall attempt providers in priority order.`
- **SR-02-309:** `When a provider successfully responds, the system shall return the response to the client.`
- **SR-02-310:** `When a provider fails, the system shall attempt the next provider in the list.`
- **SR-02-311:** `When all providers fail, the system shall return HTTP 400 with error message indicating all providers failed.`

#### State-Driven Requirements (Failover Limits)

- **SR-02-401:** `While attempting providers, if all providers have been exhausted, then the system shall stop attempting additional providers.`

---

### 3. Model Mapping Transformation

**Purpose:** Define how the relay transforms model names using provider-specific mappings.

#### Event-Driven Requirements (Model Name Mapping)

- **SR-02-501:** `When a provider has model mappings configured, the system shall check if the requested model matches a mapping key.`
- **SR-02-502:** `When a requested model matches a mapping key, the system shall replace the model name in the request body with the mapped target model.`
- **SR-02-503:** `When a provider has no model mappings, the system shall pass the requested model name unchanged.`

#### Ubiquitous Requirements

- **SR-02-504:** `The system shall apply model mapping transformations before forwarding the request to the provider.`

---

### 4. Streaming Response Handling

**Purpose:** Define how the relay handles streaming responses from providers.

#### Event-Driven Requirements (Stream Detection)

- **SR-02-601:** `When a request contains "stream": true, the system shall detect streaming mode.`
- **SR-02-602:** `When a provider returns a streaming response, the system shall forward streaming chunks to the client.`

#### State-Driven Requirements (Stream Forwarding)

- **SR-02-603:** `While a streaming response is in progress, the system shall forward each chunk to the client as it arrives.`

#### Ubiquitous Requirements

- **SR-02-604:** `The system shall preserve streaming format when forwarding responses.`

---

### 5. Token Usage Extraction

**Purpose:** Define how the relay extracts token usage from streaming responses for tracking.

#### Event-Driven Requirements (Usage Parsing)

- **SR-02-701:** `When a provider returns a streaming response, the system shall parse token usage from the response.`
- **SR-02-702:** `When streaming chunks contain token usage data, the system shall extract input_tokens, output_tokens, cache_create_tokens, cache_read_tokens, and reasoning_tokens.`
- **SR-02-703:** `When token usage is extracted, the system shall record the usage with the tenant_id, user_id, provider, model, and token counts.`

#### Ubiquitous Requirements

- **SR-02-704:** `The system shall use tool-specific parsers for token extraction (Claude, Codex, OpenCode formats).`

---

### 6. Usage Tracking

**Purpose:** Define how the relay tracks usage for analytics and cost management.

#### Event-Driven Requirements (Usage Recording)

- **SR-02-801:** `When a relay request completes, the system shall record usage data including platform, model, provider, token counts, duration, tenant_id, and user_id.`
- **SR-02-802:** `When usage recording fails, the system shall log the error but not affect the response to the client.`

#### Ubiquitous Requirements

- **SR-02-803:** `The system shall extract user_id from the authenticated request context for usage tracking.`
- **SR-02-804:** `The system shall extract tenant_id from the authenticated user for multi-tenant isolation.`

---

### 7. Request Logging

**Purpose:** Define how the relay logs requests for debugging and monitoring.

#### Event-Driven Requirements (Request Logging)

- **SR-02-901:** `When a relay request arrives, the system shall log request_id, tool, endpoint, model, is_stream, client_ip, and tenant_id.`
- **SR-02-902:** `When a relay request arrives, the system shall log the available providers and count.`
- **SR-02-903:** `When attempting a provider, the system shall log the provider name and attempt number.`
- **SR-02-904:** `When a provider succeeds, the system shall log the provider name, duration, and HTTP status.`
- **SR-02-905:** `When a provider fails, the system shall log the provider name, error message, and duration.`
- **SR-02-906:** `When all providers fail, the system shall log the total number of attempts and failure reasons.`

#### Ubiquitous Requirements

- **SR-02-907:** `The system shall generate a unique request_id using Unix nanoseconds for each relay request.`

---

### 8. Header and Query Parameter Handling

**Purpose:** Define how the relay handles HTTP headers and query parameters.

#### Ubiquitous Requirements

- **SR-02-1001:** `The system shall forward request headers to the provider, except for Authorization which is replaced with the provider's API key.`
- **SR-02-1002:** `The system shall set the Authorization header to "Bearer <provider_api_key>" when forwarding to the provider.`
- **SR-02-1003:** `The system shall set the Accept header to "application/json" if not already present.`
- **SR-02-1004:** `The system shall forward query parameters to the provider.`

---

### 9. Error Handling

**Purpose:** Define how the relay handles various error conditions.

#### Event-Driven Requirements (Provider Errors)

- **SR-02-1101:** `When a provider returns HTTP 4xx or 5xx error, the system shall consider the provider failed and attempt the next provider.`
- **SR-02-1102:** `When a provider returns HTTP 429 (rate limit), the system shall mark the provider as failed and attempt the next provider.`
- **SR-02-1103:** `When a provider connection times out or has a network error, the system shall mark the provider as failed and attempt the next provider.`

#### Unwanted Behaviour Requirements (Context Validation)

- **SR-02-1201:** `If the user context is missing from the request, then the system shall return HTTP 500 with error "User not found in context".`
- **SR-02-1202:** `If the provider service fails to load providers, then the system shall return HTTP 500 with error "failed to load providers".`
- **SR-02-1203:** `If the request body is invalid or cannot be read, then the system shall return HTTP 400 with error "invalid request body".`

---

## API Endpoints

### Relay Endpoints

| Endpoint | Method | Description | Tool Types |
|----------|--------|-------------|------------|
| `/api/v1/relay/:tool/v1/messages` | POST | Anthropic-style messages endpoint | claude, codex, opencode |
| `/api/v1/relay/:tool/v1/chat/completions` | POST | OpenAI-style chat completions | codex, opencode |

**URL Parameters:**
- `:tool` - Tool type (claude, codex, or opencode)

**Authentication:**
- JWT Bearer token required

**Request Body:**
- Varies by tool type and endpoint (forwarded to provider)

**Response:**
- Provider response forwarded to client
- Streaming responses are forwarded as-is

**Error Responses:**
| HTTP Code | Error Message | Condition |
|-----------|---------------|-----------|
| 400 | `invalid tool kind: <tool>` | Invalid tool parameter |
| 400 | `invalid request body` | Malformed request body |
| 404 | `no available provider supports model '<model>'` | No provider supports requested model |
| 404 | `no providers available` | No providers configured |
| 500 | `User not found in context` | Missing user context |
| 500 | `failed to load providers` | Provider service error |
| 400 | `all <n> providers failed (<attempts> attempts total): <details>` | All providers failed |

---

## Business Rules

- **BR-02-001:** Relay endpoints require valid JWT authentication
- **BR-02-002:** Provider selection uses level field (lower = higher priority)
- **BR-02-003:** Only enabled providers are used for relaying
- **BR-02-004:** Providers with missing API URL or API key are automatically skipped
- **BR-02-005:** Providers failing configuration validation are automatically skipped
- **BR-02-006:** Usage is tracked per-tenant for multi-tenant isolation
- **BR-02-007:** Token usage is extracted from streaming responses for cost tracking

---

## Edge Cases & Error Handling

| Scenario | System Behavior |
|----------|----------------|
| Invalid tool kind | Return 400: "invalid tool kind: <tool>" |
| No providers configured | Return 404: "no providers available" |
| No provider supports model | Return 404: "no available provider supports model '<model>'" |
| All providers fail | Return 400: "all <n> providers failed (<attempts> attempts total)" |
| User context missing | Return 500: "User not found in context" |
| Provider service error | Return 500: "failed to load providers" |
| Invalid request body | Return 400: "invalid request body" |
| Missing authentication | Return 401: Unauthorized (handled by auth middleware) |
| Expired JWT token | Return 401: Unauthorized (handled by auth middleware) |

---

## Success Criteria

- Relay endpoints accept requests for all three tool types
- Provider failover works automatically on failures
- Token usage is extracted and recorded accurately
- Multi-tenant data isolation is maintained
- Streaming responses are forwarded correctly
- Model mapping transformations are applied
- All errors are logged with sufficient context

---

**Related:** [02.01 Provider Configuration](../01_provider_configuration/) | [02.02 Request Routing](../02_request_routing/) | [02.03 Model Mapping](../03_model_mapping/) | [04.01 Usage Data Collection](../../04_usage_insights/01_data_collection/) | [Domain 02 Overview](../README.md)
