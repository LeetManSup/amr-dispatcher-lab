package simulation

import "amr-dispatcher-lab/internal/domain"

func (e *Engine) buildCandidates(tick int, snapshot domain.SystemSnapshot) []domain.TaskCandidate {
	candidates := make([]domain.TaskCandidate, 0)
	for _, task := range e.tasks {
		if task.TaskStatus != domain.TaskStatusQueued && task.TaskStatus != domain.TaskStatusPaused {
			continue
		}
		if task.DeferUntilTick > tick {
			continue
		}
		req := e.requests[task.RequestID]
		if req == nil {
			continue
		}
		path, err := e.planner.BuildPath(task.SourcePoint, task.TargetPoint)
		if err != nil {
			e.decisionLogs = append(e.decisionLogs, domain.DecisionLog{
				RunID:      e.run.ID,
				Tick:       tick,
				RequestID:  req.RequestID,
				TaskID:     task.TaskID,
				Algorithm:  e.run.Algorithm,
				Action:     domain.DecisionActionReject,
				ReasonCode: "unreachable_path",
			})
			continue
		}
		zoneLoad := 0.0
		if point := e.pointByID(task.TargetPoint); point != nil {
			zoneLoad = snapshot.ZoneLoad[point.AreaID]
		}
		candidates = append(candidates, domain.TaskCandidate{
			Request:       req,
			Task:          task,
			RouteCost:     path.Cost,
			RouteLoad:     averagePathLoad(path.Points, snapshot.SegmentLoad, e.findSegment),
			ZoneLoad:      zoneLoad,
			SystemRisk:    e.systemRisk(snapshot),
			Urgency:       computeUrgency(tick, req.Deadline),
			PriorityScore: normalizePriority(req.Priority),
		})
	}
	return candidates
}

func (e *Engine) applySchedulerDecisions(tick int, snapshot domain.SystemSnapshot, decisions []domain.Decision, candidates []domain.TaskCandidate) {
	candidateIndex := make(map[string]domain.TaskCandidate, len(candidates))
	for _, candidate := range candidates {
		candidateIndex[candidate.Task.TaskID] = candidate
	}
	availableRobots := e.availableRobots()
	for _, decision := range decisions {
		candidate, ok := candidateIndex[decision.TaskID]
		if !ok {
			continue
		}
		if decision.Action == domain.DecisionActionDefer {
			e.decisionLogs = append(e.decisionLogs, domain.DecisionLog{
				RunID:          e.run.ID,
				Tick:           tick,
				RequestID:      decision.RequestID,
				TaskID:         decision.TaskID,
				Algorithm:      e.run.Algorithm,
				Action:         decision.Action,
				ReasonCode:     decision.ReasonCode,
				Score:          decision.Score,
				DeferUntilTick: decision.DeferUntilTick,
			})
			continue
		}
		if err := e.admission.Allow(snapshot, candidate, len(availableRobots)); err != nil {
			candidate.Task.DeferUntilTick = tick + 1
			e.decisionLogs = append(e.decisionLogs, domain.DecisionLog{
				RunID:          e.run.ID,
				Tick:           tick,
				RequestID:      decision.RequestID,
				TaskID:         decision.TaskID,
				Algorithm:      e.run.Algorithm,
				Action:         domain.DecisionActionDefer,
				ReasonCode:     err.Error(),
				Score:          decision.Score,
				DeferUntilTick: tick + 1,
			})
			continue
		}
		robot, path, err := e.selectRobotAndPath(candidate.Task)
		if err != nil {
			candidate.Task.DeferUntilTick = tick + 1
			e.decisionLogs = append(e.decisionLogs, domain.DecisionLog{
				RunID:          e.run.ID,
				Tick:           tick,
				RequestID:      decision.RequestID,
				TaskID:         decision.TaskID,
				Algorithm:      e.run.Algorithm,
				Action:         domain.DecisionActionDefer,
				ReasonCode:     "robot_assignment_failed",
				Score:          decision.Score,
				DeferUntilTick: tick + 1,
			})
			continue
		}
		if err := e.dispatchTask(tick, robot, candidate.Task, path); err != nil {
			e.decisionLogs = append(e.decisionLogs, domain.DecisionLog{
				RunID:      e.run.ID,
				Tick:       tick,
				RequestID:  decision.RequestID,
				TaskID:     decision.TaskID,
				Algorithm:  e.run.Algorithm,
				Action:     domain.DecisionActionFail,
				ReasonCode: "dispatch_failed",
				Score:      decision.Score,
			})
			continue
		}
		availableRobots = e.availableRobots()
		e.decisionLogs = append(e.decisionLogs, domain.DecisionLog{
			RunID:      e.run.ID,
			Tick:       tick,
			RequestID:  decision.RequestID,
			TaskID:     decision.TaskID,
			RobotID:    robot.RobotID,
			Algorithm:  e.run.Algorithm,
			Action:     domain.DecisionActionDispatch,
			ReasonCode: "admitted",
			Score:      decision.Score,
		})
	}
}

