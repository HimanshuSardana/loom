package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/HimanshuSardana/loom/internal/runtime"
)

// Entry is a cached execution result.
type Entry struct {
	Key    string         `json:"key"`
	Result runtime.Result `json:"result"`
}

func dirFor(loomFile string) string {
	return filepath.Join(filepath.Dir(loomFile), ".loom", "cache")
}

// Key computes cache key from source+language+session+config.
func Key(blockName, source, language, session, config string) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("%s\x00%s\x00%s\x00%s\x00%s", blockName, source, language, session, config)))
	return hex.EncodeToString(h[:])
}

func pathFor(loomFile, blockName, key string) string {
	safe := ""
	for _, r := range blockName {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' || r == '.' {
			safe += string(r)
		} else {
			safe += "_"
		}
	}
	return filepath.Join(dirFor(loomFile), safe+"."+key[:16]+".json")
}

// Lookup returns cached result if key matches.
func Lookup(loomFile, blockName, key string) (runtime.Result, bool) {
	p := pathFor(loomFile, blockName, key)
	data, err := os.ReadFile(p)
	if err != nil {
		return runtime.Result{}, false
	}
	var e Entry
	if err := json.Unmarshal(data, &e); err != nil {
		return runtime.Result{}, false
	}
	if e.Key != key {
		return runtime.Result{}, false
	}
	return e.Result, true
}

// Store saves a result.
func Store(loomFile, blockName, key string, res runtime.Result) error {
	p := pathFor(loomFile, blockName, key)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	e := Entry{Key: key, Result: res}
	data, _ := json.MarshalIndent(e, "", "  ")
	return os.WriteFile(p, data, 0o644)
}

// Clean removes cache dir for a loom file.
func Clean(loomFile string) error {
	return os.RemoveAll(filepath.Join(filepath.Dir(loomFile), ".loom"))
}
