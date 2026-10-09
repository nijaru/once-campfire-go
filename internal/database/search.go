package database

import (
	"context"
	"database/sql"
	"slices"
	"strings"
	"unicode"
)

func SearchQuery(query string) string {
	return strings.Map(func(c rune) rune {
		if unicode.IsLetter(c) || unicode.IsNumber(c) || unicode.IsMark(c) ||
			unicode.Is(unicode.Pc, c) {
			return c
		}
		return ' '
	}, query)
}

// SearchReferences returns scoped identities. Use MessageRead when preparing
// misses so result membership and matching bodies remain one observation.
func (d *DB) SearchReferences(ctx context.Context, user int64, query string) ([]MessageReference, error) {
	read, err := d.BeginMessageRead(ctx)
	if err != nil {
		return nil, err
	}
	defer read.Close()
	refs, err := read.SearchReferences(ctx, user, query)
	if err != nil {
		return nil, err
	}
	if err = read.Finish(); err != nil {
		return nil, err
	}
	return refs, nil
}

// Search completes selection and matching bodies in the same scoped observation.
func (d *DB) Search(ctx context.Context, user int64, query string) ([]Message, error) {
	read, err := d.BeginMessageRead(ctx)
	if err != nil {
		return nil, err
	}
	defer read.Close()
	refs, err := read.SearchReferences(ctx, user, query)
	if err != nil {
		return nil, err
	}
	records, err := read.Records(ctx, refs)
	if err != nil {
		return nil, err
	}
	return records, read.Finish()
}

const searchProbeLimit = 1000

// Current upstream search selects the newest matching IDs, not timestamp windows.
// Limit the global FTS probe; sparse membership then uses a scoped fallback, all
// within the caller's transaction so selection cannot escape its matching bodies.
func searchReferences(ctx context.Context, tx *sql.Tx, user int64, terms string) ([]MessageReference, error) {
	rows, err := tx.QueryContext(ctx, "SELECT m.id,m.room_id,m.updated_at,member.user_id IS NOT NULL FROM message_search_index idx JOIN messages m ON m.id=idx.rowid LEFT JOIN memberships member ON member.room_id=m.room_id AND member.user_id=? WHERE idx.body MATCH ? ORDER BY idx.rowid DESC LIMIT 1000", user, terms)
	if err != nil {
		return nil, err
	}
	refs, examined, err := searchReferenceRows(rows)
	if err != nil {
		return nil, err
	}
	if len(refs) < 100 && examined == searchProbeLimit {
		rows, err = tx.QueryContext(ctx, "SELECT m.id,m.room_id,m.updated_at,1 FROM messages m JOIN message_search_index idx ON idx.rowid=m.id JOIN memberships member ON member.room_id=m.room_id WHERE member.user_id=? AND idx.body MATCH ? ORDER BY m.id DESC LIMIT 100", user, terms)
		if err != nil {
			return nil, err
		}
		refs, _, err = searchReferenceRows(rows)
		if err != nil {
			return nil, err
		}
	}
	slices.Reverse(refs)
	return refs, nil
}

func searchReferenceRows(rows *sql.Rows) ([]MessageReference, int, error) {
	defer rows.Close()
	refs := []MessageReference{}
	examined := 0
	for rows.Next() {
		examined++
		var ref MessageReference
		var version any
		var reachable bool
		if err := rows.Scan(&ref.ID, &ref.RoomID, &version, &reachable); err != nil {
			return nil, examined, err
		}
		// Inaccessible rows, including malformed foreign timestamps, cannot
		// affect a user's results or reveal their contents through an error.
		if !reachable {
			continue
		}
		if err := (timestamp{&ref.UpdatedAt}).Scan(version); err != nil {
			return nil, examined, err
		}
		refs = append(refs, ref)
		if len(refs) == 100 {
			break
		}
	}
	return refs, examined, rows.Err()
}

func searchTerms(query string) string {
	words := strings.Fields(SearchQuery(query))
	for i, word := range words {
		words[i] = "\"" + strings.ReplaceAll(word, "\"", "\"\"") + "\""
	}
	return strings.Join(words, " ")
}

func (d *DB) RecordSearch(ctx context.Context, user int64, query string) error {
	return d.Transaction(ctx, func(tx *sql.Tx) error {
		now := Stamp(d.Now())
		var id int64
		err := tx.QueryRowContext(ctx, "SELECT id FROM searches WHERE user_id=? AND query=? LIMIT 1", user, query).
			Scan(&id)
		if err == sql.ErrNoRows {
			result, e := tx.ExecContext(
				ctx,
				"INSERT INTO searches(user_id,query,created_at,updated_at) VALUES (?,?,?,?)",
				user,
				query,
				now,
				now,
			)
			if e != nil {
				return e
			}
			id, e = result.LastInsertId()
			if e != nil {
				return e
			}
			if _, e = tx.ExecContext(ctx, "DELETE FROM searches WHERE user_id=? AND id NOT IN (SELECT id FROM searches WHERE user_id=? ORDER BY updated_at DESC LIMIT 10)", user, user); e != nil {
				return e
			}
		} else if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, "UPDATE searches SET updated_at=? WHERE id=?", now, id)
		return err
	})
}

func (d *DB) ClearSearches(ctx context.Context, user int64) error {
	_, err := d.Write.ExecContext(ctx, "DELETE FROM searches WHERE user_id=?", user)
	return err
}

func (d *DB) RecentSearches(ctx context.Context, user int64) ([]string, error) {
	rows, err := d.Read.QueryContext(
		ctx,
		"SELECT query FROM searches WHERE user_id=? ORDER BY updated_at DESC",
		user,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	queries := []string{}
	for rows.Next() {
		var q string
		if err = rows.Scan(&q); err != nil {
			return nil, err
		}
		queries = append(queries, q)
	}
	return queries, rows.Err()
}
