# Section 08: Testing Strategy

## 8.1 Overview

The `td` codebase has strong test infrastructure with 134 test files containing ~67K lines of test code against 198 source files with ~59K lines of production code. This yields an overall **test-to-source line ratio of 1.14:1** -- above average for Go projects. The testing strategy employs four tiers: unit tests, integration tests, sync harness tests, and end-to-end (E2E) tests with real binaries.

**Source:** `02_code_quality.json` findings cross-referenced with actual test file enumeration.

---

## 8.2 Test Architecture

### Four-Tier Test Strategy

```
+--------------------------------------------------+
|  Tier 4: End-to-End Tests (test/e2e/)            |
|  Real td + td-sync binaries in temp directories  |
|  Multi-actor sync scenarios with real HTTP        |
+--------------------------------------------------+
|  Tier 3: Sync Harness Tests (test/syncharness/)  |
|  In-memory SQLite with real sync engine           |
|  Multi-device simulation without HTTP             |
+--------------------------------------------------+
|  Tier 2: Integration Tests                        |
|  Real SQLite DB, multi-package interactions       |
|  cmd/*_test.go, session/integration_test.go       |
+--------------------------------------------------+
|  Tier 1: Unit Tests                               |
|  Isolated package tests with mocks/stubs          |
|  models, query/parser, keymap, modal, workflow    |
+--------------------------------------------------+
```

---

## 8.3 Tier 1: Unit Tests

Unit tests operate within a single package, testing functions and types in isolation.

### Query Engine Tests

**Files (verified):**
- `/home/devuser/td/internal/query/lexer_test.go` -- Tokenizer tests
- `/home/devuser/td/internal/query/parser_test.go` -- AST construction tests
- `/home/devuser/td/internal/query/evaluator_test.go` -- In-memory evaluation tests
- `/home/devuser/td/internal/query/execute_test.go` -- End-to-end query execution
- `/home/devuser/td/internal/query/note_evaluator_test.go` -- Note-specific query tests

**Coverage:** 5 test files for 7 source files (0.71 ratio). Excellent coverage of the lexer, parser, evaluator, and executor pipeline.

### Domain Model Tests

**File:** `/home/devuser/td/internal/models/models_test.go`

Tests validation helpers (`IsValidStatus`, `IsValidType`, `IsValidPriority`, `IsValidPoints`) and normalization functions (`NormalizePriority`, `NormalizeType`, `NormalizeStatus`).

### Workflow Tests

**Files (verified):**
- `/home/devuser/td/internal/workflow/workflow_test.go` -- State machine transitions
- `/home/devuser/td/internal/workflow/guards_test.go` -- Guard implementations (DifferentReviewerGuard, HandoffRequiredGuard)
- `/home/devuser/td/internal/workflow/integration_test.go` -- Multi-step workflow scenarios

### TUI Component Tests

**Files (verified):**
- `/home/devuser/td/pkg/monitor/keymap/registry_test.go` -- Key binding lookup, sequences, context dispatch
- `/home/devuser/td/pkg/monitor/keymap/config_test.go` -- User override loading/saving
- `/home/devuser/td/pkg/monitor/modal/modal_test.go` -- Declarative modal system
- `/home/devuser/td/pkg/monitor/mouse/mouse_test.go` -- Hit-region mouse handling
- `/home/devuser/td/pkg/monitor/overlay_test.go` -- ANSI-aware overlay compositing
- `/home/devuser/td/pkg/monitor/clipboard_test.go` -- Clipboard operations
- `/home/devuser/td/pkg/monitor/markdown_test.go` -- Async markdown rendering
- `/home/devuser/td/pkg/monitor/help_modal_test.go` -- Help modal content generation

### Infrastructure Tests

