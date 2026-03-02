# Personal Analytics

All users can view their own AI usage statistics, regardless of license tier.

## Purpose

Help individual users understand their personal AI tool usage patterns and costs.

## User Stories

- As a **member**, I want to see my daily token usage so I can track consumption
- As a **member**, I want to know which models I use most
- As a **member**, I want to see my estimated costs so I can budget
- As a **member**, I want to see my request history so I can find past activity

## Available Metrics

### Summary Statistics

**Today's Usage:**
- Total requests
- Total tokens (input + output)
- Estimated cost
- Success rate (successful requests / total requests × 100%)

**Time Periods:**
- Today
- Last 7 days
- Last 30 days
- Custom date range

### Request History

**Per Request:**
- Timestamp
- AI tool used (Claude/Codex/OpenCode)
- Model name
- Provider used
- Token counts (input, output, cache)
- Duration
- Estimated cost

**Features:**
- Paginated list (100 per page)
- Sortable by any column
- Filterable by tool, provider, model
- Export to CSV/JSON

### Usage Patterns

**Visualizations:**
- Usage heatmap (GitHub-style) - shows activity over time
- Token trend chart - usage over time
- Provider breakdown - which providers used most
- Model breakdown - which models used most

## Data Availability

| Tier | Personal Data Available |
|------|------------------------|
| All tiers | ✅ Full personal analytics |

Personal analytics are available to all users, regardless of license tier.

## Functional Requirements

- **FR-001:** Users can view their usage data in real-time
- **FR-002:** Data updates in near real-time (suggested: within 5 seconds)
- **FR-003:** Users can filter by date range, tool, provider, model
- **FR-004:** Users can export their data as CSV/JSON
- **FR-005:** Historical data available per license retention policy

## Business Rules

- **BR-001:** Users see only their own data, never others'
- **BR-002:** Personal data is available immediately after first request
- **BR-003:** Data is retained per license type (see Data Retention)
- **BR-004:** Cost estimates are approximate (based on provider pricing)

## Edge Cases & Error Handling

| Scenario | System Behavior |
|----------|----------------|
| No usage data yet | Show "No data available" with help text |
| Date range exceeds retention | Show error: "Select date within retention period" |
| Export fails | Show error: "Export failed. Try again." |
| Cost data unavailable | Show "Cost calculation unavailable" |

## Success Criteria

- Personal data is accurate and complete
- Data displays in real-time (within 5 seconds)
- Filters work correctly
- Export provides complete data
- Cost estimates are reasonably accurate

---

**Related:** [04.03 Team Analytics](../03_team_analytics/) | [04.04 Cost Tracking](../04_cost_tracking/)
