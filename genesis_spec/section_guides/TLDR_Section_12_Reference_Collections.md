# Section 12: Reference Collections

## Overview

This section provides complete reference tables for all CLI commands, the TDQ query language, TUI keyboard bindings, configuration keys, database schema, and sync server API endpoints. All data is verified against source code, not documentation claims.

---

## 12.1 CLI Command Reference

**Source**: `/home/devuser/td/cmd/root.go` (command groups), all `cmd/*.go` files

Commands are organized into 7 groups, each registered via Cobra's `GroupID` field. The root command uses a custom usage template that displays aliases inline.

### Core Commands

| Command | Aliases | Description | Source |
|---------|---------|-------------|--------|
| `create [title]` | `add`, `new` | Create a new issue | `cmd/create.go` |
| `list [filters]` | `ls` | List issues matching given filters | `cmd/list.go` |
| `show [issue-id...]` | `context`, `view`, `get` | Display full details of one or more issues | `cmd/show.go` |
| `update [issue-id...]` | `edit` | Update one or more fields on existing issues | `cmd/update.go` |
| `delete [issue-id...]` | | Soft-delete one or more issues | `cmd/delete.go` |
| `restore [issue-id...]` | | Restore soft-deleted issues | `cmd/delete.go` |
| `board` | | Manage issue boards (subcommands below) | `cmd/board.go` |
| `epic` | | Shortcuts for working with epics (subcommands below) | `cmd/epic.go` |
| `task` | | Shortcuts for working with tasks (subcommands below) | `cmd/task.go` |
| `note` | | Manage freeform notes (subcommands below) | `cmd/note.go` |
| `assign <issue-id> <agent>` | `delegate` | Assign an issue to an agent | `cmd/assign.go` |
| `activity [issue-id]` | `act` | Show agent activity for an issue or agent | `cmd/activity.go` |

### Workflow Commands

| Command | Aliases | Description | Source |
|---------|---------|-------------|--------|
| `start [issue-id...]` | `begin` | Begin work on issue(s) | `cmd/start.go` |
| `unstart [issue-id...]` | `stop` | Revert issue(s) from in_progress to open | `cmd/unstart.go` |
| `review [issue-id...]` | `submit`, `finish` | Submit one or more issues for review | `cmd/review.go` |
| `approve [issue-id...]` | | Approve and close one or more issues | `cmd/review.go` |
| `reject [issue-id...]` | | Reject and return to in_progress | `cmd/review.go` |
| `close [issue-id...]` | `done`, `complete` | Close one or more issues without review | `cmd/review.go` |
| `block [issue-id...]` | | Mark issue(s) as blocked | `cmd/block.go` |
| `unblock [issue-id...]` | | Unblock issue(s) back to open status | `cmd/block.go` |
| `reopen [issue-id...]` | | Reopen closed issues | `cmd/block.go` |
| `handoff <issue-id> [message]` | | Capture structured working state | `cmd/handoff.go` |
| `log [issue-id] <message>` | | Append a log entry to the current issue | `cmd/log.go` |
| `comment [issue-id] "text"` | | Add a comment to an issue | `cmd/tree.go` |
| `comments [issue-id]` | | List comments for an issue | `cmd/tree.go` |
| `dep` | | Manage dependencies between issues (subcommands below) | `cmd/dependencies.go` |
| `tree add-child` | | Add parent/child relationship | `cmd/tree.go` |

### Query Commands

| Command | Aliases | Description | Source |
|---------|---------|-------------|--------|
| `query [expression]` | | Search issues with TDQ query language | `cmd/query.go` |
| `search [query]` | | Full-text search across issues | `cmd/search.go` |
| `tree [issue-id]` | | Visualize parent/child relationships | `cmd/tree.go` |
| `blocked-by [issue-id]` | | Show what issues are waiting on this issue | `cmd/dependencies.go` |
| `depends-on [issue-id]` | `deps`, `dependencies` | Show what this issue depends on | `cmd/dependencies.go` |
| `critical-path` | | Show sequence of issues that unblocks the most work | `cmd/dependencies.go` |

### Shortcuts

| Command | Description | Source |
|---------|-------------|--------|
| `reviewable` | Show issues awaiting review that you can review | `cmd/list.go` |
| `blocked` | List blocked issues | `cmd/list.go` |
| `in-review` (alias: `ir`) | List all issues currently in review | `cmd/list.go` |
| `ready` | List open issues sorted by priority | `cmd/list.go` |
| `next` | Show highest-priority open issue | `cmd/list.go` |
| `deleted` | Show soft-deleted issues | `cmd/list.go` |

### Session Commands

| Command | Aliases | Description | Source |
|---------|---------|-------------|--------|
| `focus [issue-id]` | | Set the current working issue | `cmd/focus.go` |
| `unfocus` | | Clear focus | `cmd/focus.go` |
| `status` | `current` | Show dashboard: session, focus, reviews, blocked, ready | `cmd/status.go` |
| `resume [issue-id]` | | Show context and set focus | `cmd/context.go` |
| `usage` | | Generate optimized context block for AI agents | `cmd/context.go` |
| `session [name]` | | Name session, or --new at context start | `cmd/system.go` |
| `whoami` | | Show current session identity | `cmd/system.go` |
| `ws` (alias: `worksession`) | | Work session commands (subcommands below) | `cmd/ws.go` |
| `check-handoff` | | Check if handoff is needed before exiting | `cmd/focus.go` |

### File Commands

| Command | Description | Source |
|---------|-------------|--------|
| `link [issue-id] [file-pattern...]` | Link files to an issue | `cmd/link.go` |
| `unlink [issue-id] [file-pattern]` | Remove file associations | `cmd/link.go` |
| `files [issue-id]` | List linked files with change status | `cmd/link.go` |

### System Commands

| Command | Aliases | Description | Source |
|---------|---------|-------------|--------|
| `init` | | Initialize a new td project | `cmd/init.go` |
| `monitor` | | Live TUI dashboard for observing agent activity | `cmd/monitor.go` |
| `version` | | Show version and check for updates | `cmd/system.go` |
| `info` | `stats` | Show database statistics and project overview | `cmd/system.go` |
| `upgrade` | | Run database migrations | `cmd/system.go` |
| `export` | | Export database | `cmd/system.go` |
| `import [file]` | | Import issues | `cmd/system.go` |
| `undo` | | Undo the last action | `cmd/undo.go` |
| `last` | | Show the last action performed | `cmd/undo.go` |
| `doctor` | | Run diagnostic checks for sync setup | `cmd/doctor.go` |
| `config` | | Manage td configuration (subcommands below) | `cmd/config.go` |
| `feature` | | Manage experimental feature flags (subcommands below) | `cmd/feature.go` |
| `workflow` | | Show issue status workflow | `cmd/workflow.go` |
| `sync` | | Sync local data with remote server | `cmd/sync.go` |
| `auth` | | Manage sync authentication (subcommands below) | `cmd/auth.go` |
| `sync-project` | `sp` | Manage sync projects (subcommands below) | `cmd/project.go` |
| `security` | | View security exception log (self-close exceptions) | `cmd/security.go` |
| `errors` | | View failed td command attempts | `cmd/errors.go` |
| `debug-stats` | | Output runtime memory and goroutine statistics (JSON) | `cmd/debug_stats.go` |

