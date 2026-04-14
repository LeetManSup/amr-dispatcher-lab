package domain

import "math"

type Request struct {
	RequestID    string        `json:"requestId"`
	SourcePoint  string        `json:"sourcePoint"`
	TargetPoint  string        `json:"targetPoint"`
	BusinessType string        `json:"businessType"`
	Priority     int           `json:"priority"`
	CreatedAt    int           `json:"createdAt"`
	Deadline     int           `json:"deadline"`
	Status       RequestStatus `json:"status"`
}

type TaskOrder struct {
	TaskID             string     `json:"taskId"`
	RequestID          string     `json:"requestId"`
	RobotID            string     `json:"robotId"`
	PlannedPath        []string   `json:"plannedPath"`
	CurrentTargetPoint string     `json:"currentTargetPoint"`
	TaskStatus         TaskStatus `json:"taskStatus"`
	CreatedAt          int        `json:"createdAt"`
	StartedAt          int        `json:"startedAt"`
	FinishedAt         int        `json:"finishedAt"`
	SourcePoint        string     `json:"sourcePoint"`
	TargetPoint        string     `json:"targetPoint"`
	PathIndex          int        `json:"pathIndex"`
	EstimatedCost      float64    `json:"estimatedCost"`
	EstimatedDuration  int        `json:"estimatedDuration"`
	LastProgressTick   int        `json:"lastProgressTick"`
	DeferUntilTick     int        `json:"deferUntilTick"`
	AssignedRobotPoint string     `json:"assignedRobotPoint"`
}

type Robot struct {
	RobotID       string     `json:"robotId"`
	RobotModel    string     `json:"robotModel"`
	State         RobotState `json:"state"`
	CurrentPoint  string     `json:"currentPoint"`
	BatteryLevel  float64    `json:"batteryLevel"`
	CurrentTaskID string     `json:"currentTaskId"`
	FaultFlag     bool       `json:"faultFlag"`
	BusyTicks     int        `json:"busyTicks"`
}

type MapPoint struct {
	PointID   string    `json:"pointId"`
	PointType PointType `json:"pointType"`
	X         float64   `json:"x"`
	Y         float64   `json:"y"`
	AreaID    string    `json:"areaId"`
}

type MapSegment struct {
	SegmentID       string           `json:"segmentId"`
	FromPoint       string           `json:"fromPoint"`
	ToPoint         string           `json:"toPoint"`
	Length          float64          `json:"length"`
	Direction       SegmentDirection `json:"direction"`
	SpeedLimit      float64          `json:"speedLimit"`
	RouteCostWeight float64          `json:"routeCostWeight"`
	SegmentType     string           `json:"segmentType"`
}

func (s MapSegment) Cost() float64 {
	speed := math.Max(s.SpeedLimit, 0.0001)
	weight := math.Max(s.RouteCostWeight, 0.0001)
	return s.Length * weight / speed
}

type MapData struct {
	Name     string                 `json:"name"`
	Points   []MapPoint             `json:"points"`
	Segments []MapSegment           `json:"segments"`
	Zones    map[string]interface{} `json:"zones,omitempty"`
}
