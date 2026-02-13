# Section 09: Monitoring & Observability

## Overview

td's observability story is unconventional: the TUI monitor itself **is** the primary observability layer. There is no external metrics pipeline, no Prometheus endpoint for the client, and no log aggregation. Instead, td provides real-time dashboard visibility through the BubbleTea TUI, CLI analytics commands, agent activity tracking, and a sync server diagnostics command (`td doctor`). The sync server (`td-sync`) does expose `/healthz` and `/metricz` HTTP endpoints.

---

## 9.1 The TUI Monitor as Observability Layer

The `td monitor` command launches a full-screen terminal UI that functions as a real-time dashboard. It is the primary way to observe what is happening across all sessions, agents, and issues.

### Three-Panel Layout

The monitor displays three vertically-stacked, resizable panels:

| Panel | Content | Data Source |
|-------|---------|-------------|
| **Current Work** (top) | Focused issue details + all in-progress issues | `config.GetFocus()` + `db.ListIssues(Status: in_progress)` |
| **Task List** (middle) | Categorized issues: Reviewable, Needs Rework, Ready, Blocked, Closed | `fetchTaskList()` in `pkg/monitor/data.go` |
| **Activity** (bottom) | Unified feed of logs, actions, and comments with timestamps | `fetchActivity()` in `pkg/monitor/data.go` |

**Source**: `/home/devuser/td/pkg/monitor/data.go:28-69` (`FetchData` function)

### Real-Time Data Refresh

The monitor uses a periodic tick-based refresh cycle following the Elm architecture:

1. `scheduleTick()` fires a `TickMsg` at configurable intervals
2. `fetchData()` dispatches as a `tea.Cmd` closure (non-blocking)
3. `FetchData()` queries the database for all panel data
4. `RefreshDataMsg` returns to `Update()` and re-renders the view

The `FetchData` function (`/home/devuser/td/pkg/monitor/data.go:29-69`) aggregates:
- Focused issue via `config.GetFocus()` + `database.GetIssue()`
- In-progress issues via `database.ListIssues(Status: in_progress)`
- Activity feed via `fetchActivity()` (logs + actions + comments, merged and sorted by timestamp)
- Categorized task list via `fetchTaskList()` (respects search query, sort mode, TDQ queries)
- Recent handoffs since monitor start
- Active sessions (activity within last 5 minutes)

### Activity Feed Construction

The activity feed (`/home/devuser/td/pkg/monitor/data.go:72-145`) merges three data sources into a unified timeline:

1. **Logs** (`database.GetRecentLogsAll(50)`) - progress, blocker, decision entries
2. **Actions** (`database.GetRecentActionsAll(50)`) - create, update, delete, start, review, approve, etc.
3. **Comments** (`database.GetRecentCommentsAll(50)`) - issue comments

Items are sorted by timestamp descending and truncated to 50 total. Issue titles are batch-fetched in a single query (`database.GetIssueTitles(issueIDs)`) to avoid N+1 queries.

Action messages are human-readable, converted via `formatActionMessage()` which maps 16 action types to strings like "created issue", "submitted for review", "approved", etc.

### Active Session Detection

The monitor tracks active sessions via `fetchActiveSessions()` (`/home/devuser/td/pkg/monitor/data.go:357-364`), which queries `database.GetActiveSessions(since)` for any sessions with activity in the last 5 minutes. This provides awareness of concurrent agents working on the same project.

---

## 9.2 DB Connection Pool for Embedded Monitors

When the monitor is embedded in an external application (sidecar), BubbleTea's value-copy semantics in `Update()` can cause database connection leaks. The `dbpool.go` module solves this.

**Source**: `/home/devuser/td/pkg/monitor/dbpool.go`

### Problem

Each `Update()` call copies the `Model` value. If the embedder holds a reference to the old model, the `*db.DB` pointer is shared between copies, but Go's `sql.DB` connection pool can grow unbounded, creating hundreds of open file descriptors on the same SQLite file.

### Solution: Singleton Pattern

```
dbPool (global) -> map[resolvedPath] -> sharedDBEntry { db: *db.DB, refs: int }
```

- `getSharedDB(baseDir)` resolves the path (handling worktree redirects), checks the pool, and either returns an existing connection or opens a new one with `MaxOpenConns(1)` for SQLite single-writer semantics.
- `releaseSharedDB(baseDir)` decrements the reference count; when it hits zero, the connection is closed and removed.
- `clearDBPool()` closes all connections (used in tests).