### Notable Subcommands

**Board subcommands** (`td board <sub>`): `list`, `create <name>`, `delete <board>`, `show <board>`, `edit <board>`, `move <board> <issue-id> <position>`, `unposition <board> <issue-id>`

**Epic subcommands** (`td epic <sub>`): `create [title]`, `list`

**Task subcommands** (`td task <sub>`): `create [title]`, `list`

**Note subcommands** (`td note <sub>`): `add <title>`, `list` (alias: `ls`), `show <id>`, `edit <id>`, `delete <id>`, `pin <id>`, `unpin <id>`, `archive <id>`, `unarchive <id>`

**Work session subcommands** (`td ws <sub>`): `start [name]`, `tag [issue-ids...]`, `untag [issue-ids...]`, `log "message"`, `current`, `handoff`, `end`, `list`, `show [session-id]`

**Dep subcommands** (`td dep <sub>`): `add <issue> <depends-on>...`, `rm <issue> <depends-on>` (alias: `remove`)

**Session subcommands** (`td session <sub>`): `list`, `cleanup`

**Config subcommands** (`td config <sub>`): `set <key> <value>`, `get <key>`, `list`

**Feature subcommands** (`td feature <sub>`): `list`, `get <name>`, `set <name> <true|false>`, `unset <name>`

**Auth subcommands** (`td auth <sub>`): `login`, `logout`, `status`

**Sync subcommands** (`td sync <sub>`): `init`, `conflicts`, `tail`

**Sync-project subcommands** (`td sync-project <sub>`): `create <name>`, `link <project-id>`, `unlink`, `list`, `members`, `invite <email> [role]`, `kick <user-id>`, `role <user-id> <role>`, `join [name-or-id]`

**Stats subcommands** (`td stats <sub>`): `analytics` (alias: `usage`), `security`, `errors`

---

## 12.2 TDQ Query Language Reference

**Source**: `/home/devuser/td/internal/query/parser.go`, `lexer.go`, `ast.go`

TDQ (td query) is a custom query language for filtering issues. It is parsed by a hand-written recursive descent parser with a separate lexer phase. The grammar supports field comparisons, boolean logic, function calls, text search, and sort clauses.

### Grammar Overview

```
query       := or_expr [sort_clause]
or_expr     := and_expr ("OR" | "||" and_expr)*
and_expr    := unary (("AND" | "&&") unary | implicit_and)*
unary       := ("NOT" | "!" | "-") unary | primary
primary     := "(" query ")" | field_expr | function_call | text_search
field_expr  := field_name operator value
function    := name "(" [args] ")"
text_search := identifier | quoted_string
sort_clause := "sort:" ["-"] field_name
```

**Implicit AND**: Two adjacent expressions are implicitly ANDed. `status = open priority <= P1` is equivalent to `status = open AND priority <= P1`.

**Maximum nesting depth**: 50 levels (enforced by `MaxQueryDepth`).

### Operators

| Operator | Lexer Token | Description |
|----------|-------------|-------------|
| `=` | `TokenEq` | Equality |
| `!=` | `TokenNeq` | Inequality |
| `<` | `TokenLt` | Less than |
| `>` | `TokenGt` | Greater than |
| `<=` | `TokenLte` | Less than or equal |
| `>=` | `TokenGte` | Greater than or equal |
| `~` | `TokenContains` | Contains (substring match) |
| `!~` | `TokenNotContains` | Does not contain |
| `IN` | `OpIn` | Value in list |
| `NOT IN` | `OpNotIn` | Value not in list |

**Legacy syntax**: The `:` character is treated as `=` for backward compatibility (e.g., `status:open`).

**Boolean operators**: `AND` / `&&`, `OR` / `||`, `NOT` / `!` / `-`

**Shell escape handling**: The lexer strips backslashes before operator characters (`\!`, `\<`, `\>`, `\=`, `\~`) to accommodate agents that escape shell metacharacters.

### Fields

**Issue Fields** (from `KnownFields` in `ast.go`):

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Issue ID |
| `title` | string | Issue title |
| `description` | string | Issue description |
| `status` | enum | `open`, `in_progress`, `blocked`, `in_review`, `closed` |
| `type` | enum | `bug`, `feature`, `task`, `epic`, `chore` |
| `priority` | ordinal | `P0`, `P1`, `P2`, `P3`, `P4` |
| `points` | number | Story points |
| `labels` | string | Comma-separated labels |
| `parent` | string | Parent issue ID |
| `epic` | string | Epic issue ID |
| `implementer` | string | Implementer session ID |
| `reviewer` | string | Reviewer session ID |
| `minor` | bool | Minor flag (self-reviewable) |
| `branch` | string | Created branch |
| `sprint` | string | Sprint identifier |
| `created` | date | Created timestamp |
| `updated` | date | Updated timestamp |
| `closed` | date | Closed timestamp |

### Cross-Entity Fields

Six entity prefixes support dot-notation queries across related data:

| Prefix | Sub-fields | Example |
|--------|-----------|---------|
| `log` | `message` (string), `type` (enum: progress/blocker/decision/hypothesis/tried/result/orchestration), `timestamp` (date), `session` (string) | `log.type = decision` |
| `comment` | `text` (string), `created` (date), `session` (string) | `comment.text ~ "bug"` |
| `handoff` | `done` (string), `remaining` (string), `decisions` (string), `uncertain` (string), `timestamp` (date) | `handoff.remaining ~ "auth"` |
| `file` | `path` (string), `role` (enum: implementation/test/reference/config) | `file.role = test` |
| `dep` | `blocks` (string), `depends_on` (string) | `dep.blocks = td-abc123` |
| `note` | `title` (string), `content` (string), `created` (date), `updated` (date), `pinned` (bool), `archived` (bool) | `note.pinned = true` |

### Functions

From `KnownFunctions` in `ast.go` (14 functions total):

