(function () {
  const palette = {
    "product-requirement": "#e66b6b",
    requirement: "#4f8bd6",
    feature: "#e29c45",
    component: "#5f9d72",
    api: "#7b6ccf",
    service: "#c7739f",
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
      ],
    });
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
    const openPath = typeof options.onOpenPath === "function"
      ? options.onOpenPath
      : (path) => defaultOpenPath(options, path);
    const canSaveLayout = Boolean(options.saveLayoutUrl || typeof options.onSaveLayout === "function");

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

    resolveGraph(options)
      .then((graph) => {
        cy = renderGraph(options, graph) || undefined;
        if (cy) {
          cy.on("tap", "node", (event) => {
            const path = event.target.data("path");
            if (path) {
              openPath(path);
            }
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