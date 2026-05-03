// Visualize webview rendering the canonical traceability graph with Cytoscape.
import * as path from "path";
import * as fs from "fs";
import * as vscode from "vscode";
import { runAndCapture, findSpecsFolder, findSpecsRoot, getOutput } from "./engine";

let currentPanel: vscode.WebviewPanel | undefined;

interface VisualizeLayout {
  x: number;
  y: number;
  locked?: boolean;
}

interface VisualizeNode {
  id: string;
  path: string;
  label: string;
  kind: string;
  layout?: VisualizeLayout;
}

interface VisualizeEdge {
  source: string;
  target: string;
  kind: string;
}

interface VisualizeGraph {
  nodes: VisualizeNode[];
  edges: VisualizeEdge[];
}

export function registerVisualizePanel(context: vscode.ExtensionContext): void {
  context.subscriptions.push(
    vscode.commands.registerCommand("specs.visualize.preview", () => openOrRefresh(context)),
    vscode.commands.registerCommand("specs.visualize.refresh", () => refresh(context)),
  );

  // Auto-refresh when markdown or canonical graph files change while the panel is open.
  const folder = findSpecsFolder();
  if (folder) {
    const root = findSpecsRoot(folder) ?? folder.uri.fsPath;
    const debounced = debounce(() => {
      if (currentPanel) {
        refresh(context);
      }
    }, 500);
    for (const pattern of [
      "model/**/*.md",
      "product/**/*.md",
      "model/traceability/**/*.yaml",
      "model/traceability/**/*.yml",
      ".specs.yaml",
    ]) {
      const watcher = vscode.workspace.createFileSystemWatcher(new vscode.RelativePattern(root, pattern));
      watcher.onDidCreate(debounced);
      watcher.onDidChange(debounced);
      watcher.onDidDelete(debounced);
      context.subscriptions.push(watcher);
    }
  }
}

async function openOrRefresh(context: vscode.ExtensionContext): Promise<void> {
  if (currentPanel) {
    currentPanel.reveal(vscode.ViewColumn.Beside);
    await refresh(context);
    return;
  }
  const folder = findSpecsFolder();
  if (!folder) {
    vscode.window.showWarningMessage("Specs: no workspace folder is open.");
    return;
  }
  const mediaRoot = vscode.Uri.joinPath(context.extensionUri, "media");
  const panel = vscode.window.createWebviewPanel(
    "specs.visualize",
    "Specs: Traceability",
    vscode.ViewColumn.Beside,
    {
      enableScripts: true,
      retainContextWhenHidden: true,
      localResourceRoots: [mediaRoot],
    },
  );
  currentPanel = panel;
  panel.onDidDispose(() => {
    currentPanel = undefined;
  });
  panel.webview.onDidReceiveMessage(async (msg: { type: string; payload?: string }) => {
    if (msg.type === "export-dot") {
      await exportFile(context, "dot");
    } else if (msg.type === "export-json") {
      await exportFile(context, "json");
    } else if (msg.type === "refresh") {
      await refresh(context);
    } else if (msg.type === "open-path" && msg.payload) {
      await openArtifact(context, msg.payload);
    }
  });
  await refresh(context);
}

async function refresh(context: vscode.ExtensionContext): Promise<void> {
  if (!currentPanel) {
    return;
  }
  const folder = findSpecsFolder();
  if (!folder) {
    return;
  }
  const cwd = findSpecsRoot(folder) ?? folder.uri.fsPath;
  const res = await runAndCapture(
    context,
    ["visualize", "traceability", "--format", "json", "--out", "-"],
    cwd,
  );
  if (res.exitCode !== 0) {
    getOutput().appendLine(res.stderr);
    currentPanel.webview.html = errorHtml(res.stderr || "visualize failed");
    return;
  }
  let graph: VisualizeGraph;
  try {
    graph = JSON.parse(res.stdout) as VisualizeGraph;
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    getOutput().appendLine(`invalid visualize json: ${message}`);
    currentPanel.webview.html = errorHtml(`invalid visualize json: ${message}`);
    return;
  }
  currentPanel.webview.html = renderHtml(currentPanel.webview, context, graph);
}

async function exportFile(
  context: vscode.ExtensionContext,
  format: "dot" | "json",
): Promise<void> {
  const folder = findSpecsFolder();
  if (!folder) {
    return;
  }
  const cwd = findSpecsRoot(folder) ?? folder.uri.fsPath;
  const ext = format === "dot" ? "dot" : "json";
  const target = await vscode.window.showSaveDialog({
    defaultUri: vscode.Uri.joinPath(folder.uri, `traceability.${ext}`),
    filters: format === "dot" ? { "Graphviz DOT": ["dot"] } : { "Traceability JSON": ["json"] },
  });
  if (!target) {
    return;
  }
  const res = await runAndCapture(
    context,
    ["visualize", "traceability", "--format", format, "--out", target.fsPath],
    cwd,
  );
  if (res.exitCode !== 0) {
    vscode.window.showErrorMessage(`visualize failed: ${res.stderr}`);
    return;
  }
  vscode.window.showInformationMessage(`Wrote ${target.fsPath}`);
}

