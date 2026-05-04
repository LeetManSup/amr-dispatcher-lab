package application

import (
	"fmt"

	"amr-dispatcher-lab/internal/domain"
	"amr-dispatcher-lab/internal/simulation"
)

func (s *Service) persistRuntime(engine *simulation.Engine, result simulation.TickResult) error {
	if err := s.persistStateOnly(engine); err != nil {
		return err
	}
	if len(result.StatusHistory) > 0 {
		if err := s.store.AppendTaskHistory(result.StatusHistory); err != nil {
			return err
		}
	}
	if len(result.Decisions) > 0 {
		if err := s.store.AppendDecisionLogs(result.Decisions); err != nil {
			return err
		}
	}
	if err := s.store.AppendMetrics([]domain.MetricsSnapshot{result.Metrics}); err != nil {
		return err
	}
	if len(result.SegmentLoads) > 0 {
		if err := s.store.AppendSegmentLoads(result.SegmentLoads); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) persistStateOnly(engine *simulation.Engine) error {
	run := engine.Run()
	if err := s.store.UpdateRun(run); err != nil {
		return err
	}
	if err := s.store.UpsertRequests(run.ID, engine.Requests()); err != nil {
		return err
	}
	if err := s.store.UpsertTasks(run.ID, engine.Tasks()); err != nil {
		return err
	}
	return s.store.UpsertRobots(run.ID, engine.Robots())
}

func (s *Service) runtime(runID string) (*runtime, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rt, ok := s.runs[runID]
	if !ok {
		return nil, fmt.Errorf("run %s not loaded", runID)
	}
	return rt, nil
}

func (s *Service) runtimeIfLoaded(runID string) (*runtime, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rt, ok := s.runs[runID]
	return rt, ok
}
