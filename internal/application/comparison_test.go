package application

import (
	"os"
	"path/filepath"
	"testing"

	"amr-dispatcher-lab/internal/assets"
	"amr-dispatcher-lab/internal/domain"
	"amr-dispatcher-lab/internal/store"
)

type comparisonResult struct {
	run       domain.Run
	metrics   []domain.MetricsSnapshot
	decisions []domain.DecisionLog
	tasks     []domain.TaskOrder
}

func TestAlgorithmComparisonFixtures(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve root: %v", err)
	}

	cases := []struct {
		name         string
		mapPath      string
		scenarioPath string
		steps        int
	}{
		{
			name:         "baseline-on-small",
			mapPath:      "fixtures/maps/factory-small.json",
			scenarioPath: "fixtures/scenarios/baseline.json",
			steps:        40,
		},
		{
			name:         "overload-on-branching",
			mapPath:      "fixtures/maps/factory-branching.json",
			scenarioPath: "fixtures/scenarios/overload.json",
			steps:        40,
		},
		{
			name:         "rush-hour-on-large-mesh",
			mapPath:      "fixtures/maps/factory-large-mesh.json",
			scenarioPath: "fixtures/scenarios/rush-hour-mixed.json",
			steps:        60,
		},
		{
			name:         "full-campus-benchmark",
			mapPath:      "fixtures/maps/factory-campus.json",
			scenarioPath: "fixtures/scenarios/full-campus-benchmark.json",
			steps:        80,
		},
		{
			name:         "campus-route-pressure",
			mapPath:      "fixtures/maps/factory-campus.json",
			scenarioPath: "fixtures/scenarios/campus-route-pressure.json",
			steps:        120,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			results := map[domain.Algorithm]comparisonResult{}
			for _, algorithm := range []domain.Algorithm{
				domain.AlgorithmFIFO,
				domain.AlgorithmPriority,
				domain.AlgorithmAdaptive,
			} {
				results[algorithm] = runFixtureComparison(t, root, algorithm, tc.mapPath, tc.scenarioPath, tc.steps)
			}

			for algorithm, result := range results {
				last := result.metrics[len(result.metrics)-1]
				t.Logf("%s -> status=%s tick=%d completed=%d cancelled=%d failed=%d throughput=%.4f avgWait=%.4f avgExec=%.4f util=%.4f decisions=%d",
					algorithm,
					result.run.Status,
					result.run.CurrentTick,
					last.CompletedTasks,
					last.CancelledTasks,
					last.FailedTasks,
					last.Throughput,
					last.AvgWaitTime,
					last.AvgExecutionTime,
					last.RobotUtilization,
					len(result.decisions),
				)
				t.Logf("%s dispatch=%v tasks=%v",
					algorithm,
					dispatchSequence(result.decisions),
					taskStates(result.tasks),
				)
				if tc.name == "full-campus-benchmark" {
					t.Logf("%s reasons=%v", algorithm, reasonCounts(result.decisions))
				}
			}

			if tc.name == "rush-hour-on-large-mesh" || tc.name == "full-campus-benchmark" || tc.name == "campus-route-pressure" {
				assertBehaviorDiverges(t, results)
			}
		})
	}
}

func TestFullCampusBenchmarkEventuallyDrainsQueue(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve root: %v", err)
	}

	for _, algorithm := range []domain.Algorithm{
		domain.AlgorithmFIFO,
		domain.AlgorithmPriority,
		domain.AlgorithmAdaptive,
	} {
		t.Run(string(algorithm), func(t *testing.T) {
			result := runFixtureComparison(
				t,
				root,
				algorithm,
				"fixtures/maps/factory-campus.json",
				"fixtures/scenarios/full-campus-benchmark.json",
				200,
			)

			if result.run.Status != domain.RunStatusCompleted {
				t.Fatalf("expected benchmark run to complete, got %s at tick %d", result.run.Status, result.run.CurrentTick)
			}

			queued := 0
			unfinished := 0
			for _, task := range result.tasks {
				if task.TaskStatus == domain.TaskStatusQueued {
					queued++
				}
				if task.TaskStatus != domain.TaskStatusCompleted &&
					task.TaskStatus != domain.TaskStatusCancelled &&
					task.TaskStatus != domain.TaskStatusFailed {
					unfinished++
				}
			}
			if queued > 0 || unfinished > 0 {
				t.Fatalf("expected queue to drain completely, got queued=%d unfinished=%d tasks=%v", queued, unfinished, taskStates(result.tasks))
			}
		})
	}
}

