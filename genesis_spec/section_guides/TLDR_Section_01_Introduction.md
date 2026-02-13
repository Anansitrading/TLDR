# Section 01: Introduction & Executive Summary

## What is td?

`td` is a command-line task management tool built in Go, designed from the ground up for AI coding agent workflows. It provides a CLI (`td`) for structured issue tracking, a rich terminal UI monitor (`td monitor`), and a sync server (`td-sync`) for multi-device collaboration.

Unlike general-purpose issue trackers, td treats AI agents as first-class users. It auto-detects which agent is running (Claude Code, Cursor, Codex, Windsurf, Zed, Aider, Copilot, Gemini) from the process tree, enforces self-review prevention so the agent that wrote code cannot approve it, and provides a structured handoff protocol that preserves working state across context windows.

**Module path:** `github.com/marcus/td`
**Language:** Go 1.25.5
**Current version:** v0.33.0 (pre-1.0, rapid iteration)
**Codebase size:** ~126K lines across 332 Go files (59K source, 67K tests)

## Who is td for?

### Primary: AI Coding Agents
AI assistants (Claude Code, Cursor, Codex, Windsurf, Zed, Aider, Copilot, Gemini) running in terminal sessions. They need structured task tracking, state handoff between context windows, and self-review prevention. td provides CLI-first interaction, automatic session detection via environment variables and process tree walking, and optimized context output (`td usage`) that fits within token budgets.

### Secondary: Human Developers and Team Leads
Humans who manage AI agent work, review agent-completed tasks, and coordinate multi-agent workflows. They use the TUI monitor for visual dashboards, the query language (TDQ) for filtering, and board/kanban views for project planning.

### Tertiary: Orchestrator Agents
Meta-agents that coordinate multiple sub-agents, assign work, track progress across issues, and manage epics. They use assignment, activity tracking, critical path analysis, and work session grouping.

## Why td exists

AI coding agents operate in ephemeral context windows. When a context window ends, all working state is lost. Traditional issue trackers (Jira, Linear, GitHub Issues) are designed for humans using web browsers -- they lack:

- **Session-scoped identity**: No concept of "which agent session touched this issue"
- **Structured handoffs**: No protocol for preserving done/remaining/decisions/uncertainties across context boundaries
- **Self-review prevention**: No mechanism to ensure the session that implemented a change cannot rubber-stamp approve it
- **CLI-first design**: Most require browser interaction that agents cannot perform
- **Context optimization**: No way to generate a minimal token-efficient status dump for an agent's limited context window

td fills this gap with a zero-dependency, per-project SQLite database that runs entirely in the terminal.

## Key Capabilities

All capabilities below are verified against the actual source code at the paths listed.

| Capability | Status | Implementation |
|---|---|---|
| Issue lifecycle (open/in_progress/blocked/in_review/closed) | Complete | `internal/workflow/transitions.go` -- 15 transitions defined in `AllTransitions()` |
| Agent auto-detection (8 agent types + terminal) | Complete | `internal/session/agent_fingerprint.go` -- process tree walking with `detectAgentAncestor()` |
| Session-scoped identity (branch + agent fingerprint) | Complete | `internal/session/session.go` -- `GetOrCreate()` with heartbeat tracking |
| Structured handoff protocol (done/remaining/decisions/uncertain) | Complete | `internal/models/models.go:144-153` -- `Handoff` struct; `cmd/handoff.go` |
| Self-review prevention | Complete | `internal/workflow/guards.go` -- `DifferentReviewerGuard` attached to in_review -> closed |
| Custom query language (TDQ) | Complete | `internal/query/` -- lexer, parser, AST, evaluator, executor (3,419 lines) |
| TUI monitor with 3-panel layout | Complete | `pkg/monitor/` -- BubbleTea Elm architecture (13,026 lines) |
| Kanban board with swimlanes/backlog views | Complete | `pkg/monitor/kanban.go`, `board_editor.go` |
| Event-sourced sync with remote server | Complete (feature-gated) | `internal/sync/`, `internal/syncclient/`, `cmd/td-sync/` |
| Dependency tracking with critical path analysis | Complete | `cmd/dependencies.go` -- Kahn's algorithm weighted by block count |
| Multi-issue work sessions | Complete | `cmd/ws.go` -- tag, log, cascading handoffs |
| Full undo system with action log snapshots | Complete | `cmd/undo.go` -- 20+ action types with previous/new data JSON snapshots |
| Embeddable TUI (sidecar support) | Complete | `pkg/monitor/model.go` -- `NewEmbedded()`, custom renderers, shared DB pool |
| Agent activity tracking (8 activity types) | Complete | `internal/models/models.go:108-117`, `cmd/activity.go` |
| Freeform notes with pin/archive | Complete | `cmd/note.go`, `internal/db/notes.go` |
| Command usage analytics | Complete | `cmd/analytics.go` |

## Architecture at a Glance

