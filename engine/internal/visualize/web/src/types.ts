import type { Core, NodeSingular, EdgeSingular } from "cytoscape";

export type { Core, NodeSingular, EdgeSingular };

export enum NodeKind {
  ProductRequirement = "product-requirement",
  Requirement = "requirement",
  Feature = "feature",
  Component = "component",
  Api = "api",
  Service = "service",
  UseCase = "use-case",
  UseCaseLegacy = "usecase",
}

export enum EdgeKind {
  Realization = "realization",
  FeatureImplementation = "feature_implementation",
  ComponentImplementation = "component_implementation",
  ServiceImplementation = "service_implementation",
  ApiImplementation = "api_implementation",
}

export enum LayoutKind {
  Layered = "layered",
  Organic = "organic",
  Grid = "grid",
  Clustered = "clustered",
}

export interface NodeData {
  id: string;
  label: string;
  path: string;
  kind: NodeKind;
  summary?: string;
}

export interface EdgeData {
  source: string;
  target: string;
  kind: EdgeKind;
}

export interface GraphData {
  nodes: NodeData[];
  edges: EdgeData[];
}

export interface RelationInfo {
  source: string;
  target: string;
  kind: EdgeKind;
  sourceLabel?: string;
  targetLabel?: string;
}

export interface RelationsPayload {
  edges: RelationInfo[];
}

export interface MountOptions {
  container: HTMLElement;
  graphUrl?: string;
  graph?: GraphData;
  saveRelationsUrl?: string;
  artifactBaseUrl?: string;
  emptyMessage?: string;
  metaElement?: HTMLElement;
  detailsElement?: HTMLElement;
  fitButton?: HTMLButtonElement;
  zoomInButton?: HTMLButtonElement;
  zoomOutButton?: HTMLButtonElement;
  layoutSelect?: HTMLSelectElement;
  filterInput?: HTMLInputElement;
  removeEdgeButton?: HTMLButtonElement;
  relationKindSelect?: HTMLSelectElement;
  onOpenPath?: (path: string) => void;
  onSaveRelations?: (payload: RelationsPayload) => Promise<void>;
  onConfirm?: (action: string, relation: RelationInfo) => boolean;
}

export interface MountHandle {
  fit(): void;
  update(graph: GraphData): void;
}

export interface DetailRow {
  label: string;
  value: string | null | undefined;
  link?: string;
}
