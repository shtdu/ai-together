# Team Analytics

Managers can view organization-wide usage analytics (Commercial license).

## Purpose

Enable managers to understand team-wide AI usage patterns, optimize provider configurations, and manage costs.

## User Stories

- As a **manager**, I want to see team usage so I can understand consumption
- As a **manager**, I want to see costs by user so I can identify heavy users
- As a **manager**, I want to compare providers so I can optimize configurations
- As a **manager**, I want to see trends over time so I can forecast needs

## Availability

| License Type | Team Analytics |
|--------------|----------------|
| Open Source | ❌ Not available |
| Commercial | ✅ Available |

## Available Metrics

### Team Summary

**Aggregate Metrics:**
- Total requests (time period)
- Total tokens (input + output)
- Total cost
- Average request duration
- Success rate (successful requests / total requests × 100%)

**Time Periods:**
- Today
- Last 7 days
- Last 30 days
- Custom date range

### Per-User Analytics

**Per-User Breakdown:**
- Total requests per user
- Total tokens per user
- Total cost per user
- Most used models per user
- Activity level (active/inactive)

**Features:**
- Sortable by any column
- Export user rankings

### Provider Analytics

**Per-Provider Breakdown:**
- Request count per provider
- Token count per provider
- Cost per provider
- Success rate per provider
- Average response time per provider

**Features:**
- Compare provider performance
- Identify most/least used providers
- Track provider health

### Model Analytics

**Per-Model Breakdown:**
- Request count per model
- Token count per model
- Cost per model
- Most popular models

**Features:**
- Identify model preferences
- Track model usage trends
- Cost optimization insights

### Request History

**Full Request Log:**
- All team requests with metadata
- Filterable by user, provider, model, time range
- Sortable by any column
- Paginated (100 per page)
- Export to CSV/JSON

## Functional Requirements

- **FR-001:** Managers can view all team usage data
- **FR-002:** Data updates in near real-time (suggested: within 30 seconds)
- **FR-003:** Multiple filter dimensions available
- **FR-004:** Export functionality for all data
- **FR-005:** Historical data available per license retention policy

## Business Rules

- **BR-001:** Only managers can access team analytics
- **BR-002:** Data includes all organization members
- **BR-003:** Managers can drill down to individual user data
- **BR-004:** Data is retained per license type

## Privacy Considerations

Managers see team data, but:
- **No prompt content** - only metadata
- **No response content** - only metadata
- Users are identified by email or name (not anonymous)

## Edge Cases & Error Handling

| Scenario | System Behavior |
|----------|----------------|
| No team usage yet | Show "No data available" with guidance |
| Filter matches no data | Show "No results for selected filters" |
| Large data set export | Export in background, notify when ready |
| Date range exceeds retention | Show error: "Select date within retention period" |

## Success Criteria

- Team data is accurate and complete
- All filters work correctly
- Data displays in reasonable time (under 5 seconds)
- Export provides complete data
- Managers can gain actionable insights

---

**Related:** [04.02 Personal Analytics](../02_personal_analytics/) | [04.04 Cost Tracking](../04_cost_tracking/)