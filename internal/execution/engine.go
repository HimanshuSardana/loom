package execution

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/HimanshuSardana/loom/internal/ast"
	"github.com/HimanshuSardana/loom/internal/cache"
	"github.com/HimanshuSardana/loom/internal/runtime"
)

// Engine executes runnable blocks with sessions, deps, cache.
type Engine struct {
	Runtimes map[string]runtime.Runtime
	Workdir  string
	LoomFile string
	UseCache bool
}

func New(workdir, loomFile string) *Engine {
	return &Engine{Runtimes: runtime.DefaultRegistry(), Workdir: workdir, LoomFile: loomFile, UseCache: true}
}

// Order returns topological order honoring {depends=}, error on missing/cycle.
func Order(doc *ast.Document, only []string) ([]*ast.Block, error) {
	byName := map[string]*ast.Block{}
	for _, b := range doc.CodeBlocks() {
		if b.Name != "" {
			byName[b.Name] = b
		}
	}
	// select set: if only given (single block), include transitive deps; else all runnable
	var targets []*ast.Block
	if len(only) > 0 {
		visited := map[string]bool{}
		var visit func(name string, stack []string) error
		visit = func(name string, stack []string) error {
			for _, s := range stack {
				if s == name {
					return fmt.Errorf("loom: circular dependency:\n%s", strings.Join(append(stack, name), " → "))
				}
			}
			if visited[name] {
				return nil
			}
			b, ok := byName[name]
			if !ok {
				return fmt.Errorf("loom: undefined block '%s'", name)
			}
			visited[name] = true
			for _, d := range b.Depends {
				if err := visit(d, append(stack, name)); err != nil {
					return err
				}
			}
			targets = append(targets, b)
			return nil
		}
		for _, n := range only {
			if err := visit(n, nil); err != nil {
				return nil, err
			}
		}
		return targets, nil
	}
	// all runnable in doc order, but deps-first via topo sort
	runnable := doc.Runnable()
	// Kahn on runnable subgraph; deps on non-runnable blocks are still executed if needed? For v1: include dep blocks even if not {run}.
	needed := map[string]*ast.Block{}
	var add func(b *ast.Block, stack []string) error
	add = func(b *ast.Block, stack []string) error {
		if b.Name != "" {
			for _, s := range stack {
				if s == b.Name {
					return fmt.Errorf("loom: circular dependency:\n%s", strings.Join(append(stack, b.Name), " → "))
				}
			}
			if _, ok := needed[b.Name]; ok && b.Name != "" {
				return nil
			}
		}
		for _, d := range b.Depends {
			t, ok := byName[d]
			if !ok {
				return fmt.Errorf("loom: undefined block '%s' (dependency of '%s')", d, b.Name)
			}
			if err := add(t, append(stack, b.Name)); err != nil {
				return err
			}
		}
		if b.Name != "" {
			needed[b.Name] = b
		} else {
			needed[fmt.Sprintf("%p", b)] = b
		}
		return nil
	}
	for _, b := range runnable {
		if err := add(b, nil); err != nil {
			return nil, err
		}
	}
	// preserve document order
	order := map[*ast.Block]int{}
	for i := range doc.Blocks {
		if doc.Blocks[i].Type == ast.BlockCode {
			order[&doc.Blocks[i]] = i
		}
	}
	var out []*ast.Block
	for _, b := range needed {
		out = append(out, b)
	}
	sort.Slice(out, func(i, j int) bool { return order[out[i]] < order[out[j]] })
	// stable topo: iterate until deps satisfied (doc order mostly satisfies)
	placed := map[*ast.Block]bool{}
	var sorted []*ast.Block
	remaining := append([]*ast.Block(nil), out...)
	for len(remaining) > 0 {
		progress := false
		var next []*ast.Block
		for _, b := range remaining {
			ok := true
			for _, d := range b.Depends {
				found := false
				for _, s := range sorted {
					if s.Name == d {
						found = true
						break
					}
				}
				// dep not in needed set? treat as satisfied check byName runnable? it is in needed if we added.
				if !found {
					// check if dep block exists but not yet placed
					ok = false
					break
				}
			}
			if ok {
				sorted = append(sorted, b)
				placed[b] = true
				progress = true
			} else {
				next = append(next, b)
			}
		}
		if !progress {
			names := []string{}
			for _, b := range next {
				names = append(names, b.Name)
			}
			return nil, fmt.Errorf("loom: circular dependency among: %s", strings.Join(names, ", "))
		}
		remaining = next
	}
	return sorted, nil
}

