package html

import (
	"fmt"
	"html"
	"strings"

	"github.com/HimanshuSardana/loom/internal/ast"
	"github.com/HimanshuSardana/loom/internal/runtime"
)

func esc(s string) string { return html.EscapeString(s) }

// Render builds a standalone HTML page. results maps block name -> result.
func Render(doc *ast.Document, results map[string]runtime.Result) string {
	var sb strings.Builder
	title := doc.Title
	if title == "" {
		title = "Loom Document"
	}
	sb.WriteString("<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n<meta charset=\"utf-8\">\n<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n<title>" + esc(title) + "</title>\n<style>\n")
	sb.WriteString("body{font-family:system-ui,sans-serif;max-width:800px;margin:2rem auto;padding:0 1rem;line-height:1.6;color:#1a1a1a}pre{background:#f5f5f5;padding:1rem;overflow:auto;border-radius:8px}.output{background:#f0fdf4;border:1px solid #bbf7d0;padding:.75rem 1rem;border-radius:8px;margin:.5rem 0}.output pre{background:none;padding:0;margin:.25rem 0}nav.toc{background:#fafafa;border:1px solid #eee;padding:1rem;border-radius:8px}code{font-family:ui-monospace,monospace}img{max-width:100%}.block-head{font-size:.8rem;color:#555}.stderr{color:#b91c1c}\n")
	sb.WriteString("</style>\n</head>\n<body>\n<article>\n")
	sb.WriteString("<h1>" + esc(title) + "</h1>\n")
	// TOC
	sb.WriteString("<nav class=\"toc\"><strong>Contents</strong><ul>\n")
	for _, b := range doc.Blocks {
		if b.Type == ast.BlockHeading && b.Level <= 3 {
			anchor := anchorOf(b.Text)
			fmt.Fprintf(&sb, "<li><a href=\"#%s\">%s</a></li>\n", anchor, esc(b.Text))
		}
	}
	sb.WriteString("</ul></nav>\n")
	for _, b := range doc.Blocks {
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
			fmt.Fprintf(&sb, "<section id=\"block-%s\">\n<div class=\"block-head\">%s", esc(id), esc(b.Language))
			if b.Name != "" {
				fmt.Fprintf(&sb, " {#%s}", esc(b.Name))
			}
			sb.WriteString("</div>\n<pre><code class=\"language-" + esc(b.Language) + "\">" + esc(b.Source) + "</code></pre>\n")
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
