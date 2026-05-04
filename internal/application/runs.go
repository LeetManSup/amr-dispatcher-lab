package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"amr-dispatcher-lab/internal/domain"
	"amr-dispatcher-lab/internal/simulation"
)

func (s *Service) Catalog() Catalog {
	maps, _ := s.assets.ListMaps()
	scenarios, _ := s.assets.ListScenarios()
	return Catalog{
		Maps:      maps,
		Scenarios: scenarios,
		MapItems: buildCatalogItems(maps, "map", func(path string) (string, string, error) {
			item, err := s.assets.LoadMap(path)
			if err != nil {
				return "", "", err
			}
			description := fmt.Sprintf("%d points, %d segments", len(item.Points), len(item.Segments))
			return pickFirstNonEmpty(item.Name, humanizePathLabel(path)), description, nil
		}),
		ScenarioItems: buildCatalogItems(scenarios, "scenario", func(path string) (string, string, error) {
			item, err := s.assets.LoadScenario(path)
			if err != nil {
				return "", "", err
			}
			return pickFirstNonEmpty(item.Name, humanizePathLabel(path)), item.Description, nil
		}),
	}
}

func (s *Service) CreateRun(input CreateRunInput) (domain.Run, error) {
	mapPath := pickFirstNonEmpty(input.MapPath, s.options.DefaultMapPath)
	scenarioPath := pickFirstNonEmpty(input.ScenarioPath, s.options.DefaultScenarioPath)
	if input.Algorithm == "" {
		input.Algorithm = domain.AlgorithmFIFO
	}
	if input.Seed == 0 {
		input.Seed = s.clock.Now().UnixNano()
	}

	mapData, err := s.assets.LoadMap(mapPath)
	if err != nil {
		s.logger.Warn("load map failed", "map_path", mapPath, "err", err)
		return domain.Run{}, err
	}
	scenario, err := s.assets.LoadScenario(scenarioPath)
	if err != nil {
		s.logger.Warn("load scenario failed", "scenario_path", scenarioPath, "err", err)
		return domain.Run{}, err
	}
	if err := validateRunConfiguration(mapData, scenario); err != nil {
		s.logger.Warn("run validation failed", "map_path", mapPath, "scenario_path", scenarioPath, "err", err)
		return domain.Run{}, err
	}
	runID := s.nextRunID()
	engine, err := simulation.NewEngine(runID, domain.RunConfig{
		Algorithm:          input.Algorithm,
		MapPath:            mapPath,
		ScenarioPath:       scenarioPath,
		Seed:               input.Seed,
		TickDurationMillis: s.options.TickDurationMillis,
		PersistenceEnabled: true,
		ExportFormats:      []string{"json", "csv"},
	}, mapData, scenario, s.clock, nil)
	if err != nil {
		s.logger.Error("create engine failed", "run_id", runID, "algorithm", input.Algorithm, "err", err)
		return domain.Run{}, err
	}
	rt := &runtime{engine: engine}

	s.mu.Lock()
	s.runs[runID] = rt
	s.mu.Unlock()

	run := engine.Run()
	if err := s.store.SaveRun(run); err != nil {
		s.logger.Error("save run failed", "run_id", run.ID, "err", err)
		return domain.Run{}, err
	}
	if err := s.store.UpsertRobots(run.ID, engine.Robots()); err != nil {
		s.logger.Error("persist robots failed", "run_id", run.ID, "err", err)
		return domain.Run{}, err
	}
	if err := s.store.UpsertTasks(run.ID, engine.Tasks()); err != nil {
		s.logger.Error("persist tasks failed", "run_id", run.ID, "err", err)
		return domain.Run{}, err
	}
	if err := s.store.UpsertRequests(run.ID, engine.Requests()); err != nil {
		s.logger.Error("persist requests failed", "run_id", run.ID, "err", err)
		return domain.Run{}, err
	}
	s.logger.Info("run created",
		"run_id", run.ID,
		"algorithm", run.Algorithm,
		"map_path", mapPath,
		"scenario_path", scenarioPath,
		"seed", input.Seed,
	)
	s.publish(LiveEvent{
		RunID:        run.ID,
		Kind:         "run.created",
		Message:      "Run created",
		CurrentTick:  run.CurrentTick,
		RunStatus:    string(run.Status),
		ActiveTasks:  0,
		WaitingTasks: len(engine.Tasks()),
	})
	return run, nil
}

