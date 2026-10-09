package application

import (
	"cmp"
	"context"
	"slices"
	"time"

	"github.com/basecamp/once-campfire-go/internal/cable"
	"github.com/basecamp/once-campfire-go/internal/database"
	"github.com/basecamp/once-campfire-go/internal/presentation"
	"github.com/basecamp/once-campfire-go/internal/rails"
)

// RoomPublications owns postcommit sidebar effects without retaining HTTP state.
type RoomPublications struct {
	DB           *database.DB
	Presentation *presentation.Renderer
	Cable        *cable.Hub
}

func (s *RoomPublications) Saved(ctx context.Context, room database.Room, update bool) error {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
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
	audience, err := s.DB.RoomPublicationAudience(ctx, room)
	if err != nil || len(audience.Active) == 0 {
		return err
	}
	if room.Type != "Rooms::Direct" {
		// A shared-room entry has no viewer-dependent data. Render it once for
		// the selected recipient set, rather than once per membership.
		markup, err := s.Presentation.Markup("sidebar-shared", presentation.RoomView{Room: room})
		if err != nil {
			return err
		}
		output := rails.TurboStream(action, target, markup)
		for _, id := range audience.Active {
			s.Cable.PublishStream(ctx, rails.UserRoomsStream(id), output)
		}
		return nil
	}
	for _, id := range audience.Active {
		// DisplayRoom needs the actual viewer for a self-ping's retained name.
		i, _ := slices.BinarySearchFunc(audience.Participants, id, func(member database.RoomParticipant, id int64) int {
			return cmp.Compare(member.ID, id)
		})
		view := presentation.DisplayRoom(room, audience.Participants, audience.Participants[i])
		markup, err := s.Presentation.Markup("sidebar-direct", view)
		if err != nil {
			return err
		}
		s.Cable.PublishStream(ctx, rails.UserRoomsStream(id), rails.TurboStream("prepend", "direct_rooms", markup))
	}
	return nil
}

func (s *RoomPublications) Removed(ctx context.Context, result RoomResult) error {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	s.Cable.PublishStream(ctx, "rooms", rails.TurboStream("remove", result.Commit.Room.DOM("list"), ""))
	return result.Processing
}

func (s *RoomPublications) Involvement(ctx context.Context, user int64, commit database.InvolvementCommit) error {
	if commit.Type == "Rooms::Direct" || commit.Value != "invisible" && commit.Previous != "invisible" {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	var output string
	if commit.Value == "invisible" {
		output = rails.TurboStream("remove", commit.DOM("list"), "")
	} else {
		markup, err := s.Presentation.Markup("sidebar-shared", presentation.RoomView{Room: commit.Room})
		if err != nil {
			return err
		}
		output = rails.TurboStream("prepend", "shared_rooms", markup)
	}
	s.Cable.PublishStream(ctx, rails.UserRoomsStream(user), output)
	return nil
}
