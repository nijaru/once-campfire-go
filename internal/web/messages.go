package web

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/basecamp/once-campfire-go/internal/rails"

	"github.com/basecamp/once-campfire-go/internal/database"
	"github.com/basecamp/once-campfire-go/internal/presentation"
	"github.com/basecamp/once-campfire-go/internal/richtext"
	"github.com/basecamp/once-campfire-go/internal/storage"
)

func (s *Server) registerMessageRoutes() {
	s.mux.HandleFunc("GET /messages", s.auth(s.messages))
	s.mux.HandleFunc("POST /messages", s.auth(s.createMessage))
	s.mux.HandleFunc("GET /messages/{message}", s.auth(s.showMessage))
	s.mux.HandleFunc("GET /messages/{message}/edit", s.auth(s.editMessage))
	s.mux.HandleFunc("PATCH /messages/{message}", s.auth(s.updateMessage))
	s.mux.HandleFunc("PUT /messages/{message}", s.auth(s.updateMessage))
	s.mux.HandleFunc("DELETE /messages/{message}", s.auth(s.deleteMessage))
	s.mux.HandleFunc("GET /rooms/{id}/messages/{message}", s.auth(s.showMessage))
	s.mux.HandleFunc("GET /rooms/{id}/messages/{message}/edit", s.auth(s.editMessage))
	s.mux.HandleFunc("PATCH /rooms/{id}/messages/{message}", s.auth(s.updateMessage))
	s.mux.HandleFunc("PUT /rooms/{id}/messages/{message}", s.auth(s.updateMessage))
	s.mux.HandleFunc("DELETE /rooms/{id}/messages/{message}", s.auth(s.deleteMessage))
	s.mux.HandleFunc("GET /messages/{message}/boosts", s.auth(s.boosts))
	s.mux.HandleFunc("GET /messages/{message}/boosts/new", s.auth(s.newBoost))
	s.mux.HandleFunc("POST /messages/{message}/boosts", s.auth(s.createBoost))
	s.mux.HandleFunc("DELETE /messages/{message}/boosts/{boost}", s.auth(s.deleteBoost))
	s.mux.HandleFunc("GET /rooms/{id}/refresh", s.auth(s.refreshRoom))
}

func pathInt(r *http.Request, key string) int64 {
	id, _ := strconv.ParseInt(r.PathValue(key), 10, 64)
	return id
}

func (s *Server) findMessage(
	r *http.Request,
	u database.User,
	administer bool,
) (database.Message, error) {
	if route, _, _ := recognizeRequest(r); route != nil &&
		strings.HasPrefix(route.Endpoint, "messages#") &&
		roomID(r) == 0 {
		return database.Message{}, sql.ErrNoRows
	}
	m, err := s.DB.ReachableMessage(r.Context(), u.ID, pathInt(r, "message"))
	if err != nil {
		return m, err
	}
	if (r.PathValue("room_id") != "" || r.Form.Has("room_id")) && m.RoomID != roomID(r) {
		return m, sql.ErrNoRows
	}
	if administer && u.Role != 1 && u.ID != m.CreatorID {
		return m, database.ErrForbidden
	}
	return m, nil
}

func (s *Server) messageViews(
	ctx context.Context,
	messages []database.Message,
) ([]messageView, error) {
	views := viewMessages(messages)
	if err := s.hydrateMessageViews(ctx, views); err != nil {
		return nil, err
	}
	return views, nil
}