**Files (verified):**
- `/home/devuser/td/internal/config/config_test.go`
- `/home/devuser/td/internal/git/git_test.go`
- `/home/devuser/td/internal/workdir/workdir_test.go`
- `/home/devuser/td/internal/features/features_test.go`
- `/home/devuser/td/internal/output/output_test.go`
- `/home/devuser/td/internal/output/tree_test.go`
- `/home/devuser/td/internal/suggest/suggest_test.go`
- `/home/devuser/td/internal/version/version_test.go`
- `/home/devuser/td/internal/version/semver_test.go`
- `/home/devuser/td/internal/version/cache_test.go`
- `/home/devuser/td/internal/version/checker_test.go`
- `/home/devuser/td/internal/crypto/crypto_test.go`
- `/home/devuser/td/internal/syncconfig/syncconfig_test.go`
- `/home/devuser/td/internal/input/input_test.go`
- `/home/devuser/td/internal/dependency/dependency_test.go`

---

## 8.4 Tier 2: Integration Tests

Integration tests use real SQLite databases and test multi-package interactions.

### Database Layer Tests

**Files (verified):**
- `/home/devuser/td/internal/db/db_test.go` -- Core DB operations, migrations
- `/home/devuser/td/internal/db/issues_logged_test.go` -- Logged issue operations with action_log verification
- `/home/devuser/td/internal/db/boards_logged_test.go` -- Board operations
- `/home/devuser/td/internal/db/relations_logged_test.go` -- Dependency/file link operations
- `/home/devuser/td/internal/db/lock_test.go` -- File lock concurrency
- `/home/devuser/td/internal/db/sessions_test.go` -- Session management
- `/home/devuser/td/internal/db/paths_test.go` -- File path resolution
- `/home/devuser/td/internal/db/activity_test.go` -- Activity feed queries
- `/home/devuser/td/internal/db/sync_history_test.go` -- Sync history tracking
- `/home/devuser/td/internal/db/agent_errors_test.go` -- Agent error logging
- `/home/devuser/td/internal/db/issue_relations_test.go` -- Issue relationships
- `/home/devuser/td/internal/db/bypass_prevention_test.go` -- Review bypass detection
- `/home/devuser/td/internal/db/multiuser_integration_test.go` -- Multi-user concurrent access
- `/home/devuser/td/internal/db/migrations_actionlog_test.go` -- Action log migration verification

**Coverage:** 14 test files for 26 source files (0.54 ratio). Strong coverage including dedicated tests for bypass prevention, multi-user scenarios, and migration edge cases.

### Session Integration Tests

**Files (verified):**
- `/home/devuser/td/internal/session/session_test.go`
- `/home/devuser/td/internal/session/agent_fingerprint_test.go`
- `/home/devuser/td/internal/session/agent_session_integration_test.go`
- `/home/devuser/td/internal/session/migration_integration_test.go`

### CLI Command Tests

**Files (verified -- 33 test files):**
- `/home/devuser/td/cmd/create_test.go` -- Issue creation
- `/home/devuser/td/cmd/list_test.go` -- Issue listing
- `/home/devuser/td/cmd/show_test.go` -- Issue detail display
- `/home/devuser/td/cmd/update_test.go` -- Issue modification
- `/home/devuser/td/cmd/delete_test.go` -- Issue deletion
- `/home/devuser/td/cmd/start_test.go` -- Start work
- `/home/devuser/td/cmd/review_test.go` -- Submit for review
- `/home/devuser/td/cmd/approve_test.go` -- Approve review
- `/home/devuser/td/cmd/handoff_test.go` -- Handoff capture
- `/home/devuser/td/cmd/block_test.go` -- Block/unblock
- `/home/devuser/td/cmd/search_test.go` -- TDQ search
- `/home/devuser/td/cmd/undo_test.go` -- Undo operations
- `/home/devuser/td/cmd/dependencies_test.go` -- Dependency management
- `/home/devuser/td/cmd/cascade_cli_test.go` -- Cascade operations
- `/home/devuser/td/cmd/epic_test.go` -- Epic workflows
- `/home/devuser/td/cmd/feature_test.go` -- Feature shortcuts
- `/home/devuser/td/cmd/focus_test.go` -- Focus management
- `/home/devuser/td/cmd/init_test.go` -- Initialization
- `/home/devuser/td/cmd/log_test.go` -- Log entries
- `/home/devuser/td/cmd/root_test.go` -- Root command
- `/home/devuser/td/cmd/security_test.go` -- Security commands
- `/home/devuser/td/cmd/status_test.go` -- Status display
- `/home/devuser/td/cmd/tree_test.go` -- Tree rendering
- `/home/devuser/td/cmd/unstart_test.go` -- Unstart work
- `/home/devuser/td/cmd/validation_test.go` -- Input validation
- `/home/devuser/td/cmd/ws_test.go` -- Work sessions
- `/home/devuser/td/cmd/context_test.go` -- Context management
- `/home/devuser/td/cmd/errors_test.go` -- Error handling
- `/home/devuser/td/cmd/minor_integration_test.go` -- Minor flag integration
- `/home/devuser/td/cmd/autosync_test.go` -- Auto-sync hooks
- `/home/devuser/td/cmd/autosync_push_test.go` -- Push sync
- `/home/devuser/td/cmd/sync_bootstrap_test.go` -- Snapshot bootstrap
- `/home/devuser/td/cmd/sync_tail_test.go` -- Sync tail display

