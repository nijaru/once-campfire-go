package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestSearchFindsSparseMembershipBehindInaccessibleMatches(t *testing.T) {
	d := testDB(t)
	ctx := context.Background()
	user, err := d.Setup(ctx, "Owner", "owner@test", "digest")
	if err != nil {
		t.Fatal(err)
	}
	rooms, err := d.Rooms(ctx, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	visible, err := d.CreateMessage(ctx, user.ID, rooms[0].ID, messageInput("", "<p>sparseneedle</p>"))
	if err != nil {
		t.Fatal(err)
	}
	private, err := d.CreateRoom(ctx, user.ID, "Rooms::Closed", &sql.NullString{String: "Private", Valid: true}, nil)
	if err != nil {
		t.Fatal(err)
	}
	// Raw SQL models another implementation writing the shared database.
	err = d.Transaction(ctx, func(tx *sql.Tx) error {
		stamp := Stamp(d.Now())
		for i := 0; i < 1100; i++ {
			result, err := tx.ExecContext(ctx, "INSERT INTO messages(room_id,creator_id,client_message_id,created_at,updated_at) VALUES (?,?,?, ?,?)", private.ID, user.ID, fmt.Sprintf("hidden-%d", i), stamp, stamp)
			if err != nil {
				return err
			}
			id, err := result.LastInsertId()
			if err != nil {
				return err
			}
			if _, err := tx.ExecContext(ctx, "INSERT INTO message_search_index(rowid,body) VALUES (?,?)", id, "sparseneedle"); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	refs, err := d.SearchReferences(ctx, user.ID, "sparseneedle")
	if err != nil || len(refs) != 1 || refs[0].ID != visible.ID {
		t.Fatalf("sparse selection: %v %v", refs, err)
	}
	records, err := d.Search(ctx, user.ID, "sparseneedle")
	if err != nil || len(records) != 1 || records[0].ID != visible.ID {
		t.Fatalf("sparse hydration: %v %v", records, err)
	}
	refs, err = d.SearchReferences(ctx, user.ID+1, "sparseneedle")
	if err != nil || len(refs) != 0 {
		t.Fatalf("inaccessible selection leaked: %v %v", refs, err)
	}
	records, err = d.Search(ctx, user.ID+1, "sparseneedle")
	if err != nil || len(records) != 0 {
		t.Fatalf("inaccessible bodies leaked: %v %v", records, err)
	}
}

func TestSearchReferencesMatchReachableResults(t *testing.T) {
	d := testDB(t)
	ctx := context.Background()
	user, err := d.Setup(ctx, "User", "user@test", "digest")
	if err != nil {
		t.Fatal(err)
	}
	rooms, err := d.Rooms(ctx, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	room := rooms[0].ID
	base := time.Date(2026, 10, 6, 12, 0, 0, 123456000, time.UTC)
	var latest MessageCommit
	for i := 0; i < 104; i++ {
		// Backdated and tied timestamps do not change the current newest-ID contract.
		d.Now = func() time.Time { return base.Add(-time.Duration(i/2) * time.Second) }
		latest, err = d.CreateMessage(ctx, user.ID, room, messageInput("", "<p>wordneedle café AND</p>"))
		if err != nil {
			t.Fatal(err)
		}
	}
	compare := func(userID int64, query string) {
		t.Helper()
		words := strings.Fields(SearchQuery(query))
		for i, word := range words {
			words[i] = "\"" + strings.ReplaceAll(word, "\"", "\"\"") + "\""
		}
		// Independent reference: upstream's membership-scoped newest-ID contract.
		rows, err := d.Read.QueryContext(ctx, messageSelect+"JOIN message_search_index idx ON idx.rowid=m.id JOIN memberships member ON member.room_id=m.room_id WHERE member.user_id=? AND idx.body MATCH ? ORDER BY m.id DESC LIMIT 100", userID, strings.Join(words, " "))
		if err != nil {
			t.Fatal(err)
		}
		want, expectedErr := scanMessages(rows)
		got, err := d.SearchReferences(ctx, userID, query)
		if (err != nil) != (expectedErr != nil) {
			t.Fatalf("query %q errors differ: %v / %v", query, err, expectedErr)
		}
		if err != nil {
			return
		}
		if len(got) != len(want) {
			t.Fatalf("query %q: got %d, want %d", query, len(got), len(want))
		}
		for i, ref := range got {
			full := want[len(want)-1-i]
			if ref.ID != full.ID || ref.RoomID != full.RoomID || !ref.UpdatedAt.Equal(full.UpdatedAt) {
				t.Fatalf("query %q: reference %d differs", query, i)
			}
		}
	}
	for _, query := range []string{"wordneedle", "WORDNEEDLE", "café", "wordneedle AND", "wordneedle/café", "absentneedle"} {
		compare(user.ID, query)
	}
	compare(user.ID+1, "wordneedle")
	for _, value := range []any{"2026-10-06 13:00:00.123456+01:00", []byte("2026-10-06T12:00:00.123456Z"), []byte("invalid")} {
		if _, err := d.Write.ExecContext(ctx, "UPDATE messages SET updated_at=? WHERE id=?", value, latest.ID); err != nil {
			t.Fatal(err)
		}
		compare(user.ID, "wordneedle")
	}
	if _, err := d.Write.ExecContext(ctx, "DELETE FROM memberships WHERE user_id=? AND room_id=?", user.ID, room); err != nil {
		t.Fatal(err)
	}
	compare(user.ID, "wordneedle") // No retained permission/result snapshot.
	if empty, err := d.SearchReferences(ctx, user.ID, " / "); err != nil || len(empty) != 0 {
		t.Fatalf("empty query: %v / %v", empty, err)
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := d.SearchReferences(cancelled, user.ID, "wordneedle"); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled query: %v", err)
	}
}