| Function | Args | Description |
|----------|------|-------------|
| `has(field)` | 1 | Field is not empty |
| `is(status)` | 1 | Shorthand for status check (e.g., `is(open)`) |
| `any(field, v1, v2, ...)` | 2+ | Field matches any listed value |
| `all(field, v1, v2, ...)` | 2+ | Field matches all listed values |
| `none(field, v1, v2, ...)` | 2+ | Field matches none of the listed values |
| `blocks(id)` | 1 | Issues that block the given ID |
| `blocked_by(id)` | 1 | Issues blocked by the given ID |
| `child_of(id)` | 1 | Direct children of issue |
| `descendant_of(id)` | 1 | All descendants (recursive) |
| `linked_to(path)` | 1 | Issues linked to file path |
| `rework()` | 0 | Issues rejected and awaiting rework |
| `is_ready()` | 0 | Issues with no open dependencies |
| `has_open_deps()` | 0 | Issues with open dependencies |
| `label(name)` / `labels(name)` | 1 | Issues with the given label |

**Note**: The `stale(N)` function is documented in the td skill file but does NOT exist in the KnownFunctions map. It is not implemented in the query engine.

### Special Values

| Value | Token | Description |
|-------|-------|-------------|
| `@me` | `TokenAtMe` | Current session identity |
| `EMPTY` | `TokenEmpty` | Field is empty/blank |
| `NULL` | `TokenNull` | Field is null |

### Date Values

**Absolute dates**: `2024-01-15` (YYYY-MM-DD format, detected by pattern length and hyphen positions)

**Relative offsets**: `-7d`, `+3w`, `-1m`, `-3h` (units: `d`=day, `w`=week, `m`=month, `h`=hour)

**Named dates**: `today`, `yesterday`, `this_week`, `last_week`, `this_month`, `last_month`

### Sort Clause

**Syntax**: `sort:field` (ascending) or `sort:-field` (descending)

**Valid sort fields** (from `SortFieldToColumn` and `scanSortClause` validation):

| User Field | DB Column |
|-----------|-----------|
| `created` | `created_at` |
| `updated` | `updated_at` |
| `closed` | `closed_at` |
| `deleted` | `deleted_at` |
| `priority` | `priority` |
| `id` | `id` |
| `title` | `title` |
| `status` | `status` |
| `points` | `points` |

Only one sort clause is allowed per query. Multiple sort clauses produce a parse error.

### Notes-Specific Sort Fields

When querying notes, a separate mapping applies (from `NoteSortFieldToColumn`):

| User Field | DB Column |
|-----------|-----------|
| `created` | `created_at` |
| `updated` | `updated_at` |
| `title` | `title` |
| `pinned` | `pinned` |
| `archived` | `archived` |

### Query Examples

```
status = open                              # Simple field comparison
type = bug AND priority <= P1              # Boolean AND with ordinal comparison
labels ~ backend AND status != closed      # Contains operator with NOT EQUAL
created >= -7d                             # Relative date
implementer = @me AND is(in_progress)      # Special value + function
rework()                                   # Zero-arg function
has(labels) AND NOT is(closed)             # Function + NOT
log.type = decision                        # Cross-entity field
status = open sort:-priority               # With descending sort
"authentication error"                     # Bare text search
```

---

## 12.3 Keyboard Shortcut Reference

**Source**: `/home/devuser/td/pkg/monitor/keymap/bindings.go`, `registry.go`

The keymap system uses a `Registry` with context-based binding resolution. Key lookup order: user overrides for active context, user overrides for global, context-specific bindings, global bindings. Multi-key sequences (e.g., `g g`) are supported with a 500ms timeout.

### 20 UI Contexts

From `registry.go`:

| Context | Description |
|---------|-------------|
| `global` | Active everywhere unless overridden |
| `main` | Main panel, no modal open |
| `modal` | Issue details modal |
| `stats` | Statistics modal |
| `search` | Search input focused |
| `confirm` | Confirmation dialog |
| `epic-tasks` | Task list in epic modal |
| `parent-epic-focused` | Parent epic row focused in modal |
| `blocked-by-focused` | Blocked-by section focused in modal |
| `blocks-focused` | Blocks section focused in modal |
| `handoffs` | Handoffs modal |
| `form` | Form modal (create/edit) |
| `help` | Help modal |
| `board-picker` | Board picker overlay |
| `board` | Board mode (swimlanes/backlog) |
| `board-kanban` | Kanban board view |
| `board-editor` | Board edit/create modal |
| `getting-started` | Getting started modal |
| `tdq-help` | TDQ query help modal |
| `close-confirm` | Close confirmation with text input |
| `td-sync-prompt` | Sync prompt modal |

### Global Bindings

| Key | Command | Description |
|-----|---------|-------------|
| `q` | quit | Quit |
| `ctrl+c` | quit | Quit |
| `?` | toggle-help | Toggle help |

### Main Panel Bindings

| Key | Command | Description |
|-----|---------|-------------|
| `tab` | next-panel | Next panel |
| `shift+tab` | prev-panel | Previous panel |
| `esc` | search-clear | Clear search filter |
| `j` / `down` | cursor-down | Move down |
| `k` / `up` | cursor-up | Move up |
| `ctrl+d` | half-page-down | Half page down |
| `ctrl+u` | half-page-up | Half page up |
| `ctrl+f` | full-page-down | Full page down |
| `ctrl+b` | full-page-up | Full page up |
| `G` | cursor-bottom | Go to bottom |
| `g g` | cursor-top | Go to top (multi-key sequence) |
| `home` | cursor-top | Go to top |
| `end` | cursor-bottom | Go to bottom |
| `enter` | open-details | Open issue details modal |
| `s` | open-stats | Open statistics |
| `h` | open-handoffs | Open handoffs modal |
| `H` | open-getting-started | Open getting started guide |
| `/` | search | Enter search mode |
| `c` | toggle-closed | Toggle closed tasks |
| `S` | cycle-sort-mode | Cycle sort mode |
| `T` | cycle-type-filter | Cycle type filter |
| `r` | mark-for-review | Review/Refresh |
| `R` | mark-for-review | Submit for review |
| `a` | approve | Approve issue |
| `x` | delete | Delete issue |
| `C` | close-issue | Close issue |
| `O` | reopen-issue | Reopen issue |
| `n` | new-issue | New issue |
| `e` | edit-issue | Edit issue |
| `y` | copy-to-clipboard | Copy issue as markdown |
| `Y` | copy-id-to-clipboard | Copy issue ID |
| `W` | send-to-worktree | Send to worktree |
| `b` | boards | Open board picker |

### Modal Bindings (Issue Details)

