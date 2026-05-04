package web

import (
	"encoding/json"
	"net/http"
)

func (s *Server) handleICS(runID string, parts []string, w http.ResponseWriter, r *http.Request) {
	if len(parts) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	switch parts[0] {
	case "addTask":
		s.handleICSAddTask(runID, w, r)
	case "getTaskOrderStatus":
		taskID := r.URL.Query().Get("taskId")
		item, err := s.service.GetTaskOrderStatus(runID, taskID)
		if err != nil {
			s.writeError(w, r, http.StatusNotFound, err)
			return
		}
		writeJSON(w, http.StatusOK, toTaskResponse(*item))
	case "cancelTask":
		taskID := r.URL.Query().Get("taskId")
		if err := s.service.CancelTask(runID, taskID); err != nil {
			s.writeError(w, r, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
	case "continueTask":
		taskID := r.URL.Query().Get("taskId")
		if err := s.service.ContinueTask(runID, taskID); err != nil {
			s.writeError(w, r, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "continued"})
	case "updateOrderPointInfo":
		taskID := r.URL.Query().Get("taskId")
		var payload updateOrderPointRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			s.writeError(w, r, http.StatusBadRequest, err)
			return
		}
		if err := s.service.UpdateOrderPointInfo(runID, taskID, payload.TargetPoint); err != nil {
			s.writeError(w, r, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
	case "deviceInfo":
		items, err := s.service.DeviceInfo(runID)
		if err != nil {
			s.writeError(w, r, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, toRobotResponses(items))
	case "getRobotTaskPath":
		taskID := r.URL.Query().Get("taskId")
		path, err := s.service.GetRobotTaskPath(runID, taskID)
		if err != nil {
			s.writeError(w, r, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string][]string{"path": path})
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (s *Server) handleICSAddTask(runID string, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req addTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, r, http.StatusBadRequest, err)
		return
	}
	task, err := s.service.AddTask(runID, toAddTaskInput(req))
	if err != nil {
		s.writeError(w, r, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, toTaskResponse(*task))
}
