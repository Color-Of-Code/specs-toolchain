package lint

import "testing"

func TestFormat_InsertsBlankLineBeforeOpeningFence(t *testing.T) {
	input := "## Section\n```text\nbody\n```\n"
	want := "## Section\n\n```text\nbody\n```\n"

	got := string(Format([]byte(input)))
	if got != want {
		t.Fatalf("Format() = %q, want %q", got, want)
	}
}

func TestFormat_DoesNotInsertBlankLineBeforeClosingFence(t *testing.T) {
	input := "```text\nbody\n```\n"
	want := "```text\nbody\n```\n"

	got := string(Format([]byte(input)))
	if got != want {
		t.Fatalf("Format() = %q, want %q", got, want)
	}
}
