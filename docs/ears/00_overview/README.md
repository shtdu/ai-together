# Product Overview

AI Together is a **team collaboration platform for AI development tools**. It helps teams manage AI service providers, track usage, and control costs across their organization.

## Purpose

AI development tools (Claude Code, Codex, OpenCode) require API keys and provider configurations. Managing these across a team is difficult:
- Each developer must configure their own tools
- Tracking who used what and how much is manual
- Switching providers requires updating every developer's setup
- Costs are opaque until bills arrive

AI Together solves these problems by:
- **Centralized configuration** - Managers set up providers once, everyone gets them automatically
- **Transparent routing** - Developers use their normal tools, requests route through configured providers
- **Automatic failover** - If a provider fails, automatically switch to backup providers
- **Usage visibility** - Track tokens, costs, and patterns across the team

## Target Users

### Primary Personas

| Persona | Role | Goals | Pain Points |
|---------|------|-------|-------------|
| **Team Manager Alex** | Engineering Manager / Tech Lead | - Control AI provider configurations for team<br>- Track team usage and costs<br>- Ensure compliance and budget adherence | - Can't see what team is spending<br>- Manual configuration distribution<br>- No visibility into usage patterns |
| **Member Sam** | Software Engineer | - Use AI tools without friction<br>- Reliable access to AI capabilities<br>- Understand personal usage | - Managing API keys is annoying<br>- Provider outages block work<br>- Unclear how much usage costs |
| **Organization Lead Pat** | CTO / VP Engineering | - Control AI tool costs across organization<br>- Ensure compliance and security<br>- Enable teams to be productive | - Shadow IT (unauthorized tool usage)<br>- Budget surprises<br>- Inconsistent tool configurations |

## Functional Requirements (EARS Format)

### 1. Core Platform Capabilities

**Purpose:** Define the fundamental platform capabilities.

#### Ubiquitous Requirements

- **OV-00-001:** `The system shall support centralized configuration of AI providers for teams.`
- **OV-00-002:** `The system shall provide transparent routing of AI tool requests through configured providers.`
- **OV-00-003:** `The system shall implement automatic failover when providers fail.`
- **OV-00-004:`The system shall track usage metrics across the team.`
- **OV-00-005:`The system shall calculate and display cost estimates for AI usage.`

---

### 2. User Management & Access Control

**Purpose:** Enable team-based user management.

#### Ubiquitous Requirements

- **OV-00-101:** `The system shall support two user roles: Manager and Member.`
- **OV-00-102:`The system shall allow managers to invite team members and assign roles.`
- **OV-00-103:`The system shall implement role-based permissions for all operations.`
- **OV-00-104:`The system shall maintain complete multi-tenant isolation between organizations.`

---

### 3. AI Provider Management

**Purpose:** Enable configuration and management of AI service providers.

#### Event-Driven Requirements

- **OV-00-201:** `When a manager configures a provider, the system shall store the provider configuration for the team.`
- **OV-00-202:`When a manager sets provider priority, the system shall route requests based on priority order.`
- **OV-00-203:`When a manager enables or disables a provider, the system shall include or exclude the provider from routing.`

#### Ubiquitous Requirements

- **OV-00-204:`The system shall support multiple AI service providers (Claude, Codex, OpenCode, and custom providers).`
- **OV-00-205:`The system shall support model name mapping for consistent model references.`

---

### 4. Configuration Distribution

**Purpose:** Automatically distribute configuration to team members.

#### Event-Driven Requirements

- **OV-00-301:** `When a manager updates provider configuration, the system shall automatically distribute the changes to all team members.`
- **OV-00-302:`When a member connects to the team server, the system shall download the latest provider configuration.`
- **OV-00-303:`When configuration changes are available, the system shall sync to members every 5 minutes.`

---

### 5. Usage Insights & Cost Tracking

**Purpose:** Provide visibility into AI usage and costs.

#### Event-Driven Requirements

- **OV-00-401:`When a user views personal analytics, the system shall display token usage and costs for the user.`
- **OV-00-402:`When a manager views team analytics, the system shall display aggregated usage data across all users (Commercial license).`
- **OV-00-403:`When an AI tool request completes, the system shall calculate estimated costs based on token usage.`

#### Ubiquitous Requirements

