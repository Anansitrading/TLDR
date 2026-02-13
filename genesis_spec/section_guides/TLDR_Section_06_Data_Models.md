# Section 06: Data Models

## 6.1 Overview

The `td` data model is defined in `internal/models/models.go` (435 lines, verified) as a pure domain package with zero internal dependencies. The database schema is defined in `internal/db/schema.go` (502 lines, verified) with 20 tables across schema version 29, evolved through 28 sequential migrations.

---

## 6.2 Core Domain Types

### Issue

**Source:** `internal/models/models.go:78-103` (verified)

The central entity. Represents a task, bug, feature, epic, or chore.

```go
type Issue struct {
    ID                 string     `json:"id"`
    Title              string     `json:"title"`
    Description        string     `json:"description,omitempty"`
    Status             Status     `json:"status"`
    Type               Type       `json:"type"`
    Priority           Priority   `json:"priority"`
    Points             int        `json:"points"`
    Labels             []string   `json:"labels,omitempty"`
    ParentID           string     `json:"parent_id,omitempty"`
    Acceptance         string     `json:"acceptance,omitempty"`
    Sprint             string     `json:"sprint,omitempty"`
    Assignee           string     `json:"assignee,omitempty"`
    LinearID           string     `json:"linear_id,omitempty"`
    LinearIdentifier   string     `json:"linear_identifier,omitempty"`
    ProjectTag         string     `json:"project_tag,omitempty"`
    ImplementerSession string     `json:"implementer_session"`
    CreatorSession     string     `json:"creator_session"`
    ReviewerSession    string     `json:"reviewer_session"`
    CreatedAt          time.Time  `json:"created_at"`
    UpdatedAt          time.Time  `json:"updated_at"`
    ClosedAt           *time.Time `json:"closed_at,omitempty"`
    DeletedAt          *time.Time `json:"deleted_at,omitempty"`
    Minor              bool       `json:"minor"`
    CreatedBranch      string     `json:"created_branch,omitempty"`
}
```

**Key fields:**
- `ID`: Text primary key (migrated from integer in v15 for sync compatibility)
- `ParentID`: Self-referential foreign key enabling epic-child hierarchy
- `ImplementerSession`, `CreatorSession`, `ReviewerSession`: Session IDs tracking who implemented, created, and reviewed the issue (used for review bypass prevention)
- `Minor`: Boolean flag indicating self-reviewable tasks (no different-session review required)
- `DeletedAt`: Soft delete support (nil = active, non-nil = deleted)
- `Labels`: Stored as comma-separated text in SQLite, exposed as `[]string` in Go
- `Points`: Fibonacci story points (1, 2, 3, 5, 8, 13, 21)
- `Assignee`, `LinearID`, `LinearIdentifier`, `ProjectTag`: Agent swarm integration fields (added in v29)

### Log

**Source:** `internal/models/models.go:133-141` (verified)

Session activity log entries attached to issues.

```go
type Log struct {
    ID            string    `json:"id"`
    IssueID       string    `json:"issue_id"`
    SessionID     string    `json:"session_id"`
    WorkSessionID string    `json:"work_session_id,omitempty"`
    Message       string    `json:"message"`
    Type          LogType   `json:"type"`
    Timestamp     time.Time `json:"timestamp"`
}
```

`IssueID` can be empty for work-session-scoped logs (e.g., general progress notes not tied to a specific issue).

### Handoff

**Source:** `internal/models/models.go:144-153` (verified)

Structured state capture for session-to-session handoffs.

```go
type Handoff struct {
    ID        string    `json:"id"`
    IssueID   string    `json:"issue_id"`
    SessionID string    `json:"session_id"`
    Done      []string  `json:"done,omitempty"`
    Remaining []string  `json:"remaining,omitempty"`
    Decisions []string  `json:"decisions,omitempty"`
    Uncertain []string  `json:"uncertain,omitempty"`
    Timestamp time.Time `json:"timestamp"`
}
```

The four list fields (`Done`, `Remaining`, `Decisions`, `Uncertain`) are stored as JSON arrays in SQLite. This structure captures the working state so a new agent/session can resume where the previous one left off.

### WorkSession

**Source:** `internal/models/models.go:184-192` (verified)

Multi-issue work sessions with optional git SHA bookmarks.

