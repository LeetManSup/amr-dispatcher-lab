package domain

import "time"

type AdmissionConfig struct {
	MaxActiveTasks      int     `json:"maxActiveTasks"`
	MaxRouteLoad        float64 `json:"maxRouteLoad"`
	MaxZoneLoad         float64 `json:"maxZoneLoad"`
	TaskTimeoutTicks    int     `json:"taskTimeoutTicks"`
	StallThresholdTicks int     `json:"stallThresholdTicks"`
}

type SchedulerWeights struct {
	Alpha float64 `json:"alpha"`
	Beta  float64 `json:"beta"`
	W1    float64 `json:"w1"`
	W2    float64 `json:"w2"`
	W3    float64 `json:"w3"`
	W4    float64 `json:"w4"`
	W5    float64 `json:"w5"`
}

type RequestPlan struct {
	RequestID    string `json:"requestId"`
	ReleaseTick  int    `json:"releaseTick"`
	SourcePoint  string `json:"sourcePoint"`
	TargetPoint  string `json:"targetPoint"`
	BusinessType string `json:"businessType"`
	Priority     int    `json:"priority"`
	Deadline     int    `json:"deadline"`
}

type RobotSpec struct {
	RobotID      string  `json:"robotId"`
	RobotModel   string  `json:"robotModel"`
	StartPoint   string  `json:"startPoint"`
	BatteryLevel float64 `json:"batteryLevel"`
}

type Scenario struct {
	Name                  string           `json:"name"`
	Description           string           `json:"description"`
	Robots                []RobotSpec      `json:"robots"`
	Requests              []RequestPlan    `json:"requests"`
	FaultProbability      float64          `json:"faultProbability"`
	ChargingProbability   float64          `json:"chargingProbability"`
	LowBatteryThreshold   float64          `json:"lowBatteryThreshold"`
	ChargeRecoveryPerTick float64          `json:"chargeRecoveryPerTick"`
	BatteryDrainPerTick   float64          `json:"batteryDrainPerTick"`
	MaxTicks              int              `json:"maxTicks"`
	Admission             AdmissionConfig  `json:"admission"`
	Weights               SchedulerWeights `json:"weights"`
}

type RunConfig struct {
	Algorithm          Algorithm        `json:"algorithm"`
	MapPath            string           `json:"mapPath"`
	ScenarioPath       string           `json:"scenarioPath"`
	Seed               int64            `json:"seed"`
	TickDurationMillis int              `json:"tickDurationMillis"`
	Admission          AdmissionConfig  `json:"admission"`
	Weights            SchedulerWeights `json:"weights"`
	PersistenceEnabled bool             `json:"persistenceEnabled"`
	ExportFormats      []string         `json:"exportFormats"`
}

type Run struct {
	ID           string     `json:"id"`
	Status       RunStatus  `json:"status"`
	Mode         RunMode    `json:"mode"`
	CurrentTick  int        `json:"currentTick"`
	Algorithm    Algorithm  `json:"algorithm"`
	MapName      string     `json:"mapName"`
	MapPath      string     `json:"mapPath"`
	ScenarioName string     `json:"scenarioName"`
	ScenarioPath string     `json:"scenarioPath"`
	Seed         int64      `json:"seed"`
	ConfigHash   string     `json:"configHash"`
	CreatedAt    time.Time  `json:"createdAt"`
	StartedAt    *time.Time `json:"startedAt,omitempty"`
	FinishedAt   *time.Time `json:"finishedAt,omitempty"`
	LastError    string     `json:"lastError,omitempty"`
}

type SystemSnapshot struct {
	Tick             int                `json:"tick"`
	ActiveTasks      int                `json:"activeTasks"`
	WaitingTasks     int                `json:"waitingTasks"`
	AvailableRobots  int                `json:"availableRobots"`
	FaultRobots      int                `json:"faultRobots"`
	ChargingRobots   int                `json:"chargingRobots"`
	PausedTasks      int                `json:"pausedTasks"`
	SegmentLoad      map[string]float64 `json:"segmentLoad"`
	ZoneLoad         map[string]float64 `json:"zoneLoad"`
	RobotUtilization float64            `json:"robotUtilization"`
}

type MetricsSnapshot struct {
	RunID               string  `json:"runId"`
	Tick                int     `json:"tick"`
	AvgWaitTime         float64 `json:"avgWaitTime"`
	AvgExecutionTime    float64 `json:"avgExecutionTime"`
	Throughput          float64 `json:"throughput"`
	CancelRate          float64 `json:"cancelRate"`
	DeadlineSuccessRate float64 `json:"deadlineSuccessRate"`
	SegmentLoadVariance float64 `json:"segmentLoadVariance"`
	RobotUtilization    float64 `json:"robotUtilization"`
	CompletedTasks      int     `json:"completedTasks"`
	CancelledTasks      int     `json:"cancelledTasks"`
	FailedTasks         int     `json:"failedTasks"`
}

type DecisionLog struct {
	RunID          string         `json:"runId"`
	Tick           int            `json:"tick"`
	RequestID      string         `json:"requestId"`
	TaskID         string         `json:"taskId"`
	RobotID        string         `json:"robotId"`
	Algorithm      Algorithm      `json:"algorithm"`
	Action         DecisionAction `json:"action"`
	ReasonCode     string         `json:"reasonCode"`
	Score          float64        `json:"score"`
	DeferUntilTick int            `json:"deferUntilTick"`
}

type TaskStatusHistory struct {
	RunID  string     `json:"runId"`
	TaskID string     `json:"taskId"`
	Tick   int        `json:"tick"`
	Status TaskStatus `json:"status"`
	Reason string     `json:"reason"`
}

type SegmentLoadSnapshot struct {
	RunID     string  `json:"runId"`
	Tick      int     `json:"tick"`
	SegmentID string  `json:"segmentId"`
	Load      float64 `json:"load"`
}

type PathResult struct {
	Points []string `json:"points"`
	Cost   float64  `json:"cost"`
	ETA    int      `json:"eta"`
}

type Decision struct {
	Action             DecisionAction `json:"action"`
	RequestID          string         `json:"requestId"`
	TaskID             string         `json:"taskId"`
	RobotID            string         `json:"robotId"`
	Score              float64        `json:"score"`
	ReasonCode         string         `json:"reasonCode"`
	DeferUntilTick     int            `json:"deferUntilTick"`
	UpdatedTargetPoint string         `json:"updatedTargetPoint,omitempty"`
}

type TaskCandidate struct {
	Request       *Request
	Task          *TaskOrder
	RouteCost     float64
	RouteLoad     float64
	ZoneLoad      float64
	SystemRisk    float64
	Urgency       float64
	PriorityScore float64
}
