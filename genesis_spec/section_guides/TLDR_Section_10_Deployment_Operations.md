# Section 10: Deployment & Operations

## Overview

td consists of two binaries: the main `td` CLI client and the optional `td-sync` server. The client is a self-contained Go binary with an embedded SQLite database. The server is a standalone HTTP service for multi-device synchronization. Both follow Go's standard build and distribution model.

---

## 10.1 Build Process

### Client Binary (`td`)

**Standard build**:
```bash
go build -o td .
```

**With version injection**:
```bash
go install -ldflags "-X main.Version=v0.3.0" ./...
```

The `main.go` entry point (`/home/devuser/td/main.go`) uses `effectiveVersion()` to determine the runtime version through a priority chain:

1. **Build-time injection**: If `-ldflags "-X main.Version=vX.Y.Z"` was used and the value is not "dev", use it directly
2. **Go module info**: If installed via `go install module@vX.Y.Z`, `debug.ReadBuildInfo().Main.Version` provides the version
3. **VCS fallback**: Extracts `vcs.revision` from build settings, truncates to 12 chars, appends "+dirty" if `vcs.modified` is true. Produces strings like `devel+a1b2c3d4e5f6`
4. **Default**: Returns "dev" if nothing else is available

### SQLite Driver Selection

The codebase includes **two** SQLite drivers in `go.mod`:
- `modernc.org/sqlite v1.41.0` - Pure Go SQLite (no CGo required)
- `github.com/mattn/go-sqlite3 v1.14.33` - CGo-based SQLite (faster)

**In production code** (`internal/db/db.go:12`), only the pure-Go driver is imported:
```go
_ "modernc.org/sqlite"
```

**In test code**, the CGo driver (`mattn/go-sqlite3`) is used by the sync test harness and E2E tests:
- `test/syncharness/harness.go`
- `test/e2e/conflicts.go`
- `internal/sync/*_test.go` files

This means the production binary is **fully pure-Go** and cross-compiles without CGo. The test suite uses the CGo driver for performance (sync tests involve significant database operations).

### Sync Server Binary (`td-sync`)

```bash
go build -o td-sync ./cmd/td-sync
```

The server has its own `main()` at `/home/devuser/td/cmd/td-sync/main.go`.

### Go Version

The project targets **Go 1.25.5** (from `go.mod`).

### Testing

```bash
go test ./...              # All tests
go test ./internal/db/...  # Just DB layer
go test ./pkg/monitor/...  # Just TUI
```

The test suite includes 134 test files with ~67K lines of test code. E2E tests (`test/e2e/`) use the actual `td` binary in temporary directories.

---

## 10.2 Installation

### Via `go install`

```bash
go install github.com/marcus/td@latest
```

Or with a specific version:
```bash
go install -ldflags "-X main.Version=v0.3.0" github.com/marcus/td@v0.3.0
```

### Via `git clone` + Build

```bash
git clone https://github.com/marcus/td.git
cd td
go build -o td .
# Optionally move to PATH
mv td /usr/local/bin/
```

### Project Initialization

After installation, initialize a project:
```bash
cd /path/to/project
td init
```

This creates the `.todos/` directory with the SQLite database, adds `.todos/` to `.gitignore` (if in a git repo), and optionally installs td usage instructions into detected agent files (CLAUDE.md, AGENTS.md, .cursorrules).

**Source**: `/home/devuser/td/cmd/init.go`

The `td init` command:
1. Checks if `.todos/` already exists (warns and exits if so)
2. Calls `db.Initialize(baseDir)` which creates the directory and runs all 29 migrations
3. Adds `.todos/` to `.gitignore`
4. Creates a session for the current agent/terminal
5. Detects agent files and offers to inject td instructions

---

## 10.3 Sync Server Deployment

### Binary

The `td-sync` server is a separate binary built from `/home/devuser/td/cmd/td-sync/main.go`.

### Configuration (Environment Variables)

All configuration is via environment variables with sensible defaults:

| Variable | Default | Description |
|----------|---------|-------------|
| `SYNC_LISTEN_ADDR` | `:8080` | HTTP listen address |
| `SYNC_SERVER_DB_PATH` | `./data/server.db` | Server metadata SQLite path |
| `SYNC_PROJECT_DATA_DIR` | `./data/projects` | Per-project event store directory |
| `SYNC_SHUTDOWN_TIMEOUT` | `30s` | Graceful shutdown timeout |
| `SYNC_ALLOW_SIGNUP` | `true` | Allow new user registration |
| `SYNC_BASE_URL` | `http://localhost:8080` | Public base URL (for auth verification pages) |
| `SYNC_LOG_FORMAT` | `json` | Log format: `json` or `text` |
| `SYNC_LOG_LEVEL` | `info` | Log level: `debug`, `info`, `warn`, `error` |
| `SYNC_RATE_LIMIT_AUTH` | `10` | Auth requests per IP per minute |
| `SYNC_RATE_LIMIT_PUSH` | `60` | Push requests per API key per minute |
| `SYNC_RATE_LIMIT_PULL` | `120` | Pull requests per API key per minute |
| `SYNC_RATE_LIMIT_OTHER` | `300` | Other requests per API key per minute |
| `SYNC_CORS_ALLOWED_ORIGINS` | (empty) | Comma-separated CORS origins for admin API |
| `SYNC_AUTH_EVENT_RETENTION` | `90d` | Auth event log retention |
| `SYNC_RATE_LIMIT_EVENT_RETENTION` | `30d` | Rate limit event log retention |

**Source**: `/home/devuser/td/internal/api/config.go`

### Server Architecture

The server uses Go's standard `net/http` with:
- **Timeouts**: Read 15s, Write 60s, Idle 120s
- **Max request body**: 10 MB
- **Graceful shutdown**: Waits for `ShutdownTimeout` before forcing close
- **Middleware chain**: recovery, requestID, logger, metrics, logging, maxBytes, authRateLimit
- **CORS**: Enabled only for admin endpoints, controlled by `SYNC_CORS_ALLOWED_ORIGINS`

### Admin CLI

The `td-sync` binary includes admin subcommands:

```bash
td-sync admin grant --email user@example.com       # Grant admin
td-sync admin revoke --email user@example.com      # Revoke admin (prevents removing last admin)
td-sync admin create-key --email admin@example.com --name td-watch --scopes "admin:read:server,sync"
```

**Source**: `/home/devuser/td/cmd/td-sync/admin.go`

### Signal Handling

The server listens for `SIGINT` and `SIGTERM` via `signal.NotifyContext()` and performs graceful shutdown, closing all project database connections via `dbPool.CloseAll()`.

---

## 10.4 Configuration Management

### Client Configuration Layers

td uses three configuration layers (in priority order):

#### 1. Environment Variables
- `TD_ANALYTICS` - Enable/disable analytics (`false` to disable)
- `TD_LOG_FILE` - Redirect slog to file
- `TD_DISABLE_EXPERIMENTAL` - Kill switch for all experimental features
- `TD_FEATURE_<NAME>` - Per-feature flag override (e.g., `TD_FEATURE_SYNC_CLI=true`)
- `TD_DISABLE_FEATURE` / `TD_ENABLE_FEATURE` - Comma-separated feature lists

#### 2. Project Config (`config.json`)

Stored at `.todos/config.json`. Managed via `internal/config/config.go`. Contains:

- **Focus**: Current focused issue ID
- **Work session**: Active work session ID
- **Pane heights**: Three-panel height ratios (persisted on drag-end)
- **Feature flags**: `map[string]bool` for feature gates
- **Filter state**: Search query, sort mode, type filter, include_closed
- **Title validation**: Min/max title length (defaults: 15-100 chars)

Config access is serialized via `flock` to prevent corruption from concurrent agents.

#### 3. Sync Config

