package database

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"
	"uuid"

	"github.com/basecamp/once-campfire-go/internal/rails"
	"github.com/basecamp/once-campfire-go/internal/richtext"
)

// MessageInput preserves absent body/attachment attributes. Search text is never
// supplied by a caller: it is derived from the state committed by this command.
type MessageInput struct {
	ClientID   string
	Body       *string
	Attachment *int64
	Upload     BlobStager
}

// MessageCommit contains the record and room observed by the successful mutation.
// Detached blobs are post-commit obligations, not database callbacks.
type MessageCommit struct {
	Message
	Room         Room
	Detached     []int64
	AttachmentID int64
}

func (d *DB) CreateMessage(ctx context.Context, user, room int64, input MessageInput) (MessageCommit, error) {
	return d.createMessage(ctx, user, room, input, true)
}

// CreateWebhookReply is only for an already-authorized queued delivery. Like the
// reference model callback, it does not reapply controller membership checks.
func (d *DB) CreateWebhookReply(ctx context.Context, user, room int64, input MessageInput) (MessageCommit, error) {
	return d.createMessage(ctx, user, room, input, false)
}

func (d *DB) createMessage(ctx context.Context, user, room int64, input MessageInput, checkMembership bool) (MessageCommit, error) {
	if input.Upload != nil {
		defer input.Upload.Discard()
	}
	if input.ClientID == "" {
		input.ClientID = uuid.NewV4().String()
	}
	if input.Body != nil {
		body := richtext.Canonical(*input.Body)
		input.Body = &body
	}
	now := d.Now().UTC()
	result := MessageCommit{Message: Message{RoomID: room, CreatorID: user, ClientID: input.ClientID, CreatedAt: now, UpdatedAt: now}}
	if input.Body != nil {
		result.Body = *input.Body
	}
	err := d.Transaction(ctx, func(tx *sql.Tx) error {
		if checkMembership {
			var allowed bool
			if err := tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM memberships m JOIN users u ON u.id=m.user_id WHERE m.room_id=? AND m.user_id=? AND u.status=0)", room, user).Scan(&allowed); err != nil {
				return err
			}
			if !allowed {
				return ErrForbidden
			}
		}
		if err := commandRoom(ctx, tx, room, &result.Room); err != nil {
			return err
		}
		if err := tx.QueryRowContext(ctx, "SELECT name FROM users WHERE id=?", user).Scan(&result.Creator); err != nil {
			return err
		}
		blob, err := messageUpload(ctx, tx, input)
		if err != nil {
			return err
		}
		result.AttachmentID = blob
		stamp := Stamp(now)
		r, err := tx.ExecContext(ctx, "INSERT INTO messages(client_message_id,creator_id,room_id,created_at,updated_at) VALUES (?,?,?,?,?)", result.ClientID, user, room, stamp, stamp)
		if err != nil {
			return err
		}
		result.ID, err = r.LastInsertId()
		if err != nil {
			return err
		}
		if input.Body != nil {
			if _, err = tx.ExecContext(ctx, "INSERT INTO action_text_rich_texts(name,record_type,record_id,body,created_at,updated_at) VALUES ('body','Message',?,?,?,?)", result.ID, result.Body, stamp, stamp); err != nil {
				return err
			}
		}
		if blob != 0 {
			if _, err = tx.ExecContext(ctx, "INSERT INTO active_storage_attachments(blob_id,record_type,record_id,name,created_at) VALUES (?,'Message',?,'attachment',?)", blob, result.ID, stamp); err != nil {
				return err
			}
		}
		plain, err := messageSearchText(ctx, tx, result.Message)
		if err != nil {
			return err
		}
		for _, q := range []struct {
			sql  string
			args []any
		}{
			{"INSERT INTO message_search_index(rowid,body) VALUES (?,?)", []any{result.ID, plain}},
			{"UPDATE rooms SET updated_at=? WHERE id=?", []any{stamp, room}},
			{"UPDATE memberships SET unread_at=?,updated_at=? WHERE room_id=? AND user_id!=? AND involvement!='invisible' AND (connected_at IS NULL OR connected_at < ?)", []any{stamp, stamp, room, user, Stamp(now.Add(-60 * time.Second))}},
		} {
			if _, err = tx.ExecContext(ctx, q.sql, q.args...); err != nil {
				return err
			}
		}
		result.Room.UpdatedAt = now
		return nil
	})
	if err != nil {
		return MessageCommit{}, err
	}
	if input.Upload != nil {
		input.Upload.Keep()
	}
	return result, nil
}

