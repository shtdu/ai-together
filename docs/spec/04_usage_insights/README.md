# Usage Insights

The system tracks AI tool usage and presents insights to users. Managers see team-wide data, while members see personal statistics.

## Purpose

Provide visibility into AI usage patterns, costs, and trends for both individuals and teams.

## Subdomains

| Subdomain | Status | Description |
|-----------|--------|-------------|
| **Data Collection** | ✅ Implemented | What data is collected and how |
| **Personal Analytics** | ✅ Implemented | Individual user insights |
| **Team Analytics** | ✅ Implemented | Organization-wide insights (Commercial) |
| **Cost Tracking** | ✅ Implemented | How costs are calculated and displayed |
| **Data Retention** | ✅ Implemented | How long data is kept by license type |

## Quick Facts

- **Privacy-first:** Only metadata collected, no prompt/response content
- **Collected data:** Model name, tokens, provider, timestamp, duration, status
- **Personal analytics:** Available to all users
- **Team analytics:** Available to managers (Commercial)
- **Cost calculation:** Estimates based on provider pricing models

## Update Intervals

Analytics dashboards update at different intervals based on scope:

| Analytics Type | Update Interval | Rationale |
|----------------|-----------------|-----------|
| **Personal Analytics** | 5 seconds | Individual data, low volume, immediate feedback valuable |
| **Team Analytics** | 30 seconds | Aggregated data, higher volume, balanced freshness and load |

**Why different intervals?**
- Personal analytics show a single user's requests - updates are quick and low-load
- Team analytics aggregate data from all users - more expensive to compute
- The 30-second team interval provides near-real-time visibility without overwhelming the server

## User Stories

- As a **member**, I want to see my token usage so I can track my consumption
- As a **manager**, I want to see team costs so I can manage budget
- As a **manager**, I want to know which providers are most used so I can optimize
- As a **member**, I want to export my data so I can analyze it further

## Related Documentation

- **Data Collection:** [`04.01 Data Collection`](01_data_collection/) - Privacy and what's tracked
- **Personal Analytics:** [`04.02 Personal Analytics`](02_personal_analytics/) - Individual insights
- **Team Analytics:** [`04.03 Team Analytics`](03_team_analytics/) - Organization insights
- **Cost Tracking:** [`04.04 Cost Tracking`](04_cost_tracking/) - Cost calculation
- **Data Retention:** [`04.05 Data Retention`](05_data_retention/) - Retention policies

---

**Next:** [User Interfaces](../05_user_interfaces/)
