package web

import (
	"context"
	"net"
	"strconv"
	"strings"

	"github.com/basecamp/once-campfire-go/internal/database"
	"github.com/basecamp/once-campfire-go/internal/rails"
	"github.com/basecamp/once-campfire-go/internal/richtext"
)

type mentionTarget struct {
	ID    int64
	Error error
}

func displayMentionTarget(token string) mentionTarget {
	gid, err := rails.UnverifiedUserGID(token)
	if err != nil {
		return mentionTarget{Error: err}
	}
	gid, _, _ = strings.Cut(gid, "?")
	parts := strings.Split(gid, "/")
	if len(parts) != 5 || parts[3] != "User" {
		return mentionTarget{}
	}
	id, err := strconv.ParseInt(parts[4], 10, 64)
	if err != nil {
		return mentionTarget{}
	}
	return mentionTarget{ID: id}
}

// The resolver below only reads owned data. Database/cancellation failures are
// returned during preparation, never mistaken for malformed-content fallbacks.
func (s *Server) resolvedRichContext(ctx context.Context, targets map[string]mentionTarget, users map[int64]database.UserDisplay) richtext.Context {
	var host string
	if info := requestMetadata(ctx); info != nil {
		host = info.host
	}
	if name, _, err := net.SplitHostPort(host); err == nil {
		host = name
	}
	resolved := make(map[int64]*richtext.Mention, len(users))
	for id, user := range users {
		mention := s.displayMention(user)
		resolved[id] = &mention
	}
	now := s.DB.Now()
	return richtext.Context{Host: host, Resolve: func(token string, verified bool) (*richtext.Mention, error) {
		target := targets[token]
		if verified {
			if _, err := s.Secrets.VerifySGID(token, "attachable", now); err != nil {
				return nil, nil
			}
		}
		return resolved[target.ID], target.Error
	}}
}
