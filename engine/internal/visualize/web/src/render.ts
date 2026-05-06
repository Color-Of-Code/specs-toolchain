import cytoscape from "cytoscape";
import type { Core, CytoscapeOptions } from "cytoscape";
import { EdgeKind, NodeKind } from "./types";
import type { GraphData, MountOptions } from "./types";
import { layoutOptions } from "./layout";
import { activeLayoutName, colorForKind, lineStyleForKind, shapeForKind } from "./utils";
import { buildElements } from "./graph";

export function updateMetaSummary(
  { metaElement }: MountOptions,
  nodeCount: number,
  edgeCount: number,
): void {
  if (!metaElement) {
    return;
  }
  const summary = `${nodeCount} nodes / ${edgeCount} edges`;
  metaElement.dataset["summary"] = summary;
  metaElement.textContent = summary;
}

export function setMetaStatus({ metaElement }: MountOptions, message: string): void {
  if (!metaElement) {
    return;
  }
  const summary =
    metaElement.dataset["summary"] ??
    metaElement.textContent ??
    "";
  metaElement.textContent = summary
    ? `${summary} • ${message}`
    : message;
}

export function renderGraph(
  options: MountOptions,
  graph: GraphData,
): Core | undefined {
  const { metaElement, container, emptyMessage } = options;
  const { nodes, edges } = graph;
  if (metaElement) {
    const summary = `${nodes.length} nodes / ${edges.length} edges`;
    metaElement.dataset["summary"] = summary;
    metaElement.textContent = summary;
  }
  if (!nodes.length) {
    container.innerHTML = `<pre style="padding: 16px; color: inherit;">${emptyMessage ?? "No traceability data found."}</pre>`;
    return undefined;
  }
  return cytoscape({
    container,
    elements: buildElements(graph) as CytoscapeOptions["elements"],
    layout: layoutOptions(graph, activeLayoutName(options)) as unknown as CytoscapeOptions["layout"],
    style: [
      {
        selector: "node",
        style: {
          label: "data(label)",
          shape: (ele) => shapeForKind(ele.data("kind") as NodeKind) as import("cytoscape").Css.NodeShape,
          width: "label",
          height: "label",
          padding: "14px",
          "text-wrap": "wrap",
          "text-max-width": "160px",
          "font-size": 12,
          "font-weight": 600,
          color: "#10222e",
          "text-valign": "center",
          "text-halign": "center",
          "border-width": 2,
          "border-color": "#173042",
          "background-color": (ele) => colorForKind(ele.data("kind") as NodeKind),
        },
      },
      {
        selector: "node:selected",
        style: {
          "border-width": 4,
          "border-color": "#f5f1c7",
          "overlay-opacity": 0,
        },
      },
      {
        selector: "edge",
        style: {
          width: 2.2,
          "curve-style": "bezier",
          "line-style": (ele) => lineStyleForKind(ele.data("kind") as EdgeKind) as import("cytoscape").Css.LineStyle,
          "line-color": "#6d7f88",
          "target-arrow-color": "#6d7f88",
          "target-arrow-shape": "triangle",
          "arrow-scale": 1.15,
        },
      },
      {
        selector: "edge:selected",
        style: {
          width: 4,
          "line-color": "#f5f1c7",
          "target-arrow-color": "#f5f1c7",
        },
      },
      {
        selector: ".traceability-create-inactive",
        style: {
          opacity: 0.28,
        },
      },
      {
        selector: ".traceability-create-source",
        style: {
          "border-width": 4,
          "border-color": "#f8d66d",
        },
      },
      {
        selector: ".traceability-create-target",
        style: {
          "border-width": 4,
          "border-color": "#91f0b5",
        },
      },
      {
        selector: ".traceability-dimmed",
        style: {
          opacity: 0.15,
        },
      },
    ],
  });
}
