(function () {
  const palette = {
    "product-requirement": "#e66b6b",
    requirement: "#4f8bd6",
    feature: "#e29c45",
    component: "#5f9d72",
    api: "#7b6ccf",
    service: "#c7739f",
  };

  const relationSpecs = {
    realization: {
      label: "Realization",
      sourceKind: "product-requirement",
      targetKind: "requirement",
    },
    feature_implementation: {
      label: "Feature",
      sourceKind: "requirement",
      targetKind: "feature",
    },
    component_implementation: {
      label: "Component",
      sourceKind: "requirement",
      targetKind: "component",
    },
    service_implementation: {
      label: "Service",
      sourceKind: "requirement",
      targetKind: "service",
    },
    api_implementation: {
      label: "API",
      sourceKind: "requirement",
      targetKind: "api",
    },
  };

  function autoRelationKindFor(sourceKind, targetKind) {
    for (const [kind, spec] of Object.entries(relationSpecs)) {
      if (spec.sourceKind === sourceKind && spec.targetKind === targetKind) {
        return kind;
      }
    }
    return null;
  }

  function resolveRelationKindForPair(options, sourceKind, targetKind) {
    const selected = options.relationKindSelect && options.relationKindSelect.value;
    if (!selected || selected === "automatic") {
      return autoRelationKindFor(sourceKind, targetKind);
    }
    return selected;
  }

  const kindOrder = {
    "product-requirement": 0,
    requirement: 1,
    feature: 2,
    api: 3,
    component: 4,
    service: 5,
  };

  // Tuning knobs for the layered (left-to-right columns) layout.
  const layeredLayoutTuning = {
    // Padding (px) passed to cy.fit() at the end of layout.
    padding: 40,
    // Horizontal distance (px) between adjacent column centres.
    columnSpacingX: 300,
    // Vertical distance (px) between adjacent node centres in a column.
    // Set this large enough so no two nodes ever overlap at any zoom level.
    nodeSpacingY: 110,
    // Number of cose iterations for the virtual organic pre-pass whose
    // y-positions are used to determine the vertical order of nodes.
    organicPrepassIterations: 800,
    // Maximum number of adjacent-pair swap sweeps per column in the
    // transposition refinement pass that runs after the organic pre-pass.
    transpositionSweeps: 8,
  };

  function shapeForKind(kind) {
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

  function lineStyleForKind(kind) {
    return kind === "realization" ? "solid" : "dashed";
  }

  function displayKind(kind) {
    return String(kind || "node").replace(/[_-]/g, " ");
  }

  function relationSpec(kind) {
    return relationSpecs[kind] || relationSpecs.realization;
  }

  function kindRank(kind) {
    return Object.prototype.hasOwnProperty.call(kindOrder, kind) ? kindOrder[kind] : Number.MAX_SAFE_INTEGER;
  }

  function escapeHTML(value) {
    return String(value == null ? "" : value).replace(/[&<>"']/g, (char) => {
      switch (char) {
        case "&":
          return "&amp;";
        case "<":
          return "&lt;";
        case ">":
          return "&gt;";
        case '"':
          return "&quot;";
        case "'":
          return "&#39;";
        default:
          return char;
      }
    });
  }

  function detailsMarkup(eyebrow, title, rows, note) {
    const renderedRows = detailsRowsMarkup(rows);
    const noteMarkup = note ? `<p class="details-note">${escapeHTML(note)}</p>` : "";
    return `<article class="details-panel"><p class="details-eyebrow">${escapeHTML(eyebrow)}</p><h2 class="details-title">${escapeHTML(title)}</h2>${renderedRows}${noteMarkup}</article>`;
  }

  function detailsRowsMarkup(rows) {
    const renderedRows = (rows || [])
      .filter((row) => row && row.value != null && row.value !== "")
      .map((row) => `<div><dt>${escapeHTML(row.label)}</dt><dd>${escapeHTML(row.value)}</dd></div>`)
      .join("");
    return renderedRows ? `<dl class="details-list">${renderedRows}</dl>` : "";
  }

  function detailsIconButton(label, path) {
    if (!path) {
      return "";
    }
    return `<button type="button" class="details-open-button details-icon-button" data-open-path="${escapeHTML(path)}" aria-label="${escapeHTML(label)}" title="${escapeHTML(label)}"><span class="details-visually-hidden">${escapeHTML(label)}</span></button>`;
  }

  function setDetails(options, markup) {
    if (!options.detailsElement) {
      return;
    }
    options.detailsElement.innerHTML = markup;
  }

  function relationDisplayLabel(kind) {
    return relationSpec(kind).label;
  }

  function nodeDisplayLabel(node) {
    if (!node) {
      return "unknown node";
    }
    return node.data("label") || node.id();
  }

  function confirmRelationChange(opts, action, relation) {
    if (typeof opts.onConfirm === "function") {
      return opts.onConfirm(action, relation);
    }
    if (typeof window === "undefined" || typeof window.confirm !== "function") {
      return true;
    }
    const sourceLabel = relation.sourceLabel || relation.source;
    const targetLabel = relation.targetLabel || relation.target;
    return window.confirm(`${action} ${relationDisplayLabel(relation.kind)} relation?\n\n${sourceLabel}\n-> ${targetLabel}`);
  }

  function emptyGraph() {
    return { nodes: [], edges: [] };
  }

  async function resolveGraph(options) {
    if (options.graph) {
      return options.graph;
    }
    if (!options.graphUrl) {
      return emptyGraph();
    }
    const response = await fetch(options.graphUrl, { cache: "no-store" });
    if (!response.ok) {
      throw new Error(`graph request failed: ${response.status}`);
    }
    return response.json();
  }

  function buildElements(graph) {
    return [
      ...(graph.nodes || []).map((node) => ({
        data: { id: node.id, label: node.label, path: node.path, kind: node.kind, summary: node.summary || "" },
      })),
      ...(graph.edges || []).map((edge, index) => ({
        data: { id: `e${index}`, source: edge.source, target: edge.target, kind: edge.kind },
      })),
    ];
  }

  function createClientEdgeId() {
    return `e${Date.now().toString(36)}${Math.random().toString(36).slice(2, 8)}`;
  }

  function roundCoord(value) {
    return Math.round(value * 100) / 100;
  }

  function compareNodeOrder(left, right) {
    const kindDiff = kindRank(left.data("kind")) - kindRank(right.data("kind"));
    if (kindDiff !== 0) {
      return kindDiff;
    }
    const labelDiff = nodeDisplayLabel(left).localeCompare(nodeDisplayLabel(right));
    if (labelDiff !== 0) {
      return labelDiff;
    }
    return String(left.id()).localeCompare(String(right.id()));
  }

  function defaultRoots(graph) {
    const nodes = graph.nodes || [];
    for (const kind of ["product-requirement", "requirement", "feature", "api", "component", "service"]) {
      const roots = nodes.filter((node) => node.kind === kind).map((node) => node.id);
      if (roots.length) {
        return roots;
      }
    }
    return undefined;
  }

  function layoutLabel(name) {
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

  function layerIndexForKind(kind) {
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

  function layoutOptions(graph, name) {
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
          padding: 40,
          avoidOverlap: true,
          sort: compareNodeOrder,
        };
      case "clustered":
        // Initial load falls back to layered; switching to clustered goes through runClusteredLayout
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
        // runLayeredLayout() overrides this with preset positions; this
        // breadthfirst config is only used as the initial Cytoscape pass before
        // runLayeredLayout repositions every node.
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

  function activeLayoutName(options) {
    return options.layoutSelect && options.layoutSelect.value ? options.layoutSelect.value : "layered";
  }

  function defaultOpenPath(options, path) {
    if (!options.artifactBaseUrl) {
      return;
    }
    const target = `${options.artifactBaseUrl}?path=${encodeURIComponent(path)}`;
    window.location.assign(target);
  }

  function renderGraph(options, graph) {
    if (options.metaElement) {
      const summary = `${graph.nodes.length} nodes / ${graph.edges.length} edges`;
      options.metaElement.dataset.summary = summary;
      options.metaElement.textContent = summary;
    }
    if (!graph.nodes.length) {
      options.container.innerHTML = `<pre style="padding: 16px; color: inherit;">${options.emptyMessage || "No traceability data found."}</pre>`;
      return undefined;
    }
    return cytoscape({
      container: options.container,
      elements: buildElements(graph),
      layout: layoutOptions(graph, activeLayoutName(options)),
      style: [
        {
          selector: "node",
          style: {
            label: "data(label)",
            shape: (ele) => shapeForKind(ele.data("kind")),
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
            "background-color": (ele) => palette[ele.data("kind")] || "#7a8791",
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
            "line-style": (ele) => lineStyleForKind(ele.data("kind")),
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

  function updateMetaSummary(options, nodeCount, edgeCount) {
    if (!options.metaElement) {
      return;
    }
    const summary = `${nodeCount} nodes / ${edgeCount} edges`;
    options.metaElement.dataset.summary = summary;
    options.metaElement.textContent = summary;
  }

  function setMetaStatus(options, message) {
    if (!options.metaElement) {
      return;
    }
    const summary = options.metaElement.dataset.summary || options.metaElement.textContent || "";
    options.metaElement.textContent = summary ? `${summary} • ${message}` : message;
  }

  function collectRelations(cy, omitEdgeId, appendEdges) {
    return cy.edges().toArray()
      .filter((edge) => edge.id() !== omitEdgeId)
      .map((edge) => ({
        source: edge.data("source"),
        target: edge.data("target"),
        kind: edge.data("kind"),
      }))
      .concat(appendEdges || [])
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

  async function persistRelations(options, cy, relationOptions) {
    const payload = { edges: collectRelations(cy, relationOptions && relationOptions.omitEdgeId, relationOptions && relationOptions.appendEdges) };
    if (typeof options.onSaveRelations === "function") {
      await options.onSaveRelations(payload);
      return;
    }
    if (!options.saveRelationsUrl) {
      return;
    }
    const response = await fetch(options.saveRelationsUrl, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
    if (!response.ok) {
      const message = await response.text();
      throw new Error(message || `relations request failed: ${response.status}`);
    }
  }

  function mount(options) {
    const container = options.container;
    if (!container) {
      throw new Error("container is required");
    }
    if (typeof cytoscape === "undefined") {
      container.innerHTML = '<pre style="padding: 16px; color: inherit;">cytoscape failed to load</pre>';
      return { fit() {} };
    }

    let cy;
    let selectedEdge;
    let selectedNode;
    let currentGraph = emptyGraph();
    const openPath = typeof options.onOpenPath === "function"
      ? options.onOpenPath
      : (path) => defaultOpenPath(options, path);
    const canSaveRelations = Boolean(options.saveRelationsUrl || typeof options.onSaveRelations === "function");

    function runLayout(layoutName, statusMessage) {
      if (!cy) {
        return;
      }
      if (layoutName === "clustered") {
        runClusteredLayout();
      } else if (layoutName === "layered") {
        runLayeredLayout();
      } else {
        cy.layout(layoutOptions(currentGraph, layoutName)).run();
      }
      if (statusMessage) {
        setMetaStatus(options, statusMessage);
      }
    }

    function runClusteredLayout() {
      if (!cy) {
        return;
      }
      // Step 1: pre-position nodes in concentric rings by kind rank
      const kindCount = Object.keys(kindOrder).length;
      const groups = {};
      cy.nodes().forEach((ele) => {
        const rank = kindRank(ele.data("kind"));
        const key = rank === Number.MAX_SAFE_INTEGER ? kindCount : rank;
        if (!groups[key]) {
          groups[key] = [];
        }
        groups[key].push(ele);
      });
      const sortedRings = Object.keys(groups).map(Number).sort((a, b) => a - b);
      const w = cy.width();
      const h = cy.height();
      const cx = w / 2;
      const ch = h / 2;
      const maxRadius = Math.min(w, h) * 0.45;
      sortedRings.forEach((ringKey, ringIdx) => {
        const r = maxRadius * (ringIdx + 1) / sortedRings.length;
        const nodes = groups[ringKey];
        nodes.forEach((node, i) => {
          const angle = (2 * Math.PI * i) / nodes.length - Math.PI / 2;
          node.position({ x: cx + r * Math.cos(angle), y: ch + r * Math.sin(angle) });
        });
      });
      // Step 2: run cose from these positions to minimize edge crossings
      cy.layout({
        name: "cose",
        fit: true,
        padding: 40,
        animate: false,
        randomize: false,
        nodeRepulsion: 4500,
        idealEdgeLength: 80,
        edgeElasticity: 100,
        gravity: 40,
        numIter: 1000,
      }).run();
    }

    function runLayeredLayout() {
      if (!cy) {
        return;
      }

      const nodes = cy.nodes().toArray();
      if (nodes.length === 0) {
        return;
      }

      // Group nodes by their fixed column (layer).
      const layers = new Map();
      nodes.forEach((node) => {
        const layer = layerIndexForKind(node.data("kind"));
        if (!layers.has(layer)) {
          layers.set(layer, []);
        }
        layers.get(layer).push(node);
      });

      const layerKeys = Array.from(layers.keys()).sort((left, right) => left - right);
      layerKeys.forEach((layer) => {
        layers.get(layer).sort(compareNodeOrder);
      });

      const edges = cy.edges().toArray().map((edge) => ({
        source: edge.data("source"),
        target: edge.data("target"),
      }));

      function indexByNodeId(layer) {
        const order = new Map();
        (layers.get(layer) || []).forEach((node, index) => {
          order.set(node.id(), index);
        });
        return order;
      }

      // Returns the sorted positions of a node's neighbours in a specific layer.
      function neighborPositions(nodeId, neighborOrder, useIncomingEdges) {
        const positions = [];
        for (const edge of edges) {
          if (useIncomingEdges) {
            if (edge.target === nodeId && neighborOrder.has(edge.source)) {
              positions.push(neighborOrder.get(edge.source));
            }
          } else if (edge.source === nodeId && neighborOrder.has(edge.target)) {
            positions.push(neighborOrder.get(edge.target));
          }
        }
        return positions;
      }

      // Counts edge crossings when node A is directly above node B.
      // An edge from A to position p and an edge from B to position q cross
      // when p > q (A's neighbour is lower than B's neighbour).
      function countCrossings(aPositions, bPositions) {
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

      // Transposition refinement: scan every adjacent pair in a column and swap
      // the pair when swapping reduces the actual crossing count against both
      // neighbour layers. Repeats until no improvement or sweep limit reached.
      function transposeLayer(layer) {
        const orderedNodes = layers.get(layer);
        if (!orderedNodes || orderedNodes.length < 2) {
          return;
        }
        const layerPos = layerKeys.indexOf(layer);
        const leftOrder = layerPos > 0 ? indexByNodeId(layerKeys[layerPos - 1]) : new Map();
        const rightOrder = layerPos < layerKeys.length - 1 ? indexByNodeId(layerKeys[layerPos + 1]) : new Map();
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
            const current = countCrossings(uLeft, vLeft) + countCrossings(uRight, vRight);
            const swapped = countCrossings(vLeft, uLeft) + countCrossings(vRight, uRight);
            if (swapped < current) {
              orderedNodes[i] = v;
              orderedNodes[i + 1] = u;
              improved = true;
            }
          }
        }
      }

      // Phase 1: organic pre-pass — run a virtual cose layout (no animation) so
      // the physics engine clusters connected nodes into coherent regions of the
      // canvas. We capture every node's resulting y-coordinate as a "semantic
      // height" signal and then discard the x-coordinates.
      // randomize: false starts from current positions for determinism.
      cy.layout({
        name: "cose",
        fit: false,
        animate: false,
        randomize: false,
        nodeRepulsion: 4500,
        idealEdgeLength: 80,
        edgeElasticity: 100,
        gravity: 40,
        numIter: layeredLayoutTuning.organicPrepassIterations,
      }).run();

      // Capture organic y-positions before they are overwritten by column placement.
      const organicY = new Map();
      nodes.forEach((node) => organicY.set(node.id(), node.position().y));

      // Sort the anchor column (requirements, layer 1) by its own organic y.
      // This is the reference: the physics engine has already arranged connected
      // requirements into coherent vertical bands, so this order is correct.
      //
      // All other columns sort by the MEAN organic y of their direct neighbours.
      // A product requirement connected to REQ-1(y=100) and REQ-2(y=200) gets
      // score=150, placing it in the same vertical band as those requirements.
      // Unconnected nodes fall back to their own organic y.
      const anchorLayer = layerKeys.includes(1) ? 1 : layerKeys[Math.floor(layerKeys.length / 2)];

      function meanNeighbourOrganicY(nodeId) {
        const ys = [];
        for (const edge of edges) {
          if (edge.source === nodeId && organicY.has(edge.target)) {
            ys.push(organicY.get(edge.target));
          }
          if (edge.target === nodeId && organicY.has(edge.source)) {
            ys.push(organicY.get(edge.source));
          }
        }
        return ys.length
          ? ys.reduce((sum, y) => sum + y, 0) / ys.length
          : (organicY.get(nodeId) || 0);
      }

      layerKeys.forEach((layer) => {
        if (layer === anchorLayer) {
          layers.get(layer).sort((a, b) => organicY.get(a.id()) - organicY.get(b.id()));
        } else {
          layers.get(layer).sort((a, b) => meanNeighbourOrganicY(a.id()) - meanNeighbourOrganicY(b.id()));
        }
      });

      // Phase 2: transposition refinement — eliminates any residual local
      // crossings introduced when nodes snap back to fixed column x-positions.
      layerKeys.forEach((layer) => transposeLayer(layer));

      // Place nodes at absolute positions using fixed spacing. The canvas
      // expands automatically; cy.fit() zooms to show all nodes with padding.
      cy.startBatch();
      layerKeys.forEach((layer, layerPosition) => {
        const x = layerPosition * layeredLayoutTuning.columnSpacingX;
        (layers.get(layer) || []).forEach((node, nodeIndex) => {
          node.position({ x: roundCoord(x), y: roundCoord(nodeIndex * layeredLayoutTuning.nodeSpacingY) });
        });
      });
      cy.endBatch();
      cy.fit(undefined, layeredLayoutTuning.padding);
    }

    function applyFilter(filterText) {
      if (!cy) {
        return;
      }
      const text = filterText.trim().toLowerCase();
      if (!text) {
        cy.nodes().removeClass("traceability-dimmed");
        return;
      }
      cy.nodes().forEach((node) => {
        const label = String(node.data("label") || node.id()).toLowerCase();
        const summary = String(node.data("summary") || "").toLowerCase();
        if (label.includes(text) || summary.includes(text)) {
          node.removeClass("traceability-dimmed");
        } else {
          node.addClass("traceability-dimmed");
        }
      });
    }

    function shortID(nodeID) {
      // Extract the PREFIX-NNN identifier from the last path segment.
      // e.g. "model/requirements/lint/LNT-001-style-rule-enforcement" -> "LNT-001"
      // Falls back to the full segment when the pattern is not found.
      const segment = String(nodeID || "").split("/").pop() || nodeID;
      const m = segment.match(/^([A-Z]+-\d+)/);
      return m ? m[1] : segment;
    }

    function describeNode(node) {
      if (!node) {
        return undefined;
      }
      const position = node.position();
      const summary = node.data("summary");
      const kind = displayKind(node.data("kind"));
      const inspectorID = shortID(node.id());
      const renderedRows = detailsRowsMarkup([
        { label: "Path", value: node.data("path") },
      ]);
      const summaryMarkup = summary ? `<p class="details-summary">${escapeHTML(summary)}</p>` : "";
      return `<article class="details-panel details-panel-node"><div class="details-header"><div class="details-header-top"><p class="details-node-meta"><span class="details-eyebrow">NODE</span><span class="details-node-kind">${escapeHTML(kind)}</span><span class="details-node-separator">:</span><span class="details-node-id">${escapeHTML(inspectorID)}</span></p>${detailsIconButton("Open artifact", node.data("path"))}</div><h2 class="details-title">${escapeHTML(node.data("label") || node.id())}</h2>${summaryMarkup}</div>${renderedRows}</article>`;
    }

    function resolveNode(nodeID) {
      if (!cy) {
        return undefined;
      }
      const matches = cy.$id(nodeID);
      return matches.length ? matches[0] : undefined;
    }

    function describeEdge(edge) {
      if (!edge) {
        return undefined;
      }
      const kind = edge.data("kind");
      const spec = relationSpec(kind);
      const sourceNode = resolveNode(edge.data("source"));
      const targetNode = resolveNode(edge.data("target"));
      return detailsMarkup("Relation", spec.label, [
        { label: "Kind", value: displayKind(kind) },
        { label: "Source", value: sourceNode ? sourceNode.data("label") || sourceNode.id() : edge.data("source") },
        { label: "Source ID", value: edge.data("source") },
        { label: "Target", value: targetNode ? targetNode.data("label") || targetNode.id() : edge.data("target") },
        { label: "Target ID", value: edge.data("target") },
      ], canSaveRelations ? "Use the remove button to delete this relation. The UI asks for confirmation before persisting the change." : "");
    }

    function updateDetailsPanel() {
      if (!options.detailsElement) {
        return;
      }
      if (selectedEdge) {
        setDetails(options, describeEdge(selectedEdge));
        return;
      }
      if (selectedNode) {
        setDetails(options, describeNode(selectedNode));
        return;
      }
      setDetails(options, detailsMarkup("Inspector", "No selection", [], "Select a node or edge to inspect its details."));
    }

    function setCreateStatus(message) {
      setMetaStatus(options, message);
    }

    function clearDragTargetClasses() {
      if (!cy) {
        return;
      }
      cy.nodes().removeClass("traceability-create-inactive traceability-create-source traceability-create-target");
    }

    function currentSelectedEdge() {
      if (selectedEdge && !selectedEdge.removed()) {
        return selectedEdge;
      }
      if (!cy) {
        return undefined;
      }
      const selectedEdges = cy.$("edge:selected");
      return selectedEdges.length > 0 ? selectedEdges[0] : undefined;
    }

    function relationAlreadyExists(source, target, kind) {
      if (!cy) {
        return false;
      }
      return cy.edges().some((edge) => edge.data("source") === source && edge.data("target") === target && edge.data("kind") === kind);
    }

    async function completeDragConnect(sourceNode, targetNode) {
      if (!cy) {
        return;
      }
      const sourceKind = sourceNode.data("kind");
      const targetKind = targetNode.data("kind");
      const kind = resolveRelationKindForPair(options, sourceKind, targetKind);
      if (!kind) {
        setMetaStatus(options, `no relation defined between ${displayKind(sourceKind)} and ${displayKind(targetKind)}`);
        return;
      }
      const edge = { source: sourceNode.id(), target: targetNode.id(), kind };
      if (relationAlreadyExists(edge.source, edge.target, edge.kind)) {
        setMetaStatus(options, "edge already exists");
        return;
      }
      try {
        await persistRelations(options, cy, { appendEdges: [edge] });
        const addedEdge = cy.add({ data: { id: createClientEdgeId(), ...edge } });
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

    function updateRemoveEdgeButton() {
      if (!options.removeEdgeButton) {
        return;
      }
      selectedEdge = currentSelectedEdge();
      options.removeEdgeButton.disabled = !canSaveRelations || !selectedEdge;
    }

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
          cy.zoom({ level: cy.zoom() * 1.25, renderedPosition: { x: cy.width() / 2, y: cy.height() / 2 } });
        }
      });
    }

    if (options.zoomOutButton) {
      options.zoomOutButton.addEventListener("click", () => {
        if (cy) {
          cy.zoom({ level: cy.zoom() * 0.8, renderedPosition: { x: cy.width() / 2, y: cy.height() / 2 } });
        }
      });
    }

    if (options.layoutSelect) {
      options.layoutSelect.addEventListener("change", () => {
        const layoutName = activeLayoutName(options);
        runLayout(layoutName, `${layoutLabel(layoutName)} layout applied`);
      });
    }

    if (options.filterInput) {
      options.filterInput.addEventListener("input", () => {
        applyFilter(options.filterInput.value);
      });
    }

    if (options.removeEdgeButton) {
      updateRemoveEdgeButton();
      options.removeEdgeButton.addEventListener("click", async () => {
        const edgeToRemove = currentSelectedEdge();
        if (!edgeToRemove || !canSaveRelations || !cy) {
          return;
        }
        const removalPreview = {
          source: edgeToRemove.data("source"),
          target: edgeToRemove.data("target"),
          kind: edgeToRemove.data("kind"),
          sourceLabel: nodeDisplayLabel(resolveNode(edgeToRemove.data("source"))),
          targetLabel: nodeDisplayLabel(resolveNode(edgeToRemove.data("target"))),
        };
        if (!confirmRelationChange(options, "Remove", removalPreview)) {
          setMetaStatus(options, "edge removal cancelled");
          updateDetailsPanel();
          return;
        }
        options.removeEdgeButton.disabled = true;
        try {
          await persistRelations(options, cy, { omitEdgeId: edgeToRemove.id() });
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
        const path = action.dataset.openPath;
        if (!path) {
          return;
        }
        openPath(path);
      });
    }

    resolveGraph(options)
      .then((graph) => {
        currentGraph = graph;
        cy = renderGraph(options, graph) || undefined;
        if (cy) {
          if (activeLayoutName(options) === "layered") {
            runLayeredLayout();
          }
          updateMetaSummary(options, cy.nodes().length, cy.edges().length);
          updateDetailsPanel();
          if (options.filterInput) {
            applyFilter(options.filterInput.value);
          }

          // ── Drag-to-connect handle ─────────────────────────────────────────
          let dragSourceNode = null;
          let isDraggingEdge = false;
          let hideHandleTimer = null;

          const handle = typeof document !== "undefined" ? document.createElement("div") : null;
          const svgNS = "http://www.w3.org/2000/svg";
          const dragSvg = typeof document !== "undefined" ? document.createElementNS(svgNS, "svg") : null;
          const dragLine = dragSvg ? document.createElementNS(svgNS, "line") : null;

          if (handle && dragSvg && dragLine && canSaveRelations) {
            handle.className = "drag-connect-handle";
            handle.style.display = "none";
            document.body.appendChild(handle);

            dragSvg.setAttribute("class", "drag-connect-overlay");
            dragSvg.style.cssText = "position:fixed;top:0;left:0;width:100vw;height:100vh;pointer-events:none;z-index:9000;display:none;";
            dragLine.setAttribute("class", "drag-connect-line");
            dragSvg.appendChild(dragLine);
            document.body.appendChild(dragSvg);

            function renderedToViewport(renderedPos) {
              const rect = container.getBoundingClientRect();
              return { x: rect.left + renderedPos.x, y: rect.top + renderedPos.y };
            }

            function findNodeAtViewport(clientX, clientY) {
              const rect = container.getBoundingClientRect();
              const rx = clientX - rect.left;
              const ry = clientY - rect.top;
              let found;
              cy.nodes().forEach((node) => {
                const bb = node.renderedBoundingBox();
                if (rx >= bb.x1 && rx <= bb.x2 && ry >= bb.y1 && ry <= bb.y2) {
                  found = node;
                }
              });
              return found;
            }

            function highlightDragTargets(srcNode) {
              cy.nodes().forEach((node) => {
                if (node.same(srcNode)) {
                  node.addClass("traceability-create-source");
                  return;
                }
                const kind = resolveRelationKindForPair(options, srcNode.data("kind"), node.data("kind"));
                if (kind) {
                  node.addClass("traceability-create-target");
                } else {
                  node.addClass("traceability-create-inactive");
                }
              });
            }

            function showHandle(node) {
              const bb = node.renderedBoundingBox();
              const rect = container.getBoundingClientRect();
              const cx = rect.left + (bb.x1 + bb.x2) / 2;
              const cy2 = rect.top + bb.y1 - 4;
              const size = Math.round(Math.max(8, Math.min(20, 20 * cy.zoom())));
              const half = size / 2;
              handle.style.width = `${size}px`;
              handle.style.height = `${size}px`;
              handle.style.left = `${cx - half}px`;
              handle.style.top = `${cy2 - half}px`;
              handle.style.display = "flex";
            }

            function hideHandle() {
              if (!isDraggingEdge) {
                handle.style.display = "none";
              }
            }

            function scheduleHideHandle() {
              hideHandleTimer = window.setTimeout(hideHandle, 120);
            }

            function cancelHideHandle() {
              if (hideHandleTimer) {
                window.clearTimeout(hideHandleTimer);
                hideHandleTimer = null;
              }
            }

            handle.addEventListener("mouseenter", cancelHideHandle);
            handle.addEventListener("mouseleave", scheduleHideHandle);

            handle.addEventListener("mousedown", (event) => {
              const handleSize = Math.round(Math.max(8, Math.min(20, 20 * cy.zoom())));
              const nodeUnder = findNodeAtViewport(event.clientX, event.clientY + handleSize * 0.7);
              if (!nodeUnder) {
                return;
              }
              event.stopPropagation();
              event.preventDefault();
              isDraggingEdge = true;
              dragSourceNode = nodeUnder;
              highlightDragTargets(dragSourceNode);
              const srcVP = renderedToViewport(dragSourceNode.renderedPosition());
              dragLine.setAttribute("x1", srcVP.x);
              dragLine.setAttribute("y1", srcVP.y);
              dragLine.setAttribute("x2", event.clientX);
              dragLine.setAttribute("y2", event.clientY);
              dragSvg.style.display = "block";

              function onMouseMove(e) {
                const srcVP2 = renderedToViewport(dragSourceNode.renderedPosition());
                dragLine.setAttribute("x1", srcVP2.x);
                dragLine.setAttribute("y1", srcVP2.y);
                dragLine.setAttribute("x2", e.clientX);
                dragLine.setAttribute("y2", e.clientY);
              }

              function onMouseUp(e) {
                document.removeEventListener("mousemove", onMouseMove);
                document.removeEventListener("mouseup", onMouseUp);
                isDraggingEdge = false;
                dragSvg.style.display = "none";
                hideHandle();
                clearDragTargetClasses();
                const targetNode = findNodeAtViewport(e.clientX, e.clientY);
                if (targetNode && !targetNode.same(dragSourceNode)) {
                  void completeDragConnect(dragSourceNode, targetNode);
                } else {
                  setMetaStatus(options, "drag cancelled");
                }
                dragSourceNode = null;
              }

              document.addEventListener("mousemove", onMouseMove);
              document.addEventListener("mouseup", onMouseUp);
            });

            cy.on("mouseover", "node", (event) => {
              if (!isDraggingEdge) {
                cancelHideHandle();
                showHandle(event.target);
              }
            });

            cy.on("mouseout", "node", () => {
              scheduleHideHandle();
            });
          }
          // ── end drag-to-connect ───────────────────────────────────────────

          cy.on("tap", (event) => {
            if (event.target === cy) {
              selectedEdge = undefined;
              selectedNode = undefined;
              updateRemoveEdgeButton();
              updateDetailsPanel();
            }
          });
          cy.on("tap", "edge", (event) => {
            selectedEdge = event.target;
            selectedNode = undefined;
            updateRemoveEdgeButton();
            updateDetailsPanel();
          });
          cy.on("tap", "node", (event) => {
            selectedEdge = undefined;
            selectedNode = event.target;
            updateRemoveEdgeButton();
            updateDetailsPanel();
          });
        }
      })
      .catch((error) => {
        container.innerHTML = `<pre style="padding: 16px; color: inherit;">${String(error)}</pre>`;
      });

    return {
      fit() {
        if (cy) {
          cy.fit(undefined, 40);
        }
      },
    };
  }

  window.TraceabilityUI = { mount };
})();