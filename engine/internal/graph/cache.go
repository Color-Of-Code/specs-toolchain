package graph

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type CacheStats struct {
	CachePath     string
	NodeCount     int
	EdgeCount     int
	BaselineCount int
	LayoutCount   int
}

func RebuildCache(cachePath string, g *Graph, dryRun bool) (*CacheStats, error) {
	if g == nil {
		return nil, fmt.Errorf("graph is nil")
	}
	absCachePath, err := filepath.Abs(cachePath)
	if err != nil {
		return nil, err
	}
	stats := &CacheStats{
		CachePath:     absCachePath,
		NodeCount:     len(g.NodeIDs()),
		EdgeCount:     relationEntryCount(g.DeriveReqt) + relationEntryCount(g.Satisfactions) + relationEntryCount(g.Refinements),
		BaselineCount: len(g.Baselines),
		LayoutCount:   len(g.Layout),
	}
	if dryRun {
		return stats, nil
	}
	if err := os.MkdirAll(filepath.Dir(absCachePath), 0o755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", absCachePath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	for _, statement := range []string{
		`PRAGMA foreign_keys = ON`,
		`DROP TABLE IF EXISTS meta`,
		`DROP TABLE IF EXISTS nodes`,
		`DROP TABLE IF EXISTS edges`,
		`DROP TABLE IF EXISTS baselines`,
		`DROP TABLE IF EXISTS layout`,
		`CREATE TABLE meta (key TEXT PRIMARY KEY, value TEXT NOT NULL)`,
		`CREATE TABLE nodes (id TEXT PRIMARY KEY, kind TEXT NOT NULL, path TEXT NOT NULL)`,
		`CREATE TABLE edges (kind TEXT NOT NULL, source_id TEXT NOT NULL, target_id TEXT NOT NULL, PRIMARY KEY (kind, source_id, target_id))`,
		`CREATE TABLE baselines (component_id TEXT PRIMARY KEY, repo TEXT NOT NULL, path TEXT NOT NULL, commit_sha TEXT NOT NULL)`,
		`CREATE TABLE layout (node_id TEXT PRIMARY KEY, x REAL NOT NULL, y REAL NOT NULL, locked INTEGER NOT NULL DEFAULT 0)`,
		`CREATE INDEX edges_source_idx ON edges (source_id, kind)`,
		`CREATE INDEX edges_target_idx ON edges (target_id, kind)`,
		`CREATE INDEX nodes_kind_idx ON nodes (kind)`,
	} {
		if _, err := tx.Exec(statement); err != nil {
			return nil, err
		}
	}

	manifest := g.Manifest
	if manifest.SchemaVersion == 0 {
		manifest = manifestForGraph(g)
	}
	metaStmt, err := tx.Prepare(`INSERT INTO meta (key, value) VALUES (?, ?)`)
	if err != nil {
		return nil, err
	}
	defer metaStmt.Close()
	for key, value := range map[string]string{
		"graph_manifest_path": g.ManifestPath,
		"schema_version":      fmt.Sprintf("%d", manifest.SchemaVersion),
		"node_id_format":      manifest.NodeIDFormat,
	} {
		if _, err := metaStmt.Exec(key, value); err != nil {
			return nil, err
		}
	}

	nodeStmt, err := tx.Prepare(`INSERT INTO nodes (id, kind, path) VALUES (?, ?, ?)`)
	if err != nil {
		return nil, err
	}
	defer nodeStmt.Close()
	for _, nodeID := range g.NodeIDs() {
		if _, err := nodeStmt.Exec(nodeID, KindForNodeID(nodeID), MarkdownPath(nodeID)); err != nil {
			return nil, err
		}
	}

	edgeStmt, err := tx.Prepare(`INSERT INTO edges (kind, source_id, target_id) VALUES (?, ?, ?)`)
	if err != nil {
		return nil, err
	}
	defer edgeStmt.Close()
	for _, group := range []struct {
		kind    PartKind
		entries []RelationEntry
	}{
		{kind: PartKindDeriveReqt, entries: g.DeriveReqt},
		{kind: PartKindRefine, entries: g.Refinements},
		{kind: PartKindSatisfy, entries: g.Satisfactions},
	} {
		for _, entry := range group.entries {
			for _, target := range entry.Targets {
				if _, err := edgeStmt.Exec(group.kind, entry.Source, target); err != nil {
					return nil, err
				}
			}
		}
	}

	baselineStmt, err := tx.Prepare(`INSERT INTO baselines (component_id, repo, path, commit_sha) VALUES (?, ?, ?, ?)`)
	if err != nil {
		return nil, err
	}
	defer baselineStmt.Close()
	for _, entry := range g.Baselines {
		if _, err := baselineStmt.Exec(entry.Component, entry.Repo, entry.Path, entry.Commit); err != nil {
			return nil, err
		}
	}

	layoutStmt, err := tx.Prepare(`INSERT INTO layout (node_id, x, y, locked) VALUES (?, ?, ?, ?)`)
	if err != nil {
		return nil, err
	}
	defer layoutStmt.Close()
	for _, entry := range g.Layout {
		locked := 0
		if entry.Locked {
			locked = 1
		}
		if _, err := layoutStmt.Exec(entry.ID, entry.X, entry.Y, locked); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return stats, nil
}

func relationEntryCount(entries []RelationEntry) int {
	total := 0
	for _, entry := range entries {
		total += len(entry.Targets)
	}
	return total
}