func (s *Service) StepRun(runID string) (StepRunResult, domain.Run, error) {
	rt, err := s.runtime(runID)
	if err != nil {
		s.logger.Warn("step run rejected", "run_id", runID, "err", err)
		return StepRunResult{}, domain.Run{}, err
	}
	logger := s.logger.With("run_id", runID)
	rt.mu.Lock()

	result, err := rt.engine.Step()
	if err != nil {
		rt.mu.Unlock()
		logger.Error("step failed", "err", err)
		return StepRunResult{}, domain.Run{}, err
	}
	if err := s.persistRuntime(rt.engine, result); err != nil {
		rt.mu.Unlock()
		logger.Error("persist runtime failed", "tick", result.Snapshot.Tick, "err", err)
		return StepRunResult{}, domain.Run{}, err
	}
	run := rt.engine.Run()
	event := LiveEvent{
		RunID:        run.ID,
		Kind:         "run.step",
		Message:      "Tick processed",
		CurrentTick:  result.Snapshot.Tick,
		RunStatus:    string(run.Status),
		ActiveTasks:  result.Snapshot.ActiveTasks,
		WaitingTasks: result.Snapshot.WaitingTasks,
	}
	rt.mu.Unlock()
	logger.Log(context.Background(), slog.LevelDebug, "tick processed",
		"tick", result.Snapshot.Tick,
		"active_tasks", result.Snapshot.ActiveTasks,
		"waiting_tasks", result.Snapshot.WaitingTasks,
		"decisions", len(result.Decisions),
	)
	s.publish(event)
	return toStepRunResult(result), run, nil
}

func (s *Service) StartRun(runID string) error {
	rt, err := s.runtime(runID)
	if err != nil {
		s.logger.Warn("start run rejected", "run_id", runID, "err", err)
		return err
	}
	logger := s.logger.With("run_id", runID)

	s.mu.Lock()
	if s.activeRunID != "" && s.activeRunID != runID {
		s.mu.Unlock()
		logger.Warn("continuous execution rejected", "active_run_id", s.activeRunID)
		return fmt.Errorf("run %s is already executing continuously", s.activeRunID)
	}
	s.activeRunID = runID
	s.mu.Unlock()

	rt.mu.Lock()
	if rt.cancel != nil {
		rt.mu.Unlock()
		logger.Log(context.Background(), slog.LevelDebug, "continuous execution already active")
		return nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	rt.cancel = cancel
	rt.engine.SetMode(domain.RunModeContinuous)
	run := rt.engine.Run()
	snapshot := rt.engine.CurrentSnapshot()
	_ = s.store.UpdateRun(run)
	rt.mu.Unlock()
	logger.Info("continuous execution started", "tick", snapshot.Tick, "status", run.Status)
	s.publish(LiveEvent{
		RunID:        run.ID,
		Kind:         "run.started",
		Message:      "Continuous execution started",
		CurrentTick:  snapshot.Tick,
		RunStatus:    string(run.Status),
		ActiveTasks:  snapshot.ActiveTasks,
		WaitingTasks: snapshot.WaitingTasks,
	})

	go func() {
		ticker := time.NewTicker(time.Duration(maxInt(1, s.options.TickDurationMillis)) * time.Millisecond)
		defer ticker.Stop()
		defer func() {
			rt.mu.Lock()
			finalRun := rt.engine.Run()
			rt.cancel = nil
			rt.engine.SetMode(domain.RunModeStep)
			_ = s.store.UpdateRun(rt.engine.Run())
			rt.mu.Unlock()
			s.mu.Lock()
			if s.activeRunID == runID {
				s.activeRunID = ""
			}
			s.mu.Unlock()
			logger.Info("continuous execution stopped", "final_status", finalRun.Status, "tick", finalRun.CurrentTick)
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				rt.mu.Lock()
				result, err := rt.engine.Step()
				if err == nil {
					err = s.persistRuntime(rt.engine, result)
				}
				run := rt.engine.Run()
				event := LiveEvent{
					RunID:        run.ID,
					Kind:         "run.step",
					Message:      "Tick processed",
					CurrentTick:  result.Snapshot.Tick,
					RunStatus:    string(run.Status),
					ActiveTasks:  result.Snapshot.ActiveTasks,
					WaitingTasks: result.Snapshot.WaitingTasks,
				}
				shouldStop := err != nil || run.Status == domain.RunStatusCompleted || run.Status == domain.RunStatusFailed || run.Status == domain.RunStatusStopped
				rt.mu.Unlock()
				if err == nil {
					s.publish(event)
				} else {
					logger.Error("continuous step failed", "tick", result.Snapshot.Tick, "err", err)
				}
				if shouldStop {
					return
				}
			}
		}
	}()
	return nil
}

