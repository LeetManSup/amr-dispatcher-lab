package application

import (
	"fmt"
	"strings"

	"amr-dispatcher-lab/internal/domain"
	"amr-dispatcher-lab/internal/planner"
)

type ValidationError struct {
	Issues []string
}

func (e ValidationError) Error() string {
	if len(e.Issues) == 0 {
		return "invalid run configuration"
	}
	return "invalid run configuration: " + strings.Join(e.Issues, "; ")
}

func validateRunConfiguration(mapData domain.MapData, scenario domain.Scenario) error {
	issues := make([]string, 0)
	pointIndex := make(map[string]domain.MapPoint, len(mapData.Points))
	for _, point := range mapData.Points {
		pointIndex[point.PointID] = point
	}

	if len(scenario.Robots) == 0 {
		issues = append(issues, "scenario has no robots")
	}

	routePlanner, err := planner.NewGraphPlanner(mapData)
	if err != nil {
		return ValidationError{Issues: []string{fmt.Sprintf("map is invalid: %v", err)}}
	}

	reachableRobotToSource := func(source string) bool {
		for _, robot := range scenario.Robots {
			if _, err := routePlanner.BuildPath(robot.StartPoint, source); err == nil {
				return true
			}
		}
		return false
	}

	for _, robot := range scenario.Robots {
		if _, ok := pointIndex[robot.StartPoint]; !ok {
			issues = append(issues, fmt.Sprintf("robot %s start point %s is not present on map %s", robot.RobotID, robot.StartPoint, pickFirstNonEmpty(mapData.Name, "selected map")))
		}
	}

	for _, request := range scenario.Requests {
		if _, ok := pointIndex[request.SourcePoint]; !ok {
			issues = append(issues, fmt.Sprintf("request %s source point %s is not present on map %s", request.RequestID, request.SourcePoint, pickFirstNonEmpty(mapData.Name, "selected map")))
			continue
		}
		if _, ok := pointIndex[request.TargetPoint]; !ok {
			issues = append(issues, fmt.Sprintf("request %s target point %s is not present on map %s", request.RequestID, request.TargetPoint, pickFirstNonEmpty(mapData.Name, "selected map")))
			continue
		}
		if _, err := routePlanner.BuildPath(request.SourcePoint, request.TargetPoint); err != nil {
			issues = append(issues, fmt.Sprintf("request %s has unreachable source -> target path (%s -> %s): %v", request.RequestID, request.SourcePoint, request.TargetPoint, err))
		}
		if !reachableRobotToSource(request.SourcePoint) {
			issues = append(issues, fmt.Sprintf("request %s source point %s is unreachable from every robot start point", request.RequestID, request.SourcePoint))
		}
	}

	if len(issues) > 0 {
		return ValidationError{Issues: issues}
	}
	return nil
}
