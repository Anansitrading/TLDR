package workflow

import (
	"testing"

	"github.com/marcus/td/internal/models"
)

func TestIsValidTransition(t *testing.T) {
	sm := DefaultMachine()

	tests := []struct {
		name     string
		from     models.Status
		to       models.Status
		expected bool
	}{
		// Valid transitions from open
		{"open → in_progress", models.StatusBacklog, models.StatusInFlight, true},
		{"open → blocked", models.StatusBacklog, models.StatusCanceled, true},
		{"open → in_review", models.StatusBacklog, models.StatusReview, true},
		{"open → closed", models.StatusBacklog, models.StatusShipped, true},

		// Valid transitions from in_progress
		{"in_progress → open", models.StatusInFlight, models.StatusBacklog, true},
		{"in_progress → blocked", models.StatusInFlight, models.StatusCanceled, true},
		{"in_progress → in_review", models.StatusInFlight, models.StatusReview, true},
		{"in_progress → closed", models.StatusInFlight, models.StatusShipped, true},

		// Valid transitions from blocked
		{"blocked → open", models.StatusCanceled, models.StatusBacklog, true},
		{"blocked → in_progress", models.StatusCanceled, models.StatusInFlight, true},
		{"blocked → closed", models.StatusCanceled, models.StatusShipped, true},

		// Invalid: blocked cannot go to in_review
		{"blocked → in_review", models.StatusCanceled, models.StatusReview, false},

		// Valid transitions from in_review
		{"in_review → open", models.StatusReview, models.StatusBacklog, true},
		{"in_review → in_progress", models.StatusReview, models.StatusInFlight, true},
		{"in_review → closed", models.StatusReview, models.StatusShipped, true},

		// Invalid: in_review cannot go to blocked
		{"in_review → blocked", models.StatusReview, models.StatusCanceled, false},

		// Valid transitions from closed
		{"closed → open", models.StatusShipped, models.StatusBacklog, true},

		// Invalid: closed cannot go anywhere else
		{"closed → in_progress", models.StatusShipped, models.StatusInFlight, false},
		{"closed → blocked", models.StatusShipped, models.StatusCanceled, false},
		{"closed → in_review", models.StatusShipped, models.StatusReview, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sm.IsValidTransition(tt.from, tt.to)
			if result != tt.expected {
				t.Errorf("IsValidTransition(%s, %s) = %v, want %v", tt.from, tt.to, result, tt.expected)
			}
		})
	}
}

func TestLiberalModeAllowsAllTransitions(t *testing.T) {
	sm := DefaultMachine()

	// Even with guards, liberal mode should allow all valid transitions
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

	results, err := sm.Validate(ctx)
	if err != nil {
		t.Errorf("Liberal mode should allow transition, got error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Liberal mode should skip guards, got %d results", len(results))
	}
}

func TestStrictModeBlocksGuardFailures(t *testing.T) {
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
		t.Error("Strict mode should block transition when BlockedGuard fails")
	}
}

func TestStrictModeAllowsWithForce(t *testing.T) {
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
		Force:      true, // Force flag set
	}

	_, err := sm.Validate(ctx)
	if err != nil {
		t.Errorf("Strict mode should allow transition with --force, got: %v", err)
	}
}

func TestAdvisoryModeReturnsWarnings(t *testing.T) {
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
		t.Errorf("Advisory mode should allow transition, got error: %v", err)
	}
	if len(results) == 0 {
		t.Error("Advisory mode should return guard results")
	}
	if results[0].Passed {
		t.Error("Advisory mode should report guard failure in results")
	}
}

