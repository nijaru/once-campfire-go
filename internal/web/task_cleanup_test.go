package web

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/basecamp/once-campfire-go/internal/jobs"
)

// An admitted purge may never start when another job exhausts the drain grace.
// Its detached root must survive queue cancellation and finish before DB teardown.
func TestCloseResumesPurgeCancelledBeforeExecution(t *testing.T) {
	t.Setenv("JOB_CONCURRENCY", "1")
	app, _, _, _ := testApp(t)
	ctx := context.Background()
	blob, err := app.Storage.Stage(ctx, "detached.txt", "text/plain", strings.NewReader("detached"))
	if err != nil {
		t.Fatal(err)
	}
	path, err := app.Storage.Path(blob.Key)
	if err != nil {
		t.Fatal(err)
	}
	// Occupy the worker so the actual cleanup cannot have begun.
	entered := make(chan struct{})
	blocked := blockedPurge{entered: entered}
	if got := app.Jobs.Enqueue(blocked); got != jobs.Accepted {
		t.Fatal(got)
	}
	<-entered
	if err = app.MessageCommands.Cleanup.Detached([]int64{blob.ID}); err != nil {
		t.Fatal(err)
	}
	app.Jobs.Close(0)
	app.Close()
	if _, err = app.DB.Blob(ctx, blob.ID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("cancelled purge lost its detached root: %v", err)
	}
	if _, err = os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("cancelled purge left its file: %v", err)
	}
}

type blockedPurge struct{ entered chan struct{} }

func (blockedPurge) Queue() string { return "purge" }
func (task blockedPurge) Run(ctx context.Context) error {
	task.entered <- struct{}{}
	<-ctx.Done()
	return nil
}
