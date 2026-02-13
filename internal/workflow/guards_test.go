package workflow

import (
	"testing"

	"github.com/marcus/td/internal/models"
)

// =============================================================================
// BlockedGuard Tests
// =============================================================================

func TestBlockedGuard_Name(t *testing.T) {
	guard := &BlockedGuard{}
	if guard.Name() != "BlockedGuard" {
		t.Errorf("Name() = %q, want %q", guard.Name(), "BlockedGuard")
	}
}

func TestBlockedGuard_WithForceFlag(t *testing.T) {
	guard := &BlockedGuard{}

	ctx := &TransitionContext{
		Issue:      &models.Issue{ID: "test-1", Status: models.StatusCanceled},
		FromStatus: models.StatusCanceled,
		ToStatus:   models.StatusInFlight,
		Force:      true,
	}

	result := guard.Check(ctx)
	if !result.Passed {
		t.Errorf("BlockedGuard should pass with force flag, got: %s", result.Message)
	}
}

func TestBlockedGuard_WithoutForceFlag(t *testing.T) {
	guard := &BlockedGuard{}

	ctx := &TransitionContext{
		Issue:      &models.Issue{ID: "test-1", Status: models.StatusCanceled},
		FromStatus: models.StatusCanceled,
		ToStatus:   models.StatusInFlight,
		Force:      false,
	}

	result := guard.Check(ctx)
	if result.Passed {
		t.Error("BlockedGuard should fail without force flag for blocked issue")
	}
	if result.Message == "" {
		t.Error("BlockedGuard should provide failure message")
	}
}

func TestBlockedGuard_VariousSourceStatuses(t *testing.T) {
	guard := &BlockedGuard{}

	tests := []struct {
		name       string
		fromStatus models.Status
		toStatus   models.Status
		force      bool
		shouldPass bool
	}{
		// Blocked → in_progress without force should fail
		{"blocked→in_progress no force", models.StatusCanceled, models.StatusInFlight, false, false},
		// Blocked → in_progress with force should pass
		{"blocked→in_progress with force", models.StatusCanceled, models.StatusInFlight, true, true},
		// Blocked → open (unblock) without force should fail
		{"blocked→open no force", models.StatusCanceled, models.StatusBacklog, false, false},
		// Blocked → closed without force should fail
		{"blocked→closed no force", models.StatusCanceled, models.StatusShipped, false, false},

		// Non-blocked statuses should always pass
		{"open→in_progress", models.StatusBacklog, models.StatusInFlight, false, true},
		{"open→blocked", models.StatusBacklog, models.StatusCanceled, false, true},
		{"in_progress→in_review", models.StatusInFlight, models.StatusReview, false, true},
		{"in_progress→blocked", models.StatusInFlight, models.StatusCanceled, false, true},
		{"in_review→closed", models.StatusReview, models.StatusShipped, false, true},
		{"closed→open", models.StatusShipped, models.StatusBacklog, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &TransitionContext{
				Issue:      &models.Issue{ID: "test-1", Status: tt.fromStatus},
				FromStatus: tt.fromStatus,
				ToStatus:   tt.toStatus,
				Force:      tt.force,
			}

			result := guard.Check(ctx)
			if result.Passed != tt.shouldPass {
				t.Errorf("BlockedGuard.Check() passed=%v, want %v (msg: %s)",
					result.Passed, tt.shouldPass, result.Message)
			}
		})
	}
}

// =============================================================================
// DifferentReviewerGuard Tests
// =============================================================================

func TestDifferentReviewerGuard_Name(t *testing.T) {
	guard := &DifferentReviewerGuard{}
	if guard.Name() != "DifferentReviewerGuard" {
		t.Errorf("Name() = %q, want %q", guard.Name(), "DifferentReviewerGuard")
	}
}

