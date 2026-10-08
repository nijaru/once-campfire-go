package web

import (
	"bytes"
	"context"
	"log/slog"
	"time"

	"github.com/basecamp/once-campfire-go/internal/database"
	"github.com/basecamp/once-campfire-go/internal/storage"
)

func (s *Server) enqueueWebhooks(message database.Message, room database.Room) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	bots, err := s.NotificationQueries.WebhookRecipients(ctx, s.presentationFacts(ctx), message, room)
	if err != nil {
		slog.Error("webhook recipients failed", "error", err)
		return
	}
	for _, bot := range bots {
		s.Jobs.Enqueue("webhook", func(ctx context.Context) error { return s.deliverWebhook(ctx, bot, message.ID) })
	}
}
func (s *Server) deliverWebhook(ctx context.Context, botID, messageID int64) error {
	delivery, err := s.NotificationQueries.Webhook(ctx, s.presentationFacts(ctx), botID, messageID)
	if err != nil {
		return err
	}
	reply, err := s.Webhooks.Deliver(ctx, delivery.Endpoint, delivery.Payload)
	if err != nil {
		return err
	}
	if reply.Text == nil && reply.Filename == "" {
		return nil
	}
	var staged *storage.Staged
	if reply.Text == nil {
		contentType := storage.Identify(reply.Attachment, reply.Filename, reply.ContentType)
		staged, err = s.Storage.StageFile(ctx, reply.Filename, contentType, bytes.NewReader(reply.Attachment))
		if err != nil {
			return err
		}
	}
	result, err := s.MessageCommands.Reply(ctx, delivery.BotID, delivery.RoomID, reply.Text, staged)
	if err != nil {
		return err
	}
	_, err = s.createdMessageEffects(ctx, result, true)
	return err
}