### Server API Tests

**Files (verified):**
- `/home/devuser/td/internal/api/server_test.go`
- `/home/devuser/td/internal/api/auth_test.go`
- `/home/devuser/td/internal/api/cors_test.go`
- `/home/devuser/td/internal/api/ratelimit_test.go`
- `/home/devuser/td/internal/api/admin_server_test.go`
- `/home/devuser/td/internal/api/admin_middleware_test.go`
- `/home/devuser/td/internal/api/admin_events_test.go`
- `/home/devuser/td/internal/api/admin_projects_test.go`
- `/home/devuser/td/internal/api/admin_scopes_test.go`
- `/home/devuser/td/internal/api/admin_users_test.go`
- `/home/devuser/td/internal/api/admin_snapshots_test.go`
- `/home/devuser/td/internal/api/admin_integration_test.go`
- `/home/devuser/td/internal/api/testharness_test.go` -- Shared test setup

### Server DB Tests

**Files (verified):**
- `/home/devuser/td/internal/serverdb/serverdb_test.go`
- `/home/devuser/td/internal/serverdb/auth_test.go`
- `/home/devuser/td/internal/serverdb/device_auth_test.go`
- `/home/devuser/td/internal/serverdb/auth_events_test.go`
- `/home/devuser/td/internal/serverdb/admin_test.go`
- `/home/devuser/td/internal/serverdb/pagination_test.go`
- `/home/devuser/td/internal/serverdb/rate_limit_events_test.go`
- `/home/devuser/td/internal/serverdb/projects_event_count_test.go`

### TUI Model Tests

**Files (verified):**
- `/home/devuser/td/pkg/monitor/model_test.go` -- 4,289 lines, the largest test file. Comprehensive integration testing of TUI model lifecycle, message handling, and state transitions.
- `/home/devuser/td/pkg/monitor/data_test.go` -- Data fetching
- `/home/devuser/td/pkg/monitor/form_test.go` -- Form creation/validation
- `/home/devuser/td/pkg/monitor/form_operations_test.go` -- Form submit operations
- `/home/devuser/td/pkg/monitor/input_test.go` -- Input handling
- `/home/devuser/td/pkg/monitor/dbpool_test.go` -- Shared DB pool
- `/home/devuser/td/pkg/monitor/modal_comments_test.go` -- Comment modal
- `/home/devuser/td/pkg/monitor/submit_to_review_test.go` -- Review submission workflow
- `/home/devuser/td/pkg/monitor/sync_prompt_test.go` -- Sync prompt modal

---

## 8.5 Tier 3: Sync Harness Tests

**Location:** `/home/devuser/td/test/syncharness/` (verified)

The sync harness provides in-process multi-device sync testing without HTTP:

### Harness Design

From `test/syncharness/harness.go` (verified):

