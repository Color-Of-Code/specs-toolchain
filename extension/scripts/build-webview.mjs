import esbuild from "esbuild";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const repoRoot = path.resolve(__dirname, "../..");
const webSrcDir = path.join(repoRoot, "engine", "internal", "visualize", "web", "src");
const webOutDir = path.join(repoRoot, "engine", "internal", "visualize", "web");

// Bundle TS → JS and CSS → CSS using esbuild.
await esbuild.build({
  entryPoints: [
    { in: path.join(webSrcDir, "index.ts"),                out: "traceability-view" },
    { in: path.join(webSrcDir, "traceability-view.css"),   out: "traceability-view" },
  ],
  bundle: true,
  outdir: webOutDir,
  format: "iife",
  globalName: "TraceabilityUI",
  platform: "browser",
  target: ["es2020"],
});

console.log(`built   ${path.join(webOutDir, "traceability-view.js")}`);
console.log(`built   ${path.join(webOutDir, "traceability-view.css")}`);
