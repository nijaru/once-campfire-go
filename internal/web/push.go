package web

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"uuid"

	"github.com/basecamp/once-campfire-go/internal/database"
)

const pushPath = "/users/me/push_subscriptions"

func subscriptionParams(r *http.Request) (map[string]*string, error) {
	attrs := map[string]*string{}
	if strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		var raw map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
			return nil, err
		}
		if nested, ok := raw["push_subscription"]; ok {
			if err := json.Unmarshal(nested, &raw); err != nil {
				return nil, err
			}
		}
		if len(raw) == 0 {
			return nil, errors.New("push_subscription is required")
		}
		for _, key := range []string{"endpoint", "p256dh_key", "auth_key"} {
			if value, ok := raw[key]; ok {
				var str *string
				if err := json.Unmarshal(value, &str); err == nil {
					attrs[key] = str
				}
			}
		}
		return attrs, nil
	}
	present := false
	for key := range r.Form {
		if strings.HasPrefix(key, "push_subscription[") {
			present = true
		}
	}
	if !present {
		return nil, errors.New("push_subscription is required")
	}
	for _, key := range []string{"endpoint", "p256dh_key", "auth_key"} {
		if r.Form.Has("push_subscription[" + key + "]") {
			str := r.Form.Get("push_subscription[" + key + "]")
			attrs[key] = &str
		}
	}
	return attrs, nil
}

func (s *Server) pushSubscriptions(w http.ResponseWriter, r *http.Request, u database.User) {
	if r.Method == "GET" || r.Method == "HEAD" {
		list, err := s.DB.PushSubscriptions(r.Context(), u.ID)
		if err != nil {
			s.fail(w, err)
			return
		}
		s.respondPage(
			w,
			r,
			"push-subscriptions",
			200,
			page{User: u, Title: "Push notification subscriptions", Subscriptions: list},
		)
		return
	}
	attrs, err := subscriptionParams(r)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	existing, err := s.DB.FindPushSubscription(r.Context(), u.ID, attrs)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		s.fail(w, err)
		return
	}
	endpoint := existing.Endpoint
	if errors.Is(err, sql.ErrNoRows) {
		if value := attrs["endpoint"]; value != nil {
			endpoint = *value
		}
	}
	if err := s.Push.Validate(r.Context(), endpoint); err != nil {
		w.WriteHeader(422)
		return
	}
	if existing.ID != 0 {
		err = s.DB.TouchPushSubscription(r.Context(), existing.ID)
	} else {
		err = s.DB.SavePushSubscription(r.Context(), u.ID, attrs, r.UserAgent())
	}
	if err != nil {
		s.fail(w, err)
		return
	}
	w.WriteHeader(200)
}

func (s *Server) deletePushSubscription(w http.ResponseWriter, r *http.Request, u database.User) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err := s.DB.DeletePushSubscription(r.Context(), u.ID, id); err != nil {
		s.fail(w, err)
		return
	}
	http.Redirect(w, r, s.origin(r)+pushPath, 302)
}

func (s *Server) testPushNotification(w http.ResponseWriter, r *http.Request, u database.User) {
	id, _ := strconv.ParseInt(r.PathValue("push_subscription_id"), 10, 64)
	delivery, err := s.NotificationQueries.TestPush(r.Context(), u.ID, id, "Campfire Test", uuid.NewV4().String(), s.origin(r)+pushPath)
	if err != nil {
		s.fail(w, err)
		return
	}
	subscription := delivery.Subscription
	err = s.Push.Send(r.Context(), subscription.Endpoint, subscription.Key, subscription.Auth, delivery.Payload)
	if err != nil {
		s.fail(w, err)
		return
	}
	http.Redirect(w, r, s.origin(r)+pushPath, 302)
}
