package gloader

import (
	"context"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/gopkgsquad/glogger"
)

type Watcher struct {
	logger         *glogger.Logger
	firstStart     bool
	reloadActive   bool
	SourceDir      string
	Interval       time.Duration
	lastCheckTime  time.Time
	server         *http.Server
	serverStopChan chan struct{}
}

func NewWatcher(server *http.Server, interval time.Duration, logger *glogger.Logger) *Watcher {
	sourceDir := GetRootPath()
	return &Watcher{
		logger:         logger,
		firstStart:     true,
		SourceDir:      sourceDir,
		Interval:       interval,
		lastCheckTime:  time.Now(),
		server:         server,
		reloadActive:   false,
		serverStopChan: make(chan struct{}),
	}
}

func (w *Watcher) Start() {
	if w.firstStart {
		go func() {
			w.firstStart = false
			w.startServer()
		}()
	}

	w.logger.Info("waiting for file changes...")

	var mainGoPath string
	filepath.Walk(w.SourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && info.Name() == "main.go" {
			mainGoPath = path
			return filepath.SkipDir // Skip further traversal
		}
		return nil
	})

	if mainGoPath == "" {
		w.logger.Fatal("main.go file not found in source directory")
	}

	for {
		w.checkChanges(mainGoPath)
		time.Sleep(w.Interval)
	}
}

func (w *Watcher) checkChanges(mainGoPath string) {
	now := time.Now()
	if now.Sub(w.lastCheckTime) >= w.Interval {
		err := filepath.Walk(w.SourceDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() && filepath.Ext(path) == ".go" {
				modTime := info.ModTime()
				if modTime.After(w.lastCheckTime) {
					w.logger.Infof("🔄 Update Alert: File '%s' has been modified. 🔄\n", info.Name())
					w.lastCheckTime = now
					// Now you can use `mainGoFullPath` in your reload function
					if err := w.reload(mainGoPath); err != nil {
						w.logger.Errorf("Error reloading application: %s", err.Error())
					}

				}
			}
			return nil
		})
		if err != nil {
			w.logger.Errorf("Error walking directory: %s", err.Error())
		}
	}
}

func (w *Watcher) reload(mainGoFullPath string) error {
	w.stopServer()
	<-w.serverStopChan
	w.logger.Warning("🔄 Reload Alert: Reloading application... 🔄")
	cmd := exec.Command("go", "run", mainGoFullPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		w.logger.Fatalf("reloading application failed due to: %s", err.Error())
	}

	if err := cmd.Wait(); err != nil {
		w.logger.Fatalf("restarting application failed due to: %s", err.Error())
	}
	return nil
}

func (w *Watcher) startServer() {
	if w.server != nil {
		w.server.ListenAndServe()
	}
}

func (w *Watcher) stopServer() {
	if w.server != nil {
		if err := w.server.Shutdown(context.Background()); err != nil {
			w.logger.Fatalf("stopping server failed due to: %s", err.Error())
		} else {
			close(w.serverStopChan)
		}
	}
}
