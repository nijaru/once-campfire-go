package web

import (
	"context"
	"regexp"
	"strconv"
	"testing"
)

func TestSearchShowsFreshResultCount(t *testing.T) {
	app, server, cookie, user := testApp(t)
	ctx := context.Background()
	rooms, err := app.DB.Rooms(ctx, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	first, err := app.DB.CreateMessage(ctx, user.ID, rooms[0].ID, "count-first", "<p>countneedle</p>", "countneedle")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.DB.CreateMessage(ctx, user.ID, rooms[0].ID, "count-second", "<p>countneedle</p>", "countneedle"); err != nil {
		t.Fatal(err)
	}
	counter := regexp.MustCompile(`(?s)class="searches__query[^"\n]*".*?<span class="flex-item-no-shrink">([0-9]+)</span>`)
	check := func(query string, want int) {
		t.Helper()
		response, body := perform(t, server, "GET", "/searches?q="+query, "", nil, cookie)
		match := counter.FindSubmatch(body)
		if response.StatusCode != 200 || len(match) != 2 || string(match[1]) != strconv.Itoa(want) {
			t.Fatalf("search %q: status %d, badge %q, want %d", query, response.StatusCode, match, want)
		}
	}
	check("countneedle", 2)
	check("countneedle", 2) // A message-list cache hit must retain the count too.
	if err := app.DB.DeleteMessage(ctx, user.ID, first.ID); err != nil {
		t.Fatal(err)
	}
	check("countneedle", 1)
	check("absentneedle", 0)
}
