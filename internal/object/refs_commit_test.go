package gitlet

import (
	"os"
	"testing"

	"gitlet/config"
	"gitlet/pkg/utils"
)

func setupGitletRepo(t *testing.T) *Commit {
	t.Helper()
	tmp := t.TempDir()
	prev, _ := os.Getwd()
	t.Cleanup(func() {
		_ = os.Chdir(prev)
	})
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	if err := os.MkdirAll(config.COMMIT, 0755); err != nil {
		t.Fatalf("mkdir commits: %v", err)
	}
	if err := os.MkdirAll(config.BLOB, 0755); err != nil {
		t.Fatalf("mkdir blobs: %v", err)
	}
	if err := os.MkdirAll(config.TREE, 0755); err != nil {
		t.Fatalf("mkdir trees: %v", err)
	}
	if err := os.MkdirAll(config.BRANCHES, 0755); err != nil {
		t.Fatalf("mkdir branches: %v", err)
	}
	if err := os.MkdirAll(config.REMOTES, 0755); err != nil {
		t.Fatalf("mkdir remotes: %v", err)
	}

	initCommit := NewInitCommit()
	initCommit.Persist()
	utils.WriteFile(config.BRANCHES+"/master", initCommit.HashId)
	utils.WriteFile(config.HEAD, config.BRANCHES+"/master")
	NewIndex().Save()
	return initCommit
}

func TestRefsBranchAndDetachedBehavior(t *testing.T) {
	initCommit := setupGitletRepo(t)

	if IsDetachedHEAD() {
		t.Fatalf("HEAD should start attached to branch")
	}
	if got := GetHEADBranch(); got != "master" {
		t.Fatalf("HEAD branch mismatch: got=%s want=master", got)
	}
	if got := GetHEAD(); got != initCommit.HashId {
		t.Fatalf("HEAD commit mismatch: got=%s want=%s", got, initCommit.HashId)
	}
	if !BranchExist("master") {
		t.Fatalf("master branch should exist")
	}
	if BranchExist("feature") {
		t.Fatalf("feature branch should not exist yet")
	}

	MoveBranchPoint("next-commit")
	if got := string(utils.ReadFile(config.BRANCHES + "/master")); got != "next-commit" {
		t.Fatalf("branch pointer should move in attached mode, got=%s", got)
	}

	DetachHEAD("detached-commit")
	if !IsDetachedHEAD() {
		t.Fatalf("HEAD should be detached")
	}
	if got := GetHEADBranch(); got != "" {
		t.Fatalf("detached HEAD should have empty branch name, got=%s", got)
	}
	if got := GetHEAD(); got != "detached-commit" {
		t.Fatalf("detached HEAD commit mismatch: got=%s", got)
	}

	MoveBranchPoint("detached-new")
	if got := string(utils.ReadFile(config.HEAD)); got != "detached-new" {
		t.Fatalf("detached mode should move HEAD directly, got=%s", got)
	}
}

func TestCommitPersistAndLoadFlattensTree(t *testing.T) {
	setupGitletRepo(t)

	flat := map[string]string{
		"a.txt":        "blob-a",
		"dir/b.txt":    "blob-b",
		"dir/sub/c.md": "blob-c",
	}
	treeID := BuildTree(flat)

	c := NewCommit("snapshot")
	c.TreeId = treeID
	c.Persist()

	got := GetCommitById(c.HashId)
	if got.TreeId != treeID {
		t.Fatalf("treeId mismatch: got=%s want=%s", got.TreeId, treeID)
	}
	if len(got.BlobIds) != len(flat) {
		t.Fatalf("blob map size mismatch: got=%d want=%d", len(got.BlobIds), len(flat))
	}
	for p, id := range flat {
		if got.BlobIds[p] != id {
			t.Fatalf("flattened blob mismatch for %s: got=%s want=%s", p, got.BlobIds[p], id)
		}
	}

	all := GetAllCommits()
	if len(all) < 2 {
		t.Fatalf("expected at least init+1 commits, got=%d", len(all))
	}
	seen := make(map[string]bool)
	for _, item := range all {
		seen[item.HashId] = true
	}
	if !seen[c.HashId] {
		t.Fatalf("new commit not found in GetAllCommits")
	}
}

func TestNewMergeCommitKeepsAllParents(t *testing.T) {
	parents := []string{"p1", "p2"}
	c := NewMergeCommit("merge branch", parents)
	if c.Message != "merge branch" {
		t.Fatalf("merge commit message mismatch: got=%s", c.Message)
	}
	if len(c.Parent) != 2 || c.Parent[0] != "p1" || c.Parent[1] != "p2" {
		t.Fatalf("merge commit parents mismatch: got=%v", c.Parent)
	}
	if c.HashId == "" {
		t.Fatalf("merge commit hash should not be empty")
	}
}