| Key | Command | Description |
|-----|---------|-------------|
| `esc` / `enter` | close | Close modal |
| `j` / `down` | scroll-down | Scroll down |
| `k` / `up` | scroll-up | Scroll up |
| `ctrl+d` | half-page-down | Half page down |
| `ctrl+u` | half-page-up | Half page up |
| `ctrl+f` / `pgdown` | full-page-down | Full page down |
| `ctrl+b` / `pgup` | full-page-up | Full page up |
| `G` | cursor-bottom | Go to bottom |
| `g g` / `home` | cursor-top | Go to top |
| `end` | cursor-bottom | Go to bottom |
| `h` / `left` | navigate-prev | Previous issue |
| `l` / `right` | navigate-next | Next issue |
| `r` | refresh | Refresh |
| `R` | mark-for-review | Submit for review |
| `y` | copy-to-clipboard | Copy to clipboard |
| `Y` | copy-id-to-clipboard | Copy issue ID |
| `n` | new-issue | New issue |
| `e` | edit-issue | Edit issue |
| `x` | delete | Delete issue |
| `C` | close-issue | Close issue |
| `O` | reopen-issue | Reopen issue |
| `W` | send-to-worktree | Send to worktree |
| `tab` | focus-task-section | Focus task list (in epic view) |

### Search Mode Bindings

| Key | Command | Description |
|-----|---------|-------------|
| `esc` | search-cancel | Cancel search |
| `enter` | search-confirm | Apply search |
| `ctrl+u` | search-clear | Clear search |
| `ctrl+w` | search-clear | Clear search |

### Confirmation Dialog Bindings

| Key | Command | Description |
|-----|---------|-------------|
| `y` / `Y` | confirm | Confirm |
| `n` / `N` | cancel | Cancel |
| `esc` | cancel | Cancel |
| `tab` | next-button | Next button |
| `shift+tab` | prev-button | Previous button |
| `enter` | select | Execute focused button |

### Form Modal Bindings

| Key | Command | Description |
|-----|---------|-------------|
| `ctrl+s` | form-submit | Submit form |
| `esc` | form-cancel | Cancel form |
| `ctrl+x` | form-toggle-extend | Toggle extended fields |
| `ctrl+o` | form-open-editor | Open in external editor |

### Board Mode Bindings (Swimlanes/Backlog)

Includes all main panel actions plus:

| Key | Command | Description |
|-----|---------|-------------|
| `J` | move-up | Move issue down in board |
| `K` | move-down | Move issue up in board |
| `ctrl+j` | move-to-bottom | Move issue to bottom |
| `ctrl+k` | move-to-top | Move issue to top |
| `v` | view | Toggle swimlanes/backlog view |
| `F` | status-filter | Cycle status filter |
| `c` | closed | Toggle closed |
| `esc` | exit | Exit to All Issues |
| `b` | boards | Open board picker |

### Board Kanban Bindings

| Key | Command | Description |
|-----|---------|-------------|
| `h` / `left` | kanban-prev-column | Previous column |
| `l` / `right` | kanban-next-column | Next column |
| `j` / `down` | cursor-down | Move down in column |
| `k` / `up` | cursor-up | Move up in column |
| `v` | view | Cycle view mode |
| `F` | status-filter | Cycle status filter |
| `c` | closed | Toggle closed |
| `esc` | exit | Exit to All Issues |

All standard issue actions (n, e, x, a, R, C, O, y, Y, W, s, S, T) are also available in kanban mode.

### Board Picker Bindings

| Key | Command | Description |
|-----|---------|-------------|
| `j` / `down` | cursor-down | Move down |
| `k` / `up` | cursor-up | Move up |
| `enter` | select-board | Select board |
| `e` | edit-board | Edit board |
| `n` | new-board | New board |
| `esc` / `q` | close-picker | Close picker |

### Board Editor Bindings

| Key | Command | Description |
|-----|---------|-------------|
| `ctrl+s` | board-editor-save | Save board |
| `esc` | board-editor-cancel | Cancel |

### Getting Started Modal Bindings

| Key | Command | Description |
|-----|---------|-------------|
| `I` | install-instructions | Install agent instructions |
| `esc` / `q` | close | Close modal |

### User-Configurable Overrides

The Registry supports user overrides via `SetUserOverride(context, key, command)`. User overrides take highest precedence in key resolution. There is infrastructure for a `keymap.json` configuration file, but no documentation or template generator exists for end users.

---

## 12.4 Configuration Reference

### Project Configuration (`config.json`)

**Source**: `/home/devuser/td/internal/config/config.go`

Stored at `.todos/config.json`, serialized with `flock`-based file locking.

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `focus` | string | `""` | Currently focused issue ID |
| `work_session_id` | string | `""` | Active work session ID |
| `pane_heights` | []float64 | `[0.25, 0.45, 0.30]` | Three-panel height ratios |
| `feature_flags` | map[string]bool | `{}` | Per-project feature flag overrides |
| `filter_state.search_query` | string | `""` | Active search/TDQ query |
| `filter_state.sort_mode` | string | `""` | Current sort mode |
| `filter_state.type_filter` | string | `""` | Type filter |
| `filter_state.include_closed` | bool | `false` | Show closed issues |
| `title_validation.min_length` | int | `15` | Minimum issue title length |
| `title_validation.max_length` | int | `100` | Maximum issue title length |

### Sync Configuration (`~/.config/td/config.json`)

**Source**: `/home/devuser/td/cmd/config.go`

System-wide (not per-project), managed via `td config set/get`. Feature-gated behind `sync_cli`.

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `sync.url` | string | (none) | Sync server URL |
| `sync.enabled` | bool | `false` | Enable sync |
| `sync.auto.enabled` | bool | `true` | Auto-sync on mutations |
| `sync.auto.debounce` | duration | `3s` | Debounce interval |
| `sync.auto.interval` | duration | `5m` | Periodic sync interval |
| `sync.auto.pull` | bool | `true` | Auto-pull on sync |
| `sync.auto.on_start` | bool | `true` | Sync on startup |
| `sync.snapshot_threshold` | int | `100` | Events before snapshot |

### Feature Flags

**Source**: `/home/devuser/td/internal/features/features.go`

Resolution order: Environment variable override > project config > default value.

| Flag | Default | Description |
|------|---------|-------------|
| `sync_cli` | `false` | Gates sync/auth CLI commands |
| `sync_autosync` | `false` | Gates auto-sync hooks |
| `sync_monitor_prompt` | `false` | Gates first-run sync prompt in monitor |
| `sync_notes` | `true` | Gates notes entity sync |

### Client Environment Variables

| Variable | Description |
|----------|-------------|
| `TD_ANALYTICS` | Set to `false` to disable command analytics |
| `TD_LOG_FILE` | Redirect slog output to file path |
| `TD_DISABLE_EXPERIMENTAL` | Kill switch for all experimental features |
| `TD_FEATURE_<NAME>` | Override specific feature flag (e.g., `TD_FEATURE_SYNC_CLI=true`) |
| `TD_DISABLE_FEATURE` | Comma-separated features to disable |
| `TD_ENABLE_FEATURE` | Comma-separated features to enable |

