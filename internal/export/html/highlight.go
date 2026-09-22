package html

import (
	"strings"
)

type langKind int

const (
	langGeneric langKind = iota
	langPython
	langShell
	langJS
)

func kindFor(lang string) langKind {
	switch strings.ToLower(lang) {
	case "python", "py":
		return langPython
	case "bash", "sh", "shell":
		return langShell
	case "javascript", "js", "node":
		return langJS
	}
	return langGeneric
}

func set(words string) map[string]bool {
	m := map[string]bool{}
	for _, w := range strings.Fields(words) {
		m[w] = true
	}
	return m
}

var kwSets = map[langKind]map[string]bool{
	langPython:  set("and as assert async await break class continue def del elif else except finally for from global if import in is lambda nonlocal not or pass raise return try while with yield True False None self"),
	langJS:      set("var let const function return if else for while do break continue new class extends super this typeof instanceof in of null undefined true false try catch finally throw switch case default import export from async await yield delete void static"),
	langShell:   set("if then else elif fi for while do done case esac function in select until time coproc local export return exit echo printf read set unset trap shift eval exec source alias readonly declare"),
	langGeneric: set(""),
}

func isDigit(c byte) bool  { return c >= '0' && c <= '9' }
func isHex(c byte) bool    { return isDigit(c) || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F') }
func isLetter(c byte) bool { return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') }
func isIdent(c byte) bool  { return isLetter(c) || isDigit(c) || c == '_' }

// highlight returns HTML-escaped source with syntax spans (tok-k/s/n/c/f).
// Stripping all tags yields exactly esc(src), so no source text is lost.
func highlight(src, lang string) string {
	k := kindFor(lang)
	kw := kwSets[k]
	var sb strings.Builder
	i, n := 0, len(src)

	for i < n {
		c := src[i]

		// Comments
		if c == '#' && k != langJS {
			j := strings.IndexByte(src[i:], '\n')
			if j < 0 {
				j = n
			} else {
				j += i
			}
			writeTok(&sb, "c", src[i:j])
			i = j
			continue
		}
		if c == '/' && i+1 < n && src[i+1] == '/' && (k == langJS || k == langGeneric) {
			j := strings.IndexByte(src[i:], '\n')
			if j < 0 {
				j = n
			} else {
				j += i
			}
			writeTok(&sb, "c", src[i:j])
			i = j
			continue
		}

		// Triple-quoted strings (python)
		if k == langPython && (strings.HasPrefix(src[i:], `"""`) || strings.HasPrefix(src[i:], `'''`)) {
			q := src[i : i+3]
			end := strings.Index(src[i+3:], q)
			var j int
			if end < 0 {
				j = n
			} else {
				j = i + 3 + end + 3
			}
			writeTok(&sb, "s", src[i:j])
			i = j
			continue
		}

		// Strings
		if c == '"' || c == '\'' || c == '`' {
			j := i + 1
			for j < n {
				if src[j] == '\\' {
					j += 2
					continue
				}
				if src[j] == c {
					j++
					break
				}
				if src[j] == '\n' && c != '`' {
					break // unterminated single-line string
				}
				j++
			}
			if j > n {
				j = n
			}
			writeTok(&sb, "s", src[i:j])
			i = j
			continue
		}

		// Numbers
		if isDigit(c) {
			j := i
			if c == '0' && i+1 < n && (src[i+1] == 'x' || src[i+1] == 'X' || src[i+1] == 'b' || src[i+1] == 'B') {
				j = i + 2
				for j < n && (isHex(src[j]) || src[j] == '_') {
					j++
				}
			} else {
				for j < n && (isDigit(src[j]) || src[j] == '.' || src[j] == '_') {
					j++
				}
				if j < n && (src[j] == 'e' || src[j] == 'E') {
					m := j + 1
					if m < n && (src[m] == '+' || src[m] == '-') {
						m++
					}
					if m < n && isDigit(src[m]) {
						for m < n && isDigit(src[m]) {
							m++
						}
						j = m
					}
				}
			}
			writeTok(&sb, "n", src[i:j])
			i = j
			continue
		}

		// Identifiers: keyword or function call
		if isLetter(c) || c == '_' {
			j := i
			for j < n && isIdent(src[j]) {
				j++
			}
			word := src[i:j]
			if kw[word] {
				writeTok(&sb, "k", word)
			} else {
				m := j
				for m < n && (src[m] == ' ' || src[m] == '\t') {
					m++
				}
				if m < n && src[m] == '(' {
					writeTok(&sb, "f", word)
				} else {
					writeTok(&sb, "", word)
				}
			}
			i = j
			continue
		}

		sb.WriteString(esc(string(c)))
		i++
	}
	return sb.String()
}

func writeTok(sb *strings.Builder, kind, text string) {
	if kind == "" {
		sb.WriteString(esc(text))
		return
	}
	sb.WriteString(`<span class="tok-` + kind + `">`)
	sb.WriteString(esc(text))
	sb.WriteString(`</span>`)
}
