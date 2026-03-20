# Documentation Guide

This directory contains all project documentation, organized into distinct categories.

## Documentation Hierarchy

```
docs/
├── ears/           # Single Source of Truth for REQUIREMENTS
├── spec/           # (Legacy) MECE product documentation
├── design/         # Technical architecture and design principles
└── review_summary.md
```

## Single Source of Truth for Requirements: `ears/`

**IMPORTANT:** The `ears/` directory is the **single source of truth for all product requirements**.

When implementing features or understanding requirements:

1. **ALWAYS check `ears/` first** for functional requirements
2. Requirements in `ears/` use the **EARS syntax** (Easy Approach to Requirements Syntax)
3. EARS requirements are precise, testable, and machine-parsable

### Why EARS?

EARS imposes strict grammatical patterns on requirements:

| Pattern | Syntax | Use Case |
|---------|--------|----------|
| Ubiquitous | `The <system> shall <response>.` | Global properties |
| Event-Driven | `When <trigger>, the <system> shall <response>.` | API calls, UI events |
| State-Driven | `While <state>, the <system> shall <response>.` | Modal behaviors |
| Optional Feature | `Where <feature>, the <system> shall <response>.` | Feature flags |
| Unwanted Behavior | `If <trigger>, then the <system> shall <response>.` | Error handling |

### EARS Structure

```
ears/
├── CLAUDE.md                    # EARS syntax guide
├── spec.md                      # Formal EARS specification
├── 00_overview/                 # Product overview
├── 01_identity_and_access/      # User accounts, roles, licensing
├── 02_provider_management/      # Provider config, routing
├── 03_configuration_sync/       # Distribution, teams, offline
├── 04_usage_insights/           # Analytics, cost tracking
├── 05_user_interfaces/          # Member app, manager dashboard
└── 06_system_behaviors/         # Privacy, security, performance
```

See `ears/CLAUDE.md` for detailed EARS syntax and writing guidelines.

## Other Documentation

### `design/` - Technical Architecture

| File | Purpose |
|------|---------|
| `constitution.md` | Design principles, technology stack, anti-patterns |
| `architecture.md` | System architecture, data flows, security, deployment |

Read `design/constitution.md` **before** making architectural decisions.

### `spec/` - Legacy Product Documentation

The `spec/` directory contains older MECE-structured documentation. **This is superseded by `ears/`.**

- Use `spec/` only for historical reference
- Do not add new requirements to `spec/`
- Migrate content to `ears/` when updating requirements

## Quick Reference

| Need | Go To |
|------|-------|
| What features to build? | `ears/` |
| How to write requirements? | `ears/CLAUDE.md` |
| Technology choices? | `design/constitution.md` |
| System architecture? | `design/architecture.md` |
| Module implementation? | `../member/CLAUDE.md`, `../server/CLAUDE.md` |

## Workflow for Implementation

1. **Start with `ears/`** - Find the feature requirements
2. **Check `design/constitution.md`** - Verify architectural alignment
3. **Reference `design/architecture.md`** - Understand system design
4. **Consult module CLAUDE.md** - Follow implementation patterns

---

**Remember:** `ears/` is the single source of truth for requirements. Always start there when implementing features.
