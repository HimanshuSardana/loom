package typst

import (
	"strings"

	"github.com/HimanshuSardana/loom/internal/ast"
	"github.com/HimanshuSardana/loom/internal/runtime"
)

func quote(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return "\"" + s + "\""
}

// Render generates Typst markup.
func Render(doc *ast.Document, results map[string]runtime.Result) string {
	var sb strings.Builder
	title := doc.Title
	if title == "" {
		title = "Loom Document"
	}
	sb.WriteString("#set document(title: " + quote(title) + ")\n")
	sb.WriteString("#set page(margin: 2cm)\n")
	sb.WriteString("= " + typstEsc(title) + "\n\n")
	for _, b := range doc.Blocks {
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
			sb.WriteString("#block(fill: luma(245), inset: 8pt, radius: 4pt)[\n")
			sb.WriteString("#text(size: 8pt, fill: gray)[" + typstEsc(b.Language+blockSuffix(&b)) + "]\n")
			sb.WriteString("#raw(" + quote(b.Source) + ", lang: " + quote(b.Language) + ")\n]\n\n")
			if r, ok := results[b.Name]; ok && b.Name != "" {
				sb.WriteString("#block(fill: green.lighten(90%), inset: 8pt, radius: 4pt)[\n*Output*\n")
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

func blockSuffix(b *ast.Block) string {
	if b.Name != "" {
		return " {# " + b.Name + "}"
	}
	return ""
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
