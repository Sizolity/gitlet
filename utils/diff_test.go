package utils

import (
	"strings"
	"testing"
)

func hasDiffLine(diffs []DiffLine, op byte, text string) bool {
	for _, d := range diffs {
		if d.Op == op && d.Text == text {
			return true
		}
	}
	return false
}

func TestDiffTextWithEmptyInput(t *testing.T) {
	diffs := DiffText("", "")
	if len(diffs) != 0 {
		t.Fatalf("expected empty diff, got=%v", diffs)
	}
}

func TestDiffTextWithIdenticalContent(t *testing.T) {
	text := "a\nb\nc\n"
	diffs := DiffText(text, text)
	if len(diffs) != 3 {
		t.Fatalf("expected 3 context lines, got=%d", len(diffs))
	}
	for _, d := range diffs {
		if d.Op != ' ' {
			t.Fatalf("expected only context ops, got=%v", diffs)
		}
	}
}

func TestDiffTextWithMixedAddDelete(t *testing.T) {
	oldText := "a\nb\nc\n"
	newText := "a\nx\nc\nd\n"
	diffs := DiffText(oldText, newText)

	if !hasDiffLine(diffs, '-', "b") {
		t.Fatalf("expected deleted line b in diff: %v", diffs)
	}
	if !hasDiffLine(diffs, '+', "x") {
		t.Fatalf("expected added line x in diff: %v", diffs)
	}
	if !hasDiffLine(diffs, '+', "d") {
		t.Fatalf("expected added line d in diff: %v", diffs)
	}
}

func TestFormatDiffReturnsEmptyWhenNoChange(t *testing.T) {
	formatted := FormatDiff("a.txt", []DiffLine{{Op: ' ', Text: "same"}})
	if formatted != "" {
		t.Fatalf("expected empty formatted diff for no changes, got=%q", formatted)
	}
}

func TestFormatDiffRendersHeadersAndChanges(t *testing.T) {
	diffs := []DiffLine{
		{Op: ' ', Text: "a"},
		{Op: '-', Text: "b"},
		{Op: '+', Text: "x"},
	}
	formatted := FormatDiff("sample.txt", diffs)
	if !strings.Contains(formatted, "--- a/sample.txt") {
		t.Fatalf("missing old file header: %q", formatted)
	}
	if !strings.Contains(formatted, "+++ b/sample.txt") {
		t.Fatalf("missing new file header: %q", formatted)
	}
	if !strings.Contains(formatted, "-b") {
		t.Fatalf("missing removed line: %q", formatted)
	}
	if !strings.Contains(formatted, "+x") {
		t.Fatalf("missing added line: %q", formatted)
	}
}

