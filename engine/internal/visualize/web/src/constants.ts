import { EdgeKind, NodeKind } from "./types";

export interface RelationSpec {
  label: string;
  sourceKind: NodeKind;
  targetKind: NodeKind;
}

export const palette: Partial<Record<NodeKind, string>> = {
  [NodeKind.ProductRequirement]: "#e66b6b",
  [NodeKind.Requirement]: "#4f8bd6",
  [NodeKind.Feature]: "#e29c45",
  [NodeKind.Component]: "#5f9d72",
  [NodeKind.Api]: "#7b6ccf",
  [NodeKind.Service]: "#c7739f",
};

export const relationSpecs: Partial<Record<EdgeKind, RelationSpec>> = {
  [EdgeKind.Realization]: {
    label: "Realization",
    sourceKind: NodeKind.ProductRequirement,
    targetKind: NodeKind.Requirement,
  },
  [EdgeKind.FeatureImplementation]: {
    label: "Feature",
    sourceKind: NodeKind.Requirement,
    targetKind: NodeKind.Feature,
  },
  [EdgeKind.ComponentImplementation]: {
    label: "Component",
    sourceKind: NodeKind.Requirement,
    targetKind: NodeKind.Component,
  },
  [EdgeKind.ServiceImplementation]: {
    label: "Service",
    sourceKind: NodeKind.Requirement,
    targetKind: NodeKind.Service,
  },
  [EdgeKind.ApiImplementation]: {
    label: "API",
    sourceKind: NodeKind.Requirement,
    targetKind: NodeKind.Api,
  },
};

export const kindOrder: Partial<Record<NodeKind, number>> = {
  [NodeKind.ProductRequirement]: 0,
  [NodeKind.Requirement]: 1,
  [NodeKind.Feature]: 2,
  [NodeKind.Api]: 3,
  [NodeKind.Component]: 4,
  [NodeKind.Service]: 5,
};

// Tuning knobs for the grid layout (transposed layered: rows per kind).
export const gridLayoutTuning = {
  // Padding (px) passed to cy.fit() at the end of layout.
  padding: 40,
  // Horizontal distance (px) between adjacent node centres in a row.
  nodeSpacingX: 260,
  // Vertical distance (px) between sub-rows within the same kind-band.
  subRowSpacingY: 140,
  // Extra vertical gap (px) added between kind-bands so types are visually
  // separated even after wrapping into multiple sub-rows.
  kindBandGapY: 80,
  // Horizontal stagger (px) applied alternately: even-index bands shift right
  // by this amount, odd-index bands shift left. Separates overlapping edges
  // between adjacent bands so they are easier to trace visually.
  kindBandXOffset: 60,
};

// Tuning knobs for the clustered (concentric-ring) layout.
export const clusteredLayoutTuning = {
  // Padding (px) passed to cy.fit() at the end of layout.
  padding: 40,
  // Radius (px) of the innermost ring (requirements around a PR centre).
  innerRingRadius: 120,
  // Radius step (px) added per additional ring outward (use-cases, leaves).
  ringRadiusStep: 100,
  // Minimum distance (px) between the bounding circles of adjacent clusters.
  // Increase to give more breathing room between PR clusters.
  clusterGap: 60,
  // Fallback radius (px) used when Cytoscape has not yet measured a node.
  defaultNodeSize: 44,
};

export const layeredLayoutTuning = {
  // Padding (px) passed to cy.fit() at the end of layout.
  padding: 40,
  // Horizontal distance (px) between adjacent column centres.
  columnSpacingX: 300,
  // Vertical distance (px) between adjacent node centres in a column.
  // Set this large enough so no two nodes ever overlap at any zoom level.
  nodeSpacingY: 60,
  // Maximum number of adjacent-pair swap sweeps per column in the
  // transposition refinement pass that runs after the cluster-walk ordering.
  // Each sweep fixes crossing pairs left by nodes connected to multiple clusters.
  transpositionSweeps: 8,
  // Minimum clear gap (px) between adjacent node bounding-boxes in a column
  // during relaxation. Prevents nodes touching even when label text is tall.
  nodeGap: 4,
};
