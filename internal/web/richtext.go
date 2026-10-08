package web

import (
	"context"

	"github.com/basecamp/once-campfire-go/internal/database"
	"github.com/basecamp/once-campfire-go/internal/presentation"
	"github.com/basecamp/once-campfire-go/internal/richtext"
)

func (s *Server) mention(u database.User) richtext.Mention {
	return presentation.Mention(s.Secrets, database.UserDisplay{ID: u.ID, Name: u.Name, Bio: u.Bio, UpdatedAt: u.UpdatedAt})
}

func (s *Server) resolvedRichContext(ctx context.Context, targets map[string]presentation.MentionTarget, users map[int64]database.UserDisplay) richtext.Context {
	var host string
	if info := requestMetadata(ctx); info != nil {
		host = info.host
	}
	return presentation.MentionContext(s.Secrets, host, s.DB.Now(), targets, users)
}

func (s *Server) resolveDocument(ctx context.Context, tokens []string) (richtext.Context, error) {
	targets := make(map[string]presentation.MentionTarget, len(tokens))
	ids := make([]int64, 0, len(tokens))
	for _, token := range tokens {
		target := presentation.MentionTargetFor(token)
		targets[token] = target
		if target.ID != 0 {
			ids = append(ids, target.ID)
		}
	}
	users, err := s.DB.UserDisplays(ctx, ids)
	if err != nil {
		return richtext.Context{}, err
	}
	return s.resolvedRichContext(ctx, targets, users), nil
}

func (s *Server) richText(ctx context.Context, body string) (richtext.Result, error) {
	doc := richtext.Prepare(body)
	resolved, err := s.resolveDocument(ctx, doc.Attachables())
	if err != nil {
		return richtext.Result{}, err
	}
	return doc.Content(resolved)
}

func (s *Server) plainText(ctx context.Context, body string) (string, error) {
	doc := richtext.Prepare(body)
	resolved, err := s.resolveDocument(ctx, doc.PlainAttachables())
	if err != nil {
		return "", err
	}
	// Only malformed-content errors remain after preparation; preserve their
	// empty-text fallback rather than swallowing database/cancellation failures.
	plain, _ := doc.PlainText(resolved)
	return plain, nil
}

func (s *Server) mentionedIDs(ctx context.Context, body string) ([]int64, error) {
	doc := richtext.Prepare(body)
	now := s.DB.Now()
	var ids []int64
	for _, token := range doc.MentionAttachables() {
		if id := presentation.VerifiedMentionID(s.Secrets, token, now); id != 0 {
			ids = append(ids, id)
		}
	}
	users, err := s.DB.UserDisplays(ctx, ids)
	if err != nil {
		return nil, err
	}
	var host string
	if info := requestMetadata(ctx); info != nil {
		host = info.host
	}
	resolved := presentation.MentionContext(s.Secrets, host, now, nil, users)
	ids, _ = doc.MentionIDs(resolved)
	return ids, nil
}