func TestDifferentReviewerGuard_SameSession(t *testing.T) {
	guard := &DifferentReviewerGuard{}

	ctx := &TransitionContext{
		Issue: &models.Issue{
			ID:                 "test-1",
			Status:             models.StatusReview,
			ImplementerSession: "session-1",
		},
		FromStatus:  models.StatusReview,
		ToStatus:    models.StatusShipped,
		SessionID:   "session-1", // Same as implementer
		WasInvolved: true,
	}

	result := guard.Check(ctx)
	if result.Passed {
		t.Error("DifferentReviewerGuard should fail when reviewer is implementer")
	}
	if result.Message != "cannot approve your own implementation" {
		t.Errorf("Unexpected message: %s", result.Message)
	}
}

func TestDifferentReviewerGuard_DifferentSession(t *testing.T) {
	guard := &DifferentReviewerGuard{}

	ctx := &TransitionContext{
		Issue: &models.Issue{
			ID:                 "test-1",
			Status:             models.StatusReview,
			ImplementerSession: "session-1",
		},
		FromStatus:  models.StatusReview,
		ToStatus:    models.StatusShipped,
		SessionID:   "session-2", // Different from implementer
		WasInvolved: false,
	}

	result := guard.Check(ctx)
	if !result.Passed {
		t.Errorf("DifferentReviewerGuard should pass for different reviewer: %s", result.Message)
	}
}

func TestDifferentReviewerGuard_MinorTask(t *testing.T) {
	guard := &DifferentReviewerGuard{}

	tests := []struct {
		name        string
		issueMinor  bool
		ctxMinor    bool
		sameSession bool
		shouldPass  bool
	}{
		// Minor tasks should always allow self-approval
		{"issue marked minor", true, false, true, true},
		{"context marked minor", false, true, true, true},
		{"both marked minor", true, true, true, true},
		// Non-minor tasks should block self-approval
		{"not minor same session", false, false, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionID := "session-1"
			implementer := "session-1"
			if !tt.sameSession {
				implementer = "session-2"
			}

			ctx := &TransitionContext{
				Issue: &models.Issue{
					ID:                 "test-1",
					Status:             models.StatusReview,
					ImplementerSession: implementer,
					Minor:              tt.issueMinor,
				},
				FromStatus:  models.StatusReview,
				ToStatus:    models.StatusShipped,
				SessionID:   sessionID,
				Minor:       tt.ctxMinor,
				WasInvolved: tt.sameSession,
			}

			result := guard.Check(ctx)
			if result.Passed != tt.shouldPass {
				t.Errorf("passed=%v, want %v (msg: %s)", result.Passed, tt.shouldPass, result.Message)
			}
		})
	}
}

func TestDifferentReviewerGuard_AdminBypass(t *testing.T) {
	guard := &DifferentReviewerGuard{}

	ctx := &TransitionContext{
		Issue: &models.Issue{
			ID:                 "test-1",
			Status:             models.StatusReview,
			ImplementerSession: "session-1",
		},
		FromStatus:  models.StatusReview,
		ToStatus:    models.StatusShipped,
		SessionID:   "session-1", // Same as implementer
		Context:     ContextAdmin,
		WasInvolved: true,
	}

	result := guard.Check(ctx)
	if !result.Passed {
		t.Errorf("DifferentReviewerGuard should pass with admin context: %s", result.Message)
	}
}