**Key detail**: `database.SetMaxOpenConns(1)` is set explicitly to prevent connection pool growth, which is critical for SQLite's single-writer constraint.

---

## 9.3 Stats and Analytics

### Stats Dashboard (TUI)

The monitor includes a stats modal (opened with `s` key) that shows extended statistics fetched via `FetchStats()` (`/home/devuser/td/pkg/monitor/data.go:427-438`). This calls `database.GetExtendedStats()` and displays the results in a scrollable declarative modal.

### CLI Analytics: `td stats analytics`

**Source**: `/home/devuser/td/cmd/stats_analytics.go`

The `td stats analytics` command (alias: `td stats usage`) provides command-usage analytics with visual charts. Analytics are enabled by default (disable with `TD_ANALYTICS=false`).

**Data collected per command invocation** (logged in `cmd/root.go:101-124` via `logAnalytics()`):
- Command name and subcommand
- Flags used (sanitized - sensitive values stripped)
- Timestamp and duration in milliseconds
- Success/failure status and error message
- Session ID

**Output sections**:
1. **Overview**: Total commands, unique commands, success rate, average duration
2. **Most Used Commands**: Bar chart (top 10, scaled to 30-char bars using `barFilled`/`barEmpty` characters)
3. **Least Used Commands**: Simple count list (bottom 5)
4. **Never Used**: Commands registered but never invoked (detected by comparing usage against `getAllCommandNames()`)
5. **Popular Flags**: Bar chart of most-used flags
6. **Daily Activity**: Last 7 days activity chart using dot characters
7. **Errors by Command**: Commands with the highest error counts
8. **Session Activity**: Top 5 most active sessions by command count

**Flags**:
- `--clear` - Wipe analytics data
- `--json` - JSON output of the full `AnalyticsSummary` struct
- `--since <duration>` - Filter by time (e.g., `7d`, `24h`)
- `--limit <n>` - Max events to analyze

### Security Audit: `td stats security`

**Source**: `/home/devuser/td/cmd/stats_security.go`

Delegates to `securityCmd.RunE` - shows audit log of issues closed using `--self-close-exception`. This tracks when agents bypassed the normal review process.

**Flags**: `--clear`, `--json`

### Error Log: `td stats errors`

**Source**: `/home/devuser/td/cmd/stats_errors.go`

Delegates to `errorsCmd.RunE` - shows failed `td` command attempts for debugging agent issues.

**Flags**: `--clear`, `--count`, `--limit`, `--session`, `--since`, `--json`

---

## 9.4 Agent Activity Tracking

### Database Schema

Agent activity is stored in the `agent_activity` table (added in schema migration v29):

