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

## Value Proposition by Persona

### For Team Managers
- **Single point of control** - Configure providers once for the entire team
- **Usage transparency** - See who's using what and how much
- **Cost management** - Track spending by provider, model, and user
- **Policy enforcement** - Set which providers and models are available

### For Members
- **Zero setup** - Install desktop app, enter credentials, start working
- **Automatic updates** - Provider changes sync automatically
- **Personal insights** - See your own usage and costs

### For Organization Leaders
- **Budget control** - Track and forecast AI spending
- **Compliance** - Ensure data stays within approved providers
- **Visibility** - Organization-wide usage analytics
- **Governance** - License tiers and seat management

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

## How It Works

### Member Experience

```
1. Install AI Together desktop app
2. Login to team account (or use standalone)
3. Desktop app auto-configures AI tools
4. Use AI tools normally - requests route through configured providers
5. View personal usage in desktop app
```

### Manager Experience

```
1. Access Manager web dashboard
2. Configure AI providers (API keys, endpoints, models)
3. Invite team members via email
4. Monitor team usage and costs
5. Adjust configurations as needed
```

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
