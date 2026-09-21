package runtime

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// snapshot lists files in dir (recursive, relative paths with sizes).
func snapshot(dir string) map[string]int64 {
	out := map[string]int64{}
	filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(dir, p)
		if err != nil {
			return nil
		}
		out[rel] = info.Size()
		return nil
	})
	return out
}

func detectArtifacts(before, after map[string]int64, workdir string) []string {
	var arts []string
	for f, sz := range after {
		bsz, ok := before[f]
		if !ok || sz != bsz {
			// ignore our own temp scripts
			if strings.HasPrefix(f, ".loom-tmp-") {
				continue
			}
			arts = append(arts, f)
		}
	}
	return arts
}

func runCmd(workdir, name string, args []string, timeoutSec int) (string, string, int) {
	if timeoutSec <= 0 {
		timeoutSec = 60
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSec)*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = workdir
	var so, se bytes.Buffer
	cmd.Stdout = &so
	cmd.Stderr = &se
	err := cmd.Run()
	code := 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			code = ee.ExitCode()
		} else {
			code = 1
		}
	}
	return so.String(), se.String(), code
}
