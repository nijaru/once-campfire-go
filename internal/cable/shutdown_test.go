package cable

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/basecamp/once-campfire-go/internal/database"
	"github.com/basecamp/once-campfire-go/internal/rails"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

func TestCloseJoinsSocketPresenceCleanup(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	db, err := database.Open(filepath.Join(t.TempDir(), "shutdown.sqlite3"), 2)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	user, err := db.Setup(ctx, "Owner", "owner@test", "digest")
	if err != nil {
		t.Fatal(err)
	}
	rooms, err := db.Rooms(ctx, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	token, err := db.StartSession(ctx, user.ID, "test", "127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}
	secrets, err := rails.NewSecrets("test-secret")
	if err != nil {
		t.Fatal(err)
	}
	hub := New(db, secrets)
	defer hub.Close()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { hub.Serve(w, r, user.User, token) }))
	defer server.Close()
	conn, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http"), &websocket.DialOptions{Subprotocols: []string{"actioncable-v1-json"}})
	if err != nil {
		t.Fatal(err)
	}
	defer conn.CloseNow()
	var frame map[string]any
	if err := wsjson.Read(ctx, conn, &frame); err != nil {
		t.Fatal(err)
	}
	identifier := fmt.Sprintf(`{"channel":"PresenceChannel","room_id":%d}`, rooms[0].ID)
	if err := wsjson.Write(ctx, conn, map[string]string{"command": "subscribe", "identifier": identifier}); err != nil {
		t.Fatal(err)
	}
	if err := wsjson.Read(ctx, conn, &frame); err != nil || frame["type"] != "confirm_subscription" {
		t.Fatalf("presence admission: %v %v", frame, err)
	}
	// Hold the sole writer so cancellation cannot finish the persisted absence.
	tx, err := db.Write.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	closed := make(chan struct{})
	go func() { hub.Close(); close(closed) }()
	select {
	case <-closed:
		t.Fatal("hub shutdown returned before persisted presence cleanup")
	case <-time.After(20 * time.Millisecond):
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	select {
	case <-closed:
	case <-ctx.Done():
		t.Fatal("hub did not join disconnected socket", ctx.Err())
	}
	var connections int
	if err := db.Read.QueryRowContext(ctx, "SELECT connections FROM memberships WHERE user_id=? AND room_id=?", user.ID, rooms[0].ID).Scan(&connections); err != nil || connections != 0 {
		t.Fatalf("presence still owned after shutdown: %d %v", connections, err)
	}
	response, err := server.Client().Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("closed hub admitted another connection: %s", response.Status)
	}
}
