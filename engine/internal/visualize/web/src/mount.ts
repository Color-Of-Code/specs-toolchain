import type { Core, NodeSingular } from "cytoscape";
import { emptyGraph, persistRelations, resolveGraph } from "./graph";
import { runClusteredLayout, runGridLayout, runLayeredLayout, runLayout } from "./layout";
import { detailsIconButton, detailsMarkup, detailsRowsMarkup, escapeHTML, setDetails } from "./markup";
import { renderGraph, setMetaStatus, updateMetaSummary } from "./render";
import { EdgeKind, NodeKind } from "./types";
import type { GraphData, MountHandle, MountOptions, RelationInfo } from "./types";
import {
  activeLayoutName,
  createClientEdgeId,
  displayKind,
  layoutLabel,
  nodeDisplayLabel,
  relationSpec,
  resolveRelationKindForPair,
  shortID,
} from "./utils";

function confirmRelationChange(
  opts: MountOptions,
  action: string,
  relation: RelationInfo,
): boolean {
  if (typeof opts.onConfirm === "function") {
    return opts.onConfirm(action, relation);
  }
  if (typeof window === "undefined" || typeof window.confirm !== "function") {
    return true;
  }
  const sourceLabel = relation.sourceLabel ?? relation.source;
  const targetLabel = relation.targetLabel ?? relation.target;
  return window.confirm(
    `${action} ${relationSpec(relation.kind).label} relation?\n\n${sourceLabel}\n-> ${targetLabel}`,
  );
}

function defaultOpenPath(options: MountOptions, path: string): void {
  if (!options.artifactBaseUrl) {
    return;
  }
  const target = `${options.artifactBaseUrl}?path=${encodeURIComponent(path)}`;
  window.location.assign(target);
}

