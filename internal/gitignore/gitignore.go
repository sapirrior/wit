package gitignore

import (
	"io"
	"os"
	"path/filepath"
	"strings"
)

type GitignoreRule struct {
	negated  bool
	dirOnly  bool
	anchored bool
	pattern  string
	basePath string
}

type GitignoreContext struct {
	rules    []GitignoreRule
	rootPath string
}

func NewGitignoreContext(rootPath string) *GitignoreContext {
	abs, err := filepath.Abs(rootPath)
	if err != nil {
		abs = rootPath
	}
	return &GitignoreContext{
		rootPath: filepath.ToSlash(abs),
	}
}

func (ctx *GitignoreContext) LoadFile(path, basePath string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	absBase, err := filepath.Abs(basePath)
	if err != nil {
		absBase = basePath
	}
	absBase = filepath.ToSlash(absBase)

	data, err := io.ReadAll(f)
	if err != nil {
		return
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		negated := false
		if strings.HasPrefix(line, "!") {
			negated = true
			line = strings.TrimSpace(line[1:])
		}
		if line == "" {
			continue
		}

		dirOnly := false
		if strings.HasSuffix(line, "/") {
			dirOnly = true
			line = line[:len(line)-1]
		}

		anchored := false
		if strings.HasPrefix(line, "/") {
			anchored = true
			line = line[1:]
		} else if strings.Contains(line, "/") {
			anchored = true
		}

		ctx.rules = append(ctx.rules, GitignoreRule{
			negated:  negated,
			dirOnly:  dirOnly,
			anchored: anchored,
			pattern:  line,
			basePath: absBase,
		})
	}
}

func (ctx *GitignoreContext) LoadAll(dirPath string) {
	filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if info.Name() == ".git" {
				return filepath.SkipDir
			}
			gi := filepath.Join(path, ".gitignore")
			if _, err := os.Stat(gi); err == nil {
				ctx.LoadFile(gi, path)
			}
		}
		return nil
	})
}

func matchPattern(pattern, relPath string) bool {
	if strings.Contains(pattern, "**") {
		if strings.HasPrefix(pattern, "**/") {
			stem := pattern[3:]
			parts := strings.Split(relPath, "/")
			for i := 0; i < len(parts); i++ {
				sub := strings.Join(parts[i:], "/")
				if matched, _ := filepath.Match(stem, sub); matched {
					return true
				}
			}
		}
		patClean := strings.ReplaceAll(pattern, "**/", "*")
		patClean = strings.ReplaceAll(patClean, "/**", "*")
		matched, _ := filepath.Match(patClean, relPath)
		return matched
	}

	if strings.Contains(pattern, "/") {
		matched, _ := filepath.Match(pattern, relPath)
		return matched
	}

	parts := strings.Split(relPath, "/")
	for _, part := range parts {
		if matched, _ := filepath.Match(pattern, part); matched {
			return true
		}
	}
	return false
}

func (ctx *GitignoreContext) IsIgnored(path string, isDir bool) bool {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	abs = filepath.ToSlash(abs)

	if !strings.HasPrefix(abs, ctx.rootPath) {
		return true
	}

	matched := false
	for _, rule := range ctx.rules {
		if rule.dirOnly && !isDir {
			continue
		}

		if !strings.HasPrefix(abs, rule.basePath) {
			continue
		}

		rel := abs[len(rule.basePath):]
		rel = strings.TrimPrefix(rel, "/")
		if rel == "" {
			continue
		}

		hit := false
		if rule.anchored {
			hit, _ = filepath.Match(rule.pattern, rel)
			if !hit && strings.HasPrefix(rel, rule.pattern+"/") {
				hit = true
			}
		} else {
			hit = matchPattern(rule.pattern, rel)
		}

		if hit {
			matched = !rule.negated
		}
	}
	return matched
}
