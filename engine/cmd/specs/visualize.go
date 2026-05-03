package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/Color-Of-Code/specs-toolchain/engine/internal/config"
	"github.com/Color-Of-Code/specs-toolchain/engine/internal/graph"
	"github.com/Color-Of-Code/specs-toolchain/engine/internal/visualize"
)

// cmdVisualize dispatches `specs visualize <subcommand>`.
func cmdVisualize(args []string) error {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: specs visualize <subcommand>")
		fmt.Fprintln(os.Stderr, "Subcommands: traceability")
		return exitWith(2, "missing subcommand")
	}
	switch args[0] {
	case "traceability":
		return cmdVisualizeTraceability(args[1:])
	case "-h", "--help", "help":
		fmt.Fprintln(os.Stderr, "Usage: specs visualize <traceability> [flags]")
		return nil
	default:
		return exitWith(2, "unknown visualize subcommand %q", args[0])
	}
}

// cmdVisualizeTraceability emits the requirement <-> implementer graph.
//
//	--format dot|mermaid|json    output format (default: dot)
//	--out <path>                 file to write (default: stdout)
//	--serve                      run a local web server instead of writing a file
//	--listen <addr>              listen address for --serve (default: 127.0.0.1:8090)
func cmdVisualizeTraceability(args []string) error {
	fs := flag.NewFlagSet("visualize traceability", flag.ContinueOnError)
	format := fs.String("format", "dot", "output format: dot | mermaid | json")
	outPath := fs.String("out", "", "write to file (use - or empty for stdout)")
	serve := fs.Bool("serve", false, "run a local web server instead of writing a file")
	listenAddr := fs.String("listen", "127.0.0.1:8090", "listen address for --serve")
	fs.StringVar(outPath, "o", "", "shorthand for --out")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: specs visualize traceability [--format dot|mermaid|json] [--out <path>|-] [--serve] [--listen <addr>]")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *serve {
		handler, err := newTraceabilityUIHandler("")
		if err != nil {
			return err
		}
		listener, err := net.Listen("tcp", *listenAddr)
		if err != nil {
			return exitWith(1, "listen %s: %v", *listenAddr, err)
		}
		defer listener.Close()
		fmt.Fprintf(os.Stderr, "serving traceability UI on http://%s\n", listener.Addr().String())
		server := &http.Server{Handler: handler}
		err = server.Serve(listener)
		if err != nil && err != http.ErrServerClosed {
			return exitWith(1, "serve traceability ui: %v", err)
		}
		return nil
	}

	_, g, err := loadTraceabilityVisualization("")
	if err != nil {
		return err
	}

	return writeTraceabilityOutput(g, *format, *outPath)
}

func loadTraceabilityVisualization(start string) (*config.Resolved, *visualize.Graph, error) {
	cfg, err := config.Load(start)
	if err != nil {
		return nil, nil, err
	}
	traceability, err := graph.Load(cfg.GraphManifest)
	if err != nil {
		return nil, nil, exitWith(1, "load graph %s: %v", cfg.GraphManifest, err)
	}
	g, err := visualize.Build(cfg.ModelDir, cfg.ProductDir, traceability)
	if err != nil {
		return nil, nil, exitWith(1, "%v", err)
	}
	return cfg, g, nil
}

func writeTraceabilityOutput(g *visualize.Graph, format, outPath string) error {
	var out *os.File = os.Stdout
	if outPath != "" && outPath != "-" {
		f, err := os.Create(outPath)
		if err != nil {
			return exitWith(1, "create %s: %v", outPath, err)
		}
		defer f.Close()
		out = f
	}

	switch format {
	case "dot":
		if err := visualize.WriteDOT(out, g); err != nil {
			return err
		}
	case "mermaid":
		if err := visualize.WriteMermaid(out, g); err != nil {
			return err
		}
	case "json":
		if err := visualize.WriteJSON(out, g); err != nil {
			return err
		}
	default:
		return exitWith(2, "unknown --format %q (want dot, mermaid, or json)", format)
	}
	if outPath != "" && outPath != "-" {
		fmt.Fprintf(os.Stderr, "wrote %s (%d node(s), %d edge(s))\n", outPath, len(g.Nodes), len(g.Edges))
	}
	return nil
}
