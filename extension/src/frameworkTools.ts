// Framework-defined LM tools — discovers skills from framework/skills/*.md and
// registers each one as a vscode.lm tool so it is available in Agent Mode
// and any other agentic context, not just @specs.
import * as vscode from "vscode";
import { runAndCapture, findSpecsFolder, findSpecsRoot } from "./engine";

interface SkillEngineArgs {
  [key: string]: string[];
}

interface SkillInfo {
  id: string;
  name: string;
  description: string;
  tags: string[];
  inputSchema?: object;
  engineArgs?: SkillEngineArgs;
  file: string;
}

interface LintInput {
  check?: "all" | "links" | "style" | "baselines";
}

// Tracks disposables for registered tools so we can re-register on framework change.
const registeredTools: vscode.Disposable[] = [];

export function registerFrameworkTools(context: vscode.ExtensionContext): void {
  // Initial registration.
  void refreshFrameworkTools(context);

  // Re-register whenever .specs.yaml changes (framework_dir may have changed).
  const watcher = vscode.workspace.createFileSystemWatcher("**/.specs.yaml");
  const reregister = () => void refreshFrameworkTools(context);
  watcher.onDidChange(reregister, undefined, context.subscriptions);
  watcher.onDidCreate(reregister, undefined, context.subscriptions);
  watcher.onDidDelete(reregister, undefined, context.subscriptions);
  context.subscriptions.push(watcher);
}

async function refreshFrameworkTools(context: vscode.ExtensionContext): Promise<void> {
  // Dispose previously registered tools before re-registering.
  for (const d of registeredTools.splice(0)) {
    d.dispose();
  }

  const folder = findSpecsFolder();
  if (!folder) {
    return;
  }
  const cwd = findSpecsRoot(folder) ?? folder.uri.fsPath;

  const result = await runAndCapture(context, ["framework", "skills", "list"], cwd);
  if (result.exitCode !== 0 || !result.stdout.trim()) {
    return;
  }

  let skills: SkillInfo[];
  try {
    skills = JSON.parse(result.stdout) as SkillInfo[];
  } catch {
    return;
  }

  for (const skill of skills) {
    const disposable = registerSkill(context, skill, cwd);
    if (disposable) {
      registeredTools.push(disposable);
      context.subscriptions.push(disposable);
    }
  }
}

function registerSkill(
  context: vscode.ExtensionContext,
  skill: SkillInfo,
  cwd: string,
): vscode.Disposable | undefined {
  try {
    return vscode.lm.registerTool<LintInput>(skill.id, {
      async invoke(
        options: vscode.LanguageModelToolInvocationOptions<LintInput>,
        token: vscode.CancellationToken,
      ): Promise<vscode.LanguageModelToolResult> {
        const engineArgs = resolveEngineArgs(skill, options.input);
        if (!engineArgs) {
          return new vscode.LanguageModelToolResult([
            new vscode.LanguageModelTextPart(
              `Skill "${skill.id}" has no engine args configured.`,
            ),
          ]);
        }

        const result = await runAndCapture(context, engineArgs, cwd);
        if (token.isCancellationRequested) {
          return new vscode.LanguageModelToolResult([
            new vscode.LanguageModelTextPart("Cancelled."),
          ]);
        }

        const output = [result.stdout, result.stderr].filter(Boolean).join("\n").trim();
        const text = output || (result.exitCode === 0 ? "Done — no output." : `Exited with code ${result.exitCode}.`);
        return new vscode.LanguageModelToolResult([
          new vscode.LanguageModelTextPart(text),
        ]);
      },
    });
  } catch {
    // registerTool throws if the id is already taken; skip silently.
    return undefined;
  }
}

function resolveEngineArgs(skill: SkillInfo, input: LintInput): string[] | undefined {
  const engineArgs = skill.engineArgs;
  if (!engineArgs) {
    return undefined;
  }
  // For skills with a "check" input parameter, look up by value; fall back to "default".
  if (input && typeof input === "object" && "check" in input && input.check) {
    const args = engineArgs[input.check] ?? engineArgs["default"];
    if (args) {
      return args;
    }
  }
  return engineArgs["default"] ?? Object.values(engineArgs)[0];
}
