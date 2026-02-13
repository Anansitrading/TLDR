# Section 11: Documentation & Knowledge

## Overview

td's documentation lives across several layers: in-code comments, a `docs/` directory with design specs and guides, a `CLAUDE.md` file for AI agent instructions, a Docusaurus website scaffold, and generated help text within the TUI. The documentation is developer-oriented and weighted toward internal specs rather than end-user guides.

---

## 11.1 Documentation Inventory

### Root-Level Files

| File | Purpose |
|------|---------|
| `CLAUDE.md` | AI agent instructions for Claude Code (mandatory td usage protocol) |
| `go.mod` | Go module definition (github.com/marcus/td, Go 1.25.5) |

### docs/ Directory

The `docs/` directory contains **41 markdown files** organized into three categories:

#### Active Guides (`docs/guides/`)

| File | Content |
|------|---------|
| `README.md` | Guide index |
| `cli-commands-guide.md` | CLI command reference |
| `collaboration.md` | Multi-user sync collaboration |
| `declarative-modal-guide.md` | How to build modals with the declarative system |
| `lipgloss-table-guide.md` | Activity table styling patterns |
| `monitor-shortcuts-guide.md` | TUI keyboard shortcuts |
| `query-guide.md` | TDQ query language guide |
| `releasing-new-version.md` | Version release process |
| `shortcuts-implementation-guide.md` | How to add new keyboard shortcuts |
| `sync-setup-guide.md` | Sync configuration walkthrough |
| `deprecated/modal-system-guide.md` | Old imperative modal system (superseded) |

#### Implemented Specs (`docs/implemented/`)

These are design documents for features that have been built:

| File | Feature |
|------|---------|
| `SPEC.md` | Original td specification |
| `auto-update-plan.md` | Version check plan (partially implemented) |
| `monitor-autosync-investigation.md` | Auto-sync reliability investigation |
| `proposal-session-identity.md` | Session scoping design |
| `spec-agent-review-bypass-prevention.md` | Self-review prevention design |
| `spec-board-swimlanes-toggle.md` | Board view mode switching |
| `spec-improve-search.md` | Search improvements |
| `spec-issue-boards-v2.md` | Query-based boards design |
| `spec-stacking-modals.md` | Modal stack implementation |
| `spec-td-state-machine.md` | Workflow state machine design |
| `sync-mvp-testing-spec.md` | Sync testing plan |
| `sync-plan-03-merged.md` | Merged sync implementation plan |

#### Operational Docs

| File | Content |
|------|---------|
| `sync-server-ops-guide.md` | Server operations guide |
| `sync-client-guide.md` | Client-side sync setup |
| `sync-agent-guide.md` | Agent sync workflow |
| `sync-testing-guide.md` | Sync test procedures |
| `sync-test-coverage.md` | Test coverage analysis |
| `sync-admin-api-plan.md` | Admin API design |
| `sync-dev-notes.md` | Developer notes for sync |
| `sync-mainline-merge-plan.md` | Branch merge strategy |
| `notes-sidecar-gated-merge-plan.md` | Notes feature gate plan |

#### Technical References

| File | Content |
|------|---------|
| `db-layer-mutation.md` | Logged vs unlogged DB operations |
| `form-wrapping-investigation.md` | Form field wrapping bug analysis |
| `modal-inventory.md` | Complete modal enumeration |
| `multi-agent-ui-review.md` | Multi-agent UI patterns |
| `pro-features-plan.md` | Future feature roadmap |
| `perf/2026-01-17-memory-leak-analysis.md` | Memory leak investigation |

#### Deprecated

| File | Content |
|------|---------|
| `deprecated/spec-turso-libsql-support.md` | Abandoned libSQL plan |
| `deprecated/spec-remote-sync-options.md` | Early sync options analysis |
| `deprecated/sync-plan-01-codex.md` | First sync plan (superseded) |
| `deprecated/sync-plan-02-opus.md` | Second sync plan (superseded) |

---

## 11.2 CLAUDE.md as AI Agent Instructions

**Source**: `/home/devuser/td/CLAUDE.md`

The `CLAUDE.md` file serves a dual purpose:
1. **Agent protocol**: Mandatory instructions for AI coding agents using td
2. **Developer quickref**: Build, test, and release instructions

### Key Agent Instructions

- **Mandatory**: Run `td usage --new-session` at conversation start
- **Session warning**: "Do NOT start a new session mid-work" (new session bypasses review)
- **Quick mode**: Use `td usage -q` after first read (compact output)

### Developer Reference

