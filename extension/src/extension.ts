import * as vscode from "vscode";
import { getOutput } from "./engine";
import { registerCommands } from "./commands";
import { registerChatParticipant } from "./chatParticipant";
import { registerCRTree } from "./crTree";
import { registerModelTree } from "./modelTree";
import { registerStatusBar } from "./statusBar";
import { registerVisualizePanel } from "./visualizePanel";

export function activate(context: vscode.ExtensionContext): void {
  const out = getOutput();
  out.appendLine(`Specs extension activated (v${context.extension.packageJSON.version})`);
  registerCommands(context);
  registerChatParticipant(context);
  registerCRTree(context);
  registerModelTree(context, "specs.productRequirements", "product", "specs.productRequirements.refresh");
  registerModelTree(context, "specs.requirements", "model/requirements", "specs.requirements.refresh");
  registerModelTree(context, "specs.features", "model/features", "specs.features.refresh");
  registerStatusBar(context);
  registerVisualizePanel(context);
}

export function deactivate(): void {
  // nothing
}
