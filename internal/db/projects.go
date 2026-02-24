package db

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Project represents a first-class project in the system.
type Project struct {
	ID             string
	Name           string
	Description    string
	ClaudeTeamName string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// CreateProject creates a new project. Returns the ID.
func (db *DB) CreateProject(name, description, claudeTeamName string) (string, error) {
	id := uuid.New().String()[:8]
	now := time.Now()
	_, err := db.conn.Exec(
		`INSERT INTO projects (id, name, description, claude_team_name, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		id, name, description, claudeTeamName, now, now,
	)
	if err != nil {
		return "", fmt.Errorf("create project: %w", err)
	}
	return id, nil
}

// ListProjects returns all non-deleted projects.
func (db *DB) ListProjects() ([]Project, error) {
	rows, err := db.conn.Query(
		`SELECT id, name, description, claude_team_name, created_at, updated_at
		 FROM projects WHERE deleted_at IS NULL ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.ClaudeTeamName, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan project: %w", err)
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

// GetProjectByName returns a project by name (non-deleted).
func (db *DB) GetProjectByName(name string) (*Project, error) {
	var p Project
	err := db.conn.QueryRow(
		`SELECT id, name, description, claude_team_name, created_at, updated_at
		 FROM projects WHERE name = ? AND deleted_at IS NULL`, name,
	).Scan(&p.ID, &p.Name, &p.Description, &p.ClaudeTeamName, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get project by name: %w", err)
	}
	return &p, nil
}

// GetProjectByClaudeTeam returns the project linked to a Claude Code team name.
func (db *DB) GetProjectByClaudeTeam(teamName string) (*Project, error) {
	var p Project
	err := db.conn.QueryRow(
		`SELECT id, name, description, claude_team_name, created_at, updated_at
		 FROM projects WHERE claude_team_name = ? AND deleted_at IS NULL`, teamName,
	).Scan(&p.ID, &p.Name, &p.Description, &p.ClaudeTeamName, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get project by claude team: %w", err)
	}
	return &p, nil
}

// LinkProjectToClaudeTeam sets the claude_team_name for a project.
func (db *DB) LinkProjectToClaudeTeam(projectName, claudeTeamName string) error {
	result, err := db.conn.Exec(
		`UPDATE projects SET claude_team_name = ?, updated_at = ?
		 WHERE name = ? AND deleted_at IS NULL`,
		claudeTeamName, time.Now(), projectName,
	)
	if err != nil {
		return fmt.Errorf("link project to claude team: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("project '%s' not found", projectName)
	}
	return nil
}

// DeleteProject soft-deletes a project by name.
func (db *DB) DeleteProject(name string) error {
	result, err := db.conn.Exec(
		`UPDATE projects SET deleted_at = ? WHERE name = ? AND deleted_at IS NULL`,
		time.Now(), name,
	)
	if err != nil {
		return fmt.Errorf("delete project: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("project '%s' not found", name)
	}
	return nil
}
