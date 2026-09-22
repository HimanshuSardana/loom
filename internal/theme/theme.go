// Package theme provides named visual themes shared by the HTML and
// Typst/PDF exporters. A Theme carries a full HTML stylesheet plus a
// Typst prelude and palette so both outputs stay in sync.
package theme

import (
	"fmt"
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

func List() []Theme {
	return []Theme{
		byName["tufte"], byName["paper"], byName["solarized"], byName["modern"],
		byName["dark"], byName["dracula"], byName["nord"],
	}
}

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

// layoutColors are the per-theme knobs of the shared HTML layout.
type layoutColors struct {
	faint, hover, active, text, muted string
	kw, str, num, com, fn             string
	border                            string
}

// layoutCSS is the shared HTML skeleton: w-7xl centered grid with a
// sticky right-margin TOC (scrollspy, left-border indicator), Iosevka
// for code, right-aligned language tags, and syntax token colors.
func layoutCSS(c layoutColors) string {
	return fmt.Sprintf(`.page{max-width:80rem;margin:0 auto;padding:2rem 1.5rem;display:grid;grid-template-columns:minmax(0,1fr) 15rem;gap:3rem;align-items:start}
.page.no-toc{grid-template-columns:minmax(0,1fr)}
.content{min-width:0}
.sidebar{position:sticky;top:2rem;max-height:calc(100vh - 4rem);overflow:auto}
.toc-title{font-size:.72rem;text-transform:uppercase;letter-spacing:.1em;color:%[1]s;margin-bottom:.6rem}
.toc ul{list-style:none;margin:0;padding:0}
.toc li a{display:block;padding:.28rem .8rem;font-size:.85rem;line-height:1.4;color:%[2]s;text-decoration:none;border-left:2px solid %[3]s;transition:border-color .15s,color .15s}
.toc li a:hover{color:%[4]s}
.toc li a.active{color:%[5]s;border-left-color:%[5]s;font-weight:600}
.toc li.l3 a{padding-left:1.6rem}
h1[id],h2[id],h3[id],section[id]{scroll-margin-top:1.5rem}
code,pre{font-family:"Iosevka","Iosevka Term","Iosevka NF","JetBrains Mono","DejaVu Sans Mono",ui-monospace,SFMono-Regular,Menlo,monospace}
pre{overflow:auto;border-left:none}
.block-head{text-align:right;font-size:.72rem;letter-spacing:.1em;margin-bottom:.4rem}
.tok-k{color:%[6]s}
.tok-s{color:%[7]s}
.tok-n{color:%[8]s}
.tok-c{color:%[9]s;font-style:italic}
.tok-f{color:%[10]s}
@media (max-width:1024px){
.page{grid-template-columns:minmax(0,1fr);gap:1.5rem}
.sidebar{position:static;max-height:none;order:2;border-top:1px solid %[3]s;padding-top:1rem}
}`,
		c.muted, c.text, c.faint, c.hover, c.active,
		c.kw, c.str, c.num, c.com, c.fn)
}

var byName = map[string]Theme{
	"tufte": {
		Name:        "tufte",
		Description: "Tufte-style essay: warm paper, serif body, brick-red accents, wide margin",
		CSS: layoutCSS(layoutColors{
			faint: "#e3ded2", hover: "#a00000", active: "#a00000",
			text: "#5c574a", muted: "#716c5e", border: "#e3ded2",
			kw: "#a00000", str: "#4a7c59", num: "#9a5b2f", com: "#99907f", fn: "#3b5ba5",
		}) + `
body{font-family:"ETBook","Palatino Linotype",Palatino,Georgia,serif;background:#fffff8;color:#111;line-height:1.7;font-size:19px}
h1,h2,h3{font-weight:400;line-height:1.2}h1{font-size:2.2rem;margin:.2rem 0 .8rem}
h2{font-style:italic;font-size:1.6rem;border-bottom:1px solid #ddd;padding-bottom:.3rem}
pre{font-size:14.5px;background:#f8f5ec;padding:1rem;border-radius:6px}
code{font-size:.85em}
.output{background:#fdf8ee;border-top:1px solid #a00000;border-bottom:1px solid #a00000;margin:1rem 0;padding:.5rem 0;font-size:.92rem}
.output pre{background:none;border:none;padding:0;margin:.25rem 0}
.output strong{font-variant:small-caps;letter-spacing:.06em;color:#a00000}
.block-head{color:#716c5e;font-variant:small-caps;letter-spacing:.1em}
img{max-width:100%;border:1px solid #ddd}a{color:#a00000}`,
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
		CSS: layoutCSS(layoutColors{
			faint: "#2c2d31", hover: "#e8e8e8", active: "#7dd3a8",
			text: "#9a9aa3", muted: "#8a8a93", border: "#2c2d31",
			kw: "#bb9af7", str: "#9ece6a", num: "#ff9e64", com: "#565f89", fn: "#7aa2f7",
		}) + `
body{font-family:"Iosevka","Iosevka Term","Iosevka NF","JetBrains Mono",Menlo,Consolas,monospace;background:#121214;color:#d6d6d6;line-height:1.65;font-size:16px}
h1,h2,h3{color:#e8e8e8;font-weight:600}h1{border-bottom:1px solid #333;padding-bottom:.4rem}a{color:#7dd3a8}
pre{background:#1a1b1e;border:1px solid #2c2d31;padding:1rem;border-radius:8px;color:#d6d6d6}
code{font-size:.88em}
.output{background:#141b16;border:1px solid #2f4a3a;color:#c9e8d4;padding:.75rem 1rem;border-radius:8px;margin:.5rem 0}
.output pre{background:none;border:none;padding:0;margin:.25rem 0;color:inherit}
.output strong{color:#7dd3a8}
.block-head{color:#8a8a93}
.stderr{color:#ff7b72}img{max-width:100%;border:1px solid #2c2d31;border-radius:8px}`,
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
		CSS: layoutCSS(layoutColors{
			faint: "#e2e8f0", hover: "#4f46e5", active: "#4f46e5",
			text: "#475569", muted: "#94a3b8", border: "#e2e8f0",
			kw: "#7c3aed", str: "#0f766e", num: "#b45309", com: "#94a3b8", fn: "#4f46e5",
		}) + `
body{font-family:system-ui,-apple-system,"Segoe UI",sans-serif;line-height:1.7;color:#1e293b;background:#fff}
h1{font-size:2rem;letter-spacing:-.02em}h2{letter-spacing:-.01em;border-bottom:2px solid #e0e7ff;padding-bottom:.3rem}
pre{background:#f6f7f9;border:1px solid #e5e7eb;padding:1rem;border-radius:12px}
code{font-size:.88em}
.output{background:#f0fdf4;border:1px solid #bbf7d0;padding:.75rem 1rem;border-radius:12px;margin:.5rem 0}
.output pre{background:none;border:none;padding:0;margin:.25rem 0}
.output strong{color:#15803d}
.block-head{color:#64748b;text-transform:uppercase}
img{max-width:100%;border-radius:12px}.stderr{color:#b91c1c}a{color:#4f46e5}`,
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
	"paper": {
		Name:        "paper",
		Description: "Warm minimal light theme: cream paper, amber accents, soft green output",
		CSS: layoutCSS(layoutColors{
			faint: "#e9e2d2", hover: "#b45309", active: "#b45309",
			text: "#57503f", muted: "#a89f8b", border: "#e9e2d2",
			kw: "#9a3412", str: "#3f6212", num: "#b45309", com: "#a89f8b", fn: "#0f766e",
		}) + `
body{font-family:system-ui,-apple-system,"Segoe UI",sans-serif;line-height:1.7;color:#37322a;background:#faf7f0;font-size:16.5px}
h1{font-size:2rem;color:#29241d;letter-spacing:-.01em}h2{color:#29241d;border-bottom:1px solid #e9e2d2;padding-bottom:.3rem}
pre{background:#f3efe4;border:1px solid #e5ddca;padding:1rem;border-radius:8px}
code{font-size:.88em}
.output{background:#eef2e2;border:1px solid #d5ddc0;padding:.75rem 1rem;border-radius:8px;margin:.5rem 0}
.output pre{background:none;border:none;padding:0;margin:.25rem 0}
.output strong{color:#4d7c0f}
.block-head{color:#a89f8b;text-transform:uppercase}
img{max-width:100%;border:1px solid #e5ddca;border-radius:8px}.stderr{color:#b91c1c}a{color:#b45309}`,
		TypstPrelude: `#set document(title: {{TITLE}})
#set page(paper: "a4", margin: 2cm, fill: rgb("#faf7f0"))
#set text(font: ("Adwaita Sans", "FreeSans"), size: 10.5pt, fill: rgb("#37322a"))
#set heading(numbering: none)
#show heading: set text(fill: rgb("#29241d"))
#show heading.where(level: 1): it => [#set text(size: 22pt); #it.body; #v(0.2em, weak: true)]
#show raw: set text(size: 9.5pt)
#show link: set text(fill: rgb("#b45309"))`,
		TypstCodeFill:   `rgb("#f3efe4")`,
		TypstOutputFill: `rgb("#eef2e2")`,
		TypstAccent:     `rgb("#b45309")`,
	},
	"solarized": {
		Name:        "solarized",
		Description: "Solarized light: cream paper, blue accents, classic terminal palette",
		CSS: layoutCSS(layoutColors{
			faint: "#eee8d5", hover: "#268bd2", active: "#268bd2",
			text: "#586e75", muted: "#93a1a1", border: "#eee8d5",
			kw: "#6c71c4", str: "#2aa198", num: "#cb4b16", com: "#93a1a1", fn: "#268bd2",
		}) + `
body{font-family:system-ui,-apple-system,"Segoe UI",sans-serif;line-height:1.7;color:#657b83;background:#fdf6e3;font-size:16px}
h1{font-size:2rem;color:#586e75}h2{color:#586e75;border-bottom:1px solid #eee8d5;padding-bottom:.3rem}
pre{background:#eee8d5;border:1px solid #d9d0b7;padding:1rem;border-radius:8px;color:#586e75}
code{font-size:.88em}
.output{background:#e8ead0;border:1px solid #cdd2ab;padding:.75rem 1rem;border-radius:8px;margin:.5rem 0}
.output pre{background:none;border:none;padding:0;margin:.25rem 0;color:inherit}
.output strong{color:#859900}
.block-head{color:#93a1a1;text-transform:uppercase}
img{max-width:100%;border:1px solid #eee8d5;border-radius:8px}.stderr{color:#dc322f}a{color:#268bd2}`,
		TypstPrelude: `#set document(title: {{TITLE}})
#set page(paper: "a4", margin: 2cm, fill: rgb("#fdf6e3"))
#set text(font: ("Adwaita Sans", "FreeSans"), size: 10.5pt, fill: rgb("#657b83"))
#set heading(numbering: none)
#show heading: set text(fill: rgb("#586e75"))
#show heading.where(level: 1): it => [#set text(size: 22pt); #it.body; #v(0.2em, weak: true)]
#show raw: set text(size: 9.5pt)
#show link: set text(fill: rgb("#268bd2"))`,
		TypstCodeFill:   `rgb("#eee8d5")`,
		TypstOutputFill: `rgb("#e8ead0")`,
		TypstAccent:     `rgb("#268bd2")`,
	},
	"dracula": {
		Name:        "dracula",
		Description: "Dracula official palette: purple accents on dark, green output",
		CSS: layoutCSS(layoutColors{
			faint: "#44475a", hover: "#f8f8f2", active: "#bd93f9",
			text: "#b8b9c4", muted: "#6272a4", border: "#44475a",
			kw: "#ff79c6", str: "#f1fa8c", num: "#bd93f9", com: "#6272a4", fn: "#50fa7b",
		}) + `
body{font-family:system-ui,-apple-system,"Segoe UI",sans-serif;background:#282a36;color:#f8f8f2;line-height:1.65;font-size:16px}
h1,h2,h3{color:#f8f8f2;font-weight:700}h1{border-bottom:1px solid #44475a;padding-bottom:.4rem}
h2{border-bottom:1px solid #44475a;padding-bottom:.3rem}
a{color:#8be9fd}
pre{background:#21222c;border:1px solid #44475a;padding:1rem;border-radius:10px;color:#f8f8f2}
code{font-size:.88em}
.output{background:#1d2b23;border:1px solid #3d5a48;color:#a7e8bd;padding:.75rem 1rem;border-radius:10px;margin:.5rem 0}
.output pre{background:none;border:none;padding:0;margin:.25rem 0;color:inherit}
.output strong{color:#50fa7b}
.block-head{color:#6272a4;text-transform:uppercase}
.stderr{color:#ff5555}img{max-width:100%;border:1px solid #44475a;border-radius:10px}`,
		TypstPrelude: `#set document(title: {{TITLE}})
#set page(paper: "a4", margin: 2cm, fill: rgb("#282a36"))
#set text(font: ("Adwaita Sans", "FreeSans"), size: 10.5pt, fill: rgb("#f8f8f2"))
#set heading(numbering: none)
#show heading: set text(fill: rgb("#bd93f9"))
#show heading.where(level: 1): it => [#set text(size: 22pt); #it.body; #v(0.2em, weak: true)]
#show raw: set text(font: ("Iosevka NFM", "Iosevka NF", "DejaVu Sans Mono"), size: 9.5pt)
#show link: set text(fill: rgb("#8be9fd"))`,
		TypstCodeFill:   `rgb("#21222c")`,
		TypstOutputFill: `rgb("#1d2b23")`,
		TypstAccent:     `rgb("#bd93f9")`,
	},
	"nord": {
		Name:        "nord",
		Description: "Nord frost palette: arctic blue-grays, cool frost accents",
		CSS: layoutCSS(layoutColors{
			faint: "#434c5e", hover: "#eceff4", active: "#88c0d0",
			text: "#b6c0cf", muted: "#8b96a8", border: "#434c5e",
			kw: "#81a1c1", str: "#a3be8c", num: "#b48ead", com: "#616e88", fn: "#88c0d0",
		}) + `
body{font-family:system-ui,-apple-system,"Segoe UI",sans-serif;background:#2e3440;color:#d8dee9;line-height:1.65;font-size:16px}
h1,h2,h3{color:#eceff4;font-weight:700}h1{border-bottom:1px solid #434c5e;padding-bottom:.4rem}
h2{border-bottom:1px solid #434c5e;padding-bottom:.3rem}
a{color:#88c0d0}
pre{background:#3b4252;border:1px solid #434c5e;padding:1rem;border-radius:8px;color:#e5e9f0}
code{font-size:.88em}
.output{background:#34403a;border:1px solid #4c6b5a;color:#cfe3d0;padding:.75rem 1rem;border-radius:8px;margin:.5rem 0}
.output pre{background:none;border:none;padding:0;margin:.25rem 0;color:inherit}
.output strong{color:#a3be8c}
.block-head{color:#8b96a8;text-transform:uppercase}
.stderr{color:#bf616a}img{max-width:100%;border:1px solid #434c5e;border-radius:8px}`,
		TypstPrelude: `#set document(title: {{TITLE}})
#set page(paper: "a4", margin: 2cm, fill: rgb("#2e3440"))
#set text(font: ("Adwaita Sans", "FreeSans"), size: 10.5pt, fill: rgb("#d8dee9"))
#set heading(numbering: none)
#show heading: set text(fill: rgb("#88c0d0"))
#show heading.where(level: 1): it => [#set text(size: 22pt); #it.body; #v(0.2em, weak: true)]
#show raw: set text(font: ("Iosevka NFM", "Iosevka NF", "DejaVu Sans Mono"), size: 9.5pt)
#show link: set text(fill: rgb("#88c0d0"))`,
		TypstCodeFill:   `rgb("#3b4252")`,
		TypstOutputFill: `rgb("#34403a")`,
		TypstAccent:     `rgb("#88c0d0")`,
	},
}
