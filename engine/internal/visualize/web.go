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
	Title         string
	Hint          string
	GraphURL      string
	SaveLayoutURL string
	DotURL        string
	JSONURL       string
	ArtifactURL   string
	Stylesheet    string
	CytoscapeJS   string
	AppJS         string
	EmptyMessage  string
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
  <button id="refresh" type="button">Refresh</button>
  <button id="fit" type="button">Fit</button>
  {{ if .SaveLayoutURL }}<button id="save-layout" type="button">Save Layout</button>{{ end }}
  <a class="toolbar-link" href="{{ .JSONURL }}">Graph JSON</a>
  <a class="toolbar-link" href="{{ .DotURL }}">Graph DOT</a>
  <div class="meta" id="meta"></div>
</div>
<p class="hint">{{ .Hint }}</p>
<div id="graph"></div>
<script src="{{ .CytoscapeJS }}"></script>
<script src="{{ .AppJS }}"></script>
<script>
  document.getElementById('refresh').addEventListener('click', () => window.location.reload());
  TraceabilityUI.mount({
    graphUrl: {{ printf "%q" .GraphURL }},
    saveLayoutUrl: {{ printf "%q" .SaveLayoutURL }},
    artifactBaseUrl: {{ printf "%q" .ArtifactURL }},
    container: document.getElementById('graph'),
    fitButton: document.getElementById('fit'),
    saveButton: document.getElementById('save-layout'),
    metaElement: document.getElementById('meta'),
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
