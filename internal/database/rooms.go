package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"
)

type Membership struct {
	ID, RoomID, UserID    int64
	Involvement           string
	UnreadAt, ConnectedAt *time.Time
	Connections           int
	UpdatedAt             time.Time
}

func (r Room) ParamKey() string {
	switch r.Type {
	case "Rooms::Closed":
		return "rooms_closed"
	case "Rooms::Direct":
		return "rooms_direct"
	default:
		return "rooms_open"
	}
}

func (r Room) DOM(prefix string) string {
	if prefix != "" {
		prefix += "_"
	}
	return fmt.Sprintf("%s%s_%d", prefix, r.ParamKey(), r.ID)
}

func (r Room) EditPath() string {
	return fmt.Sprintf("/rooms/%ss/%d/edit", strings.TrimPrefix(r.ParamKey(), "rooms_"), r.ID)
}

func (r Room) Noun() string {
	if r.Type == "Rooms::Direct" {
		return "Ping"
	}
	return "room"
}

func uniqueIDs(ids []int64) []int64 {
	out := slices.Clone(ids)
	slices.Sort(out)
	return slices.Compact(out)
}

func (d *DB) RoomInvitation(ctx context.Context, room int64) (bool, error) {
	var invitation bool
	err := d.Read.QueryRowContext(ctx, "SELECT ?=(SELECT id FROM rooms ORDER BY created_at LIMIT 1) AND NOT EXISTS(SELECT 1 FROM messages WHERE room_id=? LIMIT 1 OFFSET 40)", room, room).Scan(&invitation)
	return invitation, err
}

func (d *DB) Involvement(ctx context.Context, user, room int64) (string, error) {
	var value sql.NullString
	err := d.Read.QueryRowContext(ctx, "SELECT involvement FROM memberships WHERE user_id=? AND room_id=?", user, room).
		Scan(&value)
	return value.String, err
}

func (d *DB) SetInvolvement(ctx context.Context, user, room int64, value string) error {
	var stored any
	if strings.TrimSpace(value) != "" {
		if !slices.Contains([]string{"invisible", "nothing", "mentions", "everything"}, value) {
			return ErrValidation
		}
		stored = value
	}
	r, err := d.Write.ExecContext(
		ctx,
		"UPDATE memberships SET involvement=?,updated_at=? WHERE user_id=? AND room_id=?",
		stored,
		Stamp(d.Now()),
		user,
		room,
	)
	if err != nil {
		return err
	}
	count, err := r.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (d *DB) Presence(ctx context.Context, user, room int64, action string) error {
	return d.Transaction(ctx, func(tx *sql.Tx) error {
		now := d.Now()
		stamp, cutoff := Stamp(now), Stamp(now.Add(-60*time.Second))
		var query string
		switch action {
		case "present":
			query = "UPDATE memberships SET connections=CASE WHEN connected_at>=? THEN connections+1 ELSE 1 END,connected_at=?,unread_at=NULL WHERE user_id=? AND room_id=?"
		case "refresh":
			query = "UPDATE memberships SET connections=CASE WHEN connected_at>=? THEN connections ELSE 1 END,connected_at=? WHERE user_id=? AND room_id=?"
		case "absent":
			_, err := tx.ExecContext(
				ctx,
				"UPDATE memberships SET connections=CASE WHEN connected_at>=? THEN max(0,connections-1) ELSE 0 END WHERE user_id=? AND room_id=?",
				cutoff,
				user,
				room,
			)
			if err != nil {
				return err
			}
			_, err = tx.ExecContext(
				ctx,
				"UPDATE memberships SET connected_at=NULL WHERE user_id=? AND room_id=? AND connections<1",
				user,
				room,
			)
			return err
		default:
			return ErrValidation
		}
		_, err := tx.ExecContext(ctx, query, cutoff, stamp, user, room)
		return err
	})
}

// OriginalRoom follows Room.original (creation order, not the fixture ID order).
func (d *DB) OriginalRoom(ctx context.Context, user int64) (int64, error) {
	var id int64
	err := d.Read.QueryRowContext(ctx, "SELECT rooms.id FROM rooms JOIN memberships ON memberships.room_id=rooms.id WHERE memberships.user_id=? ORDER BY rooms.created_at LIMIT 1", user).
		Scan(&id)
	return id, err
}

// SidebarRoom owns current membership state and retained direct participants.
type SidebarRoom struct {
	Room
	Involvement string
	Unread      bool
	Members     []RoomParticipant
}

func (d *DB) SidebarRooms(ctx context.Context, user int64) ([]SidebarRoom, error) {
	rows, err := d.Read.QueryContext(
		ctx,
		"SELECT r.id,r.creator_id,coalesce(r.name,''),r.type,r.updated_at,coalesce(m.involvement,''),m.unread_at IS NOT NULL FROM rooms r JOIN memberships m ON m.room_id=r.id WHERE m.user_id=? AND m.involvement!='invisible' ORDER BY lower(r.name)",
		user,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var rooms []SidebarRoom
	for rows.Next() {
		var r SidebarRoom
		if err := rows.Scan(&r.ID, &r.CreatorID, &r.Name, &r.Type, timestamp{&r.UpdatedAt}, &r.Involvement, &r.Unread); err != nil {
			return nil, err
		}
		rooms = append(rooms, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	positions := make(map[int64]int)
	var directIDs []int64
	for i, room := range rooms {
		if room.Type == "Rooms::Direct" {
			positions[room.ID] = i
			directIDs = append(directIDs, room.ID)
		}
	}
	if len(directIDs) == 0 {
		return rooms, nil
	}
	selected, err := json.Marshal(directIDs)
	if err != nil {
		return nil, err
	}
	// Hydrate the retained room set, just as individual RoomMembers reads did.
	// Rechecking viewer visibility here could turn a concurrently hidden ping
	// into a self-ping. Room entry still performs its own fresh authorization.
	rows, err = d.Read.QueryContext(ctx, `
SELECT m.room_id,u.id,u.name,u.updated_at
FROM memberships m JOIN users u ON u.id=m.user_id
WHERE m.room_id IN (SELECT value FROM json_each(?))
ORDER BY m.room_id,m.user_id`, string(selected))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var room int64
		var participant RoomParticipant
		if err := rows.Scan(&room, &participant.ID, &participant.Name, timestamp{&participant.UpdatedAt}); err != nil {
			return nil, err
		}
		i := positions[room]
		rooms[i].Members = append(rooms[i].Members, participant)
	}
	return rooms, rows.Err()
}