Stored at `~/.config/td/config.json` (system-wide, not per-project). Managed via `internal/syncconfig/`.

**Configurable keys** (via `td config set/get`):

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `sync.url` | string | (none) | Sync server URL |
| `sync.enabled` | bool | false | Enable sync |
| `sync.auto.enabled` | bool | true | Auto-sync on mutations |
| `sync.auto.debounce` | duration | 3s | Debounce interval |
| `sync.auto.interval` | duration | 5m | Periodic sync interval |
| `sync.auto.pull` | bool | true | Auto-pull on sync |
| `sync.auto.on_start` | bool | true | Sync on startup |
| `sync.snapshot_threshold` | int | 100 | Events before snapshot |

**Source**: `/home/devuser/td/cmd/config.go`

### Feature Flags

**Source**: `/home/devuser/td/internal/features/features.go`

Four feature flags control sync-related functionality:

| Flag | Default | Description |
|------|---------|-------------|
| `sync_cli` | false | Gates sync/auth CLI commands |
| `sync_autosync` | false | Gates auto-sync hooks |
| `sync_monitor_prompt` | false | Gates first-run sync prompt in monitor |
| `sync_notes` | true | Gates notes entity sync |

Resolution order: Environment override > project config > default value.

Management:
```bash
td feature list          # Show all flags and their resolved state
td feature get sync_cli  # Check specific flag
td feature set sync_cli true   # Enable
td feature unset sync_cli      # Reset to default
```

---

## 10.5 Database Location and Management

### Location

The database is stored at `<project>/.todos/issues.db`. The `.todos/` directory is created by `td init`.

### WAL Mode

SQLite is opened with WAL (Write-Ahead Logging) mode for concurrent read access. A single connection is used (`SetMaxOpenConns(1)`) to prevent write conflicts.

### Write Locking

Cross-process write safety is handled by file-based locking (`internal/db/lock.go`):
- **Unix**: `flock()` via `lock_unix.go`
- **Windows**: `LockFileEx()` via `lock_windows.go`
- **Stale detection**: Lock file includes PID and timestamp for identifying dead locks
- **Exponential backoff**: Retries with increasing delays on lock contention

### Migrations

The schema starts at version 1 and progresses through 29 migrations (`/home/devuser/td/internal/db/schema.go`). Migrations run automatically on database open.

Migration types:
- **Pure SQL**: Most migrations (ALTER TABLE, CREATE TABLE, CREATE INDEX)
- **Custom Go code**: Complex migrations (v13/14 - session table repair, v15 - text ID migration, v18 - deterministic IDs, v19 - path normalization, v20 - action log normalization, v24 - work session IDs, v25 - board position soft delete, v26 - action log NOT NULL fix)

**Manual migration**: `td upgrade` runs migrations explicitly.

### Git Worktree Support

The `internal/workdir/` package supports git worktrees. A `.td-root` file in a worktree directory points to the main repo's `.todos` directory, so all worktrees share a single database.

---

## 10.6 Version Checking (Not Auto-Update)

The `docs/implemented/auto-update-plan.md` describes a plan for auto-updates, but the actual implementation is a **version check only**, not an auto-update mechanism.

### What is Actually Implemented

**Source**: `/home/devuser/td/internal/version/version.go` and `/home/devuser/td/internal/version/cache.go`

The `td version` command (with `--check` flag, which defaults to `true`):

1. Prints the current version
2. Checks a local cache file (avoids hitting GitHub on every invocation)
3. If cache is stale, fetches the latest release from `https://api.github.com/repos/marcus/td/releases/latest` (5-second timeout)
4. Caches the result
5. If an update is available, prints:
   ```
   Update available: v0.2.0 -> v0.3.0
   Run: go install -ldflags "-X main.Version=v0.3.0" github.com/marcus/td@v0.3.0
   ```

**Important**: There is **no automatic download or installation**. The user must manually run the `go install` command. The `UpdateCommand()` function validates the version string against a regex to prevent shell injection.

### TUI Version Check

