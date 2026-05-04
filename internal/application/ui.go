package application

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"amr-dispatcher-lab/internal/domain"
)

type CatalogItem struct {
	Path        string `json:"path"`
	Label       string `json:"label"`
	Kind        string `json:"kind"`
	Description string `json:"description,omitempty"`
}

type RunOverview struct {
	Run             domain.Run              `json:"run"`
	Snapshot        domain.SystemSnapshot   `json:"snapshot"`
	LatestMetrics   *domain.MetricsSnapshot `json:"latestMetrics,omitempty"`
	TaskCounts      map[string]int          `json:"taskCounts"`
	RobotCounts     map[string]int          `json:"robotCounts"`
	RecentDecisions []domain.DecisionLog    `json:"recentDecisions"`
	Alerts          []string                `json:"alerts,omitempty"`
}

type LiveEvent struct {
	EventID      string    `json:"eventId"`
	RunID        string    `json:"runId"`
	Kind         string    `json:"kind"`
	Message      string    `json:"message"`
	CurrentTick  int       `json:"currentTick"`
	RunStatus    string    `json:"runStatus"`
	ActiveTasks  int       `json:"activeTasks"`
	WaitingTasks int       `json:"waitingTasks"`
	OccurredAt   time.Time `json:"occurredAt"`
}

type liveSubscription struct {
	runID string
	ch    chan LiveEvent
}

func (s *Service) GetOverview(runID string) (RunOverview, error) {
	if rt, ok := s.runtimeIfLoaded(runID); ok {
		rt.mu.Lock()
		defer rt.mu.Unlock()
		return buildOverview(
			rt.engine.Run(),
			rt.engine.CurrentSnapshot(),
			rt.engine.Tasks(),
			rt.engine.Robots(),
			rt.engine.Metrics(),
			rt.engine.DecisionLogs(),
		), nil
	}

	run, err := s.store.GetRun(runID)
	if err != nil {
		return RunOverview{}, err
	}
	tasks, err := s.store.GetTasks(runID)
	if err != nil {
		return RunOverview{}, err
	}
	robots, err := s.store.GetRobots(runID)
	if err != nil {
		return RunOverview{}, err
	}
	metrics, err := s.store.GetMetrics(runID)
	if err != nil {
		return RunOverview{}, err
	}
	decisions, err := s.store.GetDecisionLogs(runID)
	if err != nil {
		return RunOverview{}, err
	}
	return buildOverview(run, deriveSnapshot(run, tasks, robots), tasks, robots, metrics, decisions), nil
}

func (s *Service) Subscribe(runID string) (<-chan LiveEvent, func()) {
	id := atomic.AddUint64(&s.subscriberCounter, 1)
	ch := make(chan LiveEvent, 16)

	s.eventMu.Lock()
	s.subscribers[id] = liveSubscription{runID: runID, ch: ch}
	s.eventMu.Unlock()

	return ch, func() {
		s.eventMu.Lock()
		sub, ok := s.subscribers[id]
		if ok {
			delete(s.subscribers, id)
			close(sub.ch)
		}
		s.eventMu.Unlock()
	}
}

func (s *Service) publish(event LiveEvent) {
	if event.RunID == "" {
		return
	}
	if event.EventID == "" {
		event.EventID = fmt.Sprintf("evt-%d-%d", s.clock.Now().UnixNano(), atomic.AddUint64(&s.eventCounter, 1))
	}
	if event.OccurredAt.IsZero() {
		event.OccurredAt = s.clock.Now()
	}

	s.eventMu.RLock()
	defer s.eventMu.RUnlock()
	for _, sub := range s.subscribers {
		if sub.runID != event.RunID {
			continue
		}
		select {
		case sub.ch <- event:
		default:
			select {
			case <-sub.ch:
			default:
			}
			select {
			case sub.ch <- event:
			default:
			}
		}
	}
}

