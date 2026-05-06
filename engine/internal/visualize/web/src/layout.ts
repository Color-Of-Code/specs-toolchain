import type { Core, NodeSingular } from "cytoscape";
import {
  clusteredLayoutTuning,
  gridLayoutTuning,
  layeredLayoutTuning,
} from "./constants";
import { NodeKind } from "./types";
import type { GraphData, MountOptions } from "./types";
import {
  compareNodeOrder,
  defaultRoots,
  layerIndexForKind,
  roundCoord,
} from "./utils";

// ── layoutOptions ──────────────────────────────────────────────────────────
// Returns the initial Cytoscape layout config for the given layout name.
// For "layered" and "grid" a breadthfirst pass is used as a starting point;
// the custom run*Layout functions then reposition every node.

export function layoutOptions(
  graph: GraphData,
  name: string,
): Record<string, unknown> {
  switch (name) {
    case "organic":
      return {
        name: "cose",
        fit: true,
        padding: 40,
        animate: false,
        nodeRepulsion: 160000,
        idealEdgeLength: 150,
        edgeElasticity: 90,
        gravity: 30,
        nestingFactor: 0.8,
      };
    case "grid":
      return {
        name: "grid",
        fit: true,
        padding: gridLayoutTuning.padding,
        avoidOverlap: true,
        sort: compareNodeOrder,
      };
    case "clustered":
      return {
        name: "breadthfirst",
        directed: true,
        roots: defaultRoots(graph),
        fit: true,
        padding: 40,
        spacingFactor: 1.2,
        avoidOverlap: true,
        depthSort: compareNodeOrder,
      };
    case "layered":
    default:
      return {
        name: "breadthfirst",
        directed: true,
        roots: defaultRoots(graph),
        fit: true,
        padding: layeredLayoutTuning.padding,
        spacingFactor: 1.2,
        avoidOverlap: true,
        depthSort: compareNodeOrder,
      };
  }
}

// ── runLayout ──────────────────────────────────────────────────────────────

export function runLayout(
  cy: Core,
  graph: GraphData,
  options: MountOptions,
  layoutName: string,
): void {
  if (layoutName === "clustered") {
    runClusteredLayout(cy);
  } else if (layoutName === "layered") {
    runLayeredLayout(cy);
  } else if (layoutName === "grid") {
    runGridLayout(cy);
  } else {
    cy.layout(layoutOptions(graph, layoutName) as unknown as Parameters<Core["layout"]>[0]).run();
  }
}

// ── runClusteredLayout ─────────────────────────────────────────────────────
// Clustered concentric-ring layout.
// Builds one sub-cluster per product-requirement node.
// Each cluster: ring 0 = the PR itself, ring 1 = connected requirements,
// ring 2 = connected use-cases, ring 3 = remaining connected leaves.
// Nodes unreachable from any PR are collected in a single extra cluster.