- Build: `go build -o td .`
- Test: `go test ./...`
- Release process with version tags
- Architecture overview (package layout)
- Settings persistence documentation
- Known issue: `saveFilterState()` in embedded sidecar mode

### Agent File Detection and Installation

The `td init` command and monitor's Getting Started modal detect existing agent files:

**Source**: `/home/devuser/td/internal/agent/`

The agent detection system searches for:
- `AGENTS.md`
- `CLAUDE.md`
- `.cursorrules`
- Other AI agent instruction files

When found, td offers to inject its usage instructions. The Getting Started modal (`pkg/monitor/getting_started.go`) shows an "Install" button if instructions are not yet present.

---

## 11.3 Docusaurus Website

**Location**: `/home/devuser/td/website/`

A Docusaurus website scaffold exists with:
- `docusaurus.config.js` - Configuration
- `package.json` / `package-lock.json` - Dependencies
- `sidebars.js` - Navigation structure
- `src/` - React components
- `static/` - Static assets
- `docs/` - Website documentation pages
- `README.md` - Website setup instructions

This appears to be a **scaffold** for a documentation website, separate from the `docs/` directory in the project root. The relationship between root `docs/` (developer specs) and `website/docs/` (user-facing docs) is not explicitly documented.

---

## 11.4 In-Code Documentation Quality

### Package Comments

Most packages have descriptive package-level comments:
- `cmd/root.go`: `// Package cmd implements all td CLI commands using cobra.`
- `internal/query/parser.go`: `// Package query implements the TDQ (td query) language parser, lexer, AST, and evaluator for filtering issues.`
- `internal/config/config.go`: `// Package config handles loading and saving td configuration from .todos/config.json.`
- `internal/version/version.go`: `// Package version provides update checking against GitHub releases and semantic version comparison.`

### Function Comments

Exported functions consistently have godoc comments. Examples:
- `FetchData()` in `data.go`: "FetchData retrieves all data needed for the monitor display"
- `getSharedDB()` in `dbpool.go`: Detailed comment explaining the singleton pattern and why it exists
- `ComputeBoardIssueCategories()`: "Sets the Category field on each BoardIssueView. This is the single source of truth..."

### Structural Comments

The codebase uses section-level comments to organize long files:
- Navigation bindings, Modal bindings, CRUD, Actions sections in `keymap/help.go`
- "Logged variants" vs "Unlogged variants" comments in the DB layer

### Areas of Weak Documentation

- The `Model` struct in `pkg/monitor/model.go` has ~200 fields with minimal per-field documentation
- Legacy fields are marked `// legacy, kept for compatibility` but without migration guidance
- The `executeCommand()` 892-line switch statement has no section-level comments
- The `Logged` suffix convention (meaning "with action_log entry for undo/sync") is not documented in the package

---

## 11.5 Help System

### TUI Help Modal

**Source**: `/home/devuser/td/pkg/monitor/keymap/help.go`

The `?` key opens a scrollable help overlay (`GenerateHelp()`) organized into 13 sections:

1. **Navigation** - Tab, j/k, Ctrl+d/u, G/gg, Enter
2. **Modals** - Scroll, navigate, refresh, copy, Tab for epic focus
3. **Epic Tasks** - j/k selection, Enter to open, Tab to exit
4. **CRUD** - n (new), e (edit), x (delete), C (close), O (reopen)
5. **Confirmation Dialogs** - Tab between buttons, Y/N shortcuts
6. **Form** - Ctrl+S (save), Esc (cancel), Ctrl+X (extended), Ctrl+O (editor)
7. **Actions** - r (review), a (approve), s (stats), S (sort), T (type), / (search), c (closed)
8. **Getting Started** - H (guide), I (install)
9. **Handoffs Modal** - j/k, Enter, r (refresh)
10. **Boards** - b (picker), J/K (move), v (view cycle), F (filter)
11. **Kanban View** - h/l (columns), j/k (cards), G/gg (top/bottom)
12. **Search** - Enter, Esc, Backspace, ? (TDQ help)
13. **Mouse** - Click, Double-click, Scroll wheel

### TDQ Query Help Modal

**Source**: `/home/devuser/td/pkg/monitor/keymap/help.go:205-309`

Pressing `?` while in search mode opens TDQ-specific help with sections:

1. **Basic Operators** - =, !=, ~, !~, <, >, <=, >=
2. **Boolean Logic** - AND, OR, NOT, parentheses
3. **Fields** - status, type, priority, points, labels, title, description, created, updated, implementer, reviewer
4. **Functions** - has(), is(), any(), descendant_of()
5. **Cross-Entity** - log.message, log.type, comment.text, file.role
6. **Special Values** - @me, today, -7d, EMPTY
7. **Sorting** - sort:priority, sort:-created, sort:-updated, sort:created
8. **Examples** - Five concrete query examples

