import type { Core } from "cytoscape";
import type {
  EdgeData,
  EdgeKind,
  GraphData,
  MountOptions,
  NodeData,
  RelationInfo,
  RelationsPayload,
} from "./types";

export function emptyGraph(): GraphData {
  return { nodes: [], edges: [] };
}

export async function resolveGraph({ graph, graphUrl }: MountOptions): Promise<GraphData> {
  if (graph) {
    return graph;
  }
  if (!graphUrl) {
    return emptyGraph();
  }
  const response = await fetch(graphUrl, { cache: "no-store" });
  if (!response.ok) {
    throw new Error(`graph request failed: ${response.status}`);
  }
  return response.json() as Promise<GraphData>;
}

export function buildElements(graph: GraphData): object[] {
  return [
    ...(graph.nodes ?? []).map((node: NodeData) => ({
      data: {
        id: node.id,
        label: node.label,
        path: node.path,
        kind: node.kind,
        summary: node.summary ?? "",
      },
    })),
    ...(graph.edges ?? []).map((edge: EdgeData, index: number) => ({
      data: {
        id: `e${index}`,
        source: edge.source,
        target: edge.target,
        kind: edge.kind,
      },
    })),
  ];
}

export function collectRelations(
  cy: Core,
  omitEdgeId: string | undefined,
  appendEdges: RelationInfo[] | undefined,
): RelationInfo[] {
  return cy
    .edges()
    .toArray()
    .filter((edge) => edge.id() !== omitEdgeId)
    .map((edge) => ({
      source: edge.data("source") as string,
      target: edge.data("target") as string,
      kind: edge.data("kind") as EdgeKind,
    }))
    .concat(appendEdges ?? [])
    .sort((left, right) => {
      if (left.kind !== right.kind) {
        return left.kind.localeCompare(right.kind);
      }
      if (left.source !== right.source) {
        return left.source.localeCompare(right.source);
      }
      return left.target.localeCompare(right.target);
    });
}

export interface PersistRelationsOptions {
  omitEdgeId?: string;
  appendEdges?: RelationInfo[];
}

export async function persistRelations(
  { onSaveRelations, saveRelationsUrl }: MountOptions,
  cy: Core,
  relationOptions?: PersistRelationsOptions,
): Promise<void> {
  const payload: RelationsPayload = {
    edges: collectRelations(
      cy,
      relationOptions?.omitEdgeId,
      relationOptions?.appendEdges,
    ),
  };
  if (typeof onSaveRelations === "function") {
    await onSaveRelations(payload);
    return;
  }
  if (!saveRelationsUrl) {
    return;
  }
  const response = await fetch(saveRelationsUrl, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
  if (!response.ok) {
    const message = await response.text();
    throw new Error(message || `relations request failed: ${response.status}`);
  }
}
