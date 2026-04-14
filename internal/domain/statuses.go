package domain

type RequestStatus string

const (
	RequestStatusCreated         RequestStatus = "created"
	RequestStatusValidated       RequestStatus = "validated"
	RequestStatusQueued          RequestStatus = "queued"
	RequestStatusConvertedToTask RequestStatus = "converted_to_task"
	RequestStatusCompleted       RequestStatus = "completed"
	RequestStatusRejected        RequestStatus = "rejected"
	RequestStatusCancelled       RequestStatus = "cancelled"
)

type TaskStatus string

const (
	TaskStatusCreated    TaskStatus = "created"
	TaskStatusQueued     TaskStatus = "queued"
	TaskStatusAssigned   TaskStatus = "assigned"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusPaused     TaskStatus = "paused"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusCancelled  TaskStatus = "cancelled"
	TaskStatusFailed     TaskStatus = "failed"
)

type RobotState string

const (
	RobotStateIdle     RobotState = "idle"
	RobotStateBusy     RobotState = "busy"
	RobotStateCharging RobotState = "charging"
	RobotStateFault    RobotState = "fault"
	RobotStatePaused   RobotState = "paused"
)

type RunStatus string

const (
	RunStatusCreated   RunStatus = "created"
	RunStatusRunning   RunStatus = "running"
	RunStatusStopped   RunStatus = "stopped"
	RunStatusCompleted RunStatus = "completed"
	RunStatusFailed    RunStatus = "failed"
)

type DecisionAction string

const (
	DecisionActionDispatch DecisionAction = "dispatch"
	DecisionActionDefer    DecisionAction = "defer"
	DecisionActionCancel   DecisionAction = "cancel"
	DecisionActionContinue DecisionAction = "continue"
	DecisionActionRetarget DecisionAction = "retarget"
	DecisionActionReject   DecisionAction = "reject"
	DecisionActionFail     DecisionAction = "fail"
)

type Algorithm string

const (
	AlgorithmFIFO     Algorithm = "fifo"
	AlgorithmPriority Algorithm = "priority"
	AlgorithmAdaptive Algorithm = "adaptive"
)

type PointType string

const (
	PointTypeRunning     PointType = "running_area"
	PointTypeQueue       PointType = "queue_area"
	PointTypeWorkbench   PointType = "workbench"
	PointTypeShelf       PointType = "shelf_area"
	PointTypeCharging    PointType = "charging_area"
	PointTypeMaintenance PointType = "maintenance_area"
	PointTypeProhibited  PointType = "prohibited_area"
)

type SegmentDirection string

const (
	SegmentDirectionBidirectional SegmentDirection = "bidirectional"
	SegmentDirectionForward       SegmentDirection = "forward"
	SegmentDirectionReverse       SegmentDirection = "reverse"
)

type RunMode string

const (
	RunModeStep       RunMode = "step"
	RunModeContinuous RunMode = "continuous"
)
