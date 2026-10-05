package web

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/basecamp/once-campfire-go/internal/front"
)

func TestCompressedRoomTracksEditsAndRechecksAuthorization(t *testing.T) {
	app, _, cookie, user := testApp(t)
	server := httptest.NewServer(front.Deflate(app))
	defer server.Close()
	server.Client().CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	ctx := context.Background()
	rooms, err := app.DB.Rooms(ctx, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	room := rooms[0].ID
	message, err := app.DB.CreateMessage(ctx, user.ID, room, "compressed-message", "before edit", "before edit")
	if err != nil {
		t.Fatal(err)
	}
	fetch := func(encoding string) (int, http.Header, []byte) {
		t.Helper()
		request, _ := http.NewRequest("GET", fmt.Sprintf("%s/rooms/%d", server.URL, room), nil)
		request.AddCookie(cookie)
		request.Header.Set("Accept-Encoding", encoding)
		response, err := server.Client().Do(request)
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		body, err := io.ReadAll(response.Body)
		if err != nil {
			t.Fatal(err)
		}
		if response.Header.Get("Content-Encoding") == "gzip" {
			reader, err := gzip.NewReader(bytes.NewReader(body))
			if err != nil {
				t.Fatal(err)
			}
			body, err = io.ReadAll(reader)
			reader.Close()
			if err != nil {
				t.Fatal(err)
			}
		}
		return response.StatusCode, response.Header, body
	}
	status, headers, first := fetch("gzip")
	_, _, plain := fetch("identity")
	_, _, cached := fetch("gzip")
	if status != 200 || headers.Get("Content-Encoding") != "gzip" || !bytes.Equal(first, plain) || !bytes.Equal(first, cached) {
		t.Fatal("compressed room differs from identity or repeated response")
	}
	if _, err := app.DB.UpdateMessage(ctx, user.ID, message.ID, "after edit", "after edit"); err != nil {
		t.Fatal(err)
	}
	_, changedHeaders, changed := fetch("gzip")
	if !strings.Contains(string(changed), "after edit") || bytes.Equal(first, changed) || headers.Get("ETag") == changedHeaders.Get("ETag") {
		t.Fatal("compressed room did not invalidate after an edit")
	}
	if _, err := app.DB.Write.ExecContext(ctx, "DELETE FROM memberships WHERE user_id=? AND room_id=?", user.ID, room); err != nil {
		t.Fatal(err)
	}
	status, _, denied := fetch("gzip")
	if status != http.StatusFound || strings.Contains(string(denied), "after edit") {
		t.Fatal("compressed-body reuse bypassed revoked membership", status)
	}
}
