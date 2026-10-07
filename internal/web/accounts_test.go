package web

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"testing"
)

func TestAccountSettingBooleanCasting(t *testing.T) {
	app, server, cookie, _ := testApp(t)
	field := "account[settings][restrict_room_creation_to_administrators]"
	for _, tc := range []struct {
		value string
		want  any
	}{
		{"1", true},
		{"off", false},
		{"OFF", false},
		{"f", false},
		{"F", false},
		{"FALSE", false},
		{"false", false},
		{"0", false},
		{"", nil},
	} {
		response, body := perform(
			t,
			server,
			"PATCH",
			"/account",
			"application/x-www-form-urlencoded",
			strings.NewReader(url.Values{field: {tc.value}}.Encode()),
			cookie,
		)
		if response.StatusCode != http.StatusFound {
			t.Fatalf("value %q: %s %s", tc.value, response.Status, body)
		}
		account, err := app.DB.Account(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		var settings map[string]any
		if err := json.Unmarshal(account.Settings, &settings); err != nil {
			t.Fatal(err)
		}
		if settings["restrict_room_creation_to_administrators"] != tc.want {
			t.Errorf("value %q: settings %s, want %v", tc.value, account.Settings, tc.want)
		}
	}
}

func TestCustomStylesSaveReturnsToEditor(t *testing.T) {
	app, server, cookie, _ := testApp(t)
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	server.Client().Jar = jar
	styles := ":root { --contract-color: red; }"
	for _, tc := range []struct{ method, field, value string }{
		{"PATCH", "account[custom_styles]", styles},
		{"PUT", "account[ignored]", "value"},
	} {
		response, body := perform(
			t,
			server,
			tc.method,
			"/account/custom_styles",
			"application/x-www-form-urlencoded",
			strings.NewReader(url.Values{tc.field: {tc.value}}.Encode()),
			cookie,
		)
		if response.StatusCode != http.StatusFound ||
			response.Header.Get("Location") != "/account/custom_styles/edit" {
			t.Errorf("save %s: %s %s", tc.field, response.Status, response.Header.Get("Location"))
		}
		account, err := app.DB.Account(context.Background())
		if err != nil || account.CustomStyles != styles {
			t.Errorf("save %s changed styles: %q, %v", tc.field, account.CustomStyles, err)
		}
		response, body = perform(t, server, "GET", "/account/custom_styles/edit", "", nil, cookie)
		if response.StatusCode != http.StatusOK || !strings.Contains(string(body), styles) ||
			!strings.Contains(string(body), "✓") {
			t.Errorf("save %s: editor missing styles or success notice", tc.field)
		}
	}
}
