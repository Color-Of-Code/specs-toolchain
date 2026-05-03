package lint

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckFileStyle_BlankLinesAroundFencesUsesOpeningFenceLine(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "doc.md")
	body := "# Doc\n\n```text\nbody\n```\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &StyleRules{BlankLinesAroundFences: true}
	violations := CheckFileStyle(path, cfg)
	for _, violation := range violations {
		if violation.Rule == "blank_lines_around_fences" {
			t.Fatalf("unexpected fence violation: %+v", violation)
		}
	}
}
