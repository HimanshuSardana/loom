package theme

import (
	"strings"
	"testing"
)

func TestGet(t *testing.T) {
	for _, th := range List() {
		got := Get(th.Name)
		if got.Name != th.Name {
			t.Fatalf("Get(%q) = %q", th.Name, got.Name)
		}
		if got.CSS == "" || got.TypstPrelude == "" {
			t.Fatalf("theme %q missing CSS/prelude", th.Name)
		}
		if !strings.Contains(got.TypstPrelude, "{{TITLE}}") {
			t.Fatalf("theme %q prelude missing {{TITLE}}", th.Name)
		}
		if !Valid(th.Name) {
			t.Fatalf("Valid(%q) = false", th.Name)
		}
	}
	if len(List()) < 7 {
		t.Fatalf("expected at least 7 themes, got %d", len(List()))
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
