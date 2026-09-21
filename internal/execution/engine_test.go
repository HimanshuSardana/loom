package execution

import (
	"testing"

	"github.com/HimanshuSardana/loom/internal/parser"
)

func TestOrderDeps(t *testing.T) {
	src := "```python {#b} {run} {depends=a}\n2\n```\n\n```python {#a} {run}\n1\n```\n"
	r := parser.Parse("t.loom", src)
	ord, err := Order(r.Doc, nil)
	if err != nil {
		t.Fatal(err)
	}
	if ord[0].Name != "a" || ord[1].Name != "b" {
		t.Fatalf("order=%v", ord)
	}
}

func TestSessionSharing(t *testing.T) {
	src := "```python {#a} {run} {session=s}\nx = 10\n```\n\n```python {#b} {run} {session=s}\nx * 2\n```\n"
	r := parser.Parse("t.loom", src)
	dir := t.TempDir()
	eng := New(dir, dir+"/t.loom")
	res, err := eng.Run(r.Doc, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(res) != 2 || res[1].Value != "20" {
		t.Fatalf("%+v", res)
	}
}
