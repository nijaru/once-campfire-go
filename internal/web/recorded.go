package web

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/basecamp/once-campfire-go/internal/database"
	"github.com/basecamp/once-campfire-go/internal/responsebody"
)

// The marker exists only during template execution. The actual response inserts
// the cached message list without copying it through template/fmt/page buffers.

func (s *Server) messageList(ctx context.Context, messages []database.Message) (responsebody.Part, error) {
	key := s.fragmentKey(ctx, messageListCacheKey(messageReferences(messages)))
	if entry, ok := s.fragments.entry(key); cacheFragments(ctx) && ok {
		return entry.part, nil
	}
	views, err := s.messageItems(ctx, messages)
	if err != nil {
		return responsebody.Part{}, err
	}
	fragments := make([]template.HTML, len(views))
	for i, view := range views {
		fragments[i] = view.Fragment
	}
	return s.recordMessageList(ctx, key, fragments), nil
}

func (s *Server) recordMessageList(ctx context.Context, key string, fragments []template.HTML) responsebody.Part {
	size := 0
	for _, fragment := range fragments {
		size += len(fragment)
	}
	body := make([]byte, size)
	offset := 0
	for _, fragment := range fragments {
		offset += copy(body[offset:], fragment)
	}
	// The Part owns unpooled bytes through eviction and outstanding responses.
	entry := fragmentEntry{key: key, part: responsebody.NewPart(body)}
	if cacheFragments(ctx) {
		entry = s.fragments.putEntry(entry)
	}
	return entry.part
}

func writeRecorded(w http.ResponseWriter, status int, rendered, marker string, part responsebody.Part) {
	before, after, found := strings.Cut(rendered, marker)
	if !found {
		http.Error(w, "Missing message insertion point", 500)
		return
	}
	parts := []responsebody.Part{responsebody.NewPart([]byte(before)), part, responsebody.NewPart([]byte(after))}
	writeParts(w, status, parts)
}

func writeParts(w http.ResponseWriter, status int, parts []responsebody.Part) {
	if w.Header().Get("ETag") == "" {
		digest := responsebody.Digest(parts)
		w.Header().Set("ETag", fmt.Sprintf("W/\"%x\"", digest[:16]))
	}
	if w.Header().Get("Cache-Control") == "" {
		w.Header().Set("Cache-Control", "max-age=0, private, must-revalidate")
	}
	w.WriteHeader(status)
	if sw, ok := w.(*sessionWriter); ok && sw.failed {
		return
	}
	target := w
	for {
		if buffered, ok := target.(*responseBuffer); ok {
			buffered.parts = parts
			return
		}
		if wrapper, ok := target.(interface{ Unwrap() http.ResponseWriter }); ok {
			target = wrapper.Unwrap()
		} else {
			break
		}
	}
	for _, part := range parts {
		part.WriteTo(w)
	}
}
