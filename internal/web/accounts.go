package web

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/basecamp/once-campfire-go/internal/application"
	"github.com/basecamp/once-campfire-go/internal/database"
	"github.com/basecamp/once-campfire-go/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

func (s *Server) registerAccountRoutes() {
	s.mux.HandleFunc("GET /account/edit", s.auth(s.accountForm))
	s.mux.HandleFunc("PATCH /account", s.auth(s.updateAccount))
	s.mux.HandleFunc("PUT /account", s.auth(s.updateAccount))
	s.mux.HandleFunc("POST /account/join_code", s.auth(s.resetJoinCode))
	s.mux.HandleFunc("GET /account/custom_styles/edit", s.auth(s.customStyles))
	s.mux.HandleFunc("PATCH /account/custom_styles", s.auth(s.customStyles))
	s.mux.HandleFunc("PUT /account/custom_styles", s.auth(s.customStyles))
	s.mux.HandleFunc("GET /users/{user}/profile", s.auth(s.profile))
	s.mux.HandleFunc("PATCH /users/{user}/profile", s.auth(s.profile))
	s.mux.HandleFunc("PUT /users/{user}/profile", s.auth(s.profile))
	s.mux.HandleFunc("GET /users/{user}", s.auth(s.showUser))
	s.mux.HandleFunc("POST /users/{user}/ban", s.auth(s.banUser))
	s.mux.HandleFunc("DELETE /users/{user}/ban", s.auth(s.banUser))
	s.mux.HandleFunc("GET /account/users", s.auth(s.accountUsers))
	s.mux.HandleFunc("PATCH /account/users/{user}", s.auth(s.manageUser))
	s.mux.HandleFunc("PUT /account/users/{user}", s.auth(s.manageUser))
	s.mux.HandleFunc("DELETE /account/users/{user}", s.auth(s.manageUser))
	s.mux.HandleFunc("GET /join/{code}", s.browserCheck(s.join))
	s.mux.HandleFunc("POST /join/{code}", s.browserCheck(s.join))
	s.mux.HandleFunc("GET /account/bots", s.auth(s.bots))
	s.mux.HandleFunc("GET /account/bots/new", s.auth(s.botForm))
	s.mux.HandleFunc("GET /account/bots/{bot}/edit", s.auth(s.botForm))
	s.mux.HandleFunc("POST /account/bots", s.auth(s.saveBot))
	s.mux.HandleFunc("PATCH /account/bots/{bot}", s.auth(s.saveBot))
	s.mux.HandleFunc("PUT /account/bots/{bot}", s.auth(s.saveBot))
	s.mux.HandleFunc("DELETE /account/bots/{bot}", s.auth(s.saveBot))
	s.mux.HandleFunc("PATCH /account/bots/{bot}/key", s.auth(s.rotateBot))
	s.mux.HandleFunc("PUT /account/bots/{bot}/key", s.auth(s.rotateBot))
	s.mux.HandleFunc("GET /session/transfers/{token}", s.browserCheck(s.transfer))
	s.mux.HandleFunc("PATCH /session/transfers/{token}", s.browserCheck(s.transfer))
	s.mux.HandleFunc("PUT /session/transfers/{token}", s.browserCheck(s.transfer))
}

// Account/profile effects are asynchronous in the reference. Preserve the
// committed redirect response while reporting processing/admission failures.
func logAccountProcessing(err error) {
	if err != nil {
		slog.Error("account post-commit processing failed", "error", err)
	}
}

func administrator(w http.ResponseWriter, u database.User) bool {
	if u.Role != 1 {
		http.Error(w, "Forbidden", 403)
		return false
	}
	return true
}

func accountPage(raw string, count int) (int64, int64) {
	var number int64
	fmt.Sscan(raw, &number)
	number = max(1, min(number, 1_000_000_000))
	last := int64(max(1, (count+499)/500))
	next := number + 1
	if number == last {
		next = 0
	}
	return number, next
}

