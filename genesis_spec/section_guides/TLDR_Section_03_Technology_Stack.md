# Section 03: Technology Stack

## Language & Runtime

| Property | Value |
|---|---|
| Language | Go |
| Version | 1.25.5 |
| Module Path | `github.com/marcus/td` |
| Entry Point | `main.go` |
| Binaries | `td` (CLI), `td-sync` (sync server) |

**Source:** `go.mod` line 3: `go 1.25.5`

---

## Core Frameworks

### CLI Framework: spf13/cobra v1.10.2

**Purpose:** Command-line subcommand routing, flag parsing, help generation, command grouping.

**Usage:** All 50+ CLI commands are individual Cobra `*cobra.Command` definitions in `cmd/`. The root command (`cmd/root.go`) organizes commands into groups (core, workflow, query, session, files, system) with `PersistentPreRun` and `PersistentPostRun` hooks for cross-cutting concerns (sync, analytics).

**Dependency:** `spf13/pflag v1.0.10` provides POSIX-compliant flag parsing (used internally by Cobra).

**Source:** `go.mod` line 14: `github.com/spf13/cobra v1.10.2`

### TUI Framework: charmbracelet/bubbletea v1.3.10

**Purpose:** Elm-architecture TUI framework providing the Model/Update/View pattern for the full-screen terminal monitor.

**Usage:** The entire TUI monitor (`pkg/monitor/`) is built on BubbleTea. The `Model` struct (~200 fields) implements `tea.Model` with `Init()`, `Update(msg tea.Msg)`, and `View()` methods. All state mutations flow through typed messages (`TickMsg`, `RefreshDataMsg`, `IssueDetailsMsg`, etc.). Side effects are encapsulated in `tea.Cmd` functions.

**Source:** `go.mod` line 7: `github.com/charmbracelet/bubbletea v1.3.10`

### TUI Styling: charmbracelet/lipgloss v1.1.1-0.20250404203927

**Purpose:** Terminal styling -- borders, colors, layout composition with `JoinHorizontal`/`JoinVertical`, padding, margins.

**Usage:** `pkg/monitor/styles.go` defines all styling constants. Every view rendering function uses lipgloss for layout composition. Status badges, priority indicators, and panel borders are all lipgloss-styled.

**Note:** This is a pre-release commit-pinned version (the `-0.20250404203927-76690c660834` suffix), indicating the project tracks lipgloss development closely.

**Source:** `go.mod` line 10: `github.com/charmbracelet/lipgloss v1.1.1-0.20250404203927-76690c660834`

### TUI Components: charmbracelet/bubbles v0.21.1-0.20250623103423

**Purpose:** Pre-built TUI components: text inputs, viewports, spinners, etc.

**Usage:** `pkg/monitor/form.go` uses text inputs and textareas for issue creation/editing. Viewports used for scrollable content in modals and panels.

**Note:** Also commit-pinned pre-release version.

**Source:** `go.mod` line 6: `github.com/charmbracelet/bubbles v0.21.1-0.20250623103423-23b8fd6302d7`

### Database (Primary): mattn/go-sqlite3 v1.14.33

**Purpose:** CGo-based SQLite3 driver implementing Go's `database/sql` interface. This is the default/primary database driver.

**Usage:** All database operations in `internal/db/` use this driver via `database/sql`. WAL mode is enabled for concurrent reads. A single connection is maintained per process to prevent corruption. Busy timeout is configured for multi-process access.

**Requires:** C compiler (CGo) for compilation.

**Source:** `go.mod` line 13: `github.com/mattn/go-sqlite3 v1.14.33`

### Database (Alternative): modernc.org/sqlite v1.41.0

**Purpose:** Pure-Go SQLite driver (CGo-free alternative) for cross-compilation scenarios where a C compiler is unavailable.

**Usage:** Selectable via Go build tags. Same `database/sql` interface as the CGo driver, allowing transparent switching.

**Dependencies (transitive):**
- `modernc.org/libc v1.66.10` -- C library reimplementation in Go
- `modernc.org/mathutil v1.7.1` -- Math utilities
- `modernc.org/memory v1.11.0` -- Memory allocation

**Source:** `go.mod` line 18: `modernc.org/sqlite v1.41.0`

---

## Supporting Libraries

### Markdown Rendering: charmbracelet/glamour v0.10.0

**Purpose:** Renders Markdown to styled terminal output.

**Usage:** `pkg/monitor/markdown.go` uses glamour to render issue descriptions with syntax highlighting in the TUI monitor. Custom Chroma themes are supported via the theme builder system.

**Source:** `go.mod` line 8: `github.com/charmbracelet/glamour v0.10.0`

