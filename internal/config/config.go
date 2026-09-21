package config

import (
	"os"
	"path/filepath"
	"strings"
)

// Config mirrors loom.toml / [project] etc.
type Config struct {
	ProjectName string
	Main        string
	TangleOut   string
	HTML        outPair
	PDF         outPair
}

type outPair struct {
	Output string
}

func defaults() Config {
	return Config{TangleOut: "src", HTML: outPair{Output: "dist/index.html"}, PDF: outPair{Output: "dist/main.pdf"}}
}

// Load searches for loom.toml next to loomFile or cwd.
func Load(loomFile string) Config {
	cfg := defaults()
	candidates := []string{
		filepath.Join(filepath.Dir(loomFile), "loom.toml"),
		"loom.toml",
	}
	for _, c := range candidates {
		data, err := os.ReadFile(c)
		if err != nil {
			continue
		}
		parseToml(string(data), &cfg)
		break
	}
	return cfg
}

// minimal TOML subset: [section], key = "value".
func parseToml(src string, cfg *Config) {
	section := ""
	for _, line := range strings.Split(src, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.Trim(line, "[]")
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		k := strings.TrimSpace(parts[0])
		v := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
		switch section {
		case "project":
			if k == "name" {
				cfg.ProjectName = v
			}
		case "source":
			if k == "main" {
				cfg.Main = v
			}
		case "tangle":
			if k == "output" {
				cfg.TangleOut = v
			}
		case "export.html":
			if k == "output" {
				cfg.HTML.Output = v
			}
		case "export.pdf":
			if k == "output" {
				cfg.PDF.Output = v
			}
		}
	}
}
