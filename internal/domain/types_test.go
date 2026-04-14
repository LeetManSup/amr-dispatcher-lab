package domain

import "testing"

func TestRequestTransitions(t *testing.T) {
	request := &Request{Status: RequestStatusCreated}
	if err := request.Transition(RequestStatusValidated); err != nil {
		t.Fatalf("expected created -> validated transition to succeed: %v", err)
	}
	if err := request.Transition(RequestStatusQueued); err != nil {
		t.Fatalf("expected validated -> queued transition to succeed: %v", err)
	}
	if err := request.Transition(RequestStatusCompleted); err == nil {
		t.Fatal("expected queued -> completed transition to fail")
	}
}

func TestTaskTransitions(t *testing.T) {
	task := &TaskOrder{TaskStatus: TaskStatusCreated}
	if err := task.Transition(TaskStatusQueued); err != nil {
		t.Fatalf("expected created -> queued transition to succeed: %v", err)
	}
	if err := task.Transition(TaskStatusAssigned); err != nil {
		t.Fatalf("expected queued -> assigned transition to succeed: %v", err)
	}
	if err := task.Transition(TaskStatusCompleted); err == nil {
		t.Fatal("expected assigned -> completed transition to fail")
	}
}

func TestRobotTransitions(t *testing.T) {
	robot := &Robot{State: RobotStateIdle}
	if err := robot.Transition(RobotStateBusy); err != nil {
		t.Fatalf("expected idle -> busy transition to succeed: %v", err)
	}
	if err := robot.Transition(RobotStateFault); err != nil {
		t.Fatalf("expected busy -> fault transition to succeed: %v", err)
	}
	if !robot.FaultFlag {
		t.Fatal("expected fault flag to be set")
	}
	if err := robot.Transition(RobotStateBusy); err == nil {
		t.Fatal("expected fault -> busy transition to fail")
	}
}
