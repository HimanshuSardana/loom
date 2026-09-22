package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/HimanshuSardana/loom/internal/ast"
	"github.com/HimanshuSardana/loom/internal/cache"
	"github.com/HimanshuSardana/loom/internal/config"
	"github.com/HimanshuSardana/loom/internal/execution"
	"github.com/HimanshuSardana/loom/internal/export/html"
	"github.com/HimanshuSardana/loom/internal/export/typst"
	"github.com/HimanshuSardana/loom/internal/parser"
	"github.com/HimanshuSardana/loom/internal/runtime"
	"github.com/HimanshuSardana/loom/internal/tangle"
	"github.com/HimanshuSardana/loom/internal/theme"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	cmd := os.Args[1]
	args := os.Args[2:]
	var err error
	switch cmd {
	case "parse":
		err = cmdParse(args)
	case "check":
		err = cmdCheck(args)
	case "tangle":
		err = cmdTangle(args)
	case "run":
		err = cmdRun(args)
	case "export":
		err = cmdExport(args)
	case "build":
		err = cmdBuild(args)
	case "clean":
		err = cmdClean(args)
	case "watch":
		err = cmdWatch(args)
	case "themes":
		err = cmdThemes(args)
	case "--help", "-h", "help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "loom: unknown command '%s'\n", cmd)
		usage()
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func usage() {
	fmt.Println(`loom — executable literate programming

Usage:
  loom parse <file> [--json]
  loom check <file>
  loom tangle <file> [--out <dir>] [--stdout]
  loom run <file> [--block <name>] [--allow-execution] [--no-cache] [--json]
  loom export html <file> [--out <path>] [--theme <name>]
  loom export pdf <file> [--out <path>] [--theme <name>]
  loom build <file> [--allow-execution] [--theme <name>]
  loom clean [<file>]
  loom watch <file> [--allow-execution]
  loom themes

Themes: tufte, dark, modern (default: modern).`)
}

// normalizeArgs moves flags before positionals so `loom run file --block x`
// works as well as `loom run --block x file` (std flag pkg stops at first arg).
func normalizeArgs(args []string) []string {
	withVal := map[string]bool{"--out": true, "--block": true, "--theme": true, "-out": true, "-block": true, "-theme": true}
	var flags, pos []string
	i := 0
	for i < len(args) {
		a := args[i]
		if strings.HasPrefix(a, "-") && a != "-" {
			name := a
			if eq := strings.Index(a, "="); eq >= 0 {
				name = a[:eq]
			}
			flags = append(flags, a)
			if withVal[name] && !strings.Contains(a, "=") && i+1 < len(args) {
				i++
				flags = append(flags, args[i])
			}
		} else {
			pos = append(pos, a)
		}
		i++
	}
	return append(flags, pos...)
}

func loadDoc(file string) (*ast.Document, error) {
	res, err := parser.ParseFile(file)
	if err != nil {
		return nil, err
	}
	if res.Diagnostics.HasErrors() {
		return nil, fmt.Errorf("%s", res.Diagnostics.Error())
	}
	return res.Doc, nil
}

func cmdParse(args []string) error {
	fs := flag.NewFlagSet("parse", flag.ContinueOnError)
	asJSON := fs.Bool("json", false, "output JSON AST")
	if err := fs.Parse(normalizeArgs(args)); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return fmt.Errorf("usage: loom parse <file>")
	}
	file := fs.Arg(0)
	res, err := parser.ParseFile(file)
	if err != nil {
		return err
	}
	if *asJSON {
		data, _ := json.MarshalIndent(res.Doc, "", "  ")
		fmt.Println(string(data))
		return nil
	}
	fmt.Printf("document: %s\ntitle: %s\nblocks: %d\n", file, res.Doc.Title, len(res.Doc.Blocks))
	for _, b := range res.Doc.Blocks {
		switch b.Type {
		case ast.BlockCode:
			fmt.Printf("  code %s [%s] run=%v session=%s file=%s (line %d)\n", b.Name, b.Language, b.Run, b.Session, b.File, b.Loc.Start.Line)
		case ast.BlockHeading:
			fmt.Printf("  h%d %s\n", b.Level, b.Text)
		default:
			fmt.Printf("  %s\n", b.Type)
		}
	}
	if res.Diagnostics.HasErrors() {
		return fmt.Errorf("%s", res.Diagnostics.Error())
	}
	return nil
}

