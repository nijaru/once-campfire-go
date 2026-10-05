package web

import (
	"container/list"
	"context"
	"encoding/binary"
	"html/template"
	"strings"
	"sync"

	"github.com/basecamp/once-campfire-go/internal/database"
	"github.com/basecamp/once-campfire-go/internal/responsebody"
)

type fragmentEntry struct {
	key   string
	html  template.HTML
	bytes int
	part  responsebody.Part
	shell *roomShellEntry
}
type fragmentCache struct {
	mu           sync.Mutex
	entries      map[string]*list.Element
	order        list.List
	bytes, limit int
}

func newFragmentCache(limit int) *fragmentCache {
	return &fragmentCache{entries: map[string]*list.Element{}, limit: limit}
}
func (c *fragmentCache) get(key string) (template.HTML, bool) {
	entry, ok := c.entry(key)
	return entry.html, ok
}
func (c *fragmentCache) entry(key string) (fragmentEntry, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.entries[key]
	if !ok {
		return fragmentEntry{}, false
	}
	c.order.MoveToFront(e)
	return e.Value.(fragmentEntry), true
}

// Like the reference's MemoryStore, retain the first rendered value for a version,
// charge 240 bytes per entry, reject oversized entries, and prune to 75% capacity.
func (c *fragmentCache) put(key string, html template.HTML) template.HTML {
	return c.putEntry(fragmentEntry{key: key, html: html}).html
}
func (c *fragmentCache) putEntry(entry fragmentEntry) fragmentEntry {
	key, html := entry.key, entry.html
	size := len(key) + len(html) + 240
	if entry.shell != nil {
		// Three Part headers/digests plus the shell allocation and size field.
		size += entry.shell.bytes + 192
	}
	if strings.HasPrefix(key, "message-list/") {
		size += len(html)
	}
	if size > c.limit/4 {
		return entry
	}
	if strings.HasPrefix(key, "message-list/") {
		entry.part = responsebody.NewPart([]byte(html))
	}
	entry.bytes = size
	// Hash/copy payloads before taking the LRU lock. Concurrent misses may do
	// duplicate preparation, but unrelated hits never wait for a full-body hash.
	c.mu.Lock()
	defer c.mu.Unlock()
	if e, ok := c.entries[key]; ok {
		c.order.MoveToFront(e)
		return e.Value.(fragmentEntry)
	}
	c.entries[key] = c.order.PushFront(entry)
	c.bytes += size
	if c.bytes > c.limit {
		for c.bytes > c.limit*3/4 {
			e := c.order.Back()
			entry := e.Value.(fragmentEntry)
			c.bytes -= entry.bytes
			delete(c.entries, entry.key)
			c.order.Remove(e)
		}
	}
	return entry
}

const messageVersionSize = 20

// Match Stamp's UTC, microsecond-truncated identity without formatting dates.
// Separate seconds and fractions avoid UnixNano/UnixMicro's narrower date range.
func messageVersion(message database.Message) [messageVersionSize]byte {
	var version [messageVersionSize]byte
	binary.LittleEndian.PutUint64(version[:8], uint64(message.ID))
	binary.LittleEndian.PutUint64(version[8:16], uint64(message.UpdatedAt.Unix()))
	binary.LittleEndian.PutUint32(version[16:], uint32(message.UpdatedAt.Nanosecond()/1000))
	return version
}

func messageCacheKey(message database.Message) string {
	version := messageVersion(message)
	return "message/" + string(version[:])
}

func messageListCacheKey(messages []database.Message) string {
	var key strings.Builder
	key.Grow(len("message-list/") + messageVersionSize*len(messages))
	key.WriteString("message-list/")
	for _, message := range messages {
		version := messageVersion(message)
		key.Write(version[:])
	}
	return key.String()
}
func (s *Server) messageItems(ctx context.Context, messages []database.Message) ([]messageView, error) {
	views := viewMessages(messages)
	var missing []int64
	for i, m := range messages {
		if html, ok := s.fragments.get(messageCacheKey(m)); ok {
			views[i].Fragment = html
		} else if m.CreatorID == 0 {
			missing = append(missing, m.ID)
		}
	}
	hydrated := make(map[int64]database.Message, len(missing))
	if len(missing) > 0 {
		loaded, err := s.DB.MessagesByID(ctx, missing)
		if err != nil {
			return nil, err
		}
		for _, message := range loaded {
			hydrated[message.ID] = message
		}
	}
	for i, m := range messages {
		if views[i].Fragment != "" {
			continue
		}
		if m.CreatorID == 0 {
			var found bool
			m, found = hydrated[m.ID]
			if !found {
				views[i].Fragment = unrenderableMessage
				continue
			}
		}

		rendered, err := s.messageViews(ctx, []database.Message{m})
		if err != nil {
			return nil, err
		}
		views[i] = rendered[0]
	}
	return views, nil
}

const unrenderableMessage template.HTML = `<div class="message message--formatted message--failed center"><div class="message__body"><div class="message__body-content txt-align-center">Failed to load message content</div></div></div>`
