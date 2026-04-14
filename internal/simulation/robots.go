package simulation

import (
	"amr-dispatcher-lab/internal/domain"
	"errors"
	"math"
	"sort"
)

func (e *Engine) updateRobots() {
	threshold := e.scenario.LowBatteryThreshold
	if threshold == 0 {
		threshold = 25
	}
	chargeRecovery := e.scenario.ChargeRecoveryPerTick
	if chargeRecovery == 0 {
		chargeRecovery = 10
	}
	drain := e.scenario.BatteryDrainPerTick
	if drain == 0 {
		drain = 3
	}
	for _, robot := range e.robots {
		switch robot.State {
		case domain.RobotStateBusy:
			robot.BusyTicks++
			robot.BatteryLevel = math.Max(0, robot.BatteryLevel-drain)
			if e.scenario.FaultProbability > 0 && e.random.Float64() < e.scenario.FaultProbability {
				_ = robot.Transition(domain.RobotStateFault)
				robot.FaultFlag = true
			}
		case domain.RobotStateCharging:
			robot.BatteryLevel = math.Min(100, robot.BatteryLevel+chargeRecovery)
			if robot.BatteryLevel >= 95 {
				_ = robot.Transition(domain.RobotStateIdle)
			}
		case domain.RobotStateIdle:
			if robot.BatteryLevel <= threshold {
				_ = robot.Transition(domain.RobotStateCharging)
			} else if e.scenario.ChargingProbability > 0 && e.random.Float64() < e.scenario.ChargingProbability {
				_ = robot.Transition(domain.RobotStateCharging)
			}
		case domain.RobotStateFault:
			if e.random.Float64() > 0.5 {
				_ = robot.Transition(domain.RobotStateIdle)
				robot.FaultFlag = false
			}
		}
	}
}

func (e *Engine) progressTasks(tick int) {
	for _, task := range e.tasks {
		if task.TaskStatus != domain.TaskStatusInProgress {
			continue
		}
		robot := e.robots[task.RobotID]
		if robot == nil || robot.State == domain.RobotStateFault {
			continue
		}
		if task.PathIndex+1 < len(task.PlannedPath) {
			task.PathIndex++
			task.LastProgressTick = tick
			robot.CurrentPoint = task.PlannedPath[task.PathIndex]
			if task.CurrentTargetPoint == task.SourcePoint && robot.CurrentPoint == task.SourcePoint {
				task.CurrentTargetPoint = task.TargetPoint
			}
		}
	}
}

func (e *Engine) finishTasks(tick int) {
	for _, task := range e.tasks {
		if task.TaskStatus != domain.TaskStatusInProgress {
			continue
		}
		if len(task.PlannedPath) == 0 || task.PathIndex < len(task.PlannedPath)-1 {
			continue
		}
		if err := task.Transition(domain.TaskStatusCompleted); err != nil {
			continue
		}
		task.FinishedAt = tick
		e.appendTaskHistory(task.TaskID, task.TaskStatus, "completed")
		if req := e.requests[task.RequestID]; req != nil {
			_ = req.Transition(domain.RequestStatusCompleted)
		}
		e.releaseRobot(task.RobotID)
	}
}

func (e *Engine) availableRobots() []*domain.Robot {
	items := make([]*domain.Robot, 0)
	for _, robot := range e.robots {
		if robot.State == domain.RobotStateIdle {
			items = append(items, robot)
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].RobotID < items[j].RobotID })
	return items
}

func (e *Engine) releaseRobot(robotID string) {
	if robotID == "" {
		return
	}
	robot := e.robots[robotID]
	if robot == nil {
		return
	}
	robot.CurrentTaskID = ""
	if robot.State != domain.RobotStateFault && robot.State != domain.RobotStateCharging {
		_ = robot.Transition(domain.RobotStateIdle)
	}
}

func (e *Engine) selectRobotAndPath(task *domain.TaskOrder) (*domain.Robot, domain.PathResult, error) {
	type candidate struct {
		robot *domain.Robot
		path  domain.PathResult
		cost  float64
	}
	options := make([]candidate, 0)
	for _, robot := range e.availableRobots() {
		path, _, err := e.buildExecutionPath(robot.CurrentPoint, task.SourcePoint, task.TargetPoint, false)
		if err != nil {
			continue
		}
		options = append(options, candidate{robot: robot, path: path, cost: path.Cost})
	}
	if len(options) == 0 {
		return nil, domain.PathResult{}, errors.New("no route from available robots")
	}
	sort.Slice(options, func(i, j int) bool {
		if options[i].cost == options[j].cost {
			return options[i].robot.RobotID < options[j].robot.RobotID
		}
		return options[i].cost < options[j].cost
	})
	return options[0].robot, options[0].path, nil
}

func (e *Engine) buildExecutionPath(startPoint, sourcePoint, targetPoint string, sourceReached bool) (domain.PathResult, string, error) {
	if sourceReached || startPoint == sourcePoint {
		path, err := e.planner.BuildPath(startPoint, targetPoint)
		if err != nil {
			return domain.PathResult{}, "", err
		}
		return path, targetPoint, nil
	}

	toSource, err := e.planner.BuildPath(startPoint, sourcePoint)
	if err != nil {
		return domain.PathResult{}, "", err
	}
	sourceToTarget, err := e.planner.BuildPath(sourcePoint, targetPoint)
	if err != nil {
		return domain.PathResult{}, "", err
	}

	points := append([]string(nil), toSource.Points...)
	if len(sourceToTarget.Points) > 1 {
		points = append(points, sourceToTarget.Points[1:]...)
	}

	return domain.PathResult{
		Points: points,
		Cost:   toSource.Cost + sourceToTarget.Cost,
		ETA:    toSource.ETA + sourceToTarget.ETA,
	}, sourcePoint, nil
}
