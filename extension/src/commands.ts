// Palette wrappers around engine subcommands.
import * as vscode from "vscode";
import { runInTerminal, runAndCapture, findSpecsFolder, findSpecsRoot, getOutput } from "./engine";
import { runInitWizard } from "./initWizard";

type ScaffoldKind = "requirement" | "feature" | "component" | "api" | "service";

// Palette commands that just shell out to the engine. Custom-handler commands
// (wizards, scaffolders, visualize) are registered separately below.
const TERMINAL_COMMANDS: ReadonlyArray<readonly [string, readonly string[]]> = [
  ["specs.lint", ["lint"]],
  ["specs.lint.links", ["lint", "--links"]],
  ["specs.lint.style", ["lint", "--style"]],
  ["specs.lint.baselines", ["lint", "--baselines"]],
  ["specs.doctor", ["doctor"]],
  ["specs.frameworkUpdate", ["framework", "update"]],
  ["specs.cr.status", ["cr", "status"]],
];

export function registerCommands(context: vscode.ExtensionContext): void {
  const reg = (id: string, fn: () => void | Promise<void>) =>
    context.subscriptions.push(vscode.commands.registerCommand(id, fn));

  for (const [id, args] of TERMINAL_COMMANDS) {
    reg(id, () => runTerminal(context, [...args]));
  }

  // Init wizard.
  reg("specs.bootstrap", () => runInitWizard(context));

  // Visualize (writes a file in model/ and opens it).
  reg("specs.visualize.dot", () => visualize(context, "dot"));
  reg("specs.visualize.mermaid", () => visualize(context, "mermaid"));

  // Scaffold a new model file.
  for (const kind of ["requirement", "feature", "component", "api", "service"] as const) {
    reg(`specs.scaffold.${kind}`, () => scaffold(context, kind));
  }

  // Change-requests.
  reg("specs.cr.new", () => crNew(context));
  reg("specs.cr.drain", () => crDrain(context));

  // Baseline.
  reg("specs.baseline.update", () => baselineUpdate(context));

  // Framework registry.
  reg("specs.registerFrameworks", () => registerFrameworks(context));
}

// --- helpers ---

function runTerminal(context: vscode.ExtensionContext, args: string[]): void {
  const folder = findSpecsFolder();
  if (!folder) {
    vscode.window.showWarningMessage("Specs: no workspace folder is open.");
    return;
  }
  const cwd = findSpecsRoot(folder) ?? folder.uri.fsPath;
  runInTerminal(context, args, cwd);
}

async function scaffold(context: vscode.ExtensionContext, kind: ScaffoldKind): Promise<void> {
  const folder = findSpecsFolder();
  if (!folder) {
    vscode.window.showWarningMessage("Specs: no workspace folder is open.");
    return;
  }
  const cwd = findSpecsRoot(folder) ?? folder.uri.fsPath;

  const placeholder =
    kind === "requirement"
      ? "e.g. core/012-some-requirement"
      : kind === "feature" || kind === "component"
      ? "e.g. core/some-slug"
      : "e.g. some-slug";
  const relPath = await vscode.window.showInputBox({
    prompt: `Relative path under model/${kind}s/ (without .md)`,
    placeHolder: placeholder,
    ignoreFocusOut: true,
    validateInput: (v) => (v.trim() ? null : "path is required"),
  });
  if (!relPath) {
    return;
  }
  const title = await vscode.window.showInputBox({
    prompt: `${kind} title (optional — defaults to derived from slug)`,
    placeHolder: "Short, descriptive title",
    ignoreFocusOut: true,
  });
  if (title === undefined) {
    return;
  }
  const args = ["scaffold", kind];
  if (title) {
    args.push("--title", title);
  }
  args.push(relPath);
  runInTerminal(context, args, cwd);
}

