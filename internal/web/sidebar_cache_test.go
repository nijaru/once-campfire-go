package web

import (
	"bytes"
	"html/template"
	"testing"
	"time"

	"github.com/basecamp/once-campfire-go/internal/database"
)

func TestSidebarPartsMatchTemplates(t *testing.T) {
	app, _, _, user := testApp(t)
	original := page{
		User: user, Screen: "sidebar", CanCreateRooms: true,
		RoomsStream: "rooms", UserRoomsStream: "user", VAPIDPublicKey: "push-key",
		Account: database.Account{ID: 1, UpdatedAt: time.Unix(1700000000, 0)},
		SidebarRooms: []sidebarRoom{
			{Room: database.Room{ID: 1, Name: "Chat", Type: "Rooms::Open"}},
		},
	}
	changes := []struct {
		name   string
		change func(*page)
	}{
		{"unchanged", func(*page) {}},
		{
			"profile and escaping",
			func(p *page) { p.User.Name = `<name "quoted">`; p.User.Role = 0 },
		},
		{
			"account logo",
			func(p *page) { p.Account.HasLogo = true; p.Account.UpdatedAt = p.Account.UpdatedAt.Add(time.Second) },
		},
		{
			"custom styles",
			func(p *page) { p.CustomStyles = template.HTML("<style>body{color:red}</style>") },
		},
		{"flash", func(p *page) { p.Notice = "Saved <changes>" }},
		{
			"permission and streams",
			func(p *page) { p.CanCreateRooms = false; p.UserRoomsStream = "fresh" },
		},
	}
	for _, enabled := range []bool{true, false} {
		if !enabled {
			app.fragments = newFragmentCache(0)
		}
		for _, frame := range []bool{false, true} {
			for _, change := range changes {
				t.Run(change.name, func(t *testing.T) {
					p := original
					p.Frame = frame
					change.change(&p)
					fragment, err := app.markup("sidebar-frame", p)
					if err != nil {
						t.Fatal(err)
					}
					input := struct {
						page
						SidebarHTML template.HTML
					}{p, template.HTML(fragment)}
					var want bytes.Buffer
					if err := app.templates.ExecuteTemplate(&want, "sidebar", input); err != nil {
						t.Fatal(err)
					}
					for range 2 {
						parts, err := app.sidebarParts(p)
						if err != nil {
							t.Fatal(err)
						}
						var got bytes.Buffer
						for _, part := range parts {
							if _, err := part.WriteTo(&got); err != nil {
								t.Fatal(err)
							}
						}
						if !bytes.Equal(got.Bytes(), want.Bytes()) {
							t.Fatalf(
								"parts differ from complete template: cache=%t frame=%t",
								enabled,
								frame,
							)
						}
					}
				})
			}
		}
	}
}

func TestSidebarCacheTracksRenderedChanges(t *testing.T) {
	app, _, _, user := testApp(t)
	makePage := func() page {
		return page{
			User: user, CanCreateRooms: true, RoomsStream: "rooms", UserRoomsStream: "user",
			SidebarRooms: []sidebarRoom{
				{Room: database.Room{ID: 1, Name: "Chat", Type: "Rooms::Open"}},
				{
					Room:    database.Room{ID: 2, Type: "Rooms::Direct"},
					Members: []database.RoomParticipant{{ID: 2, Name: "Second Person"}},
				},
			},
			Placeholders: []database.RoomParticipant{{ID: 3, Name: "Third Person"}},
		}
	}
	render := func(p page) string {
		var b bytes.Buffer
		if err := app.templates.ExecuteTemplate(&b, "sidebar-frame", p); err != nil {
			t.Fatal(err)
		}
		return b.String()
	}
	original := makePage()
	body, key := render(original), sidebarCacheKey(original)
	changes := map[string]func(*page){
		"unread":             func(p *page) { p.SidebarRooms[0].Unread = true },
		"rename":             func(p *page) { p.SidebarRooms[0].Name = "Renamed" },
		"membership removed": func(p *page) { p.SidebarRooms = p.SidebarRooms[1:] },
		"room permission":    func(p *page) { p.CanCreateRooms = false },
		"member name":        func(p *page) { p.SidebarRooms[1].Members[0].Name = "Changed Person" },
		"member avatar":      func(p *page) { p.SidebarRooms[1].Members[0].UpdatedAt = time.Now() },
		"own avatar":         func(p *page) { p.User.UpdatedAt = p.User.UpdatedAt.Add(time.Second) },
		"placeholder":        func(p *page) { p.Placeholders[0].Name = "Different Person" },
		"stream":             func(p *page) { p.UserRoomsStream = "different" },
	}
	for name, change := range changes {
		t.Run(name, func(t *testing.T) {
			p := makePage()
			change(&p)
			if render(p) == body {
				t.Fatal("test must change rendered HTML")
			}
			if sidebarCacheKey(p) == key {
				t.Fatal("changed HTML reused cache key")
			}
		})
	}
}