func (d *DB) UpdateMessage(ctx context.Context, user, id int64, input MessageInput) (MessageCommit, error) {
	if input.Upload != nil {
		defer input.Upload.Discard()
	}
	if input.Body != nil {
		body := richtext.Canonical(*input.Body)
		input.Body = &body
	}
	var result MessageCommit
	err := d.Transaction(ctx, func(tx *sql.Tx) error {
		room, err := messagePermission(ctx, tx, user, id, true)
		if err != nil {
			return err
		}
		m := &result.Message
		if err = tx.QueryRowContext(ctx, messageSelect+"WHERE m.id=?", id).Scan(&m.ID, &m.RoomID, &m.CreatorID, &m.ClientID, &m.Body, &m.Creator, timestamp{&m.CreatedAt}, timestamp{&m.UpdatedAt}); err != nil {
			return err
		}
		if err = commandRoom(ctx, tx, room, &result.Room); err != nil {
			return err
		}
		now := d.Now().UTC()
		stamp := Stamp(now)
		changed := false
		if input.Body != nil {
			var old sql.NullString
			err = tx.QueryRowContext(ctx, "SELECT body FROM action_text_rich_texts WHERE record_type='Message' AND record_id=? AND name='body'", id).Scan(&old)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return err
			}
			if errors.Is(err, sql.ErrNoRows) {
				_, err = tx.ExecContext(ctx, "INSERT INTO action_text_rich_texts(name,record_type,record_id,body,created_at,updated_at) VALUES ('body','Message',?,?,?,?)", id, *input.Body, stamp, stamp)
				changed = true
			} else if !old.Valid || old.String != *input.Body {
				_, err = tx.ExecContext(ctx, "UPDATE action_text_rich_texts SET body=?,updated_at=? WHERE record_type='Message' AND record_id=? AND name='body'", *input.Body, stamp, id)
				changed = true
			}
			if err != nil {
				return err
			}
			m.Body = *input.Body
		}
		attachment := input.Attachment
		if input.Upload != nil {
			blob, err := input.Upload.Insert(ctx, tx)
			if err != nil {
				return err
			}
			attachment = &blob
		}
		if attachment != nil {
			result.Detached, err = AttachmentBlobIDs(ctx, tx, "record_type='Message' AND record_id=? AND name='attachment'", id)
			if err != nil {
				return err
			}
			if _, err = tx.ExecContext(ctx, "DELETE FROM active_storage_attachments WHERE record_type='Message' AND record_id=? AND name='attachment'", id); err != nil {
				return err
			}
			if *attachment != 0 {
				if _, err = tx.ExecContext(ctx, "INSERT INTO active_storage_attachments(blob_id,record_type,record_id,name,created_at) VALUES (?,'Message',?,'attachment',?)", *attachment, id, stamp); err != nil {
					return err
				}
			}
			result.AttachmentID = *attachment
			changed = changed || len(result.Detached) > 0 || *attachment != 0
		}
		if !changed {
			return nil
		}
		plain, err := messageSearchText(ctx, tx, *m)
		if err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, "UPDATE message_search_index SET body=? WHERE rowid=?", plain, id); err != nil {
			return err
		}
		if err = touchMessage(ctx, tx, id, room, stamp); err != nil {
			return err
		}
		m.UpdatedAt = now
		result.Room.UpdatedAt = now
		return nil
	})
	if err != nil {
		return MessageCommit{}, err
	}
	if input.Upload != nil {
		input.Upload.Keep()
	}
	return result, nil
}

func (d *DB) DeleteMessage(ctx context.Context, user, id int64) (MessageCommit, error) {
	return d.deleteMessage(ctx, user, id, true)
}

func (d *DB) RemoveBannedMessage(ctx context.Context, id int64) (MessageCommit, error) {
	return d.deleteMessage(ctx, 0, id, false)
}

