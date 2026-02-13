package workflow

import (
	"github.com/marcus/td/internal/models"
)

// AllTransitions returns all valid status transitions.
// This defines the complete workflow state machine:
//
//	Forward flow: Triage → Backlog → Prioritized → InFlight → Review → Shipped
//	Terminal states: Canceled and Duplicate (reachable from most active states)
//	Backward transitions: Allowed for corrections (demote, reject, reopen)
func AllTransitions() []*Transition {
	return []*Transition{
		// From triage — accept into backlog, or discard
		{From: models.StatusTriage, To: models.StatusBacklog, Guards: nil},
		{From: models.StatusTriage, To: models.StatusCanceled, Guards: nil},
		{From: models.StatusTriage, To: models.StatusDuplicate, Guards: nil},

		// From backlog — promote forward, or discard
		{From: models.StatusBacklog, To: models.StatusTriage, Guards: nil},
		{From: models.StatusBacklog, To: models.StatusPrioritized, Guards: nil},
		{From: models.StatusBacklog, To: models.StatusInFlight, Guards: nil},
		{From: models.StatusBacklog, To: models.StatusCanceled, Guards: nil},
		{From: models.StatusBacklog, To: models.StatusDuplicate, Guards: nil},

		// From prioritized — start work, demote back, or discard
		{From: models.StatusPrioritized, To: models.StatusBacklog, Guards: nil},
		{From: models.StatusPrioritized, To: models.StatusInFlight, Guards: nil},
		{From: models.StatusPrioritized, To: models.StatusCanceled, Guards: nil},
		{From: models.StatusPrioritized, To: models.StatusDuplicate, Guards: nil},

		// From in_flight — submit for review, ship directly, pause, or discard
		{From: models.StatusInFlight, To: models.StatusReview, Guards: nil},
		{From: models.StatusInFlight, To: models.StatusShipped, Guards: nil},
		{From: models.StatusInFlight, To: models.StatusBacklog, Guards: nil},
		{From: models.StatusInFlight, To: models.StatusCanceled, Guards: nil},

		// From review — approve, reject back, or discard
		{From: models.StatusReview, To: models.StatusShipped, Guards: []Guard{&DifferentReviewerGuard{}}},
		{From: models.StatusReview, To: models.StatusInFlight, Guards: nil},
		{From: models.StatusReview, To: models.StatusBacklog, Guards: nil},
		{From: models.StatusReview, To: models.StatusCanceled, Guards: nil},

		// From shipped — reopen only
		{From: models.StatusShipped, To: models.StatusBacklog, Guards: nil},

		// From canceled — reopen to backlog or triage
		{From: models.StatusCanceled, To: models.StatusBacklog, Guards: nil},
		{From: models.StatusCanceled, To: models.StatusTriage, Guards: nil},

		// From duplicate — reopen to backlog or triage
		{From: models.StatusDuplicate, To: models.StatusBacklog, Guards: nil},
		{From: models.StatusDuplicate, To: models.StatusTriage, Guards: nil},
	}
}

// TransitionName returns a human-readable name for the transition
func TransitionName(from, to models.Status) string {
	switch {
	// Starting work
	case from == models.StatusBacklog && to == models.StatusInFlight:
		return "start"
	case from == models.StatusPrioritized && to == models.StatusInFlight:
		return "start"
	case from == models.StatusInFlight && to == models.StatusBacklog:
		return "unstart"

	// Terminal states
	case to == models.StatusCanceled:
		return "cancel"
	case to == models.StatusDuplicate:
		return "duplicate"

	// Review flow
	case to == models.StatusReview:
		return "review"
	case from == models.StatusReview && to == models.StatusInFlight:
		return "reject"
	case from == models.StatusReview && to == models.StatusShipped:
		return "approve"
	case from == models.StatusReview && to == models.StatusBacklog:
		return "reject"

	// Shipping
	case to == models.StatusShipped:
		return "ship"

	// Reopening from terminal/shipped states
	case from == models.StatusShipped && to == models.StatusBacklog:
		return "reopen"
	case from == models.StatusCanceled && to == models.StatusBacklog:
		return "reopen"
	case from == models.StatusCanceled && to == models.StatusTriage:
		return "reopen"
	case from == models.StatusDuplicate && to == models.StatusBacklog:
		return "reopen"
	case from == models.StatusDuplicate && to == models.StatusTriage:
		return "reopen"

	// Prioritization
	case to == models.StatusPrioritized:
		return "prioritize"
	case to == models.StatusBacklog:
		return "deprioritize"
	case to == models.StatusTriage:
		return "triage"

	default:
		return string(from) + " → " + string(to)
	}
}

// GetTransitionsFrom returns all possible transitions from a given status
func GetTransitionsFrom(status models.Status) []models.Status {
	var targets []models.Status
	for _, t := range AllTransitions() {
		if t.From == status {
			targets = append(targets, t.To)
		}
	}
	return targets
}

// GetTransitionsTo returns all statuses that can transition to the given status
func GetTransitionsTo(status models.Status) []models.Status {
	var sources []models.Status
	for _, t := range AllTransitions() {
		if t.To == status {
			sources = append(sources, t.From)
		}
	}
	return sources
}

// AllStatuses returns all valid statuses in workflow order
func AllStatuses() []models.Status {
	return []models.Status{
		models.StatusTriage,
		models.StatusBacklog,
		models.StatusPrioritized,
		models.StatusInFlight,
		models.StatusReview,
		models.StatusShipped,
		models.StatusCanceled,
		models.StatusDuplicate,
	}
}
