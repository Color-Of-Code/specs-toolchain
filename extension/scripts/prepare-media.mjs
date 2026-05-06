import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const extensionDir = path.resolve(__dirname, "..");
const repoRoot = path.resolve(extensionDir, "..");
const mediaDir = path.join(extensionDir, "media");

const assets = [
  {
    source: path.join(repoRoot, "engine", "internal", "visualize", "web", "traceability-view.js"),
    target: path.join(mediaDir, "traceability-view.js"),
  },
  {
    source: path.join(repoRoot, "engine", "internal", "visualize", "web", "traceability-view.css"),
    target: path.join(mediaDir, "traceability-view.css"),
  },
];

fs.mkdirSync(mediaDir, { recursive: true });

for (const asset of assets) {
  if (!fs.existsSync(asset.source)) {
    throw new Error(`missing webview asset source: ${asset.source}`);
  }
  fs.copyFileSync(asset.source, asset.target);
}