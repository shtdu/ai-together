# Cost Tracking

The system estimates AI usage costs based on provider pricing models and displays them to users.

## Purpose

Help users and managers understand the financial impact of their AI tool usage.

## Functional Requirements (EARS Format)

### 1. Cost Calculation

**Purpose:** Calculate estimated costs for AI usage.

#### Event-Driven Requirements (Per-Request Calculation)

- **CT-04-101:** `When an AI tool request completes, the system shall calculate the estimated cost based on token usage.`
- **CT-04-102:** `When calculating cost, the system shall use provider pricing models for the specific model.`
- **CT-04-103:** `When calculating cost, the system shall multiply input tokens by input token price.`
- **CT-04-104:** `When calculating cost, the system shall multiply output tokens by output token price.`
- **CT-04-105:** `When calculating cost, the system shall multiply cache tokens by cache token price (if applicable).`
- **CT-04-106:** `When calculating cost, the system shall sum input, output, and cache costs for total request cost.`

---

### 2. Cost Display

**Purpose:** Show cost information to users.

#### Event-Driven Requirements (Personal Cost Display)

- **CT-04-201:** `When a user views personal analytics, the system shall display estimated personal costs.`
- **CT-04-202:** `When a user views request history, the system shall display estimated cost per request.`
- **CT-04-203:** `When a user views cost summary, the system shall display cost breakdown by model.`
- **CT-04-204:** `When a user views cost summary, the system shall display cost breakdown by provider.`

#### Event-Driven Requirements (Team Cost Display - Commercial)

- **CT-04-205:** `When a manager views team analytics, the system shall display total team costs.`
- **CT-04-206:** `When a manager views team analytics, the system shall display cost per user.`
- **CT-04-207:** `When a manager views team analytics, the system shall display cost trends over time.`

---

### 3. Pricing Models

**Purpose:** Define how provider pricing is configured.

#### Ubiquitous Requirements

- **CT-04-301:** `The system shall store pricing models for each provider and model combination.`
- **CT-04-302:** `The system shall support per-token pricing (input, output, cache).`
- **CT-04-303:** `The system shall allow managers to update pricing models when providers change prices.`

#### Event-Driven Requirements (Pricing Updates)

- **CT-04-304:** `When a manager updates pricing models, the system shall apply new prices to future requests.`
- **CT-04-305:** `When a manager updates pricing models, the system shall not recalculate costs for historical requests.`

---

### 4. Cost Estimates

**Purpose:** Provide cost forecasting and budgeting.

#### Event-Driven Requirements (Cost Forecasting)

- **CT-04-401:** `When a user views cost analytics, the system shall display projected costs based on current usage trends.`
- **CT-04-402:** `When a manager views team costs, the system shall display monthly cost projections.`

#### State-Driven Requirements (Estimate Accuracy)

- **CT-04-403:** `While cost estimates are calculated, the system shall display a note indicating costs are approximate.`
- **CT-04-404:** `While cost estimates are displayed, the system shall not guarantee accuracy of estimates (actual billing may vary).`

## Cost Calculation Formula

```
Total Request Cost = (Input Tokens × Input Price) +
                     (Output Tokens × Output Price) +
                     (Cache Tokens × Cache Price)
```

## Business Rules

- **BR-04-001:** Cost estimates are approximate (based on provider pricing)
- **BR-04-002:** Actual billing may differ from estimates
- **BR-04-003:** Pricing models are configurable per provider
- **BR-04-004:** Historical costs are not recalculated when pricing changes

---

**Related:** [04.01 Data Collection](../01_data_collection/) | [04.02 Personal Analytics](../02_personal_analytics/) | [04.03 Team Analytics](../03_team_analytics/)
