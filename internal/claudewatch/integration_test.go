package claudewatch

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Anansitrading/TLDR/internal/db"
)

func TestEndToEnd_CreateTeamAndTasks(t *testing.T) {
	w, claudeDir, database := setupTestWatcher(t)
	defer database.Close()

	// 1. Create a project linked to a Claude team
	_, err := database.CreateProject("E2E-Project", "End to end test", "e2e-team")
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}

	// 2. Create team config
	teamDir := filepath.Join(claudeDir, "teams", "e2e-team")
	os.MkdirAll(teamDir, 0755)
	config := ClaudeTeamConfig{
		Name: "e2e-team",
		Members: []ClaudeTeamMember{
			{Name: "lead", AgentType: "lead"},
			{Name: "worker-1", AgentType: "general-purpose"},
			{Name: "worker-2", AgentType: "general-purpose"},
		},
	}
	configData, _ := json.Marshal(config)
	os.WriteFile(filepath.Join(teamDir, "config.json"), configData, 0644)

	// 3. Create task files (simulating Claude Code creating tasks)
	taskDir := filepath.Join(claudeDir, "tasks", "e2e-team")
	os.MkdirAll(taskDir, 0755)

	tasks := []ClaudeTask{
		{ID: "1", Subject: "Research phase", Status: "completed", Owner: "worker-1"},
		{ID: "2", Subject: "Implement feature", Status: "in_progress", Owner: "worker-2"},
		{ID: "3", Subject: "Write tests", Status: "pending", Owner: "worker-1", BlockedBy: []string{"2"}},
	}
	for _, task := range tasks {
		data, _ := json.Marshal(task)
		os.WriteFile(filepath.Join(taskDir, task.ID+".json"), data, 0644)
	}

	// 4. Scan and import
	count, err := w.ScanAndImport()
	if err != nil {
		t.Fatalf("ScanAndImport failed: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 imported tasks, got %d", count)
	}

	// 5. Verify all tasks exist in database with correct project tag
	issues, _ := database.ListIssues(db.ListIssuesOptions{ProjectTag: "E2E-Project"})
	if len(issues) != 3 {
		t.Errorf("expected 3 issues with project tag E2E-Project, got %d", len(issues))
	}

	// 6. Verify agent assignment
	agentCounts := map[string]int{}
	for _, issue := range issues {
		agentCounts[issue.Assignee]++
	}
	if agentCounts["worker-1"] != 2 {
		t.Errorf("expected 2 tasks for worker-1, got %d", agentCounts["worker-1"])
	}
	if agentCounts["worker-2"] != 1 {
		t.Errorf("expected 1 task for worker-2, got %d", agentCounts["worker-2"])
	}

	// 7. Simulate task update (worker-2 completes task 2)
	tasks[1].Status = "completed"
	data, _ := json.Marshal(tasks[1])
	os.WriteFile(filepath.Join(taskDir, "2.json"), data, 0644)

	count2, err := w.ScanAndImport()
	if err != nil {
		t.Fatalf("Second ScanAndImport failed: %v", err)
	}
	// Second scan should still import/update 3 tasks
	if count2 != 3 {
		t.Errorf("expected 3 tasks on re-scan, got %d", count2)
	}

	// Verify status updated
	issues, _ = database.ListIssues(db.ListIssuesOptions{ProjectTag: "E2E-Project"})
	for _, issue := range issues {
		if issue.LinearID == "claude:e2e-team:2" && issue.Status != "shipped" {
			t.Errorf("expected task 2 status shipped after completion, got %s", issue.Status)
		}
	}
}
