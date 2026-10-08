package web

import (
	"container/list"
	"context"
	"crypto/rand"
	"encoding/binary"
	"html/template"
	"strconv"
	"strings"
	"sync"

	"github.com/basecamp/once-campfire-go/internal/database"
	"github.com/basecamp/once-campfire-go/internal/presentation"
	"github.com/basecamp/once-campfire-go/internal/responsebody"
)

type fragmentEntry struct {
	key   string
	html  template.HTML
	bytes int
	part  responsebody.Part
	shell *templateShell
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
	key := entry.key
	size := len(key) + len(entry.html) + entry.part.Len() + 240
	if entry.shell != nil {
		// Charge owned payload once, plus the slice and Part headers/digests.
		size += entry.shell.bytes + 32 + 56*len(entry.shell.parts)
	}
	if size > c.limit/4 {
		return entry
	}
	entry.bytes = size
	// Callers prepare owned payloads before admission. Oversized/disabled-cache
	// responses remain usable, and hits never wait for copying or hashing.
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
func messageVersion(message database.MessageReference) [messageVersionSize]byte {
	var version [messageVersionSize]byte
	binary.LittleEndian.PutUint64(version[:8], uint64(message.ID))
	binary.LittleEndian.PutUint64(version[8:16], uint64(message.UpdatedAt.Unix()))
	binary.LittleEndian.PutUint32(version[16:], uint32(message.UpdatedAt.Nanosecond()/1000))
	return version
}

func messageCacheKey(message database.MessageReference) string {
	version := messageVersion(message)
	return "message/" + string(version[:])
}

func messageListCacheKey(messages []database.MessageReference) string {
	var key strings.Builder
	key.Grow(len("message-list/") + messageVersionSize*len(messages))
	key.WriteString("message-list/")
	for _, message := range messages {
		version := messageVersion(message)
		key.Write(version[:])
	}
	return key.String()
}

// Namespace timestamp fragments by the generation observed before request reads.
// An older in-flight render cannot populate a newer generation after a commit.
func (s *Server) fragmentKey(ctx context.Context, key string) string {
	version := uint64(0)
	var host, origin string
	if info := requestMetadata(ctx); info != nil {
		version = info.databaseVersion
		host, origin = info.host, info.origin
	} else if _, ok := ctx.Value(fragmentObservationKey{}).(fragmentObservation); !ok {
		version, _ = s.DB.ResponseVersion(ctx)
	}
	if observation, ok := ctx.Value(fragmentObservationKey{}).(fragmentObservation); ok {
		version = observation.generation
	}
	if version == 0 {
		return "uncached/" + rand.Text() + "/" + key
	}
	return strconv.FormatUint(version, 10) + "/" + strconv.Quote(host) + "/" + strconv.Quote(origin) + "/" + key
}

func cacheFragments(ctx context.Context) bool {
	if observation, ok := ctx.Value(fragmentObservationKey{}).(fragmentObservation); ok {
		return observation.generation != 0
	}
	info := requestMetadata(ctx)
	return info == nil || info.databaseVersion != 0
}

func messageReferences(records []database.Message) []database.MessageReference {
	refs := make([]database.MessageReference, len(records))
	for i, record := range records {
		refs[i] = record.Reference()
	}
	return refs
}

func (s *Server) messageItems(ctx context.Context, messages []database.Message) ([]presentation.MessageView, error) {
	views := presentation.ViewMessages(messages)
	for i, m := range messages {
		if html, ok := s.fragments.get(s.fragmentKey(ctx, messageCacheKey(m.Reference()))); cacheFragments(ctx) && ok {
			views[i].Fragment = html
		}
	}
	if err := s.hydrateMessageViews(ctx, messages, views); err != nil {
		return nil, err
	}
	return views, nil
}
