package web

import (
	"context"
	"database/sql"
	"errors"
	"html/template"

	"github.com/basecamp/once-campfire-go/internal/database"
	"github.com/basecamp/once-campfire-go/internal/responsebody"
)

// A query owns its cache observation separately from live request metadata. A
// changed generation disables reuse/admission; it never relabels older records.
type fragmentObservationKey struct{}
type fragmentObservation struct{ generation uint64 }

func (s *Server) messageQueryContext(ctx context.Context) context.Context {
	var version uint64
	if info := requestMetadata(ctx); info != nil {
		version = info.databaseVersion
	} else {
		version, _ = s.DB.ResponseVersion(ctx)
	}
	return context.WithValue(ctx, fragmentObservationKey{}, fragmentObservation{version})
}

func (s *Server) checkMessageObservation(ctx context.Context) context.Context {
	observation := ctx.Value(fragmentObservationKey{}).(fragmentObservation)
	if version, err := s.DB.ResponseVersion(ctx); err != nil || version != observation.generation {
		return context.WithValue(ctx, fragmentObservationKey{}, fragmentObservation{})
	}
	return ctx
}

func (s *Server) readMessagePage(ctx context.Context, user, room, anchor int64, direction string, fallback bool) (responsebody.Part, int, error) {
	ctx = s.messageQueryContext(ctx)
	read, err := s.DB.BeginMessageRead(ctx)
	if err != nil {
		return responsebody.Part{}, 0, err
	}
	defer read.Close()
	refs, err := read.PageReferences(ctx, user, room, anchor, direction)
	if fallback && errors.Is(err, sql.ErrNoRows) {
		refs, err = read.PageReferences(ctx, user, room, 0, direction)
	}
	if err != nil {
		return responsebody.Part{}, 0, err
	}
	part, err := s.readMessageList(ctx, read, refs)
	return part, len(refs), err
}

func (s *Server) readSearchMessages(ctx context.Context, user int64, query string) (responsebody.Part, int, error) {
	ctx = s.messageQueryContext(ctx)
	read, err := s.DB.BeginMessageRead(ctx)
	if err != nil {
		return responsebody.Part{}, 0, err
	}
	defer read.Close()
	refs, err := read.SearchReferences(ctx, user, query)
	if err != nil {
		return responsebody.Part{}, 0, err
	}
	part, err := s.readMessageList(ctx, read, refs)
	return part, len(refs), err
}

// Select hits first, materialize only body/association misses in the selection
// snapshot, then release it before rich-text display, signing or template work.
func (s *Server) readMessageList(ctx context.Context, read *database.MessageRead, refs []database.MessageReference) (responsebody.Part, error) {
	ctx = s.checkMessageObservation(ctx)
	key := s.fragmentKey(ctx, messageListCacheKey(refs))
	if cacheFragments(ctx) {
		if entry, ok := s.fragments.entry(key); ok {
			// Retain the actual Part, not its key: eviction cannot trigger new hydration.
			if err := read.Finish(); err != nil {
				return responsebody.Part{}, err
			}
			return entry.part, nil
		}
	}
	fragments := make([]template.HTML, len(refs))
	var missing []database.MessageReference
	var positions []int
	for i, ref := range refs {
		if cacheFragments(ctx) {
			if html, ok := s.fragments.get(s.fragmentKey(ctx, messageCacheKey(ref))); ok {
				fragments[i] = html
				continue
			}
		}
		missing = append(missing, ref)
		positions = append(positions, i)
	}
	records, err := read.Records(ctx, missing)
	if err != nil {
		return responsebody.Part{}, err
	}
	views := viewMessages(records)
	prepared := prepareMessageViews(views)
	data, users, err := read.Displays(ctx, prepared.records, prepared.mentioned)
	if err != nil {
		return responsebody.Part{}, err
	}
	if err = read.Finish(); err != nil {
		return responsebody.Part{}, err
	}
	ctx = s.checkMessageObservation(ctx)
	if err = s.presentMessageViews(ctx, views, prepared, data, users); err != nil {
		return responsebody.Part{}, err
	}
	for i, view := range views {
		fragments[positions[i]] = view.Fragment
	}
	if len(refs) == 0 {
		return responsebody.Part{}, nil
	}
	// The admission key must use the same observation as the rendered misses.
	key = s.fragmentKey(ctx, messageListCacheKey(refs))
	return s.recordMessageList(ctx, key, fragments), nil
}