func TestDifferentReviewerGuard_WasInvolved(t *testing.T) {
	guard := &DifferentReviewerGuard{}

	tests := []struct {
		name          string
		implementer   string
		sessionID     string
		wasInvolved   bool
		shouldPass    bool
		expectMessage string
	}{
		// Not involved should pass
		{"not involved", "session-1", "session-2", false, true, ""},
		// Involved but different session should fail with "involved" message
		{"involved different session", "session-1", "session-2", true, false, "cannot approve issue you were involved with"},
		// Same session as implementer should fail with "own implementation" message
		{"same session", "session-1", "session-1", true, false, "cannot approve your own implementation"},
		// No implementer, not involved should pass
		{"no implementer not involved", "", "session-2", false, true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &TransitionContext{
				Issue: &models.Issue{
					ID:                 "test-1",
					Status:             models.StatusReview,
					ImplementerSession: tt.implementer,
				},
				FromStatus:  models.StatusReview,
				ToStatus:    models.StatusShipped,
				SessionID:   tt.sessionID,
				WasInvolved: tt.wasInvolved,
			}

			result := guard.Check(ctx)
			if result.Passed != tt.shouldPass {
				t.Errorf("passed=%v, want %v (msg: %s)", result.Passed, tt.shouldPass, result.Message)
			}
			if !result.Passed && tt.expectMessage != "" && result.Message != tt.expectMessage {
				t.Errorf("message=%q, want %q", result.Message, tt.expectMessage)
			}
		})
	}
}

func TestDifferentReviewerGuard_ActionContexts(t *testing.T) {
	guard := &DifferentReviewerGuard{}

	contexts := []ActionContext{
		ContextCLI,
		ContextMonitor,
		ContextWorkSession,
		ContextAdmin,
	}

	for _, actionCtx := range contexts {
		t.Run(string(actionCtx), func(t *testing.T) {
			ctx := &TransitionContext{
				Issue: &models.Issue{
					ID:                 "test-1",
					Status:             models.StatusReview,
					ImplementerSession: "session-1",
				},
				FromStatus:  models.StatusReview,
				ToStatus:    models.StatusShipped,
				SessionID:   "session-1",
				Context:     actionCtx,
				WasInvolved: true,
			}

			result := guard.Check(ctx)

			// Only admin context should allow self-approval
			if actionCtx == ContextAdmin {
				if !result.Passed {
					t.Errorf("Admin context should bypass guard: %s", result.Message)
				}
			} else {
				if result.Passed {
					t.Errorf("Non-admin context %s should not bypass guard", actionCtx)
				}
			}
		})
	}
}

// =============================================================================
// EpicChildrenGuard Tests
// =============================================================================

func TestEpicChildrenGuard_Name(t *testing.T) {
	guard := &EpicChildrenGuard{}
	if guard.Name() != "EpicChildrenGuard" {
		t.Errorf("Name() = %q, want %q", guard.Name(), "EpicChildrenGuard")
	}
}

func TestEpicChildrenGuard_WithOpenChildren(t *testing.T) {
	guard := &EpicChildrenGuard{OpenChildCount: 5}

	ctx := &TransitionContext{
		Issue: &models.Issue{
			ID:   "epic-1",
			Type: models.TypeEpic,
		},
		FromStatus: models.StatusInFlight,
		ToStatus:   models.StatusShipped,
	}

	result := guard.Check(ctx)
	if result.Passed {
		t.Error("EpicChildrenGuard should fail when epic has open children")
	}
	if result.Message != "epic has open children" {
		t.Errorf("Unexpected message: %s", result.Message)
	}
}

func TestEpicChildrenGuard_WithoutOpenChildren(t *testing.T) {
	guard := &EpicChildrenGuard{OpenChildCount: 0}

	ctx := &TransitionContext{
		Issue: &models.Issue{
			ID:   "epic-1",
			Type: models.TypeEpic,
		},
		FromStatus: models.StatusInFlight,
		ToStatus:   models.StatusShipped,
	}

	result := guard.Check(ctx)
	if !result.Passed {
		t.Errorf("EpicChildrenGuard should pass when no open children: %s", result.Message)
	}
}

func TestEpicChildrenGuard_NonEpicTypes(t *testing.T) {
	guard := &EpicChildrenGuard{OpenChildCount: 10}

	types := []models.Type{
		models.TypeTask,
		models.TypeBug,
		models.TypeFeature,
	}

	for _, issueType := range types {
		t.Run(string(issueType), func(t *testing.T) {
			ctx := &TransitionContext{
				Issue: &models.Issue{
					ID:   "test-1",
					Type: issueType,
				},
				FromStatus: models.StatusInFlight,
				ToStatus:   models.StatusShipped,
			}

			result := guard.Check(ctx)
			if !result.Passed {
				t.Errorf("EpicChildrenGuard should pass for non-epic type %s: %s",
					issueType, result.Message)
			}
		})
	}
}