async function openArtifact(
  context: vscode.ExtensionContext,
  relativePath: string,
): Promise<void> {
  const folder = findSpecsFolder();
  if (!folder) {
    return;
  }
  const cwd = findSpecsRoot(folder) ?? folder.uri.fsPath;
  if (path.isAbsolute(relativePath)) {
    vscode.window.showErrorMessage(`Specs: expected a repo-relative path, got ${relativePath}`);
    return;
  }
  const targetPath = path.resolve(cwd, relativePath);
  const targetRel = path.relative(cwd, targetPath);
  if (targetRel.startsWith("..") || path.isAbsolute(targetRel)) {
    vscode.window.showErrorMessage(`Specs: refusing to open path outside specs root: ${relativePath}`);
    return;
  }
  const doc = await vscode.workspace.openTextDocument(vscode.Uri.file(targetPath));
  await vscode.window.showTextDocument(doc, { preview: true });
}

function renderHtml(
  webview: vscode.Webview,
  context: vscode.ExtensionContext,
  graph: VisualizeGraph,
): string {
  const mediaRoot = vscode.Uri.joinPath(context.extensionUri, "media");
  const cytoscapeUri = webview.asWebviewUri(vscode.Uri.joinPath(mediaRoot, "cytoscape.min.js"));
  const nonce = randomNonce();
  const csp = [
    `default-src 'none'`,
    `style-src ${webview.cspSource} 'unsafe-inline'`,
    `script-src 'nonce-${nonce}'`,
    `img-src ${webview.cspSource} data:`,
    `font-src ${webview.cspSource}`,
  ].join("; ");
  const safeGraph = escapeScriptText(JSON.stringify(graph));
  const fallbackInline = !fs.existsSync(
    path.join(context.extensionPath, "media", "cytoscape.min.js"),
  );
  const fallbackBanner = fallbackInline
    ? `<div class="banner">cytoscape.min.js not bundled - the diagram will not render. Run pnpm install and pnpm run compile in extension/.</div>`
    : "";

  return `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8" />
<meta http-equiv="Content-Security-Policy" content="${csp}">
<style>
  :root {
    color-scheme: light dark;
  }
  body {
    font-family: var(--vscode-font-family);
    padding: 12px;
    color: var(--vscode-foreground);
    background:
      radial-gradient(circle at top left, color-mix(in srgb, var(--vscode-textLink-foreground) 14%, transparent), transparent 28%),
      linear-gradient(180deg, color-mix(in srgb, var(--vscode-editor-background) 92%, white), var(--vscode-editor-background));
  }
  .toolbar { margin-bottom: 12px; display: flex; gap: 8px; align-items: center; flex-wrap: wrap; }
  button {
    background: var(--vscode-button-background);
    color: var(--vscode-button-foreground);
    border: none;
    padding: 6px 12px;
    cursor: pointer;
    border-radius: 999px;
  }
  button:hover { background: var(--vscode-button-hoverBackground); }
  .meta {
    margin-left: auto;
    font-size: 12px;
    color: var(--vscode-descriptionForeground);
  }
  .banner {
    background: var(--vscode-inputValidation-warningBackground);
    color: var(--vscode-inputValidation-warningForeground);
    padding: 6px 10px;
    margin-bottom: 12px;
    border-radius: 10px;
  }
  .hint {
    margin: 0 0 12px;
    font-size: 12px;
    color: var(--vscode-descriptionForeground);
  }
  #graph {
    height: calc(100vh - 140px);
    min-height: 420px;
    border-radius: 18px;
    border: 1px solid color-mix(in srgb, var(--vscode-panel-border) 70%, transparent);
    background: color-mix(in srgb, var(--vscode-editor-background) 94%, white);
    box-shadow: 0 10px 30px color-mix(in srgb, var(--vscode-editor-foreground) 10%, transparent);
  }
</style>
</head>
<body>
${fallbackBanner}
<div class="toolbar">
  <button id="refresh">Refresh</button>
  <button id="fit">Fit</button>
  <button id="export-json">Export JSON</button>
  <button id="export-dot">Export DOT</button>
  <div class="meta">${graph.nodes.length} nodes / ${graph.edges.length} edges</div>
</div>
<p class="hint">Click a node to open its markdown artifact. The preview reads canonical graph JSON from the engine.</p>
<div id="graph"></div>
${fallbackInline ? "" : `<script nonce="${nonce}" src="${cytoscapeUri}"></script>`}
<script nonce="${nonce}">
  const vscode = acquireVsCodeApi();
  const graph = ${safeGraph};
  const palette = {
    'product-requirement': '#e66b6b',
    requirement: '#4f8bd6',
    feature: '#e29c45',
    component: '#5f9d72',
    api: '#7b6ccf',
    service: '#c7739f',
  };
  function shapeForKind(kind) {
    switch (kind) {
      case 'product-requirement': return 'round-hexagon';
      case 'requirement': return 'round-rectangle';
      case 'feature': return 'ellipse';
      case 'component': return 'cut-rectangle';
      case 'api': return 'diamond';
      case 'service': return 'barrel';
      default: return 'round-rectangle';
    }
  }
  function lineStyleForKind(kind) {
    return kind === 'realization' ? 'solid' : 'dashed';
  }
  const elements = [
    ...graph.nodes.map((node) => ({
      data: { id: node.id, label: node.label, path: node.path, kind: node.kind },
      position: node.layout ? { x: node.layout.x, y: node.layout.y } : undefined,
      locked: Boolean(node.layout && node.layout.locked),
    })),
    ...graph.edges.map((edge, index) => ({
      data: { id: 'e' + index, source: edge.source, target: edge.target, kind: edge.kind },
    })),
  ];
  const hasLayout = graph.nodes.length > 0 && graph.nodes.every((node) => Boolean(node.layout));
  document.getElementById('refresh').addEventListener('click', () => vscode.postMessage({ type: 'refresh' }));
  document.getElementById('export-json').addEventListener('click', () => vscode.postMessage({ type: 'export-json' }));
  document.getElementById('export-dot').addEventListener('click', () => vscode.postMessage({ type: 'export-dot' }));
  if (typeof cytoscape !== 'undefined') {
    const cy = cytoscape({
      container: document.getElementById('graph'),
      elements,
      layout: hasLayout
        ? { name: 'preset', padding: 32, fit: true }
        : { name: 'breadthfirst', directed: true, padding: 40, spacingFactor: 1.15, avoidOverlap: true },
      style: [
        {
          selector: 'node',
          style: {
            label: 'data(label)',
            shape: (ele) => shapeForKind(ele.data('kind')),
            width: 'label',
            height: 'label',
            padding: '14px',
            'text-wrap': 'wrap',
            'text-max-width': '160px',
            'font-size': 12,
            'font-weight': 600,
            color: '#10222e',
            'text-valign': 'center',
            'text-halign': 'center',
            'border-width': 2,
            'border-color': '#173042',
            'background-color': (ele) => palette[ele.data('kind')] || '#7a8791',
          },
        },
        {
          selector: 'node:selected',
          style: {
            'border-width': 4,
            'border-color': '#f5f1c7',
            'overlay-opacity': 0,
          },
        },
        {
          selector: 'edge',
          style: {
            width: 2.2,
            'curve-style': 'bezier',
            'line-style': (ele) => lineStyleForKind(ele.data('kind')),
            'line-color': '#6d7f88',
            'target-arrow-color': '#6d7f88',
            'target-arrow-shape': 'triangle',
            'arrow-scale': 1.15,
          },
        },
      ],
    });
    document.getElementById('fit').addEventListener('click', () => cy.fit(undefined, 40));
    cy.on('tap', 'node', (event) => {
      const path = event.target.data('path');
      if (path) {
        vscode.postMessage({ type: 'open-path', payload: path });
      }
    });
  } else {
    document.getElementById('graph').innerHTML = '<pre style="padding: 16px; color: var(--vscode-errorForeground)">cytoscape failed to load</pre>';
  }
</script>
</body>
</html>`;
}

function errorHtml(message: string): string {
  return `<!doctype html><html><body><pre style="color: var(--vscode-errorForeground)">${escapeHtml(
    message,
  )}</pre></body></html>`;
}

function escapeHtml(s: string): string {
  return s
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;");
}

function escapeScriptText(s: string): string {
  return s
    .replace(/</g, "\\u003c")
    .replace(/>/g, "\\u003e")
    .replace(/&/g, "\\u0026")
    .replace(/\u2028/g, "\\u2028")
    .replace(/\u2029/g, "\\u2029");
}

function randomNonce(): string {
  let s = "";
  const cs = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";
  for (let i = 0; i < 24; i++) {
    s += cs.charAt(Math.floor(Math.random() * cs.length));
  }
  return s;
}

function debounce<T extends (...args: unknown[]) => void>(fn: T, ms: number): T {
  let t: NodeJS.Timeout | undefined;
  return ((...args: Parameters<T>) => {
    if (t) {
      clearTimeout(t);
    }
    t = setTimeout(() => fn(...args), ms);
  }) as T;
}
