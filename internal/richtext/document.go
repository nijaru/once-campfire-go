package richtext

import (
	xhtml "github.com/basecamp/once-campfire-go/internal/html"
	"strings"
)

// Document owns parsed content. Its transformations clone the tree so prepared
// content can be reused without mutation. Malformed content retains the existing
// consumer-specific fallback; resolving external data happens before rendering.
type Document struct {
	body string
	root *xhtml.Node
	err  error
}

func Prepare(body string) Document {
	root, err := load(body)
	return Document{body: body, root: root, err: err}
}

// Attachables identifies distinct mention tokens without resolving them or
// generating unused representations. Malformed content has no display targets.
func (doc Document) Attachables() []string {
	if doc.err != nil {
		return nil
	}
	var tokens []string
	seen := map[string]bool{}
	var collect func(*xhtml.Node, int)
	collect = func(root *xhtml.Node, depth int) {
		walk(root, func(n *xhtml.Node) {
			if n.Data != "action-text-attachment" {
				return
			}
			token := attr(n, "sgid")
			if token != "" && !seen[token] {
				seen[token] = true
				tokens = append(tokens, token)
			}
			if depth < 8 && strings.Contains(attr(n, "content-type"), "html") && strings.TrimSpace(attr(n, "content")) != "" {
				content, err := sanitizedAttachmentContent(attr(n, "content"))
				if err == nil {
					if nested, err := load(content); err == nil {
						collect(nested, depth+1)
					}
				}
			}
		})
	}
	collect(doc.root, 0)
	return tokens
}

func (doc Document) Display(ctx Context) (Result, error) {
	if doc.err == nil {
		doc.root = clone(doc.root)
	}
	return doc.process(ctx, displayOutput)
}