- Creates multiple in-memory SQLite databases simulating different devices
- Uses real sync engine code (`internal/sync/`) for event insertion and application
- Sets up the base schema plus all sync-extension tables (action_log, boards, board_issue_positions, sync_state, sync_conflicts, notes, agent_activity)
- Simulates push/pull cycles between devices via direct function calls

```go
// From harness.go (verified structure)
type Harness struct {
    // Multiple simulated devices with in-memory SQLite
    // Shared server event log
    // Real sync engine for event insertion and retrieval
}
```

### Sync Harness Test Files

| File | Purpose |
|------|---------|
| `harness_test.go` | Core harness smoke tests |
| `composite_sync_test.go` | Sync of composite-key entities (dependencies, files, positions) |
| `field_merge_test.go` | Field-level merge and overwrite testing |
| `session_field_sync_test.go` | Session field synchronization |
| `undo_sync_test.go` | Undo operations across synced devices |
| `board_soft_delete_test.go` | Board position soft delete sync |
| `activity_sync_test.go` | Agent activity sync |
| `server_migration_test.go` | Server schema migration compatibility |

---

## 8.6 Tier 4: End-to-End Tests

**Location:** `/home/devuser/td/test/e2e/` (verified)

The E2E harness builds real `td` and `td-sync` binaries, starts an actual HTTP server, and runs multi-actor scenarios.

### E2E Harness Design

From `test/e2e/harness.go` (verified):

```go
type Config struct {
    NumActors int     // 2 or 3 (alice, bob, optionally carol)
    AutoSync  bool    // enable auto-sync on clients
    Debounce  string  // e.g., "2s"
    Interval  string  // e.g., "10s"
}

type Harness struct {
    ServerURL  string
    ProjectID  string
    WorkDir    string
    TdBin      string    // Path to compiled td binary
    SyncBin    string    // Path to compiled td-sync binary
    clientDirs map[string]string  // actor -> working dir
    homeDirs   map[string]string  // actor -> HOME dir
    sessionIDs map[string]string  // actor -> session ID
    serverCmd  *exec.Cmd          // Running server process
}
```

**Setup flow:**
1. Builds `td` and `td-sync` binaries
2. Starts a real `td-sync` server on a random port
3. Authenticates configured actors (alice, bob, optionally carol)
4. Creates a project and links all actors
5. Uses `t.Cleanup` for teardown

### E2E Test Files

| File | Purpose |
|------|---------|
| `harness_test.go` | Harness setup/teardown verification |
| `engine_test.go` | Core sync push/pull scenarios |
| `conflicts_test.go` | Conflict resolution behavior |
| `history_test.go` | Sync history tracking |
| `ratelimit_test.go` | Rate limiting under load |
| `chaos_test.go` | Chaos testing (random failures, timing) |
| `restart_test.go` | Server restart recovery |
| `verify_test.go` | Data verification helpers |

---

## 8.7 Sync Engine Tests

**Files (verified):**
- `/home/devuser/td/internal/sync/engine_test.go` -- Server event log insertion and retrieval
- `/home/devuser/td/internal/sync/client_test.go` -- Client-side push/pull logic
- `/home/devuser/td/internal/sync/events_test.go` -- Event application (upsert, delete, restore)
- `/home/devuser/td/internal/sync/seed_parity_test.go` -- Ensures sync seeded data matches direct insertion
- `/home/devuser/td/internal/sync/backfill_test.go` -- Orphan entity backfill for initial sync

**Coverage:** 5 test files for 5 source files (1:1 ratio).

---

## 8.8 Test Patterns and Conventions

### Common Patterns

**1. Test helpers with `t.Helper()`:**
Tests consistently use Go's `t.Helper()` for helper functions, providing accurate line numbers in failure messages.

**2. Temp directory isolation:**
Integration and E2E tests use `t.TempDir()` for isolated database directories, preventing test interference.

**3. Table-driven tests:**
Parser, evaluator, and model tests use table-driven test patterns:

