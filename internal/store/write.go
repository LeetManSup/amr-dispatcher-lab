package store

import (
	"amr-dispatcher-lab/internal/domain"
	"context"
	"database/sql"
	"encoding/json"
	"time"
)

func (s *SQLiteStore) SaveRun(run domain.Run) error {
	_, err := s.db.Exec(
		`INSERT INTO runs (id, status, mode, current_tick, algorithm, map_name, map_path, scenario_name, scenario_path, seed, config_hash, created_at, started_at, finished_at, last_error)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		run.ID, run.Status, run.Mode, run.CurrentTick, run.Algorithm, run.MapName, run.MapPath, run.ScenarioName, run.ScenarioPath, run.Seed, run.ConfigHash, run.CreatedAt.Format(time.RFC3339Nano), nullableTime(run.StartedAt), nullableTime(run.FinishedAt), run.LastError,
	)
	return err
}

func (s *SQLiteStore) UpdateRun(run domain.Run) error {
	_, err := s.db.Exec(
		`UPDATE runs SET status = ?, mode = ?, current_tick = ?, algorithm = ?, map_name = ?, map_path = ?, scenario_name = ?, scenario_path = ?, seed = ?, config_hash = ?, created_at = ?, started_at = ?, finished_at = ?, last_error = ? WHERE id = ?`,
		run.Status, run.Mode, run.CurrentTick, run.Algorithm, run.MapName, run.MapPath, run.ScenarioName, run.ScenarioPath, run.Seed, run.ConfigHash, run.CreatedAt.Format(time.RFC3339Nano), nullableTime(run.StartedAt), nullableTime(run.FinishedAt), run.LastError, run.ID,
	)
	return err
}

func (s *SQLiteStore) UpsertRequests(runID string, requests []domain.Request) error {
	tx, err := s.db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(`INSERT INTO requests (run_id, request_id, source_point, target_point, business_type, priority, created_at, deadline, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(run_id, request_id) DO UPDATE SET source_point = excluded.source_point, target_point = excluded.target_point, business_type = excluded.business_type, priority = excluded.priority, created_at = excluded.created_at, deadline = excluded.deadline, status = excluded.status`)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	defer stmt.Close()
	for _, item := range requests {
		if _, err := stmt.Exec(runID, item.RequestID, item.SourcePoint, item.TargetPoint, item.BusinessType, item.Priority, item.CreatedAt, item.Deadline, item.Status); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (s *SQLiteStore) UpsertTasks(runID string, tasks []domain.TaskOrder) error {
	tx, err := s.db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(`INSERT INTO tasks (run_id, task_id, request_id, robot_id, planned_path_json, current_target_point, task_status, created_at, started_at, finished_at, source_point, target_point, path_index, estimated_cost, estimated_duration, last_progress_tick, defer_until_tick, assigned_robot_point)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(run_id, task_id) DO UPDATE SET request_id = excluded.request_id, robot_id = excluded.robot_id, planned_path_json = excluded.planned_path_json, current_target_point = excluded.current_target_point, task_status = excluded.task_status, created_at = excluded.created_at, started_at = excluded.started_at, finished_at = excluded.finished_at, source_point = excluded.source_point, target_point = excluded.target_point, path_index = excluded.path_index, estimated_cost = excluded.estimated_cost, estimated_duration = excluded.estimated_duration, last_progress_tick = excluded.last_progress_tick, defer_until_tick = excluded.defer_until_tick, assigned_robot_point = excluded.assigned_robot_point`)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	defer stmt.Close()
	for _, item := range tasks {
		pathJSON, _ := json.Marshal(item.PlannedPath)
		if _, err := stmt.Exec(runID, item.TaskID, item.RequestID, item.RobotID, string(pathJSON), item.CurrentTargetPoint, item.TaskStatus, item.CreatedAt, item.StartedAt, item.FinishedAt, item.SourcePoint, item.TargetPoint, item.PathIndex, item.EstimatedCost, item.EstimatedDuration, item.LastProgressTick, item.DeferUntilTick, item.AssignedRobotPoint); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (s *SQLiteStore) UpsertRobots(runID string, robots []domain.Robot) error {
	tx, err := s.db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(`INSERT INTO robots (run_id, robot_id, robot_model, state, current_point, battery_level, current_task_id, fault_flag, busy_ticks)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(run_id, robot_id) DO UPDATE SET robot_model = excluded.robot_model, state = excluded.state, current_point = excluded.current_point, battery_level = excluded.battery_level, current_task_id = excluded.current_task_id, fault_flag = excluded.fault_flag, busy_ticks = excluded.busy_ticks`)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	defer stmt.Close()
	for _, item := range robots {
		fault := 0
		if item.FaultFlag {
			fault = 1
		}
		if _, err := stmt.Exec(runID, item.RobotID, item.RobotModel, item.State, item.CurrentPoint, item.BatteryLevel, item.CurrentTaskID, fault, item.BusyTicks); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (s *SQLiteStore) AppendTaskHistory(items []domain.TaskStatusHistory) error {
	return execMany(s.db, `INSERT INTO task_status_history (run_id, task_id, tick, status, reason) VALUES (?, ?, ?, ?, ?)`, func(stmt *sql.Stmt) error {
		for _, item := range items {
			if _, err := stmt.Exec(item.RunID, item.TaskID, item.Tick, item.Status, item.Reason); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *SQLiteStore) AppendDecisionLogs(items []domain.DecisionLog) error {
	return execMany(s.db, `INSERT INTO decision_log (run_id, tick, request_id, task_id, robot_id, algorithm, action, reason_code, score, defer_until_tick) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, func(stmt *sql.Stmt) error {
		for _, item := range items {
			if _, err := stmt.Exec(item.RunID, item.Tick, item.RequestID, item.TaskID, item.RobotID, item.Algorithm, item.Action, item.ReasonCode, item.Score, item.DeferUntilTick); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *SQLiteStore) AppendMetrics(items []domain.MetricsSnapshot) error {
	return execMany(s.db, `INSERT INTO metrics_snapshots (run_id, tick, avg_wait_time, avg_execution_time, throughput, cancel_rate, deadline_success_rate, segment_load_variance, robot_utilization, completed_tasks, cancelled_tasks, failed_tasks) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, func(stmt *sql.Stmt) error {
		for _, item := range items {
			if _, err := stmt.Exec(item.RunID, item.Tick, item.AvgWaitTime, item.AvgExecutionTime, item.Throughput, item.CancelRate, item.DeadlineSuccessRate, item.SegmentLoadVariance, item.RobotUtilization, item.CompletedTasks, item.CancelledTasks, item.FailedTasks); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *SQLiteStore) AppendSegmentLoads(items []domain.SegmentLoadSnapshot) error {
	return execMany(s.db, `INSERT INTO segment_load (run_id, tick, segment_id, load) VALUES (?, ?, ?, ?)`, func(stmt *sql.Stmt) error {
		for _, item := range items {
			if _, err := stmt.Exec(item.RunID, item.Tick, item.SegmentID, item.Load); err != nil {
				return err
			}
		}
		return nil
	})
}
