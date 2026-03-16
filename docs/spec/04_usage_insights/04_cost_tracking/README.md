# Cost Tracking

The system calculates estimated costs based on token usage and provider pricing models.

## Purpose

Help users and managers understand AI spending and optimize costs.

## User Stories

- As a **member**, I want to see my estimated costs so I can budget
- As a **manager**, I want to see team costs so I can manage spending
- As a **manager**, I want to break down costs by provider so I can optimize
- As a **manager**, I want to see costs by user so I can identify heavy users

## How Costs Are Calculated

### Cost Formula

```
Cost = (Input Tokens / 1,000,000 × Input Price per Million)
    + (Output Tokens / 1,000,000 × Output Price per Million)
    + (Cache Tokens / 1,000,000 × Cache Read Price per Million)
```

### Pricing Data

System uses provider pricing data to calculate costs:

**Claude (Anthropic) Examples:**
- claude-3-5-sonnet: Input $3/M, Output $15/M
- claude-3-5-haiku: Input $0.80/M, Output $4/M

**OpenCode/OpenAI Examples:**
- gpt-4: Input $30/M, Output $60/M
- gpt-4-turbo: Input $10/M, Output $30/M

## Cost Accuracy

### What's Accurate

- Token counts (exact, from API responses)
- Per-request calculations (using current pricing)
- Aggregates of per-request costs

### What's Estimated

- Pricing data (may lag behind provider changes)
- Total actual billing (may include additional fees)

**Disclaimer:** Costs are estimates for reference. Actual billing may vary based on provider billing, network fees, and other factors.

## Availability

| License Type | Personal Costs | Team Costs |
|--------------|----------------|------------|
| Open Source | ✅ Available | ❌ Not available |
| Commercial | ✅ Available | ✅ Available |

## Functional Requirements

- **FR-001:** System must calculate costs for every request
- **FR-002:** Costs must be visible in near real-time
- **FR-003:** Pricing data must be updateable
- **FR-004:** Cost breakdowns must be available for filtering
- **FR-005:** Cost data must be exportable

## Business Rules

- **BR-001:** Costs are calculated per request
- **BR-002:** Costs aggregate by time period (day, week, month)
- **BR-003:** All tiers show personal costs
- **BR-004:** Team costs require Commercial license
- **BR-005:** Pricing data is updated via system configuration

## Edge Cases & Error Handling

| Scenario | System Behavior |
|----------|----------------|
| Pricing data not available | Show "Cost unavailable" |
| Provider doesn't return token counts | Estimate based on model averages |
| Cached tokens not reported | Calculate cost without cache tokens |
| Negative cost (calculation error) | Show "Cost calculation error" |

## Planned Features

> **DEFERRED** - Budget alerts and usage quotas are planned for a future release.

### Budget Alerts (Commercial)

Managers can set cost thresholds:
- Daily/monthly budget limit
- Per-user budget limit
- Per-provider budget limit

**Alerts:**
- Email when threshold exceeded
- Webhook notification (for integration)
- In-app notification

### Usage Quotas

Managers can set usage limits:
- Maximum tokens per period
- Maximum requests per period
- Maximum cost per period

**Enforcement:**
- Warning when approaching limit
- Block requests when limit exceeded (optional)

## Success Criteria

- Costs calculate correctly for all requests
- Cost displays are accurate and timely
- Pricing data stays current
- Managers can see actionable cost insights
- Users understand their personal spending

---

**Related:** [04.02 Personal Analytics](../02_personal_analytics/) | [04.03 Team Analytics](../03_team_analytics/)
