package simulation

import "amr-dispatcher-lab/internal/domain"

func (e *Engine) intakeRequests(tick int) {
	for _, plan := range e.scenario.Requests {
		if plan.ReleaseTick > tick || e.released[plan.RequestID] {
			continue
		}
		req := &domain.Request{
			RequestID:    plan.RequestID,
			SourcePoint:  plan.SourcePoint,
			TargetPoint:  plan.TargetPoint,
			BusinessType: plan.BusinessType,
			Priority:     plan.Priority,
			CreatedAt:    tick,
			Deadline:     plan.Deadline,
			Status:       domain.RequestStatusCreated,
		}
		if err := req.Transition(domain.RequestStatusValidated); err != nil {
			continue
		}
		if err := req.Transition(domain.RequestStatusQueued); err != nil {
			continue
		}
		e.requests[req.RequestID] = req
		e.released[req.RequestID] = true
		_, _ = e.createTaskForRequest(req)
	}
}

func (e *Engine) createTaskForRequest(req *domain.Request) (*domain.TaskOrder, error) {
	task := e.taskFactory.Build(req)
	if err := task.Transition(domain.TaskStatusQueued); err != nil {
		return nil, err
	}
	if err := req.Transition(domain.RequestStatusConvertedToTask); err != nil {
		return nil, err
	}
	e.tasks[task.TaskID] = task
	e.appendTaskHistory(task.TaskID, task.TaskStatus, "created_from_request")
	return task, nil
}
