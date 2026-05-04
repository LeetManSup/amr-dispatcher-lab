package web

import (
	"time"

	"amr-dispatcher-lab/internal/application"
	"amr-dispatcher-lab/internal/domain"
)

func toCreateRunInput(dto createRunRequest) application.CreateRunInput {
	return application.CreateRunInput{
		Algorithm:    domain.Algorithm(dto.Algorithm),
		MapPath:      dto.MapPath,
		ScenarioPath: dto.ScenarioPath,
		Seed:         dto.Seed,
	}
}

func toAddTaskInput(dto addTaskRequest) application.AddTaskInput {
	return application.AddTaskInput{
		RequestID:    dto.RequestID,
		SourcePoint:  dto.SourcePoint,
		TargetPoint:  dto.TargetPoint,
		BusinessType: dto.BusinessType,
		Priority:     dto.Priority,
		CreatedAt:    dto.CreatedAt,
		Deadline:     dto.Deadline,
	}
}

func toCatalogResponse(item application.Catalog) catalogResponse {
	return catalogResponse{
		Maps:          append([]string(nil), item.Maps...),
		Scenarios:     append([]string(nil), item.Scenarios...),
		MapItems:      toCatalogItemResponses(item.MapItems),
		ScenarioItems: toCatalogItemResponses(item.ScenarioItems),
	}
}

func toCatalogItemResponses(items []application.CatalogItem) []catalogItemResponse {
	out := make([]catalogItemResponse, 0, len(items))
	for _, item := range items {
		out = append(out, catalogItemResponse{
			Path:        item.Path,
			Label:       item.Label,
			Kind:        item.Kind,
			Description: item.Description,
		})
	}
	return out
}

func toRunResponse(run domain.Run) runResponse {
	return runResponse{
		ID:           run.ID,
		Status:       string(run.Status),
		Mode:         string(run.Mode),
		CurrentTick:  run.CurrentTick,
		Algorithm:    string(run.Algorithm),
		MapName:      run.MapName,
		MapPath:      run.MapPath,
		ScenarioName: run.ScenarioName,
		ScenarioPath: run.ScenarioPath,
		Seed:         run.Seed,
		ConfigHash:   run.ConfigHash,
		CreatedAt:    run.CreatedAt.Format(time.RFC3339Nano),
		StartedAt:    formatOptionalTime(run.StartedAt),
		FinishedAt:   formatOptionalTime(run.FinishedAt),
		LastError:    run.LastError,
	}
}

func toRunResponses(items []domain.Run) []runResponse {
	out := make([]runResponse, 0, len(items))
	for _, item := range items {
		out = append(out, toRunResponse(item))
	}
	return out
}

func toRunOverviewResponse(item application.RunOverview) runOverviewResponse {
	response := runOverviewResponse{
		Run:             toRunResponse(item.Run),
		Snapshot:        toSnapshotResponse(item.Snapshot),
		TaskCounts:      copyIntMap(item.TaskCounts),
		RobotCounts:     copyIntMap(item.RobotCounts),
		RecentDecisions: toDecisionResponses(item.RecentDecisions),
		Alerts:          append([]string(nil), item.Alerts...),
	}
	if item.LatestMetrics != nil {
		metrics := toMetricsResponses([]domain.MetricsSnapshot{*item.LatestMetrics})[0]
		response.LatestMetrics = &metrics
	}
	return response
}

func toTaskResponse(task domain.TaskOrder) taskResponse {
	return taskResponse{
		TaskID:             task.TaskID,
		RequestID:          task.RequestID,
		RobotID:            task.RobotID,
		PlannedPath:        append([]string(nil), task.PlannedPath...),
		CurrentTargetPoint: task.CurrentTargetPoint,
		TaskStatus:         string(task.TaskStatus),
		CreatedAt:          task.CreatedAt,
		StartedAt:          task.StartedAt,
		FinishedAt:         task.FinishedAt,
		SourcePoint:        task.SourcePoint,
		TargetPoint:        task.TargetPoint,
		PathIndex:          task.PathIndex,
		EstimatedCost:      task.EstimatedCost,
		EstimatedDuration:  task.EstimatedDuration,
		LastProgressTick:   task.LastProgressTick,
		DeferUntilTick:     task.DeferUntilTick,
		AssignedRobotPoint: task.AssignedRobotPoint,
	}
}

func toTaskResponses(items []domain.TaskOrder) []taskResponse {
	out := make([]taskResponse, 0, len(items))
	for _, item := range items {
		out = append(out, toTaskResponse(item))
	}
	return out
}

func toRequestResponses(items []domain.Request) []requestResponse {
	out := make([]requestResponse, 0, len(items))
	for _, item := range items {
		out = append(out, requestResponse{
			RequestID:    item.RequestID,
			SourcePoint:  item.SourcePoint,
			TargetPoint:  item.TargetPoint,
			BusinessType: item.BusinessType,
			Priority:     item.Priority,
			CreatedAt:    item.CreatedAt,
			Deadline:     item.Deadline,
			Status:       string(item.Status),
		})
	}
	return out
}

func toRobotResponses(items []domain.Robot) []robotResponse {
	out := make([]robotResponse, 0, len(items))
	for _, item := range items {
		out = append(out, robotResponse{
			RobotID:       item.RobotID,
			RobotModel:    item.RobotModel,
			State:         string(item.State),
			CurrentPoint:  item.CurrentPoint,
			BatteryLevel:  item.BatteryLevel,
			CurrentTaskID: item.CurrentTaskID,
			FaultFlag:     item.FaultFlag,
			BusyTicks:     item.BusyTicks,
		})
	}
	return out
}

