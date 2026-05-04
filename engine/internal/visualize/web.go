package visualize

import (
	"embed"
	"html/template"
	"io"
	"io/fs"
)

//go:embed web/*
var webAssets embed.FS

type TraceabilityPageData struct {
	Title            string
	Hint             string
	GraphURL         string
	SaveRelationsURL string
	JSONURL          string
	ArtifactURL      string
	Stylesheet       string
	CytoscapeJS      string
	AppJS            string
	EmptyMessage     string
}

type ArtifactPageData struct {
	Title string
	Path  string
	Body  string
}

var traceabilityPageTemplate = template.Must(template.New("traceability-page").Parse(`<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{ .Title }}</title>
<link rel="stylesheet" href="{{ .Stylesheet }}">
</head>
<body class="traceability-shell">
<div class="toolbar">
  <button id="refresh" type="button" class="toolbar-icon-button toolbar-refresh-button" aria-label="Refresh" title="Refresh"><span class="details-visually-hidden">Refresh</span></button>
  <button id="fit" type="button" class="toolbar-icon-button toolbar-fit-button" aria-label="Fit" title="Fit"><span class="details-visually-hidden">Fit</span></button>
  <button id="zoom-in" type="button" class="toolbar-icon-button toolbar-zoom-in-button" aria-label="Zoom in" title="Zoom in"><span class="details-visually-hidden">Zoom in</span></button>
  <button id="zoom-out" type="button" class="toolbar-icon-button toolbar-zoom-out-button" aria-label="Zoom out" title="Zoom out"><span class="details-visually-hidden">Zoom out</span></button>
  <select id="layout-mode" aria-label="Layout mode">
    <option value="layered">Layered</option>
    <option value="organic">Organic</option>
    <option value="grid">Grid</option>
    <option value="clustered">Clustered</option>
  </select>
  <input type="search" id="filter" placeholder="Filter nodes…" aria-label="Filter nodes">
  {{ if .SaveRelationsURL }}<select id="relation-kind" aria-label="Relation kind">
    <option value="automatic">Automatic</option>
    <option value="deriveReqt">Derive Req.</option>
    <option value="satisfy">Satisfy</option>
    <option value="refine">Refine</option>
  </select>{{ end }}
  {{ if .SaveRelationsURL }}<button id="remove-edge" type="button" class="toolbar-icon-button toolbar-remove-edge-button" aria-label="Remove selected edge" title="Remove selected edge"><span class="details-visually-hidden">Remove selected edge</span></button>{{ end }}
  <a class="toolbar-link" href="{{ .JSONURL }}">Graph JSON</a>
  <div class="meta" id="meta"></div>
</div>
<p class="hint">{{ .Hint }}</p>
<div class="traceability-main">
  <div id="graph"></div>
  <aside class="details" id="details"><article class="details-panel"><p class="details-eyebrow">Inspector</p><h2 class="details-title">No selection</h2><p class="details-note">Select an edge to inspect relation details. Node taps still open their markdown artifacts.</p></article></aside>
</div>
<script src="{{ .CytoscapeJS }}"></script>
<script src="{{ .AppJS }}"></script>
<script>
  document.getElementById('refresh').addEventListener('click', () => window.location.reload());
  TraceabilityUI.mount({
    graphUrl: {{ printf "%q" .GraphURL }},
    saveRelationsUrl: {{ printf "%q" .SaveRelationsURL }},
    artifactBaseUrl: {{ printf "%q" .ArtifactURL }},
    container: document.getElementById('graph'),
    fitButton: document.getElementById('fit'),
    zoomInButton: document.getElementById('zoom-in'),
    zoomOutButton: document.getElementById('zoom-out'),
    layoutSelect: document.getElementById('layout-mode'),
    filterInput: document.getElementById('filter'),
    relationKindSelect: document.getElementById('relation-kind'),
    removeEdgeButton: document.getElementById('remove-edge'),
    metaElement: document.getElementById('meta'),
    detailsElement: document.getElementById('details'),
    emptyMessage: {{ printf "%q" .EmptyMessage }},
  });
</script>
</body>
</html>`))

var artifactPageTemplate = template.Must(template.New("traceability-artifact").Parse(`<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{ .Title }}</title>
<style>
  body {
    margin: 0;
    font-family: ui-monospace, SFMono-Regular, SFMono-Regular, Consolas, monospace;
    background: #111922;
    color: #dce8f2;
  }
  header {
    padding: 16px 20px 8px;
    border-bottom: 1px solid rgba(220, 232, 242, 0.14);
  }
  h1 {
    margin: 0 0 6px;
    font-size: 18px;
  }
  p {
    margin: 0;
    color: #90a5b7;
  }
  pre {
    margin: 0;
    padding: 20px;
    overflow: auto;
    white-space: pre-wrap;
  }
</style>
</head>
<body>
<header>
  <h1>{{ .Title }}</h1>
  <p>{{ .Path }}</p>
</header>
<pre>{{ .Body }}</pre>
</body>
</html>`))

func WebAssetFS() (fs.FS, error) {
	return fs.Sub(webAssets, "web")
}

func WriteTraceabilityPage(out io.Writer, data TraceabilityPageData) error {
	return traceabilityPageTemplate.Execute(out, data)
}

func WriteArtifactPage(out io.Writer, data ArtifactPageData) error {
	return artifactPageTemplate.Execute(out, data)
}
