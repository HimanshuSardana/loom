package ast

// Position is a 1-indexed source location.
type Position struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

type Location struct {
	File  string   `json:"file"`
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// Block is a sum type; use Type discriminator.
type BlockType string

const (
	BlockHeading   BlockType = "heading"
	BlockParagraph BlockType = "paragraph"
	BlockList      BlockType = "list"
	BlockCode      BlockType = "code"
)

type Block struct {
	Type BlockType `json:"type"`

	// Heading
	Level int    `json:"level,omitempty"`
	Text  string `json:"text,omitempty"`

	// List
	Items []string `json:"items,omitempty"`

	// Code
	Language string            `json:"language,omitempty"`
	Name     string            `json:"name,omitempty"`
	Source   string            `json:"source,omitempty"`
	Options  map[string]string `json:"options,omitempty"`
	Flags    map[string]bool   `json:"flags,omitempty"`
	Run      bool              `json:"run,omitempty"`
	Session  string            `json:"session,omitempty"`
	File     string            `json:"file,omitempty"`
	Depends  []string          `json:"depends,omitempty"`
	Isolated bool              `json:"isolated,omitempty"`

	Loc Location `json:"loc"`
}

type Document struct {
	File   string  `json:"file"`
	Title  string  `json:"title,omitempty"`
	Blocks []Block `json:"blocks"`
}

// CodeBlocks returns code blocks in document order.
func (d *Document) CodeBlocks() []*Block {
	var out []*Block
	for i := range d.Blocks {
		if d.Blocks[i].Type == BlockCode {
			out = append(out, &d.Blocks[i])
		}
	}
	return out
}

// FindBlock returns the named code block or nil.
func (d *Document) FindBlock(name string) *Block {
	for i := range d.Blocks {
		if d.Blocks[i].Type == BlockCode && d.Blocks[i].Name == name {
			return &d.Blocks[i]
		}
	}
	return nil
}

// Runnable returns code blocks with {run}.
func (d *Document) Runnable() []*Block {
	var out []*Block
	for _, b := range d.CodeBlocks() {
		if b.Run {
			out = append(out, b)
		}
	}
	return out
}