func (s *Server) accountForm(w http.ResponseWriter, r *http.Request, u database.User) {
	users, err := s.DB.AccountUsers(r.Context(), u.Role == 1)
	if err != nil {
		s.fail(w, err)
		return
	}
	_, next := accountPage(r.Form.Get("page"), len(users))
	p := page{Title: "Account settings", User: u, NextPage: next}
	for _, user := range users {
		if user.Role == 1 {
			p.Administrators = append(p.Administrators, user)
		} else {
			p.Users = append(p.Users, user)
		}
	}
	s.respondPage(w, r, "account", 200, p)
}

func (s *Server) accountUsers(w http.ResponseWriter, r *http.Request, u database.User) {
	if respondFormat(w, r, "turbo_stream") == "" {
		return
	}
	users, err := s.DB.AccountUsers(r.Context(), false)
	if err != nil {
		s.fail(w, err)
		return
	}
	number, next := accountPage(r.Form.Get("page"), len(users))
	start := min(int((number-1)*500), len(users))
	body, err := s.Presentation.Markup(
		"account-users-stream",
		page{User: u, Users: users[start:min(start+500, len(users))], NextPage: next},
	)
	if err != nil {
		s.fail(w, err)
		return
	}
	writeStream(w, body)
}

func (s *Server) updateAccount(w http.ResponseWriter, r *http.Request, u database.User) {
	if !administrator(w, u) {
		return
	}
	var name *string
	if r.Form.Has("account[name]") {
		value := r.Form.Get("account[name]")
		name = &value
	}
	var restrict *sql.NullBool
	if r.Form.Has("account[settings][restrict_room_creation_to_administrators]") {
		value := r.Form.Get("account[settings][restrict_room_creation_to_administrators]")
		cast := sql.NullBool{Bool: true, Valid: true}
		// ActiveModel boolean casting is case-sensitive; blank remains JSON null.
		switch value {
		case "":
			cast = sql.NullBool{}
		case "0", "f", "F", "false", "FALSE", "off", "OFF":
			cast.Bool = false
		}
		restrict = &cast
	}
	upload, err := s.optionalUpload(r, "account[logo]")
	if err != nil {
		s.fail(w, err)
		return
	}
	logo, err := recordAttachment(r, "account[logo]", upload, false)
	if err != nil {
		s.fail(w, err)
		return
	}
	result, err := s.AccountCommands.UpdateAccount(r.Context(), u.ID, database.AccountInput{Name: name, Restrict: restrict}, logo)
	if err != nil {
		s.fail(w, err)
		return
	}
	logAccountProcessing(result.Processing)
	s.flash(r, "notice", "✓")
	http.Redirect(w, r, "/account/edit", 302)
}

func (s *Server) resetJoinCode(w http.ResponseWriter, r *http.Request, u database.User) {
	if !administrator(w, u) {
		return
	}
	if _, err := s.AccountCommands.UpdateAccount(r.Context(), u.ID, database.AccountInput{ResetJoin: true}, application.Attachment{}); err != nil {
		s.fail(w, err)
		return
	}
	http.Redirect(w, r, "/account/edit", 302)
}

func (s *Server) customStyles(w http.ResponseWriter, r *http.Request, u database.User) {
	if !administrator(w, u) {
		return
	}
	if r.Method == "GET" || r.Method == "HEAD" {
		s.respondPage(w, r, "custom-styles", 200, page{Title: "Custom styles", User: u})
		return
	}
	if r.Form.Has("account[custom_styles]") {
		value := r.Form.Get("account[custom_styles]")
		if _, err := s.AccountCommands.UpdateAccount(r.Context(), u.ID, database.AccountInput{Styles: &value}, application.Attachment{}); err != nil {
			s.fail(w, err)
			return
		}
	}
	s.flash(r, "notice", "✓")
	http.Redirect(w, r, "/account/custom_styles/edit", 302)
}

