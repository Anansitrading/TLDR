// Package workflow implements the issue status state machine.
//
// Guards are condition checks that run during status transitions in Advisory
// and Strict modes. In Liberal mode (default), guards are skipped.
//
// Currently active guards (attached to transitions):
//   - DifferentReviewerGuard: Prevents self-approval
//
// Future guards (defined but not yet attached to transitions):
//   - BlockedGuard: Requires --force to start canceled issues
//   - EpicChildrenGuard: Warns when shipping epic with open children
//   - SelfCloseGuard: Prevents self-shipping without exception
//   - InProgressRequiredGuard: Validates review source status
//
// These future guards require caller modifications to pass necessary context
// (e.g., open child count, self-close exception reason) and will be wired
// up when Advisory/Strict modes are enabled by default.
package workflow

import (
	"github.com/marcus/td/internal/models"
)

// BlockedGuard warns when starting a canceled issue without --force.
// Active: Attached to canceled → in_flight transition.
type BlockedGuard struct{}

func (g *BlockedGuard) Name() string {
	return "BlockedGuard"
}

func (g *BlockedGuard) Check(ctx *TransitionContext) GuardResult {
	// Allow if force flag is set
	if ctx.Force {
		return GuardResult{Passed: true}
	}

	// Allow if not transitioning from canceled
	if ctx.FromStatus != models.StatusCanceled {
		return GuardResult{Passed: true}
	}

	// Fail - starting canceled issue without --force
	return GuardResult{
		Passed:  false,
		Message: "cannot start canceled issue without --force",
	}
}

// DifferentReviewerGuard ensures approvals come from different session than implementer.
// Active: Attached to review → shipped transition.
type DifferentReviewerGuard struct{}

func (g *DifferentReviewerGuard) Name() string {
	return "DifferentReviewerGuard"
}

func (g *DifferentReviewerGuard) Check(ctx *TransitionContext) GuardResult {
	// Allow if minor task (self-approval permitted)
	if ctx.Minor || ctx.Issue.Minor {
		return GuardResult{Passed: true}
	}

	// Allow if admin bypass context
	if ctx.Context == ContextAdmin {
		return GuardResult{Passed: true}
	}

	// Allow if session was not involved
	if !ctx.WasInvolved {
		return GuardResult{Passed: true}
	}

	// Check if same session as implementer
	if ctx.Issue.ImplementerSession != "" && ctx.Issue.ImplementerSession == ctx.SessionID {
		return GuardResult{
			Passed:  false,
			Message: "cannot approve your own implementation",
		}
	}

	// Check if session was involved (via WasInvolved flag)
	// The caller sets this based on database.WasSessionInvolved() check
	if ctx.WasInvolved {
		return GuardResult{
			Passed:  false,
			Message: "cannot approve issue you were involved with",
		}
	}

	return GuardResult{Passed: true}
}

// EpicChildrenGuard warns when closing epic with open children.
// Future: Not yet attached to transitions. Requires caller to set OpenChildCount.
type EpicChildrenGuard struct {
	// OpenChildCount is set by caller before validation
	OpenChildCount int
}

func (g *EpicChildrenGuard) Name() string {
	return "EpicChildrenGuard"
}

func (g *EpicChildrenGuard) Check(ctx *TransitionContext) GuardResult {
	// Only applies to closing transitions
	if ctx.ToStatus != models.StatusShipped {
		return GuardResult{Passed: true}
	}

	// Only applies to epics
	if ctx.Issue.Type != models.TypeEpic {
		return GuardResult{Passed: true}
	}

	// Check if there are open children
	if g.OpenChildCount > 0 {
		return GuardResult{
			Passed:  false,
			Message: "epic has open children",
		}
	}

	return GuardResult{Passed: true}
}

// SelfCloseGuard prevents closing issues you implemented without exception.
// Future: Not yet attached to transitions. Requires caller to set SelfCloseException.
type SelfCloseGuard struct {
	// SelfCloseException is the reason provided for self-close bypass
	SelfCloseException string
}

func (g *SelfCloseGuard) Name() string {
	return "SelfCloseGuard"
}

func (g *SelfCloseGuard) Check(ctx *TransitionContext) GuardResult {
	// Allow if not closing
	if ctx.ToStatus != models.StatusShipped {
		return GuardResult{Passed: true}
	}

	// Allow if minor task
	if ctx.Minor || ctx.Issue.Minor {
		return GuardResult{Passed: true}
	}

	// Allow if self-close exception provided
	if g.SelfCloseException != "" {
		return GuardResult{Passed: true}
	}

	// Allow if not involved
	if !ctx.WasInvolved {
		return GuardResult{Passed: true}
	}

	// Check if session is the implementer
	if ctx.Issue.ImplementerSession != "" && ctx.Issue.ImplementerSession == ctx.SessionID {
		return GuardResult{
			Passed:  false,
			Message: "cannot close own implementation without --self-close-exception",
		}
	}

	// General involvement check
	if ctx.WasInvolved {
		return GuardResult{
			Passed:  false,
			Message: "cannot close issue you were involved with",
		}
	}

	return GuardResult{Passed: true}
}

// InProgressRequiredGuard ensures issue is in flight before review.
// Future: Not yet attached to transitions. Transition definitions already
// prevent most invalid paths.
type InProgressRequiredGuard struct{}

func (g *InProgressRequiredGuard) Name() string {
	return "InProgressRequiredGuard"
}

func (g *InProgressRequiredGuard) Check(ctx *TransitionContext) GuardResult {
	// Only applies when going to review
	if ctx.ToStatus != models.StatusReview {
		return GuardResult{Passed: true}
	}

	// Allow from in_progress
	if ctx.FromStatus == models.StatusInFlight {
		return GuardResult{Passed: true}
	}

	// Allow from open (direct submission)
	if ctx.FromStatus == models.StatusBacklog {
		return GuardResult{Passed: true}
	}

	return GuardResult{
		Passed:  false,
		Message: "can only submit for review from backlog or in_flight status",
	}
}