async function crNew(context: vscode.ExtensionContext): Promise<void> {
  const folder = findSpecsFolder();
  if (!folder) {
    return;
  }
  const cwd = findSpecsRoot(folder) ?? folder.uri.fsPath;

  const id = await vscode.window.showInputBox({
    prompt: "Change-request id (e.g. 042)",
    ignoreFocusOut: true,
    validateInput: (v) => (/^\d+$/.test(v) ? null : "expected a number"),
  });
  if (!id) {
    return;
  }
  const slug = await vscode.window.showInputBox({
    prompt: "Change-request slug (kebab-case)",
    placeHolder: "e.g. add-login-flow",
    ignoreFocusOut: true,
    validateInput: (v) => (/^[a-z0-9]+(-[a-z0-9]+)*$/.test(v) ? null : "expected kebab-case"),
  });
  if (!slug) {
    return;
  }
  const title = await vscode.window.showInputBox({
    prompt: "Change-request title (optional — defaults to derived from slug)",
    ignoreFocusOut: true,
  });
  if (title === undefined) {
    return;
  }
  const args = ["cr", "new", "--id", id, "--slug", slug];
  if (title) {
    args.push("--title", title);
  }
  runInTerminal(context, args, cwd);
}

async function crDrain(context: vscode.ExtensionContext): Promise<void> {
  const folder = findSpecsFolder();
  if (!folder) {
    return;
  }
  const cwd = findSpecsRoot(folder) ?? folder.uri.fsPath;

  const id = await vscode.window.showInputBox({
    prompt: "Change-request id (e.g. 042)",
    ignoreFocusOut: true,
    validateInput: (v) => (/^\d+$/.test(v) ? null : "expected a number"),
  });
  if (!id) {
    return;
  }
  runInTerminal(context, ["cr", "drain", "--id", id], cwd);
}

async function baselineUpdate(context: vscode.ExtensionContext): Promise<void> {
  const folder = findSpecsFolder();
  if (!folder) {
    return;
  }
  const cwd = findSpecsRoot(folder) ?? folder.uri.fsPath;

  const only = await vscode.window.showInputBox({
    prompt: "Component filter (substring; leave empty for all)",
    ignoreFocusOut: true,
  });
  const args = ["baseline", "update"];
  if (only) {
    args.push("--only", only);
  }
  runInTerminal(context, args, cwd);
}

async function visualize(
  context: vscode.ExtensionContext,
  format: "dot" | "mermaid",
): Promise<void> {
  const folder = findSpecsFolder();
  if (!folder) {
    return;
  }
  const cwd = findSpecsRoot(folder) ?? folder.uri.fsPath;

  const ext = format === "dot" ? "dot" : "md";
  const outPath = `model/_visualize.${ext}`;
  const out = getOutput();
  out.appendLine(`Specs: visualize -> ${outPath}`);

  const res = await runAndCapture(
    context,
    ["visualize", "traceability", "--format", format, "--out", outPath],
    cwd,
  );
  if (res.exitCode !== 0) {
    out.appendLine(res.stderr);
    out.appendLine(res.stdout);
    out.show(true);
    vscode.window.showErrorMessage(
      `specs visualize failed (exit ${res.exitCode}). See Specs output for details.`,
    );
    return;
  }
  const uri = vscode.Uri.joinPath(folder.uri, outPath);
  const doc = await vscode.workspace.openTextDocument(uri);
  await vscode.window.showTextDocument(doc, { preview: true });
}

interface FrameworkEntry {
  name: string;
  url?: string;
  ref?: string;
  path?: string;
}

/**
 * Reads specs.frameworks from settings and calls `specs framework add` for
 * each entry, so the user-level registry stays in sync with their VS Code
 * configuration without manual terminal commands.
 */
async function registerFrameworks(context: vscode.ExtensionContext): Promise<void> {
  const cfg = vscode.workspace.getConfiguration("specs");
  const entries = cfg.get<FrameworkEntry[]>("frameworks", []);

  if (entries.length === 0) {
    vscode.window.showInformationMessage(
      "No frameworks configured. Add entries to 'specs.frameworks' in Settings.",
    );
    return;
  }

  const folder = findSpecsFolder() ?? vscode.workspace.workspaceFolders?.[0];
  if (!folder) {
    vscode.window.showWarningMessage("Specs: no workspace folder is open.");
    return;
  }
  const cwd = findSpecsRoot(folder) ?? folder.uri.fsPath;

  for (const entry of entries) {
    if (!entry.name) {
      continue;
    }
    const args = ["framework", "add", entry.name];
    if (entry.url) {
      args.push("--url", entry.url);
      if (entry.ref) {
        args.push("--ref", entry.ref);
      }
    } else if (entry.path) {
      args.push("--path", entry.path);
    } else {
      continue; // neither url nor path — skip silently
    }
    runInTerminal(context, args, cwd);
  }
}