```go
type WorkSession struct {
    ID        string     `json:"id"`
    Name      string     `json:"name"`
    SessionID string     `json:"session_id"`
    StartedAt time.Time  `json:"started_at"`
    EndedAt   *time.Time `json:"ended_at,omitempty"`
    StartSHA  string     `json:"start_sha,omitempty"`
    EndSHA    string     `json:"end_sha,omitempty"`
}
```

### Board

**Source:** `internal/models/models.go:211-220` (verified)

Named views into issues defined by TDQ queries.

```go
type Board struct {
    ID           string     `json:"id"`
    Name         string     `json:"name"`
    Query        string     `json:"query"`
    IsBuiltin    bool       `json:"is_builtin"`
    ViewMode     string     `json:"view_mode"`    // "swimlanes", "backlog", or "kanban"
    LastViewedAt *time.Time `json:"last_viewed_at,omitempty"`
    CreatedAt    time.Time  `json:"created_at"`
    UpdatedAt    time.Time  `json:"updated_at"`
}
```

A built-in "All Issues" board is created at migration v10 with an empty query (matches all issues). User-created boards use TDQ queries to define their scope.

### BoardIssue and BoardIssueView

**Source:** `internal/models/models.go:223-237` (verified)

```go
type BoardIssue struct {
    BoardID  string    `json:"board_id"`
    IssueID  string    `json:"issue_id"`
    Position int       `json:"position"`
    AddedAt  time.Time `json:"added_at"`
}

type BoardIssueView struct {
    BoardID     string `json:"board_id"`
    Position    int    `json:"position"`
    HasPosition bool   `json:"has_position"`
    Issue       Issue  `json:"issue"`
    Category    string `json:"category"`    // Computed: ready/blocked/reviewable/etc
}
```

Positions use sparse integer keys with a gap of 65536 (set in migration v22), enabling O(1) insertions without reordering all rows.

### Comment

**Source:** `internal/models/models.go:240-246` (verified)

```go
type Comment struct {
    ID        string    `json:"id"`
    IssueID   string    `json:"issue_id"`
    SessionID string    `json:"session_id"`
    Text      string    `json:"text"`
    CreatedAt time.Time `json:"created_at"`
}
```

### Note

**Source:** `internal/models/models.go:249-258` (verified)

Freeform notes synced via sidecar (not attached to issues).

```go
type Note struct {
    ID        string     `json:"id"`
    Title     string     `json:"title"`
    Content   string     `json:"content"`
    CreatedAt time.Time  `json:"created_at"`
    UpdatedAt time.Time  `json:"updated_at"`
    Pinned    bool       `json:"pinned"`
    Archived  bool       `json:"archived"`
    DeletedAt *time.Time `json:"deleted_at,omitempty"`
}
```

### AgentActivity

**Source:** `internal/models/models.go:120-130` (verified)

Tracks agent actions for swarm coordination.

```go
type AgentActivity struct {
    ID           string            `json:"id"`
    IssueID      string            `json:"issue_id"`
    AgentName    string            `json:"agent_name"`
    ActivityType AgentActivityType `json:"activity_type"`
    Details      string            `json:"details,omitempty"`
    SessionID    string            `json:"session_id,omitempty"`
    WorktreePath string            `json:"worktree_path,omitempty"`
    Branch       string            `json:"branch,omitempty"`
    CreatedAt    time.Time         `json:"created_at"`
}
```

### ActionLog

**Source:** `internal/models/models.go:310-320` (verified)

The undo/sync backbone. Every local mutation is recorded here.

```go
type ActionLog struct {
    ID           string     `json:"id"`
    SessionID    string     `json:"session_id"`
    ActionType   ActionType `json:"action_type"`
    EntityType   string     `json:"entity_type"`
    EntityID     string     `json:"entity_id"`
    PreviousData string     `json:"previous_data"`  // JSON snapshot before action
    NewData      string     `json:"new_data"`        // JSON snapshot after action
    Timestamp    time.Time  `json:"timestamp"`
    Undone       bool       `json:"undone"`
}
```

---

## 6.3 Supporting Types

### GitSnapshot

**Source:** `internal/models/models.go:156-164` (verified)

Captures git state at workflow events (start, handoff).

