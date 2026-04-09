package command

import (
	"fmt"
	"gitlet/config"
	gitlet "gitlet/internal/object"
	"gitlet/pkg/utils"
)

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

	// Remove working tree files of the current commit.
	commitId := gitlet.GetHEAD()
	commit := gitlet.GetCommitById(commitId)
	for _, blobId := range commit.BlobIds {
		blob := gitlet.GetBlobById(blobId)
		if blob != nil {
			utils.RemoveFileByPath(blob.FilePath)
		}
	}

	// Switch HEAD and restore working tree.
	gitlet.MoveHEAD(branchName)
	commitId = gitlet.GetHEAD()
	commit = gitlet.GetCommitById(commitId)
	for _, blobId := range commit.BlobIds {
		blob := gitlet.GetBlobById(blobId)
		if blob != nil {
			utils.WriteFileBytes(blob.FilePath, blob.Contents)
		}
	}

	// Sync index with the new commit.
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
	// Remove working tree files of the current commit.
	commitId := gitlet.GetHEAD()
	commit := gitlet.GetCommitById(commitId)
	for _, blobId := range commit.BlobIds {
		blob := gitlet.GetBlobById(blobId)
		if blob != nil {
			utils.RemoveFileByPath(blob.FilePath)
		}
	}

	// Move branch pointer and restore working tree.
	gitlet.MoveBranchPoint(cId)
	commitId = gitlet.GetHEAD()
	commit = gitlet.GetCommitById(commitId)
	for _, blobId := range commit.BlobIds {
		blob := gitlet.GetBlobById(blobId)
		if blob != nil {
			utils.WriteFileBytes(blob.FilePath, blob.Contents)
		}
	}

	// Sync index with the new commit.
	idx := gitlet.NewIndex()
	for k, v := range commit.BlobIds {
		idx.Entries[k] = v
	}
	idx.Save()

	fmt.Printf("reset: HEAD at %s.\n", cId[:7])
}