func TestEpicChildrenGuard_NonClosingTransition(t *testing.T) {
	guard := &EpicChildrenGuard{OpenChildCount: 5}

	// Test various non-closing transitions
	transitions := []struct {
		from models.Status
		to   models.Status
	}{
		{models.StatusBacklog, models.StatusInFlight},
		{models.StatusInFlight, models.StatusReview},
		{models.StatusInFlight, models.StatusCanceled},
		{models.StatusCanceled, models.StatusBacklog},
		{models.StatusReview, models.StatusInFlight},
		{models.StatusShipped, models.StatusBacklog}, // reopening
	}

	for _, tr := range transitions {
		t.Run(string(tr.from)+"→"+string(tr.to), func(t *testing.T) {
			ctx := &TransitionContext{
				Issue: &models.Issue{
					ID:   "epic-1",
					Type: models.TypeEpic,
				},
				FromStatus: tr.from,
				ToStatus:   tr.to,
			}

			result := guard.Check(ctx)
			if !result.Passed {
				t.Errorf("EpicChildrenGuard should pass for non-closing transition: %s", result.Message)
			}
		})
	}
}

func TestEpicChildrenGuard_OpenChildCountVariations(t *testing.T) {
	tests := []struct {
		name       string
		childCount int
		shouldPass bool
	}{
		{"zero children", 0, true},
		{"one child", 1, false},
		{"many children", 100, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			guard := &EpicChildrenGuard{OpenChildCount: tt.childCount}

			ctx := &TransitionContext{
				Issue: &models.Issue{
					ID:   "epic-1",
					Type: models.TypeEpic,
				},
				FromStatus: models.StatusInFlight,
				ToStatus:   models.StatusShipped,
			}

			result := guard.Check(ctx)
			if result.Passed != tt.shouldPass {
				t.Errorf("passed=%v, want %v (children=%d)", result.Passed, tt.shouldPass, tt.childCount)
			}
		})
	}
}

// =============================================================================
// SelfCloseGuard Tests
// =============================================================================

func TestSelfCloseGuard_Name(t *testing.T) {
	guard := &SelfCloseGuard{}
	if guard.Name() != "SelfCloseGuard" {
		t.Errorf("Name() = %q, want %q", guard.Name(), "SelfCloseGuard")
	}
}

func TestSelfCloseGuard_WithoutException(t *testing.T) {
	guard := &SelfCloseGuard{SelfCloseException: ""}

	ctx := &TransitionContext{
		Issue: &models.Issue{
			ID:                 "test-1",
			Status:             models.StatusReview,
			ImplementerSession: "session-1",
		},
		FromStatus:  models.StatusReview,
		ToStatus:    models.StatusShipped,
		SessionID:   "session-1",
		WasInvolved: true,
	}

	result := guard.Check(ctx)
	if result.Passed {
		t.Error("SelfCloseGuard should fail without exception reason")
	}
	if result.Message != "cannot close own implementation without --self-close-exception" {
		t.Errorf("Unexpected message: %s", result.Message)
	}
}

func TestSelfCloseGuard_WithException(t *testing.T) {
	guard := &SelfCloseGuard{SelfCloseException: "solo project, no other reviewers"}

	ctx := &TransitionContext{
		Issue: &models.Issue{
			ID:                 "test-1",
			Status:             models.StatusReview,
			ImplementerSession: "session-1",
		},
		FromStatus:  models.StatusReview,
		ToStatus:    models.StatusShipped,
		SessionID:   "session-1",
		WasInvolved: true,
	}

	result := guard.Check(ctx)
	if !result.Passed {
		t.Errorf("SelfCloseGuard should pass with exception reason: %s", result.Message)
	}
}

