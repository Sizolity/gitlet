package gitlet

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"gitlet/config"
)

func debugLog(t *testing.T, runID, hypothesisID, location, message string, data string) {
	t.Helper()
	// #region agent log
	f, err := os.OpenFile("/home/karo/huic/gitlet/debug-66ead7.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("open debug log: %v", err)
	}
	line := fmt.Sprintf(
		`{"sessionId":"66ead7","runId":"%s","hypothesisId":"%s","location":"%s","message":"%s","data":{"detail":"%s"},"timestamp":%d}`+"\n",
		runID, hypothesisID, location, message, strings.ReplaceAll(data, `"`, `'`), time.Now().UnixMilli(),
	)
	if _, err := f.WriteString(line); err != nil {
		_ = f.Close()
		t.Fatalf("write debug log: %v", err)
	}
	_ = f.Close()
	// #endregion
}

func TestBuildTreeAndFlattenTreeRoundTrip(t *testing.T) {
	runID := "run_tree_roundtrip"
	tmp := t.TempDir()
	prev, _ := os.Getwd()
	defer os.Chdir(prev)
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	if err := os.MkdirAll(config.TREE, 0755); err != nil {
		t.Fatalf("mkdir trees: %v", err)
	}

	flat := map[string]string{
		"foo.txt":            "aaa",
		"src/main.go":        "bbb",
		"src/util/helper.go": "ccc",
	}
	debugLog(t, runID, "H1", "tree_test.go:52", "before BuildTree", "flat includes root and nested paths")

	treeID := BuildTree(flat)
	if treeID == "" {
		t.Fatalf("treeID should not be empty")
	}
	debugLog(t, runID, "H1", "tree_test.go:58", "after BuildTree", "treeID generated")

	got := FlattenTree(treeID)
	debugLog(t, runID, "H2", "tree_test.go:61", "after FlattenTree", fmt.Sprintf("flatten_count=%d", len(got)))

	if len(got) != len(flat) {
		t.Fatalf("flatten length mismatch: got=%d want=%d", len(got), len(flat))
	}
	for k, v := range flat {
		if got[k] != v {
			t.Fatalf("flatten mismatch for %s: got=%s want=%s", k, got[k], v)
		}
	}
}

func TestCommitJSONContainsTreeIdNotBlobIds(t *testing.T) {
	runID := "run_commit_json"
	c := &Commit{
		Message: "msg",
		Parent:  []string{"p1"},
		HashId:  "h1",
		TreeId:  "t1",
		BlobIds: map[string]string{"a.txt": "blob1"},
	}
	b, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("marshal commit: %v", err)
	}
	s := string(b)
	debugLog(t, runID, "H3", "tree_test.go:87", "marshal commit json", s)

	if !strings.Contains(s, `"treeId":"t1"`) {
		t.Fatalf("expected treeId in commit json: %s", s)
	}
	if strings.Contains(s, "blobIds") {
		t.Fatalf("blobIds should not be serialized: %s", s)
	}
}

func TestNewInitCommitHasPersistedTree(t *testing.T) {
	runID := "run_init_commit_tree"
	tmp := t.TempDir()
	prev, _ := os.Getwd()
	defer os.Chdir(prev)
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	if err := os.MkdirAll(config.TREE, 0755); err != nil {
		t.Fatalf("mkdir trees: %v", err)
	}

	c := NewInitCommit()
	debugLog(t, runID, "H4", "tree_test.go:110", "new init commit", fmt.Sprintf("treeId=%s", c.TreeId))

	if c.TreeId == "" {
		t.Fatalf("init commit treeId should not be empty")
	}
	tree := GetTreeById(c.TreeId)
	if tree == nil {
		t.Fatalf("init tree should be persisted and loadable")
	}
	if len(tree.Entries) != 0 {
		t.Fatalf("init tree should be empty, got %d entries", len(tree.Entries))
	}
}
