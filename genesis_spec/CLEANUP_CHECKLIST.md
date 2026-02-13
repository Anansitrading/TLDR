# Cleanup Checklist

## Priority 1: Dead Code Removal

- [ ] **Remove `internal/crypto/` package** (335 lines) — Orphaned package, never imported. Contains `crypto.go` and `crypto_test.go` with encryption utilities that are not used anywhere in the codebase.
- [ ] **Remove 19 dead exported functions** — Functions exported but never called outside their defining file. Run `go vet` and check unused exports.
- [ ] **Clean up `cmd/hello.go`** — Already deleted in latest commit, verify no references remain.

## Priority 2: Code Quality

- [ ] **Decompose `pkg/monitor/model.go`** — File is extremely large and contains too many responsibilities. Extract panel-specific logic, modal handling, and state management into separate files.
- [ ] **Decompose `pkg/monitor/commands.go`** — Command dispatch switch statement is very long. Consider command registry pattern or separate handler files.
- [ ] **Extract `pkg/monitor/view.go` renderers** — Each major view section (task list, detail, stats, board) could be its own file.
- [ ] **Review feature flags** — Check `cmd/feature_gate.go` and `internal/features/` for incomplete features hidden behind gates.

## Priority 3: Security Improvements

- [ ] **Review API key storage** — Currently stored as plaintext in sync config. Consider using OS keychain integration.
- [ ] **Add rate limiting to auth endpoints** — `POST /v1/auth/login/start` and `/poll` could be targets for brute force.
- [ ] **Validate TDQ input length** — No maximum query length enforcement in `internal/query/parser.go`.
- [ ] **Review CORS origins** — Ensure production CORS configuration restricts to expected origins.

## Priority 4: Testing Gaps

- [ ] **Work session tests** — `cmd/ws.go` and related `work_session` DB operations need more comprehensive testing.
- [ ] **Kanban view tests** — New `pkg/monitor/kanban.go` needs unit tests for column calculation, cursor movement, scroll behavior.
- [ ] **Sync conflict resolution tests** — Edge cases in `internal/sync/events.go` field-level merge logic.
- [ ] **TDQ cross-entity query tests** — `dep.` prefix is defined but non-functional per NotebookLM analysis.

## Priority 5: Documentation

- [ ] **Add README.md** — Repository needs a proper README with installation instructions, screenshots, and feature overview.
- [ ] **Sync server API documentation** — Create OpenAPI/Swagger spec for `internal/api/` endpoints.
- [ ] **Update CLAUDE.md** — Reflect kanban view addition, new commands (activity, assign, note).
- [ ] **Docusaurus site** — Verify `website/` content is up-to-date with latest features.

## Priority 6: Technical Debt

- [ ] **Dual SQLite driver maintenance** — Ensure both `mattn/go-sqlite3` (CGo) and `modernc.org/sqlite` (pure-Go) are tested in CI.
- [ ] **Schema migration complexity** — 29 migrations accumulated; consider squashing into a new baseline.
- [ ] **Review `internal/workflow/` guards** — Default mode is `ModeLiberal` which skips guards; ensure this is intentional for production.

## Validation Notes

All items in this checklist were identified through:
- 6-agent forensic code review (architecture, quality, security, orphans, TUI readiness, JTBD)
- NotebookLM codebase interrogation (state machine analysis, command inventory, schema verification)
- Cross-referencing generated documentation against actual code paths
