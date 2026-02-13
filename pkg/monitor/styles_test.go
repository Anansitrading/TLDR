package monitor

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/marcus/td/internal/models"
)

func TestKanbanStatusColorsMatchLinearScheme(t *testing.T) {
	expected := map[models.Status]lipgloss.Color{
		models.StatusTriage:      lipgloss.Color("220"), // Yellow
		models.StatusBacklog:     lipgloss.Color("245"), // Gray
		models.StatusPrioritized: lipgloss.Color("69"),  // Blue
		models.StatusInFlight:    lipgloss.Color("214"), // Orange
		models.StatusReview:      lipgloss.Color("141"), // Purple
		models.StatusShipped:     lipgloss.Color("42"),  // Green
		models.StatusCanceled:    lipgloss.Color("196"), // Red
		models.StatusDuplicate:   lipgloss.Color("238"), // Dark gray
	}

	for status, wantColor := range expected {
		gotColor, ok := kanbanStatusColors[status]
		if !ok {
			t.Errorf("kanbanStatusColors missing status %q", status)
			continue
		}
		if gotColor != wantColor {
			t.Errorf("kanbanStatusColors[%q] = %q, want %q", status, gotColor, wantColor)
		}
	}
}

func TestKanbanStatusColorsCoversAllStatuses(t *testing.T) {
	allStatuses := []models.Status{
		models.StatusTriage,
		models.StatusBacklog,
		models.StatusPrioritized,
		models.StatusInFlight,
		models.StatusReview,
		models.StatusShipped,
		models.StatusCanceled,
		models.StatusDuplicate,
	}

	for _, s := range allStatuses {
		if _, ok := kanbanStatusColors[s]; !ok {
			t.Errorf("kanbanStatusColors missing entry for status %q", s)
		}
	}

	if len(kanbanStatusColors) != len(allStatuses) {
		t.Errorf("kanbanStatusColors has %d entries, want %d", len(kanbanStatusColors), len(allStatuses))
	}
}

func TestStatusStylesMatchKanbanColors(t *testing.T) {
	// Verify that statusStyles and kanbanStatusColors cover the same statuses
	for status := range kanbanStatusColors {
		if _, ok := statusStyles[status]; !ok {
			t.Errorf("statusStyles missing status %q that exists in kanbanStatusColors", status)
		}
	}
	for status := range statusStyles {
		if _, ok := kanbanStatusColors[status]; !ok {
			t.Errorf("kanbanStatusColors missing status %q that exists in statusStyles", status)
		}
	}
}
