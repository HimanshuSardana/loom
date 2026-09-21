package tangle

import (
	"testing"

	"github.com/HimanshuSardana/loom/internal/parser"
)

func TestResolveNested(t *testing.T) {
	src := "```go {#c}\nC\n```\n\n```go {#b}\nB-<<c>>\n```\n\n```go {#a}\nA-<<b>>\n```\n"
	r := parser.Parse("t.loom", src)
	got, err := Resolve(r.Doc, "a")
	if err != nil {
		t.Fatal(err)
	}
	if got != "A-B-C" {
		t.Fatalf("got %q", got)
	}
}

func TestCircular(t *testing.T) {
	src := "```go {#a}\n<<b>>\n```\n\n```go {#b}\n<<a>>\n```\n"
	r := parser.Parse("t.loom", src)
	if _, err := Resolve(r.Doc, "a"); err == nil {
		t.Fatal("expected circular error")
	}
}

func TestMissing(t *testing.T) {
	src := "```go {#a}\n<<nope>>\n```\n"
	r := parser.Parse("t.loom", src)
	if _, err := Resolve(r.Doc, "a"); err == nil {
		t.Fatal("expected missing error")
	}
}

func TestFiles(t *testing.T) {
	src := "```go {#a} {file=\"a.go\"}\nA\n```\n\n```go {#b} {file=\"b.go\"}\nB\n```\n"
	r := parser.Parse("t.loom", src)
	files, err := Files(r.Doc, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Fatalf("files=%v", files)
	}
}
