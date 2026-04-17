package store

import (
	"amr-dispatcher-lab/internal/domain"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

func (s *SQLiteStore) GetRun(id string) (domain.Run, error) {
	row := s.db.QueryRow(`SELECT id, status, mode, current_tick, algorithm, map_name, map_path, scenario_name, scenario_path, seed, config_hash, created_at, started_at, finished_at, last_error FROM runs WHERE id = ?`, id)
	return scanRun(row)
}

func (s *SQLiteStore) ListRuns() ([]domain.Run, error) {
	rows, err := s.db.Query(`SELECT id, status, mode, current_tick, algorithm, map_name, map_path, scenario_name, scenario_path, seed, config_hash, created_at, started_at, finished_at, last_error FROM runs ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.Run
	for rows.Next() {
		run, err := scanRun(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, run)
	}
	return items, rows.Err()
}

func (s *SQLiteStore) GetRequests(runID string) ([]domain.Request, error) {
	rows, err := s.db.Query(`SELECT request_id, source_point, target_point, business_type, priority, created_at, deadline, status FROM requests WHERE run_id = ? ORDER BY created_at, request_id`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.Request
	for rows.Next() {
		var item domain.Request
		if err := rows.Scan(&item.RequestID, &item.SourcePoint, &item.TargetPoint, &item.BusinessType, &item.Priority, &item.CreatedAt, &item.Deadline, &item.Status); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func scanRun(scanner interface {
	Scan(dest ...interface{}) error
}) (domain.Run, error) {
	var run domain.Run
	var createdAt string
	var startedAt sql.NullString
	var finishedAt sql.NullString
	if err := scanner.Scan(&run.ID, &run.Status, &run.Mode, &run.CurrentTick, &run.Algorithm, &run.MapName, &run.MapPath, &run.ScenarioName, &run.ScenarioPath, &run.Seed, &run.ConfigHash, &createdAt, &startedAt, &finishedAt, &run.LastError); err != nil {
		return domain.Run{}, err
	}
	parsedCreated, err := time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return domain.Run{}, err
	}
	run.CreatedAt = parsedCreated
	if startedAt.Valid {
		parsed, err := time.Parse(time.RFC3339Nano, startedAt.String)
		if err != nil {
			return domain.Run{}, err
		}
		run.StartedAt = &parsed
	}
	if finishedAt.Valid {
		parsed, err := time.Parse(time.RFC3339Nano, finishedAt.String)
		if err != nil {
			return domain.Run{}, err
		}
		run.FinishedAt = &parsed
	}
	return run, nil
}

func (s *SQLiteStore) GetTasks(runID string) ([]domain.TaskOrder, error) {
	rows, err := s.db.Query(`SELECT task_id, request_id, robot_id, planned_path_json, current_target_point, task_status, created_at, started_at, finished_at, source_point, target_point, path_index, estimated_cost, estimated_duration, last_progress_tick, defer_until_tick, assigned_robot_point FROM tasks WHERE run_id = ? ORDER BY created_at`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.TaskOrder
	for rows.Next() {
		var item domain.TaskOrder
		var pathJSON string
		if err := rows.Scan(&item.TaskID, &item.RequestID, &item.RobotID, &pathJSON, &item.CurrentTargetPoint, &item.TaskStatus, &item.CreatedAt, &item.StartedAt, &item.FinishedAt, &item.SourcePoint, &item.TargetPoint, &item.PathIndex, &item.EstimatedCost, &item.EstimatedDuration, &item.LastProgressTick, &item.DeferUntilTick, &item.AssignedRobotPoint); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(pathJSON), &item.PlannedPath)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *SQLiteStore) GetRobots(runID string) ([]domain.Robot, error) {
	rows, err := s.db.Query(`SELECT robot_id, robot_model, state, current_point, battery_level, current_task_id, fault_flag, busy_ticks FROM robots WHERE run_id = ? ORDER BY robot_id`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.Robot
	for rows.Next() {
		var item domain.Robot
		var fault int
		if err := rows.Scan(&item.RobotID, &item.RobotModel, &item.State, &item.CurrentPoint, &item.BatteryLevel, &item.CurrentTaskID, &fault, &item.BusyTicks); err != nil {
			return nil, err
		}
		item.FaultFlag = fault == 1
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *SQLiteStore) GetMetrics(runID string) ([]domain.MetricsSnapshot, error) {
	rows, err := s.db.Query(`SELECT run_id, tick, avg_wait_time, avg_execution_time, throughput, cancel_rate, deadline_success_rate, segment_load_variance, robot_utilization, completed_tasks, cancelled_tasks, failed_tasks FROM metrics_snapshots WHERE run_id = ? ORDER BY tick`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.MetricsSnapshot
	for rows.Next() {
		var item domain.MetricsSnapshot
		if err := rows.Scan(&item.RunID, &item.Tick, &item.AvgWaitTime, &item.AvgExecutionTime, &item.Throughput, &item.CancelRate, &item.DeadlineSuccessRate, &item.SegmentLoadVariance, &item.RobotUtilization, &item.CompletedTasks, &item.CancelledTasks, &item.FailedTasks); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *SQLiteStore) GetSegmentLoads(runID string) ([]domain.SegmentLoadSnapshot, error) {
	rows, err := s.db.Query(`SELECT run_id, tick, segment_id, load FROM segment_load WHERE run_id = ? ORDER BY tick, segment_id`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.SegmentLoadSnapshot
	for rows.Next() {
		var item domain.SegmentLoadSnapshot
		if err := rows.Scan(&item.RunID, &item.Tick, &item.SegmentID, &item.Load); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *SQLiteStore) GetDecisionLogs(runID string) ([]domain.DecisionLog, error) {
	rows, err := s.db.Query(`SELECT run_id, tick, request_id, task_id, robot_id, algorithm, action, reason_code, score, defer_until_tick FROM decision_log WHERE run_id = ? ORDER BY tick, id`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.DecisionLog
	for rows.Next() {
		var item domain.DecisionLog
		if err := rows.Scan(&item.RunID, &item.Tick, &item.RequestID, &item.TaskID, &item.RobotID, &item.Algorithm, &item.Action, &item.ReasonCode, &item.Score, &item.DeferUntilTick); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *SQLiteStore) String() string {
	return fmt.Sprintf("sqlite_store(%p)", s.db)
}