export function runClusteredLayout(cy: Core): void {
  const allNodes = cy.nodes().toArray();
  const allEdges = cy.edges().toArray();

  const adj = new Map<string, Set<string>>(
    allNodes.map((n) => [n.id(), new Set<string>()]),
  );
  allEdges.forEach((e) => {
    adj.get(e.data("source") as string)?.add(e.data("target") as string);
    adj.get(e.data("target") as string)?.add(e.data("source") as string);
  });

  const prNodes = allNodes.filter(
    (n) => n.data("kind") === NodeKind.ProductRequirement,
  );
  const globalPlaced = new Set<string>();

  function buildCluster(
    pr: NodeSingular,
  ): [NodeSingular[], NodeSingular[], NodeSingular[]] {
    const ring1 = allNodes.filter(
      (n) =>
        !globalPlaced.has(n.id()) &&
        n.data("kind") === NodeKind.Requirement &&
        (adj.get(pr.id())?.has(n.id()) ?? false),
    );
    ring1.forEach((n) => globalPlaced.add(n.id()));

    const ring1ids = new Set(ring1.map((n) => n.id()));
    const ring2 = allNodes.filter(
      (n) =>
        !globalPlaced.has(n.id()) &&
        (n.data("kind") === NodeKind.UseCase || n.data("kind") === NodeKind.UseCaseLegacy) &&
        ring1.some((r) => adj.get(r.id())?.has(n.id()) ?? false),
    );
    ring2.forEach((n) => globalPlaced.add(n.id()));

    const ring12ids = new Set([...ring1ids, ...ring2.map((n) => n.id())]);
    const ring3 = allNodes.filter(
      (n) =>
        !globalPlaced.has(n.id()) &&
        [...(adj.get(n.id()) ?? [])].some((id) => ring12ids.has(id)),
    );
    ring3.forEach((n) => globalPlaced.add(n.id()));

    return [ring1, ring2, ring3];
  }

  function clusterRadius(rings: NodeSingular[][]): number {
    const nonEmpty = rings.filter((r) => r.length > 0);
    return (
      clusteredLayoutTuning.innerRingRadius +
      nonEmpty.length * clusteredLayoutTuning.ringRadiusStep
    );
  }

  function placeRing(
    ring: NodeSingular[],
    cx: number,
    cy2: number,
    r: number,
  ): void {
    ring.forEach((node, i) => {
      const angle = (2 * Math.PI * i) / ring.length - Math.PI / 2;
      node.position({
        x: roundCoord(cx + r * Math.cos(angle)),
        y: roundCoord(cy2 + r * Math.sin(angle)),
      });
    });
  }

  cy.startBatch();
  globalPlaced.clear();

  type Cluster = { pr: NodeSingular | null; rings: NodeSingular[][]; radius: number };
  const clusters: Cluster[] = prNodes.map((pr) => {
    globalPlaced.add(pr.id());
    const rings = buildCluster(pr);
    return { pr: pr, rings, radius: clusterRadius(rings) };
  });

  const orphans = allNodes.filter((n) => !globalPlaced.has(n.id()));
  if (orphans.length > 0) {
    clusters.push({
      pr: null,
      rings: [orphans, [], []],
      radius: clusterRadius([orphans, [], []]),
    });
  }

  const cols = Math.ceil(Math.sqrt(clusters.length));
  clusters.forEach((cluster, idx) => {
    const col = idx % cols;
    const row = Math.floor(idx / cols);
    const maxR = clusters
      .slice(row * cols, row * cols + cols)
      .reduce((m, c) => Math.max(m, c.radius), 0);
    const cellSize = 2 * maxR + clusteredLayoutTuning.clusterGap;
    const cx = col * cellSize + cellSize / 2;
    const cy2 = row * cellSize + cellSize / 2;

    if (cluster.pr) {
      cluster.pr.position({ x: roundCoord(cx), y: roundCoord(cy2) });
    }

    cluster.rings.forEach((ring, ringIdx) => {
      if (ring.length === 0) {
        return;
      }
      const r =
        clusteredLayoutTuning.innerRingRadius +
        ringIdx * clusteredLayoutTuning.ringRadiusStep;
      placeRing(ring, cx, cy2, r);
    });
  });

  cy.endBatch();
  cy.fit(undefined, clusteredLayoutTuning.padding);
}

// ── runLayeredLayout ───────────────────────────────────────────────────────
// Left-to-right column layout grouping nodes by kind-layer.
// Uses cluster-walk ordering and transposition refinement to minimise crossings,
// then a master-column bottom-up pass to align connected nodes vertically.

