package claudewatch

import (
	"fmt"

	"github.com/Anansitrading/TLDR/internal/models"
)

// MakeLinearID creates a dedup key for a Claude Code task.
func MakeLinearID(teamName, taskID string) string {
	return fmt.Sprintf("claude:%s:%s", teamName, taskID)
}

// MakeLinearIdentifier creates a display identifier for a Claude Code task.
func MakeLinearIdentifier(teamName, taskID string) string {
	return fmt.Sprintf("claude-%s:%s", teamName, taskID)
}

// MapStatus converts Claude Code task status to td issue status.
func MapStatus(claudeStatus string) models.Status {
	switch claudeStatus {
	case "pending":
		return models.StatusBacklog
	case "in_progress":
		return models.StatusInFlight
	case "completed":
		return models.StatusShipped
	case "deleted":
		return models.StatusCanceled
	default:
		return models.StatusBacklog
	}
}

// MapToIssue converts a Claude Code task into a td Issue.
// teamName is used for the dedup key. projectTag is the td project to assign to.
func MapToIssue(task *ClaudeTask, teamName, projectTag string) *models.Issue {
	return &models.Issue{
		Title:            task.Subject,
		Description:      task.Description,
		Status:           MapStatus(task.Status),
		Type:             models.TypeTask,
		Priority:         models.PriorityP2,
		Assignee:         task.Owner,
		ProjectTag:       projectTag,
		LinearID:         MakeLinearID(teamName, task.ID),
		LinearIdentifier: MakeLinearIdentifier(teamName, task.ID),
	}
}
