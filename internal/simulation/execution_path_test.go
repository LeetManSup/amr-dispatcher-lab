package simulation

import (
	"amr-dispatcher-lab/internal/domain"
	"testing"
)

func TestTaskFactoryTargetsSourceFirst(t *testing.T) {
	req := &domain.Request{
		RequestID:   "REQ-1",
		SourcePoint: "P2",
		TargetPoint: "P4",
		CreatedAt:   0,
	}

	task := domain.TaskFactory{}.Build(req)
	if task.CurrentTargetPoint != "P2" {
		t.Fatalf("expected current target to start at source point, got %s", task.CurrentTargetPoint)
	}
}

func TestBuildExecutionPathIncludesSourceThenTarget(t *testing.T) {
	engine := testEngineForExecutionPath(t)

	path, currentTarget, err := engine.buildExecutionPath("P1", "P2", "P4", false)
	if err != nil {
		t.Fatalf("build execution path: %v", err)
	}
	expected := []string{"P1", "P2", "P3", "P4"}
	if len(path.Points) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, path.Points)
	}
	for i := range expected {
		if path.Points[i] != expected[i] {
			t.Fatalf("expected %v, got %v", expected, path.Points)
		}
	}
	if currentTarget != "P2" {
		t.Fatalf("expected current target P2, got %s", currentTarget)
	}
}

func TestProgressSwitchesTargetAfterSourceReached(t *testing.T) {
	engine := testEngineForExecutionPath(t)
	task := domain.TaskFactory{}.Build(&domain.Request{
		RequestID:   "REQ-1",
		SourcePoint: "P2",
		TargetPoint: "P4",
		CreatedAt:   0,
	})
	if err := task.Transition(domain.TaskStatusQueued); err != nil {
		t.Fatalf("queue task: %v", err)
	}

	robot := engine.robots["R1"]
	path, _, err := engine.buildExecutionPath(robot.CurrentPoint, task.SourcePoint, task.TargetPoint, false)
	if err != nil {
		t.Fatalf("build path: %v", err)
	}
	if err := engine.dispatchTask(0, robot, task, path); err != nil {
		t.Fatalf("dispatch task: %v", err)
	}

	engine.tasks[task.TaskID] = task
	engine.progressTasks(1)

	if robot.CurrentPoint != "P2" {
		t.Fatalf("expected robot to reach source P2, got %s", robot.CurrentPoint)
	}
	if task.CurrentTargetPoint != "P4" {
		t.Fatalf("expected task target to switch to final target P4 after pickup, got %s", task.CurrentTargetPoint)
	}
}

func testEngineForExecutionPath(t *testing.T) *Engine {
	t.Helper()

	engine, err := NewEngine(
		"test-run",
		domain.RunConfig{Algorithm: domain.AlgorithmFIFO, Seed: 123},
		domain.MapData{
			Name: "Mini",
			Points: []domain.MapPoint{
				{PointID: "P1", PointType: domain.PointTypeRunning, AreaID: "A"},
				{PointID: "P2", PointType: domain.PointTypeQueue, AreaID: "A"},
				{PointID: "P3", PointType: domain.PointTypeWorkbench, AreaID: "B"},
				{PointID: "P4", PointType: domain.PointTypeShelf, AreaID: "B"},
			},
			Segments: []domain.MapSegment{
				{SegmentID: "S1", FromPoint: "P1", ToPoint: "P2", Length: 1, Direction: domain.SegmentDirectionBidirectional, SpeedLimit: 1, RouteCostWeight: 1},
				{SegmentID: "S2", FromPoint: "P2", ToPoint: "P3", Length: 1, Direction: domain.SegmentDirectionBidirectional, SpeedLimit: 1, RouteCostWeight: 1},
				{SegmentID: "S3", FromPoint: "P3", ToPoint: "P4", Length: 1, Direction: domain.SegmentDirectionBidirectional, SpeedLimit: 1, RouteCostWeight: 1},
			},
		},
		domain.Scenario{
			Name: "Mini Scenario",
			Robots: []domain.RobotSpec{
				{RobotID: "R1", RobotModel: "AMR-Lite", StartPoint: "P1", BatteryLevel: 100},
			},
		},
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}
	return engine
}