```go
type GitSnapshot struct {
    ID         string    `json:"id"`
    IssueID    string    `json:"issue_id"`
    Event      string    `json:"event"`      // "start", "handoff"
    CommitSHA  string    `json:"commit_sha"`
    Branch     string    `json:"branch"`
    DirtyFiles int       `json:"dirty_files"`
    Timestamp  time.Time `json:"timestamp"`
}
```

### IssueFile

**Source:** `internal/models/models.go:167-174` (verified)

Links files to issues with role classification.

```go
type IssueFile struct {
    ID        string    `json:"id"`
    IssueID   string    `json:"issue_id"`
    FilePath  string    `json:"file_path"`
    Role      FileRole  `json:"role"`
    LinkedSHA string    `json:"linked_sha"`
    LinkedAt  time.Time `json:"linked_at"`
}
```

### IssueDependency

**Source:** `internal/models/models.go:177-181` (verified)

```go
type IssueDependency struct {
    IssueID      string `json:"issue_id"`
    DependsOnID  string `json:"depends_on_id"`
    RelationType string `json:"relation_type"`  // "blocks", "depends_on"
}
```

### IssueSessionHistory

**Source:** `internal/models/models.go:195-201` (verified)

Audit trail of which sessions touched an issue.

```go
type IssueSessionHistory struct {
    ID        string             `json:"id"`
    IssueID   string             `json:"issue_id"`
    SessionID string             `json:"session_id"`
    Action    IssueSessionAction `json:"action"`
    CreatedAt time.Time          `json:"created_at"`
}
```

### Config

**Source:** `internal/models/models.go:261-274` (verified)

Local configuration state.

```go
type Config struct {
    FocusedIssueID    string          `json:"focused_issue_id,omitempty"`
    ActiveWorkSession string          `json:"active_work_session,omitempty"`
    PaneHeights       [3]float64      `json:"pane_heights,omitempty"`
    FeatureFlags      map[string]bool `json:"feature_flags,omitempty"`
    SearchQuery       string          `json:"search_query,omitempty"`
    SortMode          string          `json:"sort_mode,omitempty"`
    TypeFilter        string          `json:"type_filter,omitempty"`
    IncludeClosed     bool            `json:"include_closed,omitempty"`
    TitleMinLength    int             `json:"title_min_length,omitempty"`
    TitleMaxLength    int             `json:"title_max_length,omitempty"`
}
```

### ExtendedStats

**Source:** `internal/models/models.go:412-435` (verified)

Aggregated statistics for the stats modal.

```go
type ExtendedStats struct {
    Total, CreatedToday, CreatedThisWeek int
    ByStatus   map[Status]int
    ByType     map[Type]int
    ByPriority map[Priority]int
    OldestOpen, NewestTask, LastClosed *Issue
    TotalPoints int
    AvgPointsPerTask, CompletionRate float64
    TotalLogs, TotalHandoffs int
    MostActiveSession string
}
```

---

## 6.4 Enum Types and State Machines

### Status

**Source:** `internal/models/models.go:11-19` (verified)

```
open -> in_progress -> in_review -> closed
  |         |             |
  |         v             |
  |      blocked          |
  |         |             |
  |         v             v
  +-----> (any) <-----+
```

| Value | Constant | Description |
|-------|----------|-------------|
| `open` | `StatusOpen` | Initial state, not yet started |
| `in_progress` | `StatusInProgress` | Active work |
| `blocked` | `StatusBlocked` | Waiting on dependency or external factor |
| `in_review` | `StatusInReview` | Submitted for review |
| `closed` | `StatusClosed` | Completed |

**Normalization** (`models.go:398-408`, verified): Hyphens converted to underscores (`in-progress` -> `in_progress`), `review` aliased to `in_review`.

### Type

**Source:** `internal/models/models.go:22-30` (verified)

| Value | Constant | Description |
|-------|----------|-------------|
| `bug` | `TypeBug` | Defect |
| `feature` | `TypeFeature` | New capability (alias: `story`) |
| `task` | `TypeTask` | Generic work item |
| `epic` | `TypeEpic` | Parent container for child issues |
| `chore` | `TypeChore` | Maintenance/housekeeping |

**Normalization** (`models.go:388-394`, verified): `story` aliased to `feature`.

### Priority

**Source:** `internal/models/models.go:33-41` (verified)