### Server Environment Variables

**Source**: `/home/devuser/td/internal/api/config.go`

| Variable | Default | Description |
|----------|---------|-------------|
| `SYNC_LISTEN_ADDR` | `:8080` | HTTP listen address |
| `SYNC_SERVER_DB_PATH` | `./data/server.db` | Server metadata SQLite path |
| `SYNC_PROJECT_DATA_DIR` | `./data/projects` | Per-project event store directory |
| `SYNC_SHUTDOWN_TIMEOUT` | `30s` | Graceful shutdown timeout |
| `SYNC_ALLOW_SIGNUP` | `true` | Allow new user registration |
| `SYNC_BASE_URL` | `http://localhost:8080` | Public base URL (for auth pages) |
| `SYNC_LOG_FORMAT` | `json` | Log format: `json` or `text` |
| `SYNC_LOG_LEVEL` | `info` | Log level: `debug`, `info`, `warn`, `error` |
| `SYNC_RATE_LIMIT_AUTH` | `10` | Auth requests per IP per minute |
| `SYNC_RATE_LIMIT_PUSH` | `60` | Push requests per API key per minute |
| `SYNC_RATE_LIMIT_PULL` | `120` | Pull requests per API key per minute |
| `SYNC_RATE_LIMIT_OTHER` | `300` | Other requests per API key per minute |
| `SYNC_CORS_ALLOWED_ORIGINS` | (empty) | Comma-separated CORS origins for admin API |
| `SYNC_AUTH_EVENT_RETENTION` | `90d` | Auth event log retention period |
| `SYNC_RATE_LIMIT_EVENT_RETENTION` | `30d` | Rate limit event log retention period |

---

## 12.5 Database Schema Reference

**Source**: `/home/devuser/td/internal/db/schema.go`

Current schema version: **29** (29 migrations from version 1).

### Core Tables (Base Schema)

#### `issues`

| Column | Type | Default | Notes |
|--------|------|---------|-------|
| `id` | TEXT PK | | Issue ID (e.g., `td-abc123`) |
| `title` | TEXT NOT NULL | | Issue title |
| `description` | TEXT | `''` | Issue description |
| `status` | TEXT NOT NULL | `'open'` | `open`, `in_progress`, `blocked`, `in_review`, `closed` |
| `type` | TEXT NOT NULL | `'task'` | `bug`, `feature`, `task`, `epic`, `chore` |
| `priority` | TEXT NOT NULL | `'P2'` | `P0`-`P4` |
| `points` | INTEGER | `0` | Story points |
| `labels` | TEXT | `''` | Comma-separated |
| `parent_id` | TEXT | `''` | FK to `issues(id)` |
| `acceptance` | TEXT | `''` | Acceptance criteria |
| `implementer_session` | TEXT | `''` | Session that implemented |
| `reviewer_session` | TEXT | `''` | Session that reviewed |
| `created_at` | DATETIME NOT NULL | `CURRENT_TIMESTAMP` | |
| `updated_at` | DATETIME NOT NULL | `CURRENT_TIMESTAMP` | |
| `closed_at` | DATETIME | | |
| `deleted_at` | DATETIME | | Soft delete marker |
| `minor` | INTEGER | `0` | Self-reviewable flag (v4) |
| `created_branch` | TEXT | `''` | Branch at creation (v5) |
| `creator_session` | TEXT | `''` | Creating session (v6) |
| `sprint` | TEXT | `''` | Sprint identifier (v10) |
| `assignee` | TEXT | `''` | Assigned agent name (v29) |
| `linear_id` | TEXT | `''` | Linear issue UUID (v29) |
| `linear_identifier` | TEXT | `''` | Linear short ID (v29) |
| `project_tag` | TEXT | `''` | Project tag (v29) |

**Indexes**: `status`, `priority`, `type`, `parent_id`, `deleted_at`, `(deleted_at, status)`, `assignee`, `project_tag`, `linear_id`

#### `logs`

| Column | Type | Default | Notes |
|--------|------|---------|-------|
| `id` | TEXT PK | | Log entry ID |
| `issue_id` | TEXT | `''` | Can be empty for work session logs |
| `session_id` | TEXT NOT NULL | | Session that created the log |
| `work_session_id` | TEXT | `''` | Work session association |
| `message` | TEXT NOT NULL | | Log message |
| `type` | TEXT NOT NULL | `'progress'` | `progress`, `blocker`, `decision`, `hypothesis`, `tried`, `result`, `orchestration` |
| `timestamp` | DATETIME NOT NULL | `CURRENT_TIMESTAMP` | |

**Indexes**: `issue_id`, `work_session_id`, `timestamp`

#### `handoffs`

| Column | Type | Default | Notes |
|--------|------|---------|-------|
| `id` | TEXT PK | | |
| `issue_id` | TEXT NOT NULL | | FK to `issues(id)` |
| `session_id` | TEXT NOT NULL | | |
| `done` | TEXT | `'[]'` | JSON array |
| `remaining` | TEXT | `'[]'` | JSON array |
| `decisions` | TEXT | `'[]'` | JSON array |
| `uncertain` | TEXT | `'[]'` | JSON array |
| `timestamp` | DATETIME NOT NULL | `CURRENT_TIMESTAMP` | |

**Indexes**: `issue_id`, `timestamp`

#### `sessions`

| Column | Type | Default | Notes |
|--------|------|---------|-------|
| `id` | TEXT PK | | Session ID |
| `name` | TEXT | `''` | User-assigned name |
| `branch` | TEXT | `''` | Git branch |
| `agent_type` | TEXT | `''` | Agent identifier |
| `agent_pid` | INTEGER | `0` | Process ID |
| `context_id` | TEXT | `''` | |
| `previous_session_id` | TEXT | `''` | |
| `started_at` | DATETIME NOT NULL | `CURRENT_TIMESTAMP` | |
| `ended_at` | DATETIME | | |
| `last_activity` | DATETIME | | |

**Indexes**: `branch`, `(branch, agent_type, agent_pid)`

#### `comments`

| Column | Type | Default |
|--------|------|---------|
| `id` | TEXT PK | |
| `issue_id` | TEXT NOT NULL | FK to `issues(id)` |
| `session_id` | TEXT NOT NULL | |
| `text` | TEXT NOT NULL | |
| `created_at` | DATETIME NOT NULL | `CURRENT_TIMESTAMP` |

**Indexes**: `issue_id`, `created_at`

#### `git_snapshots`

