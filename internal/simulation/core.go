package simulation

import (
	"amr-dispatcher-lab/internal/domain"
	"amr-dispatcher-lab/internal/planner"
	"amr-dispatcher-lab/internal/scheduler"
	"errors"
	"math/rand"
	"sort"
)

type TickResult struct {
	Snapshot      domain.SystemSnapshot        `json:"snapshot"`
	Metrics       domain.MetricsSnapshot       `json:"metrics"`
	Decisions     []domain.DecisionLog         `json:"decisions"`
	StatusHistory []domain.TaskStatusHistory   `json:"statusHistory"`
	SegmentLoads  []domain.SegmentLoadSnapshot `json:"segmentLoads"`
}

type Engine struct {
	run           domain.Run
	config        domain.RunConfig
	mapData       domain.MapData
	scenario      domain.Scenario
	planner       planner.RoutePlanner
	scheduler     scheduler.Scheduler
	admission     domain.AdmissionController
	clock         domain.Clock
	random        domain.RandomSource
	taskFactory   domain.TaskFactory
	requests      map[string]*domain.Request
	tasks         map[string]*domain.TaskOrder
	robots        map[string]*domain.Robot
	taskHistory   []domain.TaskStatusHistory
	decisionLogs  []domain.DecisionLog
	metrics       []domain.MetricsSnapshot
	segmentLoads  []domain.SegmentLoadSnapshot
	segmentWindow []map[string]float64
	released      map[string]bool
}

type defaultRandom struct {
	rng *rand.Rand
}

func (r defaultRandom) Float64() float64 { return r.rng.Float64() }

func NewEngine(id string, cfg domain.RunConfig, m domain.MapData, s domain.Scenario, clock domain.Clock, random domain.RandomSource) (*Engine, error) {
	graphPlanner, err := planner.NewGraphPlanner(m)
	if err != nil {
		return nil, err
	}
	if clock == nil {
		clock = domain.RealClock{}
	}
	if random == nil {
		random = defaultRandom{rng: rand.New(rand.NewSource(cfg.Seed))}
	}

	engine := &Engine{
		run: domain.Run{
			ID:           id,
			Status:       domain.RunStatusCreated,
			Mode:         domain.RunModeStep,
			Algorithm:    cfg.Algorithm,
			MapName:      m.Name,
			MapPath:      cfg.MapPath,
			ScenarioName: s.Name,
			ScenarioPath: cfg.ScenarioPath,
			Seed:         cfg.Seed,
			ConfigHash:   hashRunConfig(cfg, m.Name, s.Name),
			CreatedAt:    clock.Now(),
		},
		config:        cfg,
		mapData:       m,
		scenario:      s,
		planner:       graphPlanner,
		scheduler:     scheduler.New(cfg.Algorithm, mergeWeights(cfg.Weights, s.Weights)),
		admission:     domain.AdmissionController{Config: mergeAdmission(cfg.Admission, s.Admission)},
		clock:         clock,
		random:        random,
		taskFactory:   domain.TaskFactory{},
		requests:      make(map[string]*domain.Request),
		tasks:         make(map[string]*domain.TaskOrder),
		robots:        make(map[string]*domain.Robot),
		released:      make(map[string]bool),
		segmentWindow: make([]map[string]float64, 0, 5),
	}
	if len(s.Robots) == 0 {
		return nil, errors.New("scenario has no robots")
	}
	for _, robot := range s.Robots {
		engine.robots[robot.RobotID] = &domain.Robot{
			RobotID:      robot.RobotID,
			RobotModel:   robot.RobotModel,
			State:        domain.RobotStateIdle,
			CurrentPoint: robot.StartPoint,
			BatteryLevel: robot.BatteryLevel,
		}
	}
	return engine, nil
}

func (e *Engine) Run() domain.Run { return e.run }

func (e *Engine) SetMode(mode domain.RunMode) {
	e.run.Mode = mode
}

func (e *Engine) Map() domain.MapData { return e.mapData }

func (e *Engine) Requests() []domain.Request {
	items := make([]domain.Request, 0, len(e.requests))
	for _, item := range e.requests {
		items = append(items, *item)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt < items[j].CreatedAt })
	return items
}

func (e *Engine) Tasks() []domain.TaskOrder {
	items := make([]domain.TaskOrder, 0, len(e.tasks))
	for _, item := range e.tasks {
		items = append(items, *item)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt < items[j].CreatedAt })
	return items
}

func (e *Engine) Robots() []domain.Robot {
	items := make([]domain.Robot, 0, len(e.robots))
	for _, item := range e.robots {
		items = append(items, *item)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].RobotID < items[j].RobotID })
	return items
}

func (e *Engine) DecisionLogs() []domain.DecisionLog {
	return append([]domain.DecisionLog(nil), e.decisionLogs...)
}

func (e *Engine) Metrics() []domain.MetricsSnapshot {
	return append([]domain.MetricsSnapshot(nil), e.metrics...)
}

func (e *Engine) SegmentLoads() []domain.SegmentLoadSnapshot {
	return append([]domain.SegmentLoadSnapshot(nil), e.segmentLoads...)
}

func (e *Engine) TaskHistory() []domain.TaskStatusHistory {
	return append([]domain.TaskStatusHistory(nil), e.taskHistory...)
}

func (e *Engine) CurrentSnapshot() domain.SystemSnapshot {
	segmentLoad, zoneLoad := e.computeLoads()
	return e.buildSnapshot(e.run.CurrentTick, segmentLoad, zoneLoad)
}