func TestSelfCloseGuard_MinorTask(t *testing.T) {
	guard := &SelfCloseGuard{SelfCloseException: ""}

	tests := []struct {
		name       string
		issueMinor bool
		ctxMinor   bool
		shouldPass bool
	}{
		{"issue minor", true, false, true},
		{"context minor", false, true, true},
		{"both minor", true, true, true},
		{"not minor", false, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &TransitionContext{
				Issue: &models.Issue{
					ID:                 "test-1",
					Status:             models.StatusReview,
					ImplementerSession: "session-1",
					Minor:              tt.issueMinor,
				},
				FromStatus:  models.StatusReview,
				ToStatus:    models.StatusShipped,
				SessionID:   "session-1",
				Minor:       tt.ctxMinor,
				WasInvolved: true,
			}

			result := guard.Check(ctx)
			if result.Passed != tt.shouldPass {
				t.Errorf("passed=%v, want %v (msg: %s)", result.Passed, tt.shouldPass, result.Message)
			}
		})
	}
}

func TestSelfCloseGuard_NotInvolved(t *testing.T) {
	guard := &SelfCloseGuard{SelfCloseException: ""}

	ctx := &TransitionContext{
		Issue: &models.Issue{
			ID:                 "test-1",
			Status:             models.StatusReview,
			ImplementerSession: "session-1",
		},
		FromStatus:  models.StatusReview,
		ToStatus:    models.StatusShipped,
		SessionID:   "session-2",
		WasInvolved: false, // Different reviewer not involved
	}

	result := guard.Check(ctx)
	if !result.Passed {
		t.Errorf("SelfCloseGuard should pass when not involved: %s", result.Message)
	}
}

func TestSelfCloseGuard_NonClosingTransition(t *testing.T) {
	guard := &SelfCloseGuard{SelfCloseException: ""}

	transitions := []struct {
		from models.Status
		to   models.Status
	}{
		{models.StatusBacklog, models.StatusInFlight},
		{models.StatusInFlight, models.StatusReview},
		{models.StatusInFlight, models.StatusCanceled},
		{models.StatusCanceled, models.StatusBacklog},
	}

	for _, tr := range transitions {
		t.Run(string(tr.from)+"→"+string(tr.to), func(t *testing.T) {
			ctx := &TransitionContext{
				Issue: &models.Issue{
					ID:                 "test-1",
					ImplementerSession: "session-1",
				},
				FromStatus:  tr.from,
				ToStatus:    tr.to,
				SessionID:   "session-1",
				WasInvolved: true,
			}

			result := guard.Check(ctx)
			if !result.Passed {
				t.Errorf("SelfCloseGuard should pass for non-closing transition: %s", result.Message)
			}
		})
	}
}

func TestSelfCloseGuard_WasInvolvedVariations(t *testing.T) {
	guard := &SelfCloseGuard{SelfCloseException: ""}

	tests := []struct {
		name          string
		implementer   string
		sessionID     string
		wasInvolved   bool
		shouldPass    bool
		expectMessage string
	}{
		// Not involved should pass
		{"not involved", "session-1", "session-2", false, true, ""},
		// Same session as implementer should fail
		{"same session", "session-1", "session-1", true, false, "cannot close own implementation without --self-close-exception"},
		// Involved but different session should fail with different message
		{"involved different session", "session-1", "session-2", true, false, "cannot close issue you were involved with"},
		// No implementer, not involved should pass
		{"no implementer not involved", "", "session-2", false, true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &TransitionContext{
				Issue: &models.Issue{
					ID:                 "test-1",
					ImplementerSession: tt.implementer,
				},
				FromStatus:  models.StatusReview,
				ToStatus:    models.StatusShipped,
				SessionID:   tt.sessionID,
				WasInvolved: tt.wasInvolved,
			}

			result := guard.Check(ctx)
			if result.Passed != tt.shouldPass {
				t.Errorf("passed=%v, want %v (msg: %s)", result.Passed, tt.shouldPass, result.Message)
			}
			if !result.Passed && tt.expectMessage != "" && result.Message != tt.expectMessage {
				t.Errorf("message=%q, want %q", result.Message, tt.expectMessage)
			}
		})
	}
}

