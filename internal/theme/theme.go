// Package theme provides named visual themes shared by the HTML and
// Typst/PDF exporters. A Theme carries a full HTML stylesheet plus a
// Typst prelude and palette so both outputs stay in sync.
package theme

import (
	"strings"
)

// Theme describes one named visual style.
type Theme struct {
	Name        string
	Description string
	// CSS is the full standalone stylesheet for HTML export.
	CSS string
	// TypstPrelude is emitted at the top of the Typst document
	// (page setup, fonts, heading styles). The title is available
	// as {{TITLE}} and replaced at render time.
	TypstPrelude string
	// Typst palette expressions used in the document body.
	TypstCodeFill   string
	TypstOutputFill string
	TypstAccent     string
}

const Default = "modern"

func Get(name string) Theme {
	if t, ok := byName[strings.ToLower(strings.TrimSpace(name))]; ok {
		return t
	}
	return byName[Default]
}

func List() []Theme { return []Theme{byName["tufte"], byName["dark"], byName["modern"]} }

// Valid reports whether name is a known theme.
func Valid(name string) bool {
	_, ok := byName[strings.ToLower(strings.TrimSpace(name))]
	return ok
}

func Names() string {
	var n []string
	for _, t := range List() {
		n = append(n, t.Name)
	}
	return strings.Join(n, ", ")
}

