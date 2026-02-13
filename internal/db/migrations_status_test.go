package db

import (
	"testing"
)

func TestMigrationV30StatusMapping(t *testing.T) {
	dir := t.TempDir()
	database, err := Initialize(dir)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	defer database.Close()

	// Roll back schema version to 29 so we can test the migration
	if err := database.setSchemaVersionInternal(29); err != nil {
		t.Fatalf("set schema version: %v", err)
	}

	// Insert issues with old-style statuses directly into DB
	oldStatuses := []struct {
		id     string
		status string
	}{
		{"td-open1", "open"},
		{"td-prog1", "in_progress"},
		{"td-rev1", "in_review"},
		{"td-closed1", "closed"},
		{"td-blocked1", "blocked"},
		// Also verify new statuses are NOT affected
		{"td-backlog1", "backlog"},
		{"td-inflight1", "in_flight"},
		{"td-review1", "review"},
		{"td-shipped1", "shipped"},
		{"td-triage1", "triage"},
	}

	for _, tc := range oldStatuses {
		_, err := database.conn.Exec(
			`INSERT INTO issues (id, title, status, type, priority) VALUES (?, ?, ?, 'task', 'P2')`,
			tc.id, "Test "+tc.id, tc.status,
		)
		if err != nil {
			t.Fatalf("insert issue %s: %v", tc.id, err)
		}
	}

	// Run migrations (should apply v30)
	n, err := database.RunMigrations()
	if err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected 1 migration, got %d", n)
	}

	// Verify schema version is now 30
	ver, err := database.GetSchemaVersion()
	if err != nil {
		t.Fatalf("GetSchemaVersion: %v", err)
	}
	if ver != 30 {
		t.Fatalf("expected schema version 30, got %d", ver)
	}

	// Verify old statuses were mapped correctly
	expectedMappings := map[string]string{
		"td-open1":    "backlog",
		"td-prog1":    "in_flight",
		"td-rev1":     "review",
		"td-closed1":  "shipped",
		"td-blocked1": "backlog",
		// New statuses should remain unchanged
		"td-backlog1":  "backlog",
		"td-inflight1": "in_flight",
		"td-review1":   "review",
		"td-shipped1":  "shipped",
		"td-triage1":   "triage",
	}

	for id, expected := range expectedMappings {
		var actual string
		err := database.conn.QueryRow(`SELECT status FROM issues WHERE id = ?`, id).Scan(&actual)
		if err != nil {
			t.Fatalf("query issue %s: %v", id, err)
		}
		if actual != expected {
			t.Errorf("issue %s: got status %q, want %q", id, actual, expected)
		}
	}
}

func TestMigrationV30Idempotent(t *testing.T) {
	dir := t.TempDir()
	database, err := Initialize(dir)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	defer database.Close()

	// Insert an issue with a new-style status
	_, err = database.conn.Exec(
		`INSERT INTO issues (id, title, status, type, priority) VALUES ('td-test1', 'Test', 'backlog', 'task', 'P2')`,
	)
	if err != nil {
		t.Fatalf("insert issue: %v", err)
	}

	// Roll back and re-run - should not fail even if no old statuses exist
	if err := database.setSchemaVersionInternal(29); err != nil {
		t.Fatalf("set schema version: %v", err)
	}

	n, err := database.RunMigrations()
	if err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected 1 migration, got %d", n)
	}

	// Status should remain backlog
	var status string
	err = database.conn.QueryRow(`SELECT status FROM issues WHERE id = 'td-test1'`).Scan(&status)
	if err != nil {
		t.Fatalf("query issue: %v", err)
	}
	if status != "backlog" {
		t.Errorf("got status %q, want 'backlog'", status)
	}
}
