package runtime

import "time"

// Result is the outcome of executing one block (or session step).
type Result struct {
	Block      string   `json:"block"`
	Success    bool     `json:"success"`
	Stdout     string   `json:"stdout"`
	Stderr     string   `json:"stderr"`
	Value      string   `json:"value"`
	Artifacts  []string `json:"artifacts"`
	ExitCode   int      `json:"exit_code"`
	DurationMs int64    `json:"duration_ms"`
	Error      string   `json:"error,omitempty"`
	Runtime    string   `json:"runtime"`
}

// Runtime executes code in a language.
type Runtime interface {
	Name() string
	IsAvailable() bool
	// Execute runs code (single block) with working dir; artifacts are
	// file paths (relative to workdir) created during execution.
	Execute(code string, workdir string) Result
}

func duration(start time.Time) int64 { return time.Since(start).Milliseconds() }
