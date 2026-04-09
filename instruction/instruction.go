package instruction

import (
	"fmt"
	"gitlet/config"
	"gitlet/gitlet"
	"gitlet/utils"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func Init_gitlet() {
	_, err := os.Stat(".gitlet")
	if os.IsNotExist(err) {
		os.Mkdir(".gitlet", 0755)
		os.MkdirAll(config.COMMIT, 0755)
		os.MkdirAll(config.BLOB, 0755)
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
	blobIds := make(map[string]string)
	for k, v := range idx.Entries {
		blobIds[k] = v
	}
	commit.BlobIds = blobIds
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
		// Staged a new file not in HEAD — just unstage
		idx.Remove(filename)
		idx.Save()
		fmt.Println("rm: Unstaged file.")
		return
	}

	// Tracked by HEAD — stage for removal
	idx.Remove(filename)
	idx.Save()
	if utils.FileExists(filename) {
		os.Remove(filename)
	}
	fmt.Println("rm: File removed.")
}

func Log() {
	commitIdLast := gitlet.GetHEAD()
	logHelper(commitIdLast)
}

func logHelper(commitId string) {
	commit := gitlet.GetCommitById(commitId)
	fmt.Printf("* %s %s\n", utils.Colorize(commit.HashId[0:7], utils.FgMagenta),
		utils.Colorize(commit.Message, utils.FgCyan))
	if commit.Parent != nil {
		logHelper(commit.Parent[0])
	}
}

func GlobalLog() {
	dirs := utils.ReadDir(config.COMMIT)
	for _, item := range dirs {
		fmt.Printf("* ")
		fmt.Println(utils.Colorize(item.Name()[:7], utils.FgMagenta))
	}
}

func Find(commitMessage ...string) {
	commits := gitlet.GetAllCommits()
	for _, commit := range commits {
		if commit.Message == strings.Join(commitMessage, " ") {
			fmt.Println(utils.Colorize(commit.HashId[:7], utils.FgMagenta))
		}
	}
}

func Status() {
	fmt.Printf("=== Branches ===\n")
	HEADBranch := gitlet.GetHEADBranch()
	branches := utils.ReadDir(config.BRANCHES)
	if HEADBranch == "" {
		commitId := gitlet.GetHEAD()
		fmt.Printf("*HEAD detached at %s\n", commitId[:7])
	} else {
		fmt.Printf("*%s\n", HEADBranch)
	}
	for _, branch := range branches {
		if name := branch.Name(); name != HEADBranch {
			fmt.Printf(" %s\n", name)
		}
	}

	idx := gitlet.LoadIndex()
	headCommit := gitlet.GetCommitById(gitlet.GetHEAD())

	// Staged files: in index with different blob than HEAD, or not in HEAD at all
	var staged []string
	for path, blobId := range idx.Entries {
		if headBlobId, ok := headCommit.BlobIds[path]; !ok || headBlobId != blobId {
			staged = append(staged, path)
		}
	}
	sort.Strings(staged)
	fmt.Printf("\n=== Staged Files ===\n")
	for _, f := range staged {
		fmt.Println(f)
	}

	// Removed files: in HEAD but not in index
	var removed []string
	for path := range headCommit.BlobIds {
		if !idx.Has(path) {
			removed = append(removed, path)
		}
	}
	sort.Strings(removed)
	fmt.Printf("\n=== Removed Files ===\n")
	for _, f := range removed {
		fmt.Println(f)
	}

	// Modifications not staged for commit
	var modified []string
	for path, blobId := range idx.Entries {
		if utils.FileExists(path) {
			content := utils.ReadFile(path)
			if utils.GenerateID(content) != blobId {
				modified = append(modified, path+" (modified)")
			}
		} else {
			modified = append(modified, path+" (deleted)")
		}
	}
	sort.Strings(modified)
	fmt.Printf("\n=== Modifications Not Staged For Commit ===\n")
	for _, f := range modified {
		fmt.Println(f)
	}

	// Untracked files
	var untracked []string
	workFiles := getWorkTreeFiles()
	for _, f := range workFiles {
		if !idx.Has(f) {
			untracked = append(untracked, f)
		}
	}
	sort.Strings(untracked)
	fmt.Printf("\n=== Untracked Files ===\n")
	for _, f := range untracked {
		fmt.Println(f)
	}
}

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

func Checkout(args ...string) {
	NumArgs := len(args)
	if NumArgs == 2 {
		if args[0] == "-" {
			checkoutFile(gitlet.GetHEAD(), args[1])
		} else {
			fmt.Println("checkout: Wrong argument.")
		}
	} else if NumArgs == 3 {
		if args[1] == "-" {
			checkoutFile(args[0], args[2])
		} else {
			fmt.Println("checkout: Wrong argument.")
		}
	} else if NumArgs == 1 {
		if gitlet.BranchExist(args[0]) {
			switchBranch(args[0])
		} else {
			detachCheckout(args[0])
		}
	} else {
		fmt.Println("checkout: Get wrong argument num.")
	}
}

func checkoutFile(commitId string, filename string) {
	filename = utils.NormalizePath(filename)
	commit := gitlet.GetCommitById(commitId)
	if blobId, ok := commit.BlobIds[filename]; ok {
		blob := gitlet.GetBlobById(blobId)
		if blob == nil {
			fmt.Println("checkout: Blob data missing.")
			return
		}
		utils.WriteFileBytes(blob.FilePath, blob.Contents)
		fmt.Println("checkout: Get file in Worktree.")
	} else {
		fmt.Println("checkout: Can't find target file in last commit.")
	}
}

func switchBranch(branchName string) {
	if !gitlet.BranchExist(branchName) {
		fmt.Println("checkout: Branch not exist.")
		return
	}

	// Remove working tree files of the current commit
	commitId := gitlet.GetHEAD()
	commit := gitlet.GetCommitById(commitId)
	for _, blobId := range commit.BlobIds {
		blob := gitlet.GetBlobById(blobId)
		if blob != nil {
			utils.RemoveFileByPath(blob.FilePath)
		}
	}

	// Switch HEAD and restore working tree
	gitlet.MoveHEAD(branchName)
	commitId = gitlet.GetHEAD()
	commit = gitlet.GetCommitById(commitId)
	for _, blobId := range commit.BlobIds {
		blob := gitlet.GetBlobById(blobId)
		if blob != nil {
			utils.WriteFileBytes(blob.FilePath, blob.Contents)
		}
	}

	// Sync index with the new commit
	idx := gitlet.NewIndex()
	for k, v := range commit.BlobIds {
		idx.Entries[k] = v
	}
	idx.Save()

	fmt.Printf("checkout: switch to %s.\n", branchName)
}

func detachCheckout(commitId string) {
	targetCommit := gitlet.GetCommitById(commitId)
	if targetCommit == nil {
		fmt.Println("checkout: No such branch or commit.")
		return
	}

	currentCommit := gitlet.GetCommitById(gitlet.GetHEAD())
	for _, blobId := range currentCommit.BlobIds {
		blob := gitlet.GetBlobById(blobId)
		if blob != nil {
			utils.RemoveFileByPath(blob.FilePath)
		}
	}

	gitlet.DetachHEAD(commitId)
	for _, blobId := range targetCommit.BlobIds {
		blob := gitlet.GetBlobById(blobId)
		if blob != nil {
			utils.WriteFileBytes(blob.FilePath, blob.Contents)
		}
	}

	idx := gitlet.NewIndex()
	for k, v := range targetCommit.BlobIds {
		idx.Entries[k] = v
	}
	idx.Save()

	fmt.Printf("checkout: HEAD detached at %s.\n", commitId[:7])
}

func Branch(newBranchName string) {
	commitId := gitlet.GetHEAD()
	utils.WriteFile(config.BRANCHES+"/"+newBranchName, commitId)
	fmt.Printf("branch: Create Branch(%s).\n", newBranchName)
}

func RmBranch(targetBranchName string) {
	HEADBranch := gitlet.GetHEADBranch()
	if targetBranchName == HEADBranch {
		fmt.Println("rm-branch: You can't delete the current.")
	} else if !gitlet.BranchExist(targetBranchName) {
		fmt.Println("rm-branch: Target branch not exist.")
	} else {
		utils.RemoveFileByPath(config.BRANCHES + "/" + targetBranchName)
		fmt.Println("rm-branch: Remove success.")
	}
}

func Reset(cId string) {
	// Remove working tree files of the current commit
	commitId := gitlet.GetHEAD()
	commit := gitlet.GetCommitById(commitId)
	for _, blobId := range commit.BlobIds {
		blob := gitlet.GetBlobById(blobId)
		if blob != nil {
			utils.RemoveFileByPath(blob.FilePath)
		}
	}

	// Move branch pointer and restore working tree
	gitlet.MoveBranchPoint(cId)
	commitId = gitlet.GetHEAD()
	commit = gitlet.GetCommitById(commitId)
	for _, blobId := range commit.BlobIds {
		blob := gitlet.GetBlobById(blobId)
		if blob != nil {
			utils.WriteFileBytes(blob.FilePath, blob.Contents)
		}
	}

	// Sync index with the new commit
	idx := gitlet.NewIndex()
	for k, v := range commit.BlobIds {
		idx.Entries[k] = v
	}
	idx.Save()

	fmt.Printf("reset: HEAD at %s.\n", cId[:7])
}

func Merge(targetBranchName string) {
	currentBranch := gitlet.GetHEADBranch()

	if currentBranch == "" {
		fmt.Println("merge: Cannot merge in detached HEAD state.")
		return
	}
	if !gitlet.BranchExist(targetBranchName) {
		fmt.Println("merge: Target branch does not exist.")
		return
	}
	if currentBranch == targetBranchName {
		fmt.Println("merge: Cannot merge a branch with itself.")
		return
	}

	idx := gitlet.LoadIndex()
	headCommit := gitlet.GetCommitById(gitlet.GetHEAD())
	if !mapsEqual(idx.Entries, headCommit.BlobIds) {
		fmt.Println("merge: You have uncommitted changes.")
		return
	}

	splitPointId := getSplitPoint(currentBranch, targetBranchName)
	targetCommitId := string(utils.ReadFile(config.BRANCHES + "/" + targetBranchName))
	currentCommitId := gitlet.GetHEAD()

	if splitPointId == targetCommitId {
		fmt.Println("merge: Already up-to-date.")
		return
	}

	// Fast-forward: split point equals current HEAD
	if splitPointId == currentCommitId {
		for _, blobId := range headCommit.BlobIds {
			blob := gitlet.GetBlobById(blobId)
			if blob != nil && utils.FileExists(blob.FilePath) {
				utils.RemoveFileByPath(blob.FilePath)
			}
		}
		gitlet.MoveBranchPoint(targetCommitId)
		targetCommit := gitlet.GetCommitById(targetCommitId)
		for _, blobId := range targetCommit.BlobIds {
			blob := gitlet.GetBlobById(blobId)
			if blob != nil {
				utils.WriteFileBytes(blob.FilePath, blob.Contents)
			}
		}
		newIdx := gitlet.NewIndex()
		for k, v := range targetCommit.BlobIds {
			newIdx.Entries[k] = v
		}
		newIdx.Save()
		fmt.Println("merge: Fast-forward.")
		return
	}

	// Three-way merge
	splitCommit := gitlet.GetCommitById(splitPointId)
	targetCommit := gitlet.GetCommitById(targetCommitId)

	allFiles := make(map[string]bool)
	for f := range splitCommit.BlobIds {
		allFiles[f] = true
	}
	for f := range headCommit.BlobIds {
		allFiles[f] = true
	}
	for f := range targetCommit.BlobIds {
		allFiles[f] = true
	}

	newBlobIds := make(map[string]string)
	hasConflict := false

	for file := range allFiles {
		splitBlobId, inSplit := splitCommit.BlobIds[file]
		currentBlobId, inCurrent := headCommit.BlobIds[file]
		targetBlobId, inTarget := targetCommit.BlobIds[file]

		switch {
		case inSplit && inCurrent && inTarget:
			if splitBlobId == currentBlobId && splitBlobId == targetBlobId {
				newBlobIds[file] = currentBlobId
			} else if splitBlobId == currentBlobId {
				newBlobIds[file] = targetBlobId
			} else if splitBlobId == targetBlobId {
				newBlobIds[file] = currentBlobId
			} else if currentBlobId == targetBlobId {
				newBlobIds[file] = currentBlobId
			} else {
				hasConflict = true
				cb := writeConflict(file, currentBlobId, targetBlobId)
				newBlobIds[file] = cb.HashId
			}

		case !inSplit && inCurrent && inTarget:
			if currentBlobId == targetBlobId {
				newBlobIds[file] = currentBlobId
			} else {
				hasConflict = true
				cb := writeConflict(file, currentBlobId, targetBlobId)
				newBlobIds[file] = cb.HashId
			}

		case !inSplit && inCurrent && !inTarget:
			newBlobIds[file] = currentBlobId

		case !inSplit && !inCurrent && inTarget:
			newBlobIds[file] = targetBlobId

		case inSplit && !inCurrent && inTarget:
			if splitBlobId == targetBlobId {
				// Deleted in current, not modified in target → stay deleted
			} else {
				hasConflict = true
				cb := writeConflict(file, "", targetBlobId)
				newBlobIds[file] = cb.HashId
			}

		case inSplit && inCurrent && !inTarget:
			if splitBlobId == currentBlobId {
				// Deleted in target, not modified in current → delete
			} else {
				hasConflict = true
				cb := writeConflict(file, currentBlobId, "")
				newBlobIds[file] = cb.HashId
			}

		// inSplit && !inCurrent && !inTarget → both deleted, nothing to do
		}
	}

	// Remove old working tree files
	for _, blobId := range headCommit.BlobIds {
		blob := gitlet.GetBlobById(blobId)
		if blob != nil && utils.FileExists(blob.FilePath) {
			utils.RemoveFileByPath(blob.FilePath)
		}
	}

	// Write merged files
	newIdx := gitlet.NewIndex()
	for file, blobId := range newBlobIds {
		blob := gitlet.GetBlobById(blobId)
		if blob != nil {
			utils.WriteFileBytes(blob.FilePath, blob.Contents)
			newIdx.Update(file, blobId)
		}
	}
	newIdx.Save()

	// Create merge commit
	message := fmt.Sprintf("Merged %s into %s.", targetBranchName, currentBranch)
	commit := gitlet.NewMergeCommit(message, []string{currentCommitId, targetCommitId})
	commit.BlobIds = newBlobIds
	commit.Persist()
	gitlet.MoveBranchPoint(commit.HashId)

	if hasConflict {
		fmt.Println("merge: Encountered a merge conflict.")
	} else {
		fmt.Println("merge: Merge complete.")
	}
}

func writeConflict(filePath, currentBlobId, targetBlobId string) *gitlet.Blob {
	var currentContent, targetContent string
	if currentBlobId != "" {
		blob := gitlet.GetBlobById(currentBlobId)
		if blob != nil {
			currentContent = string(blob.Contents)
		}
	}
	if targetBlobId != "" {
		blob := gitlet.GetBlobById(targetBlobId)
		if blob != nil {
			targetContent = string(blob.Contents)
		}
	}
	if currentContent != "" && !strings.HasSuffix(currentContent, "\n") {
		currentContent += "\n"
	}
	if targetContent != "" && !strings.HasSuffix(targetContent, "\n") {
		targetContent += "\n"
	}
	content := "<<<<<<< HEAD\n" + currentContent + "=======\n" + targetContent + ">>>>>>>\n"
	blob := gitlet.NewBlob(filePath, []byte(content))
	blob.Persist()
	return blob
}

// getSplitPoint finds the latest common ancestor of two branches via BFS.
func getSplitPoint(branch1, branch2 string) string {
	commitId1 := string(utils.ReadFile(config.BRANCHES + "/" + branch1))
	commitId2 := string(utils.ReadFile(config.BRANCHES + "/" + branch2))

	// Collect all ancestors of branch1
	ancestors := make(map[string]bool)
	queue := []string{commitId1}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if ancestors[cur] {
			continue
		}
		ancestors[cur] = true
		c := gitlet.GetCommitById(cur)
		if c.Parent != nil {
			queue = append(queue, c.Parent...)
		}
	}

	// BFS from branch2; the first hit in ancestors is the split point
	queue = []string{commitId2}
	visited := make(map[string]bool)
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if visited[cur] {
			continue
		}
		visited[cur] = true
		if ancestors[cur] {
			return cur
		}
		c := gitlet.GetCommitById(cur)
		if c.Parent != nil {
			queue = append(queue, c.Parent...)
		}
	}
	return ""
}