| Value | Constant | Aliases |
|-------|----------|---------|
| `P0` | `PriorityP0` | `0`, `critical`, `highest` |
| `P1` | `PriorityP1` | `1`, `high` |
| `P2` | `PriorityP2` | `2`, `medium`, `normal`, `default` |
| `P3` | `PriorityP3` | `3`, `low` |
| `P4` | `PriorityP4` | `4`, `lowest`, `none` |

**Normalization** (`models.go:367-383`, verified): Case-insensitive matching, numeric aliases (`0`-`4`), word forms.

### LogType

**Source:** `internal/models/models.go:44-55` (verified)

| Value | Description |
|-------|-------------|
| `progress` | General progress update |
| `security` | Security-related note |
| `blocker` | Blocking issue documentation |
| `decision` | Design/architecture decision |
| `hypothesis` | Working theory |
| `tried` | Attempted approach |
| `result` | Outcome of an attempt |
| `orchestration` | Multi-agent coordination note |

### ActionType

**Source:** `internal/models/models.go:277-307` (verified)

27 action types covering all mutation operations:

| Category | Actions |
|----------|---------|
| **Issue lifecycle** | `create`, `update`, `delete`, `restore` |
| **Workflow** | `start`, `review`, `approve`, `reject`, `block`, `unblock`, `close`, `reopen` |
| **Relations** | `add_dependency`, `remove_dependency`, `link_file`, `unlink_file`, `handoff` |
| **Boards** | `board_create`, `board_delete`, `board_update`, `board_add_issue`, `board_remove_issue`, `board_move_issue`, `board_set_position`, `board_unposition` |
| **Work sessions** | `work_session_tag`, `work_session_untag` |

### FileRole

**Source:** `internal/models/models.go:68-75` (verified)

| Value | Description |
|-------|-------------|
| `implementation` | Source code file |
| `test` | Test file |
| `reference` | Documentation/reference |
| `config` | Configuration file |

### AgentActivityType

**Source:** `internal/models/models.go:106-117` (verified)

| Value | Description |
|-------|-------------|
| `assigned` | Issue assigned to agent |
| `started` | Agent began work |
| `committed` | Agent made a commit |
| `pr_created` | Agent created a PR |
| `reviewed` | Agent reviewed work |
| `completed` | Agent finished work |
| `spawned_subagent` | Agent spawned a sub-agent |
| `comment` | Agent left a comment |

### IssueSessionAction

**Source:** `internal/models/models.go:58-65` (verified)

| Value | Description |
|-------|-------------|
| `created` | Session created the issue |
| `started` | Session started work on the issue |
| `unstarted` | Session unstarted the issue |
| `reviewed` | Session reviewed the issue |

### Validation Helpers

**Source:** `internal/models/models.go:323-408` (verified)

- `ValidPoints()` -> `[]int{1, 2, 3, 5, 8, 13, 21}` (Fibonacci)
- `IsValidPoints(p int) bool`
- `IsValidStatus(s Status) bool`
- `IsValidType(t Type) bool`
- `IsValidPriority(p Priority) bool`
- `NormalizePriority(p string) Priority` -- case-insensitive, numeric aliases, word forms
- `NormalizeType(t string) Type` -- `story` -> `feature`
- `NormalizeStatus(s string) Status` -- hyphens to underscores, `review` -> `in_review`

---

## 6.5 Database Schema

### Table Summary

**Source:** `internal/db/schema.go` (verified)

| Table | PK Type | Purpose | Introduced |
|-------|---------|---------|------------|
| `issues` | TEXT | Core task entities | v1 (base schema) |
| `logs` | TEXT | Session activity logs | v1 |
| `handoffs` | TEXT | Structured handoff state | v1 |
| `git_snapshots` | TEXT | Git state at events | v1 |
| `issue_files` | TEXT | File-issue links | v1 |
| `issue_dependencies` | TEXT | Issue relationships | v1 |
| `work_sessions` | TEXT | Multi-issue sessions | v1 |
| `work_session_issues` | TEXT (added v24) | Junction: session-issue | v1 |
| `comments` | TEXT | Issue comments | v1 |
| `sessions` | TEXT | Session tracking | v1 (extended v13/v14) |
| `schema_info` | TEXT | Version tracking | v1 |
| `action_log` | TEXT (fixed v26) | Undo/sync backbone | v2 |
| `boards` | TEXT | Named issue views | v9 |
| `board_issue_positions` | TEXT (added v18) | Board ordering | v9 (renamed v10) |
| `issue_session_history` | TEXT | Session-issue audit | v7 |
| `sync_state` | TEXT | Sync cursor tracking | v16 |
| `sync_conflicts` | INTEGER (auto) | Conflict audit log | v17 |
| `sync_history` | INTEGER (auto) | Sync operation log | v21 |
| `notes` | TEXT | Freeform notes | v28 |
| `agent_activity` | TEXT | Agent action tracking | v29 |

