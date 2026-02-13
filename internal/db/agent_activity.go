package db

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/marcus/td/internal/models"
)

const activityIDPrefix = "aa-"

// generateActivityID generates a unique agent activity ID
func generateActivityID() (string, error) {
	bytes := make([]byte, 4) // 8 hex characters
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return activityIDPrefix + hex.EncodeToString(bytes), nil
}

// RecordAgentActivity records an agent action on an issue
func (db *DB) RecordAgentActivity(activity *models.AgentActivity) error {
	return db.withWriteLock(func() error {
		id, err := generateActivityID()
		if err != nil {
			return err
		}
		activity.ID = id
		if activity.CreatedAt.IsZero() {
			activity.CreatedAt = time.Now()
		}

		_, err = db.conn.Exec(`
			INSERT INTO agent_activity (id, issue_id, agent_name, activity_type, details, session_id, worktree_path, branch, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, activity.ID, activity.IssueID, activity.AgentName, activity.ActivityType,
			activity.Details, activity.SessionID, activity.WorktreePath, activity.Branch, activity.CreatedAt)
		return err
	})
}

// ListAgentActivity returns activity records for an issue, newest first
func (db *DB) ListAgentActivity(issueID string, limit int) ([]models.AgentActivity, error) {
	issueID = NormalizeIssueID(issueID)
	query := `SELECT id, issue_id, agent_name, activity_type, details, session_id, worktree_path, branch, created_at
	          FROM agent_activity WHERE issue_id = ? ORDER BY created_at DESC`
	var args []interface{}
	args = append(args, issueID)
	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []models.AgentActivity
	for rows.Next() {
		var a models.AgentActivity
		var details, sessionID, worktreePath, branch sql.NullString
		if err := rows.Scan(&a.ID, &a.IssueID, &a.AgentName, &a.ActivityType,
			&details, &sessionID, &worktreePath, &branch, &a.CreatedAt); err != nil {
			return nil, err
		}
		a.Details = details.String
		a.SessionID = sessionID.String
		a.WorktreePath = worktreePath.String
		a.Branch = branch.String
		activities = append(activities, a)
	}
	return activities, nil
}

// ListAgentActivityByAgent returns all activity for a specific agent, newest first
func (db *DB) ListAgentActivityByAgent(agentName string, limit int) ([]models.AgentActivity, error) {
	query := `SELECT id, issue_id, agent_name, activity_type, details, session_id, worktree_path, branch, created_at
	          FROM agent_activity WHERE agent_name = ? ORDER BY created_at DESC`
	var args []interface{}
	args = append(args, agentName)
	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []models.AgentActivity
	for rows.Next() {
		var a models.AgentActivity
		var details, sessionID, worktreePath, branch sql.NullString
		if err := rows.Scan(&a.ID, &a.IssueID, &a.AgentName, &a.ActivityType,
			&details, &sessionID, &worktreePath, &branch, &a.CreatedAt); err != nil {
			return nil, err
		}
		a.Details = details.String
		a.SessionID = sessionID.String
		a.WorktreePath = worktreePath.String
		a.Branch = branch.String
		activities = append(activities, a)
	}
	return activities, nil
}

// AssignIssue sets the assignee on an issue and records the assignment activity
func (db *DB) AssignIssue(issueID, agentName, sessionID string) error {
	issueID = NormalizeIssueID(issueID)
	return db.withWriteLock(func() error {
		now := time.Now()
		_, err := db.conn.Exec(`UPDATE issues SET assignee = ?, updated_at = ? WHERE id = ?`,
			agentName, now, issueID)
		if err != nil {
			return fmt.Errorf("assign issue: %w", err)
		}

		// Record the activity
		id, err := generateActivityID()
		if err != nil {
			return err
		}
		activityAgent := agentName
		details := fmt.Sprintf("Assigned to %s", agentName)
		if agentName == "" {
			activityAgent = "system"
			details = "Unassigned"
		}
		_, err = db.conn.Exec(`
			INSERT INTO agent_activity (id, issue_id, agent_name, activity_type, details, session_id, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, id, issueID, activityAgent, models.ActivityAssigned,
			details, sessionID, now)
		return err
	})
}
