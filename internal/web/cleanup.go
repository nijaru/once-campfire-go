package web

import (
	"context"
	"database/sql"
	"errors"
)

func (s *Server) initCleanup() {
	s.DB.PurgeBlobs = func(ids []int64) {
		for _, id := range ids {
			s.Jobs.Enqueue("purge", func(ctx context.Context) error { return s.Storage.Purge(ctx, id) })
		}
	}
	s.DB.RemoveBannedContent = func(id int64) {
		s.Jobs.Enqueue("ban", func(ctx context.Context) error {
			messages, err := s.DB.MessagesByCreator(ctx, id)
			if err != nil {
				return err
			}
			for _, message := range messages {
				result, err := s.MessageCommands.RemoveBanned(ctx, message.ID)
				if errors.Is(err, sql.ErrNoRows) {
					continue
				} else if err != nil {
					return err
				}
				message = result.Commit.Message
				s.publish(message.RoomID, stream("remove", "message_"+message.ClientID, ""))
				if result.Processing != nil {
					return result.Processing
				}
			}
			return nil
		})
	}
}
