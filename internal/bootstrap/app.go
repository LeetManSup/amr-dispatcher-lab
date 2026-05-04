package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"amr-dispatcher-lab/internal/application"
	"amr-dispatcher-lab/internal/assets"
	"amr-dispatcher-lab/internal/config"
	"amr-dispatcher-lab/internal/logging"
	"amr-dispatcher-lab/internal/presentation/web"
	"amr-dispatcher-lab/internal/store"
)

type App struct {
	Workspace  string
	ConfigPath string
	Config     config.AppConfig
	HTTPAddr   string
	Logger     *slog.Logger
	LogPath    string
	Store      *store.SQLiteStore
	Service    *application.Service
	WebServer  *web.Server
	HTTPServer *http.Server

	closeOnce sync.Once
	logCloser io.Closer
}

func NewFromEnvironment() (*App, error) {
	workspace, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("resolve workspace: %w", err)
	}
	configPath := filepath.Join(workspace, "config", "app.json")
	if override := os.Getenv("AMR_DISP_CONFIG"); override != "" {
		configPath = override
	}
	return New(workspace, configPath)
}

func New(workspace, configPath string) (*App, error) {
	cfg, err := config.LoadAppConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	logger, logCloser, logPath, err := logging.New(workspace, cfg.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("init logger: %w", err)
	}
	sqlitePath := cfg.SQLitePath
	if !filepath.IsAbs(sqlitePath) {
		sqlitePath = filepath.Join(workspace, sqlitePath)
	}
	runStore, err := store.NewSQLiteStore(sqlitePath)
	if err != nil {
		_ = logCloser.Close()
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	service := application.NewService(application.Options{
		DefaultMapPath:      cfg.DefaultMapPath,
		DefaultScenarioPath: cfg.DefaultScenarioPath,
		TickDurationMillis:  cfg.TickDurationMillis,
	}, assets.NewFilesystemAssets(workspace), runStore, nil, logger)

	webServer := web.NewServer(service, logger)
	addr := ":" + formatPort(cfg.HTTPPort)
	app := &App{
		Workspace:  workspace,
		ConfigPath: configPath,
		Config:     cfg,
		HTTPAddr:   addr,
		Logger:     logger,
		LogPath:    logPath,
		Store:      runStore,
		Service:    service,
		WebServer:  webServer,
		HTTPServer: newHTTPServer(addr, webServer.Handler()),
		logCloser:  logCloser,
	}
	app.Logger.Info("application assembled", "http_addr", addr, "config_path", configPath, "log_path", logPath)
	return app, nil
}

func (a *App) Handler() http.Handler {
	return a.WebServer.Handler()
}

func (a *App) SetHTTPAddr(addr string) {
	a.HTTPAddr = addr
	a.HTTPServer = newHTTPServer(addr, a.Handler())
}

func (a *App) Start() error {
	if a.HTTPServer == nil {
		a.HTTPServer = newHTTPServer(a.HTTPAddr, a.Handler())
	}
	if a.Logger != nil {
		a.Logger.Info("http server starting", "http_addr", a.HTTPAddr)
	}
	err := a.HTTPServer.ListenAndServe()
	if err == nil || err == http.ErrServerClosed {
		if a.Logger != nil {
			a.Logger.Info("http server stopped", "http_addr", a.HTTPAddr)
		}
		return nil
	}
	if a.Logger != nil {
		a.Logger.Error("http server failed", "http_addr", a.HTTPAddr, "err", err)
	}
	return err
}

func (a *App) Shutdown(ctx context.Context) error {
	if a.HTTPServer != nil {
		return a.HTTPServer.Shutdown(ctx)
	}
	return nil
}

func (a *App) Close() error {
	var closeErr error
	a.closeOnce.Do(func() {
		var errs []error
		if a.Store != nil {
			if err := a.Store.Close(); err != nil {
				errs = append(errs, err)
			}
		}
		if a.logCloser != nil {
			if err := a.logCloser.Close(); err != nil {
				errs = append(errs, err)
			}
		}
		closeErr = errors.Join(errs...)
	})
	return closeErr
}

func formatPort(value int) string {
	if value == 0 {
		return "8080"
	}
	return fmt.Sprintf("%d", value)
}

func newHTTPServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:    addr,
		Handler: handler,
	}
}
