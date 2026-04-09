package gitlet

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// LoadIgnorePatterns reads .gitletignore and returns the list of glob patterns.
func LoadIgnorePatterns() []string {
	file, err := os.Open(".gitletignore")
	if err != nil {
		return nil
	}
	defer file.Close()

	var patterns []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patterns = append(patterns, line)
	}
	return patterns
}

// IsIgnored checks whether filename matches any of the ignore patterns.
func IsIgnored(filename string, patterns []string) bool {
	for _, pattern := range patterns {
		if matched, _ := filepath.Match(pattern, filename); matched {
			return true
		}
		if matched, _ := filepath.Match(pattern, filepath.Base(filename)); matched {
			return true
		}
	}
	return false
}