// Single-message forms consume different data from a displayed message. Keep
// their reads fresh without rendering and retaining an unused message fragment.
func (s *Server) messagePageViews(
	ctx context.Context,
	name string,
	records []database.Message,
) ([]messageView, error) {
	if name != "edit-message" && name != "boosts-index" && name != "new-boost" {
		return s.messageViews(ctx, records)
	}
	views := viewMessages(records)
	var creators map[int64]database.UserDisplay
	if name != "new-boost" {
		ids := make([]int64, len(records))
		for i, record := range records {
			ids[i] = record.CreatorID
		}
		var err error
		creators, err = s.DB.UserDisplays(ctx, ids)
		if err != nil {
			return nil, err
		}
	}
	for i := range views {
		if name == "new-boost" {
			continue
		}
		// Missing creators suppress attachment/boost presentation in the
		// reference. This observation is required even by these narrow views.
		_, exists := creators[views[i].CreatorID]
		missingCreator := !exists
		switch name {
		case "edit-message":
			if !missingCreator {
				if err := s.messageAttachment(ctx, &views[i]); err != nil {
					return nil, err
				}
			}
			if views[i].Attachment == nil {
				doc := richtext.Prepare(views[i].Body)
				resolved, err := s.resolveDocument(ctx, doc.EditorAttachables())
				if err != nil {
					return nil, err
				}
				views[i].Editable, _ = doc.Editable(resolved)
			}
		case "boosts-index":
			if missingCreator {
				continue
			}
			var err error
			views[i].Boosts, err = s.DB.Boosts(ctx, views[i].ID)
			if err != nil {
				return nil, err
			}
		}
	}
	return views, nil
}

// Prepare only fragment misses. Associations and mention targets are materialized
// together; presentation and signing run after the read snapshot is released.
type messagePreparation struct {
	records   []database.Message
	documents map[int]richtext.Document
	targets   map[string]presentation.MentionTarget
	mentioned []int64
}

func prepareMessageViews(views []messageView) messagePreparation {
	var records []database.Message
	documents := make(map[int]richtext.Document)
	targets := make(map[string]presentation.MentionTarget)
	var mentioned []int64
	for i := range views {
		if views[i].Fragment != "" {
			continue
		}
		records = append(records, views[i].Message)
		doc := richtext.Prepare(views[i].Body)
		documents[i] = doc
		for _, token := range doc.Attachables() {
			if _, ok := targets[token]; ok {
				continue
			}
			target := presentation.MentionTargetFor(token)
			targets[token] = target
			if target.ID != 0 {
				mentioned = append(mentioned, target.ID)
			}
		}
	}
	return messagePreparation{records: records, documents: documents, targets: targets, mentioned: mentioned}
}

func (s *Server) hydrateMessageViews(ctx context.Context, views []messageView) error {
	prepared := prepareMessageViews(views)
	data, users, err := s.DB.MessageDisplays(ctx, prepared.records, prepared.mentioned)
	if err != nil {
		return err
	}
	return s.presentMessageViews(ctx, views, prepared, data, users)
}

func (s *Server) presentMessageViews(ctx context.Context, views []messageView, prepared messagePreparation, data map[int64]database.MessageDisplay, users map[int64]database.UserDisplay) error {
	rich := s.resolvedRichContext(ctx, prepared.targets, users)
	for i := range views {
		if views[i].Fragment != "" {
			continue
		}
		detail := data[views[i].ID]
		if detail.Author == nil {
			views[i].Fragment = unrenderableMessage
			continue
		}
		creator := detail.Author
		views[i].Creator = creator.Name
		views[i].CreatorTitle = creator.Title()
		views[i].CreatorUpdatedAt = creator.UpdatedAt
		views[i].Permalink = messagePermalink(ctx, views[i].RoomID, views[i].ID)
		views[i].RoomName = displayRoom(detail.Room, detail.Participants, database.User{}).Name
		result, err := prepared.documents[i].Display(rich)
		if err != nil {
			return err
		}
		views[i].HTML = template.HTML(result.Presentation)
		views[i].AllEmoji = allEmoji(result.Plain)
		if sound := soundHTML(result.Plain); sound != "" {
			views[i].HTML = template.HTML(sound)
		}
		views[i].Boosts = detail.Boosts
		if detail.Attachment != nil {
			if err := s.prepareMessageAttachment(&views[i], *detail.Attachment); err != nil {
				return err
			}
		}
		key := s.fragmentKey(ctx, messageCacheKey(views[i].Message.Reference()))
		if html, ok := s.fragments.get(key); cacheFragments(ctx) && ok {
			views[i].Fragment = html
		} else {
			body, err := s.messageMarkup(views[i])
			if err != nil {
				return err
			}
			views[i].Fragment = template.HTML(body)
			if cacheFragments(ctx) {
				views[i].Fragment = s.fragments.put(key, views[i].Fragment)
			}
		}
	}
	return nil
}

