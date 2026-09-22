package runtime

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// PythonRuntime executes python3 code.
// Value capture: if the last non-empty line parses as an expression,
// it is evaluated and its repr printed with a marker.
type PythonRuntime struct{ TimeoutSec int }

func (PythonRuntime) Name() string { return "python" }
func (PythonRuntime) IsAvailable() bool {
	_, err := execLookPath("python3")
	return err == nil
}

const valueMarker = "__LOOM_VALUE__"

func pythonWrapper(userCode string) string {
	lines := strings.Split(userCode, "\n")
	// trim trailing empty lines
	end := len(lines)
	for end > 0 && strings.TrimSpace(lines[end-1]) == "" {
		end--
	}
	body := strings.Join(lines[:end], "\n")
	last := ""
	lastRaw := ""
	if end > 0 {
		lastRaw = lines[end-1]
		last = strings.TrimSpace(lastRaw)
	}
	// An indented final line belongs to a compound statement (for/if/with/
	// try body); never treat it as a standalone value expression.
	if lastRaw != "" && (lastRaw[0] == ' ' || lastRaw[0] == '\t') {
		return body + "\n"
	}
	// Heuristic: treat last line as expression if single-line and not assignment/import/def/class/return etc.
	isExpr := last != "" && !strings.Contains(last, "\n") &&
		!hasPrefixAny(last, []string{"import ", "from ", "def ", "class ", "for ", "while ", "if ", "elif ", "else", "try", "except", "finally", "with ", "return", "raise ", "assert ", "del ", "pass", "break", "continue"}) &&
		!strings.Contains(last, "=") || (last != "" && strings.HasPrefix(last, "[") || strings.HasPrefix(last, "{") || strings.HasPrefix(last, "("))
	// refine: if contains "=" but is "==" / "!=" / ">=" / "<=" compare, still expr
	if strings.Contains(last, "==") || strings.Contains(last, "!=") || strings.Contains(last, ">=") || strings.Contains(last, "<=") {
		isExpr = last != "" && !hasPrefixAny(last, []string{"import ", "from ", "def ", "class ", "for ", "while ", "if "})
	}
	if !isExpr || body == "" {
		return body + "\n"
	}
	head := strings.Join(lines[:end-1], "\n")
	// Use ast to safely eval last line as expression; fallback to exec.
	w := ""
	if strings.TrimSpace(head) != "" {
		w += head + "\n"
	}
	w += fmt.Sprintf("_loom_last = %q\n", last) // placeholder replaced below
	_ = w
	script := ""
	if strings.TrimSpace(head) != "" {
		script += head + "\n"
	}
	script += "import ast as _loom_ast\n"
	script += fmt.Sprintf("_loom_src = %s\n", pyQuote(last)) + "try:\n"
	script += "    _loom_tree = _loom_ast.parse(_loom_src, mode='eval')\n"
	script += "    _loom_val = eval(compile(_loom_tree, '<loom>', 'eval'))\n"
	script += fmt.Sprintf("    print(%q, repr(_loom_val))\n", valueMarker)
	script += "except SyntaxError:\n"
	script += "    exec(_loom_src)\n"
	script += "except Exception as _loom_e:\n"
	script += "    raise\n"
	return script
}

func pyQuote(s string) string {
	r := strings.ReplaceAll(s, "\\", "\\\\")
	r = strings.ReplaceAll(r, "\"", "\\\"")
	r = strings.ReplaceAll(r, "\n", "\\n")
	return "\"" + r + "\""
}

func hasPrefixAny(s string, prefs []string) bool {
	t := strings.TrimSpace(s)
	for _, p := range prefs {
		if strings.HasPrefix(t, p) {
			return true
		}
	}
	return false
}

func (r PythonRuntime) Execute(code string, workdir string) Result {
	start := time.Now()
	if workdir == "" {
		workdir = "."
	}
	before := snapshot(workdir)
	script := pythonWrapper(code)
	tmp, err := os.CreateTemp(workdir, ".loom-tmp-*.py")
	if err != nil {
		return Result{Success: false, Error: err.Error(), Runtime: "python"}
	}
	tmpName := tmp.Name()
	tmp.WriteString(script)
	tmp.Close()
	defer os.Remove(tmpName)
	// CreateTemp returns workdir-joined path; command already runs with
	// Dir=workdir, so pass only the basename to avoid doubling the dir.
	rel := filepath.Base(tmpName)
	so, se, code2 := runCmd(workdir, "python3", []string{rel}, r.TimeoutSec)
	after := snapshot(workdir)
	arts := detectArtifacts(before, after, workdir)
	val := ""
	if idx := strings.LastIndex(so, valueMarker); idx >= 0 {
		rest := strings.TrimSpace(so[idx+len(valueMarker):])
		val = strings.TrimSpace(strings.SplitN(rest, "\n", 2)[0])
		if val == "None" {
			val = "" // print(...) etc. return None; not a meaningful value
		}
		so = strings.TrimSpace(strings.Replace(so[:idx], "\n\n", "\n", -1))
		so = strings.TrimSuffix(strings.TrimSpace(so), "\n")
		if so != "" {
			so += "\n"
		} else {
			so = ""
		}
	}
	return Result{
		Success: code2 == 0, Stdout: so, Stderr: se, Value: val,
		Artifacts: arts, ExitCode: code2, DurationMs: duration(start), Runtime: "python",
	}
}