func cmdCheck(args []string) error {
	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	if err := fs.Parse(normalizeArgs(args)); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return fmt.Errorf("usage: loom check <file>")
	}
	file := fs.Arg(0)
	res, err := parser.ParseFile(file)
	if err != nil {
		return err
	}
	if res.Diagnostics.HasErrors() {
		return fmt.Errorf("%s", res.Diagnostics.Error())
	}
	// validate references & deps
	for _, b := range res.Doc.CodeBlocks() {
		if b.Name != "" {
			if _, err := tangle.Resolve(res.Doc, b.Name); err != nil {
				return err
			}
		}
		for _, d := range b.Depends {
			if res.Doc.FindBlock(d) == nil {
				return fmt.Errorf("loom: undefined block '%s' (dependency of '%s')\n%s:%d:%d", d, b.Name, file, b.Loc.Start.Line, b.Loc.Start.Column)
			}
		}
	}
	if _, err := execution.Order(res.Doc, nil); err != nil {
		return err
	}
	fmt.Println("ok")
	return nil
}

func confirmExecution(file string, allow bool) error {
	if allow {
		return nil
	}
	// Non-interactive (piped stdin or CI env) -> require flag
	fi, _ := os.Stdin.Stat()
	if fi == nil || (fi.Mode()&os.ModeCharDevice) == 0 {
		return fmt.Errorf("loom: this document contains executable blocks.\nRe-run with --allow-execution to execute code non-interactively")
	}
	if os.Getenv("CI") != "" {
		return fmt.Errorf("loom: this document contains executable blocks.\nRe-run with --allow-execution to execute code non-interactively")
	}
	fmt.Fprintf(os.Stderr, "This document contains executable blocks.\nRun code? [y/N] ")
	var ans string
	fmt.Scanln(&ans)
	ans = strings.ToLower(strings.TrimSpace(ans))
	if ans != "y" && ans != "yes" {
		return fmt.Errorf("loom: execution declined by user")
	}
	return nil
}

func lastLoomDir(file string) string { return filepath.Join(filepath.Dir(file), ".loom") }

func cmdTangle(args []string) error {
	fs := flag.NewFlagSet("tangle", flag.ContinueOnError)
	out := fs.String("out", "", "output directory")
	toStdout := fs.Bool("stdout", false, "print to stdout instead of writing files")
	if err := fs.Parse(normalizeArgs(args)); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return fmt.Errorf("usage: loom tangle <file>")
	}
	file := fs.Arg(0)
	doc, err := loadDoc(file)
	if err != nil {
		return err
	}
	cfg := config.Load(file)
	def := *out
	if def == "" {
		def = cfg.TangleOut
	}
	files, err := tangle.Files(doc, "")
	if err != nil {
		return err
	}
	if len(files) == 0 {
		// fallback: if single default desired, still nothing to do
		fmt.Fprintln(os.Stderr, "loom: nothing to tangle (no {file=} targets)")
		return nil
	}
	if *toStdout {
		for name, content := range files {
			fmt.Printf("=== %s ===\n%s\n", name, content)
		}
		return nil
	}
	base := def
	if !filepath.IsAbs(base) {
		base = filepath.Join(filepath.Dir(file), base)
	}
	for name, content := range files {
		p := filepath.Join(base, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			return err
		}
		fmt.Printf("tangled %s\n", p)
	}
	return nil
}

func cmdRun(args []string) error {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	block := fs.String("block", "", "run single block (plus its dependencies)")
	allow := fs.Bool("allow-execution", false, "allow execution without prompt")
	noCache := fs.Bool("no-cache", false, "disable cache")
	asJSON := fs.Bool("json", false, "JSON output")
	if err := fs.Parse(normalizeArgs(args)); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return fmt.Errorf("usage: loom run <file> [--block <name>]")
	}
	file := fs.Arg(0)
	doc, err := loadDoc(file)
	if err != nil {
		return err
	}
	var only []string
	if *block != "" {
		if doc.FindBlock(*block) == nil {
			return fmt.Errorf("loom: undefined block '%s'", *block)
		}
		only = []string{*block}
	} else if len(doc.Runnable()) == 0 {
		fmt.Fprintln(os.Stderr, "loom: no executable blocks ({run}) found")
		return nil
	}
	if err := confirmExecution(file, *allow); err != nil {
		return err
	}
	workdir := filepath.Dir(file)
	eng := execution.New(workdir, file)
	if *noCache {
		eng.UseCache = false
	}
	results, err := eng.Run(doc, only)
	if *asJSON {
		data, _ := json.MarshalIndent(results, "", "  ")
		fmt.Println(string(data))
	} else {
		for _, r := range results {
			printResult(r)
		}
	}
	return err
}

