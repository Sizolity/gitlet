package utils

import (
	"os"
	"testing"
)

func TestWriteFileBytesCreatesParentDirectories(t *testing.T) {
	tmp := t.TempDir()
	prev, _ := os.Getwd()
	defer os.Chdir(prev)
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	path := "dir/sub/file.txt"
	WriteFileBytes(path, []byte("hello"))

	if !FileExists(path) {
		t.Fatalf("expected file to be created: %s", path)
	}
	if got := string(ReadFile(path)); got != "hello" {
		t.Fatalf("file content mismatch: got=%s want=hello", got)
	}
}

func TestRemoveFileByPathRemovesEmptyParents(t *testing.T) {
	tmp := t.TempDir()
	prev, _ := os.Getwd()
	defer os.Chdir(prev)
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	path := "a/b/c.txt"
	WriteFileBytes(path, []byte("x"))
	RemoveFileByPath(path)

	if FileExists(path) {
		t.Fatalf("file should be removed: %s", path)
	}
	if FileExists("a/b") {
		t.Fatalf("empty parent dir should be removed: a/b")
	}
	if FileExists("a") {
		t.Fatalf("empty parent dir should be removed: a")
	}
}

