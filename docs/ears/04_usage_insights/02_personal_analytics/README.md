# Personal Analytics

All users can view their own AI usage statistics, regardless of license tier.

## Purpose

Help individual users understand their personal AI tool usage patterns and costs.

## Functional Requirements (EARS Format)

### 1. Personal Analytics Display

**Purpose:** Enable users to view their own usage statistics.

#### Event-Driven Requirements (Data Viewing)

- **PA-04-101:** `When a user requests to view personal analytics, the system shall display usage statistics for the user.`
- **PA-04-102:** `When a user views personal analytics, the system shall update the data in near real-time (within 5 seconds).`
- **PA-04-103:** `When a user views personal analytics, the system shall display only the user's own data.`

#### State-Driven Requirements (Data Availability)

- **PA-04-201:** `While a user has made at least one request, the system shall make personal data immediately available.`
- **PA-04-202:** `While personal analytics are displayed, the system shall show data within the retention period of the license type.`

---

### 2. Summary Statistics

**Purpose:** Display aggregated usage statistics.

#### Event-Driven Requirements (Statistics Display)

- **PA-04-301:** `When a user views today's usage, the system shall display total requests.`
- **PA-04-302:** `When a user views today's usage, the system shall display total tokens (input + output).`
- **PA-04-303:** `When a user views today's usage, the system shall display estimated cost.`
- **PA-04-304:** `When a user views today's usage, the system shall display success rate (successful requests / total requests × 100%).`

#### Event-Driven Requirements (Time Period Selection)

- **PA-04-305:** `When a user selects a time period (Today, Last 7 days, Last 30 days, Custom), the system shall display statistics for the selected period.`

---

### 3. Request History

**Purpose:** Provide detailed history of user's requests.

#### Event-Driven Requirements (History Display)

- **PA-04-401:** `When a user views request history, the system shall display a paginated list of requests (100 per page).`
- **PA-04-402:** `When a user views request history, the system shall display timestamp, AI tool, model name, provider, token counts, duration, and estimated cost for each request.`
- **PA-04-403:** `When a user sorts the request history, the system shall reorder the list by the selected column.`
- **PA-04-404:** `When a user filters the request history by tool, provider, or model, the system shall display only matching requests.`

#### Event-Driven Requirements (Data Export)

- **PA-04-405:** `When a user exports request history, the system shall generate a CSV file with all request data.`
- **PA-04-406:** `When a user exports request history, the system shall generate a JSON file with all request data.`

---

### 4. Usage Patterns and Visualizations

**Purpose:** Provide visual insights into usage patterns.

#### Event-Driven Requirements (Visualization Display)

- **PA-04-501:** `When a user views usage patterns, the system shall display a GitHub-style usage heatmap.`
- **PA-04-502:** `When a user views usage patterns, the system shall display a token trend chart over time.`
- **PA-04-503:** `When a user views usage patterns, the system shall display a provider breakdown chart.`
- **PA-04-504:** `When a user views usage patterns, the system shall display a model breakdown chart.`

---

### 5. Data Filtering and Customization

**Purpose:** Allow users to customize their analytics view.

#### Event-Driven Requirements (Filtering)

- **PA-04-601:** `When a user filters by date range, the system shall display data within the selected range.`
- **PA-04-602:** `When a user filters by tool, the system shall display only requests from the selected tool.`
- **PA-04-603:** `When a user filters by provider, the system shall display only requests from the selected provider.`
- **PA-04-604:** `When a user filters by model, the system shall display only requests using the selected model.`

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

## Business Rules

- **BR-04-001:** Users see only their own data, never others'
- **BR-04-002:** Personal data is available immediately after first request
- **BR-04-003:** Data is retained per license type (see Data Retention)
- **BR-04-004:** Cost estimates are approximate (based on provider pricing)

---

**Related:** [04.01 Data Collection](../01_data_collection/) | [04.03 Team Analytics](../03_team_analytics/) | [04.04 Cost Tracking](../04_cost_tracking/)
