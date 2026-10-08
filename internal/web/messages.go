package web

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/basecamp/once-campfire-go/internal/database"
	"github.com/basecamp/once-campfire-go/internal/presentation"
	"github.com/basecamp/once-campfire-go/internal/rails"
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

func (s *Server) presentationFacts(ctx context.Context) presentation.Facts {
	facts := presentation.Facts{Now: s.DB.Now()}
	if info := requestMetadata(ctx); info != nil {
		facts.Host, facts.Origin = info.host, info.origin
	}
	return facts
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
	if respondFormat(w, r, "html") == "" {
		return
	}
	views, err := s.MessageQueries.Views(r.Context(), s.presentationFacts(r.Context()), []database.Message{m})
	if err != nil {
		s.fail(w, err)
		return
	}
	s.render(w, r, "show-message", 200, page{User: u, Messages: views})
}

func (s *Server) editMessage(w http.ResponseWriter, r *http.Request, u database.User) {
	m, err := s.findMessage(r, u, true)
	if err != nil {
		s.fail(w, err)
		return
	}
	if respondFormat(w, r, "html") == "" {
		return
	}
	view, err := s.MessageQueries.Edit(r.Context(), s.presentationFacts(r.Context()), m)
	if err != nil {
		s.fail(w, err)
		return
	}
	s.render(w, r, "edit-message", 200, page{User: u, Messages: []presentation.MessageView{view}})
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
	if respondFormat(w, r, "html") == "" {
		return
	}
	view, err := s.MessageQueries.Boosts(r.Context(), m)
	if err != nil {
		s.fail(w, err)
		return
	}
	s.render(w, r, "boosts-index", 200, page{User: u, Messages: []presentation.MessageView{view}})
}

func (s *Server) newBoost(w http.ResponseWriter, r *http.Request, u database.User) {
	m, err := s.findMessage(r, u, false)
	if err != nil {
		s.fail(w, err)
		return
	}
	s.render(w, r, "new-boost", 200, page{User: u, Messages: []presentation.MessageView{presentation.ViewMessage(m)}})
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
	markup, err := s.Presentation.Markup("boost", boost)
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
		views, err := s.MessageQueries.Views(r.Context(), s.presentationFacts(r.Context()), group.messages)
		if err != nil {
			s.fail(w, err)
			return
		}
		for _, m := range views {
			markup, err := s.Presentation.Markup("message", m)
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