### Full Schema DDL (Base Tables)

**Source:** `internal/db/schema.go:6-154` (verified)

```sql
-- Issues (24 columns)
CREATE TABLE issues (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT DEFAULT '',
    status TEXT NOT NULL DEFAULT 'open',
    type TEXT NOT NULL DEFAULT 'task',
    priority TEXT NOT NULL DEFAULT 'P2',
    points INTEGER DEFAULT 0,
    labels TEXT DEFAULT '',
    parent_id TEXT DEFAULT '',
    acceptance TEXT DEFAULT '',
    implementer_session TEXT DEFAULT '',
    reviewer_session TEXT DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    closed_at DATETIME,
    deleted_at DATETIME,
    minor INTEGER DEFAULT 0,
    created_branch TEXT DEFAULT '',
    -- Added via migrations:
    sprint TEXT DEFAULT '',          -- v10
    creator_session TEXT DEFAULT '', -- v6
    assignee TEXT DEFAULT '',        -- v29
    linear_id TEXT DEFAULT '',       -- v29
    linear_identifier TEXT DEFAULT '',-- v29
    project_tag TEXT DEFAULT '',     -- v29
    FOREIGN KEY (parent_id) REFERENCES issues(id)
);
```

### Indexes

14 indexes on the base schema covering:
- Status, priority, type, parent_id (issues)
- Deleted status (composite: `deleted_at, status`)
- Issue foreign keys (logs, handoffs, git_snapshots, issue_files, comments)
- Session lookup (branch, branch+agent composite)
- Timestamp-based (handoffs, logs, comments, action_log)
- Agent activity (issue, agent_name, activity_type, created_at)
- Assignee, project_tag, linear_id (issues)

---

## 6.6 Entity Relationships

```
                    +------------+
                    |   sessions |
                    +------+-----+
                           |
            +--------------+--------------+
            |              |              |
    creator_session  implementer  reviewer_session
            |         _session         |
            v              |           v
        +------------------+----------+
        |               issues              |
        +---+----+---------+------+----+----+
            |    |         |      |    |
            |    |   parent_id    |    |
            |    |   (self-ref)   |    |
            |    |                |    |
            v    v                v    v
     +------+ +-------+   +--------+ +--------+
     | logs | |handoffs|   |comments| |git_snap|
     +------+ +-------+   +--------+ +--------+
            |
            v
    +----------------+        +-------+
    |issue_files     |        | boards|
    +----------------+        +---+---+
    |issue_deps      |            |
    +----------------+   +--------+-------+
                         |board_issue_pos |
                         +----------------+

    +----------------+        +----------+
    |work_sessions   |        | notes    |
    +------+---------+        +----------+
           |
    +------+---------+
    |work_session_   |
    |  issues        |
    +----------------+

    +----------------+        +----------+
    |issue_session_  |        |agent_    |
    |  history       |        |activity  |
    +----------------+        +----------+

    +----------------+        +----------+
    | action_log     |        |sync_state|
    +----------------+        +----------+
                              |sync_     |
                              |conflicts |
                              +----------+
                              |sync_     |
                              |history   |
                              +----------+
```

### Key Relationships

| From | To | Type | Via |
|------|----|------|-----|
| Issue | Issue | Self-referential (parent-child) | `parent_id` FK |
| Issue | Issue | Dependency graph | `issue_dependencies` junction |
| Issue | Session | Creator/Implementer/Reviewer | `creator_session`, `implementer_session`, `reviewer_session` |
| Log | Issue | Many-to-one | `issue_id` FK |
| Handoff | Issue | Many-to-one | `issue_id` FK |
| Comment | Issue | Many-to-one | `issue_id` FK |
| GitSnapshot | Issue | Many-to-one | `issue_id` FK |
| IssueFile | Issue | Many-to-one | `issue_id` FK + `UNIQUE(issue_id, file_path)` |
| Board | Issue | Many-to-many | `board_issue_positions` junction |
| WorkSession | Issue | Many-to-many | `work_session_issues` junction |
| AgentActivity | Issue | Many-to-one | `issue_id` FK |
| ActionLog | Entity | Polymorphic | `entity_type` + `entity_id` |