// Run executes ordered blocks. Sessions share state by concatenating prior
// code in the same session (Python semantics approximated by re-execution
// in a single process per session-prefix).
func (e *Engine) Run(doc *ast.Document, only []string) ([]runtime.Result, error) {
	ordered, err := Order(doc, only)
	if err != nil {
		return nil, err
	}
	// group session history
	sessionHist := map[string]string{} // session -> accumulated code
	// cache invalidation: if a dep changed (executed fresh), dependents must re-run
	dirty := map[string]bool{}
	var results []runtime.Result
	for _, b := range ordered {
		// stale if any dep dirty
		staleDep := false
		for _, d := range b.Depends {
			if dirty[d] {
				staleDep = true
			}
		}
		key := ""
		if b.Name != "" {
			key = cache.Key(b.Name, b.Source, b.Language, b.Session, strings.Join(b.Depends, ","))
			if e.UseCache && !staleDep && !b.Isolated {
				if hit, ok := cache.Lookup(e.LoomFile, b.Name, key); ok {
					// still need to extend session history for downstream blocks
					if b.Session != "" && !b.Isolated {
						sessionHist[b.Session] += "\n" + b.Source
					}
					hit.Block = b.Name
					results = append(results, hit)
					continue
				}
			}
		}
		rt, ok := e.Runtimes[strings.ToLower(b.Language)]
		if !ok {
			return results, fmt.Errorf("%s:%d:%d\nblock: %s\nruntime: %s\n\nunknown language/runtime '%s'", e.LoomFile, b.Loc.Start.Line, b.Loc.Start.Column, b.Name, b.Language, b.Language)
		}
		if !rt.IsAvailable() {
			return results, fmt.Errorf("%s:%d:%d\nblock: %s\nruntime: %s\n\nruntime '%s' is not available on PATH", e.LoomFile, b.Loc.Start.Line, b.Loc.Start.Column, b.Name, b.Language, b.Language)
		}
		code := b.Source
		if b.Session != "" && !b.Isolated {
			if hist, ok := sessionHist[b.Session]; ok && hist != "" {
				// Only python/js/shell benefit; concatenate history + current
				// Execute combined to share state, but per-block stdout separation
				// is approximated: we run combined and attribute full stdout to current block.
				// To keep per-block attribution sane, run history silently then current with history prepended?
				// Simplest correct: combined execution, strip prior runs by running history alone first (cached silently).
				code = hist + "\n" + b.Source
			}
		}
		res := rt.Execute(code, e.Workdir)
		res.Block = b.Name
		// If session-combined, try to isolate current block's stdout by re-running history alone and diffing prefix.
		// Cheap approach: if combined, run history-only to get prefix length and trim.
		// Skip for now: runtime tests use distinct outputs; combined stdout includes prior prints.
		// Better: execute incrementally in one process. For v1 we accept combined stdout but trim by running
		// history prefix measurement for python (stdout only).
		if b.Session != "" && !b.Isolated {
			if hist := sessionHist[b.Session]; hist != "" {
				prev := rt.Execute(hist, e.Workdir)
				if len(res.Stdout) >= len(prev.Stdout) && strings.HasPrefix(res.Stdout, prev.Stdout) {
					res.Stdout = res.Stdout[len(prev.Stdout):]
				}
			}
			sessionHist[b.Session] += "\n" + b.Source
		}
		if b.Name != "" && key != "" && res.Success && !b.Isolated {
			_ = cache.Store(e.LoomFile, b.Name, key, res)
		}
		if !res.Success {
			results = append(results, res)
			return results, &ExecError{File: e.LoomFile, Block: b, Result: res}
		}
		if b.Name != "" {
			dirty[b.Name] = true // executed fresh (cache miss); dependents recompute... but only if source changed?
			// Actually mark dirty only on miss; hits above `continue` so reaching here means miss.
		}
		// fix artifact paths relative note
		results = append(results, res)
	}
	_ = filepath.Separator
	return results, nil
}

type ExecError struct {
	File   string
	Block  *ast.Block
	Result runtime.Result
}

func (e *ExecError) Error() string {
	msg := fmt.Sprintf("%s:%d:%d\nblock: %s\nruntime: %s\n", e.File, e.Block.Loc.Start.Line, e.Block.Loc.Start.Column, e.Block.Name, e.Block.Language)
	if e.Result.Stderr != "" {
		msg += "\n" + strings.TrimSpace(e.Result.Stderr)
	} else if e.Result.Error != "" {
		msg += "\n" + e.Result.Error
	} else {
		msg += "\nexecution failed"
	}
	if e.Result.ExitCode != 0 {
		msg += fmt.Sprintf("\n(exit code %d)", e.Result.ExitCode)
	}
	return msg
}