// =============================================================================
// InProgressRequiredGuard Tests
// =============================================================================

func TestInProgressRequiredGuard_Name(t *testing.T) {
	guard := &InProgressRequiredGuard{}
	if guard.Name() != "InProgressRequiredGuard" {
		t.Errorf("Name() = %q, want %q", guard.Name(), "InProgressRequiredGuard")
	}
}

func TestInProgressRequiredGuard_FromInProgress(t *testing.T) {
	guard := &InProgressRequiredGuard{}

	ctx := &TransitionContext{
		Issue:      &models.Issue{ID: "test-1"},
		FromStatus: models.StatusInFlight,
		ToStatus:   models.StatusReview,
	}

	result := guard.Check(ctx)
	if !result.Passed {
		t.Errorf("InProgressRequiredGuard should pass from in_progress: %s", result.Message)
	}
}

func TestInProgressRequiredGuard_FromOpen(t *testing.T) {
	guard := &InProgressRequiredGuard{}

	ctx := &TransitionContext{
		Issue:      &models.Issue{ID: "test-1"},
		FromStatus: models.StatusBacklog,
		ToStatus:   models.StatusReview,
	}

	result := guard.Check(ctx)
	if !result.Passed {
		t.Errorf("InProgressRequiredGuard should pass from open (direct submission): %s", result.Message)
	}
}

func TestInProgressRequiredGuard_VariousSourceStatuses(t *testing.T) {
	guard := &InProgressRequiredGuard{}

	tests := []struct {
		name       string
		fromStatus models.Status
		toStatus   models.Status
		shouldPass bool
	}{
		// To in_review
		{"open→in_review", models.StatusBacklog, models.StatusReview, true},
		{"in_progress→in_review", models.StatusInFlight, models.StatusReview, true},
		{"blocked→in_review", models.StatusCanceled, models.StatusReview, false},
		{"closed→in_review", models.StatusShipped, models.StatusReview, false},

		// Non-review transitions should always pass
		{"open→in_progress", models.StatusBacklog, models.StatusInFlight, true},
		{"open→blocked", models.StatusBacklog, models.StatusCanceled, true},
		{"open→closed", models.StatusBacklog, models.StatusShipped, true},
		{"in_progress→blocked", models.StatusInFlight, models.StatusCanceled, true},
		{"in_progress→closed", models.StatusInFlight, models.StatusShipped, true},
		{"blocked→open", models.StatusCanceled, models.StatusBacklog, true},
		{"blocked→in_progress", models.StatusCanceled, models.StatusInFlight, true},
		{"in_review→closed", models.StatusReview, models.StatusShipped, true},
		{"in_review→open", models.StatusReview, models.StatusBacklog, true},
		{"closed→open", models.StatusShipped, models.StatusBacklog, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &TransitionContext{
				Issue:      &models.Issue{ID: "test-1"},
				FromStatus: tt.fromStatus,
				ToStatus:   tt.toStatus,
			}

			result := guard.Check(ctx)
			if result.Passed != tt.shouldPass {
				t.Errorf("InProgressRequiredGuard.Check() passed=%v, want %v (msg: %s)",
					result.Passed, tt.shouldPass, result.Message)
			}
		})
	}
}