func buildOverview(run domain.Run, snapshot domain.SystemSnapshot, tasks []domain.TaskOrder, robots []domain.Robot, metrics []domain.MetricsSnapshot, decisions []domain.DecisionLog) RunOverview {
	taskCounts := make(map[string]int)
	for _, task := range tasks {
		taskCounts[string(task.TaskStatus)]++
	}

	robotCounts := make(map[string]int)
	for _, robot := range robots {
		robotCounts[string(robot.State)]++
	}

	var latestMetrics *domain.MetricsSnapshot
	if len(metrics) > 0 {
		item := metrics[len(metrics)-1]
		latestMetrics = &item
	}

	recentDecisions := tailDecisions(decisions, 8)
	alerts := collectAlerts(run, tasks, robots, latestMetrics)

	return RunOverview{
		Run:             run,
		Snapshot:        snapshot,
		LatestMetrics:   latestMetrics,
		TaskCounts:      taskCounts,
		RobotCounts:     robotCounts,
		RecentDecisions: recentDecisions,
		Alerts:          alerts,
	}
}

func deriveSnapshot(run domain.Run, tasks []domain.TaskOrder, robots []domain.Robot) domain.SystemSnapshot {
	snapshot := domain.SystemSnapshot{
		Tick:        run.CurrentTick,
		SegmentLoad: map[string]float64{},
		ZoneLoad:    map[string]float64{},
	}

	totalRobots := len(robots)
	busyRobots := 0
	for _, robot := range robots {
		switch robot.State {
		case domain.RobotStateIdle:
			snapshot.AvailableRobots++
		case domain.RobotStateFault:
			snapshot.FaultRobots++
		case domain.RobotStateCharging:
			snapshot.ChargingRobots++
		}
		if robot.State == domain.RobotStateBusy {
			busyRobots++
		}
	}
	if totalRobots > 0 {
		snapshot.RobotUtilization = float64(busyRobots) / float64(totalRobots)
	}

	for _, task := range tasks {
		switch task.TaskStatus {
		case domain.TaskStatusQueued, domain.TaskStatusCreated:
			snapshot.WaitingTasks++
		case domain.TaskStatusAssigned, domain.TaskStatusInProgress:
			snapshot.ActiveTasks++
		case domain.TaskStatusPaused:
			snapshot.PausedTasks++
		}
	}
	return snapshot
}

func tailDecisions(items []domain.DecisionLog, size int) []domain.DecisionLog {
	if len(items) <= size {
		return append([]domain.DecisionLog(nil), items...)
	}
	return append([]domain.DecisionLog(nil), items[len(items)-size:]...)
}

func collectAlerts(run domain.Run, tasks []domain.TaskOrder, robots []domain.Robot, latestMetrics *domain.MetricsSnapshot) []string {
	alerts := make([]string, 0, 4)
	if run.LastError != "" {
		alerts = append(alerts, run.LastError)
	}
	failedTasks := 0
	for _, task := range tasks {
		if task.TaskStatus == domain.TaskStatusFailed {
			failedTasks++
		}
	}
	if failedTasks > 0 {
		alerts = append(alerts, fmt.Sprintf("Failed tasks: %d", failedTasks))
	}
	faultedRobots := 0
	for _, robot := range robots {
		if robot.State == domain.RobotStateFault {
			faultedRobots++
		}
	}
	if faultedRobots > 0 {
		alerts = append(alerts, fmt.Sprintf("Faulted robots: %d", faultedRobots))
	}
	if latestMetrics != nil && latestMetrics.CancelRate > 0.25 {
		alerts = append(alerts, "Cancel rate is elevated")
	}
	return alerts
}

func buildCatalogItems(paths []string, kind string, loadMetadata func(string) (string, string, error)) []CatalogItem {
	items := make([]CatalogItem, 0, len(paths))
	for _, path := range paths {
		label, description, err := loadMetadata(path)
		if err != nil {
			label = humanizePathLabel(path)
			description = ""
		}
		items = append(items, CatalogItem{
			Path:        path,
			Label:       label,
			Kind:        kind,
			Description: description,
		})
	}
	return items
}

func humanizePathLabel(path string) string {
	base := filepath.Base(path)
	base = strings.TrimSuffix(base, filepath.Ext(base))
	parts := strings.FieldsFunc(base, func(r rune) bool {
		return r == '-' || r == '_' || r == ' '
	})
	if len(parts) == 0 {
		return path
	}
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
	}
	return strings.Join(parts, " ")
}
