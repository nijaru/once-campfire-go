package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"syscall"
	"time"

	"github.com/basecamp/once-campfire-go/internal/database"
	"github.com/basecamp/once-campfire-go/internal/front"
	"github.com/basecamp/once-campfire-go/internal/rails"
	"github.com/basecamp/once-campfire-go/internal/web"
)

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
func main() {
	if err := run(); err != nil {
		slog.Error("campfire", "error", err)
		os.Exit(1)
	}
}
func run() error {
	if path := os.Getenv("GO_CPU_PROFILE"); path != "" {
		file, err := os.Create(path)
		if err != nil {
			return err
		}
		defer file.Close()
		if err := pprof.StartCPUProfile(file); err != nil {
			return err
		}
		defer pprof.StopCPUProfile()
	}

	command := "server"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}
	if command != "server" && command != "db:prepare" && command != "backup" {
		return fmt.Errorf("unknown command %q (server, db:prepare, or backup)", command)
	}
	secrets, err := rails.NewSecrets(os.Getenv("SECRET_KEY_BASE"))
	if err != nil {
		return err
	}
	storage := env("CAMPFIRE_STORAGE_PATH", "storage")
	path := env("CAMPFIRE_DATABASE_PATH", filepath.Join(storage, "db", env("RAILS_ENV", "production")+".sqlite3"))
	db, err := database.Open(path, max(1, runtime.GOMAXPROCS(0)))
	if err != nil {
		return err
	}
	if command == "backup" {
		defer db.Close()
		return db.Backup(context.Background(), filepath.Join(storage, "backups", filepath.Base(path)))
	}
	if command == "db:prepare" {
		return db.Close()
	}
	app, err := web.New(db, secrets, os.Getenv("DISABLE_SSL") == "", storage)
	if err != nil {
		db.Close()
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	closing, finished := make(chan struct{}), make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
		case <-closing:
		case <-finished:
			return
		}
		timer := time.NewTimer(30 * time.Second)
		defer timer.Stop()
		select {
		case <-finished:
		case <-timer.C:
			// Native media work cannot always be preempted. This is forced exit,
			// not successful teardown: persistence stays open under surviving work.
			slog.Error("shutdown grace expired; forcing process exit")
			os.Exit(1)
		}
	}()
	defer func() {
		close(closing)
		app.Close()
		db.Close()
		close(finished)
	}()
	return front.Serve(ctx, front.FromEnv(), app)
}
