import type { NodeSingular } from "cytoscape";
import { kindOrder, palette, relationSpecs } from "./constants";
import type { RelationSpec } from "./constants";
import { EdgeKind, LayoutKind, NodeKind } from "./types";
import type { GraphData, MountOptions } from "./types";

export function shapeForKind(kind: NodeKind): string {
  switch (kind) {
    case NodeKind.ProductRequirement:
      return "round-hexagon";
    case NodeKind.Requirement:
      return "round-rectangle";
    case NodeKind.Feature:
      return "ellipse";
    case NodeKind.Component:
      return "cut-rectangle";
    case NodeKind.Api:
      return "diamond";
    case NodeKind.Service:
      return "barrel";
    default:
      return "round-rectangle";
  }
}

export function lineStyleForKind(kind: EdgeKind): string {
  return kind === EdgeKind.Realization ? "solid" : "dashed";
}

export function displayKind(kind: NodeKind | string | null | undefined): string {
  return String(kind ?? "node").replace(/[_-]/g, " ");
}

export function relationSpec(kind: EdgeKind): RelationSpec {
  return (
    relationSpecs[kind] ??
    relationSpecs[EdgeKind.Realization] ?? {
      label: kind,
      sourceKind: NodeKind.Requirement,
      targetKind: NodeKind.Requirement,
    }
  );
}

export function kindRank(kind: NodeKind): number {
  return Object.prototype.hasOwnProperty.call(kindOrder, kind)
    ? (kindOrder[kind] as number)
    : Number.MAX_SAFE_INTEGER;
}

export function autoRelationKindFor(
  sourceKind: NodeKind,
  targetKind: NodeKind,
): EdgeKind | null {
  for (const [kind, spec] of Object.entries(relationSpecs)) {
    if (spec.sourceKind === sourceKind && spec.targetKind === targetKind) {
      return kind as EdgeKind;
    }
  }
  return null;
}

export function resolveRelationKindForPair(
  { relationKindSelect }: MountOptions,
  sourceKind: NodeKind,
  targetKind: NodeKind,
): EdgeKind | null {
  const selected =
    relationKindSelect && relationKindSelect.value;
  if (!selected || selected === "automatic") {
    return autoRelationKindFor(sourceKind, targetKind);
  }
  return selected as EdgeKind;
}

export function nodeDisplayLabel(
  node: NodeSingular | null | undefined,
): string {
  if (!node) {
    return "unknown node";
  }
  return (node.data("label") as string | undefined) ?? node.id();
}

export function layoutLabel(name: LayoutKind): string {
  switch (name) {
    case LayoutKind.Layered:
      return "layered";
    case LayoutKind.Organic:
      return "organic";
    case LayoutKind.Grid:
      return "grid";
    case LayoutKind.Clustered:
      return "clustered";
  }
}

export function layerIndexForKind(kind: NodeKind): number {
  if (kind === NodeKind.ProductRequirement) {
    return 0;
  }
  if (kind === NodeKind.Requirement) {
    return 1;
  }
  if (kind === NodeKind.UseCase || kind === NodeKind.UseCaseLegacy || kind === NodeKind.Feature) {
    return 2;
  }
  // Other kinds (api, component, service) go to a stable overflow column.
  return 3;
}

export function defaultRoots(graph: GraphData): string[] | undefined {
  const nodes = graph.nodes ?? [];
  for (const kind of [
    NodeKind.ProductRequirement,
    NodeKind.Requirement,
    NodeKind.Feature,
    NodeKind.Api,
    NodeKind.Component,
    NodeKind.Service,
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
    kindRank(left.data("kind") as NodeKind) -
    kindRank(right.data("kind") as NodeKind);
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

export function colorForKind(kind: NodeKind): string {
  return palette[kind] ?? "#7a8791";
}

export function activeLayoutName({ layoutSelect }: MountOptions): LayoutKind {
  return (layoutSelect?.value ?? LayoutKind.Layered) as LayoutKind;
}