### Terminal Forms: charmbracelet/huh v0.8.0

**Purpose:** Interactive terminal forms and prompts.

**Usage:** `pkg/monitor/form.go` and `form_modal.go` use huh for issue creation and editing forms within the TUI monitor. Provides input fields, selects, and multi-selects.

**Source:** `go.mod` line 9: `github.com/charmbracelet/huh v0.8.0`

### ANSI Utilities: charmbracelet/x/ansi v0.11.3

**Purpose:** ANSI string manipulation and truncation.

**Usage:** `pkg/monitor/kanban.go` and `view.go` use this for text truncation that respects ANSI escape sequences (preventing broken styling when truncating colored text).

**Source:** `go.mod` line 11: `github.com/charmbracelet/x/ansi v0.11.3`

### Cell Buffer: charmbracelet/x/cellbuf v0.0.14

**Purpose:** Cell buffer for terminal rendering (internal lipgloss dependency).

**Source:** `go.mod` line 12: `github.com/charmbracelet/x/cellbuf v0.0.14`

### Terminal Handling: golang.org/x/term v0.39.0

**Purpose:** Terminal size detection, raw mode handling.

**Usage:** `pkg/monitor/` uses this for detecting terminal dimensions and configuring the TUI accordingly.

**Source:** `go.mod` line 17: `golang.org/x/term v0.39.0`

### System Primitives: golang.org/x/sys v0.40.0

**Purpose:** Low-level OS primitives for terminal handling and file locking.

**Usage:** System calls for terminal handling, cross-platform file locking (`flock` on Unix, `LockFileEx` on Windows).

**Source:** `go.mod` line 16: `golang.org/x/sys v0.40.0`

---

## Full Dependency Tree

### Direct Dependencies (18)

| Package | Version | Category |
|---|---|---|
| `charmbracelet/bubbles` | v0.21.1-0.20250623 | TUI components |
| `charmbracelet/bubbletea` | v1.3.10 | TUI framework |
| `charmbracelet/glamour` | v0.10.0 | Markdown rendering |
| `charmbracelet/huh` | v0.8.0 | Terminal forms |
| `charmbracelet/lipgloss` | v1.1.1-0.20250404 | Terminal styling |
| `charmbracelet/x/ansi` | v0.11.3 | ANSI string handling |
| `charmbracelet/x/cellbuf` | v0.0.14 | Cell buffer |
| `mattn/go-sqlite3` | v1.14.33 | SQLite (CGo) |
| `spf13/cobra` | v1.10.2 | CLI framework |
| `spf13/pflag` | v1.0.10 | Flag parsing |
| `golang.org/x/sys` | v0.40.0 | System primitives |
| `golang.org/x/term` | v0.39.0 | Terminal handling |
| `modernc.org/sqlite` | v1.41.0 | SQLite (pure-Go) |

