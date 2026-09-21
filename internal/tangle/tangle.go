package tangle

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/HimanshuSardana/loom/internal/ast"
)

var refRe = regexp.MustCompile(`<<([A-Za-z][A-Za-z0-9_.\-]*)>>`)

// Resolve expands <<refs>> in the named block, detecting missing/circular refs.
func Resolve(doc *ast.Document, name string) (string, error) {
	b := doc.FindBlock(name)
	if b == nil {
		return "", fmt.Errorf("loom: undefined block '%s'", name)
	}
	return resolveBlock(doc, b, []string{name})
}

func resolveBlock(doc *ast.Document, b *ast.Block, stack []string) (string, error) {
	src := b.Source
	var sb strings.Builder
	for {
		loc := refRe.FindStringSubmatchIndex(src)
		if loc == nil {
			sb.WriteString(src)
			break
		}
		sb.WriteString(src[:loc[0]])
		refName := src[loc[2]:loc[3]]
		// cycle check
		for _, s := range stack {
			if s == refName {
				cycle := append(append([]string(nil), stack...), refName)
				return "", fmt.Errorf("loom: circular block reference:\n%s", strings.Join(cycle, " → "))
			}
		}
		target := doc.FindBlock(refName)
		if target == nil {
			return "", fmt.Errorf("loom: undefined block '%s' (referenced from block '%s')", refName, b.Name)
		}
		expanded, err := resolveBlock(doc, target, append(stack, refName))
		if err != nil {
			return "", err
		}
		sb.WriteString(expanded)
		src = src[loc[1]:]
	}
	return sb.String(), nil
}

// Files groups resolved blocks by their {file=} target.
// Blocks without file go to defaultFile (may be ""); those are skipped if empty.
func Files(doc *ast.Document, defaultFile string) (map[string]string, error) {
	out := map[string]string{}
	for _, b := range doc.CodeBlocks() {
		target := b.File
		if target == "" {
			target = defaultFile
		}
		if target == "" {
			continue
		}
		var resolved string
		var err error
		if b.Name != "" {
			resolved, err = Resolve(doc, b.Name)
		} else {
			resolved = b.Source
		}
		if err != nil {
			return nil, err
		}
		if existing, ok := out[target]; ok && existing != "" {
			out[target] = existing + "\n" + resolved
		} else {
			out[target] = resolved
		}
		// ensure trailing newline
		if !strings.HasSuffix(out[target], "\n") {
			out[target] += "\n"
		}
	}
	return out, nil
}
