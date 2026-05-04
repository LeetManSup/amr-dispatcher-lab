package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleCatalog(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, toCatalogResponse(s.service.Catalog()))
}

func (s *Server) handleRuns(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		runs, err := s.service.ListRuns()
		if err != nil {
			s.writeError(w, r, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, toRunResponses(runs))
	case http.MethodPost:
		var input createRunRequest
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			s.writeError(w, r, http.StatusBadRequest, err)
			return
		}
		run, err := s.service.CreateRun(toCreateRunInput(input))
		if err != nil {
			s.writeError(w, r, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusCreated, toRunResponse(run))
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleRunRoutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/runs/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	runID := parts[0]
	if len(parts) == 1 {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		run, err := s.service.GetRun(runID)
		if err != nil {
			s.writeError(w, r, http.StatusNotFound, err)
			return
		}
		writeJSON(w, http.StatusOK, toRunResponse(run))
		return
	}
	switch parts[1] {
	case "step":
		s.handleStepRun(runID, w, r)
	case "start":
		s.handleStartRun(runID, w, r)
	case "stop":
		s.handleStopRun(runID, w, r)
	case "tasks":
		s.handleTaskList(runID, w, r)
	case "requests":
		s.handleRequestList(runID, w, r)
	case "robots":
		s.handleRobotList(runID, w, r)
	case "metrics":
		s.handleMetrics(runID, w, r)
	case "segment-load":
		s.handleSegmentLoad(runID, w, r)
	case "map":
		s.handleMap(runID, w, r)
	case "decisions":
		s.handleDecisions(runID, w, r)
	case "export":
		s.handleExport(runID, w, r)
	case "overview":
		s.handleOverview(runID, w, r)
	case "stream":
		s.handleStream(runID, w, r)
	case "ics":
		s.handleICS(runID, parts[2:], w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (s *Server) handleStepRun(runID string, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	result, run, err := s.service.StepRun(runID)
	if err != nil {
		s.writeError(w, r, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"run": toRunResponse(run), "result": toTickResultResponse(result)})
}

func (s *Server) handleStartRun(runID string, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if err := s.service.StartRun(runID); err != nil {
		s.writeError(w, r, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]string{"status": "started"})
}

func (s *Server) handleStopRun(runID string, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if err := s.service.StopRun(runID); err != nil {
		s.writeError(w, r, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "stopped"})
}

func (s *Server) handleTaskList(runID string, w http.ResponseWriter, r *http.Request) {
	items, err := s.service.GetTasks(runID)
	if err != nil {
		s.writeError(w, r, http.StatusNotFound, err)
		return
	}
	writeJSON(w, http.StatusOK, toTaskResponses(items))
}

func (s *Server) handleRequestList(runID string, w http.ResponseWriter, r *http.Request) {
	items, err := s.service.GetRequests(runID)
	if err != nil {
		s.writeError(w, r, http.StatusNotFound, err)
		return
	}
	writeJSON(w, http.StatusOK, toRequestResponses(items))
}

func (s *Server) handleRobotList(runID string, w http.ResponseWriter, r *http.Request) {
	items, err := s.service.GetRobots(runID)
	if err != nil {
		s.writeError(w, r, http.StatusNotFound, err)
		return
	}
	writeJSON(w, http.StatusOK, toRobotResponses(items))
}

func (s *Server) handleMetrics(runID string, w http.ResponseWriter, r *http.Request) {
	items, err := s.service.GetMetrics(runID)
	if err != nil {
		s.writeError(w, r, http.StatusNotFound, err)
		return
	}
	writeJSON(w, http.StatusOK, toMetricsResponses(items))
}

func (s *Server) handleSegmentLoad(runID string, w http.ResponseWriter, r *http.Request) {
	items, err := s.service.GetSegmentLoads(runID)
	if err != nil {
		s.writeError(w, r, http.StatusNotFound, err)
		return
	}
	writeJSON(w, http.StatusOK, toSegmentLoadResponses(items))
}

func (s *Server) handleMap(runID string, w http.ResponseWriter, r *http.Request) {
	item, err := s.service.GetMap(runID)
	if err != nil {
		s.writeError(w, r, http.StatusNotFound, err)
		return
	}
	writeJSON(w, http.StatusOK, toMapResponse(item))
}

func (s *Server) handleDecisions(runID string, w http.ResponseWriter, r *http.Request) {
	items, err := s.service.GetDecisionLogs(runID)
	if err != nil {
		s.writeError(w, r, http.StatusNotFound, err)
		return
	}
	writeJSON(w, http.StatusOK, toDecisionResponses(items))
}

func (s *Server) handleExport(runID string, w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json"
	}
	data, contentType, err := s.service.ExportRun(runID, format)
	if err != nil {
		s.writeError(w, r, http.StatusBadRequest, err)
		return
	}
	w.Header().Set("Content-Type", contentType)
	_, _ = w.Write(data)
}

func (s *Server) handleOverview(runID string, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	item, err := s.service.GetOverview(runID)
	if err != nil {
		s.writeError(w, r, http.StatusNotFound, err)
		return
	}
	writeJSON(w, http.StatusOK, toRunOverviewResponse(item))
}

func (s *Server) handleStream(runID string, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	overview, err := s.service.GetOverview(runID)
	if err != nil {
		s.writeError(w, r, http.StatusNotFound, err)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch, unsubscribe := s.service.Subscribe(runID)
	defer unsubscribe()

	initialEvent := liveEventResponse{
		EventID:      "initial",
		RunID:        runID,
		Kind:         "run.snapshot",
		Message:      "Initial overview snapshot",
		CurrentTick:  overview.Snapshot.Tick,
		RunStatus:    string(overview.Run.Status),
		ActiveTasks:  overview.Snapshot.ActiveTasks,
		WaitingTasks: overview.Snapshot.WaitingTasks,
		OccurredAt:   time.Now().Format(time.RFC3339Nano),
	}
	s.writeSSE(w, "message", initialEvent)
	flusher.Flush()

	keepAlive := time.NewTicker(15 * time.Second)
	defer keepAlive.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case event, ok := <-ch:
			if !ok {
				return
			}
			s.writeSSE(w, "message", toLiveEventResponse(event))
			flusher.Flush()
		case <-keepAlive.C:
			_, _ = fmt.Fprint(w, ": keep-alive\n\n")
			flusher.Flush()
		}
	}
}

func (s *Server) writeSSE(w http.ResponseWriter, eventName string, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}
	_, _ = fmt.Fprintf(w, "event: %s\n", eventName)
	_, _ = fmt.Fprintf(w, "data: %s\n\n", data)
}
