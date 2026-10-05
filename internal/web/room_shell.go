package web

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"html/template"
	"strings"

	"github.com/basecamp/once-campfire-go/internal/database"
	"github.com/basecamp/once-campfire-go/internal/responsebody"
	"github.com/basecamp/once-campfire-go/internal/useragent"
)

type roomShellEntry struct {
	beforeTime, beforeMessages, afterMessages responsebody.Part
	bytes                                     int
}

// The same bounded input owns both the template and its cache identity. New
// room-template dependencies must enter this type, not an unrelated page field.
type roomShellPage struct {
	User                            database.User
	Room                            database.Room
	Account                         database.Account
	Platform                        useragent.Platform
	Title, BodyClass, Screen        string
	Origin, Stream, VAPIDPublicKey  string
	Notice, Error, LoadedAt         string
	CustomStyles, MessagesHTML      template.HTML
	Messages                        []messageView
	Frame, Chat, Reload, Invitation bool
}

func shellPage(p page) roomShellPage {
	return roomShellPage{
		User: p.User, Room: p.Room, Account: p.Account, Platform: p.Platform,
		Title: p.Title, BodyClass: p.BodyClass, Screen: p.Screen,
		Origin: p.Origin, Stream: p.Stream, VAPIDPublicKey: p.VAPIDPublicKey,
		Notice: p.Notice, Error: p.Error, CustomStyles: p.CustomStyles,
		Frame: p.Frame, Chat: p.Chat, Reload: p.Reload, Invitation: p.Invitation,
	}
}

// Authorization and page data are read afresh. Only static surrounding bytes
// are retained; cursor and message parts are inserted independently per request.
func (s *Server) roomParts(p page, messages responsebody.Part) ([]responsebody.Part, error) {
	loadedAt := p.LoadedAt
	input := shellPage(p)
	raw, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	key := fmt.Sprintf("room-shell/%x", sha256.Sum256(raw))
	entry, ok := s.fragments.entry(key)
	if !ok {
		messageMarker := "\x00campfire-" + rand.Text() + "\x00"
		loadedMarker := "campfire-loaded-" + rand.Text()
		input.MessagesHTML, input.LoadedAt = template.HTML(messageMarker), loadedMarker
		b := borrowBuffer()
		defer releaseBuffer(b)
		if err := s.templates.ExecuteTemplate(b, "room", input); err != nil {
			return nil, err
		}
		rendered := b.String()
		// The room controller's refresh cursor precedes its message list. Check
		// that template contract at preparation time, not on every cache hit.
		beforeTime, remainder, found := strings.Cut(rendered, loadedMarker)
		if !found || strings.Count(rendered, loadedMarker) != 1 || strings.Count(rendered, messageMarker) != 1 {
			return nil, fmt.Errorf("room template must contain one cursor and one message insertion point")
		}
		beforeMessages, afterMessages, found := strings.Cut(remainder, messageMarker)
		if !found {
			return nil, fmt.Errorf("room template cursor must precede messages")
		}
		shell := &roomShellEntry{
			beforeTime:     responsebody.NewPart([]byte(beforeTime)),
			beforeMessages: responsebody.NewPart([]byte(beforeMessages)),
			afterMessages:  responsebody.NewPart([]byte(afterMessages)),
			bytes:          len(beforeTime) + len(beforeMessages) + len(afterMessages),
		}
		entry = s.fragments.putEntry(fragmentEntry{key: key, shell: shell})
	}
	shell := entry.shell
	return []responsebody.Part{shell.beforeTime, responsebody.NewPart([]byte(loadedAt)), shell.beforeMessages, messages, shell.afterMessages}, nil
}
