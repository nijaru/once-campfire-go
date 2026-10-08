package application

import (
	"context"
	"errors"

	"github.com/basecamp/once-campfire-go/internal/database"
	"github.com/basecamp/once-campfire-go/internal/jobs"
	"github.com/basecamp/once-campfire-go/internal/storage"
)

// Attachment keeps decoded absence/deletion separate from owned staged bytes.
type Attachment struct {
	File     *storage.Staged
	Assigned bool
}

func (a Attachment) input() *database.AttachmentInput {
	if a.File != nil {
		b := a.File.Blob
		return &database.AttachmentInput{Blob: &b}
	}
	if a.Assigned {
		return &database.AttachmentInput{}
	}
	return nil
}

func (a Attachment) Discard() {
	if a.File != nil {
		a.File.Discard()
	}
}

type Attachments struct {
	DB      *database.DB
	Storage *storage.Store
	Jobs    *jobs.Runner
	Cleanup *Cleanup
}

func (s *Attachments) committed(a Attachment, commit database.AttachmentCommit) error {
	// Ownership transfers before any fallible post-commit work.
	if a.File != nil {
		a.File.Commit(*commit.Uploaded)
	}
	err := s.Cleanup.Detached(commit.Detached)
	if commit.Uploaded != nil {
		err = errors.Join(err, s.Analyze(commit.Uploaded.ID))
	}
	return err
}

func (s *Attachments) Analyze(id int64) error {
	analyze := func(ctx context.Context) error {
		blob, err := s.DB.Blob(ctx, id)
		if err != nil {
			return err
		}
		_, err = s.Storage.Analyze(ctx, blob)
		return err
	}
	if s.Jobs.Enqueue("analyze", analyze) {
		return nil
	}
	// Profile/logo analysis remains asynchronous and best-effort. Rejection is
	// explicit to the coordinator, not synchronous media work on the HTTP path.
	return errors.New("blob analysis was not admitted")
}
