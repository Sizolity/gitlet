package utils

import (
	"crypto/sha1"
	"fmt"
	"path/filepath"
)

func GetArgsNum(args []string) int {
	return len(args)
}

/* Content-addressable SHA-1 hash. */
func GenerateID(data []byte) string {
	hasher := sha1.New()
	hasher.Write(data)
	return fmt.Sprintf("%x", hasher.Sum(nil))
}

// NormalizePath cleans a file path so that "./foo.txt" and "foo.txt" resolve identically.
func NormalizePath(path string) string {
	return filepath.Clean(path)
}
