package workflow

import (
	"testing"

	"github.com/marcus/td/internal/models"
)

// TestIntegrationWorkflowScenarios tests common workflow scenarios
func TestIntegrationWorkflowScenarios(t *testing.T) {
	t.Run("typical issue lifecycle", func(t *testing.T) {
		sm := DefaultMachine()

		// open → in_progress (start)
		if !sm.IsValidTransition(models.StatusBacklog, models.StatusInFlight) {
			t.Error("Should allow open → in_progress")
		}

		// in_progress → in_review (review)
		if !sm.IsValidTransition(models.StatusInFlight, models.StatusReview) {
			t.Error("Should allow in_progress → in_review")
		}

		// in_review → closed (approve)
		if !sm.IsValidTransition(models.StatusReview, models.StatusShipped) {
			t.Error("Should allow in_review → closed")
		}
	})

	t.Run("blocked workflow", func(t *testing.T) {
		sm := DefaultMachine()

		// open → blocked
		if !sm.IsValidTransition(models.StatusBacklog, models.StatusCanceled) {
			t.Error("Should allow open → blocked")
		}

		// blocked → in_progress (with force)
		if !sm.IsValidTransition(models.StatusCanceled, models.StatusInFlight) {
			t.Error("Should allow blocked → in_progress")
		}

		// blocked → open (unblock)
		if !sm.IsValidTransition(models.StatusCanceled, models.StatusBacklog) {
			t.Error("Should allow blocked → open")
		}
	})

	t.Run("rejection workflow", func(t *testing.T) {
		sm := DefaultMachine()

		// in_review → in_progress (reject)
		if !sm.IsValidTransition(models.StatusReview, models.StatusInFlight) {
			t.Error("Should allow in_review → in_progress")
		}

		// in_review → open (reject to open)
		if !sm.IsValidTransition(models.StatusReview, models.StatusBacklog) {
			t.Error("Should allow in_review → open")
		}
	})

	t.Run("reopen workflow", func(t *testing.T) {
		sm := DefaultMachine()

		// closed → open (reopen)
		if !sm.IsValidTransition(models.StatusShipped, models.StatusBacklog) {
			t.Error("Should allow closed → open")
		}

		// closed cannot go to anything else
		if sm.IsValidTransition(models.StatusShipped, models.StatusInFlight) {
			t.Error("Should not allow closed → in_progress directly")
		}
	})

	t.Run("direct close scenarios", func(t *testing.T) {
		sm := DefaultMachine()

		// open → closed (admin close)
		if !sm.IsValidTransition(models.StatusBacklog, models.StatusShipped) {
			t.Error("Should allow open → closed")
		}

		// in_progress → closed (close without review)
		if !sm.IsValidTransition(models.StatusInFlight, models.StatusShipped) {
			t.Error("Should allow in_progress → closed")
		}

		// blocked → closed (close blocked issue)
		if !sm.IsValidTransition(models.StatusCanceled, models.StatusShipped) {
			t.Error("Should allow blocked → closed")
		}
	})
}

// TestIntegrationGuardBehavior tests guard behavior in different modes
func TestIntegrationGuardBehavior(t *testing.T) {
	t.Run("liberal mode ignores guards", func(t *testing.T) {
		sm := DefaultMachine()

		issue := &models.Issue{
			ID:                 "test-1",
			Status:             models.StatusCanceled,
			ImplementerSession: "session-1",
		}

		ctx := &TransitionContext{
			Issue:      issue,
			FromStatus: models.StatusCanceled,
			ToStatus:   models.StatusInFlight,
			SessionID:  "session-1",
			Force:      false, // Not forcing
		}

		_, err := sm.Validate(ctx)
		if err != nil {
			t.Errorf("Liberal mode should allow all transitions, got: %v", err)
		}
	})

	t.Run("advisory mode returns warnings", func(t *testing.T) {
		sm := AdvisoryMachine()

		issue := &models.Issue{
			ID:                 "test-1",
			Status:             models.StatusCanceled,
			ImplementerSession: "session-1",
		}

		ctx := &TransitionContext{
			Issue:      issue,
			FromStatus: models.StatusCanceled,
			ToStatus:   models.StatusInFlight,
			SessionID:  "session-1",
			Force:      false,
		}

		results, err := sm.Validate(ctx)
		if err != nil {
			t.Errorf("Advisory mode should allow transition, got: %v", err)
		}
		if len(results) == 0 || results[0].Passed {
			t.Error("Advisory mode should return failed guard results")
		}
	})

	t.Run("strict mode blocks on guard failure", func(t *testing.T) {
		sm := StrictMachine()

		issue := &models.Issue{
			ID:                 "test-1",
			Status:             models.StatusCanceled,
			ImplementerSession: "session-1",
		}

		ctx := &TransitionContext{
			Issue:      issue,
			FromStatus: models.StatusCanceled,
			ToStatus:   models.StatusInFlight,
			SessionID:  "session-1",
			Force:      false,
		}

		_, err := sm.Validate(ctx)
		if err == nil {
			t.Error("Strict mode should block transition when guard fails")
		}
	})
}

// TestIntegrationInvalidTransitions ensures invalid paths are blocked
func TestIntegrationInvalidTransitions(t *testing.T) {
	sm := DefaultMachine()

	invalidTransitions := []struct {
		from models.Status
		to   models.Status
	}{
		// blocked cannot go to in_review
		{models.StatusCanceled, models.StatusReview},

		// in_review cannot go to blocked
		{models.StatusReview, models.StatusCanceled},

		// closed can only go to open
		{models.StatusShipped, models.StatusInFlight},
		{models.StatusShipped, models.StatusCanceled},
		{models.StatusShipped, models.StatusReview},
	}

	for _, tt := range invalidTransitions {
		t.Run(string(tt.from)+" → "+string(tt.to), func(t *testing.T) {
			if sm.IsValidTransition(tt.from, tt.to) {
				t.Errorf("Should not allow %s → %s", tt.from, tt.to)
			}
		})
	}
}
