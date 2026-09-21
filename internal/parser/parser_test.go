package parser

import "testing"

func TestBasic(t *testing.T) {
	src := "# Hello\n\nSome text.\n\n```python {#fib} {run}\nprint(1)\n```\n"
	r := Parse("t.loom", src)
	if r.Doc.Title != "Hello" {
		t.Fatalf("title=%q", r.Doc.Title)
	}
	if len(r.Doc.Blocks) != 3 {
		t.Fatalf("blocks=%d", len(r.Doc.Blocks))
	}
	cb := r.Doc.CodeBlocks()[0]
	if cb.Name != "fib" || !cb.Run || cb.Language != "python" {
		t.Fatalf("code block: %+v", cb)
	}
	if r.Diagnostics.HasErrors() {
		t.Fatalf("diags: %s", r.Diagnostics.Error())
	}
}

func TestDuplicate(t *testing.T) {
	src := "```python {#a}\nx\n```\n\n```python {#a}\ny\n```\n"
	r := Parse("t.loom", src)
	if !r.Diagnostics.HasErrors() {
		t.Fatal("expected duplicate error")
	}
}

func TestUnclosed(t *testing.T) {
	src := "```python {#a}\nx\n"
	r := Parse("t.loom", src)
	if !r.Diagnostics.HasErrors() {
		t.Fatal("expected unclosed error")
	}
}

func TestOptions(t *testing.T) {
	src := "```python {#m} {run} {session=s1} {file=\"a.py\"} {depends=x,y} {isolated}\nx\n```\n"
	r := Parse("t.loom", src)
	cb := r.Doc.CodeBlocks()[0]
	if !cb.Run || !cb.Isolated || cb.Session != "s1" || cb.File != "a.py" || len(cb.Depends) != 2 {
		t.Fatalf("opts: %+v", cb)
	}
}
