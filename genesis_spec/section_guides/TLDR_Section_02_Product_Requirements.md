# Section 02: Product Requirements

## User Personas

### Persona 1: AI Coding Agent (Primary User)

**Who:** AI assistants (Claude Code, Cursor, Codex, Windsurf, Zed, Aider, Copilot, Gemini) running in terminal sessions.

**Key needs:**
- Structured state capture before context ends (handoff with done/remaining/decisions/uncertain)
- Session-scoped identity for review isolation (prevents self-approval)
- CLI-first interface (no GUI required)
- Optimized context output (`td usage`) for limited token budgets
- Automatic session detection from environment variables and process tree

**Evidence:** `internal/session/agent_fingerprint.go` defines 10 agent types with process tree walking. `cmd/context.go` generates optimized context blocks. CLAUDE.md mandates `td usage --new-session` at conversation start.

### Persona 2: Human Developer / Team Lead

**Who:** Humans who manage AI agent work, review agent-completed tasks, and coordinate workflows.

**Key needs:**
- Visual dashboard for tracking agent work (TUI monitor with 3 panels)
- Review/approve agent-completed work (with self-review enforcement)
- Board and swimlane/kanban views for project planning
- Query language for filtering and searching issues
- Sync across devices and agents

**Evidence:** `pkg/monitor/` provides a full BubbleTea TUI with mouse support, kanban, stats. `cmd/review.go` implements approve/reject workflow. `cmd/board.go` manages boards. `internal/query/` implements TDQ.

### Persona 3: Orchestrator Agent

**Who:** A meta-agent (e.g., Oracle) that coordinates multiple sub-agents across issues and projects.

**Key needs:**
- Assign issues to named agents
- Track agent activity across issues (8 activity types)
- Critical path analysis for prioritization
- Epic/parent-child hierarchy management
- Work session grouping for related issues