func (d *DB) deleteMessage(ctx context.Context, user, id int64, checkPermission bool) (MessageCommit, error) {
	var result MessageCommit
	err := d.Transaction(ctx, func(tx *sql.Tx) error {
		if checkPermission {
			if _, err := messagePermission(ctx, tx, user, id, true); err != nil {
				return err
			}
		}
		m := &result.Message
		if err := tx.QueryRowContext(ctx, messageSelect+"WHERE m.id=?", id).Scan(&m.ID, &m.RoomID, &m.CreatorID, &m.ClientID, &m.Body, &m.Creator, timestamp{&m.CreatedAt}, timestamp{&m.UpdatedAt}); err != nil {
			return err
		}
		if err := commandRoom(ctx, tx, m.RoomID, &result.Room); err != nil {
			return err
		}
		var err error
		result.Detached, err = AttachmentBlobIDs(ctx, tx, "(record_type='Message' AND record_id=?) OR (record_type='ActionText::RichText' AND record_id IN (SELECT id FROM action_text_rich_texts WHERE record_type='Message' AND record_id=?))", id, id)
		if err != nil {
			return err
		}
		for _, q := range []string{
			"DELETE FROM boosts WHERE message_id=?",
			"DELETE FROM message_search_index WHERE rowid=?",
			"DELETE FROM active_storage_attachments WHERE record_type='ActionText::RichText' AND record_id IN (SELECT id FROM action_text_rich_texts WHERE record_type='Message' AND record_id=?)",
			"DELETE FROM active_storage_attachments WHERE record_type='Message' AND record_id=?",
			"DELETE FROM action_text_rich_texts WHERE record_type='Message' AND record_id=?",
			"DELETE FROM messages WHERE id=?",
		} {
			if _, err = tx.ExecContext(ctx, q, id); err != nil {
				return err
			}
		}
		result.Room.UpdatedAt = d.Now().UTC()
		_, err = tx.ExecContext(ctx, "UPDATE rooms SET updated_at=? WHERE id=?", Stamp(result.Room.UpdatedAt), m.RoomID)
		return err
	})
	if err != nil {
		return MessageCommit{}, err
	}
	return result, nil
}

func messageUpload(ctx context.Context, tx *sql.Tx, input MessageInput) (int64, error) {
	if input.Upload != nil {
		return input.Upload.Insert(ctx, tx)
	}
	if input.Attachment != nil {
		return *input.Attachment, nil
	}
	return 0, nil
}

func messageSearchText(ctx context.Context, tx *sql.Tx, message Message) (string, error) {
	// Content failures deliberately retain the reference's empty-text fallback.
	// SQL/cancellation failures are different: the command must roll back.
	var lookupErr error
	cache := map[int64]*richtext.Mention{}
	plain, _ := richtext.PlainText(message.Body, richtext.Context{Resolve: func(token string, _ bool) (*richtext.Mention, error) {
		gid, err := rails.UnverifiedUserGID(token)
		if err != nil {
			return nil, err
		}
		gid, _, _ = strings.Cut(gid, "?")
		parts := strings.Split(gid, "/")
		if len(parts) != 5 || parts[3] != "User" {
			return nil, nil
		}
		id, err := strconv.ParseInt(parts[4], 10, 64)
		if err != nil {
			return nil, nil
		}
		if mention, ok := cache[id]; ok {
			return mention, nil
		}
		var name string
		err = tx.QueryRowContext(ctx, "SELECT name FROM users WHERE id=?", id).Scan(&name)
		if errors.Is(err, sql.ErrNoRows) {
			cache[id] = nil
			return nil, nil
		}
		if err != nil {
			lookupErr = err
			return nil, err
		}
		mention := &richtext.Mention{ID: id, Name: name}
		cache[id] = mention
		return mention, nil
	}})
	if lookupErr != nil {
		return "", lookupErr
	}
	if strings.TrimSpace(plain) != "" {
		return plain, nil
	}
	var filename string
	err := tx.QueryRowContext(ctx, "SELECT b.filename FROM active_storage_attachments a JOIN active_storage_blobs b ON b.id=a.blob_id WHERE a.record_type='Message' AND a.record_id=? AND a.name='attachment' ORDER BY a.id LIMIT 1", message.ID).Scan(&filename)
	if errors.Is(err, sql.ErrNoRows) {
		return plain, nil
	}
	if err != nil {
		return "", err
	}
	return rails.Filename(filename), nil
}
