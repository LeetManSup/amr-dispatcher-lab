package application

import (
	"fmt"

	"amr-dispatcher-lab/internal/domain"
)

func (s *Service) AddTask(runID string, input AddTaskInput) (*domain.TaskOrder, error) {
	rt, err := s.runtime(runID)
	if err != nil {
		s.logger.Warn("add task rejected", "run_id", runID, "request_id", input.RequestID, "err", err)
		return nil, err
	}
	logger := s.logger.With("run_id", runID, "request_id", input.RequestID)
	rt.mu.Lock()
	defer rt.mu.Unlock()

	historyBefore := len(rt.engine.TaskHistory())
	req := domain.Request{
		RequestID:    input.RequestID,
		SourcePoint:  input.SourcePoint,
		TargetPoint:  input.TargetPoint,
		BusinessType: input.BusinessType,
		Priority:     input.Priority,
		CreatedAt:    input.CreatedAt,
		Deadline:     input.Deadline,
	}
	task, err := rt.engine.AddExternalRequest(req)
	if err != nil {
		logger.Warn("add external request failed", "err", err)
		return nil, err
	}
	if err := s.store.UpsertRequests(runID, rt.engine.Requests()); err != nil {
		logger.Error("persist requests after add task failed", "err", err)
		return nil, err
	}
	if err := s.store.UpsertTasks(runID, rt.engine.Tasks()); err != nil {
		logger.Error("persist tasks after add task failed", "err", err)
		return nil, err
	}
	history := rt.engine.TaskHistory()
	if len(history) > historyBefore {
		if err := s.store.AppendTaskHistory(history[historyBefore:]); err != nil {
			logger.Error("append task history after add task failed", "err", err)
			return nil, err
		}
	}
	snapshot := rt.engine.CurrentSnapshot()
	logger.Info("task added", "task_id", task.TaskID, "tick", snapshot.Tick, "target_point", task.TargetPoint)
	s.publish(LiveEvent{
		RunID:        runID,
		Kind:         "task.added",
		Message:      fmt.Sprintf("Task %s added", task.TaskID),
		CurrentTick:  snapshot.Tick,
		RunStatus:    string(rt.engine.Run().Status),
		ActiveTasks:  snapshot.ActiveTasks,
		WaitingTasks: snapshot.WaitingTasks,
	})
	return task, nil
}

func (s *Service) CancelTask(runID, taskID string) error {
	rt, err := s.runtime(runID)
	if err != nil {
		s.logger.Warn("cancel task rejected", "run_id", runID, "task_id", taskID, "err", err)
		return err
	}
	logger := s.logger.With("run_id", runID, "task_id", taskID)
	rt.mu.Lock()
	defer rt.mu.Unlock()

	historyBefore := len(rt.engine.TaskHistory())
	if err := rt.engine.CancelTask(taskID); err != nil {
		logger.Warn("cancel task failed", "err", err)
		return err
	}
	if err := s.persistStateOnly(rt.engine); err != nil {
		logger.Error("persist state after cancel task failed", "err", err)
		return err
	}
	history := rt.engine.TaskHistory()
	if len(history) > historyBefore {
		if err := s.store.AppendTaskHistory(history[historyBefore:]); err != nil {
			logger.Error("append history after cancel task failed", "err", err)
			return err
		}
	}
	snapshot := rt.engine.CurrentSnapshot()
	logger.Info("task cancelled", "tick", snapshot.Tick)
	s.publish(LiveEvent{
		RunID:        runID,
		Kind:         "task.cancelled",
		Message:      fmt.Sprintf("Task %s cancelled", taskID),
		CurrentTick:  snapshot.Tick,
		RunStatus:    string(rt.engine.Run().Status),
		ActiveTasks:  snapshot.ActiveTasks,
		WaitingTasks: snapshot.WaitingTasks,
	})
	return nil
}

