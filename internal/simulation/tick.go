package simulation

import (
	"amr-dispatcher-lab/internal/domain"
	"fmt"
)

func (e *Engine) Step() (TickResult, error) {
	if e.run.Status == domain.RunStatusCompleted || e.run.Status == domain.RunStatusFailed {
		return TickResult{}, fmt.Errorf("run %s cannot step in status %s", e.run.ID, e.run.Status)
	}
	if e.run.Status == domain.RunStatusCreated {
		e.run.Status = domain.RunStatusRunning
		now := e.clock.Now()
		e.run.StartedAt = &now
	}

	tick := e.run.CurrentTick
	historyStart := len(e.taskHistory)
	decisionStart := len(e.decisionLogs)
	segmentStart := len(e.segmentLoads)

	e.intakeRequests(tick)
	e.updateRobots()
	e.progressTasks(tick)
	e.finishTasks(tick)

	segmentLoad, zoneLoad := e.computeLoads()
	snapshot := e.buildSnapshot(tick, segmentLoad, zoneLoad)
	candidates := e.buildCandidates(tick, snapshot)
	decisions := e.scheduler.Decide(tick, snapshot, candidates)
	e.applySchedulerDecisions(tick, snapshot, decisions, candidates)
	e.runSupervision(tick)

	segmentLoad, zoneLoad = e.computeLoads()
	snapshot = e.buildSnapshot(tick, segmentLoad, zoneLoad)
	e.appendSegmentLoads(tick, segmentLoad)
	metrics := e.collectMetrics(tick, snapshot)
	e.metrics = append(e.metrics, metrics)
	e.run.CurrentTick++

	if e.shouldFinish() {
		e.completeRun(domain.RunStatusCompleted, "")
	}

	return TickResult{
		Snapshot:      snapshot,
		Metrics:       metrics,
		Decisions:     append([]domain.DecisionLog(nil), e.decisionLogs[decisionStart:]...),
		StatusHistory: append([]domain.TaskStatusHistory(nil), e.taskHistory[historyStart:]...),
		SegmentLoads:  append([]domain.SegmentLoadSnapshot(nil), e.segmentLoads[segmentStart:]...),
	}, nil
}

func (e *Engine) Stop() {
	if e.run.Status == domain.RunStatusRunning || e.run.Status == domain.RunStatusCreated {
		e.completeRun(domain.RunStatusStopped, "")
	}
}
