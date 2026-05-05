// Resolves and invokes the bundled (or user-configured) specs engine binary.
import * as path from "path";
import * as fs from "fs";
import * as cp from "child_process";
import * as vscode from "vscode";

let outputChannel: vscode.OutputChannel | undefined;

export function getOutput(): vscode.OutputChannel {
  if (!outputChannel) {
    outputChannel = vscode.window.createOutputChannel("Specs");
  }
  return outputChannel;
}

/**
 * Resolution order:
 *   1. specs.enginePath setting (absolute path or relative to workspace)
 *   2. specs.useGlobalBinary === true: 'specs' on PATH
 *   3. workspace-local binary at <workspace>/bin/specs[.exe]
 *   4. extension's bundled binary at <extensionPath>/bin/specs[.exe]
 *   5. fallback: 'specs' on PATH
 */
export function resolveBinary(context: vscode.ExtensionContext): string {
  const cfg = vscode.workspace.getConfiguration("specs");
  const explicit = cfg.get<string>("enginePath", "").trim();
  if (explicit) {
    return resolveExplicitBinary(explicit);
  }
  const useGlobal = cfg.get<boolean>("useGlobalBinary", false);
  const exe = process.platform === "win32" ? "specs.exe" : "specs";
  if (!useGlobal) {
    const local = findWorkspaceBinary(exe);
    if (local) {
      return local;
    }
  }
  const bundled = path.join(context.extensionPath, "bin", exe);
  if (!useGlobal && fs.existsSync(bundled)) {
    return bundled;
  }
  return "specs"; // resolved via PATH
}

function getWorkspaceFolderForResolution(): vscode.WorkspaceFolder | undefined {
  return findSpecsFolder();
}

function resolveExplicitBinary(explicit: string): string {
  if (path.isAbsolute(explicit)) {
    return explicit;
  }
  const folder = getWorkspaceFolderForResolution();
  if (!folder) {
    return explicit;
  }
  return path.resolve(folder.uri.fsPath, explicit);
}

function findWorkspaceBinary(exe: string): string | undefined {
  const folder = getWorkspaceFolderForResolution();
  if (!folder) {
    return undefined;
  }
  const candidate = path.join(folder.uri.fsPath, "bin", exe);
  if (fs.existsSync(candidate)) {
    return candidate;
  }
  return undefined;
}

export interface RunResult {
  stdout: string;
  stderr: string;
  exitCode: number;
}

export interface SpecsExecutionTarget {
  folder: vscode.WorkspaceFolder;
  cwd: string;
}

/** Runs the engine and captures stdout/stderr. Logs to the Specs output channel. */
export async function runAndCapture(
  context: vscode.ExtensionContext,
  args: string[],
  cwd?: string,
): Promise<RunResult> {
  const bin = resolveBinary(context);
  const out = getOutput();
  out.appendLine(`$ ${bin} ${args.join(" ")}  (cwd=${cwd ?? "<none>"})`);
  return new Promise((resolve) => {
    const proc = cp.spawn(bin, args, { cwd, env: process.env });
    let stdout = "";
    let stderr = "";
    proc.stdout.on("data", (d) => (stdout += d.toString()));
    proc.stderr.on("data", (d) => (stderr += d.toString()));
    proc.on("error", (err) => {
      resolve({ stdout, stderr: stderr + String(err), exitCode: 127 });
    });
    proc.on("close", (code) => {
      out.appendLine(`  -> exit ${code}`);
      resolve({ stdout, stderr, exitCode: code ?? 0 });
    });
  });
}

/** Runs the engine in a dedicated VS Code terminal so the user can interact with it. */
export function runInTerminal(
  context: vscode.ExtensionContext,
  args: string[],
  cwd?: string,
  name = "Specs",
): vscode.Terminal {
  const bin = resolveBinary(context);
  const term = vscode.window.createTerminal({ name, cwd });
  // Quote args containing spaces; engine args here are simple flags/values.
  const quoted = args.map((a) => (/[\s"'$]/.test(a) ? JSON.stringify(a) : a));
  term.sendText(`${bin} ${quoted.join(" ")}`);
  term.show();
  return term;
}

/** Returns the workspace folder containing a .specs.yaml (or specs/.specs.yaml), or undefined. */
export function findSpecsFolder(): vscode.WorkspaceFolder | undefined {
  const folders = vscode.workspace.workspaceFolders ?? [];
  for (const f of folders) {
    const root = f.uri.fsPath;
    if (
      fs.existsSync(path.join(root, ".specs.yaml")) ||
      fs.existsSync(path.join(root, "specs", ".specs.yaml"))
    ) {
      return f;
    }
  }
  return folders[0];
}

export function getSpecsExecutionTarget(options?: {
  warnIfMissing?: boolean;
}): SpecsExecutionTarget | undefined {
  const folder = findSpecsFolder();
  if (!folder) {
    if (options?.warnIfMissing) {
      vscode.window.showWarningMessage("Specs: no workspace folder is open.");
    }
    return undefined;
  }

  return {
    folder,
    cwd: findSpecsRoot(folder) ?? folder.uri.fsPath,
  };
}

/** Returns the resolved specs root from .specs.yaml when present. */
export function findSpecsRoot(folder: vscode.WorkspaceFolder): string | undefined {
  const cfgPath = findSpecsConfigPath(folder.uri.fsPath);
  if (!cfgPath) {
    return undefined;
  }
  const configDir = path.dirname(cfgPath);
  const configuredRoot = readConfigPathValue(cfgPath, "specs_root");
  if (!configuredRoot) {
    return configDir;
  }
  return path.resolve(configDir, configuredRoot);
}

function findSpecsConfigPath(workspaceRoot: string): string | undefined {
  const candidates = [workspaceRoot, path.join(workspaceRoot, "specs")];
  for (const candidate of candidates) {
    const cfgPath = path.join(candidate, ".specs.yaml");
    if (fs.existsSync(cfgPath)) {
      return cfgPath;
    }
  }
  return undefined;
}

function readConfigPathValue(cfgPath: string, key: string): string | undefined {
  let text: string;
  try {
    text = fs.readFileSync(cfgPath, "utf8");
  } catch {
    return undefined;
  }
  const pattern = new RegExp(`^${key}:\\s*(.+?)\\s*$`, "m");
  const match = text.match(pattern);
  if (!match) {
    return undefined;
  }
  const raw = match[1].trim();
  if (!raw || raw === "null") {
    return undefined;
  }
  if (
    (raw.startsWith('"') && raw.endsWith('"')) ||
    (raw.startsWith("'") && raw.endsWith("'"))
  ) {
    return raw.slice(1, -1);
  }
  return raw;
}
