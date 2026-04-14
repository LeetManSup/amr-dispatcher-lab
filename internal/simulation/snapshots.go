package simulation

import "amr-dispatcher-lab/internal/domain"

func (e *Engine) computeLoads() (map[string]float64, map[string]float64) {
	segmentLoad := make(map[string]float64)
	zoneLoad := make(map[string]float64)
	points := make(map[string]domain.MapPoint, len(e.mapData.Points))
	for _, point := range e.mapData.Points {
		points[point.PointID] = point
	}
	for _, task := range e.tasks {
		if task.TaskStatus != domain.TaskStatusInProgress && task.TaskStatus != domain.TaskStatusAssigned {
			continue
		}
		if task.PathIndex+1 < len(task.PlannedPath) {
			from := task.PlannedPath[task.PathIndex]
			to := task.PlannedPath[task.PathIndex+1]
			if segmentID := e.findSegment(from, to); segmentID != "" {
				segmentLoad[segmentID]++
			}
		}
		if point, ok := points[task.CurrentTargetPoint]; ok {
			zoneLoad[point.AreaID]++
		}
	}
	return segmentLoad, zoneLoad
}

func (e *Engine) buildSnapshot(tick int, segmentLoad, zoneLoad map[string]float64) domain.SystemSnapshot {
	snapshot := domain.SystemSnapshot{
		Tick:             tick,
		SegmentLoad:      segmentLoad,
		ZoneLoad:         zoneLoad,
		RobotUtilization: e.robotUtilization(float64(tick + 1)),
	}
	for _, task := range e.tasks {
		switch task.TaskStatus {
		case domain.TaskStatusQueued:
			snapshot.WaitingTasks++
		case domain.TaskStatusAssigned, domain.TaskStatusInProgress:
			snapshot.ActiveTasks++
		case domain.TaskStatusPaused:
			snapshot.PausedTasks++
		}
	}
	for _, robot := range e.robots {
		switch robot.State {
		case domain.RobotStateIdle:
			snapshot.AvailableRobots++
		case domain.RobotStateFault:
			snapshot.FaultRobots++
		case domain.RobotStateCharging:
			snapshot.ChargingRobots++
		}
	}
	return snapshot
}

func (e *Engine) pointByID(id string) *domain.MapPoint {
	for _, point := range e.mapData.Points {
		if point.PointID == id {
			copy := point
			return &copy
		}
	}
	return nil
}

func (e *Engine) findSegment(from, to string) string {
	for _, segment := range e.mapData.Segments {
		if segment.FromPoint == from && segment.ToPoint == to {
			return segment.SegmentID
		}
		if segment.Direction == domain.SegmentDirectionBidirectional && segment.FromPoint == to && segment.ToPoint == from {
			return segment.SegmentID
		}
		if segment.Direction == domain.SegmentDirectionReverse && segment.ToPoint == from && segment.FromPoint == to {
			return segment.SegmentID
		}
	}
	return ""
}