| Column | Type | Default |
|--------|------|---------|
| `id` | TEXT PK | |
| `issue_id` | TEXT NOT NULL | FK to `issues(id)` |
| `event` | TEXT NOT NULL | |
| `commit_sha` | TEXT NOT NULL | |
| `branch` | TEXT NOT NULL | |
| `dirty_files` | INTEGER | `0` |
| `timestamp` | DATETIME NOT NULL | `CURRENT_TIMESTAMP` |

#### `issue_files`

| Column | Type | Default | Notes |
|--------|------|---------|-------|
| `id` | TEXT PK | | |
| `issue_id` | TEXT NOT NULL | | FK to `issues(id)` |
| `file_path` | TEXT NOT NULL | | Repo-relative path (normalized in v19) |
| `role` | TEXT NOT NULL | `'implementation'` | `implementation`, `test`, `reference`, `config` |
| `linked_sha` | TEXT | `''` | |
| `linked_at` | DATETIME NOT NULL | `CURRENT_TIMESTAMP` | |

**Constraint**: `UNIQUE(issue_id, file_path)`

#### `issue_dependencies`

| Column | Type | Default |
|--------|------|---------|
| `id` | TEXT PK | |
| `issue_id` | TEXT NOT NULL | FK to `issues(id)` |
| `depends_on_id` | TEXT NOT NULL | FK to `issues(id)` |
| `relation_type` | TEXT NOT NULL | `'depends_on'` |

**Constraint**: `UNIQUE(issue_id, depends_on_id, relation_type)`

#### `work_sessions`

| Column | Type | Default |
|--------|------|---------|
| `id` | TEXT PK | |
| `name` | TEXT NOT NULL | |
| `session_id` | TEXT NOT NULL | |
| `started_at` | DATETIME NOT NULL | `CURRENT_TIMESTAMP` |
| `ended_at` | DATETIME | |
| `start_sha` | TEXT | `''` |
| `end_sha` | TEXT | `''` |

#### `work_session_issues`

| Column | Type | Default |
|--------|------|---------|
| `id` | TEXT PK | |
| `work_session_id` | TEXT NOT NULL | FK to `work_sessions(id)` |
| `issue_id` | TEXT NOT NULL | FK to `issues(id)` |
| `tagged_at` | DATETIME NOT NULL | `CURRENT_TIMESTAMP` |

**Constraint**: `UNIQUE(work_session_id, issue_id)`

#### `schema_info`

| Column | Type |
|--------|------|
| `key` | TEXT PK |
| `value` | TEXT NOT NULL |

### Migration-Added Tables

#### `action_log` (v2)

| Column | Type | Default | Notes |
|--------|------|---------|-------|
| `id` | INTEGER PK | AUTOINCREMENT | Recreated with NOT NULL in v26 |
| `session_id` | TEXT NOT NULL | | |
| `action_type` | TEXT NOT NULL | | create, update, delete, start, etc. |
| `entity_type` | TEXT NOT NULL | | issue, log, handoff, etc. |
| `entity_id` | TEXT NOT NULL | | |
| `previous_data` | TEXT | `''` | JSON snapshot of previous state |
| `new_data` | TEXT | `''` | JSON snapshot of new state |
| `timestamp` | DATETIME NOT NULL | `CURRENT_TIMESTAMP` | |
| `undone` | INTEGER | `0` | Undo marker |
| `synced_at` | DATETIME | | Added in v16 |
| `server_seq` | INTEGER | | Added in v16 |

**Indexes**: `session_id`, `timestamp`, `(entity_id, action_type)`

#### `issue_session_history` (v7)

| Column | Type | Default |
|--------|------|---------|
| `id` | TEXT PK | |
| `issue_id` | TEXT NOT NULL | FK to `issues(id)` |
| `session_id` | TEXT NOT NULL | |
| `action` | TEXT NOT NULL | |
| `created_at` | DATETIME | `CURRENT_TIMESTAMP` |

**Indexes**: `issue_id`, `session_id`

#### `boards` (v9, modified in v10, v11, v23)

| Column | Type | Default | Notes |
|--------|------|---------|-------|
| `id` | TEXT PK | | |
| `name` | TEXT NOT NULL | | COLLATE NOCASE, was UNIQUE until v23 |
| `last_viewed_at` | DATETIME | | |
| `created_at` | DATETIME NOT NULL | `CURRENT_TIMESTAMP` | |
| `updated_at` | DATETIME NOT NULL | `CURRENT_TIMESTAMP` | |
| `query` | TEXT NOT NULL | `''` | TDQ query (v10) |
| `is_builtin` | INTEGER NOT NULL | `0` | Built-in flag (v10) |
| `view_mode` | TEXT NOT NULL | `'swimlanes'` | swimlanes/backlog/kanban (v11) |

#### `board_issue_positions` (v9, renamed from `board_issues` in v10)

| Column | Type | Default | Notes |
|--------|------|---------|-------|
| `board_id` | TEXT NOT NULL | | FK to `boards(id)` ON DELETE CASCADE |
| `issue_id` | TEXT NOT NULL | | FK to `issues(id)` ON DELETE CASCADE |
| `position` | INTEGER NOT NULL | | Sparse: re-spaced * 65536 in v22 |
| `added_at` | DATETIME NOT NULL | `CURRENT_TIMESTAMP` | |
| `deleted_at` | | | Soft delete (v25) |

**PK**: `(board_id, issue_id)`

#### `sync_state` (v16)

| Column | Type | Default |
|--------|------|---------|
| `project_id` | TEXT PK | |
| `last_pushed_action_id` | INTEGER | `0` |
| `last_pulled_server_seq` | INTEGER | `0` |
| `last_sync_at` | DATETIME | |
| `sync_disabled` | INTEGER | `0` |

#### `sync_conflicts` (v17)

| Column | Type | Default |
|--------|------|---------|
| `id` | INTEGER PK | AUTOINCREMENT |
| `entity_type` | TEXT NOT NULL | |
| `entity_id` | TEXT NOT NULL | |
| `server_seq` | INTEGER NOT NULL | |
| `local_data` | JSON | |
| `remote_data` | JSON | |
| `overwritten_at` | DATETIME NOT NULL | `CURRENT_TIMESTAMP` |

**Indexes**: `(entity_type, entity_id)`, `overwritten_at`, `server_seq`

#### `sync_history` (v21)

| Column | Type | Default |
|--------|------|---------|
| `id` | INTEGER PK | AUTOINCREMENT |
| `direction` | TEXT NOT NULL | push/pull |
| `action_type` | TEXT NOT NULL | |
| `entity_type` | TEXT NOT NULL | |
| `entity_id` | TEXT NOT NULL | |
| `server_seq` | INTEGER | |
| `device_id` | TEXT | `''` |
| `timestamp` | DATETIME | `CURRENT_TIMESTAMP` |

