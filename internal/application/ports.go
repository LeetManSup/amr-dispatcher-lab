package application

import "amr-dispatcher-lab/internal/domain"

type RunStore interface {
	SaveRun(run domain.Run) error
	UpdateRun(run domain.Run) error
	UpsertRequests(runID string, requests []domain.Request) error
	UpsertTasks(runID string, tasks []domain.TaskOrder) error
	UpsertRobots(runID string, robots []domain.Robot) error
	AppendTaskHistory(items []domain.TaskStatusHistory) error
	AppendDecisionLogs(items []domain.DecisionLog) error
	AppendMetrics(items []domain.MetricsSnapshot) error
	AppendSegmentLoads(items []domain.SegmentLoadSnapshot) error
	GetRun(id string) (domain.Run, error)
	ListRuns() ([]domain.Run, error)
	GetRequests(runID string) ([]domain.Request, error)
	GetTasks(runID string) ([]domain.TaskOrder, error)
	GetRobots(runID string) ([]domain.Robot, error)
	GetMetrics(runID string) ([]domain.MetricsSnapshot, error)
	GetSegmentLoads(runID string) ([]domain.SegmentLoadSnapshot, error)
	GetDecisionLogs(runID string) ([]domain.DecisionLog, error)
	Close() error
}

type ExperimentAssets interface {
	ListMaps() ([]string, error)
	ListScenarios() ([]string, error)
	LoadMap(path string) (domain.MapData, error)
	LoadScenario(path string) (domain.Scenario, error)
}
