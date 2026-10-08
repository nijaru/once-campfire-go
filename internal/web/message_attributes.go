package web

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/basecamp/once-campfire-go/internal/application"
	"github.com/basecamp/once-campfire-go/internal/database"
	"github.com/basecamp/once-campfire-go/internal/rails"
	"github.com/basecamp/once-campfire-go/internal/storage"
)

func requireMessage(w http.ResponseWriter, r *http.Request) bool {
	for key := range r.Form {
		if strings.HasPrefix(key, "message[") {
			return true
		}
	}
	if r.MultipartForm != nil {
		for key := range r.MultipartForm.File {
			if strings.HasPrefix(key, "message[") {
				return true
			}
		}
	}
	http.Error(w, "Missing message parameter", 400)
	return false
}

// Matches MessagesController#update and Messages::ByBotsController#message_params.
func (s *Server) updateMessageAttributes(r *http.Request, user database.User, message database.Message, bodyField, attachmentField string, rawBody *string) (application.MessageResult, error) {
	body := rawBody
	if body == nil && r.Form.Has(bodyField) && !nullParam(r, bodyField) {
		value := r.Form.Get(bodyField)
		body = &value
	}
	var attachment *int64
	var uploaded *storage.Staged
	if r.MultipartForm != nil && len(r.MultipartForm.File[attachmentField]) > 0 {
		var err error
		uploaded, err = s.stageAttachment(r, attachmentField)
		if err != nil {
			return application.MessageResult{}, err
		}
	} else if r.Form.Has(attachmentField) {
		if r.Form.Get(attachmentField) != "" {
			return application.MessageResult{}, errors.New("could not find or build blob: expected attachable")
		}
		attachment = new(int64)
	}
	return s.MessageCommands.Update(r.Context(), user.ID, message.ID, body, attachment, uploaded)
}

// Presentation and transports remain here until their ownership migrations. All
// creation entry points share this post-commit sequence; failures in processing
// or presentation cannot suppress the independent notification obligations.
func (s *Server) createdMessageEffects(ctx context.Context, result application.MessageResult, reply bool) (string, error) {
	commit := result.Commit
	s.messageCreated(commit.Message, commit.Room)
	if !reply {
		s.enqueueWebhooks(commit.Message, commit.Room)
	}
	// Notification preparation has its own bounds; it cannot exhaust display
	// or publication's independently owned deadline.
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	views, err := s.messageViews(ctx, []database.Message{commit.Message})
	if err != nil {
		return "", errors.Join(result.Processing, err)
	}
	markup, err := s.Presentation.Markup("message", views[0])
	if err != nil {
		return "", errors.Join(result.Processing, err)
	}
	output := rails.TurboStream("append", commit.Room.DOM("messages"), markup)
	s.publish(commit.Room.ID, output)
	return output, result.Processing
}

func (s *Server) updatedMessageEffects(ctx context.Context, result application.MessageResult) error {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	message := result.Commit.Message
	views, err := s.messageViews(ctx, []database.Message{message})
	if err != nil {
		return errors.Join(result.Processing, err)
	}
	markup, err := s.Presentation.Markup("presentation", views[0])
	if err != nil {
		return errors.Join(result.Processing, err)
	}
	s.publish(message.RoomID, rails.TurboStream("replace", "presentation_message_"+message.ClientID, markup))
	return result.Processing
}
