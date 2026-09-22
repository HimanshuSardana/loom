package html

import (
	"html"
	"regexp"
	"strings"
	"testing"
)

var tagRe = regexp.MustCompile(`<[^>]*>`)

func TestHighlightRoundTrip(t *testing.T) {
	sources := map[string]string{
		"python":     "# comment\nx = \"hi # there\"\ndef foo(a):\n    return a + 42\nprint(foo(1))\n",
		"bash":       "echo 'single' # trail\ncount=3\n",
		"javascript": "const x = 1; // c\nfunction f(y) { return y * 2; }\n",
		"go":         "package main\nvar x = 10\n",
	}
	for lang, src := range sources {
		out := highlight(src, lang)
		stripped := tagRe.ReplaceAllString(out, "")
		if want := html.EscapeString(src); stripped != want {
			t.Errorf("lang=%s round trip mismatch:\n got %q\nwant %q", lang, stripped, want)
		}
	}
}

func TestHighlightTokens(t *testing.T) {
	out := highlight("def foo():\n    return 42", "python")
	for _, want := range []string{`tok-k`, `tok-f`, `tok-n`} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %s in %s", want, out)
		}
	}
	if !strings.Contains(out, `class="tok-k">def<`) {
		t.Errorf("def not keyword-colored: %s", out)
	}
}

func TestHighlightEscapes(t *testing.T) {
	out := highlight("<script>alert('x')</script>", "generic")
	if strings.Contains(out, "<script>") {
		t.Errorf("not escaped: %s", out)
	}
}
