package database

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"sync"
	"testing"
	"testing/synctest"
)

func TestResponseVersionWaitingReadIsCancellable(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		d := testDB(t)
		d.version.gate <- struct{}{}
		defer func() { <-d.version.gate }()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		result := make(chan error, 1)
		go func() { _, err := d.ResponseVersion(ctx); result <- err }()
		synctest.Wait()
		cancel()
		synctest.Wait()
		select {
		case err := <-result:
			if !errors.Is(err, context.Canceled) {
				t.Fatal(err)
			}
		default:
			t.Fatal("cancelled query waited for the active generation read")
		}
	})
}

func TestResponseVersionObservesForeignCommitsDuringConcurrentReads(t *testing.T) {
	d := testDB(t)
	ctx := context.Background()
	if _, err := d.Setup(ctx, "User", "user@test", "digest"); err != nil {
		t.Fatal(err)
	}
	var index int
	var name, path string
	if err := d.Read.QueryRow("PRAGMA database_list").Scan(&index, &name, &path); err != nil {
		t.Fatal(err)
	}
	foreign, err := sql.Open("sqlite3", path+"?_busy_timeout=5000")
	if err != nil {
		t.Fatal(err)
	}
	defer foreign.Close()
	stop := make(chan struct{})
	failures := make(chan error, 8)
	var readers sync.WaitGroup
	for range 8 {
		readers.Go(func() {
			for {
				select {
				case <-stop:
					return
				default:
				}
				if _, err := d.ResponseVersion(ctx); err != nil {
					failures <- err
					return
				}
			}
		})
	}
	defer func() { close(stop); readers.Wait() }()
	for i := range 100 {
		before, err := d.ResponseVersion(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := foreign.ExecContext(ctx, "UPDATE accounts SET name=?", strconv.Itoa(i)); err != nil {
			t.Fatal(err)
		}
		after, err := d.ResponseVersion(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if after <= before {
			t.Fatalf("commit %d: generation %d did not advance from %d", i, after, before)
		}
	}
	select {
	case err := <-failures:
		t.Fatal(err)
	default:
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := d.ResponseVersion(cancelled); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled generation query: %v", err)
	}
}
