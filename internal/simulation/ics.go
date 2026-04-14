package simulation

import (
	"amr-dispatcher-lab/internal/domain"
	"fmt"
)

func (e *Engine) AddExternalRequest(req domain.Request) (*domain.TaskOrder, error) {
	if existing, ok := e.requests[req.RequestID]; ok {
		return e.tasks["task-"+existing.RequestID], nil
	}
	req.Status = domain.RequestStatusCreated
	if err := req.Transition(domain.RequestStatusValidated); err != nil {
		return nil, err
	}
	if err := req.Transition(domain.RequestStatusQueued); err != nil {
		return nil, err
	}
	e.requests[req.RequestID] = &req
	return e.createTaskForRequest(&req)
}

func (e *Engine) CancelTask(taskID string) error {
	task, ok := e.tasks[taskID]
	if !ok {
		return fmt.Errorf("task %s not found", taskID)
	}
	if task.TaskStatus == domain.TaskStatusCancelled || task.TaskStatus == domain.TaskStatusCompleted {
		return nil
	}
	if err := task.Transition(domain.TaskStatusCancelled); err != nil {
		return err
	}
	task.FinishedAt = e.run.CurrentTick
	e.appendTaskHistory(task.TaskID, task.TaskStatus, "cancelled_via_ics")
	if req := e.requests[task.RequestID]; req != nil && req.Status != domain.RequestStatusCancelled {
		_ = req.Transition(domain.RequestStatusCancelled)
	}
	e.releaseRobot(task.RobotID)
	return nil
}

func (e *Engine) ContinueTask(taskID string) error {
	task, ok := e.tasks[taskID]
	if !ok {
		return fmt.Errorf("task %s not found", taskID)
	}
	if task.TaskStatus != domain.TaskStatusPaused {
		return nil
	}
	if err := task.Transition(domain.TaskStatusQueued); err != nil {
		return err
	}
	task.DeferUntilTick = 0
	e.appendTaskHistory(task.TaskID, task.TaskStatus, "continued_via_ics")
	return nil
}

func (e *Engine) UpdateTaskTarget(taskID, targetPoint string) error {
	task, ok := e.tasks[taskID]
	if !ok {
		return fmt.Errorf("task %s not found", taskID)
	}
	req := e.requests[task.RequestID]
	sourceReached := task.CurrentTargetPoint == task.TargetPoint
	task.TargetPoint = targetPoint
	if req != nil {
		req.TargetPoint = targetPoint
	}
	task.CurrentTargetPoint = task.SourcePoint
	if sourceReached {
		task.CurrentTargetPoint = targetPoint
	}
	if task.RobotID != "" {
		robot := e.robots[task.RobotID]
		path, currentTarget, err := e.buildExecutionPath(robot.CurrentPoint, task.SourcePoint, targetPoint, sourceReached)
		if err != nil {
			return err
		}
		task.PlannedPath = path.Points
		task.PathIndex = 0
		task.EstimatedCost = path.Cost
		task.EstimatedDuration = path.ETA
		task.CurrentTargetPoint = currentTarget
	}
	return nil
}

func (e *Engine) GetTask(taskID string) (*domain.TaskOrder, error) {
	task, ok := e.tasks[taskID]
	if !ok {
		return nil, fmt.Errorf("task %s not found", taskID)
	}
	copy := *task
	return &copy, nil
}

func (e *Engine) GetTaskPath(taskID string) ([]string, error) {
	task, ok := e.tasks[taskID]
	if !ok {
		return nil, fmt.Errorf("task %s not found", taskID)
	}
	return append([]string(nil), task.PlannedPath...), nil
}
