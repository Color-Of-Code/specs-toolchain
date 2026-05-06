import type { NodeSingular } from "cytoscape";
import { kindOrder, palette, relationSpecs } from "./constants";
import type { RelationSpec } from "./constants";
import type { GraphData, MountOptions } from "./types";

export function shapeForKind(kind: string): string {
  switch (kind) {
    case "product-requirement":
      return "round-hexagon";
    case "requirement":
      return "round-rectangle";
    case "feature":
      return "ellipse";
    case "component":
      return "cut-rectangle";
    case "api":
      return "diamond";
    case "service":
      return "barrel";
    default:
      return "round-rectangle";
  }
}

export function lineStyleForKind(kind: string): string {
  return kind === "realization" ? "solid" : "dashed";
}

export function displayKind(kind: string | null | undefined): string {
  return String(kind ?? "node").replace(/[_-]/g, " ");
}

export function relationSpec(kind: string): RelationSpec {
  return relationSpecs[kind] ?? relationSpecs["realization"];
}

export function kindRank(kind: string): number {
  return Object.prototype.hasOwnProperty.call(kindOrder, kind)
    ? (kindOrder[kind])
    : Number.MAX_SAFE_INTEGER;
}

export function autoRelationKindFor(
  sourceKind: string,
  targetKind: string,
): string | null {
  for (const [kind, spec] of Object.entries(relationSpecs)) {
    if (spec.sourceKind === sourceKind && spec.targetKind === targetKind) {
      return kind;
    }
  }
  return null;
}

export function resolveRelationKindForPair(
  options: MountOptions,
  sourceKind: string,
  targetKind: string,
): string | null {
  const selected =
    options.relationKindSelect && options.relationKindSelect.value;
  if (!selected || selected === "automatic") {
    return autoRelationKindFor(sourceKind, targetKind);
  }
  return selected;
}

export function nodeDisplayLabel(
  node: NodeSingular | null | undefined,
): string {
  if (!node) {
    return "unknown node";
  }
  return (node.data("label") as string | undefined) ?? node.id();
}

export function layoutLabel(name: string): string {
  switch (name) {
    case "layered":
      return "layered";
    case "organic":
      return "organic";
    case "grid":
      return "grid";
    case "clustered":
      return "clustered";
    default:
      return String(name || "layout");
  }
}

export function layerIndexForKind(kind: string): number {
  if (kind === "product-requirement") {
    return 0;
  }
  if (kind === "requirement") {
    return 1;
  }
  if (kind === "use-case" || kind === "usecase" || kind === "feature") {
    return 2;
  }
  // Other kinds (api, component, service) go to a stable overflow column.
  return 3;
}

export function defaultRoots(graph: GraphData): string[] | undefined {
  const nodes = graph.nodes ?? [];
  for (const kind of [
    "product-requirement",
    "requirement",
    "feature",
    "api",
    "component",
    "service",
  ]) {
    const roots = nodes.filter((node) => node.kind === kind).map((node) => node.id);
    if (roots.length) {
      return roots;
    }
  }
  return undefined;
}

export function compareNodeOrder(
  left: NodeSingular,
  right: NodeSingular,
): number {
  const kindDiff =
    kindRank(left.data("kind") as string) -
    kindRank(right.data("kind") as string);
  if (kindDiff !== 0) {
    return kindDiff;
  }
  const labelDiff = nodeDisplayLabel(left).localeCompare(nodeDisplayLabel(right));
  if (labelDiff !== 0) {
    return labelDiff;
  }
  return String(left.id()).localeCompare(String(right.id()));
}

export function roundCoord(value: number): number {
  return Math.round(value * 100) / 100;
}

export function createClientEdgeId(): string {
  return `e${Date.now().toString(36)}${Math.random().toString(36).slice(2, 8)}`;
}

export function shortID(nodeID: string): string {
  // Extract the PREFIX-NNN identifier from the last path segment.
  // e.g. "model/requirements/lint/LNT-001-style-rule-enforcement" -> "LNT-001"
  // Falls back to the full segment when the pattern is not found.
  const segment = String(nodeID ?? "").split("/").pop() ?? nodeID;
  const m = segment.match(/^([A-Z]+-\d+)/);
  return m ? m[1] : segment;
}

export function colorForKind(kind: string): string {
  return palette[kind] ?? "#7a8791";
}

export function activeLayoutName(options: MountOptions): string {
  return options.layoutSelect?.value ?? "layered";
}
