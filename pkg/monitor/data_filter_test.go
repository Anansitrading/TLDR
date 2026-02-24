package monitor

import (
	"testing"

	"github.com/Anansitrading/TLDR/internal/models"
)

func TestFilterTaskListByTeamProject_NoFilter(t *testing.T) {
	m := &Model{
		Nav: TeamProjectNav{},
		IssueTeamMap:    map[string]string{"1": "KIJ", "2": "DEV"},
		IssueProjectMap: map[string]string{"1": "ProjectA", "2": "ProjectB"},
	}
	data := TaskListData{
		Ready: []models.Issue{{ID: "1"}, {ID: "2"}},
	}
	result := m.filterTaskListByTeamProject(data)
	if len(result.Ready) != 2 {
		t.Errorf("expected 2 ready issues with no filter, got %d", len(result.Ready))
	}
}

func TestFilterTaskListByTeamProject_ProjectFilter(t *testing.T) {
	m := &Model{
		Nav: TeamProjectNav{
			SelectedTeam:    "KIJ",
			SelectedProject: "ProjectA",
		},
		IssueTeamMap:    map[string]string{"1": "KIJ", "2": "KIJ", "3": "DEV"},
		IssueProjectMap: map[string]string{"1": "ProjectA", "2": "ProjectB", "3": "ProjectA"},
	}
	data := TaskListData{
		Ready:       []models.Issue{{ID: "1"}, {ID: "2"}, {ID: "3"}},
		Reviewable:  []models.Issue{{ID: "1"}},
		Blocked:     []models.Issue{{ID: "2"}},
		NeedsRework: []models.Issue{{ID: "3"}},
		Closed:      []models.Issue{{ID: "1"}, {ID: "2"}},
	}
	result := m.filterTaskListByTeamProject(data)
	if len(result.Ready) != 1 || result.Ready[0].ID != "1" {
		t.Errorf("expected only issue 1 in Ready, got %v", result.Ready)
	}
	if len(result.Reviewable) != 1 {
		t.Errorf("expected 1 reviewable, got %d", len(result.Reviewable))
	}
	if len(result.Blocked) != 0 {
		t.Errorf("expected 0 blocked (issue 2 is ProjectB), got %d", len(result.Blocked))
	}
	if len(result.NeedsRework) != 0 {
		t.Errorf("expected 0 needs rework (issue 3 is DEV team), got %d", len(result.NeedsRework))
	}
	if len(result.Closed) != 1 {
		t.Errorf("expected 1 closed, got %d", len(result.Closed))
	}
}

func TestFilterTaskListByTeamProject_TeamOnlyFilter(t *testing.T) {
	m := &Model{
		Nav: TeamProjectNav{
			SelectedTeam: "KIJ",
		},
		IssueTeamMap:    map[string]string{"1": "KIJ", "2": "DEV"},
		IssueProjectMap: map[string]string{"1": "ProjectA", "2": "ProjectB"},
	}
	data := TaskListData{
		Ready: []models.Issue{{ID: "1"}, {ID: "2"}},
	}
	result := m.filterTaskListByTeamProject(data)
	if len(result.Ready) != 1 || result.Ready[0].ID != "1" {
		t.Errorf("expected only issue 1 (KIJ team), got %v", result.Ready)
	}
}
