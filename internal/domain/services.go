package domain

import (
	"errors"
	"time"
)

type Clock interface {
	Now() time.Time
}

type RealClock struct{}

func (RealClock) Now() time.Time { return time.Now().UTC() }

type RandomSource interface {
	Float64() float64
}

type MetricsCollector struct{}

type AdmissionController struct {
	Config AdmissionConfig
}

type Supervisor interface {
	Inspect(tick int, tasks []*TaskOrder, robots []*Robot) []Decision
}

type TaskFactory struct{}

func (TaskFactory) Build(request *Request) *TaskOrder {
	return &TaskOrder{
		TaskID:             "task-" + request.RequestID,
		RequestID:          request.RequestID,
		CurrentTargetPoint: request.SourcePoint,
		TaskStatus:         TaskStatusCreated,
		CreatedAt:          request.CreatedAt,
		StartedAt:          -1,
		FinishedAt:         -1,
		SourcePoint:        request.SourcePoint,
		TargetPoint:        request.TargetPoint,
	}
}

func (a AdmissionController) Allow(snapshot SystemSnapshot, candidate TaskCandidate, availableRobots int) error {
	if availableRobots == 0 {
		return errors.New("no_available_robot")
	}
	if a.Config.MaxActiveTasks > 0 && snapshot.ActiveTasks >= a.Config.MaxActiveTasks {
		return errors.New("active_limit_reached")
	}
	if a.Config.MaxRouteLoad > 0 && candidate.RouteLoad > a.Config.MaxRouteLoad {
		return errors.New("route_overloaded")
	}
	if a.Config.MaxZoneLoad > 0 && candidate.ZoneLoad > a.Config.MaxZoneLoad {
		return errors.New("zone_overloaded")
	}
	return nil
}
