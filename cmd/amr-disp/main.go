package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"amr-dispatcher-lab/internal/bootstrap"
)

func main() {
	app, err := bootstrap.NewFromEnvironment()
	if err != nil {
		log.Fatalf("bootstrap app: %v", err)
	}

	serverErrCh := make(chan error, 1)
	go func() {
		serverErrCh <- app.Start()
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signals)

	app.Logger.Info("AMR Dispatch Lab listening", "http_addr", app.HTTPAddr, "log_path", app.LogPath)
	exitCode := 0
	select {
	case err := <-serverErrCh:
		if err != nil {
			app.Logger.Error("server stopped", "err", err)
			exitCode = 1
		}
	case sig := <-signals:
		app.Logger.Info("shutdown signal received", "signal", sig.String())
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := app.Shutdown(ctx); err != nil {
			app.Logger.Error("graceful shutdown failed", "err", err)
			exitCode = 1
			break
		}
		if err := <-serverErrCh; err != nil {
			app.Logger.Error("server stopped after shutdown", "err", err)
			exitCode = 1
		}
	}
	if err := app.Close(); err != nil {
		log.Printf("close app: %v", err)
		if exitCode == 0 {
			exitCode = 1
		}
	}
	if exitCode != 0 {
		os.Exit(exitCode)
	}
}
