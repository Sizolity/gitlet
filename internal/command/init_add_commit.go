package command

import (
	"fmt"
	"gitlet/config"
	gitlet "gitlet/internal/object"
	"gitlet/pkg/utils"
	"os"
	"strings"
)

func Init_gitlet() {
	_, err := os.Stat(".gitlet")
	if os.IsNotExist(err) {
		os.Mkdir(".gitlet", 0755)
		os.MkdirAll(config.COMMIT, 0755)
		os.MkdirAll(config.BLOB, 0755)
		os.MkdirAll(config.TREE, 0755)
		os.MkdirAll(config.BRANCHES, 0755)
		os.MkdirAll(config.REMOTES, 0755)
		os.Create(config.HEAD)
		os.Create(config.BRANCHES + "/master")

		commit := gitlet.NewInitCommit()
		utils.WriteFile(config.BRANCHES+"/master", commit.HashId)
		utils.WriteFile(config.HEAD, config.BRANCHES+"/master")
		commit.Persist()

		idx := gitlet.NewIndex()
		idx.Save()

		fmt.Println("Gitlet init success.")
	} else {
		fmt.Println("A Gitlet version-control system already exists in the current directory.")
	}
}

func Add(filename string) {
	filename = utils.NormalizePath(filename)

	if !utils.FileExists(filename) {
		fmt.Println("add: File does not exist.")
		return
	}

	patterns := gitlet.LoadIgnorePatterns()
	if gitlet.IsIgnored(filename, patterns) {
		fmt.Println("add: File is ignored by .gitletignore.")
		return
	}

	data := utils.ReadFile(filename)
	blob := gitlet.NewBlob(filename, data)

	idx := gitlet.LoadIndex()
	headCommit := gitlet.GetCommitById(gitlet.GetHEAD())

	if headBlobId, ok := headCommit.BlobIds[filename]; ok && headBlobId == blob.HashId {
		idx.Update(filename, headBlobId)
		idx.Save()
		return
	}

	blob.Persist()
	idx.Update(filename, blob.HashId)
	idx.Save()

	fmt.Println("Adding files succeed.")
}

func Commit(messages ...string) {
	message := strings.Join(messages, " ")

	idx := gitlet.LoadIndex()
	headCommit := gitlet.GetCommitById(gitlet.GetHEAD())

	if mapsEqual(idx.Entries, headCommit.BlobIds) {
		fmt.Println("Commit: Nothing to do.")
		return
	}

	commit := gitlet.NewCommit(message)
	commit.TreeId = gitlet.BuildTree(idx.Entries)
	commit.Persist()
	gitlet.MoveBranchPoint(commit.HashId)

	fmt.Println("Commit succeed.")
}

func Rm(filename string) {
	filename = utils.NormalizePath(filename)

	idx := gitlet.LoadIndex()
	headCommit := gitlet.GetCommitById(gitlet.GetHEAD())

	headBlobId, trackedByHead := headCommit.BlobIds[filename]
	indexBlobId := idx.GetBlobId(filename)
	inIndex := idx.Has(filename)

	stagedForAdd := inIndex && (!trackedByHead || indexBlobId != headBlobId)

	if !stagedForAdd && !trackedByHead {
		fmt.Println("rm: No reason to remove the file.")
		return
	}

	if stagedForAdd && !trackedByHead {
		// Staged a new file not in HEAD - just unstage.
		idx.Remove(filename)
		idx.Save()
		fmt.Println("rm: Unstaged file.")
		return
	}

	// Tracked by HEAD - stage for removal.
	idx.Remove(filename)
	idx.Save()
	if utils.FileExists(filename) {
		os.Remove(filename)
	}
	fmt.Println("rm: File removed.")
}
