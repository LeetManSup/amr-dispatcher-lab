package web

import (
	"html/template"
	"io"
	"log/slog"
	"net/http"

	"amr-dispatcher-lab/internal/application"
)

type Server struct {
	service       *application.Service
	logger        *slog.Logger
	mux           *http.ServeMux
	indexTemplate *template.Template
	uiStatic      http.Handler
}

func NewServer(service *application.Service, logger *slog.Logger) *Server {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	uiFS := mustUIFS()
	server := &Server{
		service:       service,
		logger:        logger.With("component", "web"),
		mux:           http.NewServeMux(),
		indexTemplate: template.Must(template.ParseFS(uiFS, "templates/index.html")),
		uiStatic:      http.StripPrefix("/ui/", http.FileServer(http.FS(uiFS))),
	}
	server.routes()
	return server
}

func (s *Server) Handler() http.Handler { return s.mux }

func (s *Server) routes() {
	s.mux.HandleFunc("/", s.handleIndex)
	s.mux.Handle("/ui/", s.uiStatic)
	s.mux.HandleFunc("/health", s.handleHealth)
	s.mux.HandleFunc("/api/catalog", s.handleCatalog)
	s.mux.HandleFunc("/api/runs", s.handleRuns)
	s.mux.HandleFunc("/api/runs/", s.handleRunRoutes)
}
