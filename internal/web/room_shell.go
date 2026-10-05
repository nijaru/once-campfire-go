package web

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"html/template"
	"strings"

	"github.com/basecamp/once-campfire-go/internal/responsebody"
)

type roomShellEntry struct {
	beforeTime, beforeMessages, afterMessages responsebody.Part
	bytes                                     int
}

// Authorization and page data are read afresh. Only static surrounding bytes
// are retained; cursor and message parts are inserted independently per request.
func (s *Server) roomParts(p page, messages responsebody.Part) ([]responsebody.Part, error) {
	loadedAt := p.LoadedAt
	p.Messages, p.MessagesHTML, p.LoadedAt = nil, "", ""
	// Preserve all page dependencies, including future template inputs.
	raw, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	key := fmt.Sprintf("room-shell/%x", sha256.Sum256(raw))
	entry, ok := s.fragments.entry(key)
	if !ok {
		messageMarker := "\x00campfire-" + rand.Text() + "\x00"
		loadedMarker := "campfire-loaded-" + rand.Text()
		p.MessagesHTML, p.LoadedAt = template.HTML(messageMarker), loadedMarker
		b := borrowBuffer()
		defer releaseBuffer(b)
		if err := s.templates.ExecuteTemplate(b, "room", p); err != nil {
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
