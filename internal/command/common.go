package command

import (
	gitlet "gitlet/internal/object"
	"os"
	"path/filepath"
	"sort"
)

func getWorkTreeFiles() []string {
	var files []string
	patterns := gitlet.LoadIgnorePatterns()
	filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if d.Name() == ".gitlet" {
				return filepath.SkipDir
			}
			if gitlet.IsIgnored(d.Name(), patterns) {
				return filepath.SkipDir
			}
			return nil
		}
		path = filepath.ToSlash(path)
		if gitlet.IsIgnored(path, patterns) {
			return nil
		}
		files = append(files, path)
		return nil
	})
	return files
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func mapsEqual(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if bv, ok := b[k]; !ok || bv != v {
			return false
		}
	}
	return true
}
