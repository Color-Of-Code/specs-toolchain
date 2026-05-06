import { createRequire } from "node:module";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// Plugins are installed under extension/node_modules; use createRequire so
// that Node resolves them from there regardless of where ESLint is invoked.
const require = createRequire(
  path.join(__dirname, "extension", "node_modules", "marker.js"),
);
const tseslint = require("@typescript-eslint/eslint-plugin");
const tsparser = require("@typescript-eslint/parser");

const webSrcDir = "engine/internal/visualize/web/src";
const extSrcDir = "extension/src";

// Rules shared by both packages.
const sharedRules = {
  // ── Type-aware recommended rules ────────────────────────────────────
  ...tseslint.configs["recommended-type-checked"].rules,

  // ── Extra strictness ────────────────────────────────────────────────
  // Unhandled promise rejections are easy to miss in both packages.
  "@typescript-eslint/no-floating-promises": "error",
  // Explicit return types make public API contracts clear.
  "@typescript-eslint/explicit-function-return-type": [
    "error",
    { allowExpressions: true, allowHigherOrderFunctions: true },
  ],
  // Force consistent use of `??` over `||` for nullish checks.
  "@typescript-eslint/prefer-nullish-coalescing": "error",
  // Disallow gratuitous non-null assertions; fix the types instead.
  "@typescript-eslint/no-non-null-assertion": "error",
  // No unnecessary type casts.
  "@typescript-eslint/no-unnecessary-type-assertion": "error",
  // Prefer for-of over indexed loops where only the value is used.
  "@typescript-eslint/prefer-for-of": "error",
  // Catch accidental await on non-Promise values.
  "@typescript-eslint/await-thenable": "error",
  // Unused variables are noise.
  "@typescript-eslint/no-unused-vars": [
    "error",
    { argsIgnorePattern: "^_", varsIgnorePattern: "^_" },
  ],
  // Prefer destructuring assignment over property access.
  "prefer-destructuring": [
    "error",
    {
      VariableDeclarator: { array: false, object: true },
      AssignmentExpression: { array: false, object: false },
    },
    { enforceForRenamedProperties: false },
  ],
};

export default [
  // ── Webview (browser) TypeScript ─────────────────────────────────────────
  // cytoscape.data() returns `any` by design; relax unsafe rules to warnings.
  {
    files: [`${webSrcDir}/**/*.ts`],
    languageOptions: {
      parser: tsparser,
      parserOptions: {
        project: path.join(__dirname, webSrcDir, "tsconfig.json"),
        tsconfigRootDir: path.join(__dirname, webSrcDir),
      },
    },
    plugins: { "@typescript-eslint": tseslint },
    rules: {
      ...sharedRules,
      "@typescript-eslint/no-unsafe-argument": "warn",
      "@typescript-eslint/no-unsafe-assignment": "warn",
      "@typescript-eslint/no-unsafe-member-access": "warn",
      "@typescript-eslint/no-unsafe-call": "warn",
    },
  },

  // ── Extension (Node / VS Code) TypeScript ────────────────────────────────
  // No Cytoscape `any` usage here — enforce the full unsafe rules as errors.
  {
    files: [`${extSrcDir}/**/*.ts`],
    languageOptions: {
      parser: tsparser,
      parserOptions: {
        project: path.join(__dirname, extSrcDir, "../tsconfig.json"),
        tsconfigRootDir: path.join(__dirname, extSrcDir, ".."),
      },
    },
    plugins: { "@typescript-eslint": tseslint },
    rules: {
      ...sharedRules,
      "@typescript-eslint/no-unsafe-argument": "error",
      "@typescript-eslint/no-unsafe-assignment": "error",
      "@typescript-eslint/no-unsafe-member-access": "error",
      "@typescript-eslint/no-unsafe-call": "error",
    },
  },
];
