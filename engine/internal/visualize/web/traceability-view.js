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
    const renderedRows = (rows || [])
      .filter((row) => row && row.value != null && row.value !== "")
      .map((row) => `<div><dt>${escapeHTML(row.label)}</dt><dd>${escapeHTML(row.value)}</dd></div>`)
      .join("");
    const noteMarkup = note ? `<p class="details-note">${escapeHTML(note)}</p>` : "";
    return `<article class="details-panel"><p class="details-eyebrow">${escapeHTML(eyebrow)}</p><h2 class="details-title">${escapeHTML(title)}</h2>${renderedRows ? `<dl class="details-list">${renderedRows}</dl>` : ""}${noteMarkup}</article>`;
  }

  function detailsActionButton(label, path) {
    if (!path) {
      return "";
    }
    return `<div class="details-actions"><button type="button" class="details-open-button" data-open-path="${escapeHTML(path)}">${escapeHTML(label)}</button></div>`;
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

  function confirmRelationChange(action, relation) {
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
        data: { id: node.id, label: node.label, path: node.path, kind: node.kind },
        position: node.layout ? { x: node.layout.x, y: node.layout.y } : undefined,
        locked: Boolean(node.layout && node.layout.locked),
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

  function hasLayout(graph) {
    return (graph.nodes || []).length > 0 && (graph.nodes || []).every((node) => Boolean(node.layout));
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
      layout: hasLayout(graph)
        ? { name: "preset", padding: 32, fit: true }
        : { name: "breadthfirst", directed: true, padding: 40, spacingFactor: 1.15, avoidOverlap: true },
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

  function collectLayout(cy) {
    return cy.nodes().map((node) => {
      const position = node.position();
      return {
        id: node.id(),
        x: roundCoord(position.x),
        y: roundCoord(position.y),
        locked: Boolean(node.locked()),
      };
    }).sort((left, right) => left.id.localeCompare(right.id));
  }

  async function persistLayout(options, cy) {
    const payload = { nodes: collectLayout(cy) };
    if (typeof options.onSaveLayout === "function") {
      await options.onSaveLayout(payload);
      return;
    }
    if (!options.saveLayoutUrl) {
      return;
    }
    const response = await fetch(options.saveLayoutUrl, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
    if (!response.ok) {
      const message = await response.text();
      throw new Error(message || `layout request failed: ${response.status}`);
    }
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
    let createEdgeMode = false;
    let edgeSourceNode;
    const openPath = typeof options.onOpenPath === "function"
      ? options.onOpenPath
      : (path) => defaultOpenPath(options, path);
    const canSaveLayout = Boolean(options.saveLayoutUrl || typeof options.onSaveLayout === "function");
    const canSaveRelations = Boolean(options.saveRelationsUrl || typeof options.onSaveRelations === "function");

    function currentRelationKind() {
      return options.relationKindSelect && options.relationKindSelect.value
        ? options.relationKindSelect.value
        : "realization";
    }

    function currentRelationSpec() {
      return relationSpec(currentRelationKind());
    }

    function describeNode(node) {
      if (!node) {
        return undefined;
      }
      const position = node.position();
      return `${detailsMarkup("Node", node.data("label") || node.id(), [
        { label: "Kind", value: displayKind(node.data("kind")) },
        { label: "ID", value: node.id() },
        { label: "Path", value: node.data("path") },
        { label: "Position", value: `${roundCoord(position.x)}, ${roundCoord(position.y)}` },
        { label: "Locked", value: node.locked() ? "yes" : "no" },
      ])}${detailsActionButton("Open artifact", node.data("path"))}`;
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
      ], canSaveRelations ? "Use Remove Selected Edge to delete this relation. The UI asks for confirmation before persisting the change." : "");
    }

    function describeCreateMode() {
      const spec = currentRelationSpec();
      if (edgeSourceNode) {
        return detailsMarkup("Add Edge", edgeSourceNode.data("label") || edgeSourceNode.id(), [
          { label: "Relation", value: spec.label },
          { label: "Source kind", value: displayKind(spec.sourceKind) },
          { label: "Target kind", value: displayKind(spec.targetKind) },
          { label: "Source ID", value: edgeSourceNode.id() },
        ], `Select a ${displayKind(spec.targetKind)} target to create this relation. The UI asks for confirmation before saving it.`);
      }
      return detailsMarkup("Add Edge", spec.label, [
        { label: "Source kind", value: displayKind(spec.sourceKind) },
        { label: "Target kind", value: displayKind(spec.targetKind) },
      ], `Select a ${displayKind(spec.sourceKind)} source node to start. The UI asks for confirmation before saving the relation.`);
    }

    function updateDetailsPanel() {
      if (!options.detailsElement) {
        return;
      }
      if (createEdgeMode) {
        setDetails(options, describeCreateMode());
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

    function clearCreateClasses() {
      if (!cy) {
        return;
      }
      cy.nodes().removeClass("traceability-create-inactive traceability-create-source traceability-create-target");
    }

    function updateCreateClasses() {
      if (!cy) {
        return;
      }
      clearCreateClasses();
      if (!createEdgeMode) {
        return;
      }
      const spec = currentRelationSpec();
      const candidateKind = edgeSourceNode ? spec.targetKind : spec.sourceKind;
      cy.nodes().forEach((node) => {
        if (edgeSourceNode && edgeSourceNode.same(node)) {
          node.addClass("traceability-create-source");
          return;
        }
        if (node.data("kind") === candidateKind) {
          node.addClass(edgeSourceNode ? "traceability-create-target" : "traceability-create-source");
          return;
        }
        node.addClass("traceability-create-inactive");
      });
    }

    function invalidCreateSelectionMessage(stage) {
      const spec = currentRelationSpec();
      if (stage === "source") {
        return `choose a ${displayKind(spec.sourceKind)} source for ${spec.label.toLowerCase()}`;
      }
      return `choose a ${displayKind(spec.targetKind)} target for ${spec.label.toLowerCase()}`;
    }

    function clearEdgeSourceSelection() {
      if (edgeSourceNode) {
        edgeSourceNode.unselect();
      }
      edgeSourceNode = undefined;
      updateCreateClasses();
    }

    function exitCreateEdgeMode() {
      createEdgeMode = false;
      clearEdgeSourceSelection();
      clearCreateClasses();
      if (options.addEdgeButton) {
        options.addEdgeButton.textContent = "Add Edge";
      }
      updateDetailsPanel();
    }

    function enterCreateEdgeMode() {
      createEdgeMode = true;
      selectedEdge = undefined;
      selectedNode = undefined;
      updateRemoveEdgeButton();
      if (options.addEdgeButton) {
        options.addEdgeButton.textContent = "Cancel Add";
      }
      updateCreateClasses();
      setCreateStatus(invalidCreateSelectionMessage("source").replace("choose a ", "select a "));
      updateDetailsPanel();
    }

    function relationAlreadyExists(source, target, kind) {
      if (!cy) {
        return false;
      }
      return cy.edges().some((edge) => edge.data("source") === source && edge.data("target") === target && edge.data("kind") === kind);
    }

    async function addRelation(targetNode) {
      if (!cy || !edgeSourceNode) {
        return;
      }
      const nextEdge = {
        source: edgeSourceNode.id(),
        target: targetNode.id(),
        kind: currentRelationKind(),
      };
      const nextEdgePreview = {
        ...nextEdge,
        sourceLabel: nodeDisplayLabel(edgeSourceNode),
        targetLabel: nodeDisplayLabel(targetNode),
      };
      if (nextEdge.source === nextEdge.target) {
        setMetaStatus(options, "edge source and target must differ");
        return;
      }
      if (relationAlreadyExists(nextEdge.source, nextEdge.target, nextEdge.kind)) {
        setMetaStatus(options, "edge already exists");
        return;
      }
      if (!confirmRelationChange("Add", nextEdgePreview)) {
        setMetaStatus(options, "edge creation cancelled");
        return;
      }
      const originalLabel = options.addEdgeButton ? options.addEdgeButton.textContent : "Add Edge";
      let saveSucceeded = false;
      if (options.addEdgeButton) {
        options.addEdgeButton.disabled = true;
        options.addEdgeButton.textContent = "Adding...";
      }
      try {
        await persistRelations(options, cy, { appendEdges: [nextEdge] });
        const addedEdge = cy.add({ data: { id: createClientEdgeId(), source: nextEdge.source, target: nextEdge.target, kind: nextEdge.kind } });
        selectedNode = undefined;
        selectedEdge = addedEdge[0];
        updateMetaSummary(options, cy.nodes().length, cy.edges().length);
        updateRemoveEdgeButton();
        setMetaStatus(options, "edge added");
        saveSucceeded = true;
        exitCreateEdgeMode();
      } catch (error) {
        console.error(error);
        setMetaStatus(options, "edge creation failed");
        if (options.addEdgeButton) {
          options.addEdgeButton.textContent = "Add Failed";
        }
        window.setTimeout(() => {
          if (options.addEdgeButton) {
            options.addEdgeButton.textContent = originalLabel;
            options.addEdgeButton.disabled = !canSaveRelations;
          }
        }, 1200);
        return;
      }
      if (options.addEdgeButton) {
        options.addEdgeButton.textContent = saveSucceeded ? "Add Edge" : originalLabel;
        options.addEdgeButton.disabled = !canSaveRelations;
      }
    }

    function updateRemoveEdgeButton() {
      if (!options.removeEdgeButton) {
        return;
      }
      options.removeEdgeButton.disabled = !canSaveRelations || !selectedEdge;
    }

    if (options.relationKindSelect) {
      options.relationKindSelect.disabled = !canSaveRelations;
      options.relationKindSelect.addEventListener("change", () => {
        if (!createEdgeMode) {
          return;
        }
        clearEdgeSourceSelection();
        updateCreateClasses();
        setCreateStatus(invalidCreateSelectionMessage("source").replace("choose a ", "select a "));
        updateDetailsPanel();
      });
    }

    if (options.addEdgeButton) {
      options.addEdgeButton.disabled = !canSaveRelations;
      options.addEdgeButton.addEventListener("click", () => {
        if (!canSaveRelations) {
          return;
        }
        if (createEdgeMode) {
          exitCreateEdgeMode();
          setMetaStatus(options, "edge creation cancelled");
          return;
        }
        enterCreateEdgeMode();
      });
    }

    if (options.fitButton) {
      options.fitButton.addEventListener("click", () => {
        if (cy) {
          cy.fit(undefined, 40);
        }
      });
    }

    if (options.saveButton) {
      options.saveButton.disabled = !canSaveLayout;
      options.saveButton.addEventListener("click", async () => {
        if (!cy || !canSaveLayout) {
          return;
        }
        const originalLabel = options.saveButton.textContent;
        options.saveButton.disabled = true;
        options.saveButton.textContent = "Saving...";
        try {
          await persistLayout(options, cy);
          options.saveButton.textContent = "Saved";
          setMetaStatus(options, "layout saved");
        } catch (error) {
          console.error(error);
          options.saveButton.textContent = "Save Failed";
          setMetaStatus(options, "layout save failed");
        } finally {
          window.setTimeout(() => {
            options.saveButton.disabled = !canSaveLayout;
            options.saveButton.textContent = originalLabel;
          }, 1200);
        }
      });
    }

    if (options.removeEdgeButton) {
      updateRemoveEdgeButton();
      options.removeEdgeButton.addEventListener("click", async () => {
        if (!selectedEdge || !canSaveRelations || !cy) {
          return;
        }
        const edgeToRemove = selectedEdge;
        const removalPreview = {
          source: edgeToRemove.data("source"),
          target: edgeToRemove.data("target"),
          kind: edgeToRemove.data("kind"),
          sourceLabel: nodeDisplayLabel(resolveNode(edgeToRemove.data("source"))),
          targetLabel: nodeDisplayLabel(resolveNode(edgeToRemove.data("target"))),
        };
        if (!confirmRelationChange("Remove", removalPreview)) {
          setMetaStatus(options, "edge removal cancelled");
          updateDetailsPanel();
          return;
        }
        const originalLabel = options.removeEdgeButton.textContent;
        options.removeEdgeButton.disabled = true;
        options.removeEdgeButton.textContent = "Removing...";
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
          options.removeEdgeButton.textContent = "Remove Failed";
          setMetaStatus(options, "edge removal failed");
        } finally {
          window.setTimeout(() => {
            options.removeEdgeButton.textContent = originalLabel;
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
        const path = target.dataset.openPath;
        if (!path) {
          return;
        }
        openPath(path);
      });
    }

    resolveGraph(options)
      .then((graph) => {
        cy = renderGraph(options, graph) || undefined;
        if (cy) {
          updateMetaSummary(options, cy.nodes().length, cy.edges().length);
          updateDetailsPanel();
          cy.on("tap", (event) => {
            if (event.target === cy) {
              if (createEdgeMode) {
                clearEdgeSourceSelection();
                setCreateStatus(invalidCreateSelectionMessage("source").replace("choose a ", "select a "));
              }
              selectedEdge = undefined;
              selectedNode = undefined;
              updateRemoveEdgeButton();
              updateDetailsPanel();
            }
          });
          cy.on("tap", "edge", (event) => {
            if (createEdgeMode) {
              return;
            }
            selectedEdge = event.target;
            selectedNode = undefined;
            updateRemoveEdgeButton();
            updateDetailsPanel();
          });
          cy.on("tap", "node", (event) => {
            if (createEdgeMode) {
              const tappedNode = event.target;
              const spec = currentRelationSpec();
              if (!edgeSourceNode) {
                if (tappedNode.data("kind") !== spec.sourceKind) {
                  setCreateStatus(invalidCreateSelectionMessage("source"));
                  return;
                }
                edgeSourceNode = tappedNode;
                edgeSourceNode.select();
                updateCreateClasses();
                setCreateStatus(`select a ${displayKind(spec.targetKind)} target for ${spec.label.toLowerCase()}`);
                updateDetailsPanel();
                return;
              }
              if (edgeSourceNode.same(tappedNode)) {
                setCreateStatus("choose a different target node");
                return;
              }
              if (tappedNode.data("kind") !== spec.targetKind) {
                setCreateStatus(invalidCreateSelectionMessage("target"));
                return;
              }
              void addRelation(tappedNode);
              return;
            }
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