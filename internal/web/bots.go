package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/basecamp/once-campfire-go/internal/database"
	"github.com/basecamp/once-campfire-go/internal/presentation"
	"github.com/basecamp/once-campfire-go/internal/rails"
	"github.com/basecamp/once-campfire-go/internal/storage"
)

var botRoute = regexp.MustCompile(`^/rooms/([^/]+)/([^/]+)/messages(?:/([^/]+))?(?:/boosts(?:/([^/]+))?)?$`)

func (s *Server) botRequest(w http.ResponseWriter, r *http.Request) bool {
	values := botRoute.FindStringSubmatch(r.URL.Path)
	if values == nil {
		return false
	}
	// The normal nested message route comes first in config/routes.rb.
	if values[2] == "messages" {
		return false
	}
	switch r.Method {
	case "GET", "HEAD", "POST", "PATCH", "PUT", "DELETE":
	default:
		return false
	}
	r.SetPathValue("id", values[1])
	r.SetPathValue("message", values[3])
	r.SetPathValue("boost", values[4])
	var user database.User
	var err error
	fromCookie := false
	if cookie, e := r.Cookie("session_token"); e == nil {
		var token string
		if s.Secrets.VerifyCookie("session_token", rails.UnescapeCookie(cookie.Value), s.DB.Now(), &token) == nil {
			user, err = s.DB.SessionUser(r.Context(), token)
			fromCookie = err == nil
		}
	}
	if !fromCookie {
		user, err = s.DB.Bot(r.Context(), values[2])
		if err != nil {
			s.requestAuthentication(w, r)
			return true
		}
	} else if r.Method != "GET" && r.Method != "HEAD" && !s.browserWriteAllowed(r) {
		http.Error(w, "Invalid request origin", 422)
		return true
	}
	if s.blockBrowser(w, r) {
		return true
	}
	room, err := s.DB.Room(r.Context(), user.ID, roomID(r))
	if err != nil {
		http.NotFound(w, r)
		return true
	}
	var raw []byte
	if boundary := multipartBoundary(r); boundary != "" {
		r.Body = http.MaxBytesReader(w, r.Body, maxMultipartBody)
		cleanup, e := parseMultipart(r, boundary)
		defer cleanup()
		if e != nil {
			var limit *http.MaxBytesError
			if errors.As(e, &limit) {
				http.Error(w, "Request too large", 413)
				return true
			}
			http.Error(w, "Invalid upload", 400)
			return true
		}
	} else {
		r.Body = http.MaxBytesReader(w, r.Body, MaxBody)
		raw, err = io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Request too large", 413)
			return true
		}
		r.Body = io.NopCloser(bytes.NewReader(raw))
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid form", 400)
			return true
		}
	}
	if strings.Contains(r.URL.Path, "/boosts") {
		s.botBoost(w, r, user, room, string(raw))
		return true
	}
	switch r.Method {
	case "GET", "HEAD":
		before, _ := strconv.ParseInt(r.URL.Query().Get("before"), 10, 64)
		direction := "before"
		if before == 0 {
			before, _ = strconv.ParseInt(r.URL.Query().Get("after"), 10, 64)
			direction = "after"
		}
		linkAfter := r.URL.Query().Get("after") != ""
		result, err := s.MessageQueries.APIPage(r.Context(), s.presentationFacts(r.Context()), user.ID, room.ID, before, direction, linkAfter)
		if err != nil {
			s.fail(w, err)
			return true
		}
		w.Header().Set("X-Total-Count", strconv.Itoa(result.Count))
		if result.Next != 0 {
			param := "before"
			if linkAfter {
				param = "after"
			}
			w.Header().Set("Link", fmt.Sprintf(`<%s/rooms/%d/%s/messages?%s=%d>; rel="next"`, s.origin(r), room.ID, values[2], param, result.Next))
		}
		writeJSON(w, 200, result.Messages)
	case "POST":
		var staged *storage.Staged
		var body *string
		if r.MultipartForm != nil && len(r.MultipartForm.File["attachment"]) > 0 {
			staged, err = s.stageAttachment(r, "attachment")
			if err != nil {
				s.fail(w, err)
				return true
			}
		} else if r.Form.Get("attachment") != "" {
			http.Error(w, "Invalid attachment", 500)
			return true
		} else if strings.TrimSpace(string(raw)) == "" {
			w.WriteHeader(422)
			return true
		} else {
			value := strings.ToValidUTF8(string(raw), "�")
			body = &value
		}
		result, err := s.MessageCommands.Create(r.Context(), user.ID, room.ID, "", body, staged)
		if err != nil {
			s.fail(w, err)
			return true
		}
		if _, err = s.MessageEffects.Created(r.Context(), s.presentationFacts(r.Context()), result); err != nil {
			s.fail(w, err)
			return true
		}
		message := result.Commit.Message
		w.Header().Set("Location", fmt.Sprintf("%s/messages/%d", s.origin(r), message.ID))
		w.WriteHeader(201)
	case "PATCH", "PUT":
		message, err := s.findMessage(r, user, true)
		if err != nil {
			s.fail(w, err)
			return true
		}
		body := strings.ToValidUTF8(string(raw), "�")
		var rawBody *string = &body
		if r.Form.Has("attachment") || r.MultipartForm != nil && len(r.MultipartForm.File["attachment"]) > 0 {
			rawBody = nil
		}
		result, err := s.updateMessageAttributes(r, user, message, "", "attachment", rawBody)
		if err != nil {
			s.fail(w, err)
			return true
		}
		message = result.Commit.Message
		if err = s.MessagePublications.Updated(r.Context(), s.presentationFacts(r.Context()), result); err != nil {
			s.fail(w, err)
			return true
		}
		value, err := s.MessageQueries.APIRecord(r.Context(), s.presentationFacts(r.Context()), message)
		if err != nil {
			s.fail(w, err)
			return true
		}
		writeJSON(w, 200, value)
	case "DELETE":
		message, err := s.findMessage(r, user, true)
		if err != nil {
			s.fail(w, err)
			return true
		}
		result, err := s.MessageCommands.Delete(r.Context(), user.ID, message.ID)
		if err != nil {
			s.fail(w, err)
			return true
		}
		message = result.Commit.Message
		s.publish(message.RoomID, rails.TurboStream("remove", "message_"+message.ClientID, ""))
		if result.Processing != nil {
			s.fail(w, result.Processing)
			return true
		}
		w.WriteHeader(204)
	}
	return true
}
func (s *Server) botBoost(w http.ResponseWriter, r *http.Request, u database.User, room database.Room, body string) {
	message, err := s.findMessage(r, u, false)
	if err != nil {
		s.fail(w, err)
		return
	}
	if r.Method == "DELETE" {
		id := pathInt(r, "boost")
		if err = s.DB.DeleteBoost(r.Context(), u.ID, message.ID, id); err != nil {
			s.fail(w, err)
			return
		}
		s.publish(room.ID, rails.TurboStream("remove", fmt.Sprintf("boost_%d", id), ""))
		w.WriteHeader(204)
		return
	}
	if r.Method != "POST" {
		http.NotFound(w, r)
		return
	}
	if strings.TrimSpace(body) == "" {
		w.WriteHeader(422)
		return
	}
	boost, err := s.DB.CreateBoost(r.Context(), u.ID, message.ID, body)
	if err != nil {
		s.fail(w, err)
		return
	}
	markup, err := s.Presentation.Markup("boost", boost)
	if err != nil {
		s.fail(w, err)
		return
	}
	s.publish(room.ID, rails.TurboStream("append", "boosts_message_"+message.ClientID, markup))
	writeJSON(w, 201, struct {
		ID        int64                `json:"id"`
		Content   string               `json:"content"`
		CreatedAt string               `json:"created_at"`
		Booster   presentation.APIUser `json:"booster"`
		Message   struct {
			ID  int64  `json:"id"`
			URL string `json:"url"`
		} `json:"message"`
	}{boost.ID, boost.Content, jsonTime(boost.CreatedAt), s.Presentation.APIUser(s.presentationFacts(r.Context()), database.APIAuthor{UserDisplay: database.UserDisplay{ID: u.ID, Name: u.Name, UpdatedAt: u.UpdatedAt}, Role: u.Role}), struct {
		ID  int64  `json:"id"`
		URL string `json:"url"`
	}{message.ID, fmt.Sprintf("%s/rooms/%d/messages/%d", s.origin(r), room.ID, message.ID)}})
}
func jsonTime(t time.Time) string { return t.UTC().Format("2006-01-02T15:04:05.000Z") }
func writeJSON(w http.ResponseWriter, status int, value any) {
	data, err := json.Marshal(value)
	if err != nil {
		http.Error(w, "Internal server error", 500)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	w.Write(data)
}
