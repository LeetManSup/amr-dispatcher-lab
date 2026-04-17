package store

func (s *SQLiteStore) migrate() error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS runs (
			id TEXT PRIMARY KEY,
			status TEXT NOT NULL,
			mode TEXT NOT NULL,
			current_tick INTEGER NOT NULL,
			algorithm TEXT NOT NULL,
			map_name TEXT NOT NULL,
			map_path TEXT NOT NULL,
			scenario_name TEXT NOT NULL,
			scenario_path TEXT NOT NULL,
			seed INTEGER NOT NULL,
			config_hash TEXT NOT NULL,
			created_at TEXT NOT NULL,
			started_at TEXT,
			finished_at TEXT,
			last_error TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS requests (
			run_id TEXT NOT NULL,
			request_id TEXT NOT NULL,
			source_point TEXT NOT NULL,
			target_point TEXT NOT NULL,
			business_type TEXT NOT NULL,
			priority INTEGER NOT NULL,
			created_at INTEGER NOT NULL,
			deadline INTEGER NOT NULL,
			status TEXT NOT NULL,
			PRIMARY KEY (run_id, request_id)
		)`,
		`CREATE TABLE IF NOT EXISTS tasks (
			run_id TEXT NOT NULL,
			task_id TEXT NOT NULL,
			request_id TEXT NOT NULL,
			robot_id TEXT NOT NULL,
			planned_path_json TEXT NOT NULL,
			current_target_point TEXT NOT NULL,
			task_status TEXT NOT NULL,
			created_at INTEGER NOT NULL,
			started_at INTEGER NOT NULL,
			finished_at INTEGER NOT NULL,
			source_point TEXT NOT NULL,
			target_point TEXT NOT NULL,
			path_index INTEGER NOT NULL,
			estimated_cost REAL NOT NULL,
			estimated_duration INTEGER NOT NULL,
			last_progress_tick INTEGER NOT NULL,
			defer_until_tick INTEGER NOT NULL,
			assigned_robot_point TEXT NOT NULL,
			PRIMARY KEY (run_id, task_id)
		)`,
		`CREATE TABLE IF NOT EXISTS robots (
			run_id TEXT NOT NULL,
			robot_id TEXT NOT NULL,
			robot_model TEXT NOT NULL,
			state TEXT NOT NULL,
			current_point TEXT NOT NULL,
			battery_level REAL NOT NULL,
			current_task_id TEXT NOT NULL,
			fault_flag INTEGER NOT NULL,
			busy_ticks INTEGER NOT NULL,
			PRIMARY KEY (run_id, robot_id)
		)`,
		`CREATE TABLE IF NOT EXISTS task_status_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			run_id TEXT NOT NULL,
			task_id TEXT NOT NULL,
			tick INTEGER NOT NULL,
			status TEXT NOT NULL,
			reason TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS decision_log (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			run_id TEXT NOT NULL,
			tick INTEGER NOT NULL,
			request_id TEXT NOT NULL,
			task_id TEXT NOT NULL,
			robot_id TEXT NOT NULL,
			algorithm TEXT NOT NULL,
			action TEXT NOT NULL,
			reason_code TEXT NOT NULL,
			score REAL NOT NULL,
			defer_until_tick INTEGER NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS metrics_snapshots (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			run_id TEXT NOT NULL,
			tick INTEGER NOT NULL,
			avg_wait_time REAL NOT NULL,
			avg_execution_time REAL NOT NULL,
			throughput REAL NOT NULL,
			cancel_rate REAL NOT NULL,
			deadline_success_rate REAL NOT NULL,
			segment_load_variance REAL NOT NULL,
			robot_utilization REAL NOT NULL,
			completed_tasks INTEGER NOT NULL,
			cancelled_tasks INTEGER NOT NULL,
			failed_tasks INTEGER NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS segment_load (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			run_id TEXT NOT NULL,
			tick INTEGER NOT NULL,
			segment_id TEXT NOT NULL,
			load REAL NOT NULL
		)`,
	}
	for _, statement := range statements {
		if _, err := s.db.Exec(statement); err != nil {
			return err
		}
	}
	return nil
}