```go
tests := []struct {
    name     string
    input    string
    expected SomeType
}{
    {"description", "input", expected},
    // ...
}
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        // test body
    })
}
```

**4. Real database in tests:**
Most tests use real SQLite databases rather than mocks. The `db.Open` and `db.Init` functions are called in test setup to ensure schema migrations run correctly.

**5. Test-specific base directory override:**
`cmd/` tests use `baseDirOverride` pointer to redirect database operations to temp directories without modifying the global `baseDir`.

**6. Named actors in E2E:**
E2E tests use named actors (alice, bob, carol) for multi-device scenarios, making test output readable.

### Error Handling in Tests

Tests generally check `err != nil` with descriptive failure messages. The `cmd/` tests capture stderr/stdout for assertion.

---

## 8.9 Coverage Analysis

### Per-Package Test Ratios

| Package | Source Files | Test Files | Ratio | Assessment |
|---------|-------------|------------|-------|------------|
| `internal/sync/` | 5 | 5 | **1.00** | Excellent |
| `internal/query/` | 7 | 5 | **0.71** | Excellent |
| `cmd/` | 50 | 33 | **0.66** | Good |
| `internal/db/` | 26 | 14 | **0.54** | Good |
| `internal/api/` | ~15 | 13 | **0.87** | Excellent |
| `internal/serverdb/` | ~8 | 8 | **1.00** | Excellent |
| `internal/session/` | ~3 | 4 | **1.33** | Excellent |
| `internal/workflow/` | ~4 | 3 | **0.75** | Good |
| `pkg/monitor/` | 35 | 17 | **0.49** | Moderate |
| `test/e2e/` | ~5 (harness) | 8 | N/A | Dedicated E2E |
| `test/syncharness/` | ~3 (harness) | 8 | N/A | Dedicated harness |

### Overall Metrics

| Metric | Value |
|--------|-------|
| Total test files | 134 |
| Total source files | 198 |
| Test file ratio | 0.68 |
| Total test lines | ~67,123 |
| Total source lines | ~58,753 |
| Test-to-source line ratio | **1.14** |
| Largest test file | `pkg/monitor/model_test.go` (4,289 lines) |

---

## 8.10 Coverage Gaps

### Critical Gaps

| Area | Lines | Issue |
|------|-------|-------|
| `pkg/monitor/view.go` | 2,399 | **Largest file, no dedicated test file.** Scroll indicator arithmetic, modal dimension calculations, and overlay composition untested directly. Covered indirectly via `model_test.go`. |
| `pkg/monitor/commands.go` | 2,166 | **Second largest file, no dedicated test file.** The 60+ switch cases in `executeCommand()` have no targeted unit tests. |

### Moderate Gaps

| Area | Issue |
|------|-------|
| `cmd/monitor.go` | No test file for the monitor launch command |
| `cmd/board.go` | No test file for board management commands |
| `cmd/activity.go` | No test file for activity command |
| `cmd/assign.go` | No test file for assign command |
| `internal/agent/instructions.go` | Agent file detection/installation untested |
| `internal/syncclient/client.go` | HTTP sync client untested (covered by E2E) |
| `pkg/monitor/kanban.go` | No test file for kanban view |
| `pkg/monitor/board_editor.go` | No test file for board editor |
| `pkg/monitor/getting_started.go` | No test file for onboarding modal |
| `pkg/monitor/sync_prompt.go` | Has test file, but limited coverage |
| `pkg/monitor/actions.go` | No test file for issue actions |
| `pkg/monitor/activity_table.go` | No test file for activity table rendering |

### Monitor Package Detail

The `pkg/monitor/` package has the weakest test ratio (0.49). While `model_test.go` at 4,289 lines provides extensive integration testing, the following files lack dedicated unit tests:

- `view.go` (2,399 lines) -- all rendering logic
- `commands.go` (2,166 lines) -- all command dispatch
- `input.go` (1,535 lines) -- mouse/input handling
- `kanban.go` (~406 lines) -- kanban view
- `board_editor.go` -- board CRUD modals
- `getting_started.go` -- onboarding
- `actions.go` -- issue mutation actions
- `activity_table.go` -- activity feed rendering
- `styles.go` -- style definitions

