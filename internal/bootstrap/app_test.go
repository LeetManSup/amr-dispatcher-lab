package bootstrap

import (
	"context"
	"net"
	"net/http"
	"path/filepath"
	"testing"
	"time"
)

func TestAppStartAndShutdown(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve root: %v", err)
	}
	app, err := New(root, filepath.Join(root, "config", "app.json"))
	if err != nil {
		t.Fatalf("bootstrap new: %v", err)
	}
	defer func() { _ = app.Close() }()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := listener.Addr().String()
	_ = listener.Close()
	app.SetHTTPAddr(addr)

	serverErrCh := make(chan error, 1)
	go func() {
		serverErrCh <- app.Start()
	}()

	client := &http.Client{Timeout: 2 * time.Second}
	var healthy bool
	for i := 0; i < 20; i++ {
		time.Sleep(100 * time.Millisecond)
		resp, err := client.Get("http://" + addr + "/health")
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				healthy = true
				break
			}
		}
	}
	if !healthy {
		t.Fatal("server did not become healthy")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := app.Shutdown(ctx); err != nil {
		t.Fatalf("shutdown: %v", err)
	}
	if err := <-serverErrCh; err != nil {
		t.Fatalf("start returned error after shutdown: %v", err)
	}
}
