package api

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestSnapshotQuerySource_GetRejectedInProgressIssueIDs(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	// Create minimal schema needed for the query
	_, err = db.Exec(`
		CREATE TABLE issues (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'backlog',
			deleted_at DATETIME
		);
		CREATE TABLE action_log (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT NOT NULL,
			action_type TEXT NOT NULL,
			entity_type TEXT NOT NULL,
			entity_id TEXT NOT NULL,
			timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			undone INTEGER DEFAULT 0
		);
	`)
	if err != nil {
		t.Fatalf("schema: %v", err)
	}

	// issue1: in_flight + rejected, no subsequent review → should be detected
	_, err = db.Exec(`INSERT INTO issues (id, title, status) VALUES ('i1', 'Issue 1', 'in_flight')`)
	if err != nil {
		t.Fatalf("insert i1: %v", err)
	}
	_, err = db.Exec(`INSERT INTO action_log (session_id, action_type, entity_type, entity_id) VALUES ('ses_rev', 'reject', 'issue', 'i1')`)
	if err != nil {
		t.Fatalf("insert action i1: %v", err)
	}

	// issue2: in_flight + rejected, then re-submitted → should NOT be detected
	_, err = db.Exec(`INSERT INTO issues (id, title, status) VALUES ('i2', 'Issue 2', 'in_flight')`)
	if err != nil {
		t.Fatalf("insert i2: %v", err)
	}
	_, err = db.Exec(`INSERT INTO action_log (session_id, action_type, entity_type, entity_id, timestamp) VALUES ('ses_rev', 'reject', 'issue', 'i2', '2025-01-01 00:00:00')`)
	if err != nil {
		t.Fatalf("insert reject i2: %v", err)
	}
	_, err = db.Exec(`INSERT INTO action_log (session_id, action_type, entity_type, entity_id, timestamp) VALUES ('ses_impl', 'review', 'issue', 'i2', '2025-01-01 00:01:00')`)
	if err != nil {
		t.Fatalf("insert review i2: %v", err)
	}

	// issue3: backlog + rejected → should NOT be detected (wrong status)
	_, err = db.Exec(`INSERT INTO issues (id, title, status) VALUES ('i3', 'Issue 3', 'backlog')`)
	if err != nil {
		t.Fatalf("insert i3: %v", err)
	}
	_, err = db.Exec(`INSERT INTO action_log (session_id, action_type, entity_type, entity_id) VALUES ('ses_rev', 'reject', 'issue', 'i3')`)
	if err != nil {
		t.Fatalf("insert action i3: %v", err)
	}

	// issue4: in_flight, never rejected → should NOT be detected
	_, err = db.Exec(`INSERT INTO issues (id, title, status) VALUES ('i4', 'Issue 4', 'in_flight')`)
	if err != nil {
		t.Fatalf("insert i4: %v", err)
	}

	// issue5: deleted + in_flight + rejected → should NOT be detected
	_, err = db.Exec(`INSERT INTO issues (id, title, status, deleted_at) VALUES ('i5', 'Issue 5', 'in_flight', '2025-01-01')`)
	if err != nil {
		t.Fatalf("insert i5: %v", err)
	}
	_, err = db.Exec(`INSERT INTO action_log (session_id, action_type, entity_type, entity_id) VALUES ('ses_rev', 'reject', 'issue', 'i5')`)
	if err != nil {
		t.Fatalf("insert action i5: %v", err)
	}

	src := NewSnapshotQuerySource(db)
	result, err := src.GetRejectedInProgressIssueIDs()
	if err != nil {
		t.Fatalf("GetRejectedInProgressIssueIDs: %v", err)
	}

	if !result["i1"] {
		t.Error("i1 should be detected: in_flight + rejected without re-review")
	}
	if result["i2"] {
		t.Error("i2 should NOT be detected: was re-submitted after rejection")
	}
	if result["i3"] {
		t.Error("i3 should NOT be detected: not in in_flight status")
	}
	if result["i4"] {
		t.Error("i4 should NOT be detected: never rejected")
	}
	if result["i5"] {
		t.Error("i5 should NOT be detected: deleted issue")
	}
	if len(result) != 1 {
		t.Errorf("expected exactly 1 result, got %d: %v", len(result), result)
	}
}
