# Section 05: Architecture Design

## 5.1 Overview

`td` is a Go CLI/TUI task management tool with 332 files across ~126K lines of code. It follows a layered architecture with clear package boundaries: entry points delegate to CLI commands, which invoke business logic, which persists through a SQLite data access layer. A BubbleTea-based TUI provides a rich terminal interface, and a sync engine enables multi-device replication via event sourcing.

**Go version:** 1.25.5
**Module path:** `github.com/marcus/td`

---

## 5.2 Package Dependency Map

### Architectural Layers

```
+------------------------------------------------------------+
|                     ENTRY POINTS                            |
|  main.go              cmd/td-sync/main.go                  |
+------------------------------------------------------------+
         |                        |
         v                        v
+------------------+    +---------------------+
|   CLI LAYER      |    |   SERVER LAYER      |
|   cmd/ (83 files)|    | internal/api/       |
|   13,529 lines   |    | internal/serverdb/  |
+------------------+    +---------------------+
    |     |                   |
    v     v                   v
+------------------+    +---------------------+
| TUI LAYER        |    | BUSINESS LOGIC      |
| pkg/monitor/     |    | internal/query/     |
| (34 files)       |    | internal/workflow/  |
| 13,026 lines     |    | internal/session/   |
+------------------+    | internal/sync/      |
    |                   +---------------------+
    v                        |
+----------------------------+--+
|       DATA ACCESS             |
|    internal/db/ (26 files)    |
|    7,835 lines                |
+-------------------------------+
         |
         v
+-------------------------------+
|       DOMAIN MODELS           |
|    internal/models/ (435 LOC) |
|    Zero internal dependencies |
+-------------------------------+
```

### Full Dependency Graph by Layer

| Layer | Packages | Depends On |
|-------|----------|------------|
| **Entry Points** | `main`, `cmd/td-sync` | `cmd`, `internal/db`, `internal/query` |
| **CLI Layer** | `cmd/` | `internal/db`, `internal/models`, `internal/session`, `internal/output`, `internal/config`, `internal/git`, `internal/suggest`, `internal/workdir`, `internal/query`, `internal/workflow`, `internal/syncclient`, `internal/syncconfig`, `internal/features`, `pkg/monitor` |
| **TUI Layer** | `pkg/monitor`, `pkg/monitor/keymap`, `pkg/monitor/modal`, `pkg/monitor/mouse` | `internal/db`, `internal/models`, `internal/session`, `internal/config`, `internal/query`, `internal/syncclient`, `internal/syncconfig`, `internal/features`, `internal/agent`, `internal/version` |
| **Business Logic** | `internal/query`, `internal/workflow`, `internal/session`, `internal/sync` | `internal/db`, `internal/models`, `internal/git` |
| **Server** | `internal/api`, `internal/serverdb` | `internal/sync`, `internal/db`, `internal/crypto`, `internal/query` |
| **Data Access** | `internal/db` | `internal/models`, `internal/workdir` |
| **Infrastructure** | `internal/config`, `internal/git`, `internal/workdir`, `internal/output`, `internal/crypto`, `internal/agent`, `internal/suggest`, `internal/version`, `internal/features`, `internal/syncclient`, `internal/syncconfig` | `internal/models` |
| **Domain Models** | `internal/models` | *(none)* |

### Import Cycle Resolution

The `db` package needs to validate TDQ queries (for board creation), but importing `query` from `db` would create a cycle (`query` already imports `db` for `ListIssuesOptions`). This is resolved in `main.go` via dependency injection through a function variable:

```go
// main.go:17-21 (verified)
func init() {
    db.QueryValidator = func(queryStr string) error {
        _, err := query.Parse(queryStr)
        return err
    }
}
```

The `db.QueryValidator` variable is declared in `internal/db/db.go:17` as `var QueryValidator func(queryStr string) error`.

### Key Interfaces

| Interface | Package | Implementors | Purpose |
|-----------|---------|-------------|---------|
| `QuerySource` | `internal/query` | `db.DB`, `api.snapshotQuerySource` | Abstracts DB for query engine; enables server-side query on snapshots |
| `NoteQuerySource` | `internal/query` | `db.DB` | Abstracts note-related DB operations for TDQ note queries |
| `Guard` | `internal/workflow` | `DifferentReviewerGuard`, `HandoffRequiredGuard` | Pluggable transition guards for workflow state machine |