#### `notes` (v28)

| Column | Type | Default |
|--------|------|---------|
| `id` | TEXT PK | |
| `title` | TEXT NOT NULL | |
| `content` | TEXT NOT NULL | |
| `created_at` | TEXT NOT NULL | |
| `updated_at` | TEXT NOT NULL | |
| `pinned` | INTEGER | `0` |
| `archived` | INTEGER | `0` |
| `deleted_at` | TEXT | |

**Indexes**: `updated_at DESC`, `deleted_at`

#### `agent_activity` (v29)

| Column | Type | Default |
|--------|------|---------|
| `id` | TEXT PK | `aa-XXXXXXXX` hex |
| `issue_id` | TEXT NOT NULL | FK to `issues(id)` |
| `agent_name` | TEXT NOT NULL | |
| `activity_type` | TEXT NOT NULL | assigned/started/committed/pr_created/reviewed/completed/spawned_subagent/comment |
| `details` | TEXT | `''` |
| `session_id` | TEXT | `''` |
| `worktree_path` | TEXT | `''` |
| `branch` | TEXT | `''` |
| `created_at` | DATETIME NOT NULL | `CURRENT_TIMESTAMP` |

**Indexes**: `issue_id`, `agent_name`, `activity_type`, `created_at`

### Migration History Summary

| Version | Type | Description |
|---------|------|-------------|
| 1 | Base | Initial schema (13 tables) |
| 2 | SQL | Add `action_log` table |
| 3 | SQL | Recreate `logs` allowing empty `issue_id` |
| 4 | SQL | Add `minor` column to issues |
| 5 | SQL | Add `created_branch` to issues |
| 6 | SQL | Add `creator_session` to issues |
| 7 | SQL | Add `issue_session_history` table |
| 8 | SQL | Add timestamp indexes |
| 9 | SQL | Add `boards` and `board_issues` tables |
| 10 | SQL | Query-based boards, rename to `board_issue_positions`, add `sprint` |
| 11 | SQL | Add `view_mode` to boards |
| 12 | SQL | Add missing performance indexes |
| 13 | Go | Extend sessions table for DB-backed storage |
| 14 | Go | Repair sessions table (fixup for v13) |
| 15 | Go | Migrate integer PKs to text IDs for sync |
| 16 | SQL | Add `sync_state` table, sync columns to `action_log` |
| 17 | SQL | Add `sync_conflicts` table |
| 18 | Go | Add deterministic IDs to composite-key tables |
| 19 | Go | Convert absolute file paths to repo-relative |
| 20 | Go | Normalize legacy `action_log` composite IDs |
| 21 | SQL | Add `sync_history` table |
| 22 | SQL | Sparse positioning (drop unique index, re-space) |
| 23 | SQL | Drop UNIQUE(name) on boards |
| 24 | Go | Add deterministic IDs to `work_session_issues` |
| 25 | Go | Add `deleted_at` to `board_issue_positions` |
| 26 | Go | Enforce NOT NULL on `action_log.id` |
| 27 | SQL | Normalize NULL session fields in issues |
| 28 | SQL | Add `notes` table |
| 29 | SQL | Add agent swarm tracking columns and `agent_activity` table |

---

## 12.6 Sync Server API Endpoint Reference

**Source**: `/home/devuser/td/internal/api/server.go:161-217` (`routes()` function)

### Middleware Chain

All requests pass through (in order): `recovery`, `requestID`, `logger`, `metrics`, `logging`, `maxBytes(10MB)`, `authRateLimit`