func TestInProgressRequiredGuard_ErrorMessage(t *testing.T) {
	guard := &InProgressRequiredGuard{}

	ctx := &TransitionContext{
		Issue:      &models.Issue{ID: "test-1"},
		FromStatus: models.StatusCanceled,
		ToStatus:   models.StatusReview,
	}

	result := guard.Check(ctx)
	if result.Passed {
		t.Error("Should fail for blocked→in_review")
	}
	expectedMsg := "can only submit for review from backlog or in_flight status"
	if result.Message != expectedMsg {
		t.Errorf("message=%q, want %q", result.Message, expectedMsg)
	}
}

// =============================================================================
// Guard Integration Tests
// =============================================================================

func TestGuardsIntegration_StrictMode(t *testing.T) {
	sm := StrictMachine()

	// Test that DifferentReviewerGuard is wired up for review→shipped
	ctx := &TransitionContext{
		Issue: &models.Issue{
			ID:                 "test-1",
			Status:             models.StatusReview,
			ImplementerSession: "session-1",
		},
		FromStatus:  models.StatusReview,
		ToStatus:    models.StatusShipped,
		SessionID:   "session-1",
		WasInvolved: true,
	}

	_, err := sm.Validate(ctx)
	if err == nil {
		t.Error("Strict mode should block self-approval of review→shipped")
	}

	// Different reviewer should pass
	ctx.SessionID = "session-2"
	ctx.WasInvolved = false
	_, err = sm.Validate(ctx)
	if err != nil {
		t.Errorf("Strict mode should allow review→shipped with different reviewer: %v", err)
	}
}

func TestGuardsIntegration_DifferentReviewerInStrict(t *testing.T) {
	sm := StrictMachine()

	ctx := &TransitionContext{
		Issue: &models.Issue{
			ID:                 "test-1",
			Status:             models.StatusReview,
			ImplementerSession: "session-1",
		},
		FromStatus:  models.StatusReview,
		ToStatus:    models.StatusShipped,
		SessionID:   "session-1",
		WasInvolved: true,
	}

	_, err := sm.Validate(ctx)
	if err == nil {
		t.Error("Strict mode should block self-approval")
	}

	// Different reviewer should pass
	ctx.SessionID = "session-2"
	ctx.WasInvolved = false
	_, err = sm.Validate(ctx)
	if err != nil {
		t.Errorf("Strict mode should allow different reviewer: %v", err)
	}
}

func TestGuardsIntegration_LiberalModeSkipsGuards(t *testing.T) {
	sm := DefaultMachine()

	// In liberal mode, even guard-failing scenarios should pass
	ctx := &TransitionContext{
		Issue: &models.Issue{
			ID:                 "test-1",
			Status:             models.StatusReview,
			ImplementerSession: "session-1",
		},
		FromStatus:  models.StatusReview,
		ToStatus:    models.StatusShipped,
		SessionID:   "session-1",
		WasInvolved: true, // Would fail DifferentReviewerGuard
	}

	results, err := sm.Validate(ctx)
	if err != nil {
		t.Errorf("Liberal mode should allow transition: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Liberal mode should not run guards, got %d results", len(results))
	}
}

func TestGuardsIntegration_AdvisoryModeReturnsWarnings(t *testing.T) {
	sm := AdvisoryMachine()

	ctx := &TransitionContext{
		Issue: &models.Issue{
			ID:                 "test-1",
			Status:             models.StatusReview,
			ImplementerSession: "session-1",
		},
		FromStatus:  models.StatusReview,
		ToStatus:    models.StatusShipped,
		SessionID:   "session-1",
		WasInvolved: true,
	}

	results, err := sm.Validate(ctx)
	if err != nil {
		t.Errorf("Advisory mode should allow transition: %v", err)
	}
	if len(results) == 0 {
		t.Error("Advisory mode should return guard results")
	}

	// Should have DifferentReviewerGuard failure
	found := false
	for _, r := range results {
		if r.Guard == "DifferentReviewerGuard" && !r.Passed {
			found = true
			break
		}
	}
	if !found {
		t.Error("Advisory mode should report DifferentReviewerGuard failure")
	}
}