export function mount(options: MountOptions): MountHandle {
  const container = options.container;
  if (!container) {
    throw new Error("container is required");
  }

  let cy: Core | undefined;
  let selectedEdge: ReturnType<Core["$"]>[0] | undefined;
  let selectedNode: NodeSingular | undefined;
  let currentGraph: GraphData = emptyGraph();

  const openPath =
    typeof options.onOpenPath === "function"
      ? options.onOpenPath
      : (path: string) => defaultOpenPath(options, path);
  const canSaveRelations = Boolean(
    options.saveRelationsUrl ??
      typeof options.onSaveRelations === "function",
  );

  function applyFilter(filterText: string): void {
    if (!cy) {
      return;
    }
    const text = filterText.trim().toLowerCase();
    if (!text) {
      cy.nodes().removeClass("traceability-dimmed");
      return;
    }
    cy.nodes().forEach((node) => {
      const label = String(node.data("label") ?? node.id()).toLowerCase();
      const summary = String(node.data("summary") ?? "").toLowerCase();
      if (label.includes(text) || summary.includes(text)) {
        node.removeClass("traceability-dimmed");
      } else {
        node.addClass("traceability-dimmed");
      }
    });
  }

  function resolveNode(nodeID: string): NodeSingular | undefined {
    if (!cy) {
      return undefined;
    }
    const matches = cy.$id(nodeID);
    return matches.length ? (matches[0]) : undefined;
  }

  function describeNode(node: NodeSingular | undefined): string | undefined {
    if (!node) {
      return undefined;
    }
    const summary = node.data("summary") as string | undefined;
    const kind = displayKind(node.data("kind") as NodeKind);
    const inspectorID = shortID(node.id());
    const renderedRows = detailsRowsMarkup([
      { label: "Path", value: node.data("path") as string },
    ]);
    const summaryMarkup = summary
      ? `<p class="details-summary">${escapeHTML(summary)}</p>`
      : "";
    return `<article class="details-panel details-panel-node"><div class="details-header"><div class="details-header-top"><p class="details-node-meta"><span class="details-eyebrow">NODE</span><span class="details-node-kind">${escapeHTML(kind)}</span><span class="details-node-separator">:</span><span class="details-node-id">${escapeHTML(inspectorID)}</span></p>${detailsIconButton("Open artifact", node.data("path") as string)}</div><h2 class="details-title">${escapeHTML((node.data("label") as string | undefined) ?? node.id())}</h2>${summaryMarkup}</div>${renderedRows}</article>`;
  }

  function describeEdge(
    edge: ReturnType<Core["$"]>[0] | undefined,
  ): string | undefined {
    if (!edge) {
      return undefined;
    }
    const kind = edge.data("kind") as EdgeKind;
    const spec = relationSpec(kind);
    const sourceNode = resolveNode(edge.data("source") as string);
    const targetNode = resolveNode(edge.data("target") as string);
    return detailsMarkup(
      "Relation",
      spec.label,
      [
        { label: "Kind", value: displayKind(kind) },
        {
          label: "Source",
          value: sourceNode
            ? ((sourceNode.data("label") as string | undefined) ?? sourceNode.id())
            : (edge.data("source") as string),
          link: sourceNode ? (sourceNode.data("path") as string | undefined) : undefined,
        },
        {
          label: "Target",
          value: targetNode
            ? ((targetNode.data("label") as string | undefined) ?? targetNode.id())
            : (edge.data("target") as string),
          link: targetNode ? (targetNode.data("path") as string | undefined) : undefined,
        },
      ],
      canSaveRelations
        ? "Use the remove button to delete this relation. The UI asks for confirmation before persisting the change."
        : "",
    );
  }

  function updateDetailsPanel(): void {
    if (!options.detailsElement) {
      return;
    }
    if (selectedEdge) {
      setDetails(options, describeEdge(selectedEdge) ?? "");
      return;
    }
    if (selectedNode) {
      setDetails(options, describeNode(selectedNode) ?? "");
      return;
    }
    setDetails(
      options,
      detailsMarkup(
        "Inspector",
        "No selection",
        [],
        "Select a node or edge to inspect its details.",
      ),
    );
  }

  function currentSelectedEdge(): ReturnType<Core["$"]>[0] | undefined {
    if (selectedEdge && !selectedEdge.removed()) {
      return selectedEdge;
    }
    if (!cy) {
      return undefined;
    }
    const selectedEdges = cy.$("edge:selected");
    return selectedEdges.length > 0
      ? (selectedEdges[0])
      : undefined;
  }

  function updateRemoveEdgeButton(): void {
    if (!options.removeEdgeButton) {
      return;
    }
    selectedEdge = currentSelectedEdge();
    options.removeEdgeButton.disabled = !canSaveRelations || !selectedEdge;
  }

  function relationAlreadyExists(
    source: string,
    target: string,
    kind: EdgeKind,
  ): boolean {
    if (!cy) {
      return false;
    }
    return cy
      .edges()
      .some(
        (edge) =>
          edge.data("source") === source &&
          edge.data("target") === target &&
          edge.data("kind") === kind,
      );
  }

  async function completeDragConnect(
    sourceNode: NodeSingular,
    targetNode: NodeSingular,
  ): Promise<void> {
    if (!cy) {
      return;
    }
    const sourceKind = sourceNode.data("kind") as NodeKind;
    const targetKind = targetNode.data("kind") as NodeKind;
    const kind = resolveRelationKindForPair(options, sourceKind, targetKind);
    if (!kind) {
      setMetaStatus(
        options,
        `no relation defined between ${displayKind(sourceKind)} and ${displayKind(targetKind)}`,
      );
      return;
    }
    const edge: RelationInfo = {
      source: sourceNode.id(),
      target: targetNode.id(),
      kind,
    };
    if (relationAlreadyExists(edge.source, edge.target, edge.kind)) {
      setMetaStatus(options, "edge already exists");
      return;
    }
    try {
      await persistRelations(options, cy, { appendEdges: [edge] });
      const addedEdge = cy.add({
        data: { id: createClientEdgeId(), ...edge },
      });
      selectedNode = undefined;
      selectedEdge = addedEdge[0];
      updateMetaSummary(options, cy.nodes().length, cy.edges().length);
      updateRemoveEdgeButton();
      setMetaStatus(options, "edge added");
      updateDetailsPanel();
    } catch (error) {
      console.error(error);
      setMetaStatus(options, "edge creation failed");
    }
  }

  function clearDragTargetClasses(): void {
    if (!cy) {
      return;
    }
    cy.nodes().removeClass(
      "traceability-create-inactive traceability-create-source traceability-create-target",
    );
  }

  // ── Wire toolbar controls ──────────────────────────────────────────────

  if (options.relationKindSelect) {
    options.relationKindSelect.disabled = !canSaveRelations;
  }

  if (options.fitButton) {
    options.fitButton.addEventListener("click", () => {
      if (cy) {
        cy.fit(undefined, 40);
      }
    });
  }

  if (options.zoomInButton) {
    options.zoomInButton.addEventListener("click", () => {
      if (cy) {
        cy.zoom({
          level: cy.zoom() * 1.25,
          renderedPosition: { x: cy.width() / 2, y: cy.height() / 2 },
        });
      }
    });
  }

  if (options.zoomOutButton) {
    options.zoomOutButton.addEventListener("click", () => {
      if (cy) {
        cy.zoom({
          level: cy.zoom() * 0.8,
          renderedPosition: { x: cy.width() / 2, y: cy.height() / 2 },
        });
      }
    });
  }

  if (options.layoutSelect) {
    options.layoutSelect.addEventListener("change", () => {
      if (!cy) {
        return;
      }
      const layoutName = activeLayoutName(options);
      runLayout(cy, currentGraph, options, layoutName);
      setMetaStatus(options, `${layoutLabel(layoutName)} layout applied`);
    });
  }

  if (options.filterInput) {
    const filterInput = options.filterInput;
    filterInput.addEventListener("input", () => {
      applyFilter(filterInput.value);
    });
  }

  if (options.removeEdgeButton) {
    updateRemoveEdgeButton();
    const removeEdgeButton = options.removeEdgeButton;
    removeEdgeButton.addEventListener("click", () => {
      void (async () => {
        const edgeToRemove = currentSelectedEdge();
        if (!edgeToRemove || !canSaveRelations || !cy) {
          return;
        }
        const removalPreview: RelationInfo = {
          source: edgeToRemove.data("source") as string,
          target: edgeToRemove.data("target") as string,
          kind: edgeToRemove.data("kind") as EdgeKind,
          sourceLabel: nodeDisplayLabel(
            resolveNode(edgeToRemove.data("source") as string),
          ),
          targetLabel: nodeDisplayLabel(
            resolveNode(edgeToRemove.data("target") as string),
          ),
        };
        if (!confirmRelationChange(options, "Remove", removalPreview)) {
          setMetaStatus(options, "edge removal cancelled");
          updateDetailsPanel();
          return;
        }
        removeEdgeButton.disabled = true;
        try {
          await persistRelations(options, cy, {
            omitEdgeId: edgeToRemove.id(),
          });
          edgeToRemove.remove();
          selectedEdge = undefined;
          selectedNode = undefined;
          updateMetaSummary(options, cy.nodes().length, cy.edges().length);
          setMetaStatus(options, "edge removed");
          updateDetailsPanel();
        } catch (error) {
          console.error(error);
          setMetaStatus(options, "edge removal failed");
        } finally {
          window.setTimeout(() => {
            updateRemoveEdgeButton();
          }, 1200);
        }
      })();
    });
  }

  if (options.detailsElement) {
    options.detailsElement.addEventListener("click", (event) => {
      const target = event.target;
      if (!(target instanceof HTMLElement)) {
        return;
      }
      const action = target.closest("[data-open-path]");
      if (!(action instanceof HTMLElement)) {
        return;
      }
      const path = action.dataset["openPath"];
      if (!path) {
        return;
      }
      openPath(path);
    });
  }

  // ── Initialise graph ───────────────────────────────────────────────────

  resolveGraph(options)
    .then((graph) => {
      currentGraph = graph;
      cy = renderGraph(options, graph) ?? undefined;
      if (cy) {
        const layoutName = activeLayoutName(options);
        if (layoutName === "layered") {
          runLayeredLayout(cy);
        } else if (layoutName === "grid") {
          runGridLayout(cy);
        } else if (layoutName === "clustered") {
          runClusteredLayout(cy);
        }
        updateMetaSummary(options, cy.nodes().length, cy.edges().length);
        updateDetailsPanel();
        if (options.filterInput) {
          applyFilter(options.filterInput.value);
        }

        // ── Drag-to-connect handle ───────────────────────────────────────
        let dragSourceNode: NodeSingular | null = null;
        let isDraggingEdge = false;
        let hideHandleTimer: ReturnType<typeof window.setTimeout> | null = null;

        const handle =
          typeof document !== "undefined"
            ? document.createElement("div")
            : null;
        const svgNS = "http://www.w3.org/2000/svg";
        const dragSvg =
          typeof document !== "undefined"
            ? document.createElementNS(svgNS, "svg")
            : null;
        const dragLine = dragSvg
          ? document.createElementNS(svgNS, "line")
          : null;

        if (handle && dragSvg && dragLine && canSaveRelations && cy) {
          const cyRef = cy;
          const h = handle;
          const svg = dragSvg;
          const line = dragLine;
          h.setAttribute("class", "drag-connect-handle");
          h.style.display = "none";
          document.body.appendChild(h);

          svg.setAttribute("class", "drag-connect-overlay");
          svg.style.cssText =
            "position:fixed;top:0;left:0;width:100vw;height:100vh;pointer-events:none;z-index:9000;display:none;";
          line.setAttribute("class", "drag-connect-line");
          svg.appendChild(line);
          document.body.appendChild(svg);

          function renderedToViewport(renderedPos: {
            x: number;
            y: number;
          }): { x: number; y: number } {
            const rect = container.getBoundingClientRect();
            return {
              x: rect.left + renderedPos.x,
              y: rect.top + renderedPos.y,
            };
          }

          function findNodeAtViewport(
            clientX: number,
            clientY: number,
          ): NodeSingular | undefined {
            const rect = container.getBoundingClientRect();
            const rx = clientX - rect.left;
            const ry = clientY - rect.top;
            let found: NodeSingular | undefined;
            cyRef.nodes().forEach((node) => {
              const bb = node.renderedBoundingBox();
              if (
                rx >= bb.x1 &&
                rx <= bb.x2 &&
                ry >= bb.y1 &&
                ry <= bb.y2
              ) {
                found = node;
              }
            });
            return found;
          }

          function highlightDragTargets(srcNode: NodeSingular): void {
            cyRef.nodes().forEach((node) => {
              if (node.same(srcNode)) {
                node.addClass("traceability-create-source");
                return;
              }
              const kind = resolveRelationKindForPair(
                options,
                srcNode.data("kind") as NodeKind,
                node.data("kind") as NodeKind,
              );
              if (kind) {
                node.addClass("traceability-create-target");
              } else {
                node.addClass("traceability-create-inactive");
              }
            });
          }

          function showHandle(node: NodeSingular): void {
            const bb = node.renderedBoundingBox();
            const rect = container.getBoundingClientRect();
            const cx = rect.left + (bb.x1 + bb.x2) / 2;
            const cy2 = rect.top + bb.y1 - 4;
            const size = Math.round(
              Math.max(8, Math.min(20, 20 * cyRef.zoom())),
            );
            const half = size / 2;
            h.style.width = `${size}px`;
            h.style.height = `${size}px`;
            h.style.left = `${cx - half}px`;
            h.style.top = `${cy2 - half}px`;
            h.style.display = "flex";
          }

          function hideHandle(): void {
            if (!isDraggingEdge) {
              h.style.display = "none";
            }
          }

          function scheduleHideHandle(): void {
            hideHandleTimer = window.setTimeout(hideHandle, 120);
          }

          function cancelHideHandle(): void {
            if (hideHandleTimer) {
              window.clearTimeout(hideHandleTimer);
              hideHandleTimer = null;
            }
          }

          h.addEventListener("mouseenter", cancelHideHandle);
          h.addEventListener("mouseleave", scheduleHideHandle);

          h.addEventListener("mousedown", (event) => {
            const handleSize = Math.round(
              Math.max(8, Math.min(20, 20 * cyRef.zoom())),
            );
            const nodeUnder = findNodeAtViewport(
              event.clientX,
              event.clientY + handleSize * 0.7,
            );
            if (!nodeUnder) {
              return;
            }
            event.stopPropagation();
            event.preventDefault();
            isDraggingEdge = true;
            dragSourceNode = nodeUnder;
            highlightDragTargets(dragSourceNode);
            const srcVP = renderedToViewport(dragSourceNode.renderedPosition());
            line.setAttribute("x1", String(srcVP.x));
            line.setAttribute("y1", String(srcVP.y));
            line.setAttribute("x2", String(event.clientX));
            line.setAttribute("y2", String(event.clientY));
            svg.style.display = "block";

            function onMouseMove(e: MouseEvent): void {
              if (!dragSourceNode) {
                return;
              }
              const srcVP2 = renderedToViewport(
                dragSourceNode.renderedPosition(),
              );
              line.setAttribute("x1", String(srcVP2.x));
              line.setAttribute("y1", String(srcVP2.y));
              line.setAttribute("x2", String(e.clientX));
              line.setAttribute("y2", String(e.clientY));
            }

            function onMouseUp(e: MouseEvent): void {
              document.removeEventListener("mousemove", onMouseMove);
              document.removeEventListener("mouseup", onMouseUp);
              isDraggingEdge = false;
              svg.style.display = "none";
              hideHandle();
              clearDragTargetClasses();
              const targetNode = findNodeAtViewport(e.clientX, e.clientY);
              if (
                targetNode &&
                dragSourceNode &&
                !targetNode.same(dragSourceNode)
              ) {
                void completeDragConnect(dragSourceNode, targetNode);
              } else {
                setMetaStatus(options, "drag cancelled");
              }
              dragSourceNode = null;
            }

            document.addEventListener("mousemove", onMouseMove);
            document.addEventListener("mouseup", onMouseUp);
          });

          cyRef.on("mouseover", "node", (event) => {
            if (!isDraggingEdge) {
              cancelHideHandle();
              showHandle(event.target as NodeSingular);
            }
          });

          cyRef.on("mouseout", "node", () => {
            scheduleHideHandle();
          });
        }
        // ── end drag-to-connect ──────────────────────────────────────────

        cy.on("tap", (event) => {
          if (event.target === cy) {
            selectedEdge = undefined;
            selectedNode = undefined;
            updateRemoveEdgeButton();
            updateDetailsPanel();
          }
        });
        cy.on("tap", "edge", (event) => {
          selectedEdge = event.target as ReturnType<Core["$"]>[0];
          selectedNode = undefined;
          updateRemoveEdgeButton();
          updateDetailsPanel();
        });
        cy.on("tap", "node", (event) => {
          selectedEdge = undefined;
          selectedNode = event.target as NodeSingular;
          updateRemoveEdgeButton();
          updateDetailsPanel();
        });
      }
    })
    .catch((error: unknown) => {
      container.innerHTML = `<pre style="padding: 16px; color: inherit;">${String(error)}</pre>`;
    });

  // ── Public handle ──────────────────────────────────────────────────────

  function updateGraph(graph: GraphData): void {
    if (!cy) {
      return;
    }
    const cyRef = cy;
    const savedZoom = cyRef.zoom();
    const savedPan = cyRef.pan();

    const newNodeIds = new Set((graph.nodes ?? []).map((n) => n.id));
    cyRef.nodes().forEach((n) => {
      if (!newNodeIds.has(n.id())) {
        n.remove();
      }
    });
    (graph.nodes ?? []).forEach((node) => {
      if (!cyRef.$id(node.id).length) {
        cyRef.add({
          data: {
            id: node.id,
            label: node.label,
            path: node.path,
            kind: node.kind,
            summary: node.summary ?? "",
          },
        });
      }
    });

    const edgeKey = (source: string, target: string, kind: string): string =>
      `${source}\0${target}\0${kind}`;
    const newEdgeKeys = new Set(
      (graph.edges ?? []).map((e) => edgeKey(e.source, e.target, e.kind)),
    );
    cyRef.edges().forEach((e) => {
      if (
        !newEdgeKeys.has(
          edgeKey(
            e.data("source") as string,
            e.data("target") as string,
            e.data("kind") as EdgeKind,
          ),
        )
      ) {
        e.remove();
      }
    });
    const existingEdgeKeys = new Set<string>();
    cyRef.edges().forEach((e) => {
      existingEdgeKeys.add(
        edgeKey(
          e.data("source") as string,
          e.data("target") as string,
          e.data("kind") as EdgeKind,
        ),
      );
    });
    (graph.edges ?? []).forEach((edge) => {
      if (!existingEdgeKeys.has(edgeKey(edge.source, edge.target, edge.kind))) {
        cyRef.add({
          data: {
            id: createClientEdgeId(),
            source: edge.source,
            target: edge.target,
            kind: edge.kind,
          },
        });
      }
    });

    currentGraph = graph;
    cyRef.viewport({ zoom: savedZoom, pan: savedPan });

    selectedEdge = undefined;
    selectedNode = undefined;
    updateMetaSummary(options, cyRef.nodes().length, cyRef.edges().length);
    updateRemoveEdgeButton();
    updateDetailsPanel();
    if (options.filterInput) {
      applyFilter(options.filterInput.value);
    }
  }

  return {
    fit() {
      if (cy) {
        cy.fit(undefined, 40);
      }
    },
    update(graph: GraphData) {
      updateGraph(graph);
    },
  };
}