func (e *Engine) dispatchTask(tick int, robot *domain.Robot, task *domain.TaskOrder, path domain.PathResult) error {
	if err := robot.Transition(domain.RobotStateBusy); err != nil {
		return err
	}
	robot.CurrentTaskID = task.TaskID
	task.RobotID = robot.RobotID
	task.AssignedRobotPoint = robot.CurrentPoint
	task.PlannedPath = path.Points
	task.PathIndex = 0
	task.EstimatedCost = path.Cost
	task.EstimatedDuration = path.ETA
	task.LastProgressTick = tick
	if robot.CurrentPoint == task.SourcePoint {
		task.CurrentTargetPoint = task.TargetPoint
	} else {
		task.CurrentTargetPoint = task.SourcePoint
	}
	if task.TaskStatus == domain.TaskStatusPaused {
		if err := task.Transition(domain.TaskStatusInProgress); err != nil {
			return err
		}
	} else {
		if err := task.Transition(domain.TaskStatusAssigned); err != nil {
			return err
		}
		e.appendTaskHistory(task.TaskID, task.TaskStatus, "assigned")
		if err := task.Transition(domain.TaskStatusInProgress); err != nil {
			return err
		}
	}
	task.StartedAt = chooseStartTick(task.StartedAt, tick)
	task.DeferUntilTick = 0
	e.appendTaskHistory(task.TaskID, task.TaskStatus, "dispatched")
	return nil
}

func (e *Engine) runSupervision(tick int) {
	timeout := e.admission.Config.TaskTimeoutTicks
	if timeout == 0 {
		timeout = 12
	}
	stall := e.admission.Config.StallThresholdTicks
	if stall == 0 {
		stall = 4
	}
	for _, task := range e.tasks {
		if task.TaskStatus != domain.TaskStatusInProgress && task.TaskStatus != domain.TaskStatusAssigned {
			continue
		}
		if tick-task.StartedAt > timeout {
			e.failTask(task, tick, "timeout")
			continue
		}
		if tick-task.LastProgressTick > stall {
			e.failTask(task, tick, "stalled")
			continue
		}
		robot := e.robots[task.RobotID]
		if robot != nil && robot.State == domain.RobotStateFault {
			e.failTask(task, tick, "robot_fault")
		}
	}
	for _, robot := range e.robots {
		if robot.State == domain.RobotStateBusy && robot.CurrentTaskID == "" {
			_ = robot.Transition(domain.RobotStateIdle)
		}
	}
}

func (e *Engine) failTask(task *domain.TaskOrder, tick int, reason string) {
	if task.TaskStatus == domain.TaskStatusFailed || task.TaskStatus == domain.TaskStatusCancelled || task.TaskStatus == domain.TaskStatusCompleted {
		return
	}
	_ = task.Transition(domain.TaskStatusFailed)
	task.FinishedAt = tick
	e.appendTaskHistory(task.TaskID, task.TaskStatus, reason)
	e.decisionLogs = append(e.decisionLogs, domain.DecisionLog{
		RunID:      e.run.ID,
		Tick:       tick,
		RequestID:  task.RequestID,
		TaskID:     task.TaskID,
		RobotID:    task.RobotID,
		Algorithm:  e.run.Algorithm,
		Action:     domain.DecisionActionFail,
		ReasonCode: reason,
	})
	e.releaseRobot(task.RobotID)
}

func (e *Engine) shouldFinish() bool {
	if e.scenario.MaxTicks > 0 && e.run.CurrentTick >= e.scenario.MaxTicks {
		return true
	}
	hasFutureRequests := false
	for _, plan := range e.scenario.Requests {
		if !e.released[plan.RequestID] {
			hasFutureRequests = true
			break
		}
	}
	for _, task := range e.tasks {
		switch task.TaskStatus {
		case domain.TaskStatusQueued, domain.TaskStatusAssigned, domain.TaskStatusInProgress, domain.TaskStatusPaused:
			return false
		}
	}
	return !hasFutureRequests
}

func (e *Engine) completeRun(status domain.RunStatus, lastErr string) {
	e.run.Status = status
	e.run.LastError = lastErr
	now := e.clock.Now()
	e.run.FinishedAt = &now
}
