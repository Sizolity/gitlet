package command

import (
	"fmt"
	"gitlet/config"
	gitlet "gitlet/internal/object"
	"gitlet/pkg/utils"
	"sort"
	"strings"
)

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

	// Staged files: in index with different blob than HEAD, or not in HEAD at all.
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

	// Removed files: in HEAD but not in index.
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

	// Modifications not staged for commit.
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

	// Untracked files.
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

	// Fast-forward: split point equals current HEAD.
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

	// Three-way merge.
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
				// Deleted in current, not modified in target -> stay deleted.
			} else {
				hasConflict = true
				cb := writeConflict(file, "", targetBlobId)
				newBlobIds[file] = cb.HashId
			}

		case inSplit && inCurrent && !inTarget:
			if splitBlobId == currentBlobId {
				// Deleted in target, not modified in current -> delete.
			} else {
				hasConflict = true
				cb := writeConflict(file, currentBlobId, "")
				newBlobIds[file] = cb.HashId
			}

			// inSplit && !inCurrent && !inTarget -> both deleted, nothing to do.
		}
	}

	// Remove old working tree files.
	for _, blobId := range headCommit.BlobIds {
		blob := gitlet.GetBlobById(blobId)
		if blob != nil && utils.FileExists(blob.FilePath) {
			utils.RemoveFileByPath(blob.FilePath)
		}
	}

	// Write merged files.
	newIdx := gitlet.NewIndex()
	for file, blobId := range newBlobIds {
		blob := gitlet.GetBlobById(blobId)
		if blob != nil {
			utils.WriteFileBytes(blob.FilePath, blob.Contents)
			newIdx.Update(file, blobId)
		}
	}
	newIdx.Save()

	// Create merge commit.
	message := fmt.Sprintf("Merged %s into %s.", targetBranchName, currentBranch)
	commit := gitlet.NewMergeCommit(message, []string{currentCommitId, targetCommitId})
	commit.TreeId = gitlet.BuildTree(newBlobIds)
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

	// Collect all ancestors of branch1.
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

	// BFS from branch2; first hit in ancestors is split point.
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

		// Files in index that differ from HEAD (staged additions/modifications).
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
		// Files removed from index (staged deletions).
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
		// Working tree vs index (unstaged changes).
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