func (s *Server) messageAttachment(ctx context.Context, view *messageView) error {
	blob, err := s.DB.AttachedBlob(ctx, "Message", view.ID, "attachment")
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	return s.prepareMessageAttachment(view, blob)
}

func (s *Server) prepareMessageAttachment(view *messageView, blob database.Blob) error {
	view.Attachment = &blob
	view.BlobURL = s.Storage.BlobURL(blob)
	view.DownloadURL = view.BlobURL + "?disposition=attachment"
	view.Image = storage.Variable(blob.Type())
	if view.Image || storage.Previewable(blob.Type()) {
		variation := storage.Resize(1200, 800, "")
		if storage.Previewable(blob.Type()) {
			variation = storage.Variation{
				{Key: "format", Value: storage.Symbol("webp")},
				{Key: "resize_to_limit", Value: []any{int64(1200), int64(800)}},
			}
		}
		var err error
		view.PreviewURL, err = s.Storage.RepresentationURL(blob, variation)
		if err != nil {
			return err
		}
	}
	view.HTML = template.HTML(attachmentHTML(blob, view.BlobURL, view.DownloadURL, view.PreviewURL))
	return nil
}

func (s *Server) markup(name string, data any) (string, error) {
	var b bytes.Buffer
	err := s.templates.ExecuteTemplate(&b, name, data)
	return b.String(), err
}

func messagePermalink(ctx context.Context, room, message int64) string {
	var origin string
	if info := requestMetadata(ctx); info != nil {
		origin = info.origin
	}
	if origin == "" {
		origin = "http://example.org"
	}
	return fmt.Sprintf("%s/rooms/%d/@%d", origin, room, message)
}

func (s *Server) publish(room int64, markup string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s.Cable.Publish(ctx, room, markup)
}

func writeStream(w http.ResponseWriter, markup string) {
	w.Header().Set("Content-Type", "text/vnd.turbo-stream.html; charset=utf-8")
	fmt.Fprint(w, markup)
}

func (s *Server) showMessage(w http.ResponseWriter, r *http.Request, u database.User) {
	m, err := s.findMessage(r, u, false)
	if err != nil {
		s.fail(w, err)
		return
	}
	s.render(w, r, "show-message", 200, page{User: u, messageRecords: []database.Message{m}})
}

func (s *Server) editMessage(w http.ResponseWriter, r *http.Request, u database.User) {
	m, err := s.findMessage(r, u, true)
	if err != nil {
		s.fail(w, err)
		return
	}
	s.render(w, r, "edit-message", 200, page{User: u, messageRecords: []database.Message{m}})
}

