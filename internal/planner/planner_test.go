package planner

import (
	"amr-dispatcher-lab/internal/domain"
	"testing"
)

func TestBuildPathRespectsDirection(t *testing.T) {
	graph, err := NewGraphPlanner(domain.MapData{
		Name: "test",
		Points: []domain.MapPoint{
			{PointID: "A", PointType: domain.PointTypeRunning},
			{PointID: "B", PointType: domain.PointTypeRunning},
			{PointID: "C", PointType: domain.PointTypeRunning},
		},
		Segments: []domain.MapSegment{
			{SegmentID: "S1", FromPoint: "A", ToPoint: "B", Length: 1, SpeedLimit: 1, RouteCostWeight: 1, Direction: domain.SegmentDirectionForward},
			{SegmentID: "S2", FromPoint: "B", ToPoint: "C", Length: 1, SpeedLimit: 1, RouteCostWeight: 1, Direction: domain.SegmentDirectionBidirectional},
		},
	})
	if err != nil {
		t.Fatalf("new graph planner: %v", err)
	}

	if _, err := graph.BuildPath("B", "A"); err == nil {
		t.Fatal("expected reverse path over forward-only edge to fail")
	}
	path, err := graph.BuildPath("A", "C")
	if err != nil {
		t.Fatalf("build path: %v", err)
	}
	if len(path.Points) != 3 {
		t.Fatalf("expected three points in path, got %v", path.Points)
	}
}
