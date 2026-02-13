package workflow

import (
	"testing"

	"github.com/marcus/td/internal/models"
)

// TestIntegrationWorkflowScenarios tests common workflow scenarios
func TestIntegrationWorkflowScenarios(t *testing.T) {
	t.Run("full forward lifecycle", func(t *testing.T) {
		sm := DefaultMachine()

		// triage → backlog (accept)
		if !sm.IsValidTransition(models.StatusTriage, models.StatusBacklog) {
			t.Error("Should allow triage → backlog")
		}

		// backlog → prioritized (prioritize)
		if !sm.IsValidTransition(models.StatusBacklog, models.StatusPrioritized) {
			t.Error("Should allow backlog → prioritized")
		}

		// prioritized → in_flight (start)
		if !sm.IsValidTransition(models.StatusPrioritized, models.StatusInFlight) {
			t.Error("Should allow prioritized → in_flight")
		}

		// in_flight → review (submit)
		if !sm.IsValidTransition(models.StatusInFlight, models.StatusReview) {
			t.Error("Should allow in_flight → review")
		}

		// review → shipped (approve)
		if !sm.IsValidTransition(models.StatusReview, models.StatusShipped) {
			t.Error("Should allow review → shipped")
		}
	})

	t.Run("quick start from backlog", func(t *testing.T) {
		sm := DefaultMachine()

		// backlog → in_flight (skip prioritized)
		if !sm.IsValidTransition(models.StatusBacklog, models.StatusInFlight) {
			t.Error("Should allow backlog → in_flight")
		}
	})

	t.Run("cancellation workflow", func(t *testing.T) {
		sm := DefaultMachine()

		// Cancel from various states
		if !sm.IsValidTransition(models.StatusBacklog, models.StatusCanceled) {
			t.Error("Should allow backlog → canceled")
		}
		if !sm.IsValidTransition(models.StatusInFlight, models.StatusCanceled) {
			t.Error("Should allow in_flight → canceled")
		}
		if !sm.IsValidTransition(models.StatusReview, models.StatusCanceled) {
			t.Error("Should allow review → canceled")
		}

		// Reopen from canceled
		if !sm.IsValidTransition(models.StatusCanceled, models.StatusBacklog) {
			t.Error("Should allow canceled → backlog")
		}
		if !sm.IsValidTransition(models.StatusCanceled, models.StatusTriage) {
			t.Error("Should allow canceled → triage")
		}
	})

	t.Run("rejection workflow", func(t *testing.T) {
		sm := DefaultMachine()

		// review → in_flight (reject back to work)
		if !sm.IsValidTransition(models.StatusReview, models.StatusInFlight) {
			t.Error("Should allow review → in_flight")
		}

		// review → backlog (reject back to backlog)
		if !sm.IsValidTransition(models.StatusReview, models.StatusBacklog) {
			t.Error("Should allow review → backlog")
		}
	})

	t.Run("reopen workflow", func(t *testing.T) {
		sm := DefaultMachine()

		// shipped → backlog (reopen)
		if !sm.IsValidTransition(models.StatusShipped, models.StatusBacklog) {
			t.Error("Should allow shipped → backlog")
		}

		// shipped cannot go to anything else
		if sm.IsValidTransition(models.StatusShipped, models.StatusInFlight) {
			t.Error("Should not allow shipped → in_flight directly")
		}
	})

	t.Run("direct ship from in_flight", func(t *testing.T) {
		sm := DefaultMachine()

		// in_flight → shipped (skip review)
		if !sm.IsValidTransition(models.StatusInFlight, models.StatusShipped) {
			t.Error("Should allow in_flight → shipped")
		}
	})
}

// TestIntegrationGuardBehavior tests guard behavior in different modes
func TestIntegrationGuardBehavior(t *testing.T) {
	t.Run("liberal mode ignores guards", func(t *testing.T) {
		sm := DefaultMachine()

		issue := &models.Issue{
			ID:                 "test-1",
			Status:             models.StatusReview,
			ImplementerSession: "session-1",
		}

		ctx := &TransitionContext{
			Issue:       issue,
			FromStatus:  models.StatusReview,
			ToStatus:    models.StatusShipped,
			SessionID:   "session-1",
			Force:       false,
			WasInvolved: true,
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
			Status:             models.StatusReview,
			ImplementerSession: "session-1",
		}

		ctx := &TransitionContext{
			Issue:       issue,
			FromStatus:  models.StatusReview,
			ToStatus:    models.StatusShipped,
			SessionID:   "session-1",
			Force:       false,
			WasInvolved: true,
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
			Status:             models.StatusReview,
			ImplementerSession: "session-1",
		}

		ctx := &TransitionContext{
			Issue:       issue,
			FromStatus:  models.StatusReview,
			ToStatus:    models.StatusShipped,
			SessionID:   "session-1",
			Force:       false,
			WasInvolved: true,
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
		// Cannot skip directly to review from backlog
		{models.StatusBacklog, models.StatusReview},
		// Cannot ship directly from backlog
		{models.StatusBacklog, models.StatusShipped},

		// Canceled is terminal — can only reopen to backlog/triage
		{models.StatusCanceled, models.StatusInFlight},
		{models.StatusCanceled, models.StatusReview},
		{models.StatusCanceled, models.StatusShipped},

		// Duplicate is terminal — can only reopen to backlog/triage
		{models.StatusDuplicate, models.StatusInFlight},
		{models.StatusDuplicate, models.StatusReview},
		{models.StatusDuplicate, models.StatusShipped},

		// Shipped can only reopen to backlog
		{models.StatusShipped, models.StatusInFlight},
		{models.StatusShipped, models.StatusCanceled},
		{models.StatusShipped, models.StatusReview},
		{models.StatusShipped, models.StatusTriage},

		// Triage cannot jump forward past backlog
		{models.StatusTriage, models.StatusInFlight},
		{models.StatusTriage, models.StatusReview},
		{models.StatusTriage, models.StatusShipped},
		{models.StatusTriage, models.StatusPrioritized},
	}

	for _, tt := range invalidTransitions {
		t.Run(string(tt.from)+" → "+string(tt.to), func(t *testing.T) {
			if sm.IsValidTransition(tt.from, tt.to) {
				t.Errorf("Should not allow %s → %s", tt.from, tt.to)
			}
		})
	}
}