```
                    +------------------+
                    |     main.go      |  Entry point: wires QueryValidator,
                    |                  |  resolves version, calls cmd.Execute()
                    +--------+---------+
                             |
              +--------------+---------------+
              |                              |
     +--------v--------+         +----------v-----------+
     |    cmd/ (CLI)    |         |  cmd/td-sync/ (Sync  |
     |  50+ Cobra cmds |         |     Server binary)   |
     +--------+---------+         +----------+-----------+
              |                              |
   +----------+----------+         +---------v---------+
   |                     |         |   internal/api/   |
   v                     v         |  HTTP REST API    |
+--+------+    +--------+---+     +---+------+--------+
|  pkg/   |    | internal/  |         |      |
| monitor |    |   db/      |    +----v----+ |
| (TUI)   |    |  SQLite    |    |internal/| |
| 13K LOC |    |  7.8K LOC  |    | sync/   | |
+----+----+    +-----+------+    | engine  | |
     |               |          +----+----+ |
     |               |               |      |
     |          +----v----+     +----v------v--+
     +--------->| internal|     | internal/    |
                | /models |     | serverdb/    |
                | (domain)|     | (server DB)  |
                +---------+     +--------------+

Storage: <project>/.todos/issues.db (SQLite, WAL mode)
Config:  <project>/.todos/config.json
```

**Layered dependency flow:**
- Entry Points (`main`, `cmd/td-sync`) -> CLI Layer (`cmd/`)
- CLI Layer -> Business Logic (`internal/query`, `internal/workflow`, `internal/session`, `internal/sync`)
- Business Logic -> Data Access (`internal/db`)
- Data Access -> Domain Models (`internal/models`, zero dependencies)
- TUI Layer (`pkg/monitor`) -> Business Logic + Data Access (public package for embedding)
- Server (`internal/api`, `internal/serverdb`) -> Sync Engine (`internal/sync`)

**Key architectural patterns:**
- Elm Architecture (BubbleTea Model/Update/View for TUI)
- Event Sourcing (sync protocol with action_log)
- State Machine (issue lifecycle with pluggable guards)
- Repository Pattern (db.DB wrapping SQLite)
- Logged/Unlogged operation split (local mutations vs. remote event application)
- Dependency injection via function variable (main.go wires `db.QueryValidator` to break import cycle)

## Version History Highlights

td has 72 tagged releases from v0.1.0 to v0.33.0, reflecting rapid iteration on a pre-1.0 codebase.

### Major Milestones

| Version | Theme | Notable Changes |
|---|---|---|
| **v0.1.0 -- v0.4.x** | Foundation | Core CLI (create, list, show, update, delete), basic monitor TUI, handoff protocol, session management, dependencies, work sessions |
| **v0.5.0 -- v0.9.0** | Stabilization | Refactored dependencies, single-writer queue for SQLite concurrency, action shortcuts in monitor, WIP limit warnings, interactive search |
| **v0.10.0 -- v0.12.x** | DB & TUI Maturity | Split db.go into smaller files, sidecar worktree integration, embeddable monitor, query improvements, mouse click fixes |
| **v0.13.0** | Sidecar Integration | Send-to-worktree command for external sidecar embedding |
| **v0.14.0 -- v0.15.0** | Theming | Markdown theme support with custom Chroma style builder |
| **v0.16.0 -- v0.20.0** | Polish & Workflow | Workflow state machine, file locking, declarative modal system migration, improved onboarding |
| **v0.21.0 -- v0.22.x** | Declarative Modals | Full migration to declarative modal system (stats, handoffs, board picker, delete/close confirmation) |
| **v0.23.0 -- v0.25.0** | UX Refinements | Better focus management, footer view controls, auto-unblock dependents on approval, embeddable OpenIssueByIDMsg |
| **v0.26.0 -- v0.28.x** | Session & Workflow | Session improvements, agent fingerprinting via process tree, enhanced review isolation |
| **v0.29.0 -- v0.30.0** | Sync Infrastructure | Event-sourced sync engine, sync client, auto-sync, E2E test harness, entity backfill |
| **v0.31.0 -- v0.32.0** | Admin API & TDQ Server | Admin panel for sync server, snapshot query endpoint, QuerySource interface extraction, integration test harness |
| **v0.33.0** | Notes & Kanban | Note model and CLI commands, note query support in TDQ, Linear-style kanban board view |

### Release Cadence

- **v0.1.0 through v0.4.26:** 28 releases during rapid prototyping (heavy patch releases)
- **v0.5.0 through v0.33.0:** 29 minor releases with more structured feature development
- Total: 72 releases, all pre-1.0

### Schema Evolution

The SQLite database has undergone 29 sequential migrations (v1 through v29), adding tables for issues, logs, handoffs, git snapshots, issue files, dependencies, work sessions, comments, sessions, boards, board positions, action log, session history, sync state, sync conflicts, sync history, notes, and agent activity. Several migrations use custom Go code for complex operations like text-to-ID migration, deterministic ID generation for sync, and file path normalization.

### Binaries

- `td` -- Main CLI tool (issue tracking, TUI monitor, query engine)
- `td-sync` -- Standalone sync server for multi-device collaboration
