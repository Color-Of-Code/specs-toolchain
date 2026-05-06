// @specs Chat Participant — central AI entry point for the specs toolchain.
import * as vscode from "vscode";
import { runAndCapture, getSpecsExecutionTarget } from "./engine";

const PARTICIPANT_ID = "specs-toolchain.specs";

// Map slash commands to engine subcommand args.
const COMMAND_ARGS: Record<string, string[]> = {
  lint: ["lint"],
  doctor: ["doctor"],
  trace: ["graph", "validate"],
};

interface AgentCommand {
  name: string;
  description: string;
}

interface AgentInfo {
  id: string;
  name: string;
  description: string;
  commands?: AgentCommand[];
  systemPrompt: string;
  file: string;
}

// In-process cache of agents loaded from the framework.
let cachedAgents: AgentInfo[] = [];

export function registerChatParticipant(context: vscode.ExtensionContext): void {
  const participant = vscode.chat.createChatParticipant(PARTICIPANT_ID, makeHandler(context));
  participant.iconPath = new vscode.ThemeIcon("book");
  participant.followupProvider = { provideFollowups };
  context.subscriptions.push(participant);

  // Load agents from the framework initially and on config changes.
  void loadAgents(context);
  const watcher = vscode.workspace.createFileSystemWatcher("**/.specs.yaml");
  const reload = (): void => { void loadAgents(context); };
  watcher.onDidChange(reload, undefined, context.subscriptions);
  watcher.onDidCreate(reload, undefined, context.subscriptions);
  watcher.onDidDelete(reload, undefined, context.subscriptions);
  context.subscriptions.push(watcher);
}

async function loadAgents(context: vscode.ExtensionContext): Promise<void> {
  const target = getSpecsExecutionTarget();
  if (!target) {
    cachedAgents = [];
    return;
  }
  const result = await runAndCapture(context, ["framework", "agents", "list"], target.cwd);
  if (result.exitCode !== 0 || !result.stdout.trim()) {
    cachedAgents = [];
    return;
  }
  try {
    cachedAgents = JSON.parse(result.stdout) as AgentInfo[];
  } catch {
    cachedAgents = [];
  }
}

/** Returns the agent whose command list includes the given slash-command name, if any. */
function findAgentForCommand(command: string): AgentInfo | undefined {
  return cachedAgents.find((a) => a.commands?.some((c) => c.name === command));
}

function makeHandler(
  context: vscode.ExtensionContext,
): vscode.ChatRequestHandler {
  return async (
    request: vscode.ChatRequest,
    _chatContext: vscode.ChatContext,
    stream: vscode.ChatResponseStream,
    token: vscode.CancellationToken,
  ): Promise<vscode.ChatResult> => {
    const target = getSpecsExecutionTarget();
    if (!target) {
      stream.markdown(
        "No specs workspace found. Open a folder containing `.specs.yaml` first.",
      );
      return {};
    }
    const {cwd} = target;

    // Slash-command dispatch.
    if (request.command) {
      return handleCommand(request, context, stream, token, cwd);
    }

    // Free-form: forward to the language model with workspace context.
    return handleFreeForm(request, stream, token, context, cwd);
  };
}

async function handleCommand(
  request: vscode.ChatRequest,
  context: vscode.ExtensionContext,
  stream: vscode.ChatResponseStream,
  token: vscode.CancellationToken,
  cwd: string,
): Promise<vscode.ChatResult> {
  const cmd = request.command as string;

  // Scaffold: delegate to LLM with scaffolding instructions.
  if (cmd === "scaffold") {
    return handleScaffold(request, stream, token);
  }

  // CR: delegate to LLM with CR context.
  if (cmd === "cr") {
    return handleCR(request, stream, token);
  }

  const engineArgs = COMMAND_ARGS[cmd];
  if (!engineArgs) {
    // Check whether a framework agent owns this command.
    const agent = findAgentForCommand(cmd);
    if (agent) {
      return handleAgentCommand(request, stream, token, agent, context, cwd);
    }
    stream.markdown(`Unknown command \`/${cmd}\`.`);
    return {};
  }

  stream.progress(`Running \`specs ${engineArgs.join(" ")}\`…`);
  const result = await runAndCapture(context, engineArgs, cwd);
  if (token.isCancellationRequested) {
    return {};
  }

  const combined = [result.stdout, result.stderr].filter(Boolean).join("\n").trim();
  if (combined) {
    stream.markdown("```\n" + combined + "\n```");
  } else {
    stream.markdown("Done — no output.");
  }
  if (result.exitCode !== 0) {
    stream.markdown(`\n_Engine exited with code ${result.exitCode}._`);
  }
  return { metadata: { command: cmd } };
}