---

## 8.11 Test Infrastructure Quality

### Sync Test Harness (`test/syncharness/`)

**Strengths:**
- Creates a complete sync-ready schema by combining the base schema with sync extension tables
- Simulates multi-device sync without HTTP overhead
- Tests include composite-key sync, field merging, undo sync, soft delete sync
- Server migration compatibility testing

**Verified from source:** The harness (`test/syncharness/harness.go:20-100`) creates the full schema including `action_log`, `boards`, `board_issue_positions`, `issue_session_history`, `sync_state`, `sync_conflicts`, and `notes` tables.

### E2E Test Harness (`test/e2e/`)

**Strengths:**
- Builds real binaries (`go build`) for authentic testing
- Starts a real HTTP server on a random port
- Named actors (alice, bob, carol) for clear multi-device scenarios
- Configurable auto-sync, debounce, and interval settings
- `t.Cleanup` for reliable teardown
- Chaos and restart tests for resilience

**Verified from source:** The harness (`test/e2e/harness.go:83-100`) shows real binary compilation, server startup, actor authentication, and project creation.

### Database Test Pattern

Integration tests consistently:
1. Create temp directory via `t.TempDir()`
2. Initialize DB via `db.Init(tempDir)`
3. Run full migration chain
4. Execute test operations on real SQLite
5. Verify with SQL queries or DB methods

This ensures migrations are continuously tested as part of every integration test.

---

## 8.12 Recommendations

### High Priority

| # | Recommendation | Rationale | Effort |
|---|---------------|-----------|--------|
| 1 | Add unit tests for `view.go` scroll arithmetic | The effectiveMaxLines calculation is duplicated 4 times and is regression-prone during refactoring | Medium |
| 2 | Add unit tests for `commands.go` dispatch branches | 60+ switch cases untested individually; targeted tests would catch regressions when refactoring the monolithic switch | Medium |
| 3 | Add modal dimension calculation tests | Magic numbers (80%, cap 100/40, floor 40/50) are repeated in 5+ places with slight variations | Low |

### Medium Priority

| # | Recommendation | Rationale | Effort |
|---|---------------|-----------|--------|
| 4 | Add `cmd/board.go` and `cmd/monitor.go` tests | Board management and monitor launch are user-facing commands without test coverage | Medium |
| 5 | Add `internal/agent/instructions.go` tests | Agent file detection affects first-run experience | Low |
| 6 | Add kanban view tests | Kanban column navigation and rendering has complex state management | Medium |
| 7 | Run full migration chain test (v1 to current) on both empty and populated databases | Some migrations (v13, v14) repair previous migration failures; ensuring the full chain works prevents silent regressions | Medium |

### Low Priority

| # | Recommendation | Rationale | Effort |
|---|---------------|-----------|--------|
| 8 | Add syncclient HTTP tests | Currently only tested via E2E; unit-level HTTP client tests would catch API contract violations faster | Medium |
| 9 | Extract and test scroll indicator helper | Duplicated logic would be easier to test once extracted into a helper | Low |
| 10 | Add property-based testing for TDQ parser | Parser edge cases (empty input, max depth, unicode, escape sequences) benefit from fuzz/property testing | Medium |

### Testing Conventions to Maintain

The following existing patterns should be preserved and extended:

1. **Real SQLite in tests** -- not mocks. This continuously validates migrations and SQL correctness.
2. **Table-driven tests** -- for parsers, validators, and normalizers. Easy to extend.
3. **Named actors in E2E** -- alice/bob/carol make multi-device test output readable.
4. **`t.Helper()` usage** -- maintains accurate line numbers in failure output.
5. **Temp directory isolation** -- prevents test cross-contamination.
6. **1:1 test ratio for sync engine** -- the sync subsystem's excellent coverage should be maintained as it evolves.