**Note:** 5 of the 18 direct dependencies are from the Charm ecosystem (charmbracelet/*), reflecting the deep investment in TUI quality.

### Indirect Dependencies (33)

| Package | Version | Pulled By |
|---|---|---|
| `alecthomas/chroma/v2` | v2.14.0 | glamour (syntax highlighting) |
| `atotto/clipboard` | v0.1.4 | bubbles (clipboard support) |
| `aymanbagabas/go-osc52/v2` | v2.0.1 | termenv (clipboard via OSC52) |
| `aymerick/douceur` | v0.2.0 | bluemonday (CSS parsing) |
| `catppuccin/go` | v0.3.0 | glamour (theme) |
| `charmbracelet/colorprofile` | v0.3.3 | lipgloss (color detection) |
| `charmbracelet/x/exp/slice` | v0.0.0-20250327 | huh |
| `charmbracelet/x/exp/strings` | v0.0.0-20240722 | huh |
| `charmbracelet/x/term` | v0.2.2 | bubbletea |
| `clipperhouse/displaywidth` | v0.6.1 | text width calculation |
| `clipperhouse/stringish` | v0.1.1 | string utilities |
| `clipperhouse/uax29/v2` | v2.3.0 | Unicode text segmentation |
| `dlclark/regexp2` | v1.11.0 | chroma (regex engine) |
| `dustin/go-humanize` | v1.0.1 | modernc.org/sqlite |
| `erikgeiser/coninput` | v0.0.0-20211004 | bubbletea (Windows console) |
| `google/uuid` | v1.6.0 | modernc.org/sqlite |
| `gorilla/css` | v1.0.1 | bluemonday |
| `inconshreveable/mousetrap` | v1.1.0 | cobra (Windows) |
| `lucasb-eyer/go-colorful` | v1.3.0 | lipgloss (color space) |
| `mattn/go-isatty` | v0.0.20 | termenv (TTY detection) |
| `mattn/go-localereader` | v0.0.1 | bubbletea |
| `mattn/go-runewidth` | v0.0.19 | lipgloss (CJK width) |
| `microcosm-cc/bluemonday` | v1.0.27 | glamour (HTML sanitization) |
| `mitchellh/hashstructure/v2` | v2.0.2 | huh |
| `muesli/ansi` | v0.0.0-20230316 | termenv |
| `muesli/cancelreader` | v0.2.2 | bubbletea |
| `muesli/reflow` | v0.3.0 | glamour (text reflow) |
| `muesli/termenv` | v0.16.0 | lipgloss (terminal env) |
| `ncruces/go-strftime` | v0.1.9 | modernc.org/sqlite |
| `remyoudompheng/bigfft` | v0.0.0-20230129 | modernc.org/sqlite |
| `rivo/uniseg` | v0.4.7 | runewidth (grapheme clusters) |
| `xo/terminfo` | v0.0.0-20220910 | termenv |
| `yuin/goldmark` | v1.7.8 | glamour (Markdown parser) |
| `yuin/goldmark-emoji` | v1.0.5 | glamour (emoji support) |
| `golang.org/x/crypto` | v0.47.0 | modernc.org/sqlite |
| `golang.org/x/exp` | v0.0.0-20250620 | modernc.org/sqlite |
| `golang.org/x/net` | v0.48.0 | bluemonday |
| `golang.org/x/text` | v0.33.0 | various |
| `modernc.org/libc` | v1.66.10 | modernc.org/sqlite |
| `modernc.org/mathutil` | v1.7.1 | modernc.org/sqlite |
| `modernc.org/memory` | v1.11.0 | modernc.org/sqlite |

---

## Build System

### Build Commands

```bash
# Build the main CLI binary
go build -o td .

# Build the sync server binary
go build -o td-sync ./cmd/td-sync/

# Run all tests
go test ./...

# Install with version injection
go install -ldflags "-X main.Version=v0.33.0" ./...
```

### Version Injection

Version is determined by a three-tier fallback system (verified in `main.go:24-63`):

1. **Build-time injection:** `-ldflags "-X main.Version=vX.Y.Z"` sets the version at compile time
2. **Go install info:** `debug.ReadBuildInfo()` reads `info.Main.Version` when installed via `go install module@vX.Y.Z`
3. **VCS revision fallback:** Extracts `vcs.revision` and `vcs.modified` from build settings to produce `devel+<short-sha>[+dirty]`

```go
// main.go
var Version = "dev"

func effectiveVersion(v string) string {
    if v != "" && v != "dev" { return v }
    info, ok := debug.ReadBuildInfo()
    // ... fallback logic
}
```

### Dual SQLite Driver Strategy

td supports two SQLite drivers, selectable via Go build tags:

| Driver | Package | Requires | Use Case |
|---|---|---|---|
| **Primary (default)** | `mattn/go-sqlite3` v1.14.33 | CGo (C compiler) | Normal builds, maximum compatibility with SQLite |
| **Alternative** | `modernc.org/sqlite` v1.41.0 | Nothing (pure Go) | Cross-compilation, environments without C compiler |

Both implement Go's `database/sql` interface, so the rest of the codebase is driver-agnostic. The selection is done at the `internal/db/` level via build tag conditional imports.

### Database Configuration

- **Mode:** WAL (Write-Ahead Logging) for concurrent reads
- **Connections:** Single connection per process (prevents SQLite corruption)
- **Busy timeout:** Configured for multi-process access
- **Write safety:** OS-level file locking (`flock` on Unix, `LockFileEx` on Windows) with exponential backoff and stale lock detection
- **Location:** `<project>/.todos/issues.db`
- **Schema:** 29 sequential migrations, 20 tables

---

## Internal Systems (Not External Dependencies)

These are substantial subsystems built within the td codebase itself:

### TDQ (td Query Language) -- `internal/query/`

A full custom query language with:
- **Lexer** (`lexer.go`): Tokenizes TDQ syntax into tokens
- **Parser** (`parser.go`): Recursive descent parser producing AST; handles implicit AND, operator precedence, nested expressions up to depth 50
- **AST** (`ast.go`): Node types -- `BinaryExpr`, `UnaryExpr`, `FieldExpr`, `FunctionCall`, `TextSearch`, `SortClause`
- **Evaluator** (`evaluator.go`): In-memory evaluation checking each issue against the AST
- **Executor** (`execute.go`): Orchestrates parse -> validate -> fetch -> evaluate pipeline
- **QuerySource** (`source.go`): Interface abstracting DB operations for testability

**Size:** 3,419 source lines across 6 files

### Sync Engine -- `internal/sync/`

Event-sourced synchronization with:
- **Event types** (`types.go`): `Event`, `PushResult`, `PullResult`, `ApplyResult`, `ConflictRecord`
- **Server engine** (`engine.go`): Server-side event insertion with deduplication by `(device_id, session_id, client_action_id)`, monotonic `server_seq` assignment
- **Client logic** (`client.go`): Builds events from `action_log`, pushes to server, pulls and applies remote events with conflict detection
- **Entity mapping**: Supports 11 entity types (issues, handoffs, boards, logs, comments, work_sessions, board_issue_positions, issue_dependencies, issue_files, work_session_issues, notes)

**Size:** 1,520 source lines

### Session Management -- `internal/session/`

- **Agent fingerprinting** (`agent_fingerprint.go`): Walks process tree (up to 15 levels) to identify agent type from process names; detects 10 agent types
- **Session scoping** (`session.go`): Sessions scoped by `branch + agent fingerprint`; heartbeat tracking for liveness
- **Migration**: One-time migration from legacy filesystem-based sessions to DB-backed sessions

**Size:** 493 source lines + 180 lines for agent fingerprinting

### Modal System -- `pkg/monitor/modal/`

Declarative modal framework:
- Sections: text, list, button, custom, input
- Internal scrolling and focus cycling
- Mouse interaction support
- Stacking support (modals can open modals)

**Size:** 1,557 source lines

### Keymap System -- `pkg/monitor/keymap/`

Context-aware keyboard binding registry:
- 17 contexts (main, modal, search, board, kanban, form, confirm, etc.)
- 60+ command constants
- Key sequences (e.g., `gg` for jump to top)
- Configurable bindings
- Help text generation

**Size:** 1,715 source lines

### Workflow State Machine -- `internal/workflow/`

Issue status transition management:
- 15 valid transitions across 5 statuses
- 3 modes: Liberal (default, skips guards), Advisory (warns), Strict (blocks)
- Pluggable guards: `DifferentReviewerGuard`, `BlockedGuard` (active); `EpicChildrenGuard`, `SelfCloseGuard`, `InProgressRequiredGuard` (defined, not yet wired)

**Size:** 588 source lines

---

## Development Environment Requirements

### Minimum Requirements

| Tool | Version | Purpose |
|---|---|---|
| Go | 1.25.5+ | Compilation |
| Git | Any recent | Version control, branch detection, git snapshots |
| C compiler (gcc/clang) | Any | Required for default `mattn/go-sqlite3` CGo driver |

### Optional (for pure-Go builds)

If a C compiler is unavailable, use the `modernc.org/sqlite` pure-Go driver via build tags. This removes the CGo requirement at the cost of slightly different SQLite behavior.

### Testing

| Tool | Purpose |
|---|---|
| `go test` | Standard Go test runner |
| Real `td` binary | E2E tests in `test/e2e/` build and run the actual binary in temp directories |
| Custom sync harness | `test/syncharness/` simulates concurrent agent sessions with chaos testing |

### Documentation Site

| Tool | Version | Purpose |
|---|---|---|
| Node.js | (for Docusaurus) | Documentation website at `website/` |
| Docusaurus | Configured in `website/docusaurus.config.js` | Static site generation for docs |

---

## Codebase Metrics

| Metric | Value |
|---|---|
| Total Go files | 332 |
| Total lines of code | ~126K |
| Source files (non-test) | 198 |
| Source lines | ~59K |
| Test files | 134 |
| Test lines | ~67K |
| Test-to-source ratio | 1.14:1 |
| Packages | 29 |
| SQLite schema version | 29 |
| SQLite tables | 20 |
| Cobra commands | ~50 (non-test) |
| TUI message types | ~25 |
| Keymap commands | ~60 |
| Keymap contexts | 17 |
| Direct Go module dependencies | 18 |
| Indirect Go module dependencies | 33 |

### Largest Packages by Source Lines

| Package | Lines | Description |
|---|---|---|
| `cmd/` | 13,529 | CLI commands (50+ Cobra commands) |
| `pkg/monitor/` | 13,026 | TUI monitor (BubbleTea) |
| `internal/db/` | 7,835 | SQLite persistence layer |
| `test/e2e/` | 3,975 | End-to-end tests |
| `internal/api/` | 3,922 | Sync server HTTP API |
| `internal/query/` | 3,419 | TDQ query language |
| `internal/serverdb/` | 1,947 | Server-side database |
| `pkg/monitor/keymap/` | 1,715 | Keymap system |
| `pkg/monitor/modal/` | 1,557 | Modal system |
| `internal/sync/` | 1,520 | Sync engine |
