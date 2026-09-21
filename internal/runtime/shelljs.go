package runtime

import (
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func execLookPath(name string) (string, error) { return exec.LookPath(name) }

// ShellRuntime runs bash code.
type ShellRuntime struct{ TimeoutSec int }

func (ShellRuntime) Name() string { return "shell" }
func (ShellRuntime) IsAvailable() bool {
	_, err := execLookPath("bash")
	return err == nil
}

func (r ShellRuntime) Execute(code string, workdir string) Result {
	start := time.Now()
	if workdir == "" {
		workdir = "."
	}
	before := snapshot(workdir)
	tmp, err := os.CreateTemp(workdir, ".loom-tmp-*.sh")
	if err != nil {
		return Result{Success: false, Error: err.Error(), Runtime: "shell"}
	}
	tmp.WriteString(code + "\n")
	tmp.Close()
	defer os.Remove(tmp.Name())
	so, se, c := runCmd(workdir, "bash", []string{filepath.Base(tmp.Name())}, r.TimeoutSec)
	after := snapshot(workdir)
	return Result{Success: c == 0, Stdout: so, Stderr: se, Artifacts: detectArtifacts(before, after, workdir), ExitCode: c, DurationMs: duration(start), Runtime: "shell"}
}

// JS runtime via node.
type JavascriptRuntime struct{ TimeoutSec int }

func (JavascriptRuntime) Name() string { return "javascript" }
func (JavascriptRuntime) IsAvailable() bool {
	if _, err := execLookPath("node"); err == nil {
		return true
	}
	_, err := execLookPath("bun")
	return err == nil
}

func (r JavascriptRuntime) Execute(code string, workdir string) Result {
	start := time.Now()
	if workdir == "" {
		workdir = "."
	}
	bin := "node"
	if _, err := execLookPath("node"); err != nil {
		bin = "bun"
	}
	before := snapshot(workdir)
	tmp, err := os.CreateTemp(workdir, ".loom-tmp-*.js")
	if err != nil {
		return Result{Success: false, Error: err.Error(), Runtime: "javascript"}
	}
	tmp.WriteString(code + "\n")
	tmp.Close()
	defer os.Remove(tmp.Name())
	so, se, c := runCmd(workdir, bin, []string{filepath.Base(tmp.Name())}, r.TimeoutSec)
	after := snapshot(workdir)
	return Result{Success: c == 0, Stdout: so, Stderr: se, Artifacts: detectArtifacts(before, after, workdir), ExitCode: c, DurationMs: duration(start), Runtime: "javascript"}
}

// Registry maps language -> runtime.
func DefaultRegistry() map[string]Runtime {
	return map[string]Runtime{
		"python":     PythonRuntime{},
		"py":         PythonRuntime{},
		"shell":      ShellRuntime{},
		"bash":       ShellRuntime{},
		"sh":         ShellRuntime{},
		"javascript": JavascriptRuntime{},
		"js":         JavascriptRuntime{},
		"node":       JavascriptRuntime{},
	}
}
