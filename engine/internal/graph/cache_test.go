package graph

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

func TestRebuildCacheWritesSQLiteTables(t *testing.T) {
	cachePath := filepath.Join(t.TempDir(), "traceability.sqlite")
	g := &Graph{
		ManifestPath: "model/traceability/graph.yaml",
		Manifest:     manifestForGraph(&Graph{}),
		Relations: map[PartKind][]RelationEntry{
			PartKindDeriveReqt: {{Source: "product/alpha", Targets: []string{"model/requirements/alpha-requirement"}}},
			PartKindRefine:     {{Source: "model/requirements/alpha-requirement", Targets: []string{"model/use-cases/alpha-feature"}}},
		},
		Layout: []LayoutEntry{{ID: "model/use-cases/alpha-feature", X: 12.5, Y: 8.75, Locked: true}},
	}

	stats, err := RebuildCache(cachePath, g, false)
	if err != nil {
		t.Fatalf("RebuildCache() error = %v", err)
	}
	if stats.NodeCount != 3 || stats.EdgeCount != 2 || stats.LayoutCount != 1 {
		t.Fatalf("unexpected stats: %+v", stats)
	}

	db, err := sql.Open("sqlite", cachePath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	assertQueryCount(t, db, `SELECT COUNT(*) FROM nodes`, 3)
	assertQueryCount(t, db, `SELECT COUNT(*) FROM edges`, 2)
	assertQueryCount(t, db, `SELECT COUNT(*) FROM layout`, 1)
	assertQueryCount(t, db, `SELECT COUNT(*) FROM meta`, 3)

	var kind string
	if err := db.QueryRow(`SELECT kind FROM nodes WHERE id = ?`, "product/alpha").Scan(&kind); err != nil {
		t.Fatal(err)
	}
	if kind != "product-requirement" {
		t.Fatalf("node kind = %q, want product-requirement", kind)
	}
}

func TestRebuildCacheDryRunDoesNotCreateFile(t *testing.T) {
	cachePath := filepath.Join(t.TempDir(), "traceability.sqlite")
	g := &Graph{Relations: map[PartKind][]RelationEntry{
		PartKindDeriveReqt: {{Source: "product/alpha", Targets: []string{"model/requirements/alpha-requirement"}}},
	}}
	if _, err := RebuildCache(cachePath, g, true); err != nil {
		t.Fatalf("RebuildCache() dry-run error = %v", err)
	}
	if _, err := os.Stat(cachePath); !os.IsNotExist(err) {
		t.Fatalf("expected no cache file, stat err = %v", err)
	}
}

func assertQueryCount(t *testing.T, db *sql.DB, query string, want int) {
	t.Helper()
	var got int
	if err := db.QueryRow(query).Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("%s = %d, want %d", query, got, want)
	}
}
