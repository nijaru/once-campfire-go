package web

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/basecamp/once-campfire-go/internal/presentation"
	"github.com/basecamp/once-campfire-go/internal/responsebody"
)

// The marker exists only during template execution. The actual response inserts
// the cached message list without copying it through template/fmt/page buffers.

func (s *Server) recordMessageList(ctx context.Context, key string, fragments []template.HTML) responsebody.Part {
	entry := fragmentEntry{key: key, part: presentation.FragmentList(fragments)}
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
