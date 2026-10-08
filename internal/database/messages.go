package database

import (
	"context"
	"database/sql"
	"slices"
	"time"
)

// ReachableMessage applies the same membership scope used by message and boost controllers.
func (d *DB) ReachableMessage(ctx context.Context, user, id int64) (Message, error) {
	rows, err := d.Read.QueryContext(
		ctx,
		messageSelect+"JOIN memberships member ON member.room_id=m.room_id WHERE member.user_id=? AND m.id=?",
		user,
		id,
	)
	if err != nil {
		return Message{}, err
	}
	messages, err := scanMessages(rows)
	if err != nil {
		return Message{}, err
	}
	if len(messages) == 0 {
		return Message{}, sql.ErrNoRows
	}
	return messages[0], nil
}

func (d *DB) RefreshedMessages(
	ctx context.Context,
	room int64,
	since time.Time,
) (created, updated []Message, err error) {
	rows, err := d.Read.QueryContext(
		ctx,
		messageSelect+"WHERE m.room_id=? AND m.created_at>? ORDER BY m.created_at LIMIT 40",
		room,
		Stamp(since),
	)
	if err != nil {
		return
	}
	created, err = scanMessages(rows)
	if err != nil {
		return
	}
	rows, err = d.Read.QueryContext(
		ctx,
		messageSelect+"WHERE m.room_id=? AND m.updated_at>? ORDER BY m.created_at DESC LIMIT 40",
		room,
		Stamp(since),
	)
	if err != nil {
		return
	}
	updated, err = scanMessages(rows)
	if err != nil {
		return
	}
	slices.Reverse(updated)
	ids := make(map[int64]bool, len(created))
	for _, m := range created {
		ids[m.ID] = true
	}
	updated = slices.DeleteFunc(updated, func(m Message) bool { return ids[m.ID] })
	return
}

func messagePermission(
	ctx context.Context,
	tx *sql.Tx,
	user, id int64,
	administer bool,
) (room int64, err error) {
	var creator int64
	var role int
	err = tx.QueryRowContext(ctx, "SELECT m.room_id,m.creator_id,u.role FROM messages m JOIN memberships member ON member.room_id=m.room_id JOIN users u ON u.id=member.user_id WHERE m.id=? AND u.id=? AND u.status=0", id, user).
		Scan(&room, &creator, &role)
	if err == nil && administer && creator != user && role != 1 {
		err = ErrForbidden
	}
	return
}

func touchMessage(ctx context.Context, tx *sql.Tx, id, room int64, now string) error {
	if _, err := tx.ExecContext(ctx, "UPDATE messages SET updated_at=? WHERE id=?", now, id); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx, "UPDATE rooms SET updated_at=? WHERE id=?", now, room)
	return err
}

type Boost struct {
	BoosterTitle             string
	BoosterUpdatedAt         time.Time
	ID, MessageID, BoosterID int64
	Content, Booster         string
	CreatedAt, UpdatedAt     time.Time
}

func (d *DB) Boosts(ctx context.Context, message int64) ([]Boost, error) {
	rows, err := d.Read.QueryContext(
		ctx,
		"SELECT b.id,b.message_id,b.booster_id,b.content,u.name,coalesce(u.bio,''),u.updated_at,b.created_at,b.updated_at FROM boosts b JOIN users u ON u.id=b.booster_id WHERE b.message_id=? ORDER BY b.created_at",
		message,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []Boost{}
	for rows.Next() {
		var b Boost
		var bio string
		if err = rows.Scan(&b.ID, &b.MessageID, &b.BoosterID, &b.Content, &b.Booster, &bio, timestamp{&b.BoosterUpdatedAt}, timestamp{&b.CreatedAt}, timestamp{&b.UpdatedAt}); err != nil {
			return nil, err
		}
		b.BoosterTitle = (User{Name: b.Booster, Bio: bio}).Title()
		result = append(result, b)
	}
	return result, rows.Err()
}

func (d *DB) CreateBoost(ctx context.Context, user, message int64, content string) (Boost, error) {
	now := d.Now()
	b := Boost{
		MessageID: message,
		BoosterID: user,
		Content:   content,
		CreatedAt: now,
		UpdatedAt: now,
	}
	err := d.Transaction(ctx, func(tx *sql.Tx) error {
		room, err := messagePermission(ctx, tx, user, message, false)
		if err != nil {
			return err
		}
		var bio string
		if err = tx.QueryRowContext(ctx, "SELECT name,coalesce(bio,''),updated_at FROM users WHERE id=?", user).Scan(&b.Booster, &bio, timestamp{&b.BoosterUpdatedAt}); err != nil {
			return err
		}
		b.BoosterTitle = (User{Name: b.Booster, Bio: bio}).Title()
		result, err := tx.ExecContext(
			ctx,
			"INSERT INTO boosts(message_id,booster_id,content,created_at,updated_at) VALUES (?,?,?,?,?)",
			message,
			user,
			content,
			Stamp(now),
			Stamp(now),
		)
		if err != nil {
			return err
		}
		b.ID, err = result.LastInsertId()
		if err != nil {
			return err
		}
		return touchMessage(ctx, tx, message, room, Stamp(now))
	})
	return b, err
}

func (d *DB) DeleteBoost(ctx context.Context, user, message, id int64) error {
	return d.Transaction(ctx, func(tx *sql.Tx) error {
		room, err := messagePermission(ctx, tx, user, message, false)
		if err != nil {
			return err
		}
		result, err := tx.ExecContext(
			ctx,
			"DELETE FROM boosts WHERE id=? AND message_id=? AND booster_id=?",
			id,
			message,
			user,
		)
		if err != nil {
			return err
		}
		n, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if n == 0 {
			return sql.ErrNoRows
		}
		return touchMessage(ctx, tx, message, room, Stamp(d.Now()))
	})
}

// Message is the unscoped model lookup used by background jobs.
func (d *DB) Message(ctx context.Context, id int64) (Message, error) {
	rows, err := d.Read.QueryContext(ctx, messageSelect+"WHERE m.id=?", id)
	if err != nil {
		return Message{}, err
	}
	messages, err := scanMessages(rows)
	if err != nil {
		return Message{}, err
	}
	if len(messages) == 0 {
		return Message{}, sql.ErrNoRows
	}
	return messages[0], nil
}

func (d *DB) FindRoom(ctx context.Context, id int64) (Room, error) {
	var room Room
	err := d.Read.QueryRowContext(ctx, "SELECT id,creator_id,coalesce(name,''),type,updated_at FROM rooms WHERE id=?", id).
		Scan(&room.ID, &room.CreatorID, &room.Name, &room.Type, timestamp{&room.UpdatedAt})
	return room, err
}