async function handleScaffold(
  request: vscode.ChatRequest,
  stream: vscode.ChatResponseStream,
  token: vscode.CancellationToken,
): Promise<vscode.ChatResult> {
  const models = await vscode.lm.selectChatModels({ family: "gpt-4o" });
  const model = models[0] ?? request.model;
  const prompt = request.prompt.trim() || "requirement";
  const messages = [
    vscode.LanguageModelChatMessage.User(
      `You are a specs toolchain assistant. The user wants to scaffold a new model artifact.\n` +
        `Available kinds: requirement, use-case, component.\n` +
        `Based on their description, suggest the most appropriate kind and a kebab-case slug path ` +
        `(e.g. "core/some-slug" for requirements, "some-slug" for others). ` +
        `Then show the exact \`Specs: Scaffold\` command to run in VS Code.\n\n` +
        `User request: ${prompt}`,
    ),
  ];

  if (token.isCancellationRequested) {
    return {};
  }

  const response = await model.sendRequest(messages, {}, token);
  for await (const chunk of response.text) {
    stream.markdown(chunk);
  }
  stream.button({ command: "specs.scaffold.requirement", title: "Scaffold: Requirement" });
  stream.button({ command: "specs.scaffold.use-case", title: "Scaffold: Use Case" });
  return { metadata: { command: "scaffold" } };
}

async function handleCR(
  request: vscode.ChatRequest,
  stream: vscode.ChatResponseStream,
  token: vscode.CancellationToken,
): Promise<vscode.ChatResult> {
  const models = await vscode.lm.selectChatModels({ family: "gpt-4o" });
  const model = models[0] ?? request.model;
  const prompt = request.prompt.trim() || "list open change requests";
  const messages = [
    vscode.LanguageModelChatMessage.User(
      `You are a specs toolchain assistant helping manage change requests.\n` +
        `Change requests live under the CR directory (e.g. specs/change-requests/).\n` +
        `Each CR is a numbered directory, e.g. CR-001-some-description.\n` +
        `The engine commands are:\n` +
        `  specs cr new --title "..." [--id NNN]\n` +
        `  specs cr status\n` +
        `  specs cr drain --id NNN [--yes] [--dry-run]\n\n` +
        `Based on the user's request, explain what to do and show the appropriate command.\n\n` +
        `User request: ${prompt}`,
    ),
  ];

  if (token.isCancellationRequested) {
    return {};
  }

  const response = await model.sendRequest(messages, {}, token);
  for await (const chunk of response.text) {
    stream.markdown(chunk);
  }
  stream.button({ command: "specs.cr.new", title: "New Change Request" });
  stream.button({ command: "specs.cr.status", title: "CR Status" });
  return { metadata: { command: "cr" } };
}

