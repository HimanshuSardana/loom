package typst

import (
	"strings"

	"github.com/HimanshuSardana/loom/internal/ast"
	"github.com/HimanshuSardana/loom/internal/runtime"
	"github.com/HimanshuSardana/loom/internal/theme"
)

func quote(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return "\"" + s + "\""
}

// Render generates Typst markup. themeName selects a named theme
// (see internal/theme); "" means default.
func Render(doc *ast.Document, results map[string]runtime.Result, themeName string) string {
	th := theme.Get(themeName)
	var sb strings.Builder
	title := doc.Title
	if title == "" {
		title = "Loom Document"
	}
	sb.WriteString(strings.ReplaceAll(th.TypstPrelude, "{{TITLE}}", quote(title)) + "\n")
	sb.WriteString("= " + typstEsc(title) + "\n\n")
	blocks := doc.Blocks
	// The document title already renders as the top-level heading; a
	// leading H1 with identical text would print the title twice.
	if len(blocks) > 0 && blocks[0].Type == ast.BlockHeading && blocks[0].Level == 1 && blocks[0].Text == doc.Title {
		blocks = blocks[1:]
	}
	for _, b := range blocks {
		switch b.Type {
		case ast.BlockHeading:
			sb.WriteString(strings.Repeat("=", b.Level) + " " + typstEsc(b.Text) + "\n\n")
		case ast.BlockParagraph:
			sb.WriteString(typstEsc(b.Text) + "\n\n")
		case ast.BlockList:
			for _, it := range b.Items {
				sb.WriteString("- " + typstEsc(it) + "\n")
			}
			sb.WriteString("\n")
		case ast.BlockCode:
			sb.WriteString("#block(width: 100%, fill: " + th.TypstCodeFill + ", inset: 8pt, radius: 4pt)[\n")
			sb.WriteString("#align(right)[#text(size: 8pt, weight: \"bold\", fill: " + th.TypstAccent + ")[" + typstEsc(strings.ToUpper(b.Language)) + "]]\n")
			sb.WriteString("#raw(" + quote(b.Source) + ", lang: " + quote(b.Language) + ", block: true)\n]\n\n")
			if r, ok := results[b.Name]; ok && b.Name != "" {
				sb.WriteString("#block(width: 100%, fill: " + th.TypstOutputFill + ", inset: 8pt, radius: 4pt)[\n*Output* \\\n")
				if r.Stdout != "" {
					sb.WriteString("#raw(" + quote(r.Stdout) + ")\n")
				}
				if r.Value != "" {
					sb.WriteString("#raw(" + quote(r.Value) + ")\n")
				}
				for _, a := range r.Artifacts {
					if isImage(a) {
						sb.WriteString("#image(" + quote(a) + ")\n")
					}
				}
				sb.WriteString("]\n\n")
			}
		}
	}
	return sb.String()
}

func typstEsc(s string) string {
	// escape typst markup chars minimally: use raw inline for safety? keep plain
	r := strings.ReplaceAll(s, "\\", "\\\\")
	r = strings.ReplaceAll(r, "#", "\\#")
	return r
}

func isImage(p string) bool {
	l := strings.ToLower(p)
	return strings.HasSuffix(l, ".png") || strings.HasSuffix(l, ".jpg") || strings.HasSuffix(l, ".jpeg") || strings.HasSuffix(l, ".svg") || strings.HasSuffix(l, ".gif") || strings.HasSuffix(l, ".webp")
}
