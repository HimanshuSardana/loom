package runtime

import "testing"

func TestPython(t *testing.T) {
	rt := PythonRuntime{}
	if !rt.IsAvailable() {
		t.Skip("no python3")
	}
	res := rt.Execute("print('hi')", t.TempDir())
	if !res.Success || res.Stdout != "hi\n" {
		t.Fatalf("%+v", res)
	}
}

func TestPythonValue(t *testing.T) {
	rt := PythonRuntime{}
	if !rt.IsAvailable() {
		t.Skip("no python3")
	}
	res := rt.Execute("x = 10\nx * 2", t.TempDir())
	if !res.Success || res.Value != "20" {
		t.Fatalf("%+v", res)
	}
}

func TestPythonFail(t *testing.T) {
	rt := PythonRuntime{}
	if !rt.IsAvailable() {
		t.Skip("no python3")
	}
	res := rt.Execute("raise ValueError('boom')", t.TempDir())
	if res.Success {
		t.Fatalf("expected failure: %+v", res)
	}
}

func TestShell(t *testing.T) {
	rt := ShellRuntime{}
	if !rt.IsAvailable() {
		t.Skip("no bash")
	}
	res := rt.Execute("echo hello", t.TempDir())
	if !res.Success {
		t.Fatalf("%+v", res)
	}
}