var byName = map[string]Theme{
	"tufte": {
		Name:        "tufte",
		Description: "Tufte-style essay: warm paper, serif body, brick-red accents, wide margin",
		CSS: `body{font-family:"ETBook","Palatino Linotype",Palatino,Georgia,serif;background:#fffff8;color:#111;max-width:760px;margin:2.5rem auto;padding:0 3rem 0 1rem;line-height:1.7;font-size:19px}h1,h2,h3{font-weight:400;line-height:1.2}h1{font-size:2.2rem;margin-bottom:.2rem}h2{font-style:italic;font-size:1.6rem;border-bottom:1px solid #ddd;padding-bottom:.3rem}nav.toc{font-size:.9rem;border-top:1px solid #ccc;border-bottom:1px solid #ccc;padding:.5rem 0;background:none;border-left:none;border-right:none;border-radius:0}nav.toc a{color:#a00000;text-decoration:none}pre{font-size:15px;background:#f8f5ec;padding:1rem;overflow:auto;border-left:2px solid #a00000;border-radius:0}code{font-family:"DejaVu Sans Mono",ui-monospace,monospace;font-size:.85em}.output{background:none;border-top:1px solid #a00000;border-bottom:1px solid #a00000;border-left:none;border-right:none;border-radius:0;margin:1rem 0;padding:.5rem 0;font-size:.92rem}.output strong{font-variant:small-caps;letter-spacing:.06em;color:#a00000}.block-head{font-size:.8rem;color:#716c5e;font-variant:small-caps;letter-spacing:.08em}img{max-width:100%;border:1px solid #ddd}a{color:#a00000}
section[id]{scroll-margin-top:1rem}`,
		TypstPrelude: `#set document(title: {{TITLE}})
#set page(paper: "a4", margin: (left: 2cm, right: 4.5cm, top: 2.5cm, bottom: 2.5cm))
#set text(font: ("Libertinus Serif", "FreeSerif"), size: 11pt, fill: rgb("#111111"))
#set heading(numbering: none)
#show heading.where(level: 1): it => [#set text(size: 22pt, weight: "regular"); #v(0.5em); #it.body; #v(0.2em)]
#show heading.where(level: 2): it => [#set text(size: 15pt, style: "italic", weight: "regular", fill: rgb("#a00000")); #it.body]
#show raw: set text(size: 9.5pt)
#set par(justify: true, leading: 0.7em)`,
		TypstCodeFill:   `rgb("#f8f5ec")`,
		TypstOutputFill: `rgb("#fdf8ee")`,
		TypstAccent:     `rgb("#a00000")`,
	},
	"dark": {
		Name:        "dark",
		Description: "Dark monospace theme using Iosevka: near-black background, terminal feel",
		CSS:         `body{font-family:Iosevka,"Iosevka Term","JetBrains Mono",Menlo,Consolas,monospace;background:#121214;color:#d6d6d6;max-width:860px;margin:0 auto;padding:2rem 1.5rem;line-height:1.65;font-size:16px}h1,h2,h3{color:#e8e8e8;font-weight:600}h1{border-bottom:1px solid #333;padding-bottom:.4rem}a{color:#7dd3a8}nav.toc{background:#1a1b1e;border:1px solid #2c2d31;border-radius:8px;padding:1rem}nav.toc a{color:#7dd3a8;text-decoration:none}pre{background:#1a1b1e;border:1px solid #2c2d31;padding:1rem;overflow:auto;border-radius:8px;color:#d6d6d6}code{font-family:inherit}.output{background:#141b16;border:1px solid #2f4a3a;color:#c9e8d4;padding:.75rem 1rem;border-radius:8px;margin:.5rem 0}.output pre{background:none;border:none;padding:0;margin:.25rem 0;color:inherit}.output strong{color:#7dd3a8}.block-head{font-size:.8rem;color:#8a8a93}.stderr{color:#ff7b72}img{max-width:100%;border:1px solid #2c2d31;border-radius:8px}`,
		TypstPrelude: `#set document(title: {{TITLE}})
#set page(paper: "a4", margin: 2cm, fill: rgb("#121214"))
#set text(font: ("Iosevka NFM", "Iosevka NF", "DejaVu Sans Mono"), size: 10.5pt, fill: rgb("#d6d6d6"))
#set heading(numbering: none)
#show heading: set text(fill: rgb("#e8e8e8"))
#show raw: set text(size: 9.5pt)
#show link: set text(fill: rgb("#7dd3a8"))`,
		TypstCodeFill:   `rgb("#1a1b1e")`,
		TypstOutputFill: `rgb("#141b16")`,
		TypstAccent:     `rgb("#7dd3a8")`,
	},
	"modern": {
		Name:        "modern",
		Description: "Modern clean light theme: airy sans-serif, indigo accents",
		CSS:         `body{font-family:system-ui,-apple-system,"Segoe UI",sans-serif;max-width:800px;margin:2rem auto;padding:0 1.5rem;line-height:1.7;color:#1e293b;background:#fff}h1{font-size:2rem;letter-spacing:-.02em}h2{letter-spacing:-.01em;border-bottom:2px solid #e0e7ff;padding-bottom:.3rem}nav.toc{background:#f8faff;border:1px solid #e0e7ff;padding:1rem 1.25rem;border-radius:12px}nav.toc a{color:#4f46e5;text-decoration:none}pre{background:#f6f7f9;border:1px solid #e5e7eb;padding:1rem;overflow:auto;border-radius:12px}code{font-family:ui-monospace,"SF Mono",monospace;font-size:.88em}.output{background:#f0fdf4;border:1px solid #bbf7d0;padding:.75rem 1rem;border-radius:12px;margin:.5rem 0}.output pre{background:none;border:none;padding:0;margin:.25rem 0}.output strong{color:#15803d}.block-head{font-size:.8rem;color:#64748b;text-transform:uppercase;letter-spacing:.05em}img{max-width:100%;border-radius:12px}.stderr{color:#b91c1c}a{color:#4f46e5}`,
		TypstPrelude: `#set document(title: {{TITLE}})
#set page(paper: "a4", margin: 2cm)
#set text(font: ("Adwaita Sans", "FreeSans"), size: 10.5pt, fill: rgb("#1e293b"))
#set heading(numbering: none)
#show heading: set text(fill: rgb("#312e81"))
#show heading.where(level: 1): it => [#set text(size: 22pt); #it.body; #v(0.2em, weak: true)]
#show raw: set text(size: 9.5pt)
#show link: set text(fill: rgb("#4f46e5"))`,
		TypstCodeFill:   `rgb("#f6f7f9")`,
		TypstOutputFill: `rgb("#f0fdf4")`,
		TypstAccent:     `rgb("#4f46e5")`,
	},
}
