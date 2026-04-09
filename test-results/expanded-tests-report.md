# Expanded Tests Report

- Date: 2026-04-09
- Command: `go test ./...`
- Result: PASS

## Test Output

```text
?   	gitlet	[no test files]
?   	gitlet/config	[no test files]
ok  	gitlet/gitlet	0.002s
ok  	gitlet/instruction	0.003s
ok  	gitlet/utils	0.001s
```

## Added Coverage

- `gitlet/refs_commit_test.go`
  - `refs` branch/detached transitions (`GetHEAD`, `GetHEADBranch`, `MoveBranchPoint`, `DetachHEAD`, `BranchExist`)
  - `commit` persistence and tree flatten hydration (`GetCommitById`, `GetAllCommits`)
  - merge commit parent behavior (`NewMergeCommit`)
- `instruction/instruction_edge_test.go`
  - detached HEAD merge rejection
  - self-merge rejection
  - detached status output
  - checkout by commit detaches HEAD
  - reset restores old commit and worktree content
- `utils/diff_test.go`
  - empty/identical/mixed add-delete diff behavior
  - formatted diff header and change rendering
- `utils/io_test.go`
  - nested path write with parent auto-create
  - file removal with empty parent directory cleanup

## Saved Artifacts

- `test-results/expanded-tests.txt`
- `test-results/expanded-tests-report.md`
