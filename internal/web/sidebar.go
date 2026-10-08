package web

import (
	"context"

	"github.com/basecamp/once-campfire-go/internal/database"
	"github.com/basecamp/once-campfire-go/internal/presentation"
	"github.com/basecamp/once-campfire-go/internal/rails"
)

func (s *Server) broadcastRoom(ctx context.Context, room database.Room, update bool) error {
	action, target := "prepend", "shared_rooms"
	if update {
		action, target = "replace", room.DOM("list")
	}
	if room.Type == "Rooms::Open" {
		markup, err := s.Presentation.Markup("sidebar-shared", presentation.RoomView{Room: room})
		if err != nil {
			return err
		}
		s.Cable.PublishStream(ctx, "rooms", rails.TurboStream(action, target, markup))
		return nil
	}
	members, err := s.DB.ActiveUsers(ctx, room.ID)
	if err != nil {
		return err
	}
	var participants []database.RoomParticipant
	if room.Type == "Rooms::Direct" {
		participants, err = s.DB.RoomParticipants(ctx, room.ID)
		if err != nil {
			return err
		}
	}
	for _, user := range members {
		view := presentation.DisplayRoom(room, participants, user.Participant())
		name := "sidebar-shared"
		if room.Type == "Rooms::Direct" {
			name, target, action = "sidebar-direct", "direct_rooms", "prepend"
		}
		markup, err := s.Presentation.Markup(name, view)
		if err != nil {
			return err
		}
		s.Cable.PublishStream(ctx, rails.UserRoomsStream(user.ID), rails.TurboStream(action, target, markup))
	}
	return nil
}
