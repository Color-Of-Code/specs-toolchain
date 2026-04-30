package registry

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoad_MissingFile_EmptyRegistry(t *testing.T) {
	dir := t.TempDir()
	r, err := Load(filepath.Join(dir, "frameworks.yaml"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(r.Frameworks) != 0 {
		t.Errorf("want empty registry, got %d entries", len(r.Frameworks))
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "frameworks.yaml")
	r := &Registry{Frameworks: map[string]Entry{
		"default": {URL: "https://example.com/a.git", Ref: "v1"},
		"local":   {Path: "/tmp/fw"},
	}}
	if err := r.Save(path); err != nil {
		t.Fatalf("save: %v", err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got.Frameworks["default"].URL != "https://example.com/a.git" {
		t.Errorf("default URL mismatch: %+v", got.Frameworks["default"])
	}
	if got.Frameworks["local"].Path != "/tmp/fw" {
		t.Errorf("local path mismatch: %+v", got.Frameworks["local"])
	}
}

func TestEntryValidate(t *testing.T) {
	cases := []struct {
		name    string
		entry   Entry
		wantErr bool
	}{
		{"url only", Entry{URL: "u"}, false},
		{"url+ref", Entry{URL: "u", Ref: "main"}, false},
		{"path only", Entry{Path: "/p"}, false},
		{"empty", Entry{}, true},
		{"both url and path", Entry{URL: "u", Path: "/p"}, true},
		{"path+ref", Entry{Path: "/p", Ref: "main"}, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.entry.Validate()
			if c.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !c.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestAddRemoveResolve(t *testing.T) {
	r := &Registry{}
	if err := r.Add("a", Entry{URL: "u"}); err != nil {
		t.Fatalf("add: %v", err)
	}
	got, err := r.Resolve("a")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if got.URL != "u" {
		t.Errorf("URL mismatch: %+v", got)
	}
	if _, err := r.Resolve("missing"); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("expected os.ErrNotExist, got %v", err)
	}
	if err := r.Remove("a"); err != nil {
		t.Fatalf("remove: %v", err)
	}
	if err := r.Remove("a"); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("expected os.ErrNotExist on second remove, got %v", err)
	}
}

func TestNamesSorted(t *testing.T) {
	r := &Registry{Frameworks: map[string]Entry{
		"zebra": {URL: "u"},
		"alpha": {URL: "u"},
		"mango": {URL: "u"},
	}}
	got := r.Names()
	want := []string{"alpha", "mango", "zebra"}
	if len(got) != len(want) {
		t.Fatalf("length mismatch: %v", got)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("at %d: got %q want %q", i, got[i], want[i])
		}
	}
}

func TestLoad_InvalidEntry(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "frameworks.yaml")
	bad := []byte("frameworks:\n  bad:\n    url: u\n    path: /p\n")
	if err := os.WriteFile(path, bad, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Error("expected error for invalid entry")
	}
}

// withDefaultPath redirects user config lookup to a tempdir and writes the
// supplied registry. Lookup() reads the result.
func withDefaultPath(t *testing.T, r *Registry) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	t.Setenv("LocalAppData", filepath.Join(home, "AppData", "Local"))
	if r != nil {
		path, err := DefaultPath()
		if err != nil {
			t.Fatal(err)
		}
		if err := r.Save(path); err != nil {
			t.Fatal(err)
		}
	}
}

func TestLookup_RegisteredURLEntry(t *testing.T) {
	withDefaultPath(t, &Registry{Frameworks: map[string]Entry{
		"acme": {URL: "https://example.com/fw.git", Ref: "v1"},
	}})
	got, err := Lookup("acme")
	if err != nil {
		t.Fatalf("Lookup: %v", err)
	}
	if got.URL != "https://example.com/fw.git" || got.Ref != "v1" || got.Path != "" {
		t.Errorf("got %+v; want URL+Ref entry", got)
	}
}

func TestLookup_RegisteredPathEntry(t *testing.T) {
	withDefaultPath(t, &Registry{Frameworks: map[string]Entry{
		"local-dev": {Path: "/tmp/fw"},
	}})
	got, err := Lookup("local-dev")
	if err != nil {
		t.Fatalf("Lookup: %v", err)
	}
	if got.Path != "/tmp/fw" || got.URL != "" || got.Ref != "" {
		t.Errorf("got %+v; want Path entry", got)
	}
}

func TestLookup_RefOverride(t *testing.T) {
	withDefaultPath(t, &Registry{Frameworks: map[string]Entry{
		"acme": {URL: "https://example.com/fw.git", Ref: "v1"},
	}})
	got, err := Lookup("acme@v2")
	if err != nil {
		t.Fatalf("Lookup: %v", err)
	}
	if got.Ref != "v2" {
		t.Errorf("ref override not applied: %+v", got)
	}
}

func TestLookup_RefOverrideOnPathEntryFails(t *testing.T) {
	withDefaultPath(t, &Registry{Frameworks: map[string]Entry{
		"local-dev": {Path: "/tmp/fw"},
	}})
	if _, err := Lookup("local-dev@main"); err == nil {
		t.Error("expected error for @ref on path-based entry")
	}
}

func TestLookup_EmptySpecResolvesDefault(t *testing.T) {
	withDefaultPath(t, &Registry{Frameworks: map[string]Entry{
		"default": {URL: "https://example.com/default.git"},
	}})
	got, err := Lookup("")
	if err != nil {
		t.Fatalf("Lookup: %v", err)
	}
	if got.URL != "https://example.com/default.git" {
		t.Errorf("default lookup failed: %+v", got)
	}
}

func TestLookup_UnknownNameFails(t *testing.T) {
	withDefaultPath(t, &Registry{})
	_, err := Lookup("nope")
	if err == nil {
		t.Fatal("expected error for unknown framework")
	}
	if !strings.Contains(err.Error(), "specs framework add") {
		t.Errorf("expected hint about `specs framework add` in error: %v", err)
	}
}

func TestLookup_EmptySpecWithEmptyRegistryFails(t *testing.T) {
	withDefaultPath(t, &Registry{})
	_, err := Lookup("")
	if err == nil {
		t.Fatal("expected error when no default is registered")
	}
	if !strings.Contains(err.Error(), `"default"`) {
		t.Errorf("expected error to mention 'default': %v", err)
	}
}