func printResult(r runtime.Result) {
	status := "ok"
	if !r.Success {
		status = "FAILED"
	}
	fmt.Printf("━━ block %s [%s] %s (%dms, exit %d) ━━\n", r.Block, r.Runtime, status, r.DurationMs, r.ExitCode)
	if r.Stdout != "" {
		fmt.Printf("%s", ensureNL(r.Stdout))
	}
	if r.Value != "" {
		fmt.Printf("=> %s\n", r.Value)
	}
	if r.Stderr != "" {
		fmt.Fprintf(os.Stderr, "%s", ensureNL(r.Stderr))
	}
	for _, a := range r.Artifacts {
		fmt.Printf("(artifact: %s)\n", a)
	}
}

func ensureNL(s string) string {
	if s == "" || strings.HasSuffix(s, "\n") {
		return s
	}
	return s + "\n"
}

// loadResults reads cached results for export (best effort).
func loadResults(file string, doc *ast.Document) map[string]runtime.Result {
	out := map[string]runtime.Result{}
	for _, b := range doc.CodeBlocks() {
		if b.Name == "" || !b.Run {
			continue
		}
		key := cache.Key(b.Name, b.Source, b.Language, b.Session, strings.Join(b.Depends, ","))
		if hit, ok := cache.Lookup(file, b.Name, key); ok {
			out[b.Name] = hit
		}
	}
	return out
}

func cmdExport(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: loom export html|pdf <file>")
	}
	format := args[0]
	rest := args[1:]
	fs := flag.NewFlagSet("export", flag.ContinueOnError)
	out := fs.String("out", "", "output path")
	themeFlag := fs.String("theme", "", "theme name ("+theme.Names()+")")
	if err := fs.Parse(normalizeArgs(rest)); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return fmt.Errorf("usage: loom export html|pdf <file>")
	}
	file := fs.Arg(0)
	doc, err := loadDoc(file)
	if err != nil {
		return err
	}
	cfg := config.Load(file)
	results := loadResults(file, doc)
	switch format {
	case "html":
		dest := *out
		if dest == "" {
			dest = cfg.HTML.Output
		}
		th := *themeFlag
		if th == "" {
			th = cfg.HTML.Theme
		}
		if th != "" && !theme.Valid(th) {
			return fmt.Errorf("loom: unknown theme '%s' (available: %s)", th, theme.Names())
		}
		if !filepath.IsAbs(dest) {
			dest = filepath.Join(filepath.Dir(file), dest)
		}
		page := html.Render(doc, results, th)
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(dest, []byte(page), 0o644); err != nil {
			return err
		}
		// copy image artifacts next to output
		for _, r := range results {
			for _, a := range r.Artifacts {
				src := filepath.Join(filepath.Dir(file), a)
				dst := filepath.Join(filepath.Dir(dest), filepath.Base(a))
				if data, err := os.ReadFile(src); err == nil {
					_ = os.WriteFile(dst, data, 0o644)
				}
			}
		}
		fmt.Printf("exported %s\n", dest)
	case "pdf":
		dest := *out
		if dest == "" {
			dest = cfg.PDF.Output
		}
		th := *themeFlag
		if th == "" {
			th = cfg.PDF.Theme
		}
		if th != "" && !theme.Valid(th) {
			return fmt.Errorf("loom: unknown theme '%s' (available: %s)", th, theme.Names())
		}
		if !filepath.IsAbs(dest) {
			dest = filepath.Join(filepath.Dir(file), dest)
		}
		markup := typst.Render(doc, results, th)
		tmp, err := os.CreateTemp("", "loom-*.typ")
		if err != nil {
			return err
		}
		tmp.WriteString(markup)
		tmp.Close()
		defer os.Remove(tmp.Name())
		if _, err := exec.LookPath("typst"); err != nil {
			// still write .typ for inspection, clear diagnostic
			typPath := strings.TrimSuffix(dest, filepath.Ext(dest)) + ".typ"
			_ = os.MkdirAll(filepath.Dir(typPath), 0o755)
			_ = os.WriteFile(typPath, []byte(markup), 0o644)
			return fmt.Errorf("loom: typst not found on PATH (wrote %s instead).\nInstall typst: https://github.com/typst/typst", typPath)
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}
		cmd := exec.Command("typst", "compile", tmp.Name(), dest)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("loom: typst compile failed: %v", err)
		}
		fmt.Printf("exported %s\n", dest)
	default:
		return fmt.Errorf("usage: loom export html|pdf <file>")
	}
	return nil
}