**Verified from source:** `internal/query/source.go` defines `QuerySource` with 9 methods including `ListIssues`, `GetIssue`, `GetLogs`, `GetComments`, `GetLatestHandoff`, `GetLinkedFiles`, `GetDependencies`, `GetRejectedInProgressIssueIDs`, and `GetIssuesWithOpenDeps`.

---

## 5.3 CLI Architecture: Cobra Command Pattern

### Entry Point

`main.go` resolves the version (build flags, VCS info, or "dev"), sets it on the root command, and calls `cmd.Execute()`:

```go
// main.go:65-68 (verified)
func main() {
    cmd.SetVersion(effectiveVersion(Version))
    cmd.Execute()
}
```

### Root Command and Lifecycle Hooks

The root command (`cmd/root.go:34-49`, verified) defines persistent pre-run and post-run hooks that run on every command:

- **`PersistentPreRun`**: Records start time, runs `runGatedSyncStartupHook` (auto-pulls from sync server if configured)
- **`PersistentPostRun`**: Captures the executed command reference, runs `runGatedSyncMutationHook` (auto-pushes if configured)

### Command Groups

Commands are organized into 7 groups (`cmd/root.go:304-312`, verified):

| Group | ID | Commands |
|-------|----|----------|
| Core | `core` | `create`, `list`, `show`, `update`, `delete`, `search`, `tree` |
| Workflow | `workflow` | `start`, `handoff`, `review`, `approve`, `reject`, `block`, `unblock`, `close`, `reopen` |
| Query | `query` | `query`, `filter` |
| Shortcuts | `shortcuts` | `task`, `bug`, `feature`, `epic`, `chore` |
| Session | `session` | `session`, `focus`, `ws` (work session), `context` |
| Files | `files` | `files`, `link`, `unlink` |
| System | `system` | `init`, `system`, `config`, `version`, `usage`, `security`, `sync`, `undo`, `monitor` |

### Command Execution Flow

```
User runs "td create ..."
    |
    v
main.go:cmd.Execute()
    |
    v
Cobra routes to cmd/create.go RunE
    |
    v
PersistentPreRun: record start time, sync pull (if configured)
    |
    v
RunE handler:
  1. db.Open(baseDir) -- opens SQLite, runs migrations
  2. session.Get(db)  -- resolves current session (branch + agent)
  3. db.CreateIssueLogged() -- atomically writes issue + action_log
  4. output.Success() -- formatted terminal output
    |
    v
PersistentPostRun: sync push (if configured), capture command for analytics
    |
    v
Execute() logs analytics, handles errors with workflow hints
```

### Analytics and Error Handling

`cmd/root.go:66-98` (verified) shows that `Execute()` performs:
1. Analytics logging via `logAnalytics(err)` using `db.LogCommandUsage`
2. Agent error logging via `logAgentError` for failed commands
3. Unknown flag suggestions via `handleUnknownFlagError` with fuzzy matching
4. Workflow hints via `handleWorkflowHint` (e.g., "done" suggests "review")

---

## 5.4 TUI Architecture: Elm/BubbleTea

### Model-Update-View Pattern

The TUI monitor (`pkg/monitor/model.go`, verified) follows the Elm architecture mandated by BubbleTea:

- **Model** (`model.go:22-203`): Single struct with ~90 exported fields covering all UI state
- **Init** (`model.go:361-376`): Returns initial batch of commands (fetch data, schedule tick, restore board, restore filters, check first run, version check)
- **Update** (`model.go:422-845`): Dispatches on `tea.Msg` type with 25+ cases
- **View** (`model.go:854-856`): Delegates to `renderView()` in `view.go`

### Message Types

The Update function handles these verified message types:

| Message | Source | Purpose |
|---------|--------|---------|
| `tea.KeyMsg` | BubbleTea | Keyboard input, delegated to `handleKey()` |
| `tea.WindowSizeMsg` | BubbleTea | Terminal resize, updates panel bounds |
| `tea.MouseMsg` | BubbleTea | Mouse input, delegated to `handleMouse()` |
| `TickMsg` | `scheduleTick()` | Periodic data refresh (configurable interval) |
| `RefreshDataMsg` | `fetchData()` | Updates panel data (focused issue, task list, activity) |
| `IssueDetailsMsg` | `fetchIssueDetails()` | Modal issue detail data |
| `MarkdownRenderedMsg` | `renderMarkdownAsync()` | Async-rendered markdown for modals |
| `StatsDataMsg` | `fetchStats()` | Statistics modal data |
| `HandoffsDataMsg` | `fetchHandoffs()` | Handoffs modal data |
| `BoardIssuesMsg` | `fetchBoardIssues()` | Board view issue data |
| `RestoreLastBoardMsg` | `restoreLastViewedBoard()` | Restores board state on launch |
| `RestoreFilterMsg` | `restoreFilterState()` | Restores search/sort/filter on launch |
| `OpenIssueByIDMsg` | External (sidecar) | Programmatic modal opening |

### Command Dispatch

Key handling flows through a centralized dispatch chain (`commands.go`, verified):

1. **Form mode**: Forward all messages to `huh` form
2. **Board editor mode**: Forward non-key messages to textarea/textinput
3. **Close confirm mode**: Forward non-key messages to textinput
4. **Search mode**: Forward non-key messages to textinput
5. **`handleKey()`** (`commands.go:203-434`): Modal-type-specific dispatch (sync prompt, getting started, TDQ help, stats, handoffs, board editor, board picker, delete/close confirm, search)
6. **`executeCommand()`** (`commands.go:437-1329`): Centralized switch on ~60 keymap commands

### Keymap Registry

`pkg/monitor/keymap/registry.go` (verified) implements a context-aware key binding system:

- **19 UI contexts**: global, main, modal, stats, search, confirm, epic-tasks, parent-epic-focused, blocked-by-focused, blocks-focused, handoffs, form, help, board-picker, board, getting-started, tdq-help, board-editor, board-kanban, close-confirm, sync-prompt
- **Multi-key sequences**: Support for `gg` (go to top) with 500ms timeout
- **Precedence chain**: User overrides (from `keymap.json`) > context bindings > global bindings
- **Thread-safe**: Uses `sync.RWMutex` for concurrent access

### Modal System

The declarative modal system (`pkg/monitor/modal/`, verified) provides:

- **Section types**: Text, Spacer, Buttons, Checkbox, Custom, When, List, Input
- **Automatic hit regions**: For mouse interaction
- **Focus cycling**: Tab/Shift-Tab between focusable sections
- **Scroll viewport**: Internal scroll management with clamping
- **Modal stack**: Push/pop with breadcrumb trail for stacked navigation (e.g., epic -> child task -> child's dependency)

### Panel Layout

The main view renders three vertically-stacked, resizable panels:

```
+-----------------------------------+
| Current Work Panel                |
| (focused issue + in-progress)     |
+-----------------------------------+ <-- drag-to-resize divider
| Task List Panel                   |
| (categorized or board view)       |
+-----------------------------------+ <-- drag-to-resize divider
| Activity Log Panel                |
| (logs, actions, comments table)   |
+-----------------------------------+
```

Panel heights are stored as `[3]float64` ratios (sum=1.0), persisted to `config.json`. Drag-to-resize via mouse divider interaction.

### Embedded API

The monitor is in `pkg/` (not `internal/`) specifically to support embedding:

- `NewEmbedded(baseDir, interval, version)` -- uses shared DB pool
- `NewEmbeddedWithOptions(EmbeddedOptions)` -- adds custom `PanelRenderer`, `ModalRenderer`, `MarkdownThemeConfig`
- Shared DB pool (`dbpool.go`) prevents connection leaks from BubbleTea's value-copy semantics
- Reference counting ensures the actual connection is closed only when all references are released

---

## 5.5 Sync Architecture

### Event Sourcing Model

All local mutations use "logged" DB operation variants that atomically write to both the entity table and the `action_log` table within a write lock:

```
Local mutation flow:
  CLI/TUI action
    -> db.CreateIssueLogged() (or UpdateIssueLogged, DeleteIssueLogged, etc.)
    -> withWriteLock()
      -> INSERT/UPDATE into entity table
      -> INSERT into action_log (before/after JSON snapshots)
    -> release lock
```

Remote events use "unlogged" variants to avoid double-logging:

```
Remote event application:
  Pull from server
    -> ApplyRemoteEvents()
    -> db.CreateIssue() (unlogged -- no action_log entry)
```

This split is implemented across multiple files in `internal/db/`:
- `issues.go` -- unlogged CRUD operations
- `issues_logged.go` -- logged variants with `action_log` entries

### Sync Protocol

**Push flow** (verified in `internal/sync/client.go:62-147`):

1. `GetPendingEvents()` reads unsynced `action_log` rows (backfills orphan entities first)
2. Events sent as POST to `/v1/projects/:id/sync/push`
3. Server assigns monotonic `server_seq` numbers via `InsertServerEvents()` (verified in `engine.go:36-111`)
4. Server returns `PushResult` with `Acks` (client_action_id -> server_seq mapping)
5. Client calls `MarkEventsSynced()` to update `action_log.synced_at` and `server_seq`

**Pull flow** (verified in `internal/sync/client.go:149-254`):

1. `GET /v1/projects/:id/sync/pull?after=N` returns events since cursor
2. `ApplyRemoteEvents()` applies each event via unlogged DB operations
3. Conflict detection: if local row was modified after `lastSyncAt`, records conflict in `sync_conflicts` table
4. Updates `sync_state.last_pulled_server_seq`

### Event Structure

From `internal/sync/types.go:9-19` (verified):

```go
type Event struct {
    ClientActionID  int64
    DeviceID        string
    SessionID       string
    ActionType      string      // create, update, soft_delete, delete, restore
    EntityType      string      // issues, handoffs, boards, logs, comments, etc.
    EntityID        string
    Payload         []byte      // JSON with schema_version, new_data, previous_data
    ClientTimestamp time.Time
    ServerSeq       int64       // assigned by server
}
```

### Conflict Resolution

The sync engine uses **last-writer-wins** conflict resolution:

1. Server is a thin event store -- it accepts all events and assigns sequence numbers
2. During pull, `ApplyRemoteEvents()` overwrites local data with remote data
3. If the local row was modified after the last sync timestamp, a `ConflictRecord` is stored in `sync_conflicts` for auditing
4. Deduplication uses `UNIQUE(device_id, session_id, client_action_id)` -- duplicate pushes are silently rejected with their existing `server_seq` returned

### Entity Type Normalization

The sync client normalizes entity types (verified in `sync/client.go:29-56`):

| Input | Canonical Table |
|-------|----------------|
| `issue` / `issues` | `issues` |
| `handoff` / `handoffs` | `handoffs` |
| `board` / `boards` | `boards` |
| `log` / `logs` | `logs` |
| `comment` / `comments` | `comments` |
| `work_session` / `work_sessions` | `work_sessions` |
| `board_position` / `board_issue_positions` | `board_issue_positions` |
| `dependency` / `issue_dependencies` | `issue_dependencies` |
| `file_link` / `issue_files` | `issue_files` |
| `note` / `notes` | `notes` |

### Auto-Sync

Background sync is configured via `PersistentPreRun`/`PersistentPostRun` hooks in `cmd/autosync.go`:

- **Pre-run**: Pull (on read commands to get latest data)
- **Post-run**: Push (on mutation commands to share changes)
- **TUI backup path**: Independent goroutine in `cmd/monitor.go` for when BubbleTea Cmd dispatch stalls under certain PTYs

---

## 5.6 Database Architecture

### Storage

- **Single-file SQLite** at `<project>/.todos/issues.db`
- **WAL mode** for concurrent reads (`PRAGMA journal_mode=WAL`)
- **Single connection** (`MaxOpenConns(1)`) to prevent corruption
- **Busy timeout**: 5 seconds (`PRAGMA busy_timeout=5000`)
- **Synchronous mode**: `NORMAL` (balanced speed/safety with WAL)

### Write Locking

Cross-process write safety via OS file locks (`internal/db/lock.go`, verified):

```
withWriteLock(timeout)
  -> writeLocker.acquire(timeout)
    -> os.OpenFile("db.lock")
    -> tryLock() (flock on Unix, LockFileEx on Windows)
    -> exponential backoff (5ms initial, 50ms cap)
    -> writeHolder() (PID + timestamp for diagnostics)
  -> execute callback
  -> writeLocker.release()
    -> truncate holder info
    -> unlock()
    -> close file
```

Stale lock detection reads PID from the lock file and checks if the process is alive.

### Schema Migration

29 sequential migrations (`internal/db/schema.go`, verified), with a mix of:
- **SQL-only migrations** (e.g., `ALTER TABLE`, `CREATE TABLE`)
- **Custom Go migrations** (v13, v14, v15, v18, v19, v20, v24, v25, v26) for operations that SQLite DDL cannot express (table recreation, data migration, format changes)

Notable migrations:
- **v15**: Migrated integer primary keys to text IDs for sync compatibility
- **v18**: Added deterministic ID columns to composite-key tables for sync
- **v22**: Sparse positioning with 65536 gap for O(1) board issue ordering
- **v23**: Dropped UNIQUE(name) on boards to prevent sync data loss

### Git Worktree Support

`internal/workdir/` (verified) enables multiple git worktrees to share a single database:
- A `.td-root` file in a worktree contains the path to the main repo root
- `ResolveBaseDir()` checks for this file and redirects to the main repo's `.todos` directory

---

## 5.7 Design Patterns Summary

| Pattern | Location | Description |
|---------|----------|-------------|
| **Elm Architecture** | `pkg/monitor/model.go` | Model/Update/View with typed messages and commands |
| **Command Pattern** | `cmd/*.go` | One Cobra `*cobra.Command` per file with `RunE` handler |
| **Repository Pattern** | `internal/db/db.go` | `DB` struct encapsulates all SQLite operations |
| **State Machine** | `internal/workflow/workflow.go` | Configurable transition modes (Liberal/Advisory/Strict) with pluggable guards |
| **Event Sourcing** | `internal/sync/` | Action log captures all mutations with before/after snapshots |
| **Logged/Unlogged Split** | `internal/db/issues_logged.go` | Local mutations write action_log; remote events do not |
| **Dependency Injection** | `main.go:17-21` | Function variable breaks import cycle between `db` and `query` |
| **Shared Pool** | `pkg/monitor/dbpool.go` | Reference-counted DB pool for BubbleTea value-copy semantics |
| **Declarative UI** | `pkg/monitor/modal/` | Composable sections with auto-focus and scroll management |
| **Context-Aware Keymap** | `pkg/monitor/keymap/registry.go` | Multi-key sequences, user overrides, 19 UI contexts |
| **Sparse Positioning** | `internal/db/schema.go` (v22) | 65536-gap integer positions for O(1) board reordering |
| **Cross-Platform Locking** | `internal/db/lock.go` | `lock_unix.go` (flock) / `lock_windows.go` (LockFileEx) |

---

## 5.8 Architectural Risks

| Severity | Risk | Location | Recommendation |
|----------|------|----------|----------------|
| **High** | Model struct has ~200 fields covering 15+ UI states | `pkg/monitor/model.go:22-203` | Extract sub-models: `BoardState`, `FormState`, `SearchState`, `SyncState` |
| **High** | Duplicated 24-column issue scan code across 4 locations | `internal/db/issues.go`, `issues_logged.go` | Extract shared `scanIssue()` helper |
| **Medium** | `executeCommand()` is a 900-line switch statement with ~60 cases | `pkg/monitor/commands.go:437-1329` | Split into handler groups; consider dispatch table |
| **Medium** | Manual SQL building in `ListIssues` with 33 filter conditions | `internal/db/issues.go:317-571` | Extract filter composition helpers |
| **Medium** | TDQ evaluates all issues in-memory (10K cap) | `internal/query/execute.go` | Push simple field queries to SQL WHERE |
| **Medium** | Monitor depends on 10+ internal packages | `pkg/monitor/commands.go` imports | Introduce service/facade layer |
| **Low** | Legacy state fields coexist with declarative modal system | `pkg/monitor/model.go:76-94` | Complete migration to declarative modals |