func (s *Server) profile(w http.ResponseWriter, r *http.Request, u database.User) {
	if r.Method == "GET" || r.Method == "HEAD" {
		data, err := s.PageQueries.Profile(r.Context(), u.Participant())
		if err != nil {
			s.fail(w, err)
			return
		}
		p := page{
			Title: "My settings", User: u, Transfer: s.origin(r) + s.transferPath(u),
			AvatarAttached: data.AvatarAttached, Memberships: data.Memberships,
			DirectMemberships: data.DirectMemberships,
		}
		s.respondPage(w, r, "profile", 200, p)
		return
	}
	attrs := database.UserChanges{}
	for _, field := range []struct {
		key    string
		target **string
	}{{"name", &attrs.Name}, {"email_address", &attrs.Email}, {"bio", &attrs.Bio}} {
		if r.Form.Has("user["+field.key+"]") && !nullParam(r, "user["+field.key+"]") {
			value := r.Form.Get("user[" + field.key + "]")
			*field.target = &value
		}
	}
	if password := r.Form.Get("user[password]"); password != "" {
		digest, err := bcrypt.GenerateFromPassword([]byte(password), 12)
		if err != nil {
			http.Error(w, "Invalid password", 422)
			return
		}
		value := string(digest)
		attrs.Password = &value
	}
	upload, err := s.optionalUpload(r, "user[avatar]")
	if err != nil {
		s.fail(w, err)
		return
	}
	avatar, err := recordAttachment(r, "user[avatar]", upload, true)
	if err != nil {
		s.fail(w, err)
		return
	}
	result, err := s.AccountCommands.UpdateUser(r.Context(), u.ID, u.ID, attrs, avatar)
	if err != nil {
		s.fail(w, err)
		return
	}
	logAccountProcessing(result.Processing)
	notice := "✓"
	if upload != nil || r.Form.Has("user[avatar]") && !nullParam(r, "user[avatar]") {
		notice = "It may take up to 30 minutes to change everywhere."
	}
	s.flash(r, "notice", notice)
	http.Redirect(w, r, "/users/me/profile", 302)
}

func (s *Server) showUser(w http.ResponseWriter, r *http.Request, u database.User) {
	subject, err := s.DB.User(r.Context(), pathInt(r, "user"))
	if err != nil {
		s.fail(w, err)
		return
	}
	s.respondPage(
		w,
		r,
		"user",
		200,
		page{
			Title:    subject.Name,
			User:     u,
			Subject:  subject,
			Transfer: s.origin(r) + s.transferPath(subject),
		},
	)
}

func (s *Server) manageUser(w http.ResponseWriter, r *http.Request, u database.User) {
	if !administrator(w, u) {
		return
	}
	id := pathInt(r, "user")
	subject, err := s.DB.User(r.Context(), id)
	if err != nil {
		s.fail(w, err)
		return
	}
	if subject.Status != 0 {
		http.NotFound(w, r)
		return
	}
	if r.Method == "DELETE" {
		err = s.AccountCommands.Deactivate(r.Context(), u.ID, id)
	} else {
		role := 0
		if r.Form.Get("user[role]") == "administrator" {
			role = 1
		}
		_, err = s.AccountCommands.UpdateUser(r.Context(), u.ID, id, database.UserChanges{Role: &role}, application.Attachment{})
	}
	if err != nil {
		s.fail(w, err)
		return
	}
	http.Redirect(w, r, "/account/edit", 302)
}

func (s *Server) banUser(w http.ResponseWriter, r *http.Request, u database.User) {
	if !administrator(w, u) {
		return
	}
	id := pathInt(r, "user")
	if _, err := s.DB.User(r.Context(), id); err != nil {
		s.fail(w, err)
		return
	}
	result, err := s.AccountCommands.Ban(r.Context(), u.ID, id, r.Method == "POST")
	if err != nil {
		s.fail(w, err)
		return
	}
	logAccountProcessing(result.Processing)
	http.Redirect(w, r, fmt.Sprintf("/users/%d", id), 302)
}

