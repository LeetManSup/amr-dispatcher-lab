package web

import (
	"encoding/json"
	"html/template"
	"net/http"
)

type uiShellData struct {
	Title       string
	AppName     string
	Description string
	ConfigJSON  template.JS
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	configBytes, err := json.Marshal(map[string]string{
		"apiBase":  "/api",
		"uiBase":   "/ui",
		"appTitle": "AMR Dispatch Lab",
	})
	if err != nil {
		s.writeError(w, r, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = s.indexTemplate.Execute(w, uiShellData{
		Title:       "AMR Dispatch Lab",
		AppName:     "AMR Dispatch Lab",
		Description: "Research workbench for task dispatch, map traffic and robot fleet analytics.",
		ConfigJSON:  template.JS(configBytes),
	})
	if err != nil {
		s.writeError(w, r, http.StatusInternalServerError, err)
	}
}
