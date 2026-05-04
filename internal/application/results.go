package application

import (
	"amr-dispatcher-lab/internal/domain"
	"amr-dispatcher-lab/internal/simulation"
)

type StepRunResult struct {
	Snapshot      domain.SystemSnapshot
	Metrics       domain.MetricsSnapshot
	Decisions     []domain.DecisionLog
	StatusHistory []domain.TaskStatusHistory
	SegmentLoads  []domain.SegmentLoadSnapshot
}

func toStepRunResult(item simulation.TickResult) StepRunResult {
	return StepRunResult{
		Snapshot:      item.Snapshot,
		Metrics:       item.Metrics,
		Decisions:     append([]domain.DecisionLog(nil), item.Decisions...),
		StatusHistory: append([]domain.TaskStatusHistory(nil), item.StatusHistory...),
		SegmentLoads:  append([]domain.SegmentLoadSnapshot(nil), item.SegmentLoads...),
	}
}
