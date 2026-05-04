package application

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"
)

func (s *Service) ExportRun(runID, format string) ([]byte, string, error) {
	run, err := s.GetRun(runID)
	if err != nil {
		return nil, "", err
	}
	tasks, _ := s.GetTasks(runID)
	robots, _ := s.GetRobots(runID)
	metrics, _ := s.GetMetrics(runID)
	decisions, _ := s.GetDecisionLogs(runID)

	switch strings.ToLower(format) {
	case "csv":
		buffer := &bytes.Buffer{}
		writer := csv.NewWriter(buffer)
		_ = writer.Write([]string{"run_id", "tick", "avg_wait_time", "avg_execution_time", "throughput", "cancel_rate", "deadline_success_rate", "segment_load_variance", "robot_utilization", "completed_tasks", "cancelled_tasks", "failed_tasks"})
		for _, item := range metrics {
			_ = writer.Write([]string{
				item.RunID,
				fmt.Sprintf("%d", item.Tick),
				fmt.Sprintf("%.4f", item.AvgWaitTime),
				fmt.Sprintf("%.4f", item.AvgExecutionTime),
				fmt.Sprintf("%.4f", item.Throughput),
				fmt.Sprintf("%.4f", item.CancelRate),
				fmt.Sprintf("%.4f", item.DeadlineSuccessRate),
				fmt.Sprintf("%.4f", item.SegmentLoadVariance),
				fmt.Sprintf("%.4f", item.RobotUtilization),
				fmt.Sprintf("%d", item.CompletedTasks),
				fmt.Sprintf("%d", item.CancelledTasks),
				fmt.Sprintf("%d", item.FailedTasks),
			})
		}
		writer.Flush()
		return buffer.Bytes(), "text/csv", writer.Error()
	default:
		payload := map[string]interface{}{
			"run":       run,
			"tasks":     tasks,
			"robots":    robots,
			"metrics":   metrics,
			"decisions": decisions,
		}
		data, err := json.MarshalIndent(payload, "", "  ")
		return data, "application/json", err
	}
}
