package instruction

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"gitlet/gitlet"
	"gitlet/utils"
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

func TestInitAddCommitWithNestedFileCreatesTreeSnapshot(t *testing.T) {
	runID := "run_instruction_nested_commit"
	tmp := t.TempDir()
	prev, _ := os.Getwd()
	defer os.Chdir(prev)
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	Init_gitlet()

	path := "src/lib/a.txt"
	utils.WriteFileBytes(path, []byte("hello-tree"))
	debugLog(t, runID, "H5", "instruction_test.go:45", "created nested file", path)

	Add(path)
	Commit("add nested file")
	head := gitlet.GetCommitById(gitlet.GetHEAD())
	debugLog(t, runID, "H5", "instruction_test.go:50", "created commit", fmt.Sprintf("treeId=%s", head.TreeId))

	if head.TreeId == "" {
		t.Fatalf("expected non-empty treeId in HEAD commit")
	}
	flat := gitlet.FlattenTree(head.TreeId)
	debugLog(t, runID, "H6", "instruction_test.go:56", "flatten tree after commit", fmt.Sprintf("tracked=%d", len(flat)))

	if got := flat[path]; got == "" {
		t.Fatalf("expected tracked nested path %s in flattened tree", path)
	}
}