The monitor also checks for updates asynchronously on launch and displays a notification in the footer if an update is available. This is passive -- no actionable upgrade path from within the TUI.

### What the Plan Proposed but is NOT Implemented

The auto-update plan (`docs/implemented/auto-update-plan.md`) proposed:
- An `internal/update/update.go` package (does not exist)
- An `update` subcommand (does not exist)
- Automatic `go install` execution (not implemented)
- Root-level async check hook in `PersistentPostRun` (not implemented)

The actual implementation lives in `internal/version/` (not `internal/update/`), and only performs checking with caching.

---

## 10.7 Version Management

**Source**: `/home/devuser/td/main.go:24-63`

### Version Resolution Chain

```
Build-time -ldflags  ->  go install @version  ->  VCS revision  ->  "dev"
```

The `effectiveVersion()` function:

1. If `Version` (set via `-ldflags`) is non-empty and not "dev", use it
2. If `debug.ReadBuildInfo().Main.Version` is set and not "(devel)", use it (happens with `go install module@version`)
3. If `vcs.revision` exists in build settings, construct `devel+<12-char-hash>[+dirty]`
4. Fall back to the raw `Version` variable (typically "dev")

### Release Process (from CLAUDE.md)

```bash
# 1. Commit with proper message
git commit -m "feat: description"

# 2. Create version tag
git tag -a v0.3.0 -m "Release v0.3.0: description"

# 3. Push commit and tag
git push origin main
git push origin v0.3.0

# 4. Install locally
go install -ldflags "-X main.Version=v0.3.0" ./...
```

There is no CI/CD pipeline, GitHub Actions workflow, or binary release automation visible in the codebase. Distribution relies on `go install` from the module path.

---

## 10.8 Operational Concerns

### Disk Usage

The SQLite database grows with usage. Key tables that accumulate data:
- `action_log` - Every mutation is logged (for undo + sync)
- `agent_activity` - Every agent action recorded
- `sync_history` - Every sync operation logged
- `sync_conflicts` - Every conflict preserved

There is no built-in pruning or compaction for client databases.

### Concurrency

Multiple agents can safely write concurrently thanks to:
- File-based write locks with exponential backoff
- WAL mode for concurrent reads
- Busy timeout on the SQLite connection

### Backup

No built-in backup command. The database is a single file (`.todos/issues.db`) that can be copied. WAL mode means `.todos/issues.db-wal` and `.todos/issues.db-shm` must also be included for a consistent backup.

### Environment Variables Summary

| Variable | Purpose | Scope |
|----------|---------|-------|
| `TD_ANALYTICS` | Enable/disable analytics | Client |
| `TD_LOG_FILE` | Redirect slog to file | Client |
| `TD_DISABLE_EXPERIMENTAL` | Kill switch for experimental features | Client |
| `TD_FEATURE_*` | Individual feature flag overrides | Client |
| `TD_DISABLE_FEATURE` | Comma-separated features to disable | Client |
| `TD_ENABLE_FEATURE` | Comma-separated features to enable | Client |
| `SYNC_LISTEN_ADDR` | Server listen address | Server |
| `SYNC_SERVER_DB_PATH` | Server DB path | Server |
| `SYNC_PROJECT_DATA_DIR` | Project data directory | Server |
| `SYNC_LOG_FORMAT` | Log format (json/text) | Server |
| `SYNC_LOG_LEVEL` | Log level | Server |
| `SYNC_ALLOW_SIGNUP` | Allow user registration | Server |
| `SYNC_BASE_URL` | Public URL for auth pages | Server |
| `SYNC_RATE_LIMIT_*` | Rate limiting thresholds | Server |
| `SYNC_CORS_ALLOWED_ORIGINS` | Admin API CORS origins | Server |
| `SYNC_AUTH_EVENT_RETENTION` | Auth event log retention | Server |
| `SYNC_RATE_LIMIT_EVENT_RETENTION` | Rate limit event retention | Server |
