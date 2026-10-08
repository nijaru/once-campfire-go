package application

import (
	"context"
	"errors"
	"slices"
	"time"

	"github.com/basecamp/once-campfire-go/internal/jobs"
	"github.com/basecamp/once-campfire-go/internal/storage"
)

// Cleanup owns detached identities through admission, including a saturated queue.
type Cleanup struct {
	Storage *storage.Store
	Jobs    *jobs.Runner
}

func (s *Cleanup) Detached(ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	ids = slices.Clone(ids)
	purge := func(ctx context.Context) error {
		var failures error
		for _, id := range ids {
			failures = errors.Join(failures, s.Storage.Purge(ctx, id))
		}
		return failures
	}
	if s.Jobs.Enqueue("purge", purge) {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return purge(ctx)
}