async function handleAgentCommand(
  request: vscode.ChatRequest,
  stream: vscode.ChatResponseStream,
  token: vscode.CancellationToken,
  agent: AgentInfo,
  context: vscode.ExtensionContext,
  cwd: string,
): Promise<vscode.ChatResult> {
  // Collect graph context to enrich the agent's prompt.
  const graphResult = await runAndCapture(context, ["graph", "validate"], cwd);
  const graphSummary = graphResult.stdout.split("\n").slice(0, 10).join("\n");

  const messages = [
    vscode.LanguageModelChatMessage.User(
      agent.systemPrompt +
        `\n\nWorkspace graph summary:\n${graphSummary || "(not available)"}`,
    ),
    vscode.LanguageModelChatMessage.User(
      request.prompt || `Please perform the /${request.command} analysis.`,
    ),
  ];

  if (token.isCancellationRequested) {
    return {};
  }

  const response = await request.model.sendRequest(messages, {}, token);
  for await (const chunk of response.text) {
    stream.markdown(chunk);
  }
  return { metadata: { command: request.command, agentId: agent.id } };
}

async function handleFreeForm(
  request: vscode.ChatRequest,
  stream: vscode.ChatResponseStream,
  token: vscode.CancellationToken,
  context: vscode.ExtensionContext,
  cwd: string,
): Promise<vscode.ChatResult> {
  // Gather lightweight workspace context from the engine.
  const graphResult = await runAndCapture(context, ["graph", "validate"], cwd);
  const graphSummary = graphResult.stdout.split("\n").slice(0, 10).join("\n");

  // Pick the best matching agent for the user's prompt, or fall back to the
  // default specs assistant persona.
  const agentSystemPrompt = selectAgentForPrompt(request.prompt);
  const baseSystemPrompt =
    `You are the Specs assistant, an expert in requirements engineering and ` +
    `the specs-toolchain. You help authors write, link, and validate technical ` +
    `specifications stored as Markdown files.\n\n` +
    `The workspace uses the specs-toolchain engine (binary: specs). Key concepts:\n` +
    `- Product requirements live under specs/product/\n` +
    `- Technical requirements under specs/model/requirements/\n` +
    `- Use cases under specs/model/use-cases/\n` +
    `- Traceability edges in YAML files under specs/model/traceability/\n` +
    `- Change requests under specs/change-requests/\n\n` +
    `Current graph summary:\n${graphSummary || "(not available)"}`;

  const systemPrompt = agentSystemPrompt
    ? agentSystemPrompt + `\n\nWorkspace graph summary:\n${graphSummary || "(not available)"}`
    : baseSystemPrompt;

  const messages = [
    vscode.LanguageModelChatMessage.User(systemPrompt),
    vscode.LanguageModelChatMessage.User(request.prompt),
  ];

  if (token.isCancellationRequested) {
    return {};
  }

  const response = await request.model.sendRequest(messages, {}, token);
  for await (const chunk of response.text) {
    stream.markdown(chunk);
  }
  return {};
}

/**
 * Naively selects the first agent whose disambiguation examples contain a
 * keyword match against the user's prompt. Falls back to undefined (default
 * specs persona) when no agent matches.
 */
function selectAgentForPrompt(prompt: string): string | undefined {
  const lower = prompt.toLowerCase();
  for (const agent of cachedAgents) {
    for (const dis of agent.commands ?? []) {
      // Check agent command names as keywords.
      if (lower.includes(dis.name.toLowerCase())) {
        return agent.systemPrompt;
      }
    }
  }
  return undefined;
}

function provideFollowups(
  result: vscode.ChatResult,
  _context: vscode.ChatContext,
  _token: vscode.CancellationToken,
): vscode.ChatFollowup[] | undefined {
  const cmdRaw: unknown = result.metadata?.["command"];
  const cmd = typeof cmdRaw === "string" ? cmdRaw : undefined;
  if (cmd === "lint") {
    return [
      { prompt: "@specs /trace", label: "Check traceability" },
      { prompt: "@specs /doctor", label: "Run diagnostics" },
    ];
  }
  if (cmd === "scaffold") {
    return [{ prompt: "@specs /lint", label: "Lint after scaffolding" }];
  }
  if (cmd === "cr") {
    return [{ prompt: "@specs /trace", label: "Verify traceability" }];
  }
  return [
    { prompt: "@specs /lint", label: "Lint the model" },
    { prompt: "@specs /trace", label: "Validate traceability" },
  ];
}
