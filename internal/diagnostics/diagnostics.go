package diagnostics

import "fmt"

type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
)

type Diagnostic struct {
	File     string   `json:"file"`
	Line     int      `json:"line"`
	Column   int      `json:"column"`
	Block    string   `json:"block,omitempty"`
	Severity Severity `json:"severity"`
	Message  string   `json:"message"`
}

func (d Diagnostic) Error() string { return d.String() }

func (d Diagnostic) String() string {
	loc := d.File
	if d.Line > 0 {
		loc += fmt.Sprintf(":%d:%d", d.Line, d.Column)
	}
	if d.Block != "" {
		return fmt.Sprintf("loom: %s\nblock: %s\n%s", d.Message, d.Block, loc)
	}
	return fmt.Sprintf("loom: %s\n%s", d.Message, loc)
}

type List []Diagnostic

func (l List) HasErrors() bool {
	for _, d := range l {
		if d.Severity == SeverityError {
			return true
		}
	}
	return false
}

func (l List) Error() string {
	msg := ""
	for i, d := range l {
		if i > 0 {
			msg += "\n"
		}
		msg += d.String()
	}
	return msg
}
