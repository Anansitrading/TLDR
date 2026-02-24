package db

import (
	"testing"
)

func setupTestDB(t *testing.T) *DB {
	t.Helper()
	tmpDir := t.TempDir()
	database, err := Initialize(tmpDir)
	if err != nil {
		t.Fatalf("init db: %v", err)
	}
	return database
}

func TestProjectsCRUD(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	// Create
	id, err := database.CreateProject("TestProject", "A test project", "")
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}
	if id == "" {
		t.Fatal("expected non-empty project ID")
	}

	// Get by name
	proj, err := database.GetProjectByName("TestProject")
	if err != nil {
		t.Fatalf("GetProjectByName failed: %v", err)
	}
	if proj.Name != "TestProject" {
		t.Errorf("expected name 'TestProject', got '%s'", proj.Name)
	}
	if proj.Description != "A test project" {
		t.Errorf("expected description 'A test project', got '%s'", proj.Description)
	}

	// List
	projects, err := database.ListProjects()
	if err != nil {
		t.Fatalf("ListProjects failed: %v", err)
	}
	if len(projects) != 1 {
		t.Errorf("expected 1 project, got %d", len(projects))
	}

	// Link to claude team
	err = database.LinkProjectToClaudeTeam("TestProject", "my-team")
	if err != nil {
		t.Fatalf("LinkProjectToClaudeTeam failed: %v", err)
	}
	proj, _ = database.GetProjectByName("TestProject")
	if proj.ClaudeTeamName != "my-team" {
		t.Errorf("expected claude_team_name 'my-team', got '%s'", proj.ClaudeTeamName)
	}

	// Delete (soft)
	err = database.DeleteProject("TestProject")
	if err != nil {
		t.Fatalf("DeleteProject failed: %v", err)
	}
	projects, _ = database.ListProjects()
	if len(projects) != 0 {
		t.Errorf("expected 0 projects after delete, got %d", len(projects))
	}
}

func TestProjectDuplicateName(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	_, err := database.CreateProject("Dup", "", "")
	if err != nil {
		t.Fatalf("first create failed: %v", err)
	}
	_, err = database.CreateProject("Dup", "", "")
	if err == nil {
		t.Fatal("expected error on duplicate project name")
	}
}

func TestGetProjectByClaudeTeam(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	database.CreateProject("Linked", "", "team-alpha")
	proj, err := database.GetProjectByClaudeTeam("team-alpha")
	if err != nil {
		t.Fatalf("GetProjectByClaudeTeam failed: %v", err)
	}
	if proj.Name != "Linked" {
		t.Errorf("expected 'Linked', got '%s'", proj.Name)
	}
}