**Evidence:** `cmd/assign.go`, `cmd/activity.go`, `cmd/dependencies.go` (critical-path with Kahn's algorithm), `cmd/epic.go`, `cmd/tree.go`, `cmd/ws.go`.

---

## Feature Inventory (Derived from JTBD Analysis)

### Job J01: Track work state across AI agent context windows

**Importance:** Critical

| Feature | Command(s) | Status | Implementation Path |
|---|---|---|---|
| Structured handoff capture | `td handoff` | Complete | `cmd/handoff.go` -- flags for --done/--remaining/--decision/--uncertain, YAML stdin, file input |
| Optimized context generation | `td usage` | Complete | `cmd/context.go` -- generates session, focused issue, handoff, in-progress, reviewable, ready issues |
| Session rotation | `td usage --new-session` | Complete | `cmd/context.go` calls `session.ForceNewSession()` |
| Quiet mode (compact) | `td usage -q` | Complete | `cmd/context.go` -- reduced output after first read |
| Resume work on issue | `td resume` | Complete | `cmd/resume.go` -- shows context and sets focus |
| Check handoff hook | `td check-handoff` | Complete | `cmd/focus.go` -- returns exit code 1 if no handoff exists |
| Cascading handoffs to descendants | `td handoff` | Complete | `cmd/handoff.go` -- iterates children and creates per-child handoffs |
| Git diff stats since start | `td handoff` | Complete | `cmd/handoff.go` -- compares git state from start snapshot |

### Job J02: Prevent self-review bypass in automated workflows

**Importance:** Critical

| Feature | Command(s) | Status | Implementation Path |
|---|---|---|---|
| Session isolation (branch + agent) | Automatic | Complete | `internal/session/session.go:GetOrCreate()` scopes by `branch + AgentFingerprint.String()` |
| Self-approval prevention | `td approve` | Complete | `cmd/review.go` -- `WasSessionInvolved()` + creator/implementer session checks |
| Self-close prevention | `td close` | Complete | `cmd/close.go` -- blocks self-close; `--self-close-exception` override with security audit log |
| Minor task bypass | `td create --minor` | Complete | `cmd/create.go` -- `--minor` flag sets `Issue.Minor=true`, skips review requirement |
| Security audit logging | `td security` | Complete | `cmd/security.go` -- views security exception log (self-close exceptions) |
| Workflow state machine guards | Automatic | Complete | `internal/workflow/guards.go` -- `DifferentReviewerGuard` on in_review->closed, `BlockedGuard` on blocked->in_progress |
| Liberal/Advisory/Strict modes | Config | Complete | `internal/workflow/workflow.go` -- `ModeLiberal` (default, skips guards), `ModeAdvisory` (warns), `ModeStrict` (blocks) |

### Job J03: Manage issue lifecycle from creation to closure

**Importance:** Critical

| Feature | Command(s) | Status | Implementation Path |
|---|---|---|---|
| Create issues | `td create` (aliases: add, new) | Complete | `cmd/create.go` -- type prefix parsing, title validation, git snapshot, creator session |
| Start work | `td start` (alias: begin) | Complete | `cmd/start.go` -- multi-issue, WIP limit warning (>4), git snapshot, auto-focus |
| Submit for review | `td review` (aliases: submit, finish) | Complete | `cmd/review.go` -- auto-handoff, cascades to descendants |
| Approve | `td approve` | Complete | `cmd/review.go` -- self-review prevention, --all bulk, auto-unblock dependents |
| Reject | `td reject` | Complete | `cmd/review.go` -- records reason, returns to in_progress |
| Close directly | `td close` (aliases: done, complete) | Complete | `cmd/close.go` -- self-close prevention with exception override |
| Block/Unblock | `td block`, `td unblock` | Complete | `cmd/block.go` |
| Reopen | `td reopen` | Complete | `cmd/reopen.go` |
| Unstart | `td unstart` (alias: stop) | Complete | `cmd/unstart.go` -- reverts in_progress to open |
| Update fields | `td update` (alias: edit) | Complete | `cmd/update.go` -- bulk update, append mode, inline comment |
| Soft delete/Restore | `td delete`, `td restore` | Complete | `cmd/delete.go` -- soft delete with deleted_at timestamp |
| Undo | `td undo` | Complete | `cmd/undo.go` -- 20+ action types with JSON snapshots |

### Job J04: Provide real-time visual dashboard

**Importance:** High

| Feature | Command(s) | Status | Implementation Path |
|---|---|---|---|
| Three-panel TUI layout | `td monitor` | Complete | `pkg/monitor/model.go` -- Current Work (top), Task List (middle), Activity (bottom) |
| Categorized task view | `td monitor` | Complete | `pkg/monitor/view.go` -- groups by Review, Rework, Ready, Blocked, Closed |
| Kanban board view | `td monitor` | Complete | `pkg/monitor/kanban.go` -- Linear-style side-by-side status columns |
| Swimlane board view | `td monitor` | Complete | `pkg/monitor/view.go` -- status-grouped board issues |
| Backlog board view | `td monitor` | Complete | `pkg/monitor/view.go` -- flat ordered list with positions |
| Mouse support (click, scroll, drag) | `td monitor` | Complete | `pkg/monitor/mouse/` -- click, double-click, scroll, drag-to-resize panels |
| Stacked modal details | `td monitor` | Complete | `pkg/monitor/model.go` -- modal stack allows opening issues from within modals |
| Declarative modal system | `td monitor` | Complete | `pkg/monitor/modal/` -- sections (text, list, button, custom, input) |
| Keyboard navigation | `td monitor` | Complete | `pkg/monitor/keymap/` -- 60+ commands across 17 contexts |
| Statistics dashboard | `td monitor` | Complete | `pkg/monitor/stats.go` -- bar charts, distribution, velocity |
| Inline issue creation/editing | `td monitor` | Complete | `pkg/monitor/form.go` -- charmbracelet/huh forms |
| Getting Started onboarding | `td monitor` | Complete | `pkg/monitor/getting_started.go` -- first-run modal |
| Auto-restore filter/board state | `td monitor` | Complete | `pkg/monitor/model.go` -- loads from config.json and boards.last_viewed_at |
| Embeddable in external apps | API | Complete | `pkg/monitor/model.go:NewEmbedded()` -- custom renderers, shared DB pool |

### Job J05: Organize and query issues with flexible filtering

**Importance:** High

| Feature | Command(s) | Status | Implementation Path |
|---|---|---|---|
| TDQ query language | `td query` (alias: q) | Complete | `internal/query/` -- lexer, parser, AST, evaluator (3,419 source lines) |
| Field comparisons (15+ fields) | `td query` | Complete | `internal/query/evaluator.go` -- status, type, priority, labels, assignee, sprint, etc. |
| Boolean logic (AND/OR/NOT) | `td query` | Complete | `internal/query/parser.go` -- implicit AND, operator precedence |
| Functions (has, is, any, blocks, blocked_by, descendant_of, rework, stale) | `td query` | Complete | `internal/query/evaluator.go` -- 8 built-in functions |
| Cross-entity search (log.*, comment.*, handoff.*, file.*) | `td query` | Complete | `internal/query/evaluator.go` -- queries across related entities |
| Relative dates (-7d, -24h, -1w) | `td query` | Complete | `internal/query/evaluator.go` -- date arithmetic |
| @me session reference | `td query` | Complete | `internal/query/evaluator.go` -- resolves to current session |
| Explain mode | `td query --explain` | Complete | `cmd/query.go` -- shows parsed query AST |
| Full-text search with ranking | `td search` | Complete | `cmd/search.go` -- `SearchIssuesRanked` |
| Extensive list filtering (25+ flags) | `td list` (alias: ls) | Complete | `cmd/list.go` -- status, type, labels, priority, search, assignee, project, etc. |
| Named query-based boards | `td board` | Complete | `cmd/board.go` -- TDQ query defines board membership |

### Job J06: Manage dependencies and critical path

**Importance:** High

| Feature | Command(s) | Status | Implementation Path |
|---|---|---|---|
| Add/remove dependencies | `td dep add`, `td dep rm` | Complete | `cmd/dependencies.go` |
| Show dependency tree | `td dep`, `td blocked-by` | Complete | `cmd/dependencies.go` -- transitive resolution |
| Critical path analysis | `td critical-path` | Complete | `cmd/dependencies.go` -- Kahn's algorithm weighted by block count |
| Auto-unblock on close/approve | `td approve`, `td close` | Complete | `cmd/review.go` -- checks and unblocks dependents |

### Job J07: Coordinate multi-agent work sessions

**Importance:** High

| Feature | Command(s) | Status | Implementation Path |
|---|---|---|---|
| Start work session | `td ws start` | Complete | `cmd/ws.go` |
| Tag/untag issues | `td ws tag`, `td ws untag` | Complete | `cmd/ws.go` -- auto-starts open issues on tag |
| Log to all tagged issues | `td ws log` | Complete | `cmd/ws.go` |
| Cascading handoffs | `td ws handoff` | Complete | `cmd/ws.go` -- generates per-issue handoffs |
| Bulk review submission | `td ws handoff --review` | Complete | `cmd/ws.go` -- submits all for review on handoff |
| End session | `td ws end` | Complete | `cmd/ws.go` |
| List/show sessions | `td ws list`, `td ws show` | Complete | `cmd/ws.go` |

### Job J08: Assign and track agent activity

**Importance:** Medium

| Feature | Command(s) | Status | Implementation Path |
|---|---|---|---|
| Assign to named agent | `td assign` (alias: delegate) | Complete | `cmd/assign.go` -- records in agent activity log |
| View activity by issue/agent | `td activity` (alias: act) | Complete | `cmd/activity.go` -- 8 activity types |
| Activity types | N/A | Complete | `internal/models/models.go:108-117` -- assigned, started, committed, pr_created, reviewed, completed, spawned_subagent, comment |

### Job J09: Sync issue state across devices

**Importance:** Medium

| Feature | Command(s) | Status | Implementation Path |
|---|---|---|---|
| Push/pull sync | `td sync push`, `td sync pull` | Complete (gated) | `cmd/sync.go` -- event-sourced with 500/batch |
| Snapshot bootstrap | `td sync init` | Complete (gated) | `cmd/sync.go` -- downloads server snapshot for first sync |
| Auto-sync on mutations | Automatic | Complete (gated) | `cmd/autosync.go` -- debounced push/pull in PersistentPreRun/PostRun |
| Device code auth | `td auth login` | Complete (gated) | `cmd/auth.go` -- device code flow similar to GitHub |
| Project management | `td project create/join/invite/kick/role` | Complete (gated) | `cmd/project.go` |
| Conflict detection | `td sync conflicts` | Complete (gated) | `cmd/sync_conflicts.go` -- shows conflicts, `--resolve` flag |
| Diagnostics | `td doctor` | Complete (gated) | `cmd/doctor.go` -- checks auth, server, DB, sync state |
| Feature gates | `td feature set sync_cli true` | Complete | `cmd/feature.go` -- gates: sync_cli, sync_autosync, sync_notes |

### Job J10: Track progress with structured logging

**Importance:** Medium

| Feature | Command(s) | Status | Implementation Path |
|---|---|---|---|
| Typed log entries | `td log` | Complete | `cmd/log.go` -- 7 types: progress, blocker, decision, hypothesis, tried, result, orchestration |
| Flexible arg parsing | `td log` | Complete | `cmd/log.go` -- auto-detects issue ID vs. message |
| Stdin support | `td log` | Complete | `cmd/log.go` -- reads from stdin for long messages |
| Auto-use focused issue | `td log` | Complete | `cmd/log.go` -- falls back to config.FocusedIssueID |

### Job J11: Epic hierarchy management

**Importance:** Medium

| Feature | Command(s) | Status | Implementation Path |
|---|---|---|---|
| Create epics | `td epic create` | Complete | `cmd/epic.go` -- delegates to `create --type=epic` |
| List epics | `td epic list` | Complete | `cmd/epic.go` |
| Add child relationships | `td tree add-child` | Complete | `cmd/tree.go` |
| Visualize tree hierarchy | `td tree` | Complete | `cmd/tree.go` -- depth control |
| Cascade status on review | `td review` | Complete | `cmd/review.go` -- cascades to descendants |

### Job J12: Undo accidental changes

**Importance:** Medium

| Feature | Command(s) | Status | Implementation Path |
|---|---|---|---|
| Undo last action | `td undo` | Complete | `cmd/undo.go` -- 20+ action types with full snapshot restore |
| View recent actions | `td last` | Complete | `cmd/last.go` -- shows action log |
| Action log viewer | `td undo --list` | Complete | `cmd/undo.go` |

### Job J13: Track files associated with issues

**Importance:** Low

| Feature | Command(s) | Status | Implementation Path |
|---|---|---|---|
| Link files with glob | `td link` | Complete | `cmd/link.go` -- glob expansion, recursive directory, SHA recording |
| Unlink files | `td unlink` | Complete | `cmd/unlink.go` |
| Show file status | `td files` | Complete | `cmd/files.go` -- SHA change detection, git status integration |

### Job J14: Import and export issue data

**Importance:** Low

| Feature | Command(s) | Status | Implementation Path |
|---|---|---|---|
| Export to JSON/Markdown | `td export` | Complete | `cmd/system.go` -- format selection, --all flag |
| Import from JSON/Markdown | `td import` | Complete | `cmd/system.go` -- dry-run, force overwrite, auto-detect format |

---

## Half-Implemented Features

These features have partial implementation -- enough to be visible in the codebase but not fully functional end-to-end.

| Feature | Current Status | Details | Evidence |
|---|---|---|---|
| **Linear Integration** | Fields present, no sync | Issue model has `LinearID` and `LinearIdentifier` fields. Activity log displays them. But no commands exist for Linear API sync. | `internal/models/models.go:91-92` |
| **Sprint Management** | Field only | Issues have a `Sprint` string field settable via `td update --sprint`, but no sprint lifecycle commands (create, close, burndown). | `internal/models/models.go:89` |
| **Worktree Integration** | Message defined, no handler | `SendTaskToWorktreeMsg` defined in TUI types for sidecar embedding. No command-level worktree management. | `pkg/monitor/types.go` |
| **Orchestration Log Type** | Type defined, no special handling | `LogTypeOrchestration` exists as a constant but receives no special display, filtering, or query treatment. | `internal/models/models.go:54` |
| **Future Workflow Guards** | Defined, not attached | `EpicChildrenGuard`, `SelfCloseGuard`, `InProgressRequiredGuard` are fully implemented but not wired into `AllTransitions()`. | `internal/workflow/guards.go:94-207` |

---

## Requirements Traceability Matrix

This matrix maps user needs to the code that implements them, enabling auditors to verify each requirement has a corresponding implementation.

| Requirement | Persona | JTBD | Commands | Key Source Files |
|---|---|---|---|---|
| Auto-detect AI agent type | AI Agent | J01 | Automatic | `internal/session/agent_fingerprint.go` |
| Preserve state across context windows | AI Agent | J01 | handoff, usage, resume | `cmd/handoff.go`, `cmd/context.go` |
| Prevent self-review | AI Agent, Team Lead | J02 | approve, close | `internal/workflow/guards.go`, `cmd/review.go` |
| Full issue lifecycle | All | J03 | create, start, review, approve, close, etc. | `internal/workflow/transitions.go` (15 transitions) |
| Visual project dashboard | Team Lead | J04 | monitor | `pkg/monitor/` (34 Go files, 13K LOC) |
| Flexible issue querying | All | J05 | query, search, list | `internal/query/` (3.4K LOC) |
| Dependency management | Orchestrator | J06 | dep, critical-path | `cmd/dependencies.go` |
| Multi-issue coordination | Orchestrator | J07 | ws start/tag/log/handoff | `cmd/ws.go` |
| Agent activity tracking | Orchestrator | J08 | assign, activity | `cmd/assign.go`, `cmd/activity.go` |
| Multi-device sync | Team Lead | J09 | sync, auth, project | `internal/sync/`, `internal/syncclient/` |
| Structured progress logging | AI Agent | J10 | log | `cmd/log.go` |
| Epic hierarchy | Orchestrator | J11 | epic, tree | `cmd/epic.go`, `cmd/tree.go` |
| Undo mistakes | All | J12 | undo, last | `cmd/undo.go` |
| File tracking | AI Agent | J13 | link, files | `cmd/link.go`, `cmd/files.go` |
| Data portability | Team Lead | J14 | export, import | `cmd/system.go` |

---

## Command Count Summary

| Group | Count | Examples |
|---|---|---|
| Core | 9 | create, list, show, update, delete, board, activity, assign, note |
| Workflow | 11 | start, handoff, review, approve, reject, close, block, unblock, reopen, unstart, log |
| Query | 2 | query, search |
| Shortcuts | 8 | reviewable, blocked, in-review, ready, next, deleted, epic, task |
| Session | 9 | session, status, focus, unfocus, check-handoff, resume, usage, whoami, ws |
| Files | 3 | link, unlink, files |
| System | 13 | init, monitor, info, version, export, import, upgrade, undo, last, workflow, security, feature, hello |
| Sync (gated) | 12 | sync push/pull/status/conflicts/init/tail, auth, config, project, doctor |
| Analytics | 3 | stats analytics, stats security, stats errors |
| Dependencies | 5 | dep add, dep rm, dep, blocked-by, critical-path |
| Hierarchy | 3 | tree add-child, tree, comment |
| **Total** | **~72** | |

---

## Identified Gaps

| Gap ID | Title | Severity | Who Wants It |
|---|---|---|---|
| G01 | No time tracking or estimation vs. actual comparison | Low | Team leads analyzing velocity |
| G02 | No notification or webhook system | Medium | Orchestrator agents needing real-time events |
| G03 | No issue templates or recurring tasks | Low | Teams with standardized workflows |
| G04 | No git branch/PR linkage beyond snapshot | Low | Teams wanting GitHub/GitLab integration |
| G05 | No multi-project support in single database | Low | Organizations managing multiple repos |
| G06 | No configurable workflow state machine | Low | Teams with custom statuses |
| G07 | No rich attachment support | Low | Bug reporters needing screenshots |
| G08 | Limited Linear integration (fields present, no sync) | Medium | Teams using Linear as primary tracker |
