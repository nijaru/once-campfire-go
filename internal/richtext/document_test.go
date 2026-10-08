package richtext

import (
	"html"
	"slices"
	"strings"
	"testing"
)

func TestPreparedNestedMentionsUseRenderingSanitization(t *testing.T) {
	for _, trix := range []string{`{"caption":"x"}`, `["invalid trix merge"]`} {
		nested := `<action-text-attachment sgid="nested-user" data-trix-attachment='` + trix + `'></action-text-attachment>`
		body := `<action-text-attachment content-type="text/html" content="` + html.EscapeString(nested) + `"></action-text-attachment>`
		doc := Prepare(body)
		tokens := doc.Attachables()
		if !slices.Equal(tokens, []string{"nested-user"}) {
			t.Fatalf("nested token lost after sanitization: %v", tokens)
		}
		ctx := Context{Resolve: func(token string, _ bool) (*Mention, error) {
			if !slices.Contains(tokens, token) {
				return nil, nil
			}
			return &Mention{ID: 7, Name: "Mentioned", SGID: token, Path: "/users/7", Avatar: "/avatar"}, nil
		}}
		first, err := doc.Display(ctx)
		if err != nil || !strings.Contains(first.Presentation, "Mentioned") {
			t.Fatalf("nested mention lost: %+v %v", first, err)
		}
		second, err := doc.Display(ctx)
		if err != nil || first.Presentation != second.Presentation {
			t.Fatal("prepared tree was mutated", err)
		}
	}
}
