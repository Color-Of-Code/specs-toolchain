import type { Core, NodeSingular, EdgeSingular } from "cytoscape";

export type { Core, NodeSingular, EdgeSingular };

export interface NodeData {
  id: string;
  label: string;
  path: string;
  kind: string;
  summary?: string;
}

export interface EdgeData {
  source: string;
  target: string;
  kind: string;
}

export interface GraphData {
  nodes: NodeData[];
  edges: EdgeData[];
}

export interface RelationInfo {
  source: string;
  target: string;
  kind: string;
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