func TestDifferentReviewerGuard(t *testing.T) {
	sm := StrictMachine()

	tests := []struct {
		name        string
		implementer string
		reviewer    string
		minor       bool
		wasInvolved bool
		shouldPass  bool
	}{
		{"different reviewer", "session-1", "session-2", false, false, true},
		{"same reviewer (blocked)", "session-1", "session-1", false, true, false},
		{"same reviewer minor (allowed)", "session-1", "session-1", true, true, true},
		{"no implementer", "", "session-2", false, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issue := &models.Issue{
				ID:                 "test-1",
				Status:             models.StatusReview,
				ImplementerSession: tt.implementer,
				Minor:              tt.minor,
			}

			ctx := &TransitionContext{
				Issue:       issue,
				FromStatus:  models.StatusReview,
				ToStatus:    models.StatusShipped,
				SessionID:   tt.reviewer,
				Minor:       tt.minor,
				WasInvolved: tt.wasInvolved,
			}

			_, err := sm.Validate(ctx)
			passed := err == nil

			if passed != tt.shouldPass {
				t.Errorf("DifferentReviewerGuard: expected pass=%v, got pass=%v (err=%v)",
					tt.shouldPass, passed, err)
			}
		})
	}
}

func TestBlockedGuard(t *testing.T) {
	guard := &BlockedGuard{}

	tests := []struct {
		name       string
		fromStatus models.Status
		force      bool
		shouldPass bool
	}{
		{"blocked without force", models.StatusCanceled, false, false},
		{"blocked with force", models.StatusCanceled, true, true},
		{"open (not blocked)", models.StatusBacklog, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &TransitionContext{
				Issue:      &models.Issue{ID: "test-1", Status: tt.fromStatus},
				FromStatus: tt.fromStatus,
				ToStatus:   models.StatusInFlight,
				Force:      tt.force,
			}

			result := guard.Check(ctx)
			if result.Passed != tt.shouldPass {
				t.Errorf("BlockedGuard: expected pass=%v, got pass=%v", tt.shouldPass, result.Passed)
			}
		})
	}
}

func TestEpicChildrenGuard(t *testing.T) {
	tests := []struct {
		name           string
		issueType      models.Type
		toStatus       models.Status
		openChildCount int
		shouldPass     bool
	}{
		{"epic with open children", models.TypeEpic, models.StatusShipped, 3, false},
		{"epic with no open children", models.TypeEpic, models.StatusShipped, 0, true},
		{"task (not epic)", models.TypeTask, models.StatusShipped, 3, true},
		{"epic not closing", models.TypeEpic, models.StatusInFlight, 3, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			guard := &EpicChildrenGuard{OpenChildCount: tt.openChildCount}
			ctx := &TransitionContext{
				Issue:    &models.Issue{ID: "test-1", Type: tt.issueType},
				ToStatus: tt.toStatus,
			}

			result := guard.Check(ctx)
			if result.Passed != tt.shouldPass {
				t.Errorf("EpicChildrenGuard: expected pass=%v, got pass=%v", tt.shouldPass, result.Passed)
			}
		})
	}
}

func TestGetAllowedTransitions(t *testing.T) {
	sm := DefaultMachine()

	tests := []struct {
		from     models.Status
		expected int
	}{
		{models.StatusTriage, 3},       // backlog, canceled, duplicate
		{models.StatusBacklog, 6},      // triage, prioritized, in_flight, review, canceled, shipped
		{models.StatusPrioritized, 3},  // backlog, in_flight, canceled
		{models.StatusInFlight, 4},     // backlog, review, shipped, canceled
		{models.StatusReview, 3},       // backlog, in_flight, shipped
		{models.StatusShipped, 1},      // backlog only
		{models.StatusCanceled, 3},     // backlog, in_flight, shipped
		{models.StatusDuplicate, 1},    // backlog only
	}

	for _, tt := range tests {
		t.Run(string(tt.from), func(t *testing.T) {
			allowed := sm.GetAllowedTransitions(tt.from)
			if len(allowed) != tt.expected {
				t.Errorf("GetAllowedTransitions(%s) = %d transitions, want %d", tt.from, len(allowed), tt.expected)
			}
		})
	}
}

