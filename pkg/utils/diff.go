package utils

import (
	"fmt"
	"strings"
)

type DiffLine struct {
	Op   byte // '+', '-', ' '
	Text string
}

// DiffText computes a line-level diff between two texts using LCS.
func DiffText(oldText, newText string) []DiffLine {
	oldLines := splitLines(oldText)
	newLines := splitLines(newText)

	m, n := len(oldLines), len(newLines)
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if oldLines[i-1] == newLines[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else if dp[i-1][j] >= dp[i][j-1] {
				dp[i][j] = dp[i-1][j]
			} else {
				dp[i][j] = dp[i][j-1]
			}
		}
	}

	var stack []DiffLine
	i, j := m, n
	for i > 0 || j > 0 {
		if i > 0 && j > 0 && oldLines[i-1] == newLines[j-1] {
			stack = append(stack, DiffLine{' ', oldLines[i-1]})
			i--
			j--
		} else if j > 0 && (i == 0 || dp[i][j-1] >= dp[i-1][j]) {
			stack = append(stack, DiffLine{'+', newLines[j-1]})
			j--
		} else {
			stack = append(stack, DiffLine{'-', oldLines[i-1]})
			i--
		}
	}

	result := make([]DiffLine, len(stack))
	for k := range stack {
		result[len(stack)-1-k] = stack[k]
	}
	return result
}

func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	lines := strings.Split(s, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// FormatDiff renders a diff with ANSI colors in unified-diff style.
func FormatDiff(filename string, diffs []DiffLine) string {
	hasChange := false
	for _, d := range diffs {
		if d.Op != ' ' {
			hasChange = true
			break
		}
	}
	if !hasChange {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(Colorize("--- a/"+filename, FgRed) + "\n")
	sb.WriteString(Colorize("+++ b/"+filename, FgGreen) + "\n")

	const ctx = 3
	lines := diffs
	n := len(lines)
	printed := make([]bool, n)

	for i := 0; i < n; i++ {
		if lines[i].Op != ' ' {
			lo := i - ctx
			if lo < 0 {
				lo = 0
			}
			hi := i + ctx
			if hi >= n {
				hi = n - 1
			}
			for k := lo; k <= hi; k++ {
				printed[k] = true
			}
		}
	}

	inBlock := false
	for i := 0; i < n; i++ {
		if !printed[i] {
			if inBlock {
				sb.WriteString(Colorize("...", FgYellow) + "\n")
				inBlock = false
			}
			continue
		}
		inBlock = true
		d := lines[i]
		switch d.Op {
		case '+':
			sb.WriteString(Colorize(fmt.Sprintf("+%s", d.Text), FgGreen) + "\n")
		case '-':
			sb.WriteString(Colorize(fmt.Sprintf("-%s", d.Text), FgRed) + "\n")
		default:
			sb.WriteString(fmt.Sprintf(" %s", d.Text) + "\n")
		}
	}
	return sb.String()
}