func TestAdaptiveTunedBenchmarksOutperformPriorityOnTargetMetrics(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve root: %v", err)
	}

	t.Run("rush-hour-mixed", func(t *testing.T) {
		priority := runFixtureComparison(
			t,
			root,
			domain.AlgorithmPriority,
			"fixtures/maps/factory-large-mesh.json",
			"fixtures/scenarios/rush-hour-mixed.json",
			60,
		)
		adaptive := runFixtureComparison(
			t,
			root,
			domain.AlgorithmAdaptive,
			"fixtures/maps/factory-large-mesh.json",
			"fixtures/scenarios/rush-hour-mixed.json",
			60,
		)

		priorityLast := priority.metrics[len(priority.metrics)-1]
		adaptiveLast := adaptive.metrics[len(adaptive.metrics)-1]

		if adaptiveLast.Throughput < priorityLast.Throughput {
			t.Fatalf("expected adaptive throughput >= priority on rush-hour benchmark, got adaptive=%.4f priority=%.4f", adaptiveLast.Throughput, priorityLast.Throughput)
		}
		if adaptiveLast.AvgWaitTime >= priorityLast.AvgWaitTime {
			t.Fatalf("expected adaptive observed wait < priority on rush-hour benchmark, got adaptive=%.4f priority=%.4f", adaptiveLast.AvgWaitTime, priorityLast.AvgWaitTime)
		}
	})

	t.Run("full-campus-benchmark", func(t *testing.T) {
		priority := runFixtureComparison(
			t,
			root,
			domain.AlgorithmPriority,
			"fixtures/maps/factory-campus.json",
			"fixtures/scenarios/full-campus-benchmark.json",
			120,
		)
		adaptive := runFixtureComparison(
			t,
			root,
			domain.AlgorithmAdaptive,
			"fixtures/maps/factory-campus.json",
			"fixtures/scenarios/full-campus-benchmark.json",
			120,
		)

		priorityLast := priority.metrics[len(priority.metrics)-1]
		adaptiveLast := adaptive.metrics[len(adaptive.metrics)-1]

		if adaptiveLast.Throughput < priorityLast.Throughput {
			t.Fatalf("expected adaptive throughput >= priority on full-campus benchmark, got adaptive=%.4f priority=%.4f", adaptiveLast.Throughput, priorityLast.Throughput)
		}
		if adaptive.run.CurrentTick > priority.run.CurrentTick {
			t.Fatalf("expected adaptive to finish no slower than priority on full-campus benchmark, got adaptive tick=%d priority tick=%d", adaptive.run.CurrentTick, priority.run.CurrentTick)
		}
	})
}

func TestCampusRoutePressureIsTrafficConstrained(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve root: %v", err)
	}

	for _, algorithm := range []domain.Algorithm{
		domain.AlgorithmFIFO,
		domain.AlgorithmPriority,
		domain.AlgorithmAdaptive,
	} {
		t.Run(string(algorithm), func(t *testing.T) {
			result := runFixtureComparison(
				t,
				root,
				algorithm,
				"fixtures/maps/factory-campus.json",
				"fixtures/scenarios/campus-route-pressure.json",
				160,
			)

			reasons := reasonCounts(result.decisions)
			trafficRejects := reasons["route_overloaded"] + reasons["zone_overloaded"]
			resourceRejects := reasons["no_available_robot"]

			if trafficRejects <= resourceRejects {
				t.Fatalf("expected traffic pressure to dominate resource scarcity for %s, got traffic=%d resource=%d reasons=%v", algorithm, trafficRejects, resourceRejects, reasons)
			}
		})
	}
}

