package web

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/basecamp/once-campfire-go/internal/database"
)

func TestMessageControllersRenderFreshRecords(t *testing.T) {
	app, server, cookie, user := testApp(t)
	ctx := context.Background()
	rooms, err := app.DB.Rooms(ctx, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	message, err := app.DB.CreateMessage(ctx, user.ID, rooms[0].ID, "record-input", "<p>record before</p>", "record before")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.DB.CreateBoost(ctx, user.ID, message.ID, "record boost"); err != nil {
		t.Fatal(err)
	}
	check := func(body string) {
		t.Helper()
		for _, endpoint := range []struct{ path, content string }{
			{fmt.Sprintf("/rooms/%d/messages/%d", message.RoomID, message.ID), body},
			{fmt.Sprintf("/rooms/%d/messages/%d/edit", message.RoomID, message.ID), body},
			{fmt.Sprintf("/messages/%d/boosts", message.ID), "record boost"},
			{fmt.Sprintf("/messages/%d/boosts/new", message.ID), "new_boost_message_" + message.ClientID},
		} {
			response, data := perform(t, server, "GET", endpoint.path, "", nil, cookie)
			if response.StatusCode != 200 || !strings.Contains(string(data), endpoint.content) {
				t.Fatalf("%s: status %d, missing %q", endpoint.path, response.StatusCode, endpoint.content)
			}
		}
	}
	check("record before")
	if _, err := app.DB.UpdateMessage(ctx, user.ID, message.ID, "<p>record after</p>", "record after"); err != nil {
		t.Fatal(err)
	}
	check("record after")
}

func TestRecordedMessagesPreserveBodyAndInvalidate(t *testing.T) {
	app, _, _, user := testApp(t)
	ctx := context.Background()
	rooms, _ := app.DB.Rooms(ctx, user.ID)
	message, err := app.DB.CreateMessage(ctx, user.ID, rooms[0].ID, "recorded", "<p>one &amp; two</p>", "one & two")
	if err != nil {
		t.Fatal(err)
	}
	list := []database.Message{message}
	views, err := app.messageItems(ctx, list)
	if err != nil {
		t.Fatal(err)
	}
	var original bytes.Buffer
	if err := app.templates.ExecuteTemplate(&original, "messages", page{Messages: views}); err != nil {
		t.Fatal(err)
	}
	fragment, err := app.messageList(ctx, list)
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest("GET", "/", nil)
	first := httptest.NewRecorder()
	buffered := &responseBuffer{ResponseWriter: first}
	writeRecorded(buffered, 200, "before\x00marker\x00after", "\x00marker\x00", fragment)
	buffered.finish(request)
	if first.Body.String() != "before"+original.String()+"after" {
		t.Fatal("recorded rendering changed bytes")
	}
	// Keep the existing wire validator even though compression now shares its
	// full strong part identity. This is the pre-optimization construction.
	hash := sha256.New()
	for _, value := range []string{"before", original.String(), "after"} {
		var size [8]byte
		binary.LittleEndian.PutUint64(size[:], uint64(len(value)))
		hash.Write(size[:])
		digest := sha256.Sum256([]byte(value))
		hash.Write(digest[:])
	}
	if first.Header().Get("ETag") != fmt.Sprintf("W/\"%x\"", hash.Sum(nil)[:16]) {
		t.Fatal("part identity changed the wire validator")
	}
	request.Header.Set("If-None-Match", first.Header().Get("ETag"))
	second := httptest.NewRecorder()
	buffered = &responseBuffer{ResponseWriter: second}
	writeRecorded(buffered, 200, "before\x00marker\x00after", "\x00marker\x00", fragment)
	buffered.finish(request)
	if second.Code != 304 || second.Body.Len() != 0 {
		t.Fatal("unchanged parts were not conditional", second.Code)
	}
	list[0].UpdatedAt = list[0].UpdatedAt.Add(time.Second)
	list[0].Body = "<p>changed</p>"
	changed, err := app.messageList(ctx, list)
	if err != nil || !strings.Contains(string(changed.html), "changed") || changed.part.Digest() == fragment.part.Digest() {
		t.Fatal("stale message list", err)
	}
	if changed.part.Digest() != sha256.Sum256([]byte(changed.html)) {
		t.Fatal("incorrect cached digest")
	}
}
