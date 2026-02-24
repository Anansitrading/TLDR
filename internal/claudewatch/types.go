package claudewatch

import "time"

// ClaudeTask represents a task from Claude Code's JSON task file
type ClaudeTask struct {
	ID          string            `json:"id"`
	Subject     string            `json:"subject"`
	Description string            `json:"description"`
	ActiveForm  string            `json:"activeForm"`
	Status      string            `json:"status"` // pending, in_progress, completed, deleted
	Blocks      []string          `json:"blocks"`
	BlockedBy   []string          `json:"blockedBy"`
	Owner       string            `json:"owner"`
	Metadata    map[string]string `json:"metadata"`
}

// ClaudeTeamConfig represents Claude Code's team config.json
type ClaudeTeamConfig struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	CreatedAt   int64              `json:"createdAt"`
	Members     []ClaudeTeamMember `json:"members"`
}

// ClaudeTeamMember represents a member in a Claude Code team
type ClaudeTeamMember struct {
	AgentID   string `json:"agentId"`
	Name      string `json:"name"`
	AgentType string `json:"agentType"`
	Model     string `json:"model"`
	JoinedAt  int64  `json:"joinedAt"`
}

// WatchEvent represents a change detected by the watcher
type WatchEvent struct {
	TeamName string
	TaskID   string
	Task     *ClaudeTask // nil on delete
	Time     time.Time
}
