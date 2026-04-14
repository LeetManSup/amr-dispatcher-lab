package domain

import "fmt"

var requestTransitions = map[RequestStatus]map[RequestStatus]bool{
	RequestStatusCreated: {
		RequestStatusValidated: true,
		RequestStatusRejected:  true,
	},
	RequestStatusValidated: {
		RequestStatusQueued:    true,
		RequestStatusRejected:  true,
		RequestStatusCancelled: true,
	},
	RequestStatusQueued: {
		RequestStatusConvertedToTask: true,
		RequestStatusCancelled:       true,
	},
	RequestStatusConvertedToTask: {
		RequestStatusCompleted: true,
		RequestStatusCancelled: true,
	},
	RequestStatusCompleted: {},
	RequestStatusRejected:  {},
	RequestStatusCancelled: {},
}

func (r *Request) Transition(next RequestStatus) error {
	if !requestTransitions[r.Status][next] {
		return fmt.Errorf("request transition %s -> %s is not allowed", r.Status, next)
	}
	r.Status = next
	return nil
}

var taskTransitions = map[TaskStatus]map[TaskStatus]bool{
	TaskStatusCreated: {
		TaskStatusQueued:    true,
		TaskStatusCancelled: true,
	},
	TaskStatusQueued: {
		TaskStatusAssigned:  true,
		TaskStatusPaused:    true,
		TaskStatusCancelled: true,
		TaskStatusFailed:    true,
	},
	TaskStatusAssigned: {
		TaskStatusInProgress: true,
		TaskStatusCancelled:  true,
		TaskStatusFailed:     true,
	},
	TaskStatusInProgress: {
		TaskStatusCompleted: true,
		TaskStatusPaused:    true,
		TaskStatusCancelled: true,
		TaskStatusFailed:    true,
	},
	TaskStatusPaused: {
		TaskStatusQueued:     true,
		TaskStatusInProgress: true,
		TaskStatusCancelled:  true,
		TaskStatusFailed:     true,
	},
	TaskStatusCompleted: {},
	TaskStatusCancelled: {},
	TaskStatusFailed:    {},
}

func (t *TaskOrder) Transition(next TaskStatus) error {
	if !taskTransitions[t.TaskStatus][next] {
		return fmt.Errorf("task transition %s -> %s is not allowed", t.TaskStatus, next)
	}
	t.TaskStatus = next
	return nil
}

var robotTransitions = map[RobotState]map[RobotState]bool{
	RobotStateIdle: {
		RobotStateBusy:     true,
		RobotStateCharging: true,
		RobotStateFault:    true,
		RobotStatePaused:   true,
	},
	RobotStateBusy: {
		RobotStateIdle:     true,
		RobotStateCharging: true,
		RobotStateFault:    true,
		RobotStatePaused:   true,
	},
	RobotStateCharging: {
		RobotStateIdle:  true,
		RobotStateFault: true,
	},
	RobotStateFault: {
		RobotStateIdle: true,
	},
	RobotStatePaused: {
		RobotStateIdle: true,
		RobotStateBusy: true,
	},
}

func (r *Robot) Transition(next RobotState) error {
	if !robotTransitions[r.State][next] {
		return fmt.Errorf("robot transition %s -> %s is not allowed", r.State, next)
	}
	r.State = next
	r.FaultFlag = next == RobotStateFault
	return nil
}
