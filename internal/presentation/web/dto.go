package web

type createRunRequest struct {
	Algorithm    string `json:"algorithm"`
	MapPath      string `json:"mapPath"`
	ScenarioPath string `json:"scenarioPath"`
	Seed         int64  `json:"seed"`
}

type addTaskRequest struct {
	RequestID    string `json:"requestId"`
	SourcePoint  string `json:"sourcePoint"`
	TargetPoint  string `json:"targetPoint"`
	BusinessType string `json:"businessType"`
	Priority     int    `json:"priority"`
	CreatedAt    int    `json:"createdAt"`
	Deadline     int    `json:"deadline"`
}

type updateOrderPointRequest struct {
	TargetPoint string `json:"targetPoint"`
}

type catalogItemResponse struct {
	Path        string `json:"path"`
	Label       string `json:"label"`
	Kind        string `json:"kind"`
	Description string `json:"description,omitempty"`
}

type catalogResponse struct {
	Maps          []string              `json:"maps"`
	Scenarios     []string              `json:"scenarios"`
	MapItems      []catalogItemResponse `json:"mapItems,omitempty"`
	ScenarioItems []catalogItemResponse `json:"scenarioItems,omitempty"`
}

type runResponse struct {
	ID           string  `json:"id"`
	Status       string  `json:"status"`
	Mode         string  `json:"mode"`
	CurrentTick  int     `json:"currentTick"`
	Algorithm    string  `json:"algorithm"`
	MapName      string  `json:"mapName"`
	MapPath      string  `json:"mapPath"`
	ScenarioName string  `json:"scenarioName"`
	ScenarioPath string  `json:"scenarioPath"`
	Seed         int64   `json:"seed"`
	ConfigHash   string  `json:"configHash"`
	CreatedAt    string  `json:"createdAt"`
	StartedAt    *string `json:"startedAt,omitempty"`
	FinishedAt   *string `json:"finishedAt,omitempty"`
	LastError    string  `json:"lastError,omitempty"`
}

type taskResponse struct {
	TaskID             string   `json:"taskId"`
	RequestID          string   `json:"requestId"`
	RobotID            string   `json:"robotId"`
	PlannedPath        []string `json:"plannedPath"`
	CurrentTargetPoint string   `json:"currentTargetPoint"`
	TaskStatus         string   `json:"taskStatus"`
	CreatedAt          int      `json:"createdAt"`
	StartedAt          int      `json:"startedAt"`
	FinishedAt         int      `json:"finishedAt"`
	SourcePoint        string   `json:"sourcePoint"`
	TargetPoint        string   `json:"targetPoint"`
	PathIndex          int      `json:"pathIndex"`
	EstimatedCost      float64  `json:"estimatedCost"`
	EstimatedDuration  int      `json:"estimatedDuration"`
	LastProgressTick   int      `json:"lastProgressTick"`
	DeferUntilTick     int      `json:"deferUntilTick"`
	AssignedRobotPoint string   `json:"assignedRobotPoint"`
}

type requestResponse struct {
	RequestID    string `json:"requestId"`
	SourcePoint  string `json:"sourcePoint"`
	TargetPoint  string `json:"targetPoint"`
	BusinessType string `json:"businessType"`
	Priority     int    `json:"priority"`
	CreatedAt    int    `json:"createdAt"`
	Deadline     int    `json:"deadline"`
	Status       string `json:"status"`
}

type robotResponse struct {
	RobotID       string  `json:"robotId"`
	RobotModel    string  `json:"robotModel"`
	State         string  `json:"state"`
	CurrentPoint  string  `json:"currentPoint"`
	BatteryLevel  float64 `json:"batteryLevel"`
	CurrentTaskID string  `json:"currentTaskId"`
	FaultFlag     bool    `json:"faultFlag"`
	BusyTicks     int     `json:"busyTicks"`
}

type metricsResponse struct {
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

type decisionLogResponse struct {
	RunID          string  `json:"runId"`
	Tick           int     `json:"tick"`
	RequestID      string  `json:"requestId"`
	TaskID         string  `json:"taskId"`
	RobotID        string  `json:"robotId"`
	Algorithm      string  `json:"algorithm"`
	Action         string  `json:"action"`
	ReasonCode     string  `json:"reasonCode"`
	Score          float64 `json:"score"`
	DeferUntilTick int     `json:"deferUntilTick"`
}

type segmentLoadResponse struct {
	RunID     string  `json:"runId"`
	Tick      int     `json:"tick"`
	SegmentID string  `json:"segmentId"`
	Load      float64 `json:"load"`
}

type mapPointResponse struct {
	PointID   string  `json:"pointId"`
	PointType string  `json:"pointType"`
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	AreaID    string  `json:"areaId"`
}

type mapSegmentResponse struct {
	SegmentID       string  `json:"segmentId"`
	FromPoint       string  `json:"fromPoint"`
	ToPoint         string  `json:"toPoint"`
	Length          float64 `json:"length"`
	Direction       string  `json:"direction"`
	SpeedLimit      float64 `json:"speedLimit"`
	RouteCostWeight float64 `json:"routeCostWeight"`
	SegmentType     string  `json:"segmentType"`
}

type mapResponse struct {
	Name     string               `json:"name"`
	Points   []mapPointResponse   `json:"points"`
	Segments []mapSegmentResponse `json:"segments"`
	Zones    map[string]any       `json:"zones,omitempty"`
}

type snapshotResponse struct {
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

type statusHistoryResponse struct {
	RunID  string `json:"runId"`
	TaskID string `json:"taskId"`
	Tick   int    `json:"tick"`
	Status string `json:"status"`
	Reason string `json:"reason"`
}

type tickResultResponse struct {
	Snapshot      snapshotResponse        `json:"snapshot"`
	Metrics       metricsResponse         `json:"metrics"`
	Decisions     []decisionLogResponse   `json:"decisions"`
	StatusHistory []statusHistoryResponse `json:"statusHistory"`
	SegmentLoads  []segmentLoadResponse   `json:"segmentLoads"`
}

type runOverviewResponse struct {
	Run             runResponse           `json:"run"`
	Snapshot        snapshotResponse      `json:"snapshot"`
	LatestMetrics   *metricsResponse      `json:"latestMetrics,omitempty"`
	TaskCounts      map[string]int        `json:"taskCounts"`
	RobotCounts     map[string]int        `json:"robotCounts"`
	RecentDecisions []decisionLogResponse `json:"recentDecisions"`
	Alerts          []string              `json:"alerts,omitempty"`
}

type liveEventResponse struct {
	EventID      string `json:"eventId"`
	RunID        string `json:"runId"`
	Kind         string `json:"kind"`
	Message      string `json:"message"`
	CurrentTick  int    `json:"currentTick"`
	RunStatus    string `json:"runStatus"`
	ActiveTasks  int    `json:"activeTasks"`
	WaitingTasks int    `json:"waitingTasks"`
	OccurredAt   string `json:"occurredAt"`
}

type comparisonResponse struct {
	RunID       string  `json:"runId"`
	Algorithm   string  `json:"algorithm"`
	Scenario    string  `json:"scenario"`
	Throughput  float64 `json:"throughput"`
	WaitTime    float64 `json:"waitTime"`
	Utilization float64 `json:"utilization"`
}
