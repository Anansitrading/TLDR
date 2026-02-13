# TLDR (td) — Executive Summary

## Forensic Genesis Analysis — 2026-02-13

### What is td?

**td** is a Go CLI task management tool built specifically for AI coding agents. It provides issue tracking, structured handoffs between AI context windows, session-isolated review enforcement, a custom query language (TDQ), multi-device sync, and a rich terminal UI (TUI) with Linear-style kanban board views.

### Key Metrics

| Metric | Value |
|--------|-------|
| Language | Go 1.25.5 |
| Total Go files | 332 |
| Total lines of code | ~126,000 |
| Source files | 198 |
| Test files | 134 (0.68 ratio) |
| CLI commands | 60+ (including subcommands) |
| Database tables | 19 (client) + 5 (server) |
| Schema migrations | 29 versions |
| TUI view modes | 3 (swimlanes, backlog, kanban) |
| Direct dependencies | 18 |

### Architecture Overview

```
┌──────────────────────────────────────────────────────────────────┐
│  CLI Layer (cmd/*.go — Cobra)                                    │
│  60+ commands: create, start, handoff, review, approve, query... │
├──────────────────────────────────────────────────────────────────┤
│  TUI Layer (pkg/monitor/ — Bubbletea Elm Architecture)           │
│  Model → Update → View with modal stack, keymap, mouse support   │
├──────────────────────────────────────────────────────────────────┤
│  Business Logic                                                  │
│  ├── internal/models/     — Domain entities (Issue, Handoff, etc)│
│  ├── internal/session/    — Session management, agent detection   │
│  ├── internal/query/      — TDQ lexer/parser/evaluator           │
│  ├── internal/workflow/   — State machine guards                 │
│  └── internal/sync/       — Event sourcing sync engine           │
├──────────────────────────────────────────────────────────────────┤
│  Data Layer (internal/db/ — SQLite with WAL mode)                │
│  Parameterized queries, action logging, dual CGo/pure-Go drivers │
├──────────────────────────────────────────────────────────────────┤
│  Sync Server (cmd/td-sync/ — HTTP API)                           │
│  Device auth, RBAC, rate limiting, event-sourced sync            │
└──────────────────────────────────────────────────────────────────┘
```

### Unique Differentiators

1. **AI-Agent-First Design**: Session isolation prevents AI agents from approving their own work. Agent fingerprinting detects 10+ AI agent types from process trees.

2. **Structured Handoffs**: When an AI agent's context window ends, `td handoff` captures done/remaining/decisions/uncertain state for the next session.

3. **Custom Query Language (TDQ)**: Full lexer/parser/evaluator with 15+ fields, 13 functions, boolean logic, date arithmetic, cross-entity queries, and SQL pushdown optimization.

4. **Linear-Style Kanban TUI**: Side-by-side status columns with per-column cursors, colored headers, scroll indicators, and vim-style navigation — all in the terminal.

5. **Local-First with Sync**: SQLite database per project with event-sourced sync, conflict resolution, and a standalone HTTP sync server.

### Security Posture

**Overall: Medium-Low Risk** (as expected for a local-first CLI tool)

- SQL injection: Properly mitigated with parameterized queries throughout
- Authentication: Bearer token with SHA-256 hashed API keys, device-flow auth
- Authorization: RBAC (owner/writer/reader) + scope-based admin access
- Notable: No public API key exposure in error messages, proper CORS configuration
- Gap: Token storage uses plaintext in config file (acceptable for local tool)

### Code Quality Assessment

**Overall Score: 7/10**

**Strengths:**
- Strong test coverage (134 test files, custom E2E simulation framework)
- Clean DB abstraction with action logging baked in
- Well-designed query language with SQL optimization
- Mature modal system with declarative API

**Areas for Improvement:**
- TUI model.go is very large (needs decomposition)
- Some dead code identified (internal/crypto package, 19 unused exports)
- Commands.go complexity could benefit from extraction
- Feature flags system exists but may have incomplete features behind gates

### Documentation Inventory

This forensic analysis produced:

| Category | Files | Total Size |
|----------|-------|------------|
| Findings (JSON) | 7 | ~206K |
| Section Guides (MD) | 12 | ~250K |
| Reports | 3 | ~30K |
| **Total** | **22** | **~486K** |

### NotebookLM Validation

The codebase was uploaded to NotebookLM (notebook: `6ecddb1b-f0c3-4447-9fba-9132cf6a9954`) in 5 chunks totaling 2.1MB. Validation queries confirmed:

- All 10 most important files correctly identified
- Complete state machine transitions verified against code
- 60+ CLI commands inventoried from actual registrations
- 19 database tables with full column listings confirmed
- TDQ language operators, fields, and functions verified
- Authentication/authorization middleware paths confirmed
- Gap found: `cmd/ws.go` not included in upload (work sessions command)

### Recommended Next Steps

1. **Decompose TUI model** — Split `model.go` and `commands.go` into focused files
2. **Remove dead code** — Clean up `internal/crypto` (335 lines), unused exports
3. **Complete test coverage** — Work sessions and sync edge cases
4. **Add README.md** — The repo lacks a proper README for GitHub
5. **Document API** — Sync server API needs OpenAPI/Swagger spec
