package html

import (
	"fmt"
	"html"
	"strings"

	"github.com/HimanshuSardana/loom/internal/ast"
	"github.com/HimanshuSardana/loom/internal/runtime"
	"github.com/HimanshuSardana/loom/internal/theme"
)

func esc(s string) string { return html.EscapeString(s) }

// Render builds a standalone HTML page. results maps block name -> result.
// themeName selects a named theme (see internal/theme); "" means default.
func Render(doc *ast.Document, results map[string]runtime.Result, themeName string) string {
	th := theme.Get(themeName)
	var sb strings.Builder
	title := doc.Title
	if title == "" {
		title = "Loom Document"
	}
	sb.WriteString("<!DOCTYPE html>\n<html lang=\"en\" data-theme=\"" + esc(th.Name) + "\">\n<head>\n<meta charset=\"utf-8\">\n<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n<title>" + esc(title) + "</title>\n<style>\n")
	sb.WriteString(th.CSS + "\n")
	sb.WriteString("</style>\n</head>\n<body class=\"theme-" + esc(th.Name) + "\">\n<article>\n")
	sb.WriteString("<h1>" + esc(title) + "</h1>\n")
	// TOC
	blocks := doc.Blocks
	// The document title already renders as the page H1; a leading H1
	// with identical text would print the title twice.
	if len(blocks) > 0 && blocks[0].Type == ast.BlockHeading && blocks[0].Level == 1 && blocks[0].Text == doc.Title {
		blocks = blocks[1:]
	}
	sb.WriteString("<nav class=\"toc\"><strong>Contents</strong><ul>\n")
	for _, b := range blocks {
		if b.Type == ast.BlockHeading && b.Level <= 3 {
			anchor := anchorOf(b.Text)
			fmt.Fprintf(&sb, "<li><a href=\"#%s\">%s</a></li>\n", anchor, esc(b.Text))
		}
	}
	sb.WriteString("</ul></nav>\n")
	for _, b := range blocks {
		switch b.Type {
		case ast.BlockHeading:
			anchor := anchorOf(b.Text)
			fmt.Fprintf(&sb, "<h%d id=\"%s\">%s</h%d>\n", b.Level, anchor, esc(b.Text), b.Level)
		case ast.BlockParagraph:
			sb.WriteString("<p>" + esc(b.Text) + "</p>\n")
		case ast.BlockList:
			sb.WriteString("<ul>\n")
			for _, it := range b.Items {
				sb.WriteString("<li>" + esc(it) + "</li>\n")
			}
			sb.WriteString("</ul>\n")
		case ast.BlockCode:
			id := b.Name
			if id == "" {
				id = fmt.Sprintf("block-%d-%d", b.Loc.Start.Line, b.Loc.Start.Column)
			}
			fmt.Fprintf(&sb, "<section id=\"block-%s\">\n<div class=\"block-head\">%s</div>\n<pre><code class=\"language-"+esc(b.Language)+"\">"+esc(b.Source)+"</code></pre>\n", esc(id), esc(b.Language))
			// references
			// execution output
			if r, ok := results[b.Name]; ok && b.Name != "" {
				sb.WriteString("<div class=\"output\">\n<strong>Output</strong>\n")
				if r.Stdout != "" {
					sb.WriteString("<pre>" + esc(r.Stdout) + "</pre>\n")
				}
				if r.Value != "" {
					sb.WriteString("<pre class=\"value\">" + esc(r.Value) + "</pre>\n")
				}
				if r.Stderr != "" {
					sb.WriteString("<pre class=\"stderr\">" + esc(r.Stderr) + "</pre>\n")
				}
				for _, a := range r.Artifacts {
					if isImage(a) {
						fmt.Fprintf(&sb, "<img src=\"%s\" alt=\"%s\">\n", esc(a), esc(a))
					} else if a != "" {
						fmt.Fprintf(&sb, "<div>artifact: <code>%s</code></div>\n", esc(a))
					}
				}
				sb.WriteString("</div>\n")
			}
			sb.WriteString("</section>\n")
		}
	}
	sb.WriteString("</article>\n</body>\n</html>\n")
	return sb.String()
}

func anchorOf(s string) string {
	s = strings.ToLower(s)
	var out strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			out.WriteRune(r)
		} else if r == ' ' || r == '-' || r == '_' {
			out.WriteRune('-')
		}
	}
	return out.String()
}

func isImage(p string) bool {
	l := strings.ToLower(p)
	return strings.HasSuffix(l, ".png") || strings.HasSuffix(l, ".jpg") || strings.HasSuffix(l, ".jpeg") || strings.HasSuffix(l, ".svg") || strings.HasSuffix(l, ".gif") || strings.HasSuffix(l, ".webp")
}
