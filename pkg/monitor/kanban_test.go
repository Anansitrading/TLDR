package monitor

import (
	"testing"

	"github.com/marcus/td/internal/models"
)

func TestKanbanStatusOrder(t *testing.T) {
	expected := []models.Status{
		models.StatusTriage,
		models.StatusBacklog,
		models.StatusPrioritized,
		models.StatusInFlight,
		models.StatusReview,
		models.StatusShipped,
		models.StatusCanceled,
		models.StatusDuplicate,
	}

	if len(kanbanStatusOrder) != len(expected) {
		t.Fatalf("kanbanStatusOrder has %d entries, want %d", len(kanbanStatusOrder), len(expected))
	}

	for i, want := range expected {
		if kanbanStatusOrder[i] != want {
			t.Errorf("kanbanStatusOrder[%d] = %q, want %q", i, kanbanStatusOrder[i], want)
		}
	}
}

func TestKanbanStatusOrderCoversAllStatuses(t *testing.T) {
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

	seen := make(map[models.Status]bool)
	for _, s := range kanbanStatusOrder {
		if seen[s] {
			t.Errorf("duplicate status in kanbanStatusOrder: %q", s)
		}
		seen[s] = true
	}

	for _, s := range allStatuses {
		if !seen[s] {
			t.Errorf("status %q missing from kanbanStatusOrder", s)
		}
	}
}
