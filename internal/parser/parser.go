// Package parser parses .loom documents (Markdown-like with fenced code blocks).
package parser

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/HimanshuSardana/loom/internal/ast"
	"github.com/HimanshuSardana/loom/internal/diagnostics"
)

var (
	fenceOpenRe = regexp.MustCompile("^```(\\S*)\\s*(.*)$")
	nameRe      = regexp.MustCompile(`\{#([A-Za-z][A-Za-z0-9_.\-]*)\}`)
	optRe       = regexp.MustCompile(`\{([^}]*)\}`)
	validName   = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_.\-]*$`)
)

type Result struct {
	Doc         *ast.Document
	Diagnostics diagnostics.List
}

// ParseFile reads and parses a .loom file.
func ParseFile(path string) (*Result, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Parse(path, string(data)), nil
}

// Parse parses source text.
func Parse(filename, src string) *Result {
	r := &Result{Doc: &ast.Document{File: filename}}
	lines := strings.Split(src, "\n")
	n := len(lines)
	i := 0 // 0-indexed

	pos := func(line int, col int) ast.Position { return ast.Position{Line: line, Column: col} }
	loc := func(sLine, sCol, eLine, eCol int) ast.Location {
		return ast.Location{File: filename, Start: pos(sLine, sCol), End: pos(eLine, eCol)}
	}

	seen := map[string]int{} // block name -> line

	paraBuf := []string{}
	paraStart := 0
	flushPara := func(endLine int) {
		if len(paraBuf) == 0 {
			return
		}
		text := strings.Join(paraBuf, "\n")
		if strings.TrimSpace(text) == "" {
			paraBuf = nil
			return
		}
		r.Doc.Blocks = append(r.Doc.Blocks, ast.Block{
			Type: ast.BlockParagraph, Text: text,
			Loc: loc(paraStart, 1, endLine, 1),
		})
		paraBuf = nil
	}

	listBuf := []string{}
	listStart := 0
	flushList := func(endLine int) {
		if len(listBuf) == 0 {
			return
		}
		r.Doc.Blocks = append(r.Doc.Blocks, ast.Block{
			Type: ast.BlockList, Items: append([]string(nil), listBuf...),
			Loc: loc(listStart, 1, endLine, 1),
		})
		listBuf = nil
	}

	for i < n {
		line := lines[i]
		lineNo := i + 1
		trimmed := strings.TrimSpace(line)

		// Fenced code block?
		if m := fenceOpenRe.FindStringSubmatch(line); m != nil && strings.HasPrefix(trimmed, "```") {
			// Only treat as fence if opening ``` at line start (allow indent? keep simple: trimmed starts with ```)
			flushPara(lineNo)
			flushList(lineNo)
			lang := m[1]
			rest := m[2]
			codeStart := lineNo
			var codeLines []string
			j := i + 1
			closed := false
			for j < n {
				if strings.TrimSpace(lines[j]) == "```" || strings.HasPrefix(strings.TrimSpace(lines[j]), "```") && strings.TrimSpace(strings.TrimSpace(lines[j])) == "```" {
					closed = true
					break
				}
				// closing fence: line whose trimmed == ``` exactly
				if strings.TrimSpace(lines[j]) == "```" {
					closed = true
					break
				}
				codeLines = append(codeLines, lines[j])
				j++
			}
			// handle closing detection: above loop breaks at j pointing to closing or n
			if j < n && strings.TrimSpace(lines[j]) == "```" {
				closed = true
			}
			if !closed {
				r.Diagnostics = append(r.Diagnostics, diagnostics.Diagnostic{
					File: filename, Line: codeStart, Column: 1,
					Severity: diagnostics.SeverityError, Message: "unclosed code fence",
				})
				i = j
				continue
			}
			endLine := j + 1
			b := ast.Block{
				Type: ast.BlockCode, Language: lang,
				Source:  strings.Join(codeLines, "\n"),
				Options: map[string]string{}, Flags: map[string]bool{},
				Loc: loc(codeStart, 1, endLine, 3),
			}
			// Parse {#name} and {options} from rest
			if nm := nameRe.FindStringSubmatch(rest); nm != nil {
				b.Name = nm[1]
			}
			for _, om := range optRe.FindAllStringSubmatch(rest, -1) {
				inner := strings.TrimSpace(om[1])
				if inner == "" {
					continue
				}
				if strings.HasPrefix(inner, "#") {
					continue // name, already handled
				}
				if strings.Contains(inner, "=") {
					parts := strings.SplitN(inner, "=", 2)
					k := strings.TrimSpace(parts[0])
					v := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
					// depends may be comma list; keep raw
					b.Options[k] = v
				} else {
					b.Flags[inner] = true
				}
			}
			// Derive conveniences
			if _, ok := b.Flags["run"]; ok {
				b.Run = true
			}
			if s, ok := b.Options["session"]; ok {
				b.Session = s
			} else if b.Run && b.Language != "" {
				b.Session = "default:" + b.Language
			}
			if f, ok := b.Options["file"]; ok {
				b.File = f
			}
			if _, ok := b.Flags["isolated"]; ok {
				b.Isolated = true
			}
			if d, ok := b.Options["depends"]; ok && d != "" {
				for _, dep := range strings.Split(d, ",") {
					dep = strings.TrimSpace(dep)
					if dep != "" {
						b.Depends = append(b.Depends, dep)
					}
				}
			}
			// Validation
			if b.Language == "" {
				r.Diagnostics = append(r.Diagnostics, diagnostics.Diagnostic{
					File: filename, Line: codeStart, Column: 1, Block: b.Name,
					Severity: diagnostics.SeverityError, Message: "code block missing language",
				})
			}
			if b.Name != "" {
				if !validName.MatchString(b.Name) {
					r.Diagnostics = append(r.Diagnostics, diagnostics.Diagnostic{
						File: filename, Line: codeStart, Column: 1, Block: b.Name,
						Severity: diagnostics.SeverityError, Message: fmt.Sprintf("invalid block name '%s'", b.Name),
					})
				} else if prev, dup := seen[b.Name]; dup {
					r.Diagnostics = append(r.Diagnostics, diagnostics.Diagnostic{
						File: filename, Line: codeStart, Column: 1, Block: b.Name,
						Severity: diagnostics.SeverityError,
						Message:  fmt.Sprintf("duplicate block name '%s' (first defined at line %d)", b.Name, prev),
					})
				} else {
					seen[b.Name] = codeStart
				}
			}
			r.Doc.Blocks = append(r.Doc.Blocks, b)
			i = j + 1
			continue
		}

		// Heading
		if strings.HasPrefix(trimmed, "#") {
			flushPara(lineNo)
			flushList(lineNo)
			level := 0
			for level < len(trimmed) && trimmed[level] == '#' {
				level++
			}
			text := strings.TrimSpace(trimmed[level:])
			if r.Doc.Title == "" && level == 1 {
				r.Doc.Title = text
			}
			r.Doc.Blocks = append(r.Doc.Blocks, ast.Block{
				Type: ast.BlockHeading, Level: level, Text: text,
				Loc: loc(lineNo, 1, lineNo, len(line)+1),
			})
			i++
			continue
		}

		// List item
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
			flushPara(lineNo)
			if len(listBuf) == 0 {
				listStart = lineNo
			}
			listBuf = append(listBuf, strings.TrimSpace(trimmed[2:]))
			i++
			continue
		}

		// Blank line flushes paragraph/list boundaries
		if trimmed == "" {
			flushPara(lineNo)
			flushList(lineNo)
			i++
			continue
		}

		flushList(lineNo)
		if len(paraBuf) == 0 {
			paraStart = lineNo
		}
		paraBuf = append(paraBuf, line)
		i++
	}
	flushPara(n)
	flushList(n)
	return r
}
