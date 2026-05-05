package lint

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLinksInFile_FiltersURLsAndAnchors(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "doc.md")
	body := strings.Join([]string{
		"# Doc",
		"",
		"See [external](https://example.com) and [anchor](#section).",
		"Also [neighbour](./neighbour.md) and [parent](../up.md#frag).",
		"",
		"```markdown",
		"[ignored](should-not-match.md)",
		"```",
		"",
		"And [mailto](mailto:a@b.c).",
	}, "\n")
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := linksInFile(p)
	if err != nil {
		t.Fatalf("linksInFile: %v", err)
	}
	want := []string{"./neighbour.md", "../up.md#frag"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("link[%d]=%q want %q", i, got[i], want[i])
		}
	}
}

func TestStripFragment(t *testing.T) {
	cases := map[string]string{
		"a/b.md":         "a/b.md",
		"a/b.md#section": "a/b.md",
		"a/b.md?x=1":     "a/b.md",
		"a/b.md#s?x":     "a/b.md",
	}
	for in, want := range cases {
		if got := stripFragment(in); got != want {
			t.Errorf("stripFragment(%q)=%q want %q", in, got, want)
		}
	}
}
