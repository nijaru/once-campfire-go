package richtext

import (
	"reflect"
	"strings"
	"testing"

	xhtml "github.com/basecamp/once-campfire-go/internal/html"
	"golang.org/x/net/html/atom"
)

// The literal-text shortcut must produce the parser's tree, not just equivalent
// display text. Non-body contexts and special input retain HTML parsing.
func TestPresentationTreeMatchesParser(t *testing.T) {
	for _, body := range []string{"", "hello", " a\n b ", "a&gt;b &amp; c", "\ufefftext", "é🙂", "<b>text</b>", "<table><s>reopened<tr><td>cell</s>", "<span>a</span><span>b</span>"} {
		root, err := parse(body)
		if err != nil {
			t.Fatal(err)
		}
		got, err := presentationTree(root)
		if err != nil {
			t.Fatal(err)
		}
		want, err := parse("<div class=\"lexxy-content\">\n  " + serialize(root) + "\n</div>\n")
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("presentation wrapper differs from parser: %q", body)
		}
	}
}

func TestFragmentTextMatchesParser(t *testing.T) {
	bodies := []string{"", "hello", " \t\n\v\f ", "\ufefftext", "a>b", "é\u00a0🙂", "a\r\nb", "a\x00b", string([]byte{0xff}), "&amp;", "<b>text</b>", "<table>text<tr><td>cell", strings.Repeat("text ", 10000)}
	contexts := []*xhtml.Node{nil, {Type: xhtml.ElementNode, Data: "body", DataAtom: atom.Body}, {Type: xhtml.ElementNode, Data: "div", DataAtom: atom.Div}, {Type: xhtml.ElementNode, Data: "table", DataAtom: atom.Table}, {Type: xhtml.ElementNode, Data: "textarea", DataAtom: atom.Textarea}, {Type: xhtml.ElementNode, Data: "script", DataAtom: atom.Script}}
	for _, context := range contexts {
		for _, body := range bodies {
			got, err := parseIn(body, context)
			if err != nil {
				t.Fatal(err)
			}
			reference := context
			if reference == nil {
				reference = &xhtml.Node{Type: xhtml.ElementNode, Data: "body", DataAtom: atom.Body}
			}
			nodes, err := xhtml.ParseFragmentWithOptions(strings.NewReader(strings.TrimPrefix(body, "\ufeff")), reference, xhtml.ParseOptionEnableScripting(false))
			if err != nil {
				t.Fatal(err)
			}
			want := &xhtml.Node{Type: xhtml.DocumentNode}
			for _, n := range nodes {
				want.AppendChild(n)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("fragment differs from parser: context=%v body=%q", context, body)
			}
		}
	}
}