func (s *Server) updateMessage(w http.ResponseWriter, r *http.Request, u database.User) {
	m, err := s.findMessage(r, u, true)
	if err != nil {
		s.fail(w, err)
		return
	}
	if !requireMessage(w, r) {
		return
	}
	result, err := s.updateMessageAttributes(r, u, m, "message[body]", "message[attachment]", nil)
	if err != nil {
		s.fail(w, err)
		return
	}
	m = result.Commit.Message
	if err = s.updatedMessageEffects(r.Context(), result); err != nil {
		s.fail(w, err)
		return
	}
	format := respondFormat(w, r, "html", "json")
	if format == "" {
		return
	}
	if format == "json" {
		http.Error(w, "Missing template messages/show", 500)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/rooms/%d/messages/%d", m.RoomID, m.ID), 302)
}

func (s *Server) deleteMessage(w http.ResponseWriter, r *http.Request, u database.User) {
	m, err := s.findMessage(r, u, true)
	if err != nil {
		s.fail(w, err)
		return
	}
	result, err := s.MessageCommands.Delete(r.Context(), u.ID, m.ID)
	if err != nil {
		s.fail(w, err)
		return
	}
	m = result.Commit.Message
	markup := rails.TurboStream("remove", "message_"+m.ClientID, "")
	s.publish(m.RoomID, markup)
	if result.Processing != nil {
		s.fail(w, result.Processing)
		return
	}
	if respondFormat(w, r, "turbo_stream") != "" {
		writeStream(w, markup)
	}
}

func (s *Server) boosts(w http.ResponseWriter, r *http.Request, u database.User) {
	m, err := s.findMessage(r, u, false)
	if err != nil {
		s.fail(w, err)
		return
	}
	s.render(w, r, "boosts-index", 200, page{User: u, messageRecords: []database.Message{m}})
}

func (s *Server) newBoost(w http.ResponseWriter, r *http.Request, u database.User) {
	m, err := s.findMessage(r, u, false)
	if err != nil {
		s.fail(w, err)
		return
	}
	s.render(w, r, "new-boost", 200, page{User: u, messageRecords: []database.Message{m}})
}

func (s *Server) createBoost(w http.ResponseWriter, r *http.Request, u database.User) {
	m, err := s.findMessage(r, u, false)
	if err != nil {
		s.fail(w, err)
		return
	}
	boost, err := s.DB.CreateBoost(r.Context(), u.ID, m.ID, r.Form.Get("boost[content]"))
	if err != nil {
		s.fail(w, err)
		return
	}
	markup, err := s.markup("boost", boost)
	if err != nil {
		s.fail(w, err)
		return
	}
	s.publish(m.RoomID, rails.TurboStream("append", "boosts_message_"+m.ClientID, markup))
	http.Redirect(w, r, fmt.Sprintf("/messages/%d/boosts", m.ID), 302)
}

func (s *Server) deleteBoost(w http.ResponseWriter, r *http.Request, u database.User) {
	m, err := s.findMessage(r, u, false)
	if err != nil {
		s.fail(w, err)
		return
	}
	id := pathInt(r, "boost")
	if err = s.DB.DeleteBoost(r.Context(), u.ID, m.ID, id); err != nil {
		s.fail(w, err)
		return
	}
	markup := rails.TurboStream("remove", fmt.Sprintf("boost_%d", id), "")
	s.publish(m.RoomID, markup)
	w.WriteHeader(204)
}

func (s *Server) refreshRoom(w http.ResponseWriter, r *http.Request, u database.User) {
	room, err := s.DB.Room(r.Context(), u.ID, roomID(r))
	if err != nil {
		s.fail(w, err)
		return
	}
	seconds, err := strconv.ParseFloat(r.URL.Query().Get("since"), 64)
	since := time.UnixMicro(int64(seconds * 1000))
	if err != nil {
		http.Error(w, "Invalid timestamp", 400)
		return
	}
	created, updated, err := s.DB.RefreshedMessages(r.Context(), room.ID, since)
	if err != nil {
		s.fail(w, err)
		return
	}
	var result strings.Builder
	for _, group := range []struct {
		action   string
		messages []database.Message
	}{{"append", created}, {"replace", updated}} {
		views, err := s.messageViews(r.Context(), group.messages)
		if err != nil {
			s.fail(w, err)
			return
		}
		for _, m := range views {
			markup, err := s.markup("message", m)
			if err != nil {
				s.fail(w, err)
				return
			}
			target := room.DOM("messages")
			if group.action == "replace" {
				target = "message_" + m.ClientID
			}
			result.WriteString(rails.TurboStream(group.action, target, markup))
		}
	}
	writeStream(w, result.String())
}