func (s *Server) join(w http.ResponseWriter, r *http.Request) {
	if !s.requireUnauthenticated(w, r) {
		return
	}
	account, err := s.DB.Account(r.Context())
	if err != nil {
		s.fail(w, err)
		return
	}
	if account.JoinCode != r.PathValue("code") {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if r.Method == "GET" || r.Method == "HEAD" {
		s.respondPage(w, r, "join", 200, page{Title: "Join " + account.Name, JoinCode: account.JoinCode})
		return
	}
	banned, err := s.DB.BannedIP(r.Context(), remoteIP(r))
	if err != nil {
		s.fail(w, err)
		return
	}
	if banned {
		http.Error(w, "Forbidden", 403)
		return
	}
	password := r.Form.Get("user[password]")
	if password == "" {
		http.Error(w, "Password is required", 422)
		return
	}
	digest, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		http.Error(w, "Invalid password", 422)
		return
	}
	email := r.Form.Get("user[email_address]")
	upload, err := s.optionalUpload(r, "user[avatar]")
	if err != nil {
		s.fail(w, err)
		return
	}
	result, err := s.AccountCommands.Join(r.Context(), r.PathValue("code"), remoteIP(r), database.UserInput{Name: r.Form.Get("user[name]"), Email: email, Password: string(digest)}, application.Attachment{File: upload})
	if err != nil {
		if existing, e := s.DB.UserByEmail(r.Context(), email); e == nil && existing.ID != 0 {
			http.Redirect(w, r, "/session/new?email_address="+url.QueryEscape(email), 302)
			return
		}
		s.fail(w, err)
		return
	}
	logAccountProcessing(result.Processing)
	s.startSession(w, r, result.Commit.User)
}

func (s *Server) bots(w http.ResponseWriter, r *http.Request, u database.User) {
	if !administrator(w, u) {
		return
	}
	bots, err := s.DB.Users(r.Context(), 0, true)
	if err != nil {
		s.fail(w, err)
		return
	}
	p := page{Title: "Bots", User: u, Users: bots}
	for _, bot := range bots {
		rooms, err := s.DB.AllRooms(r.Context(), bot.ID)
		if err != nil {
			s.fail(w, err)
			return
		}
		for i, room := range rooms {
			view, err := s.PageQueries.DisplayRoom(r.Context(), room, bot.Participant())
			if err != nil {
				s.fail(w, err)
				return
			}
			rooms[i] = view.Room
		}
		shared := rooms[:0]
		for _, room := range rooms {
			if room.Type != "Rooms::Direct" {
				shared = append(shared, room)
			}
		}
		p.Bots = append(p.Bots, botView{bot, shared})
	}
	s.respondPage(w, r, "bots", 200, p)
}

func (s *Server) botForm(w http.ResponseWriter, r *http.Request, u database.User) {
	if !administrator(w, u) {
		return
	}
	bot := database.User{Role: 2}
	webhook := ""
	if id := pathInt(r, "bot"); id != 0 {
		var err error
		bot, err = s.DB.User(r.Context(), id)
		if err != nil {
			s.fail(w, err)
			return
		}
		if bot.Role != 2 || bot.Status != 0 {
			http.NotFound(w, r)
			return
		}
		err = s.DB.Read.QueryRowContext(r.Context(), "SELECT url FROM webhooks WHERE user_id=?", id).
			Scan(&webhook)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			s.fail(w, err)
			return
		}
	}
	avatarURL := ""
	if bot.ID != 0 {
		blob, e := s.DB.AttachedBlob(r.Context(), "User", bot.ID, "avatar")
		if e == nil {
			avatarURL = storage.BlobURL(s.Storage.Verifier, blob)
		} else if !errors.Is(e, sql.ErrNoRows) {
			s.fail(w, e)
			return
		}
	}
	s.respondPage(
		w,
		r,
		"bot-form",
		200,
		page{AvatarURL: avatarURL, Title: "Bot settings", User: u, Subject: bot, Webhook: webhook},
	)
}