```sql
CREATE TABLE IF NOT EXISTS agent_activity (
    id TEXT PRIMARY KEY,           -- aa-XXXXXXXX random hex
    issue_id TEXT NOT NULL,
    agent_name TEXT NOT NULL,
    activity_type TEXT NOT NULL,   -- assigned, started, committed, etc.
    details TEXT DEFAULT '',
    session_id TEXT DEFAULT '',
    worktree_path TEXT DEFAULT '',
    branch TEXT DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

**Source**: `/home/devuser/td/internal/db/agent_activity.go`

### Activity Types

Eight activity types are tracked (defined in `internal/models/models.go`):
- `assigned` - Issue assigned to an agent
- `started` - Agent started work
- `committed` - Agent committed code
- `pr_created` - Agent created a pull request
- `reviewed` - Agent reviewed work
- `completed` - Agent completed work
- `spawned_subagent` - Agent spawned a sub-agent
- `comment` - Agent left a comment

### CLI: `td activity`

**Source**: `/home/devuser/td/cmd/activity.go`

Two modes:
1. **By issue**: `td activity td-abc123` - Shows all agent activity for a specific issue
2. **By agent**: `td activity --agent oracle` - Shows all activity by a named agent

Output includes timestamp, agent name, activity type, branch, and details.

**Flags**: `--agent`, `--limit/-n` (default 20), `--json`

### Recording Activity

The `AssignIssue()` function (`/home/devuser/td/internal/db/agent_activity.go:115-143`) atomically updates the issue's assignee and records an `assigned` activity entry within a single write lock. Unassignment is recorded with agent name "system" and details "Unassigned".

**Note**: The exported `RecordAgentActivity()` function exists but is dead code (zero callers). `AssignIssue()` directly inserts into the table with its own SQL instead of calling it.

---

## 9.5 Doctor / Diagnostics Command

**Source**: `/home/devuser/td/cmd/doctor.go`

The `td doctor` command runs a sequential diagnostic check for sync setup. It is feature-gated behind the `sync_cli` flag.

### Diagnostic Checks (in order)

| # | Check | How | Pass | Fail |
|---|-------|-----|------|------|
| 1 | Auth config | `syncconfig.LoadAuth()` | Shows email | Shows error |
| 2 | Server reachable | `client.HealthCheck()` (hits `/healthz`) | Shows server URL | Shows error |
| 3 | Auth valid | `client.ListProjects()` | OK | "invalid or expired API key" |
| 4 | Local database | `db.Open(baseDir)` | OK | Shows error |
| 5 | Sync linked | `database.GetSyncState()` | Shows project ID | "not linked to a project" |
| 6 | Pending events | `database.CountPendingEvents()` | Shows count (0 = clean) | Shows error |

Checks are sequential with dependencies: auth valid is skipped if auth config or server check fails; sync linked and pending events are skipped if local database fails.

Output uses aligned dot leaders for readability:
```
Auth config ............ OK (user@example.com)
Server reachable ....... OK (https://sync.example.com)
Auth valid ............. OK
Local database ......... OK
Sync linked ............ OK (project abc123)
Pending events ......... 0
```

---

## 9.6 Sync Server Observability

The `td-sync` server (separate binary at `cmd/td-sync/main.go`) provides its own observability:

### Health Endpoint

`GET /healthz` - Pings the server database and returns `{"status": "ok"}` or `{"status": "error", "detail": "db unreachable"}`.

### Metrics Endpoint

`GET /metricz` - Returns a snapshot of server metrics from the `Metrics` struct. The server tracks request counts, latencies, and error rates via `metricsMiddleware`.

### Structured Logging

The server uses Go's `slog` package with configurable format and level:
- Format: JSON (default) or text, controlled by `SYNC_LOG_FORMAT`
- Level: debug/info/warn/error, controlled by `SYNC_LOG_LEVEL`

### Background Cleanup

Three background goroutines run periodic maintenance:
1. **Expired auth requests**: Every 5 minutes, logs "expired" auth events then cleans up
2. **Old auth events**: Every hour, removes auth events older than `AuthEventRetention` (default: 90 days)
3. **Rate limit events**: Every hour, removes rate limit events older than `RateLimitEventRetention` (default: 30 days)

### Admin API

The admin API (`/v1/admin/*`) provides deep inspection endpoints:
- `GET /v1/admin/server/overview` - Server overview
- `GET /v1/admin/server/config` - Running configuration
- `GET /v1/admin/server/rate-limit-violations` - Rate limit violation history
- `GET /v1/admin/users` / `GET /v1/admin/users/{id}` - User management
- `GET /v1/admin/auth/events` - Auth event log
- `GET /v1/admin/projects/{id}/sync/status` - Per-project sync status
- `GET /v1/admin/projects/{id}/sync/cursors` - Sync cursor positions
- `GET /v1/admin/projects/{id}/events` - Event log browser

All admin endpoints require authentication with scoped API keys (e.g., `admin:read:server`, `admin:read:projects`, `admin:read:events`, `admin:read:snapshots`).

---

## 9.7 Client-Side Logging

### TD_LOG_FILE

Setting the `TD_LOG_FILE` environment variable redirects `slog` output to a file. This is primarily useful for debugging auto-sync errors while `td monitor` is running (since the TUI consumes stdout).

**Source**: `/home/devuser/td/cmd/root.go:52-64`

```go
func initLogFile() *os.File {
    path := os.Getenv("TD_LOG_FILE")
    if path == "" { return nil }
    f, _ := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
    slog.SetDefault(slog.New(slog.NewTextHandler(f, &slog.HandlerOptions{Level: slog.LevelDebug})))
    return f
}
```

### Agent Error Logging

Failed command invocations are logged via `logAgentError()` (`cmd/root.go:136-158`). This records the full argument list, error message, and session ID for analysis via `td stats errors`.

---

## 9.8 Observability Gaps

| Gap | Description | Impact |
|-----|-------------|--------|
| No metrics export | Client has no Prometheus/StatsD/OTLP export | Cannot integrate with external monitoring stacks |
| Silent error discarding in TUI | `fetchIssueDetails()` silently discards errors from 7 DB calls | Modal shows partial data without warning |
| No alerting | No notification when agents fail, sessions go stale, or sync falls behind | Requires manual monitoring |
| No tracing | No distributed tracing between client sync and server | Difficult to debug sync timing issues |
| Limited sync server metrics | `/metricz` exists but content/format is not documented in code | Unknown what metrics are exposed |
