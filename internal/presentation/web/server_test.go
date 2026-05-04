package web

import (
	"bufio"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"amr-dispatcher-lab/internal/application"
	"amr-dispatcher-lab/internal/assets"
	"amr-dispatcher-lab/internal/config"
	"amr-dispatcher-lab/internal/store"
)

func TestRunLifecycleEndpoints(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	createReq := httptest.NewRequest(http.MethodPost, "/api/runs", bytes.NewBufferString(`{"algorithm":"fifo","mapPath":"fixtures/maps/factory-small.json","scenarioPath":"fixtures/scenarios/baseline.json","seed":123}`))
	createRes := httptest.NewRecorder()
	server.Handler().ServeHTTP(createRes, createReq)
	if createRes.Code != http.StatusCreated {
		t.Fatalf("expected 201 from create run, got %d: %s", createRes.Code, createRes.Body.String())
	}

	var run struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createRes.Body).Decode(&run); err != nil {
		t.Fatalf("decode create run response: %v", err)
	}

	stepReq := httptest.NewRequest(http.MethodPost, "/api/runs/"+run.ID+"/step", nil)
	stepRes := httptest.NewRecorder()
	server.Handler().ServeHTTP(stepRes, stepReq)
	if stepRes.Code != http.StatusOK {
		t.Fatalf("expected 200 from step run, got %d: %s", stepRes.Code, stepRes.Body.String())
	}

	taskReq := httptest.NewRequest(http.MethodPost, "/api/runs/"+run.ID+"/ics/addTask", bytes.NewBufferString(`{"requestId":"manual-100","sourcePoint":"P1","targetPoint":"P4","businessType":"manual","priority":8,"createdAt":0,"deadline":10}`))
	taskRes := httptest.NewRecorder()
	server.Handler().ServeHTTP(taskRes, taskReq)
	if taskRes.Code != http.StatusOK {
		t.Fatalf("expected 200 from addTask, got %d: %s", taskRes.Code, taskRes.Body.String())
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/runs/"+run.ID+"/tasks", nil)
	getRes := httptest.NewRecorder()
	server.Handler().ServeHTTP(getRes, getReq)
	if getRes.Code != http.StatusOK {
		t.Fatalf("expected 200 from get tasks, got %d", getRes.Code)
	}
}

func TestUIRoutesAndCatalog(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	indexReq := httptest.NewRequest(http.MethodGet, "/", nil)
	indexRes := httptest.NewRecorder()
	server.Handler().ServeHTTP(indexRes, indexReq)
	if indexRes.Code != http.StatusOK {
		t.Fatalf("expected 200 from index, got %d", indexRes.Code)
	}
	if !strings.Contains(indexRes.Body.String(), "/ui/static/js/main.js") {
		t.Fatalf("index page did not reference modular frontend entrypoint")
	}

	assetReq := httptest.NewRequest(http.MethodGet, "/ui/static/css/app.css", nil)
	assetRes := httptest.NewRecorder()
	server.Handler().ServeHTTP(assetRes, assetReq)
	if assetRes.Code != http.StatusOK {
		t.Fatalf("expected 200 from static asset, got %d", assetRes.Code)
	}

	catalogReq := httptest.NewRequest(http.MethodGet, "/api/catalog", nil)
	catalogRes := httptest.NewRecorder()
	server.Handler().ServeHTTP(catalogRes, catalogReq)
	if catalogRes.Code != http.StatusOK {
		t.Fatalf("expected 200 from catalog, got %d", catalogRes.Code)
	}

	var catalog catalogResponse
	if err := json.NewDecoder(catalogRes.Body).Decode(&catalog); err != nil {
		t.Fatalf("decode catalog: %v", err)
	}
	if len(catalog.MapItems) == 0 || len(catalog.ScenarioItems) == 0 {
		t.Fatalf("expected catalog items in response")
	}
	if strings.Contains(catalog.MapItems[0].Path, ":/") || filepath.IsAbs(catalog.MapItems[0].Path) {
		t.Fatalf("expected relative catalog path, got %s", catalog.MapItems[0].Path)
	}
}

