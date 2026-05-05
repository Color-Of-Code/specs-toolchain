// Palette wrappers around engine subcommands.
import * as vscode from "vscode";
import { runInTerminal, runAndCapture, getSpecsExecutionTarget, getOutput } from "./engine";
import { runInitWizard } from "./initWizard";

type ScaffoldKind = "requirement" | "use-case" | "component";

// Palette commands that just shell out to the engine. Custom-handler commands
// (wizards, scaffolders, visualize) are registered separately below.
const TERMINAL_COMMANDS: ReadonlyArray<readonly [string, readonly string[]]> = [
  ["specs.lint", ["lint"]],
  ["specs.lint.links", ["lint", "--links"]],
  ["specs.lint.style", ["lint", "--style"]],
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
	reg("specs.visualize.mermaid", () => visualize(context));

  // Scaffold a new model file.
  for (const kind of ["requirement", "use-case", "component"] as const) {
    reg(`specs.scaffold.${kind}`, () => scaffold(context, kind));
  }

  // Change-requests.
  reg("specs.cr.new", () => crNew(context));
  reg("specs.cr.drain", () => crDrain(context));
}

// --- helpers ---

function runTerminal(context: vscode.ExtensionContext, args: string[]): void {
  const target = getSpecsExecutionTarget({ warnIfMissing: true });
  if (!target) {
    return;
  }
  runInTerminal(context, args, target.cwd);
}

async function scaffold(context: vscode.ExtensionContext, kind: ScaffoldKind): Promise<void> {
  const target = getSpecsExecutionTarget({ warnIfMissing: true });
  if (!target) {
    return;
  }

  const placeholder =
    kind === "requirement"
      ? "e.g. core/012-some-requirement"
      : kind === "use-case" || kind === "component"
      ? "e.g. core/some-slug"
      : "e.g. some-slug";
  const relPath = await vscode.window.showInputBox({
    prompt: `Relative path under model/${kind === "use-case" ? "use-cases" : kind + "s"}/ (without .md)`,
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
  runInTerminal(context, args, target.cwd);
}

async function crNew(context: vscode.ExtensionContext): Promise<void> {
  const target = getSpecsExecutionTarget({ warnIfMissing: true });
  if (!target) {
    return;
  }

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
  runInTerminal(context, args, target.cwd);
}

async function crDrain(context: vscode.ExtensionContext): Promise<void> {
  const target = getSpecsExecutionTarget({ warnIfMissing: true });
  if (!target) {
    return;
  }

  const id = await vscode.window.showInputBox({
    prompt: "Change-request id (e.g. 042)",
    ignoreFocusOut: true,
    validateInput: (v) => (/^\d+$/.test(v) ? null : "expected a number"),
  });
  if (!id) {
    return;
  }
  runInTerminal(context, ["cr", "drain", "--id", id], target.cwd);
}

async function visualize(context: vscode.ExtensionContext): Promise<void> {
  const target = getSpecsExecutionTarget({ warnIfMissing: true });
  if (!target) {
    return;
  }

	const outPath = "model/_visualize.md";
  const out = getOutput();
  out.appendLine(`Specs: visualize -> ${outPath}`);

  const res = await runAndCapture(
    context,
		["visualize", "traceability", "--format", "mermaid", "--out", outPath],
    target.cwd,
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
  const uri = vscode.Uri.joinPath(target.folder.uri, outPath);
  const doc = await vscode.workspace.openTextDocument(uri);
  await vscode.window.showTextDocument(doc, { preview: true });
}

