package theme

import (
	"strings"
	"testing"
)

func TestGet(t *testing.T) {
	for _, name := range []string{"tufte", "dark", "modern"} {
		th := Get(name)
		if th.Name != name {
			t.Fatalf("Get(%q) = %q", name, th.Name)
		}
		if th.CSS == "" || th.TypstPrelude == "" {
			t.Fatalf("theme %q missing CSS/prelude", name)
		}
		if !strings.Contains(th.TypstPrelude, "{{TITLE}}") {
			t.Fatalf("theme %q prelude missing {{TITLE}}", name)
		}
	}
}

func TestFallback(t *testing.T) {
	if Get("nope").Name != Default {
		t.Fatal("expected default fallback")
	}
	if Get("").Name != Default {
		t.Fatal("expected default for empty")
	}
	if Valid("nope") || !Valid("dark") {
		t.Fatal("Valid mismatch")
	}
}