func TestOverviewAndStreamEndpoints(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	httpServer := httptest.NewServer(server.Handler())
	defer httpServer.Close()

	createResp, err := http.Post(httpServer.URL+"/api/runs", "application/json", bytes.NewBufferString(`{"algorithm":"adaptive","mapPath":"fixtures/maps/factory-small.json","scenarioPath":"fixtures/scenarios/baseline.json","seed":999}`))
	if err != nil {
		t.Fatalf("create run request: %v", err)
	}
	defer createResp.Body.Close()
	if createResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 from create run, got %d", createResp.StatusCode)
	}

	var run runResponse
	if err := json.NewDecoder(createResp.Body).Decode(&run); err != nil {
		t.Fatalf("decode run: %v", err)
	}

	overviewResp, err := http.Get(httpServer.URL + "/api/runs/" + run.ID + "/overview")
	if err != nil {
		t.Fatalf("overview request: %v", err)
	}
	defer overviewResp.Body.Close()
	if overviewResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 from overview, got %d", overviewResp.StatusCode)
	}

	var overview runOverviewResponse
	if err := json.NewDecoder(overviewResp.Body).Decode(&overview); err != nil {
		t.Fatalf("decode overview: %v", err)
	}
	if overview.Run.ID != run.ID {
		t.Fatalf("expected overview for run %s, got %s", run.ID, overview.Run.ID)
	}

	streamReq, err := http.NewRequest(http.MethodGet, httpServer.URL+"/api/runs/"+run.ID+"/stream", nil)
	if err != nil {
		t.Fatalf("create stream request: %v", err)
	}
	streamResp, err := http.DefaultClient.Do(streamReq)
	if err != nil {
		t.Fatalf("open stream: %v", err)
	}
	defer streamResp.Body.Close()
	if streamResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 from stream, got %d", streamResp.StatusCode)
	}

	reader := bufio.NewReader(streamResp.Body)
	initial, err := readSSEPayload(reader)
	if err != nil {
		t.Fatalf("read initial stream event: %v", err)
	}
	if !strings.Contains(initial, `"kind":"run.snapshot"`) {
		t.Fatalf("expected initial snapshot event, got %s", initial)
	}

	stepResp, err := http.Post(httpServer.URL+"/api/runs/"+run.ID+"/step", "application/json", nil)
	if err != nil {
		t.Fatalf("step run request: %v", err)
	}
	_ = stepResp.Body.Close()

	streamEvent, err := readSSEPayload(reader)
	if err != nil {
		t.Fatalf("read streamed step event: %v", err)
	}
	if !strings.Contains(streamEvent, `"kind":"run.step"`) {
		t.Fatalf("expected run.step event, got %s", streamEvent)
	}
}

func TestCreateRunRejectsIncompatibleMapAndScenario(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/runs",
		bytes.NewBufferString(`{"algorithm":"fifo","mapPath":"fixtures/maps/factory-branching.json","scenarioPath":"fixtures/scenarios/baseline.json","seed":123}`),
	)
	res := httptest.NewRecorder()
	server.Handler().ServeHTTP(res, req)
	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for incompatible map/scenario, got %d: %s", res.Code, res.Body.String())
	}
	if !strings.Contains(res.Body.String(), "is not present on map") {
		t.Fatalf("expected compatibility error in response, got %s", res.Body.String())
	}
}

func setupTestServer(t *testing.T) (*Server, func()) {
	t.Helper()

	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatalf("resolve root: %v", err)
	}
	cfg := config.AppConfig{
		HTTPPort:            8080,
		SQLitePath:          filepath.Join(root, "data", "test-api.db"),
		DefaultMapPath:      "fixtures/maps/factory-small.json",
		DefaultScenarioPath: "fixtures/scenarios/baseline.json",
		TickDurationMillis:  1,
		LogLevel:            "debug",
	}
	_ = os.Remove(cfg.SQLitePath)
	runStore, err := store.NewSQLiteStore(cfg.SQLitePath)
	if err != nil {
		t.Fatalf("create sqlite store: %v", err)
	}

	service := application.NewService(application.Options{
		DefaultMapPath:      cfg.DefaultMapPath,
		DefaultScenarioPath: cfg.DefaultScenarioPath,
		TickDurationMillis:  cfg.TickDurationMillis,
	}, assets.NewFilesystemAssets(root), runStore, nil, nil)

	return NewServer(service, nil), func() {
		_ = runStore.Close()
		_ = os.Remove(cfg.SQLitePath)
	}
}

func readSSEPayload(reader *bufio.Reader) (string, error) {
	deadline := time.Now().Add(2 * time.Second)
	var lines []string
	for time.Now().Before(deadline) {
		line, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		line = strings.TrimSpace(line)
		if line == "" {
			if len(lines) > 0 {
				return strings.Join(lines, "\n"), nil
			}
			continue
		}
		if strings.HasPrefix(line, ":") {
			continue
		}
		lines = append(lines, line)
	}
	return "", os.ErrDeadlineExceeded
}
