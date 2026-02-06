package core

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

var (
	todoPatterns    = []string{"**/TODO*.md", "**/todo*.md"}
	defaultTodoFile = "TODO.md"
)

// FoundFile represents a discovered TODO file
type FoundFile struct {
	Path         string
	RelativePath string
}

// FindTodoFiles finds TODO files in a directory
func FindTodoFiles(cwd string, recursive bool) ([]FoundFile, error) {
	var patterns []string
	if recursive {
		patterns = todoPatterns
	} else {
		patterns = []string{"TODO*.md", "todo*.md"}
	}

	var files []FoundFile
	seen := make(map[string]bool)

	for _, pattern := range patterns {
		fullPattern := filepath.Join(cwd, pattern)
		matches, err := doublestar.FilepathGlob(fullPattern)
		if err != nil {
			return nil, err
		}

		for _, match := range matches {
			// Skip node_modules and .git
			if strings.Contains(match, "node_modules") || strings.Contains(match, ".git") {
				continue
			}

			if seen[match] {
				continue
			}
			seen[match] = true

			relPath, err := filepath.Rel(cwd, match)
			if err != nil {
				relPath = match
			}

			files = append(files, FoundFile{
				Path:         match,
				RelativePath: relPath,
			})
		}
	}

	return files, nil
}

// FindDefaultTodoFile finds or returns the default TODO.md path
func FindDefaultTodoFile(cwd string) (string, error) {
	defaultPath := filepath.Join(cwd, defaultTodoFile)

	if _, err := os.Stat(defaultPath); err == nil {
		return defaultPath, nil
	}

	// Look for any TODO*.md in current directory
	files, err := FindTodoFiles(cwd, false)
	if err != nil {
		return "", err
	}

	if len(files) > 0 {
		return files[0].Path, nil
	}

	// Return default path (will be created if needed)
	return defaultPath, nil
}

// FindTodoFilesInSubdirs finds TODO files grouped by directory
func FindTodoFilesInSubdirs(cwd string) (map[string][]FoundFile, error) {
	files, err := FindTodoFiles(cwd, true)
	if err != nil {
		return nil, err
	}

	byDir := make(map[string][]FoundFile)

	for _, file := range files {
		dir := filepath.Dir(file.RelativePath)
		if dir == "." {
			dir = "(root)"
		}

		byDir[dir] = append(byDir[dir], file)
	}

	return byDir, nil
}