- **OV-00-404:`The system shall retain usage metadata for 7 days (Open Source) or 90 days (Commercial).`
- **OV-00-405:`The system shall only collect metadata (no prompt or response content).`

---

### 6. User Interfaces

**Purpose:** Provide appropriate interfaces for different user types.

#### Ubiquitous Requirements

- **OV-00-501:`The system shall provide a desktop application for members (Windows, macOS, Linux).`
- **OV-00-502:`The system shall provide a web dashboard for managers.`
- **OV-00-503:`The desktop application shall run in the background and provide system tray access.`
- **OV-00-504:`The web dashboard shall require authentication and manager role access.`

---

### 7. Privacy & Security

**Purpose:** Ensure privacy-first design and secure operation.

#### Ubiquitous Requirements

- **OV-00-601:`The system shall never store prompt or response content from AI tool requests.`
- **OV-00-602:`The system shall only collect metadata about AI tool requests.`
- **OV-00-603:`The system shall implement secure authentication with password-based login.`
- **OV-00-604:`The system shall isolate each organization's data at the database level.`
- **OV-00-605:`The system shall encrypt data in transit using TLS 1.3 or higher.`

## Product Capabilities Overview

### 1. User Management & Access Control

Managers can invite team members and assign roles:
- **Managers** - Full access to configure providers and view team analytics
- **Members** - Can use providers and view personal usage

Multi-tenant isolation ensures each organization's data is completely separate.

**See:** [`01_identity_and_access/`](../01_identity_and_access/)

### 2. AI Provider Management

Configure multiple AI service providers:
- Support for Claude, Codex, OpenCode/OpenAI, and custom providers
- Set priority order for automatic selection
- Enable/disable providers without deletion
- Model name mapping (e.g., "claude-sonnet" → "claude-3-5-sonnet-20241022")

**See:** [`02_provider_management/`](../02_provider_management/)

### 3. Configuration Distribution

Providers configured by managers automatically distribute to all team members:
- Members install desktop app and connect to team server
- Configuration syncs every 5 minutes
- Members can manually pull latest configuration
- Managers can push changes immediately

**See:** [`03_configuration_sync/`](../03_configuration_sync/)

### 4. Usage Insights & Cost Tracking

Track AI usage across the team:
- **Personal analytics** (all users) - Your tokens, requests, costs
- **Team analytics** (Commercial) - Aggregated team data with filtering
- **Cost calculation** - Per-model pricing with cost breakdown
- **Data retention** - 7 days (Open Source) to 90 days (Commercial)

**See:** [`04_usage_insights/`](../04_usage_insights/)

### 5. User Interfaces

Two interfaces for different user types:

**Member Desktop App** - For developers
- Windows, macOS, Linux
- Runs in background, provides system tray access
- Shows personal usage statistics
- Displays provider status and health

**Manager Web Dashboard** - For team administration
- Browser-based, requires manager role
- Team analytics and user management
- Provider configuration interface
- License management

**See:** [`05_user_interfaces/`](../05_user_interfaces/)

### 6. Privacy & Security

System behaviors that protect users and organizations:
- **Privacy-first** - Only metadata collected, no prompt/response content
- **Multi-tenant isolation** - Each tenant's data completely separated
- **Secure authentication** - Password-based with session management

**See:** [`06_system_behaviors/`](../06_system_behaviors/)

## License Types

| Type | Seats | Teams | Key Features |
|------|-------|-------|--------------|
| **Open Source** | Unlimited | 1 | Basic features, 7-day retention |
| **Commercial** | Unlimited | 1 (planned: unlimited) | Team analytics, 90-day retention |

**See:** [`01_identity_and_access/04_licensing/`](../01_identity_and_access/04_licensing/)

## Quick Links

- **Getting Started:** [`05_user_interfaces/01_member_app/`](../05_user_interfaces/01_member_app/) - Member client setup
- **Administration:** [`05_user_interfaces/02_manager_dashboard/`](../05_user_interfaces/02_manager_dashboard/) - Manager dashboard guide
- **Providers:** [`02_provider_management/`](../02_provider_management/) - Configure AI providers
- **Analytics:** [`04_usage_insights/`](../04_usage_insights/) - Usage and cost tracking

---

**Next:** [Identity & Access](../01_identity_and_access/)
