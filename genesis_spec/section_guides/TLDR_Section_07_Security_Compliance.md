# Section 07: Security & Compliance

## 7.1 Overview

**Overall Risk Level:** Medium

The `td` codebase demonstrates generally strong security practices. The core database operations use parameterized queries consistently, API keys are SHA-256 hashed before storage, authentication follows an OAuth 2.0-style device flow, and rate limiting is applied per-IP on auth endpoints and per-key on sync operations. No critical vulnerabilities were identified. The primary concerns involve defense-in-depth gaps in the sync engine's dynamic SQL construction, missing LIKE wildcard escaping in one code path, path traversal risks in the project DB pool, and the sync server defaulting to plaintext HTTP.

**Audit scope:** 332 Go files, ~126K lines. Source: `03_security.json` findings cross-referenced with actual code.

---

## 7.2 Authentication

### Mechanism

The sync server uses a **Device Authorization Flow** (similar to GitHub's device flow):

1. Client sends email to `POST /v1/auth/login/start`
2. Server returns a 6-character `user_code` and `device_code`
3. User opens verification URI in browser and enters `user_code`
4. Client polls `POST /v1/auth/login/poll` with `device_code`
5. On verification, server returns a Bearer API key (`td_live_` prefix + 32 base62 characters)

**Source:** `internal/api/auth.go` (verified), `internal/serverdb/apikeys.go` (verified)

### API Key Security

| Measure | Implementation | Location |
|---------|---------------|----------|
| Keys hashed before storage | SHA-256 hash, plaintext never stored | `serverdb/apikeys.go:67-68` |
| Credential file permissions | `auth.json` written with `0600` (owner-only) | `syncconfig/syncconfig.go:125` |
| Key prefix for identification | `td_live_` prefix enables key scanning/detection | `api/auth.go` |
| User code character set | Excludes ambiguous chars: `ABCDEFGHJKMNPQRSTUVWXYZ23456789` (no I, L, O, 0, 1) | `api/auth.go:282` |
| Auth event logging | Both successful and failed attempts logged for audit trail | `api/auth.go:268-278` |
| Auth request reuse prevention | Status tracking prevents code reuse (`already_used` check) | `api/auth.go:133-136` |
| Expired request cleanup | Automatic cleanup of expired auth requests | `server.go:77-95` |

### Authorization Model

- **Admin scope-based access**: Admin endpoints require both `is_admin` flag and specific scopes (`admin:read:server`, `admin:read:projects`, `admin:read:events`, `admin:read:snapshots`, `admin:export`)
- **Project-level roles**: `owner` > `writer` > `reader` hierarchy
- **Membership validation**: `requireProjectAuth` middleware checks membership before processing

**Source:** `internal/api/admin_middleware.go:14-17` (verified), `internal/api/middleware.go:189-207` (verified)

---

## 7.3 SQL Injection Prevention

### Parameterized Queries

The core database layer (`internal/db/`) uses parameterized queries with `?` placeholders consistently throughout. All user-provided values (issue IDs, titles, descriptions, status values, session IDs) are passed as parameters, not interpolated into SQL strings.

**Verified in:** `internal/db/issues.go` -- all CRUD operations use `?` placeholders.

### Sort Column Validation

`ListIssues` validates sort columns against an explicit allowlist of 10 permitted columns before interpolation:

**Source:** `internal/db/issues.go:497-501` (verified)

### TDQ Query Language as SQL Injection Defense

The TDQ parser provides a safe abstraction over database queries:
- Users write TDQ expressions (`status=open AND priority=P0`)
- The parser validates field names, operators, and values
- The evaluator generates parameterized SQL or performs in-memory filtering
- Users never write raw SQL

**Source:** `internal/query/parser.go`, `internal/query/evaluator.go` (verified)

### LIKE Wildcard Escaping

The `evaluator.go:29-32` defines an `escapeSQLWildcards` function and uses it correctly for field `contains` operations at line 261.

---

## 7.4 Identified Vulnerabilities

### SEC-001: Dynamic Table/Column Name Interpolation in Sync Engine

| Attribute | Value |
|-----------|-------|
| **Severity** | Medium |
| **CWE** | CWE-89 (SQL Injection) |
| **Location** | `internal/sync/events.go:232,353,412,433,442,451` |

**Description:** The sync engine constructs SQL queries by interpolating `entityType` and column names using `fmt.Sprintf` (e.g., `fmt.Sprintf("UPDATE %s SET %s WHERE id = ?", entityType, ...)`). While `entityType` is validated against an allowlist in the API layer (`api/sync.go:38-40`) and column names are validated with `validColumnName = regexp.MustCompile('^[a-zA-Z_][a-zA-Z0-9_]*$')` at `events.go:15` (verified), the entity type parameter itself is not regex-validated in the functions that interpolate it.

**Exploit scenario:** If a caller bypasses the `isValidEntityType` check (e.g., via a future code path or direct function call), a crafted entity type could inject SQL.

**Recommendation:** Add `validColumnName.MatchString(entityType)` validation at the start of each function that interpolates entity type into SQL.

**Effort:** Small (5-line fix per function)

---

### SEC-002: Missing LIKE Wildcard Escaping in Text Search

| Attribute | Value |
|-----------|-------|
| **Severity** | Medium |
| **CWE** | CWE-89 (SQL Injection -- LIKE pattern injection) |
| **Location** | `internal/query/evaluator.go:363-368` |

**Description:** The `textSearchToSQL` method constructs LIKE patterns as `"%" + node.Text + "%"` without escaping `%` and `_` wildcards in user input. The `escapeSQLWildcards` function exists at `evaluator.go:29-32` but is not applied to text search patterns.

**Exploit scenario:** A user searching for `%` or `_` characters gets wildcard expansion instead of literal matching. Not a data breach vector but violates least-surprise.

**Recommendation:** Apply `escapeSQLWildcards` to `TextSearch.Text` and add `ESCAPE` clause. The function already exists.

**Effort:** Small (2-line fix)

---

### SEC-003: Hardcoded Table List with String Concatenation

| Attribute | Value |
|-----------|-------|
| **Severity** | Medium |
| **CWE** | CWE-89 (SQL Injection) |
| **Location** | `internal/api/admin_snapshots.go:109` |

**Description:** `countSnapshotEntities` constructs SQL via `db.QueryRow("SELECT COUNT(*) FROM " + table)`. Table names come from a hardcoded slice, but the pattern is fragile.

**Recommendation:** Validate table name against `validColumnName` regex before interpolation.

**Effort:** Small

---

### SEC-004: Path Traversal in Project DB Pool

| Attribute | Value |
|-----------|-------|
| **Severity** | Medium |
| **CWE** | CWE-22 (Path Traversal) |
| **Location** | `internal/api/dbpool.go:47,71` |

**Description:** `ProjectDBPool.Get` and `Create` construct file paths using `filepath.Join(p.dataDir, projectID, "events.db")` where `projectID` comes from the HTTP request path. While `requireProjectAuth` middleware validates membership, the authorization and file path resolution are not atomic.

**Exploit scenario:** A request with `projectID = "../../etc"` would attempt to open `/etc/events.db`. The auth middleware would likely reject it, but there is a TOCTOU gap.

**Recommendation:** Reject any `projectID` containing `/`, `\`, or `..` sequences. A UUID format validation would be ideal since project IDs should be UUIDs.

**Effort:** Small (3-line guard clause)

---

### SEC-005: Unauthenticated Metrics Endpoint

| Attribute | Value |
|-----------|-------|
| **Severity** | Medium |
| **CWE** | CWE-200 (Information Disclosure) |
| **Location** | `internal/api/server.go:165-166,229-231` |

**Description:** The `/metricz` endpoint is publicly accessible without authentication, exposing request counts, error rates, push event counts, and pull request counts.

**Recommendation:** Add authentication to `/metricz`. Keep `/healthz` public for load balancer checks.

**Effort:** Small

---

### SEC-006: Default HTTP (No TLS)

| Attribute | Value |
|-----------|-------|
| **Severity** | Medium |
| **CWE** | CWE-319 (Cleartext Transmission) |
| **Location** | `internal/api/config.go:40`, `internal/syncconfig/syncconfig.go:47` |

**Description:** The sync server defaults to `http://localhost:8080` with no TLS configuration. API keys are transmitted as Bearer tokens in plaintext. No certificate verification in the sync client.

**Recommendation:** Add TLS configuration or document reverse proxy requirement. Add startup warning when BaseURL is non-HTTPS for non-localhost deployments.

**Effort:** Medium

---

### SEC-007: Permissive Config Directory Permissions

| Attribute | Value |
|-----------|-------|
| **Severity** | Low |
| **CWE** | CWE-276 (Incorrect Default Permissions) |
| **Location** | `internal/syncconfig/syncconfig.go:56,92` |

**Description:** Config directory (`~/.config/td/`) created with `0755` and `config.json` with `0644`. While `auth.json` correctly uses `0600`, the config file may expose sync server URLs.

**Recommendation:** Use `0700` for directory and `0600` for `config.json`.

---

### SEC-008: Long-Lived API Keys Without Rotation

| Attribute | Value |
|-----------|-------|
| **Severity** | Low |
| **CWE** | CWE-798 (Hard-coded Credentials) |
| **Location** | `internal/api/auth.go:156` |

**Description:** API keys have 1-year expiry with no rotation mechanism, no expiry warnings, and no CLI-based key revocation.

**Recommendation:** Shorter default lifetimes (90 days), refresh mechanism, or at minimum expiry warnings.

---

### SEC-009: X-Forwarded-For Trust Without Proxy Validation

| Attribute | Value |
|-----------|-------|
| **Severity** | Low |
| **CWE** | CWE-346 (Origin Validation Error) |
| **Location** | `internal/api/ratelimit.go:140-153` |

**Description:** The `clientIP` function trusts `X-Forwarded-For` without validating the proxy source. Attackers can spoof IP to bypass IP-based rate limiting on auth endpoints.

**Recommendation:** Add trusted proxy configuration. Ignore `X-Forwarded-For` when not behind a trusted proxy.

---

### SEC-010: Snapshot Bootstrap Trusts Server Database

| Attribute | Value |
|-----------|-------|
| **Severity** | Low |
| **CWE** | CWE-494 (Download Without Integrity Check) |
| **Location** | `cmd/sync.go:202,223` |

**Description:** Snapshot bootstrap downloads a full SQLite database and writes it to disk with only a header validation. A malicious server could craft a database with triggers.

**Recommendation:** Validate schema matches expected td schema. Consider signing snapshots.

---

### SEC-011: SQL Injection in Test Code

| Attribute | Value |
|-----------|-------|
| **Severity** | Low |
| **CWE** | CWE-89 |
| **Location** | `test/syncharness/harness.go`, `test/e2e/verify.go` |

**Description:** Test code uses `fmt.Sprintf` for SQL queries with direct interpolation. Low direct risk since it is test-only, but sets a bad precedent.

**Recommendation:** Refactor test SQL to use parameterized queries.

---

### SEC-012: Manual JSON Escaping

| Attribute | Value |
|-----------|-------|
| **Severity** | Low |
| **CWE** | CWE-117 (Improper Output Neutralization for Logs) |
| **Location** | `cmd/security.go:44`, `cmd/errors.go:113-119` |

**Description:** Manual `escapeJSON` function only handles `\`, `"`, `\n`, `\t`. Does not escape other control characters, potentially producing invalid JSON.

**Recommendation:** Replace with `encoding/json.Marshal`.

---

### SEC-013: Generous TDQ Parser Depth Limit

| Attribute | Value |
|-----------|-------|
| **Severity** | Low |
| **CWE** | CWE-400 (Denial of Service) |
| **Location** | `internal/query/parser.go:15,203` |

**Description:** `MaxQueryDepth=50` is generous. No maximum query length or token count limit.

**Recommendation:** Add query length limit (4096 chars), reduce depth to 20, add token count limit.

---

### SEC-014: Wildcard CORS Configuration Risk

| Attribute | Value |
|-----------|-------|
| **Severity** | Low |
| **CWE** | CWE-942 (Permissive CORS) |
| **Location** | `internal/api/cors.go:24` |

**Description:** Wildcard origin (`*`) is accepted as configuration. Would allow all origins to make cross-origin requests.

**Recommendation:** Log warning for wildcard configuration. Reject in production mode.

---

### SEC-015: Non-Constant-Time API Key Comparison

| Attribute | Value |
|-----------|-------|
| **Severity** | Low |
| **CWE** | CWE-208 (Timing Side-Channel) |
| **Location** | `internal/serverdb/apikeys.go:64-68` |

**Description:** API key verification uses SHA-256 hash lookup via SQL WHERE clause rather than `crypto/subtle.ConstantTimeCompare`. Theoretical timing attack, but infeasible in practice since the DB query time dominates.

**Recommendation:** Very low priority. Could add constant-time comparison after DB retrieval for defense-in-depth.

---

## 7.5 Positive Security Findings

The following security strengths were verified in the source code:

| Finding | Location (verified) |
|---------|-------------------|
| All core DB operations use parameterized queries with `?` | `internal/db/issues.go` |
| API keys SHA-256 hashed before storage | `serverdb/apikeys.go:67` |
| Auth credentials file has `0600` permissions | `syncconfig/syncconfig.go:125` |
| Per-IP rate limiting on auth endpoints | `ratelimit.go:68-73` |
| Request body size limited to 10MB | `middleware.go:226` |
| Panic recovery middleware returns generic 500 | `middleware.go:79-89` |
| Request ID tracking for audit | `middleware.go:101-108` |
| Admin requires both admin flag AND scope | `admin_middleware.go:14-17` |
| Sort columns validated against explicit allowlist | `issues.go:497-501` |
| Column names regex-validated in sync engine | `events.go:15` |
| TDQ parser has depth limiting | `parser.go:15,203` |
| LIKE wildcards escaped in field contains | `evaluator.go:29-32` |
| Entity types validated against allowlist on push | `api/sync.go:140-145` |
| CORS uses explicit origin list by default | `cors.go:11-14` |
| Sensitive analytics keywords redacted | `db/analytics.go:40-42` |
| Auth events logged including failures | `api/auth.go:268-278` |
| Expired auth requests auto-cleaned | `server.go:77-95` |
| HTTP server has read/write/idle timeouts | `server.go:41-43` |
| Signup can be disabled via configuration | `auth.go:69-80` |

---

## 7.6 Input Validation Summary

### Comprehensive Validation Patterns

| Input | Validation | Location |
|-------|-----------|----------|
| SQL query parameters | Parameterized queries (`?`) | `internal/db/*.go` |
| Sort columns | Explicit allowlist (10 columns) | `db/issues.go:497-501` |
| Sync entity types | Allowlist check on push | `api/sync.go:38-40` |
| Config keys | Explicit allowlist (8 keys) | `cmd/config.go:16-25` |
| Email addresses | `net/mail.ParseAddress` | `api/auth.go:64` |
| User codes | 6 chars from restricted set | `api/auth.go:208-209` |
| Pull limit | Capped at 10,000 | `api/sync.go:247-249` |
| Push batch | Capped at 1,000 | `api/sync.go:134` |
| TDQ nesting | Max depth 50 | `query/parser.go:15` |
| Snapshot query limit | Capped at 200 | `api/admin_snapshots.go:132` |
| Snapshot header | SQLite header validated | `cmd/sync.go:202` |
| Column names | Regex: `^[a-zA-Z_][a-zA-Z0-9_]*$` | `sync/events.go:15` |

### Validation Gaps

| Gap | Risk | Location |
|-----|------|----------|
| TDQ text search: no LIKE wildcard escaping | Low (pattern expansion) | `evaluator.go:363-368` |
| No TDQ query length limit | Low (DoS) | `parser.go` |
| `projectID` not validated for path traversal | Medium (file access) | `api/dbpool.go:47` |
| `mapFieldToColumn` returns unknown fields verbatim | Low (lexer limits chars) | `evaluator.go:392-413` |
| No format validation on device_id, session_id | Low | Push requests |

---

## 7.7 Dependency Security

All dependencies are recent versions with no known CVEs at time of audit:

| Package | Version | Risk |
|---------|---------|------|
| `modernc.org/sqlite` | v1.41.0 | Low -- pure Go SQLite |
| `mattn/go-sqlite3` | v1.14.33 | Low -- CGo SQLite binding |
| `golang.org/x/crypto` | v0.47.0 | Low |
| `golang.org/x/net` | v0.48.0 | Low |
| `spf13/cobra` | v1.10.2 | Low |
| `charmbracelet/bubbletea` | v1.3.10 | Low |
| `microcosm-cc/bluemonday` | v1.0.27 | Low -- HTML sanitizer |
| Go | 1.25.5 | Current |

**Note:** Two SQLite drivers are present (`mattn/go-sqlite3` for client, `modernc.org/sqlite` for server). This increases the dependency surface but allows CGo-free server builds.

---

## 7.8 Prioritized Remediation

| Priority | Finding | Action | Effort |
|----------|---------|--------|--------|
| 1 | SEC-001 | Add regex validation for `entityType` in sync event functions | Small |
| 2 | SEC-004 | Add path traversal guard in `ProjectDBPool.Get/Create` | Small |
| 3 | SEC-002 | Apply `escapeSQLWildcards` to text search | Small |
| 4 | SEC-006 | Add TLS support or document reverse proxy requirement | Medium |
| 5 | SEC-005 | Authenticate `/metricz` endpoint | Small |
| 6 | SEC-009 | Add trusted proxy configuration for rate limiting | Medium |
| 7 | SEC-007 | Use `0700`/`0600` for config directory and files | Small |
| 8 | SEC-012 | Replace manual JSON escaping with `encoding/json.Marshal` | Small |