export function runLayeredLayout(cy: Core): void {
  const nodes = cy.nodes().toArray();
  if (nodes.length === 0) {
    return;
  }

  const layers = new Map<number, NodeSingular[]>();
  nodes.forEach((node) => {
    const layer = layerIndexForKind(node.data("kind") as NodeKind);
    let arr = layers.get(layer);
    if (!arr) {
      arr = [];
      layers.set(layer, arr);
    }
    arr.push(node);
  });

  const layerKeys = Array.from(layers.keys()).sort((a, b) => a - b);
  layerKeys.forEach((layer) => {
    layers.get(layer)?.sort(compareNodeOrder);
  });

  const edges = cy.edges().toArray().map((edge) => ({
    source: edge.data("source") as string,
    target: edge.data("target") as string,
  }));

  function indexByNodeId(layer: number): Map<string, number> {
    const order = new Map<string, number>();
    (layers.get(layer) ?? []).forEach((node, index) => {
      order.set(node.id(), index);
    });
    return order;
  }

  function neighborPositions(
    nodeId: string,
    neighborOrder: Map<string, number>,
    useIncomingEdges: boolean,
  ): number[] {
    const positions: number[] = [];
    for (const edge of edges) {
      if (useIncomingEdges) {
        if (edge.target === nodeId && neighborOrder.has(edge.source)) {
          positions.push(neighborOrder.get(edge.source) ?? 0);
        }
      } else if (edge.source === nodeId && neighborOrder.has(edge.target)) {
        positions.push(neighborOrder.get(edge.target) ?? 0);
      }
    }
    return positions;
  }

  function countCrossings(
    aPositions: number[],
    bPositions: number[],
  ): number {
    let count = 0;
    for (const p of aPositions) {
      for (const q of bPositions) {
        if (p > q) {
          count += 1;
        }
      }
    }
    return count;
  }

  function transposeLayer(layer: number): void {
    const orderedNodes = layers.get(layer);
    if (!orderedNodes || orderedNodes.length < 2) {
      return;
    }
    const layerPos = layerKeys.indexOf(layer);
    const leftOrder =
      layerPos > 0 ? indexByNodeId(layerKeys[layerPos - 1]) : new Map<string, number>();
    const rightOrder =
      layerPos < layerKeys.length - 1
        ? indexByNodeId(layerKeys[layerPos + 1])
        : new Map<string, number>();
    let improved = true;
    let sweeps = 0;
    while (improved && sweeps < layeredLayoutTuning.transpositionSweeps) {
      improved = false;
      sweeps += 1;
      for (let i = 0; i < orderedNodes.length - 1; i += 1) {
        const u = orderedNodes[i];
        const v = orderedNodes[i + 1];
        const uid = u.id();
        const vid = v.id();
        const uLeft = neighborPositions(uid, leftOrder, true);
        const vLeft = neighborPositions(vid, leftOrder, true);
        const uRight = neighborPositions(uid, rightOrder, false);
        const vRight = neighborPositions(vid, rightOrder, false);
        const current =
          countCrossings(uLeft, vLeft) + countCrossings(uRight, vRight);
        const swapped =
          countCrossings(vLeft, uLeft) + countCrossings(vRight, uRight);
        if (swapped < current) {
          orderedNodes[i] = v;
          orderedNodes[i + 1] = u;
          improved = true;
        }
      }
    }
  }

  // Cluster-walk ordering.
  for (let i = 0; i < layerKeys.length - 1; i += 1) {
    const leftLayer = layerKeys[i];
    const rightLayer = layerKeys[i + 1];
    const rightSet = new Map<string, NodeSingular>(
      (layers.get(rightLayer) ?? []).map((node) => [node.id(), node]),
    );
    const placed = new Set<string>();
    const ordered: NodeSingular[] = [];

    (layers.get(leftLayer) ?? []).forEach((leftNode) => {
      const lid = leftNode.id();
      for (const edge of edges) {
        let rightId: string | null = null;
        if (edge.source === lid && rightSet.has(edge.target)) {
          rightId = edge.target;
        } else if (edge.target === lid && rightSet.has(edge.source)) {
          rightId = edge.source;
        }
        if (rightId !== null) {
          const rightNode = rightSet.get(rightId);
          if (rightNode && !placed.has(rightId)) {
            ordered.push(rightNode);
            placed.add(rightId);
          }
        }
      }
    });

    (layers.get(rightLayer) ?? []).forEach((node) => {
      if (!placed.has(node.id())) {
        ordered.push(node);
      }
    });

    layers.set(rightLayer, ordered);
  }

  layerKeys.forEach((layer) => transposeLayer(layer));

  // Find master column (most nodes).
  let masterLayer = layerKeys[0];
  layerKeys.forEach((layer) => {
    if (
      (layers.get(layer) ?? []).length > (layers.get(masterLayer) ?? []).length
    ) {
      masterLayer = layer;
    }
  });
  const masterNodes = layers.get(masterLayer) ?? [];

  const adjacent = new Map<string, Set<string>>();
  nodes.forEach((n) => adjacent.set(n.id(), new Set<string>()));
  for (const edge of edges) {
    adjacent.get(edge.source)?.add(edge.target);
    adjacent.get(edge.target)?.add(edge.source);
  }

  const colTopY = new Map<number, number>(
    layerKeys.map((layer) => [layer, Infinity]),
  );
  const placed = new Set<string>(masterNodes.map((n) => n.id()));

  cy.startBatch();

  layerKeys.forEach((layer, layerPos) => {
    const x = layerPos * layeredLayoutTuning.columnSpacingX;
    (layers.get(layer) ?? []).forEach((node, nodeIndex) => {
      node.position({
        x: roundCoord(x),
        y:
          layer === masterLayer
            ? roundCoord(nodeIndex * layeredLayoutTuning.nodeSpacingY)
            : 0,
      });
    });
  });

  for (let mi = masterNodes.length - 1; mi >= 0; mi -= 1) {
    const mNode = masterNodes[mi];
    const mY = mi * layeredLayoutTuning.nodeSpacingY;

    layerKeys.forEach((layer) => {
      if (layer === masterLayer) {
        return;
      }
      const colNodes = layers.get(layer) ?? [];
      const group = colNodes.filter(
        (n) => !placed.has(n.id()) && (adjacent.get(mNode.id())?.has(n.id()) ?? false),
      );
      if (group.length === 0) {
        return;
      }

      const prevTop = colTopY.get(layer) ?? Infinity;
      const actualBottom =
        prevTop === Infinity
          ? mY
          : Math.min(mY, prevTop - layeredLayoutTuning.nodeGap);
      const count = group.length;
      const groupTop =
        actualBottom - (count - 1) * layeredLayoutTuning.nodeSpacingY;

      group.forEach((n, i) => {
        n.position({
          x: n.position().x,
          y: roundCoord(
            actualBottom - (count - 1 - i) * layeredLayoutTuning.nodeSpacingY,
          ),
        });
        placed.add(n.id());
      });

      colTopY.set(
        layer,
        prevTop === Infinity ? groupTop : Math.min(prevTop, groupTop),
      );
    });
  }

  layerKeys.forEach((layer) => {
    if (layer === masterLayer) {
      return;
    }
    const unplaced = (layers.get(layer) ?? []).filter(
      (n) => !placed.has(n.id()),
    );
    if (unplaced.length === 0) {
      return;
    }
    const top = colTopY.get(layer) ?? Infinity;
    let cursor = top === Infinity ? 0 : top - layeredLayoutTuning.nodeGap;
    for (let i = unplaced.length - 1; i >= 0; i -= 1) {
      unplaced[i].position({
        x: unplaced[i].position().x,
        y: roundCoord(cursor),
      });
      placed.add(unplaced[i].id());
      cursor -= layeredLayoutTuning.nodeSpacingY;
    }
  });

  cy.endBatch();
  cy.fit(undefined, layeredLayoutTuning.padding);
}