func cmdBuild(args []string) error {
	fs := flag.NewFlagSet("build", flag.ContinueOnError)
	allow := fs.Bool("allow-execution", false, "allow execution without prompt")
	themeFlag := fs.String("theme", "", "theme for HTML/PDF export ("+theme.Names()+")")
	if err := fs.Parse(normalizeArgs(args)); err != nil {
		return err
	}
	if *themeFlag != "" && !theme.Valid(*themeFlag) {
		return fmt.Errorf("loom: unknown theme '%s' (available: %s)", *themeFlag, theme.Names())
	}
	if fs.NArg() < 1 {
		return fmt.Errorf("usage: loom build <file>")
	}
	file := fs.Arg(0)
	if err := cmdCheck([]string{file}); err != nil {
		return err
	}
	doc, err := loadDoc(file)
	if err != nil {
		return err
	}
	if len(doc.Runnable()) > 0 {
		if err := confirmExecution(file, *allow); err != nil {
			return err
		}
		eng := execution.New(filepath.Dir(file), file)
		results, err := eng.Run(doc, nil)
		for _, r := range results {
			printResult(r)
		}
		if err != nil {
			return err
		}
	}
	if err := cmdTangle([]string{file}); err != nil {
		fmt.Fprintln(os.Stderr, "tangle:", err)
	}
	exportArgs := func(format string) []string {
		a := []string{format, file}
		if *themeFlag != "" {
			a = append(a, "--theme", *themeFlag)
		}
		return a
	}
	if err := cmdExport(exportArgs("html")); err != nil {
		fmt.Fprintln(os.Stderr, "export html:", err)
	}
	if err := cmdExport(exportArgs("pdf")); err != nil {
		fmt.Fprintln(os.Stderr, "export pdf:", err)
	}
	fmt.Println("build complete")
	return nil
}

func cmdThemes(args []string) error {
	for _, th := range theme.List() {
		fmt.Printf("%s\n    %s\n", th.Name, th.Description)
	}
	return nil
}

func cmdClean(args []string) error {
	target := ""
	if len(args) > 0 {
		target = args[0]
	}
	if target == "" {
		// clean .loom dirs under cwd
		matches, _ := filepath.Glob(".loom")
		_ = matches
		if _, err := os.Stat(".loom"); err == nil {
			return os.RemoveAll(".loom")
		}
		fmt.Println("nothing to clean")
		return nil
	}
	return cache.Clean(target)
}

func cmdWatch(args []string) error {
	fs := flag.NewFlagSet("watch", flag.ContinueOnError)
	allow := fs.Bool("allow-execution", false, "allow execution without prompt")
	if err := fs.Parse(normalizeArgs(args)); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return fmt.Errorf("usage: loom watch <file>")
	}
	file := fs.Arg(0)
	fmt.Printf("watching %s (Ctrl-C to stop)\n", file)
	var last time.Time
	for {
		fi, err := os.Stat(file)
		if err != nil {
			return err
		}
		if fi.ModTime().After(last) {
			last = fi.ModTime()
			fmt.Println("── change detected, rebuilding ──")
			_ = cmdBuild(append([]string{}, func() []string {
				if *allow {
					return []string{"--allow-execution", file}
				}
				return []string{file}
			}()...))
		}
		time.Sleep(700 * time.Millisecond)
	}
}
