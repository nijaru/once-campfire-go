package web

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/basecamp/once-campfire-go/internal/database"
	"github.com/basecamp/once-campfire-go/internal/presentation"
)

// CreatedStream may use commit-owned room/blob inputs instead of re-reading all
// associations. The rendered append must stay byte-identical to the generic path.
func TestCreatedStreamMatchesGenericDisplay(t *testing.T) {
	app, _, _, owner := testApp(t)
	ctx := context.Background()
	rooms, err := app.DB.Rooms(ctx, owner.ID)
	if err != nil {
		t.Fatal(err)
	}
	var open database.Room
	for _, room := range rooms {
		if room.Type != "Rooms::Direct" {
			open = room
			break
		}
	}
	if open.ID == 0 {
		t.Fatal("no open room")
	}
	target, err := app.DB.CreateUser(ctx, owner.ID, database.UserInput{Name: "Target", Email: "target@test"})
	if err != nil {
		t.Fatal(err)
	}
	mention := presentation.Mention(app.Secrets, database.UserDisplay{ID: target.ID, Name: target.Name})
	contentType := "text/plain"
	upload := &database.Blob{Filename: "note.txt", ContentType: &contentType, ByteSize: 1, Metadata: json.RawMessage("{}")}

	direct, err := app.DB.CreateRoom(ctx, owner.ID, "Rooms::Direct", nil, []int64{target.ID})
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name  string
		room  int64
		body  string
		blob  *database.Blob
		boost bool
	}{
		{"open", open.ID, "<p>hello</p>", nil, false},
		{"mentioned", open.ID, fmt.Sprintf(`<p>hi <action-text-attachment sgid="%s"></action-text-attachment></p>`, mention.SGID), nil, false},
		{"upload", open.ID, "<p>with file</p>", upload, false},
		{"open-raced-boost", open.ID, "<p>boosted</p>", nil, true},
		{"direct-fallback", direct.ID, "<p>direct</p>", nil, false},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			body := test.body
			commit, err := app.DB.CreateMessage(ctx, owner.ID, test.room, database.MessageInput{Body: &body, Upload: test.blob})
			if err != nil {
				t.Fatal(err)
			}
			if test.boost {
				if _, err = app.DB.CreateBoost(ctx, target.ID, commit.ID, "🔥"); err != nil {
					t.Fatal(err)
				}
			}
			facts := presentation.Facts{Host: "chat.test", Origin: "http://chat.test", Now: app.DB.Now()}
			got, err := app.MessageQueries.CreatedStream(ctx, facts, commit)
			if err != nil {
				t.Fatal(err)
			}
			prepared := presentation.PrepareMessages([]database.Message{commit.Message})
			data, mentions, err := app.DB.MessageDisplays(ctx, prepared.Records, prepared.Mentioned)
			if err != nil {
				t.Fatal(err)
			}
			views, err := app.Presentation.MessageViews(facts, prepared, data, mentions)
			if err != nil {
				t.Fatal(err)
			}
			want, err := app.Presentation.AppendMessage(commit.Room.DOM("messages"), views[0])
			if err != nil {
				t.Fatal(err)
			}
			if got != want {
				t.Fatalf("created stream differs from generic display:\n got %q\nwant %q", got, want)
			}
		})
	}
}