func (s *Service) ContinueTask(runID, taskID string) error {
	rt, err := s.runtime(runID)
	if err != nil {
		s.logger.Warn("continue task rejected", "run_id", runID, "task_id", taskID, "err", err)
		return err
	}
	logger := s.logger.With("run_id", runID, "task_id", taskID)
	rt.mu.Lock()
	defer rt.mu.Unlock()

	historyBefore := len(rt.engine.TaskHistory())
	if err := rt.engine.ContinueTask(taskID); err != nil {
		logger.Warn("continue task failed", "err", err)
		return err
	}
	if err := s.persistStateOnly(rt.engine); err != nil {
		logger.Error("persist state after continue task failed", "err", err)
		return err
	}
	history := rt.engine.TaskHistory()
	if len(history) > historyBefore {
		if err := s.store.AppendTaskHistory(history[historyBefore:]); err != nil {
			logger.Error("append history after continue task failed", "err", err)
			return err
		}
	}
	snapshot := rt.engine.CurrentSnapshot()
	logger.Info("task continued", "tick", snapshot.Tick)
	s.publish(LiveEvent{
		RunID:        runID,
		Kind:         "task.continued",
		Message:      fmt.Sprintf("Task %s continued", taskID),
		CurrentTick:  snapshot.Tick,
		RunStatus:    string(rt.engine.Run().Status),
		ActiveTasks:  snapshot.ActiveTasks,
		WaitingTasks: snapshot.WaitingTasks,
	})
	return nil
}

func (s *Service) UpdateOrderPointInfo(runID, taskID, targetPoint string) error {
	rt, err := s.runtime(runID)
	if err != nil {
		s.logger.Warn("retarget task rejected", "run_id", runID, "task_id", taskID, "target_point", targetPoint, "err", err)
		return err
	}
	logger := s.logger.With("run_id", runID, "task_id", taskID)
	rt.mu.Lock()
	defer rt.mu.Unlock()

	if err := rt.engine.UpdateTaskTarget(taskID, targetPoint); err != nil {
		logger.Warn("retarget task failed", "target_point", targetPoint, "err", err)
		return err
	}
	if err := s.persistStateOnly(rt.engine); err != nil {
		logger.Error("persist state after retarget task failed", "err", err)
		return err
	}
	snapshot := rt.engine.CurrentSnapshot()
	logger.Info("task retargeted", "tick", snapshot.Tick, "target_point", targetPoint)
	s.publish(LiveEvent{
		RunID:        runID,
		Kind:         "task.retargeted",
		Message:      fmt.Sprintf("Task %s retargeted to %s", taskID, targetPoint),
		CurrentTick:  snapshot.Tick,
		RunStatus:    string(rt.engine.Run().Status),
		ActiveTasks:  snapshot.ActiveTasks,
		WaitingTasks: snapshot.WaitingTasks,
	})
	return nil
}

func (s *Service) GetTaskOrderStatus(runID, taskID string) (*domain.TaskOrder, error) {
	rt, err := s.runtime(runID)
	if err == nil {
		rt.mu.Lock()
		defer rt.mu.Unlock()
		return rt.engine.GetTask(taskID)
	}
	tasks, storeErr := s.store.GetTasks(runID)
	if storeErr != nil {
		return nil, storeErr
	}
	for _, task := range tasks {
		if task.TaskID == taskID {
			copy := task
			return &copy, nil
		}
	}
	return nil, fmt.Errorf("task %s not found", taskID)
}

func (s *Service) DeviceInfo(runID string) ([]domain.Robot, error) {
	return s.GetRobots(runID)
}

func (s *Service) GetRobotTaskPath(runID, taskID string) ([]string, error) {
	rt, err := s.runtime(runID)
	if err == nil {
		rt.mu.Lock()
		defer rt.mu.Unlock()
		return rt.engine.GetTaskPath(taskID)
	}
	tasks, storeErr := s.store.GetTasks(runID)
	if storeErr != nil {
		return nil, storeErr
	}
	for _, task := range tasks {
		if task.TaskID == taskID {
			return task.PlannedPath, nil
		}
	}
	return nil, fmt.Errorf("task %s not found", taskID)
}