func TestTransitionName(t *testing.T) {
	tests := []struct {
		from     models.Status
		to       models.Status
		expected string
	}{
		{models.StatusBacklog, models.StatusInFlight, "start"},
		{models.StatusPrioritized, models.StatusInFlight, "start"},
		{models.StatusInFlight, models.StatusBacklog, "unstart"},
		{models.StatusBacklog, models.StatusCanceled, "cancel"},
		{models.StatusCanceled, models.StatusBacklog, "reopen"},
		{models.StatusInFlight, models.StatusReview, "review"},
		{models.StatusReview, models.StatusInFlight, "reject"},
		{models.StatusReview, models.StatusShipped, "approve"},
		{models.StatusInFlight, models.StatusShipped, "ship"},
		{models.StatusShipped, models.StatusBacklog, "reopen"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			name := TransitionName(tt.from, tt.to)
			if name != tt.expected {
				t.Errorf("TransitionName(%s, %s) = %q, want %q", tt.from, tt.to, name, tt.expected)
			}
		})
	}
}

func TestAllStatuses(t *testing.T) {
	statuses := AllStatuses()
	if len(statuses) != 8 {
		t.Errorf("AllStatuses() returned %d statuses, want 8", len(statuses))
	}

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

	for i, s := range expected {
		if statuses[i] != s {
			t.Errorf("AllStatuses()[%d] = %s, want %s", i, statuses[i], s)
		}
	}
}

func TestCanTransition(t *testing.T) {
	sm := DefaultMachine()

	issue := &models.Issue{
		ID:     "test-1",
		Status: models.StatusBacklog,
	}

	ctx := &TransitionContext{
		Issue:      issue,
		FromStatus: models.StatusBacklog,
		ToStatus:   models.StatusInFlight,
		SessionID:  "session-1",
	}

	can, _ := sm.CanTransition(ctx)
	if !can {
		t.Error("CanTransition should return true for valid transition")
	}
}

func TestTransitionError(t *testing.T) {
	err := &TransitionError{
		From:    models.StatusShipped,
		To:      models.StatusInFlight,
		IssueID: "test-123",
		Reason:  "transition not allowed",
	}

	msg := err.Error()
	if msg != "cannot transition test-123 from shipped to in_flight: transition not allowed" {
		t.Errorf("TransitionError.Error() = %q", msg)
	}

	// Test without issue ID
	err.IssueID = ""
	msg = err.Error()
	if msg != "cannot transition from shipped to in_flight: transition not allowed" {
		t.Errorf("TransitionError.Error() = %q", msg)
	}
}

func TestGuardError(t *testing.T) {
	err := &GuardError{
		GuardName: "BlockedGuard",
		IssueID:   "test-123",
		Reason:    "cannot start blocked issue",
	}

	msg := err.Error()
	if msg != "guard BlockedGuard failed for test-123: cannot start blocked issue" {
		t.Errorf("GuardError.Error() = %q", msg)
	}
}

func TestValidationError(t *testing.T) {
	ve := &ValidationError{}

	if ve.HasErrors() {
		t.Error("Empty ValidationError should not have errors")
	}

	ve.Add(&GuardError{GuardName: "G1", Reason: "failed"})
	if !ve.HasErrors() {
		t.Error("ValidationError with error should have errors")
	}
	if ve.Error() != "guard G1 failed: failed" {
		t.Errorf("Single error message: %q", ve.Error())
	}

	ve.Add(&GuardError{GuardName: "G2", Reason: "also failed"})
	if ve.Error() != "2 validation errors" {
		t.Errorf("Multiple errors message: %q", ve.Error())
	}
}

func TestValidateNilContext(t *testing.T) {
	sm := DefaultMachine()

	// Test nil context
	_, err := sm.Validate(nil)
	if err == nil {
		t.Error("Expected error for nil context")
	}
	if te, ok := err.(*TransitionError); !ok {
		t.Errorf("Expected TransitionError, got %T", err)
	} else if te.Reason != "nil context" {
		t.Errorf("Expected 'nil context' reason, got %q", te.Reason)
	}

	// Test nil issue in context
	ctx := &TransitionContext{
		Issue:      nil,
		FromStatus: models.StatusBacklog,
		ToStatus:   models.StatusInFlight,
	}
	_, err = sm.Validate(ctx)
	if err == nil {
		t.Error("Expected error for nil issue")
	}
	if te, ok := err.(*TransitionError); !ok {
		t.Errorf("Expected TransitionError, got %T", err)
	} else if te.Reason != "nil issue in context" {
		t.Errorf("Expected 'nil issue in context' reason, got %q", te.Reason)
	}
}
