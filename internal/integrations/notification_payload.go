package integrations

import (
	"encoding/json"
	"fmt"

	"github.com/basecamp/once-campfire-go/internal/rails"
)

func NotificationJSON(title, body, path string, badge int64) []byte {
	var payload struct {
		Title   string `json:"title"`
		Options struct {
			Body string `json:"body"`
			Icon string `json:"icon"`
			Data struct {
				Path  string `json:"path"`
				Badge int64  `json:"badge"`
			} `json:"data"`
		} `json:"options"`
	}
	payload.Title = title
	payload.Options.Body, payload.Options.Icon = body, "/account/logo"
	payload.Options.Data.Path, payload.Options.Data.Badge = path, badge
	raw, _ := json.Marshal(payload)
	encoded, _ := rails.CanonicalJSON(raw, false)
	return encoded
}

// WebhookContent owns provider payload inputs. HTML remains nullable and raw;
// plaintext is prepared separately for this consumer.
type WebhookContent struct {
	CreatorID, RoomID, MessageID     int64
	Creator, RoomName, BotKey, Plain string
	HTML                             *string
}

func (content WebhookContent) JSON() ([]byte, error) {
	var roomName *string
	if content.RoomName != "" {
		roomName = &content.RoomName
	}
	payload := struct {
		User struct {
			ID   int64  `json:"id"`
			Name string `json:"name"`
		} `json:"user"`
		Room struct {
			ID   int64   `json:"id"`
			Name *string `json:"name"`
			Path string  `json:"path"`
		} `json:"room"`
		Message struct {
			ID   int64 `json:"id"`
			Body struct {
				HTML  *string `json:"html"`
				Plain string  `json:"plain"`
			} `json:"body"`
			Path string `json:"path"`
		} `json:"message"`
	}{}
	payload.User.ID = content.CreatorID
	payload.User.Name = content.Creator
	payload.Room.ID = content.RoomID
	payload.Room.Name = roomName
	payload.Room.Path = fmt.Sprintf("/rooms/%d/%s/messages", content.RoomID, content.BotKey)
	payload.Message.ID = content.MessageID
	payload.Message.Body.HTML = content.HTML
	payload.Message.Body.Plain = content.Plain
	payload.Message.Path = fmt.Sprintf("/rooms/%d/@%d", content.RoomID, content.MessageID)
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	raw, err = rails.CanonicalJSON(raw, true)
	if err != nil {
		return nil, err
	}

	return raw, nil
}
