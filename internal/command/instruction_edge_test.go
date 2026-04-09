package command

import (
	"io"
	"os"
	"strings"
	"testing"

	gitlet "gitlet/internal/object"
	"gitlet/pkg/utils"
)

func captureOutput(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w
	fn()
	_ = w.Close()
	os.Stdout = old
	out, _ := io.ReadAll(r)
	return string(out)
}

func setupInstructionRepo(t *testing.T) {
	t.Helper()
	tmp := t.TempDir()
	prev, _ := os.Getwd()
	t.Cleanup(func() {
		_ = os.Chdir(prev)
	})
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	Init_gitlet()
}

func TestMergeRejectsDetachedHEAD(t *testing.T) {
	setupInstructionRepo(t)
	gitlet.DetachHEAD(gitlet.GetHEAD())
	out := captureOutput(t, func() { Merge("feature") })
	if !strings.Contains(out, "merge: Cannot merge in detached HEAD state.") {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestMergeRejectsMergingSelfBranch(t *testing.T) {
	setupInstructionRepo(t)
	out := captureOutput(t, func() { Merge("master") })
	if !strings.Contains(out, "merge: Cannot merge a branch with itself.") {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestStatusShowsDetachedHead(t *testing.T) {
	setupInstructionRepo(t)
	gitlet.DetachHEAD(gitlet.GetHEAD())
	out := captureOutput(t, func() { Status() })
	if !strings.Contains(out, "*HEAD detached at ") {
		t.Fatalf("status should show detached head, got: %q", out)
	}
}

func TestCheckoutCommitDetachesHead(t *testing.T) {
	setupInstructionRepo(t)

	path := "docs/readme.txt"
	utils.WriteFileBytes(path, []byte("v1"))
	Add(path)
	Commit("first")

	commitID := gitlet.GetHEAD()
	out := captureOutput(t, func() { Checkout(commitID) })
	if !strings.Contains(out, "checkout: HEAD detached at ") {
		t.Fatalf("unexpected checkout output: %q", out)
	}
	if got := gitlet.GetHEADBranch(); got != "" {
		t.Fatalf("HEAD should be detached after checkout commit, got branch=%s", got)
	}
}

func TestResetRestoresOldCommitAndFileContent(t *testing.T) {
	setupInstructionRepo(t)

	path := "a.txt"
	utils.WriteFileBytes(path, []byte("v1"))
	Add(path)
	Commit("c1")
	c1 := gitlet.GetHEAD()

	utils.WriteFileBytes(path, []byte("v2"))
	Add(path)
	Commit("c2")

	out := captureOutput(t, func() { Reset(c1) })
	if !strings.Contains(out, "reset: HEAD at ") {
		t.Fatalf("unexpected reset output: %q", out)
	}
	if got := gitlet.GetHEAD(); got != c1 {
		t.Fatalf("HEAD should move to c1: got=%s want=%s", got, c1)
	}
	if got := string(utils.ReadFile(path)); got != "v1" {
		t.Fatalf("worktree file should be restored to v1, got=%s", got)
	}
}