func (s *Server) saveBot(w http.ResponseWriter, r *http.Request, u database.User) {
	if !administrator(w, u) {
		return
	}
	id := pathInt(r, "bot")
	webhook := r.Form.Get("user[webhook_url]")
	name := r.Form.Get("user[name]")
	upload, err := s.optionalUpload(r, "user[avatar]")
	if err != nil {
		s.fail(w, err)
		return
	}
	if upload != nil {
		defer upload.Discard()
	}
	if id == 0 {
		avatar, e := recordAttachment(r, "user[avatar]", upload, false)
		if e != nil {
			s.fail(w, e)
			return
		}
		result, e := s.AccountCommands.CreateUser(r.Context(), u.ID, database.UserInput{Name: name, Role: 2, Webhook: &webhook}, avatar)
		err = e
		logAccountProcessing(result.Processing)
	} else {
		bot, e := s.DB.User(r.Context(), id)
		if e != nil {
			s.fail(w, e)
			return
		}
		if bot.Role != 2 || bot.Status != 0 {
			http.NotFound(w, r)
			return
		}
		if r.Method == "DELETE" {
			err = s.AccountCommands.DeactivateBot(r.Context(), u.ID, id)
		} else {
			avatar, e := recordAttachment(r, "user[avatar]", upload, false)
			if e != nil {
				s.fail(w, e)
				return
			}
			input := botChanges(r)
			input.Webhook = botWebhook(r)
			result, e := s.AccountCommands.UpdateBot(r.Context(), u.ID, id, input, avatar)
			err = e
			logAccountProcessing(result.Processing)
		}
	}
	if err != nil {
		s.fail(w, err)
		return
	}
	http.Redirect(w, r, "/account/bots", 302)
}

func botChanges(r *http.Request) database.UserChanges {
	fields := database.UserChanges{}
	if r.Form.Has("user[name]") {
		value := r.Form.Get("user[name]")
		fields.Name = &value
	}
	return fields
}

func botWebhook(r *http.Request) *string {
	if !r.Form.Has("user[webhook_url]") {
		return nil
	}
	value := r.Form.Get("user[webhook_url]")
	return &value
}

func (s *Server) rotateBot(w http.ResponseWriter, r *http.Request, u database.User) {
	if !administrator(w, u) {
		return
	}
	id := pathInt(r, "bot")
	bot, err := s.DB.User(r.Context(), id)
	if err != nil {
		s.fail(w, err)
		return
	}
	if bot.Role != 2 || bot.Status != 0 {
		http.NotFound(w, r)
		return
	}
	value := database.RandomToken(12)
	if _, err = s.AccountCommands.UpdateBot(r.Context(), u.ID, id, database.UserChanges{BotToken: &value}, application.Attachment{}); err != nil {
		s.fail(w, err)
		return
	}
	http.Redirect(w, r, "/account/bots", 302)
}

func (s *Server) transfer(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" || r.Method == "HEAD" {
		s.respondPage(w, r, "transfer", 200, page{Title: "Sign in", Transfer: r.PathValue("token")})
		return
	}
	id, err := s.Secrets.VerifyID("User", r.PathValue("token"), "transfer", s.DB.Now())
	if err != nil {
		http.Error(w, "Bad request", 400)
		return
	}
	u, err := s.DB.User(r.Context(), id)
	if err != nil || u.Status != 0 {
		http.Error(w, "Bad request", 400)
		return
	}
	s.startSession(w, r, u)
}

// Transfer links expire after the same four-hour window as User#transfer_id.
func (s *Server) transferPath(u database.User) string {
	return "/session/transfers/" + s.Secrets.SignedID(
		"User",
		u.ID,
		"transfer",
		s.DB.Now().Add(4*time.Hour),
	)
}