### Soft Delete Pattern

Issues use soft deletion via `deleted_at`:
- `NULL` / `nil` = active
- Non-null timestamp = deleted
- Board positions also support soft delete via `deleted_at` (added in v25 for sync)
- Notes support soft delete via `deleted_at`
- `action_log.undone` is a boolean flag, not a timestamp-based soft delete

---

## 6.7 Migration History

| Version | Type | Description |
|---------|------|-------------|
| v1 | Base | Initial schema (13 tables) |
| v2 | SQL | Add `action_log` for undo support |
| v3 | SQL | Allow logs without `issue_id` (work session logs) |
| v4 | SQL | Add `minor` flag to issues |
| v5 | SQL | Add `created_branch` to issues |
| v6 | SQL | Add `creator_session` for review enforcement |
| v7 | SQL | Add `issue_session_history` table |
| v8 | SQL | Add timestamp indexes for activity queries |
| v9 | SQL | Add `boards` and `board_issues` tables |
| v10 | SQL | Query-based boards, sparse ordering, sprint field, built-in "All Issues" board |
| v11 | SQL | Add `view_mode` to boards |
| v12 | SQL | Add performance indexes for monitor queries |
| v13 | **Go** | Extend sessions table for DB-backed storage |
| v14 | **Go** | Repair sessions table (fixes v13 edge cases) |
| v15 | **Go** | Migrate integer PKs to text IDs for sync |
| v16 | SQL | Add `sync_state` table, sync columns on `action_log` |
| v17 | SQL | Add `sync_conflicts` table |
| v18 | **Go** | Add deterministic IDs to composite-key tables for sync |
| v19 | **Go** | Convert absolute file paths to repo-relative |
| v20 | **Go** | Normalize legacy action_log entries for composite-key entities |
| v21 | SQL | Add `sync_history` table |
| v22 | SQL | Sparse positioning (multiply positions by 65536, drop unique index) |
| v23 | SQL | Drop UNIQUE(name) on boards (prevent sync data loss) |
| v24 | **Go** | Add deterministic ID to `work_session_issues` |
| v25 | **Go** | Add `deleted_at` to `board_issue_positions` for soft delete sync |
| v26 | **Go** | Enforce NOT NULL on `action_log.id` |
| v27 | SQL | Normalize NULL session fields to empty strings |
| v28 | SQL | Add `notes` table for sidecar notes sync |
| v29 | SQL | Add agent swarm fields (`assignee`, `linear_id`, etc.) and `agent_activity` table |

---

## 6.8 Sync-Related Data

### Sync State

```sql
CREATE TABLE sync_state (
    project_id TEXT PRIMARY KEY,
    last_pushed_action_id INTEGER DEFAULT 0,
    last_pulled_server_seq INTEGER DEFAULT 0,
    last_sync_at DATETIME,
    sync_disabled INTEGER DEFAULT 0
);
```

Tracks per-project sync cursor position. `last_pushed_action_id` marks the last `action_log` row pushed. `last_pulled_server_seq` marks the server sequence cursor for pull.

### Action Log (Extended for Sync)

The `action_log` table has two additional columns added by v16:
- `synced_at DATETIME` -- timestamp when the event was pushed to the server
- `server_seq INTEGER` -- server-assigned sequence number

### Sync Event (Server-Side)

From `internal/sync/engine.go:12-27` (verified):

```sql
CREATE TABLE events (
    server_seq        INTEGER PRIMARY KEY AUTOINCREMENT,
    device_id         TEXT NOT NULL,
    session_id        TEXT NOT NULL,
    client_action_id  INTEGER NOT NULL,
    action_type       TEXT NOT NULL,
    entity_type       TEXT NOT NULL,
    entity_id         TEXT NOT NULL,
    payload           JSON NOT NULL,
    client_timestamp  DATETIME NOT NULL,
    server_timestamp  DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(device_id, session_id, client_action_id)
);
```

The UNIQUE constraint on `(device_id, session_id, client_action_id)` prevents duplicate event insertion. Duplicates are silently rejected with their existing `server_seq` returned.
