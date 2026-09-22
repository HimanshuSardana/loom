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

type tocEntry struct {
	anchor string
	text   string
	level  int
}

// scrollspyJS highlights the TOC entry for the heading nearest the viewport top.
const scrollspyJS = `<script>
(function () {
  var links = Array.prototype.slice.call(document.querySelectorAll('.toc a[href^="#"]'));
  if (!links.length) return;
  var heads = links
    .map(function (a) { return document.getElementById(a.getAttribute('href').slice(1)); })
    .filter(Boolean);
  function update() {
    var y = window.scrollY + 140;
    var cur = heads[0];
    for (var i = 0; i < heads.length; i++) {
      if (heads[i].getBoundingClientRect().top + window.scrollY <= y) cur = heads[i];
    }
    links.forEach(function (a) {
      a.classList.toggle('active', cur && a.getAttribute('href') === '#' + cur.id);
    });
  }
  window.addEventListener('scroll', update, { passive: true });
  window.addEventListener('hashchange', update);
  update();
})();
</script>
`

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
	sb.WriteString("</style>\n</head>\n<body class=\"theme-" + esc(th.Name) + "\">\n")

	blocks := doc.Blocks
	// The document title already renders as the page H1; a leading H1
	// with identical text would print the title twice.
	if len(blocks) > 0 && blocks[0].Type == ast.BlockHeading && blocks[0].Level == 1 && blocks[0].Text == doc.Title {
		blocks = blocks[1:]
	}

	// Collect TOC entries first so we know whether a sidebar is needed.
	var toc []tocEntry
	for _, b := range blocks {
		if b.Type == ast.BlockHeading && b.Level <= 3 {
			toc = append(toc, tocEntry{anchor: anchorOf(b.Text), text: b.Text, level: b.Level})
		}
	}

	pageClass := "page"
	if len(toc) == 0 {
		pageClass = "page no-toc"
	}
	sb.WriteString("<div class=\"" + pageClass + "\">\n<main class=\"content\">\n<article>\n")
	sb.WriteString("<h1>" + esc(title) + "</h1>\n")

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
			fmt.Fprintf(&sb, "<section id=\"block-%s\">\n<div class=\"block-head\">%s</div>\n<pre><code class=\"language-%s\">%s</code></pre>\n",
				esc(id), esc(b.Language), esc(b.Language), highlight(b.Source, b.Language))
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
	sb.WriteString("</article>\n</main>\n")

	if len(toc) > 0 {
		sb.WriteString("<aside class=\"sidebar\">\n<nav class=\"toc\">\n<div class=\"toc-title\">Contents</div>\n<ul>\n")
		for _, e := range toc {
			fmt.Fprintf(&sb, "<li class=\"l%d\"><a href=\"#%s\">%s</a></li>\n", e.level, e.anchor, esc(e.text))
		}
		sb.WriteString("</ul>\n</nav>\n</aside>\n")
	}

	sb.WriteString("</div>\n")
	if len(toc) > 0 {
		sb.WriteString(scrollspyJS)
	}
	sb.WriteString("</body>\n</html>\n")
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
