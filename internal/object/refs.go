package gitlet

import (
	"gitlet/config"
	"gitlet/pkg/utils"
	"path/filepath"
	"strings"
)

func IsDetachedHEAD() bool {
	data := string(utils.ReadFile(config.HEAD))
	return !strings.HasPrefix(data, ".gitlet/")
}

func GetHEAD() string {
	data := string(utils.ReadFile(config.HEAD))
	if strings.HasPrefix(data, ".gitlet/") {
		return string(utils.ReadFile(data))
	}
	return data
}

func GetHEADBranch() string {
	data := string(utils.ReadFile(config.HEAD))
	if strings.HasPrefix(data, ".gitlet/") {
		return filepath.Base(data)
	}
	return ""
}

func MoveHEAD(branchName string) {
	branchPath := config.BRANCHES + "/" + branchName
	utils.WriteFile(config.HEAD, branchPath)
}

func DetachHEAD(commitId string) {
	utils.WriteFile(config.HEAD, commitId)
}

func MoveBranchPoint(commitId string) {
	data := string(utils.ReadFile(config.HEAD))
	if strings.HasPrefix(data, ".gitlet/") {
		utils.WriteFile(data, commitId)
	} else {
		utils.WriteFile(config.HEAD, commitId)
	}
}

func BranchExist(branchName string) bool {
	return utils.FileExists(config.BRANCHES + "/" + branchName)
}