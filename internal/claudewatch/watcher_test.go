package claudewatch

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Anansitrading/TLDR/internal/db"
)

func setupTestWatcher(t *testing.T) (*Watcher, string, *db.DB) {
	t.Helper()
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	os.MkdirAll(filepath.Join(claudeDir, "tasks"), 0755)
	os.MkdirAll(filepath.Join(claudeDir, "teams"), 0755)

	dbDir := filepath.Join(tmpDir, "project")
	database, err := db.Initialize(dbDir)
	if err != nil {
		t.Fatalf("failed to init test db: %v", err)
	}

	w := NewWatcher(claudeDir, database)
	return w, claudeDir, database
}

func TestScanTeamTasks(t *testing.T) {
	w, claudeDir, database := setupTestWatcher(t)
	defer database.Close()

	database.CreateProject("TestProj", "", "test-team")

	teamTaskDir := filepath.Join(claudeDir, "tasks", "test-team")
	os.MkdirAll(teamTaskDir, 0755)

	task := ClaudeTask{
		ID:      "1",
		Subject: "Do the thing",
		Status:  "in_progress",
		Owner:   "agent-1",
	}
	data, _ := json.Marshal(task)
	os.WriteFile(filepath.Join(teamTaskDir, "1.json"), data, 0644)

	count, err := w.ScanAndImport()
	if err != nil {
		t.Fatalf("ScanAndImport failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 imported task, got %d", count)
	}

	issues, _ := database.ListIssues(db.ListIssuesOptions{})
	found := false
	for _, issue := range issues {
		if issue.LinearID == "claude:test-team:1" {
			found = true
			if issue.Title != "Do the thing" {
				t.Errorf("expected title 'Do the thing', got '%s'", issue.Title)
			}
			if issue.Assignee != "agent-1" {
				t.Errorf("expected assignee 'agent-1', got '%s'", issue.Assignee)
			}
		}
	}
	if !found {
		t.Error("imported issue not found in database")
	}
}

func TestScanUpdatesExistingTask(t *testing.T) {
	w, claudeDir, database := setupTestWatcher(t)
	defer database.Close()

	database.CreateProject("TestProj", "", "test-team")
	teamTaskDir := filepath.Join(claudeDir, "tasks", "test-team")
	os.MkdirAll(teamTaskDir, 0755)

	task := ClaudeTask{ID: "1", Subject: "Initial", Status: "pending", Owner: "a1"}
	data, _ := json.Marshal(task)
	os.WriteFile(filepath.Join(teamTaskDir, "1.json"), data, 0644)
	w.ScanAndImport()

	task.Subject = "Updated"
	task.Status = "completed"
	data, _ = json.Marshal(task)
	os.WriteFile(filepath.Join(teamTaskDir, "1.json"), data, 0644)

	w.ScanAndImport()

	issues, _ := database.ListIssues(db.ListIssuesOptions{})
	for _, issue := range issues {
		if issue.LinearID == "claude:test-team:1" {
			if issue.Title != "Updated" {
				t.Errorf("expected updated title, got '%s'", issue.Title)
			}
		}
	}
}

func TestScanSkipsEmptySubject(t *testing.T) {
	w, claudeDir, database := setupTestWatcher(t)
	defer database.Close()

	database.CreateProject("TestProj", "", "test-team")
	teamTaskDir := filepath.Join(claudeDir, "tasks", "test-team")
	os.MkdirAll(teamTaskDir, 0755)

	task := ClaudeTask{ID: "1", Subject: "", Status: "pending"}
	data, _ := json.Marshal(task)
	os.WriteFile(filepath.Join(teamTaskDir, "1.json"), data, 0644)

	count, err := w.ScanAndImport()
	if err != nil {
		t.Fatalf("ScanAndImport failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 imported tasks for empty subject, got %d", count)
	}
}

func TestScanNoTasksDir(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	// Don't create the tasks dir

	dbDir := filepath.Join(tmpDir, "project")
	database, err := db.Initialize(dbDir)
	if err != nil {
		t.Fatalf("failed to init test db: %v", err)
	}
	defer database.Close()

	w := NewWatcher(claudeDir, database)
	count, err := w.ScanAndImport()
	if err != nil {
		t.Fatalf("ScanAndImport should not error on missing dir: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 tasks, got %d", count)
	}
}

func TestScanMultipleTeams(t *testing.T) {
	w, claudeDir, database := setupTestWatcher(t)
	defer database.Close()

	database.CreateProject("Proj1", "", "team-a")
	database.CreateProject("Proj2", "", "team-b")

	for _, team := range []string{"team-a", "team-b"} {
		teamDir := filepath.Join(claudeDir, "tasks", team)
		os.MkdirAll(teamDir, 0755)
		task := ClaudeTask{ID: "1", Subject: "Task for " + team, Status: "pending"}
		data, _ := json.Marshal(task)
		os.WriteFile(filepath.Join(teamDir, "1.json"), data, 0644)
	}

	count, err := w.ScanAndImport()
	if err != nil {
		t.Fatalf("ScanAndImport failed: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 tasks from 2 teams, got %d", count)
	}
}

func TestScanTeamWithoutProject(t *testing.T) {
	w, claudeDir, database := setupTestWatcher(t)
	defer database.Close()

	// No project created for this team
	teamTaskDir := filepath.Join(claudeDir, "tasks", "orphan-team")
	os.MkdirAll(teamTaskDir, 0755)

	task := ClaudeTask{ID: "1", Subject: "Orphan task", Status: "in_progress", Owner: "bot"}
	data, _ := json.Marshal(task)
	os.WriteFile(filepath.Join(teamTaskDir, "1.json"), data, 0644)

	count, err := w.ScanAndImport()
	if err != nil {
		t.Fatalf("ScanAndImport failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 task even without project, got %d", count)
	}

	issues, _ := database.ListIssues(db.ListIssuesOptions{})
	for _, issue := range issues {
		if issue.LinearID == "claude:orphan-team:1" {
			if issue.ProjectTag != "" {
				t.Errorf("expected empty project tag for orphan, got '%s'", issue.ProjectTag)
			}
		}
	}
}

func TestScanInvalidJSON(t *testing.T) {
	w, claudeDir, database := setupTestWatcher(t)
	defer database.Close()

	teamTaskDir := filepath.Join(claudeDir, "tasks", "test-team")
	os.MkdirAll(teamTaskDir, 0755)

	// Write invalid JSON
	os.WriteFile(filepath.Join(teamTaskDir, "1.json"), []byte("{bad json"), 0644)

	// Should not error, just skip
	count, err := w.ScanAndImport()
	if err != nil {
		t.Fatalf("ScanAndImport should not fail on bad JSON: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 tasks for invalid JSON, got %d", count)
	}
}

func TestScanSkipsNonJSONFiles(t *testing.T) {
	w, claudeDir, database := setupTestWatcher(t)
	defer database.Close()

	teamTaskDir := filepath.Join(claudeDir, "tasks", "test-team")
	os.MkdirAll(teamTaskDir, 0755)

	// Write non-json file
	os.WriteFile(filepath.Join(teamTaskDir, "readme.txt"), []byte("hello"), 0644)

	count, err := w.ScanAndImport()
	if err != nil {
		t.Fatalf("ScanAndImport failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 tasks for non-JSON file, got %d", count)
	}
}
