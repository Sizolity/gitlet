package main

import (
	"fmt"
	"gitlet/internal/command"
	"gitlet/pkg/utils"
	"os"
)

func main() {
	args := os.Args[1:]

	if len(args) < 1 {
		fmt.Println("You need input at least an argument.")
		return
	}

	switch args[0] {
	case "init":
		command.Init_gitlet()
	case "add":
		if utils.GetArgsNum(args) == 2 {
			command.Add(args[1])
		} else {
			fmt.Println("add: You need input at least a file.")
		}
	case "commit":
		if utils.GetArgsNum(args) >= 2 {
			command.Commit(args[1:]...)
		} else {
			fmt.Println("commit: Get wrong argument num.")
		}
	case "rm":
		if utils.GetArgsNum(args) == 2 {
			command.Rm(args[1])
		} else {
			fmt.Println("rm: Get wrong argument num.")
		}
	case "log":
		if utils.GetArgsNum(args) == 1 {
			command.Log()
		} else {
			fmt.Println("log: Get wrong argument num.")
		}
	case "global-log":
		if utils.GetArgsNum(args) == 1 {
			command.GlobalLog()
		} else {
			fmt.Println("log: Get wrong argument num.")
		}
	case "find":
		if utils.GetArgsNum(args) >= 2 {
			command.Find(args[1:]...)
		} else {
			fmt.Println("find: Get wrong argument num.")
		}
	case "status":
		if utils.GetArgsNum(args) == 1 {
			command.Status()
		} else {
			fmt.Println("status: Get wrong argument num.")
		}
	case "checkout":
		if utils.GetArgsNum(args) >= 2 {
			command.Checkout(args[1:]...)
		} else {
			fmt.Println("checkout: Get wrong argument num.")
		}
	case "branch":
		if utils.GetArgsNum(args) == 2 {
			command.Branch(args[1])
		} else {
			fmt.Println("branch: Get wrong argument num.")
		}
	case "rm-branch":
		if utils.GetArgsNum(args) == 2 {
			command.RmBranch(args[1])
		} else {
			fmt.Println("rm-branch: Get wrong argument num.")
		}
	case "reset":
		if utils.GetArgsNum(args) == 2 {
			command.Reset(args[1])
		} else {
			fmt.Println("reset: Get wrong argument num.")
		}
	case "merge":
		if utils.GetArgsNum(args) == 2 {
			command.Merge(args[1])
		} else {
			fmt.Println("merge: Get wrong argument num.")
		}
	case "diff":
		command.Diff(args[1:]...)
	default:
		fmt.Println("Please input a valid instruction.")
		return
	}
}