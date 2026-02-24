package claudewatch

import (
	"testing"

	"github.com/Anansitrading/TLDR/internal/models"
)

func TestMapClaudeTaskToIssue(t *testing.T) {
	task := &ClaudeTask{
		ID:          "3",
		Subject:     "Implement retry logic",
		Description: "Add exponential backoff",
		Status:      "in_progress",
		Owner:       "researcher",
	}

	issue := MapToIssue(task, "my-team", "MyProject")

	if issue.Title != "Implement retry logic" {
		t.Errorf("expected title 'Implement retry logic', got '%s'", issue.Title)
	}
	if issue.Status != models.StatusInFlight {
		t.Errorf("expected status in_flight, got %s", issue.Status)
	}
	if issue.Assignee != "researcher" {
		t.Errorf("expected assignee 'researcher', got '%s'", issue.Assignee)
	}
	if issue.ProjectTag != "MyProject" {
		t.Errorf("expected project tag 'MyProject', got '%s'", issue.ProjectTag)
	}
	if issue.LinearID != "claude:my-team:3" {
		t.Errorf("expected linear_id 'claude:my-team:3', got '%s'", issue.LinearID)
	}
}

func TestMapClaudeStatusToTdStatus(t *testing.T) {
	tests := []struct {
		claude string
		td     models.Status
	}{
		{"pending", models.StatusBacklog},
		{"in_progress", models.StatusInFlight},
		{"completed", models.StatusShipped},
		{"deleted", models.StatusCanceled},
		{"unknown", models.StatusBacklog},
	}
	for _, tt := range tests {
		got := MapStatus(tt.claude)
		if got != tt.td {
			t.Errorf("MapStatus(%q) = %s, want %s", tt.claude, got, tt.td)
		}
	}
}

func TestLinearIDFormat(t *testing.T) {
	id := MakeLinearID("team-alpha", "5")
	if id != "claude:team-alpha:5" {
		t.Errorf("expected 'claude:team-alpha:5', got '%s'", id)
	}
}