// ── runGridLayout ──────────────────────────────────────────────────────────
// Grid layout: the layered layout transposed.
// Rows contain nodes of a single kind. Within each row nodes are ordered by
// the same cluster-walk used in the layered layout so that columns of connected
// nodes align vertically, minimising crossing angles.

export function runGridLayout(cy: Core): void {
  const nodes = cy.nodes().toArray();
  if (nodes.length === 0) {
    return;
  }

  const rows = new Map<number, NodeSingular[]>();
  nodes.forEach((node) => {
    const row = layerIndexForKind(node.data("kind") as NodeKind);
    let arr = rows.get(row);
    if (!arr) {
      arr = [];
      rows.set(row, arr);
    }
    arr.push(node);
  });

  const rowKeys = Array.from(rows.keys()).sort((a, b) => a - b);
  rowKeys.forEach((row) => rows.get(row)?.sort(compareNodeOrder));

  const edges = cy.edges().toArray().map((e) => ({
    source: e.data("source") as string,
    target: e.data("target") as string,
  }));

  // Cluster-walk: walk each row left-to-right; for every node pull its
  // unvisited neighbours in the next row into position immediately following
  // nodes already placed by previous pulls.
  for (let i = 0; i < rowKeys.length - 1; i += 1) {
    const topRow = rowKeys[i];
    const btmRow = rowKeys[i + 1];
    const btmSet = new Map<string, NodeSingular>(
      (rows.get(btmRow) ?? []).map((node) => [node.id(), node]),
    );
    const placed = new Set<string>();
    const ordered: NodeSingular[] = [];

    (rows.get(topRow) ?? []).forEach((topNode) => {
      for (const edge of edges) {
        let btmId: string | null = null;
        if (edge.source === topNode.id() && btmSet.has(edge.target)) {
          btmId = edge.target;
        } else if (edge.target === topNode.id() && btmSet.has(edge.source)) {
          btmId = edge.source;
        }
        if (btmId !== null) {
          const btmNode = btmSet.get(btmId);
          if (btmNode && !placed.has(btmId)) {
            ordered.push(btmNode);
            placed.add(btmId);
          }
        }
      }
    });

    (rows.get(btmRow) ?? []).forEach((node) => {
      if (!placed.has(node.id())) {
        ordered.push(node);
      }
    });

    rows.set(btmRow, ordered);
  }

  const colCount = Math.max(1, Math.ceil(Math.sqrt(nodes.length)));

  const adjIds = new Map<string, Set<string>>(
    nodes.map((n) => [n.id(), new Set<string>()]),
  );
  for (const edge of edges) {
    adjIds.get(edge.source)?.add(edge.target);
    adjIds.get(edge.target)?.add(edge.source);
  }
  const nodeById = new Map<string, NodeSingular>(
    nodes.map((n) => [n.id(), n]),
  );

  cy.startBatch();
  let currentY = 0;
  rowKeys.forEach((row, bandIdx) => {
    const kindNodes = rows.get(row) ?? [];
    const rowsPerBand = Math.max(1, Math.ceil(kindNodes.length / colCount));
    const xOffset =
      bandIdx % 2 === 0
        ? gridLayoutTuning.kindBandXOffset
        : -gridLayoutTuning.kindBandXOffset;

    // Step 1: column-major initial placement with alternating x-offset.
    kindNodes.forEach((node, idx) => {
      const col = Math.floor(idx / rowsPerBand);
      const subRow = idx % rowsPerBand;
      node.position({
        x: roundCoord(col * gridLayoutTuning.nodeSpacingX + xOffset),
        y: roundCoord(currentY + subRow * gridLayoutTuning.subRowSpacingY),
      });
    });

    // Step 2: greedy right-shift pass — improve edge verticality.
    const usedCells = new Set<string>(
      kindNodes.map((n) => {
        const c = Math.round(
          (n.position().x - xOffset) / gridLayoutTuning.nodeSpacingX,
        );
        const r = Math.round(
          (n.position().y - currentY) / gridLayoutTuning.subRowSpacingY,
        );
        return `${c},${r}`;
      }),
    );

    for (let idx = kindNodes.length - 1; idx >= 0; idx -= 1) {
      const node = kindNodes[idx];
      const pos = node.position();
      const curCol = Math.round(
        (pos.x - xOffset) / gridLayoutTuning.nodeSpacingX,
      );
      const curSubRow = Math.round(
        (pos.y - currentY) / gridLayoutTuning.subRowSpacingY,
      );

      const nbrs = [...(adjIds.get(node.id()) ?? [])]
        .map((id) => nodeById.get(id))
        .filter((n): n is NodeSingular => n !== undefined);
      if (nbrs.length === 0) {
        continue;
      }

      const curDx =
        nbrs.reduce((s, n2) => s + Math.abs(n2.position().x - pos.x), 0) /
        nbrs.length;
      let bestCol = curCol;
      let bestDx = curDx;

      for (let c = curCol + 1; c < colCount; c += 1) {
        if (usedCells.has(`${c},${curSubRow}`)) {
          break;
        }
        const testX = c * gridLayoutTuning.nodeSpacingX + xOffset;
        const dx =
          nbrs.reduce((s, n2) => s + Math.abs(n2.position().x - testX), 0) /
          nbrs.length;
        if (dx < bestDx) {
          bestDx = dx;
          bestCol = c;
        }
      }

      if (bestCol !== curCol) {
        usedCells.delete(`${curCol},${curSubRow}`);
        usedCells.add(`${bestCol},${curSubRow}`);
        node.position({
          x: roundCoord(bestCol * gridLayoutTuning.nodeSpacingX + xOffset),
          y: pos.y,
        });
      }
    }

    currentY +=
      rowsPerBand * gridLayoutTuning.subRowSpacingY + gridLayoutTuning.kindBandGapY;
  });
  cy.endBatch();
  cy.fit(undefined, gridLayoutTuning.padding);
}
