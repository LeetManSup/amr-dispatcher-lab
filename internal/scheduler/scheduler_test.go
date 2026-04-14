package scheduler

import (
	"amr-dispatcher-lab/internal/domain"
	"testing"
)

func TestPrioritySchedulerRanksByPriorityAndUrgency(t *testing.T) {
	s := New(domain.AlgorithmPriority, domain.SchedulerWeights{Alpha: 0.7, Beta: 0.3})
	candidates := []domain.TaskCandidate{
		{
			Request:       &domain.Request{RequestID: "low", CreatedAt: 0},
			Task:          &domain.TaskOrder{TaskID: "task-low"},
			PriorityScore: 0.2,
			Urgency:       0.1,
		},
		{
			Request:       &domain.Request{RequestID: "high", CreatedAt: 1},
			Task:          &domain.TaskOrder{TaskID: "task-high"},
			PriorityScore: 0.9,
			Urgency:       0.7,
		},
	}

	decisions := s.Decide(0, domain.SystemSnapshot{}, candidates)
	if decisions[0].TaskID != "task-high" {
		t.Fatalf("expected high priority task first, got %s", decisions[0].TaskID)
	}
}

func TestDeferredTaskIsRankedAgainWhenItsDeferTickHasElapsed(t *testing.T) {
	s := New(domain.AlgorithmFIFO, domain.SchedulerWeights{})
	candidates := []domain.TaskCandidate{
		{
			Request: &domain.Request{RequestID: "ready", CreatedAt: 0},
			Task: &domain.TaskOrder{
				TaskID:         "task-ready",
				DeferUntilTick: 1,
			},
		},
	}

	decisions := s.Decide(1, domain.SystemSnapshot{}, candidates)
	if len(decisions) != 1 {
		t.Fatalf("expected exactly one decision, got %d", len(decisions))
	}
	if decisions[0].Action != domain.DecisionActionDispatch {
		t.Fatalf("expected deferred task to be reconsidered for dispatch, got %s", decisions[0].Action)
	}
	if decisions[0].ReasonCode != "ranked_for_dispatch" {
		t.Fatalf("expected ranked_for_dispatch reason, got %s", decisions[0].ReasonCode)
	}
}
