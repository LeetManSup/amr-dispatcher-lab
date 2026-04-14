package simulation

import "amr-dispatcher-lab/internal/domain"

func (e *Engine) appendTaskHistory(taskID string, status domain.TaskStatus, reason string) {
	e.taskHistory = append(e.taskHistory, domain.TaskStatusHistory{
		RunID:  e.run.ID,
		TaskID: taskID,
		Tick:   e.run.CurrentTick,
		Status: status,
		Reason: reason,
	})
}

func (e *Engine) appendSegmentLoads(tick int, segmentLoad map[string]float64) {
	copyLoad := make(map[string]float64, len(segmentLoad))
	for segmentID, load := range segmentLoad {
		copyLoad[segmentID] = load
		e.segmentLoads = append(e.segmentLoads, domain.SegmentLoadSnapshot{
			RunID:     e.run.ID,
			Tick:      tick,
			SegmentID: segmentID,
			Load:      load,
		})
	}
	e.segmentWindow = append(e.segmentWindow, copyLoad)
	if len(e.segmentWindow) > 5 {
		e.segmentWindow = e.segmentWindow[1:]
	}
}

func (e *Engine) collectMetrics(tick int, snapshot domain.SystemSnapshot) domain.MetricsSnapshot {
	total := 0
	waitTotal := 0.0
	executionTotal := 0.0
	completed := 0
	cancelled := 0
	failed := 0
	deadlineSuccess := 0
	for _, task := range e.tasks {
		total++
		waitTotal += float64(observedWait(task, tick))
		switch task.TaskStatus {
		case domain.TaskStatusCompleted:
			completed++
			if task.StartedAt >= 0 {
				executionTotal += float64(task.FinishedAt - task.StartedAt)
			}
			if req := e.requests[task.RequestID]; req != nil && req.Deadline > 0 && task.FinishedAt <= req.Deadline {
				deadlineSuccess++
			}
		case domain.TaskStatusCancelled:
			cancelled++
		case domain.TaskStatusFailed:
			failed++
		}
	}
	avgWait := 0.0
	avgExecution := 0.0
	if total > 0 {
		avgWait = waitTotal / float64(total)
	}
	if completed > 0 {
		avgExecution = executionTotal / float64(completed)
	}
	return domain.MetricsSnapshot{
		RunID:               e.run.ID,
		Tick:                tick,
		AvgWaitTime:         avgWait,
		AvgExecutionTime:    avgExecution,
		Throughput:          float64(completed) / float64(maxInt(1, tick+1)),
		CancelRate:          float64(cancelled) / float64(maxInt(1, total)),
		DeadlineSuccessRate: float64(deadlineSuccess) / float64(maxInt(1, completed)),
		SegmentLoadVariance: variance(segmentRollingValues(e.segmentWindow)),
		RobotUtilization:    snapshot.RobotUtilization,
		CompletedTasks:      completed,
		CancelledTasks:      cancelled,
		FailedTasks:         failed,
	}
}

func observedWait(task *domain.TaskOrder, tick int) int {
	switch {
	case task.StartedAt >= 0:
		return maxInt(0, task.StartedAt-task.CreatedAt)
	case task.FinishedAt >= 0:
		return maxInt(0, task.FinishedAt-task.CreatedAt)
	default:
		return maxInt(0, tick-task.CreatedAt)
	}
}