### Footer Help Strings

Four context-specific footer help strings provide condensed shortcut references:

| Context | Content |
|---------|---------|
| Main panel | `n:new e:edit x:del a:approve r:review  S:sort T:type c:closed b:boards  /:search s:stats tab:panel ?:help` |
| Board mode | Adds `v:view F:filter` |
| Kanban | `h/l:columns j/k:cards  v:view n:new e:edit  F:filter c:closed b:boards  /:search s:stats ?:help` |
| Modal | `updown:scroll  leftright:prev/next  y:copy  esc:close  r:refresh` |
| Stats | `updown:scroll  Ctrl+d/u:halfpage  esc:close  r:refresh` |

### CommandHelp()

The `CommandHelp()` function (`help.go:338-439`) provides per-command descriptions for ~40 TUI commands (e.g., `CmdQuit` -> "Exit the monitor", `CmdCopyToClipboard` -> "Copy issue as markdown to clipboard").

### CLI Help

Cobra provides built-in `--help` for all commands. The root command uses a custom usage template (`cmd/root.go:267-296`) that shows aliases inline:

```
Core Commands:
  create, add, new          Create a new issue
  list, ls                  List issues
  show, context, view, get  Show issue details
```

Commands are organized into 7 groups: Core, Workflow, Query, Shortcuts, Session, Files, System.

---

## 11.6 Getting Started Experience

### `td init`

**Source**: `/home/devuser/td/cmd/init.go`

The initialization flow:
1. Creates `.todos/` directory with SQLite database
2. Adds `.todos/` to `.gitignore`
3. Creates initial session
4. Detects agent files (CLAUDE.md, AGENTS.md, .cursorrules)
5. If found: asks if user wants td instructions injected
6. If not found: prints suggested text for manual addition

### Monitor Getting Started Modal

**Source**: `/home/devuser/td/pkg/monitor/getting_started.go`

On first launch, the monitor shows a Getting Started modal:

```
Welcome to td!

Task management for AI agents.

[checkmark] Agent instructions installed
  -- or --
Press I to install td instructions to AGENTS.md

PROMPT: "Use td to plan my feature and implement it."

Press ? for help . H to reopen this modal

[Install] [Close]
```

The modal is built declaratively using `modal.New()` with text sections, spacers, and buttons. It is designed to fit on 80x24 terminals.

Key features:
- Detects if agent file already has td instructions (shows checkmark if yes)
- `I` key installs instructions directly
- `H` key reopens the modal at any time
- Chains to sync prompt if authenticated (feature-gated behind `sync_monitor_prompt`)

---

## 11.7 Documentation Gaps

### Missing Documentation

| Gap | Description | Impact |
|-----|-------------|--------|
| **No end-user quickstart** | No single "getting started" guide for humans (only CLAUDE.md for agents) | New users must discover features by exploring `td --help` |
| **No API documentation** | Sync server API endpoints not documented outside code | Server operators must read `internal/api/server.go` |
| **No schema documentation** | 20 tables across 29 migrations with no ER diagram | Contributors must trace through `schema.go` and `migrations.go` |
| **No architecture decision records** | Design decisions are scattered across implemented specs | Rationale for choices is hard to reconstruct |
| **No changelog** | No CHANGELOG.md or release notes | Users cannot discover what changed between versions |
| **No contribution guide** | No CONTRIBUTING.md | External contributors lack onboarding path |
| **Stale sync docs** | Multiple sync plans (01, 02, 03) exist; unclear which is current | Confusing for anyone studying the sync architecture |
| **Duplicate CLAUDE.md sections** | CLAUDE.md has the "MANDATORY" section duplicated verbatim | Copy-paste error, minor but noticeable |
| **Website incomplete** | Docusaurus scaffold exists but content relationship to `docs/` is unclear | Two documentation trees without clear boundary |
| **No keymap.json documentation** | User-configurable key bindings exist but no guide or template generator | Users must discover the feature via source code |

### Documentation that Diverges from Code

| Claim | Reality |
|-------|---------|
| `auto-update-plan.md` describes `internal/update/update.go` | Package does not exist; actual implementation is `internal/version/` |
| `auto-update-plan.md` describes an `update` subcommand | No such command exists |
| `modal-system-guide.md` describes the old imperative modal system | Partially superseded by `declarative-modal-guide.md` |
| CLAUDE.md references `docs/modal-system.md` | File does not exist at that path (it is at `docs/guides/deprecated/modal-system-guide.md`) |