func (s *Service) StopRun(runID string) error {
	rt, err := s.runtime(runID)
	if err != nil {
		s.logger.Warn("stop run rejected", "run_id", runID, "err", err)
		return err
	}
	logger := s.logger.With("run_id", runID)
	rt.mu.Lock()
	if rt.cancel != nil {
		rt.cancel()
		rt.cancel = nil
	}
	rt.engine.Stop()
	run := rt.engine.Run()
	snapshot := rt.engine.CurrentSnapshot()
	err = s.store.UpdateRun(run)
	rt.mu.Unlock()
	if err != nil {
		logger.Error("persist stopped run failed", "err", err)
		return err
	}
	logger.Info("run stopped", "tick", snapshot.Tick, "status", run.Status)
	s.publish(LiveEvent{
		RunID:        run.ID,
		Kind:         "run.stopped",
		Message:      "Run stopped",
		CurrentTick:  snapshot.Tick,
		RunStatus:    string(run.Status),
		ActiveTasks:  snapshot.ActiveTasks,
		WaitingTasks: snapshot.WaitingTasks,
	})
	return nil
}

func (s *Service) GetRun(runID string) (domain.Run, error) {
	if rt, ok := s.runtimeIfLoaded(runID); ok {
		rt.mu.Lock()
		defer rt.mu.Unlock()
		return rt.engine.Run(), nil
	}
	return s.store.GetRun(runID)
}

func (s *Service) ListRuns() ([]domain.Run, error) {
	return s.store.ListRuns()
}

func (s *Service) GetRequests(runID string) ([]domain.Request, error) {
	if rt, ok := s.runtimeIfLoaded(runID); ok {
		rt.mu.Lock()
		defer rt.mu.Unlock()
		return rt.engine.Requests(), nil
	}
	return s.store.GetRequests(runID)
}

func (s *Service) GetTasks(runID string) ([]domain.TaskOrder, error) {
	if rt, ok := s.runtimeIfLoaded(runID); ok {
		rt.mu.Lock()
		defer rt.mu.Unlock()
		return rt.engine.Tasks(), nil
	}
	return s.store.GetTasks(runID)
}

func (s *Service) GetRobots(runID string) ([]domain.Robot, error) {
	if rt, ok := s.runtimeIfLoaded(runID); ok {
		rt.mu.Lock()
		defer rt.mu.Unlock()
		return rt.engine.Robots(), nil
	}
	return s.store.GetRobots(runID)
}

func (s *Service) GetMetrics(runID string) ([]domain.MetricsSnapshot, error) {
	if rt, ok := s.runtimeIfLoaded(runID); ok {
		rt.mu.Lock()
		defer rt.mu.Unlock()
		return rt.engine.Metrics(), nil
	}
	return s.store.GetMetrics(runID)
}

func (s *Service) GetSegmentLoads(runID string) ([]domain.SegmentLoadSnapshot, error) {
	if rt, ok := s.runtimeIfLoaded(runID); ok {
		rt.mu.Lock()
		defer rt.mu.Unlock()
		return rt.engine.SegmentLoads(), nil
	}
	return s.store.GetSegmentLoads(runID)
}

func (s *Service) GetDecisionLogs(runID string) ([]domain.DecisionLog, error) {
	if rt, ok := s.runtimeIfLoaded(runID); ok {
		rt.mu.Lock()
		defer rt.mu.Unlock()
		return rt.engine.DecisionLogs(), nil
	}
	return s.store.GetDecisionLogs(runID)
}

func (s *Service) GetMap(runID string) (domain.MapData, error) {
	if rt, ok := s.runtimeIfLoaded(runID); ok {
		rt.mu.Lock()
		defer rt.mu.Unlock()
		return rt.engine.Map(), nil
	}
	run, err := s.store.GetRun(runID)
	if err != nil {
		return domain.MapData{}, err
	}
	return s.assets.LoadMap(run.MapPath)
}