### Health and Metrics

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/healthz` | None | Health check (pings server DB) |
| GET | `/metricz` | None | Server metrics snapshot |

### Authentication (Public)

| Method | Path | Auth | Rate Limit | Description |
|--------|------|------|------------|-------------|
| POST | `/v1/auth/login/start` | None | Auth limit | Start login flow (sends email) |
| POST | `/v1/auth/login/poll` | None | Auth limit | Poll for login completion |
| GET | `/auth/verify` | None | | Render email verification page |
| POST | `/auth/verify` | None | | Submit email verification |

### Projects

| Method | Path | Auth | Min Role | Rate Limit | Description |
|--------|------|------|----------|------------|-------------|
| POST | `/v1/projects` | API Key | (any) | Other | Create a project |
| GET | `/v1/projects` | API Key | (any) | Other | List projects |
| GET | `/v1/projects/{id}` | Project Auth | Reader | Other | Get project details |
| PATCH | `/v1/projects/{id}` | Project Auth | Writer | Other | Update project |
| DELETE | `/v1/projects/{id}` | Project Auth | Owner | Other | Delete project |

### Members

| Method | Path | Auth | Min Role | Rate Limit | Description |
|--------|------|------|----------|------------|-------------|
| POST | `/v1/projects/{id}/members` | Project Auth | Owner | Other | Add member |
| GET | `/v1/projects/{id}/members` | Project Auth | Reader | Other | List members |
| PATCH | `/v1/projects/{id}/members/{userID}` | Project Auth | Owner | Other | Update member role |
| DELETE | `/v1/projects/{id}/members/{userID}` | Project Auth | Owner | Other | Remove member |

### Sync

| Method | Path | Auth | Min Role | Rate Limit | Description |
|--------|------|------|----------|------------|-------------|
| POST | `/v1/projects/{id}/sync/push` | Project Auth | Writer | Push (60/min) | Push local events to server |
| GET | `/v1/projects/{id}/sync/pull` | Project Auth | Reader | Pull (120/min) | Pull events from server |
| GET | `/v1/projects/{id}/sync/status` | Project Auth | Reader | Other | Get sync status |
| GET | `/v1/projects/{id}/sync/snapshot` | Project Auth | Reader | Other | Get full snapshot |

### Admin API

All admin endpoints require authentication with scoped API keys. CORS is enabled for admin endpoints only, controlled by `SYNC_CORS_ALLOWED_ORIGINS`.

**Server Endpoints** (scope: `admin:read:server`):

| Method | Path | Description |
|--------|------|-------------|
| GET | `/v1/admin/server/overview` | Server overview |
| GET | `/v1/admin/server/config` | Running configuration |
| GET | `/v1/admin/server/rate-limit-violations` | Rate limit violation history |
| GET | `/v1/admin/users` | List all users |
| GET | `/v1/admin/users/{id}` | Get user details |
| GET | `/v1/admin/users/{id}/keys` | List user's API keys |
| GET | `/v1/admin/auth/events` | Auth event log |

**Project Endpoints** (scope: `admin:read:projects`):

| Method | Path | Description |
|--------|------|-------------|
| GET | `/v1/admin/projects` | List all projects |
| GET | `/v1/admin/projects/{id}` | Get project details |
| GET | `/v1/admin/projects/{id}/members` | Project members |
| GET | `/v1/admin/projects/{id}/sync/status` | Per-project sync status |
| GET | `/v1/admin/projects/{id}/sync/cursors` | Sync cursor positions |

**Event Endpoints** (scope: `admin:read:events`):

| Method | Path | Description |
|--------|------|-------------|
| GET | `/v1/admin/projects/{id}/events` | Event log browser |
| GET | `/v1/admin/projects/{id}/events/{server_seq}` | Specific event by sequence |
| GET | `/v1/admin/entity-types` | List known entity types |

**Snapshot Endpoints** (scope: `admin:read:snapshots`):

| Method | Path | Description |
|--------|------|-------------|
| GET | `/v1/admin/projects/{id}/snapshot/meta` | Snapshot metadata |
| GET | `/v1/admin/projects/{id}/snapshot/query` | Query snapshot data |

### Admin CLI Commands

**Source**: `/home/devuser/td/cmd/td-sync/admin.go`

```bash
td-sync admin grant --email user@example.com        # Grant admin role
td-sync admin revoke --email user@example.com       # Revoke admin (prevents removing last admin)
td-sync admin create-key --email admin@example.com --name td-watch --scopes "admin:read:server,sync"
```

### Admin API Scopes

| Scope | Description |
|-------|-------------|
| `admin:read:server` | Server overview, config, users, auth events, rate limits |
| `admin:read:projects` | Project listing, members, sync status/cursors |
| `admin:read:events` | Event log browsing, entity types |
| `admin:read:snapshots` | Snapshot metadata and queries |
| `sync` | Standard sync operations |

---

## 12.7 Command Constants Reference

**Source**: `/home/devuser/td/pkg/monitor/keymap/registry.go:44-155`

All 55 named command constants used by the keymap system:

| Command ID | Category | Description |
|-----------|----------|-------------|
| `quit` | Global | Exit the monitor |
| `toggle-help` | Global | Toggle help overlay |
| `refresh` | Global | Refresh data |
| `next-panel` | Navigation | Next panel |
| `prev-panel` | Navigation | Previous panel |
| `cursor-down` | Navigation | Move cursor down |
| `cursor-up` | Navigation | Move cursor up |
| `cursor-top` | Navigation | Go to top |
| `cursor-bottom` | Navigation | Go to bottom |
| `half-page-down` | Navigation | Half page down |
| `half-page-up` | Navigation | Half page up |
| `full-page-down` | Navigation | Full page down |
| `full-page-up` | Navigation | Full page up |
| `scroll-down` | Navigation | Scroll down (in modal) |
| `scroll-up` | Navigation | Scroll up (in modal) |
| `select` | Navigation | Select focused item |
| `back` | Navigation | Go back |
| `close` | Navigation | Close current view |
| `navigate-prev` | Navigation | Navigate to previous item |
| `navigate-next` | Navigation | Navigate to next item |
| `open-details` | Action | Open issue details modal |
| `open-stats` | Action | Open statistics modal |
| `search` | Action | Enter search mode |
| `toggle-closed` | Action | Toggle closed issues |
| `mark-for-review` | Action | Submit issue for review |
| `approve` | Action | Approve issue |
| `delete` | Action | Delete issue |
| `confirm` | Action | Confirm dialog action |
| `cancel` | Action | Cancel dialog action |
| `cycle-sort-mode` | Action | Cycle sort mode |
| `search-confirm` | Search | Apply search query |
| `search-cancel` | Search | Cancel search |
| `search-clear` | Search | Clear search input |
| `search-backspace` | Search | Backspace in search |
| `search-input` | Search | Text input in search |
| `focus-task-section` | Epic | Toggle task section focus |
| `open-epic-task` | Epic | Open task within epic |
| `open-parent-epic` | Epic | Open parent epic |
| `open-blocked-by-issue` | Dependency | Open blocked-by issue |
| `open-blocks-issue` | Dependency | Open blocking issue |
| `open-handoffs` | Handoff | Open handoffs modal |
| `copy-to-clipboard` | Clipboard | Copy as markdown |
| `copy-id-to-clipboard` | Clipboard | Copy issue ID |
| `new-issue` | Form | Create new issue |
| `edit-issue` | Form | Edit current issue |
| `form-submit` | Form | Submit form |
| `form-cancel` | Form | Cancel form |
| `form-toggle-extend` | Form | Toggle extended fields |
| `form-open-editor` | Form | Open external editor |
| `close-issue` | Issue | Close issue |
| `reopen-issue` | Issue | Reopen issue |
| `cycle-type-filter` | Filter | Cycle type filter |
| `next-button` | Button | Next button in dialog |
| `prev-button` | Button | Previous button |
| `boards` | Board | Open board picker |
| `select-board` | Board | Select a board |
| `close-picker` | Board | Close board picker |
| `move-up` | Board | Move issue up |
| `move-down` | Board | Move issue down |
| `move-to-top` | Board | Move issue to top |
| `move-to-bottom` | Board | Move issue to bottom |
| `exit` | Board | Exit board mode |
| `closed` | Board | Toggle closed in board |
| `status-filter` | Board | Cycle status filter |
| `view` | Board | Toggle view mode |
| `kanban-prev-column` | Kanban | Previous column |
| `kanban-next-column` | Kanban | Next column |
| `send-to-worktree` | External | Send to worktree |
| `edit-board` | Board Editor | Edit board |
| `new-board` | Board Editor | Create new board |
| `board-editor-save` | Board Editor | Save board edit |
| `board-editor-cancel` | Board Editor | Cancel board edit |
| `board-editor-delete` | Board Editor | Delete board |
| `open-getting-started` | Getting Started | Open guide |
| `install-instructions` | Getting Started | Install agent instructions |

---

## 12.8 Issue Lifecycle State Machine

**Source**: `/home/devuser/td/cmd/workflow.go`, verified against `internal/models/models.go`

```
                    +---------+
                    |  open   |
                    +----+----+
                         |
              td start   |
                         v
                  +------+------+
                  | in_progress |<-------+
                  +------+------+        |
                         |               |
              td review  |     td reject |
                         v               |
                  +------+------+        |
                  |  in_review  |--------+
                  +------+------+
                         |
             td approve  |
                         v
                    +----+----+
                    |  closed |
                    +---------+

  Any state --td block--> blocked
  blocked --td unblock--> open
  closed --td reopen--> open
  Any state --td close (with --self-close-exception)--> closed
```

Valid status values: `open`, `in_progress`, `blocked`, `in_review`, `closed`

Valid types: `bug`, `feature`, `task`, `epic`, `chore`

Valid priorities: `P0`, `P1`, `P2`, `P3`, `P4` (ordinal, P0 = highest)