func Diff(args ...string) {
	idx := gitlet.LoadIndex()

	if len(args) > 0 && args[0] == "--staged" {
		headCommit := gitlet.GetCommitById(gitlet.GetHEAD())

		// Files in index that differ from HEAD (staged additions / modifications)
		paths := sortedKeys(idx.Entries)
		for _, path := range paths {
			blobId := idx.Entries[path]
			headBlobId, inHead := headCommit.BlobIds[path]
			if !inHead {
				blob := gitlet.GetBlobById(blobId)
				if blob != nil {
					diffs := utils.DiffText("", string(blob.Contents))
					fmt.Print(utils.FormatDiff(path, diffs))
				}
			} else if headBlobId != blobId {
				headBlob := gitlet.GetBlobById(headBlobId)
				newBlob := gitlet.GetBlobById(blobId)
				if headBlob != nil && newBlob != nil {
					diffs := utils.DiffText(string(headBlob.Contents), string(newBlob.Contents))
					fmt.Print(utils.FormatDiff(path, diffs))
				}
			}
		}
		// Files removed from index (staged deletions)
		removedPaths := sortedKeys(headCommit.BlobIds)
		for _, path := range removedPaths {
			if !idx.Has(path) {
				headBlob := gitlet.GetBlobById(headCommit.BlobIds[path])
				if headBlob != nil {
					diffs := utils.DiffText(string(headBlob.Contents), "")
					fmt.Print(utils.FormatDiff(path, diffs))
				}
			}
		}
	} else {
		// Working tree vs index (unstaged changes)
		paths := sortedKeys(idx.Entries)
		for _, path := range paths {
			blobId := idx.Entries[path]
			if utils.FileExists(path) {
				content := utils.ReadFile(path)
				if utils.GenerateID(content) != blobId {
					indexBlob := gitlet.GetBlobById(blobId)
					if indexBlob != nil {
						diffs := utils.DiffText(string(indexBlob.Contents), string(content))
						fmt.Print(utils.FormatDiff(path, diffs))
					}
				}
			} else {
				indexBlob := gitlet.GetBlobById(blobId)
				if indexBlob != nil {
					diffs := utils.DiffText(string(indexBlob.Contents), "")
					fmt.Print(utils.FormatDiff(path, diffs))
				}
			}
		}
	}
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
