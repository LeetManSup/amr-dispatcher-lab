package scheduler

import (
	"amr-dispatcher-lab/internal/domain"
	"sort"
)

type Scheduler interface {
	Algorithm() domain.Algorithm
	Decide(tick int, snapshot domain.SystemSnapshot, candidates []domain.TaskCandidate) []domain.Decision
}

func New(algorithm domain.Algorithm, weights domain.SchedulerWeights) Scheduler {
	switch algorithm {
	case domain.AlgorithmPriority:
		return PriorityScheduler{weights: weights}
	case domain.AlgorithmAdaptive:
		return AdaptiveScheduler{weights: weights}
	default:
		return FIFOScheduler{}
	}
}

type FIFOScheduler struct{}

func (FIFOScheduler) Algorithm() domain.Algorithm { return domain.AlgorithmFIFO }

func (FIFOScheduler) Decide(_ int, _ domain.SystemSnapshot, candidates []domain.TaskCandidate) []domain.Decision {
	ordered := append([]domain.TaskCandidate(nil), candidates...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].Request.CreatedAt != ordered[j].Request.CreatedAt {
			return ordered[i].Request.CreatedAt < ordered[j].Request.CreatedAt
		}
		return ordered[i].Task.TaskID < ordered[j].Task.TaskID
	})
	return toDecisions(ordered, func(_ domain.TaskCandidate) float64 { return 0 })
}

type PriorityScheduler struct {
	weights domain.SchedulerWeights
}

func (PriorityScheduler) Algorithm() domain.Algorithm { return domain.AlgorithmPriority }

func (s PriorityScheduler) Decide(_ int, _ domain.SystemSnapshot, candidates []domain.TaskCandidate) []domain.Decision {
	ordered := append([]domain.TaskCandidate(nil), candidates...)
	sort.SliceStable(ordered, func(i, j int) bool {
		si := s.score(ordered[i])
		sj := s.score(ordered[j])
		if si == sj {
			return ordered[i].Request.CreatedAt < ordered[j].Request.CreatedAt
		}
		return si > sj
	})
	return toDecisions(ordered, s.score)
}

func (s PriorityScheduler) score(candidate domain.TaskCandidate) float64 {
	alpha := fallback(s.weights.Alpha, 0.65)
	beta := fallback(s.weights.Beta, 0.35)
	return alpha*candidate.PriorityScore + beta*candidate.Urgency
}

type AdaptiveScheduler struct {
	weights domain.SchedulerWeights
}

func (AdaptiveScheduler) Algorithm() domain.Algorithm { return domain.AlgorithmAdaptive }

func (s AdaptiveScheduler) Decide(_ int, _ domain.SystemSnapshot, candidates []domain.TaskCandidate) []domain.Decision {
	ordered := append([]domain.TaskCandidate(nil), candidates...)
	sort.SliceStable(ordered, func(i, j int) bool {
		si := s.score(ordered[i])
		sj := s.score(ordered[j])
		if si == sj {
			return ordered[i].Request.CreatedAt < ordered[j].Request.CreatedAt
		}
		return si > sj
	})
	return toDecisions(ordered, s.score)
}

func (s AdaptiveScheduler) score(candidate domain.TaskCandidate) float64 {
	loadPressure := 0.25 + 0.75*adaptivePenalty(candidate.SystemRisk)
	return fallback(s.weights.W1, 0.35)*candidate.PriorityScore +
		fallback(s.weights.W2, 0.35)*candidate.Urgency -
		fallback(s.weights.W3, 0.15)*loadPressure*adaptivePenalty(candidate.RouteLoad) -
		fallback(s.weights.W4, 0.10)*loadPressure*adaptivePenalty(candidate.ZoneLoad) -
		fallback(s.weights.W5, 0.05)*candidate.SystemRisk
}

func adaptivePenalty(value float64) float64 {
	if value <= 0 {
		return 0
	}
	if value >= 1 {
		return 1
	}
	return value
}

func toDecisions(candidates []domain.TaskCandidate, scoreFn func(domain.TaskCandidate) float64) []domain.Decision {
	decisions := make([]domain.Decision, 0, len(candidates))
	for _, candidate := range candidates {
		decisions = append(decisions, domain.Decision{
			Action:     domain.DecisionActionDispatch,
			RequestID:  candidate.Request.RequestID,
			TaskID:     candidate.Task.TaskID,
			Score:      scoreFn(candidate),
			ReasonCode: "ranked_for_dispatch",
		})
	}
	return decisions
}

func fallback(value, defaultValue float64) float64 {
	if value == 0 {
		return defaultValue
	}
	return value
}
