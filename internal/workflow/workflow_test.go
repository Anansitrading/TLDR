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
		// Valid transitions from triage
		{"triage → backlog", models.StatusTriage, models.StatusBacklog, true},
		{"triage → canceled", models.StatusTriage, models.StatusCanceled, true},
		{"triage → duplicate", models.StatusTriage, models.StatusDuplicate, true},
		{"triage → in_flight (invalid)", models.StatusTriage, models.StatusInFlight, false},

		// Valid transitions from backlog
		{"backlog → triage", models.StatusBacklog, models.StatusTriage, true},
		{"backlog → prioritized", models.StatusBacklog, models.StatusPrioritized, true},
		{"backlog → in_flight", models.StatusBacklog, models.StatusInFlight, true},
		{"backlog → canceled", models.StatusBacklog, models.StatusCanceled, true},
		{"backlog → duplicate", models.StatusBacklog, models.StatusDuplicate, true},
		{"backlog → review (invalid)", models.StatusBacklog, models.StatusReview, false},
		{"backlog → shipped (invalid)", models.StatusBacklog, models.StatusShipped, false},

		// Valid transitions from prioritized
		{"prioritized → backlog", models.StatusPrioritized, models.StatusBacklog, true},
		{"prioritized → in_flight", models.StatusPrioritized, models.StatusInFlight, true},
		{"prioritized → canceled", models.StatusPrioritized, models.StatusCanceled, true},
		{"prioritized → duplicate", models.StatusPrioritized, models.StatusDuplicate, true},
		{"prioritized → review (invalid)", models.StatusPrioritized, models.StatusReview, false},

		// Valid transitions from in_flight
		{"in_flight → review", models.StatusInFlight, models.StatusReview, true},
		{"in_flight → shipped", models.StatusInFlight, models.StatusShipped, true},
		{"in_flight → backlog", models.StatusInFlight, models.StatusBacklog, true},
		{"in_flight → canceled", models.StatusInFlight, models.StatusCanceled, true},
		{"in_flight → duplicate (invalid)", models.StatusInFlight, models.StatusDuplicate, false},

		// Valid transitions from review
		{"review → shipped", models.StatusReview, models.StatusShipped, true},
		{"review → in_flight", models.StatusReview, models.StatusInFlight, true},
		{"review → backlog", models.StatusReview, models.StatusBacklog, true},
		{"review → canceled", models.StatusReview, models.StatusCanceled, true},
		{"review → duplicate (invalid)", models.StatusReview, models.StatusDuplicate, false},

		// Valid transitions from shipped (terminal — reopen only)
		{"shipped → backlog", models.StatusShipped, models.StatusBacklog, true},
		{"shipped → in_flight (invalid)", models.StatusShipped, models.StatusInFlight, false},
		{"shipped → canceled (invalid)", models.StatusShipped, models.StatusCanceled, false},
		{"shipped → review (invalid)", models.StatusShipped, models.StatusReview, false},

		// Valid transitions from canceled (terminal — reopen only)
		{"canceled → backlog", models.StatusCanceled, models.StatusBacklog, true},
		{"canceled → triage", models.StatusCanceled, models.StatusTriage, true},
		{"canceled → in_flight (invalid)", models.StatusCanceled, models.StatusInFlight, false},
		{"canceled → review (invalid)", models.StatusCanceled, models.StatusReview, false},

		// Valid transitions from duplicate (terminal — reopen only)
		{"duplicate → backlog", models.StatusDuplicate, models.StatusBacklog, true},
		{"duplicate → triage", models.StatusDuplicate, models.StatusTriage, true},
		{"duplicate → in_flight (invalid)", models.StatusDuplicate, models.StatusInFlight, false},
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
		Status:             models.StatusReview,
		ImplementerSession: "session-1",
	}

	ctx := &TransitionContext{
		Issue:       issue,
		FromStatus:  models.StatusReview,
		ToStatus:    models.StatusShipped,
		SessionID:   "session-1",
		Force:       false,
		WasInvolved: true, // Would normally trigger DifferentReviewerGuard
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
		t.Error("Strict mode should block transition when DifferentReviewerGuard fails")
	}
}

func TestStrictModeAllowsWithMinor(t *testing.T) {
	sm := StrictMachine()

	issue := &models.Issue{
		ID:                 "test-1",
		Status:             models.StatusReview,
		ImplementerSession: "session-1",
		Minor:              true,
	}

	ctx := &TransitionContext{
		Issue:       issue,
		FromStatus:  models.StatusReview,
		ToStatus:    models.StatusShipped,
		SessionID:   "session-1",
		Minor:       true, // Minor flag allows self-approval
		WasInvolved: true,
	}

	_, err := sm.Validate(ctx)
	if err != nil {
		t.Errorf("Strict mode should allow self-approval for minor issues, got: %v", err)
	}
}

func TestAdvisoryModeReturnsWarnings(t *testing.T) {
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
		{models.StatusTriage, 3},      // backlog, canceled, duplicate
		{models.StatusBacklog, 5},     // triage, prioritized, in_flight, canceled, duplicate
		{models.StatusPrioritized, 4}, // backlog, in_flight, canceled, duplicate
		{models.StatusInFlight, 4},    // review, shipped, backlog, canceled
		{models.StatusReview, 4},      // shipped, in_flight, backlog, canceled
		{models.StatusShipped, 1},     // backlog only
		{models.StatusCanceled, 2},    // backlog, triage
		{models.StatusDuplicate, 2},   // backlog, triage
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
		name     string
		from     models.Status
		to       models.Status
		expected string
	}{
		{"start from backlog", models.StatusBacklog, models.StatusInFlight, "start"},
		{"start from prioritized", models.StatusPrioritized, models.StatusInFlight, "start"},
		{"unstart", models.StatusInFlight, models.StatusBacklog, "unstart"},
		{"cancel from backlog", models.StatusBacklog, models.StatusCanceled, "cancel"},
		{"cancel from in_flight", models.StatusInFlight, models.StatusCanceled, "cancel"},
		{"cancel from review", models.StatusReview, models.StatusCanceled, "cancel"},
		{"duplicate from triage", models.StatusTriage, models.StatusDuplicate, "duplicate"},
		{"duplicate from backlog", models.StatusBacklog, models.StatusDuplicate, "duplicate"},
		{"review", models.StatusInFlight, models.StatusReview, "review"},
		{"reject to in_flight", models.StatusReview, models.StatusInFlight, "reject"},
		{"reject to backlog", models.StatusReview, models.StatusBacklog, "reject"},
		{"approve", models.StatusReview, models.StatusShipped, "approve"},
		{"ship", models.StatusInFlight, models.StatusShipped, "ship"},
		{"reopen from shipped", models.StatusShipped, models.StatusBacklog, "reopen"},
		{"reopen from canceled", models.StatusCanceled, models.StatusBacklog, "reopen"},
		{"reopen canceled to triage", models.StatusCanceled, models.StatusTriage, "reopen"},
		{"reopen from duplicate", models.StatusDuplicate, models.StatusBacklog, "reopen"},
		{"reopen duplicate to triage", models.StatusDuplicate, models.StatusTriage, "reopen"},
		{"prioritize", models.StatusBacklog, models.StatusPrioritized, "prioritize"},
		{"deprioritize", models.StatusPrioritized, models.StatusBacklog, "deprioritize"},
		{"triage", models.StatusBacklog, models.StatusTriage, "triage"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
