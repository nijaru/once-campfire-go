package web

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func TestRoomUpdatesPreserveOmittedName(t *testing.T) {
	app, server, cookie, owner := testApp(t)
	ctx := context.Background()
	room, err := app.DB.CreateRoom(
		ctx,
		owner.ID,
		"Rooms::Open",
		&sql.NullString{String: "Preserved", Valid: true},
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		method, contentType, body, kind string
		name                            sql.NullString
	}{
		{"PATCH", "application/x-www-form-urlencoded", "room[ignored]=value", "Rooms::Open", sql.NullString{String: "Preserved", Valid: true}},
		{"PUT", "application/json", `{"room":{"ignored":true}}`, "Rooms::Open", sql.NullString{String: "Preserved", Valid: true}},
		{"POST", "application/x-www-form-urlencoded", "_method=patch&room[name]=must-not-overwrite", "Rooms::Open", sql.NullString{String: "Preserved", Valid: true}},
		{"PATCH", "application/x-www-form-urlencoded", fmt.Sprintf("room[ignored]=value&user_ids[]=%d", owner.ID), "Rooms::Closed", sql.NullString{String: "Preserved", Valid: true}},
		{"PATCH", "application/x-www-form-urlencoded", "room[name]=", "Rooms::Open", sql.NullString{Valid: true}},
		{"PATCH", "application/json", `{"room":{"name":null}}`, "Rooms::Open", sql.NullString{}},
		{"PUT", "application/json", `{"room":{"ignored":true}}`, "Rooms::Open", sql.NullString{}},
	} {
		namespace := "opens"
		if tc.kind == "Rooms::Closed" {
			namespace = "closeds"
		}
		before, err := app.DB.FindRoom(ctx, room.ID)
		if err != nil {
			t.Fatal(err)
		}
		path := fmt.Sprintf("/rooms/%s/%d", namespace, room.ID)
		if tc.method == "POST" {
			path += "?room[ignored]=query"
		}
		response, body := perform(
			t,
			server,
			tc.method,
			path,
			tc.contentType,
			strings.NewReader(tc.body),
			cookie,
		)
		if response.StatusCode != http.StatusFound {
			t.Fatalf("%s: %s %s", tc.body, response.Status, body)
		}
		var stored sql.NullString
		if err := app.DB.Read.QueryRowContext(ctx, "SELECT name FROM rooms WHERE id=?", room.ID).Scan(&stored); err != nil {
			t.Fatal(err)
		}
		after, err := app.DB.FindRoom(ctx, room.ID)
		if err != nil || stored != tc.name || after.Type != tc.kind {
			t.Errorf(
				"%s: stored name %+v/%+v, room %+v, error %v",
				tc.body,
				stored,
				tc.name,
				after,
				err,
			)
		}
		if strings.Contains(tc.body, "ignored") && before.Type == after.Type &&
			!before.UpdatedAt.Equal(after.UpdatedAt) {
			t.Errorf(
				"omitted name changed room freshness: %v -> %v",
				before.UpdatedAt,
				after.UpdatedAt,
			)
		}
	}
}

func TestRoomUpdatesRejectMissingRoot(t *testing.T) {
	app, server, cookie, owner := testApp(t)
	ctx := context.Background()
	room, err := app.DB.CreateRoom(
		ctx,
		owner.ID,
		"Rooms::Open",
		&sql.NullString{String: "Protected", Valid: true},
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	before, err := app.DB.RoomMembers(ctx, room.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, namespace := range []string{"opens", "closeds"} {
		for _, tc := range []struct{ contentType, body string }{
			{"application/x-www-form-urlencoded", ""},
			{"application/json", `{"room":{}}`},
		} {
			response, _ := perform(
				t,
				server,
				"PATCH",
				fmt.Sprintf("/rooms/%s/%d", namespace, room.ID),
				tc.contentType,
				strings.NewReader(tc.body),
				cookie,
			)
			if response.StatusCode != http.StatusBadRequest {
				t.Errorf("%s missing root: %s", namespace, response.Status)
			}
			stored, err := app.DB.FindRoom(ctx, room.ID)
			if err != nil || stored != room {
				t.Errorf("missing root changed room: %+v, %v", stored, err)
			}
			after, err := app.DB.RoomMembers(ctx, room.ID)
			if err != nil || !reflect.DeepEqual(before, after) {
				t.Errorf("missing root changed membership: %+v, %v", after, err)
			}
		}
	}
}

func TestRoomUpdatesRedirectWhenMembershipIsMissing(t *testing.T) {
	app, server, cookie, owner := testApp(t)
	ctx := context.Background()
	member, err := app.DB.CreateUser(ctx, "Member", "member@test", "digest", "", 0, nil)
	if err != nil {
		t.Fatal(err)
	}
	room, err := app.DB.CreateRoom(
		ctx,
		owner.ID,
		"Rooms::Open",
		&sql.NullString{String: "Protected", Valid: true},
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := app.DB.UpdateRoom(ctx, room.ID, "Rooms::Closed", nil, []int64{member.ID}); err != nil {
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
