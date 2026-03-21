#!/bin/bash
# QMD Setup Script for AI Together Documentation
# https://github.com/tobi/qmd
#
# Prerequisites:
#   npm >= 22.0.0 (qmd must be installed via npm, NOT bun)
#   Install: npm install -g @tobilu/qmd
#
# Usage:
#   ./qmd-setup.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COLLECTION_NAME="docs"

# Workaround: QMD launcher bug - it checks $BUN_INSTALL which causes
# ABI mismatch with native modules (better-sqlite3, sqlite-vec)
# https://github.com/tobi/qmd/issues/361
qmd() {
    BUN_INSTALL="" command qmd "$@"
}

echo "=== QMD Setup for AI Together Docs ==="
echo ""

# Remove existing collection if present
echo "Checking for existing collection..."
if qmd collection list 2>/dev/null | grep -q "$COLLECTION_NAME"; then
    echo "Removing existing collection '$COLLECTION_NAME'..."
    qmd collection remove "$COLLECTION_NAME" || true
fi

# Add collection
echo "Adding collection '$COLLECTION_NAME' from $SCRIPT_DIR..."
# qmd collection add "$SCRIPT_DIR" --name "$COLLECTION_NAME"

# # Add context to help with search results
# # Context works as a tree - each piece of context will be returned when matching
# # sub-documents are returned. This allows LLMs to make better contextual choices.
# echo "Adding context descriptions..."
# qmd context add "qmd://$COLLECTION_NAME" "AI Together - Team collaboration platform for managing AI coding tools (Claude Code, Codex, OpenCode). Provides centralized provider management, team-level settings, usage statistics, and automatic configuration distribution."

# qmd context add "qmd://$COLLECTION_NAME/ears" "EARS Product Requirements - Single source of truth for functional requirements using EARS syntax (Easy Approach to Requirements Syntax). Contains user stories, acceptance criteria, and business rules."

# qmd context add "qmd://$COLLECTION_NAME/ears/00_overview" "Product overview - Value proposition, user personas (Team Admin, Team Member), capabilities overview, and system scope."

# qmd context add "qmd://$COLLECTION_NAME/ears/01_identity_and_access" "Identity and Access Management - User accounts, roles/permissions (RBAC), multi-tenancy, JWT authentication, and licensing."

# qmd context add "qmd://$COLLECTION_NAME/ears/02_provider_management" "Provider Management - AI provider configuration (Anthropic, OpenAI, etc.), request routing, model mapping, and API key management."

# qmd context add "qmd://$COLLECTION_NAME/ears/03_configuration_sync" "Configuration Sync - Distribution of provider configs to members, team management, and offline mode support."

# qmd context add "qmd://$COLLECTION_NAME/ears/04_usage_insights" "Usage Insights - Data collection, personal/team analytics, cost tracking, and data retention policies."

# qmd context add "qmd://$COLLECTION_NAME/ears/05_user_interfaces" "User Interfaces - Member desktop app (Wails + Vue), Manager dashboard (React), and accessibility requirements."

# qmd context add "qmd://$COLLECTION_NAME/ears/06_system_behaviors" "System Behaviors - Privacy-first design, security architecture, performance requirements, and compliance."

# qmd context add "qmd://$COLLECTION_NAME/design" "Technical Design - Architecture decisions, design constitution, technology stack (Go, React, PostgreSQL), and implementation patterns."

# qmd context add "qmd://$COLLECTION_NAME/design/constitution.md" "Design Constitution - Foundational principles: Privacy-First, Multi-Tenant Isolation, Stateless Auth, API-First, Graceful Degradation."

# qmd context add "qmd://$COLLECTION_NAME/design/architecture.md" "System Architecture - Three-tier architecture, data flows, security design, deployment patterns, and scalability."

# qmd context add "qmd://$COLLECTION_NAME/spec" "Legacy Specifications - MECE product documentation (superseded by ears/). Use for historical reference only."

# Generate embeddings for semantic search
echo ""
echo "Generating embeddings for semantic search (this may take a few minutes)..."
qmd embed -f

# # Show status
# echo ""
# echo "=== Setup Complete ==="
# echo ""
# qmd status

# echo ""
# echo "=== Usage Examples ==="
# echo ""
# echo "# Fast keyword search"
# echo "  qmd search \"authentication\""
qmd search "authentcation"
# echo ""
# echo "# Semantic search"
# echo "  qmd vsearch \"how to configure providers\""
# echo ""
# echo "# Hybrid + reranking (best quality)"
# echo "  qmd query \"user roles and permissions\""
# echo ""
# echo "# Search within this collection"
# echo "  qmd search \"API\" -c $COLLECTION_NAME"
# echo ""
# echo "# Get a specific document"
# echo "  qmd get \"ears/01_identity_and_access/README.md\""
# echo ""
# echo "# Get document by docid (shown in search results)"
# echo "  qmd get \"#abc123\""
# echo ""
# echo "# Get multiple documents by glob pattern"
# echo "  qmd multi-get \"ears/**/*.md\""
# echo ""
# echo "# Export all matches for an agent"
# echo "  qmd search \"authentication\" --all --files --min-score 0.3"
# echo ""
