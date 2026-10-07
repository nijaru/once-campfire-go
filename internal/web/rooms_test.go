package web

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestRoomUpdatesRedirectWhenMembershipIsMissing(t *testing.T) {
	app, server, cookie, owner := testApp(t)
	ctx := context.Background()
	member, err := app.DB.CreateUser(ctx, "Member", "member@test", "digest", "", 0, nil)
	if err != nil {
		t.Fatal(err)
	}
	room, err := app.DB.CreateRoom(ctx, owner.ID, "Rooms::Open", "Protected", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := app.DB.UpdateRoom(ctx, room.ID, "Rooms::Closed", room.Name, []int64{member.ID}); err != nil {
		t.Fatal(err)
	}
	for _, namespace := range []string{"opens", "closeds"} {
		for _, id := range []int64{room.ID, room.ID + 1} {
			path := fmt.Sprintf("/rooms/%s/%d", namespace, id)
			response, body := perform(t, server, "PATCH", path, "application/x-www-form-urlencoded",
				strings.NewReader("room%5Bname%5D=forbidden"), cookie)
			if response.StatusCode != http.StatusFound || response.Header.Get("Location") != "/" {
				t.Errorf("inaccessible update %s: %s %s", path, response.Status, body)
			}
		}
	}
	stored, err := app.DB.FindRoom(ctx, room.ID)
	if err != nil || stored.Name != room.Name || stored.Type != "Rooms::Closed" {
		t.Fatalf("inaccessible updates changed the room: %+v %v", stored, err)
	}
}
