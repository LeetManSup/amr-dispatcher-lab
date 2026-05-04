package application

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"amr-dispatcher-lab/internal/assets"
	"amr-dispatcher-lab/internal/domain"
	"amr-dispatcher-lab/internal/store"
)

func TestServiceRunCompletesWithoutBusyRobotLeak(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve root: %v", err)
	}
	dbPath := filepath.Join(root, "data", "test-service.db")
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
		DefaultMapPath:      "fixtures/maps/factory-small.json",
		DefaultScenarioPath: "fixtures/scenarios/baseline.json",
		TickDurationMillis:  1,
	}, assets.NewFilesystemAssets(root), runStore, nil, nil)

	run, err := service.CreateRun(CreateRunInput{
		Algorithm:    domain.AlgorithmFIFO,
		MapPath:      "fixtures/maps/factory-small.json",
		ScenarioPath: "fixtures/scenarios/baseline.json",
		Seed:         123,
	})
	if err != nil {
		t.Fatalf("create run: %v", err)
	}

	for step := 0; step < 30; step++ {
		current, err := service.GetRun(run.ID)
		if err != nil {
			t.Fatalf("get run: %v", err)
		}
		if current.Status == domain.RunStatusCompleted {
			break
		}
		if _, _, err := service.StepRun(run.ID); err != nil {
			t.Fatalf("step run: %v", err)
		}
	}

	current, err := service.GetRun(run.ID)
	if err != nil {
		t.Fatalf("get final run: %v", err)
	}
	if current.Status != domain.RunStatusCompleted {
		t.Fatalf("expected run to complete, got %s", current.Status)
	}

	robots, err := service.GetRobots(run.ID)
	if err != nil {
		t.Fatalf("get robots: %v", err)
	}
	for _, robot := range robots {
		if robot.State == domain.RobotStateBusy && robot.CurrentTaskID == "" {
			t.Fatalf("robot %s leaked busy state without current task", robot.RobotID)
		}
	}
}

func TestCreateRunRejectsIncompatibleMapAndScenario(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve root: %v", err)
	}
	dbPath := filepath.Join(root, "data", "test-service-validation.db")
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
		DefaultMapPath:      "fixtures/maps/factory-branching.json",
		DefaultScenarioPath: "fixtures/scenarios/baseline.json",
		TickDurationMillis:  1,
	}, assets.NewFilesystemAssets(root), runStore, nil, nil)

	_, err = service.CreateRun(CreateRunInput{
		Algorithm:    domain.AlgorithmFIFO,
		MapPath:      "fixtures/maps/factory-branching.json",
		ScenarioPath: "fixtures/scenarios/baseline.json",
		Seed:         123,
	})
	if err == nil {
		t.Fatalf("expected validation error for incompatible map and scenario")
	}
	if !strings.Contains(err.Error(), "is not present on map") {
		t.Fatalf("expected point compatibility error, got %v", err)
	}
}