func toMetricsResponses(items []domain.MetricsSnapshot) []metricsResponse {
	out := make([]metricsResponse, 0, len(items))
	for _, item := range items {
		out = append(out, metricsResponse{
			RunID:               item.RunID,
			Tick:                item.Tick,
			AvgWaitTime:         item.AvgWaitTime,
			AvgExecutionTime:    item.AvgExecutionTime,
			Throughput:          item.Throughput,
			CancelRate:          item.CancelRate,
			DeadlineSuccessRate: item.DeadlineSuccessRate,
			SegmentLoadVariance: item.SegmentLoadVariance,
			RobotUtilization:    item.RobotUtilization,
			CompletedTasks:      item.CompletedTasks,
			CancelledTasks:      item.CancelledTasks,
			FailedTasks:         item.FailedTasks,
		})
	}
	return out
}

func toDecisionResponses(items []domain.DecisionLog) []decisionLogResponse {
	out := make([]decisionLogResponse, 0, len(items))
	for _, item := range items {
		out = append(out, decisionLogResponse{
			RunID:          item.RunID,
			Tick:           item.Tick,
			RequestID:      item.RequestID,
			TaskID:         item.TaskID,
			RobotID:        item.RobotID,
			Algorithm:      string(item.Algorithm),
			Action:         string(item.Action),
			ReasonCode:     item.ReasonCode,
			Score:          item.Score,
			DeferUntilTick: item.DeferUntilTick,
		})
	}
	return out
}

func toSegmentLoadResponses(items []domain.SegmentLoadSnapshot) []segmentLoadResponse {
	out := make([]segmentLoadResponse, 0, len(items))
	for _, item := range items {
		out = append(out, segmentLoadResponse{
			RunID:     item.RunID,
			Tick:      item.Tick,
			SegmentID: item.SegmentID,
			Load:      item.Load,
		})
	}
	return out
}

func toMapResponse(item domain.MapData) mapResponse {
	points := make([]mapPointResponse, 0, len(item.Points))
	for _, point := range item.Points {
		points = append(points, mapPointResponse{
			PointID:   point.PointID,
			PointType: string(point.PointType),
			X:         point.X,
			Y:         point.Y,
			AreaID:    point.AreaID,
		})
	}
	segments := make([]mapSegmentResponse, 0, len(item.Segments))
	for _, segment := range item.Segments {
		segments = append(segments, mapSegmentResponse{
			SegmentID:       segment.SegmentID,
			FromPoint:       segment.FromPoint,
			ToPoint:         segment.ToPoint,
			Length:          segment.Length,
			Direction:       string(segment.Direction),
			SpeedLimit:      segment.SpeedLimit,
			RouteCostWeight: segment.RouteCostWeight,
			SegmentType:     segment.SegmentType,
		})
	}
	return mapResponse{
		Name:     item.Name,
		Points:   points,
		Segments: segments,
		Zones:    item.Zones,
	}
}

func toTickResultResponse(item application.StepRunResult) tickResultResponse {
	return tickResultResponse{
		Snapshot:      toSnapshotResponse(item.Snapshot),
		Metrics:       toMetricsResponses([]domain.MetricsSnapshot{item.Metrics})[0],
		Decisions:     toDecisionResponses(item.Decisions),
		StatusHistory: toStatusHistoryResponses(item.StatusHistory),
		SegmentLoads:  toSegmentLoadResponses(item.SegmentLoads),
	}
}

func toLiveEventResponse(item application.LiveEvent) liveEventResponse {
	return liveEventResponse{
		EventID:      item.EventID,
		RunID:        item.RunID,
		Kind:         item.Kind,
		Message:      item.Message,
		CurrentTick:  item.CurrentTick,
		RunStatus:    item.RunStatus,
		ActiveTasks:  item.ActiveTasks,
		WaitingTasks: item.WaitingTasks,
		OccurredAt:   item.OccurredAt.Format(time.RFC3339Nano),
	}
}

func toStatusHistoryResponses(items []domain.TaskStatusHistory) []statusHistoryResponse {
	out := make([]statusHistoryResponse, 0, len(items))
	for _, item := range items {
		out = append(out, statusHistoryResponse{
			RunID:  item.RunID,
			TaskID: item.TaskID,
			Tick:   item.Tick,
			Status: string(item.Status),
			Reason: item.Reason,
		})
	}
	return out
}

func toSnapshotResponse(item domain.SystemSnapshot) snapshotResponse {
	return snapshotResponse{
		Tick:             item.Tick,
		ActiveTasks:      item.ActiveTasks,
		WaitingTasks:     item.WaitingTasks,
		AvailableRobots:  item.AvailableRobots,
		FaultRobots:      item.FaultRobots,
		ChargingRobots:   item.ChargingRobots,
		PausedTasks:      item.PausedTasks,
		SegmentLoad:      copyFloatMap(item.SegmentLoad),
		ZoneLoad:         copyFloatMap(item.ZoneLoad),
		RobotUtilization: item.RobotUtilization,
	}
}

func copyFloatMap(items map[string]float64) map[string]float64 {
	if len(items) == 0 {
		return map[string]float64{}
	}
	out := make(map[string]float64, len(items))
	for key, value := range items {
		out[key] = value
	}
	return out
}

func copyIntMap(items map[string]int) map[string]int {
	if len(items) == 0 {
		return map[string]int{}
	}
	out := make(map[string]int, len(items))
	for key, value := range items {
		out[key] = value
	}
	return out
}

func formatOptionalTime(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := value.Format(time.RFC3339Nano)
	return &formatted
}
