import * as vscode from "vscode";
import { getOutput } from "./cli";
import { registerCommands } from "./commands";

export function activate(context: vscode.ExtensionContext): void {
    const out = getOutput();
    out.appendLine(`Specs extension activated (v${context.extension.packageJSON.version})`);
    registerCommands(context);
}

export function deactivate(): void {
    // nothing
}
