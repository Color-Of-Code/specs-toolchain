// Visualize webview rendering the canonical traceability graph with Cytoscape.
import * as path from "path";
import * as os from "os";
import * as fs from "fs";
import * as vscode from "vscode";
import { runAndCapture, findSpecsFolder, findSpecsRoot, getOutput } from "./engine";

let currentPanel: vscode.WebviewPanel | undefined;
let suppressRefreshUntil = 0;

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
  summary?: string;
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

interface SaveRelationsEdge {
  source: string;
  target: string;
  kind: string;
}

interface SaveRelationsPayload {
  edges: SaveRelationsEdge[];
}

interface PendingRequest {
  resolve: () => void;
  reject: (reason: Error) => void;
}

export function registerVisualizePanel(context: vscode.ExtensionContext): void {
  context.subscriptions.push(
    vscode.commands.registerCommand("specs.visualize.preview", () => openOrRefresh(context)),
    vscode.commands.registerCommand("specs.visualize.previewIcon", () => openOrRefresh(context)),
    vscode.commands.registerCommand("specs.visualize.refresh", () => refresh(context)),
  );

  // Auto-refresh when markdown or canonical graph files change while the panel is open.
  const folder = findSpecsFolder();
  if (folder) {
    const root = findSpecsRoot(folder) ?? folder.uri.fsPath;
    const debounced = debounce(() => {
      if (currentPanel && Date.now() > suppressRefreshUntil) {
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
  panel.webview.onDidReceiveMessage(async (msg: { type: string; requestId?: string; payload?: string | SaveRelationsPayload }) => {
    if (msg.type === "export-json") {
      await exportFile(context, "json");
    } else if (msg.type === "refresh") {
      await refresh(context);
    } else if (msg.type === "open-path" && msg.payload) {
      await openArtifact(context, msg.payload as string);
    } else if (msg.type === "save-relations" && msg.requestId && msg.payload) {
      await saveGraphPayload(context, panel, msg.requestId, msg.payload as SaveRelationsPayload, "save-relations");
    }
  });
  await refresh(context);
}

async function saveGraphPayload(
  context: vscode.ExtensionContext,
  panel: vscode.WebviewPanel,
  requestId: string,
  payload: SaveRelationsPayload,
  subcommand: "save-relations",
): Promise<void> {
  const folder = findSpecsFolder();
  if (!folder) {
    await panel.webview.postMessage({ type: "save-relations-result", requestId, ok: false, error: "Specs: no workspace folder is open." });
    return;
  }
  const cwd = findSpecsRoot(folder) ?? folder.uri.fsPath;
  const tempDir = await fs.promises.mkdtemp(path.join(os.tmpdir(), "specs-relations-"));
  const inputPath = path.join(tempDir, "relations.json");
  try {
    await fs.promises.writeFile(inputPath, JSON.stringify(payload), "utf8");
    suppressRefreshUntil = Date.now() + 3000;
    const res = await runAndCapture(context, ["graph", subcommand, "--in", inputPath], cwd);
    if (res.exitCode !== 0) {
      const error = res.stderr || `graph ${subcommand} failed`;
      getOutput().appendLine(error);
      await panel.webview.postMessage({ type: `${subcommand}-result`, requestId, ok: false, error });
      return;
    }
    await panel.webview.postMessage({ type: `${subcommand}-result`, requestId, ok: true });
  } finally {
    await fs.promises.rm(tempDir, { recursive: true, force: true });
  }
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
	format: "json",
): Promise<void> {
  const folder = findSpecsFolder();
  if (!folder) {
    return;
  }
  const cwd = findSpecsRoot(folder) ?? folder.uri.fsPath;
	const ext = "json";
  const target = await vscode.window.showSaveDialog({
    defaultUri: vscode.Uri.joinPath(folder.uri, `traceability.${ext}`),
		filters: { "Traceability JSON": ["json"] },
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
  const appUri = webview.asWebviewUri(vscode.Uri.joinPath(mediaRoot, "traceability-view.js"));
  const styleUri = webview.asWebviewUri(vscode.Uri.joinPath(mediaRoot, "traceability-view.css"));
  const nonce = randomNonce();
  const csp = [
    `default-src 'none'`,
    `style-src ${webview.cspSource} 'unsafe-inline'`,
    `script-src 'nonce-${nonce}'`,
    `img-src ${webview.cspSource} data:`,
    `font-src ${webview.cspSource}`,
  ].join("; ");
  const safeGraph = escapeScriptText(JSON.stringify(graph));
  const requiredAssets = [
    path.join(context.extensionPath, "media", "cytoscape.min.js"),
    path.join(context.extensionPath, "media", "traceability-view.js"),
    path.join(context.extensionPath, "media", "traceability-view.css"),
  ];
  const fallbackInline = requiredAssets.some((current) => !fs.existsSync(current));
  const fallbackBanner = fallbackInline
    ? `<div class="banner">traceability web assets are not bundled - the diagram will not render. Run pnpm run compile in extension/.</div>`
    : "";

  return `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8" />
<meta http-equiv="Content-Security-Policy" content="${csp}">
<link rel="stylesheet" href="${styleUri}">
<style>
  body {
    color: var(--vscode-foreground);
    background: var(--vscode-editor-background);
  }
  .traceability-shell {
    --traceability-bg: var(--vscode-editor-background);
    --traceability-fg: var(--vscode-foreground);
    --traceability-link: var(--vscode-textLink-foreground);
    --traceability-muted: var(--vscode-descriptionForeground);
    --traceability-border: var(--vscode-panel-border);
    --traceability-button-bg: var(--vscode-button-background);
    --traceability-button-fg: var(--vscode-button-foreground);
    --traceability-button-hover: var(--vscode-button-hoverBackground);
    --traceability-warning-bg: var(--vscode-inputValidation-warningBackground);
    --traceability-warning-fg: var(--vscode-inputValidation-warningForeground);
    font-family: var(--vscode-font-family);
  }
</style>
</head>
<body class="traceability-shell">
${fallbackBanner}
<div class="toolbar">
  <button id="refresh">Refresh</button>
  <button id="fit">Fit</button>
  <button id="zoom-in" class="toolbar-icon-button toolbar-zoom-in-button" aria-label="Zoom in" title="Zoom in"><span class="details-visually-hidden">Zoom in</span></button>
  <button id="zoom-out" class="toolbar-icon-button toolbar-zoom-out-button" aria-label="Zoom out" title="Zoom out"><span class="details-visually-hidden">Zoom out</span></button>
  <select id="layout-mode" aria-label="Layout mode">
    <option value="layered">Layered</option>
    <option value="organic">Organic</option>
    <option value="grid">Grid</option>
  </select>
  <button id="relayout">Relayout</button>
  <input type="search" id="filter" placeholder="Filter nodes…" aria-label="Filter nodes">
  <select id="relation-kind" aria-label="Relation kind">
    <option value="realization">Realization</option>
    <option value="feature_implementation">Feature</option>
    <option value="component_implementation">Component</option>
    <option value="service_implementation">Service</option>
    <option value="api_implementation">API</option>
  </select>
  <button id="add-edge" class="toolbar-icon-button toolbar-add-edge-button" aria-label="Add edge" title="Add edge"><span class="details-visually-hidden">Add edge</span></button>
  <button id="remove-edge" class="toolbar-icon-button toolbar-remove-edge-button" aria-label="Remove selected edge" title="Remove selected edge"><span class="details-visually-hidden">Remove selected edge</span></button>
  <button id="export-json">Export JSON</button>
  <div class="meta" id="meta">${graph.nodes.length} nodes / ${graph.edges.length} edges</div>
</div>
<p class="hint">Select a node or edge to inspect its details. Use the inspector to open markdown artifacts.</p>
<div class="traceability-main">
  <div id="graph"></div>
  <aside class="details" id="details"><article class="details-panel"><p class="details-eyebrow">Inspector</p><h2 class="details-title">No selection</h2><p class="details-note">Select a node or edge to inspect its details.</p></article></aside>
</div>
${fallbackInline ? "" : `<script nonce="${nonce}" src="${cytoscapeUri}"></script><script nonce="${nonce}" src="${appUri}"></script>`}
<script nonce="${nonce}">
  const vscode = acquireVsCodeApi();
  const graph = ${safeGraph};
  let nextSaveRequestId = 0;
  const pendingRequests = new Map();
  window.addEventListener('message', (event) => {
    const message = event.data;
    if (!message || !message.requestId || message.type !== 'save-relations-result') {
      return;
    }
    const pending = pendingRequests.get(message.requestId);
    if (!pending) {
      return;
    }
    pendingRequests.delete(message.requestId);
    if (message.ok) {
      pending.resolve();
      return;
    }
    pending.reject(new Error(message.error || 'save relations failed'));
  });
  document.getElementById('refresh').addEventListener('click', () => vscode.postMessage({ type: 'refresh' }));
  document.getElementById('export-json').addEventListener('click', () => vscode.postMessage({ type: 'export-json' }));
  if (typeof TraceabilityUI !== 'undefined') {
    const ui = TraceabilityUI.mount({
      graph,
      container: document.getElementById('graph'),
      fitButton: document.getElementById('fit'),
      zoomInButton: document.getElementById('zoom-in'),
      zoomOutButton: document.getElementById('zoom-out'),
      layoutSelect: document.getElementById('layout-mode'),
      relayoutButton: document.getElementById('relayout'),
      filterInput: document.getElementById('filter'),
      addEdgeButton: document.getElementById('add-edge'),
      relationKindSelect: document.getElementById('relation-kind'),
      removeEdgeButton: document.getElementById('remove-edge'),
      metaElement: document.getElementById('meta'),
      detailsElement: document.getElementById('details'),
      onOpenPath: (path) => vscode.postMessage({ type: 'open-path', payload: path }),
      onSaveRelations: (payload) => new Promise((resolve, reject) => {
        const requestId = String(++nextSaveRequestId);
        pendingRequests.set(requestId, { resolve, reject });
        vscode.postMessage({ type: 'save-relations', requestId, payload });
      }),
      onConfirm: () => true,
      emptyMessage: 'No traceability data found.',
    });
    void ui;
  } else {
    document.getElementById('graph').innerHTML = '<pre style="padding: 16px; color: var(--vscode-errorForeground)">traceability UI failed to load</pre>';
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
