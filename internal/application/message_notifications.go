package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/basecamp/once-campfire-go/internal/cable"
	"github.com/basecamp/once-campfire-go/internal/database"
	"github.com/basecamp/once-campfire-go/internal/integrations"
	"github.com/basecamp/once-campfire-go/internal/jobs"
	"github.com/basecamp/once-campfire-go/internal/presentation"
)

// MessageNotifications owns unread publication and push selection/admission.
// Neither obligation depends on media processing or message presentation success.
type MessageNotifications struct {
	DB      *database.DB
	Queries *NotificationQueries
	Cable   *cable.Hub
	Push    *integrations.PushSender
	Jobs    *jobs.Runner
}

func (s *MessageNotifications) Created(message database.Message, room database.Room) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	members, err := s.DB.RoomMemberIDs(ctx, room.ID)
	if err != nil {
		slog.Error("unread notification failed", "error", err)
	} else {
		for _, id := range members {
			s.Cable.PublishStream(ctx, fmt.Sprintf("user_%d_unreads", id), map[string]any{"roomId": room.ID})
		}
	}
	if s.Push.VAPID == nil {
		return
	}
	deliveries, err := s.Queries.Push(ctx, presentation.Facts{Now: s.DB.Now()}, message, room)
	if err != nil {
		slog.Error("push preparation failed", "error", err)
		return
	}
	for _, delivery := range deliveries {
		subscription, payload := delivery.Subscription, delivery.Payload
		s.Jobs.Enqueue("push", func(ctx context.Context) error {
			err := s.Push.Send(ctx, subscription.Endpoint, subscription.Key, subscription.Auth, payload)
			if errors.Is(err, integrations.ErrPushGone) || errors.Is(err, integrations.ErrPushPoint) {
				return s.DB.DeletePushSubscription(ctx, subscription.UserID, subscription.ID)
			}
			return err
		})
	}
}