func runFixtureComparison(t *testing.T, root string, algorithm domain.Algorithm, mapPath, scenarioPath string, steps int) comparisonResult {
	t.Helper()

	dbPath := filepath.Join(root, "data", "comparison-"+string(algorithm)+".db")
	_ = os.Remove(dbPath)
	runStore, err := store.NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("create store: %v", err)
	}
	defer func() {
		_ = runStore.Close()
		_ = os.Remove(dbPath)
	}()

	service := NewService(Options{
		DefaultMapPath:      mapPath,
		DefaultScenarioPath: scenarioPath,
		TickDurationMillis:  1,
	}, assets.NewFilesystemAssets(root), runStore, nil, nil)

	run, err := service.CreateRun(CreateRunInput{
		Algorithm:    algorithm,
		MapPath:      mapPath,
		ScenarioPath: scenarioPath,
		Seed:         123,
	})
	if err != nil {
		t.Fatalf("create run: %v", err)
	}

	for step := 0; step < steps; step++ {
		current, err := service.GetRun(run.ID)
		if err != nil {
			t.Fatalf("get run: %v", err)
		}
		if current.Status == domain.RunStatusCompleted || current.Status == domain.RunStatusFailed {
			break
		}
		if _, _, err := service.StepRun(run.ID); err != nil {
			t.Fatalf("step run: %v", err)
		}
	}

	finalRun, err := service.GetRun(run.ID)
	if err != nil {
		t.Fatalf("get final run: %v", err)
	}
	metrics, err := service.GetMetrics(run.ID)
	if err != nil {
		t.Fatalf("get metrics: %v", err)
	}
	decisions, err := service.GetDecisionLogs(run.ID)
	if err != nil {
		t.Fatalf("get decisions: %v", err)
	}
	tasks, err := service.GetTasks(run.ID)
	if err != nil {
		t.Fatalf("get tasks: %v", err)
	}
	if len(metrics) == 0 {
		t.Fatalf("expected metrics snapshots")
	}
	return comparisonResult{
		run:       finalRun,
		metrics:   metrics,
		decisions: decisions,
		tasks:     tasks,
	}
}

func dispatchSequence(items []domain.DecisionLog) []string {
	seq := make([]string, 0)
	for _, item := range items {
		if item.Action == domain.DecisionActionDispatch {
			seq = append(seq, item.TaskID)
		}
	}
	return seq
}

func taskStates(items []domain.TaskOrder) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, item.TaskID+"="+string(item.TaskStatus))
	}
	return out
}

func assertBehaviorDiverges(t *testing.T, results map[domain.Algorithm]comparisonResult) {
	t.Helper()

	fifo := results[domain.AlgorithmFIFO].metrics[len(results[domain.AlgorithmFIFO].metrics)-1]
	priority := results[domain.AlgorithmPriority].metrics[len(results[domain.AlgorithmPriority].metrics)-1]
	adaptive := results[domain.AlgorithmAdaptive].metrics[len(results[domain.AlgorithmAdaptive].metrics)-1]

	allSame := fifo.AvgWaitTime == priority.AvgWaitTime &&
		priority.AvgWaitTime == adaptive.AvgWaitTime &&
		fifo.Throughput == priority.Throughput &&
		priority.Throughput == adaptive.Throughput &&
		fifo.DeadlineSuccessRate == priority.DeadlineSuccessRate &&
		priority.DeadlineSuccessRate == adaptive.DeadlineSuccessRate &&
		fifo.AvgExecutionTime == priority.AvgExecutionTime &&
		priority.AvgExecutionTime == adaptive.AvgExecutionTime &&
		fifo.RobotUtilization == priority.RobotUtilization &&
		priority.RobotUtilization == adaptive.RobotUtilization

	dispatchSame := sameStrings(dispatchSequence(results[domain.AlgorithmFIFO].decisions), dispatchSequence(results[domain.AlgorithmPriority].decisions)) &&
		sameStrings(dispatchSequence(results[domain.AlgorithmPriority].decisions), dispatchSequence(results[domain.AlgorithmAdaptive].decisions))

	if allSame && dispatchSame {
		t.Fatalf("expected large rush-hour benchmark to differentiate algorithms, got identical final metrics: fifo=%+v priority=%+v adaptive=%+v", fifo, priority, adaptive)
	}
}

func sameStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func reasonCounts(items []domain.DecisionLog) map[string]int {
	out := map[string]int{}
	for _, item := range items {
		out[item.ReasonCode]++
	}
	return out
}
