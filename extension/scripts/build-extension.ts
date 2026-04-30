#!/usr/bin/env node

const { execFileSync } = require("node:child_process");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");

type Target = "linux-x64" | "linux-arm64" | "darwin-x64" | "darwin-arm64" | "win32-x64";

type TargetConfig = {
	archiveFileName: (version: string) => string;
	archiveBinary: string;
	bundledBinary: string;
};

const TARGETS: Record<Target, TargetConfig> = {
	"linux-x64": {
		archiveFileName: (version) => `specs_${version}_linux_amd64.tar.gz`,
		archiveBinary: "specs",
		bundledBinary: "specs",
	},
	"linux-arm64": {
		archiveFileName: (version) => `specs_${version}_linux_arm64.tar.gz`,
		archiveBinary: "specs",
		bundledBinary: "specs",
	},
	"darwin-x64": {
		archiveFileName: (version) => `specs_${version}_darwin_amd64.tar.gz`,
		archiveBinary: "specs",
		bundledBinary: "specs",
	},
	"darwin-arm64": {
		archiveFileName: (version) => `specs_${version}_darwin_arm64.tar.gz`,
		archiveBinary: "specs",
		bundledBinary: "specs",
	},
	"win32-x64": {
		archiveFileName: (version) => `specs_${version}_windows_amd64.zip`,
		archiveBinary: "specs.exe",
		bundledBinary: "specs.exe",
	},
};

function usage(): void {
	console.error("usage: pnpm run package:bundled -- [<target>] [--version <version>]");
	console.error("targets: linux-x64 linux-arm64 darwin-x64 darwin-arm64 win32-x64");
	console.error("defaults to the host platform/architecture if <target> is omitted");
	console.error("version defaults to SPECS_VERSION, then extension/package.json version");
}

function hostTarget(): Target {
	const platform = os.platform();
	const arch = os.arch();
	if (platform === "linux" && arch === "x64") return "linux-x64";
	if (platform === "linux" && arch === "arm64") return "linux-arm64";
	if (platform === "darwin" && arch === "x64") return "darwin-x64";
	if (platform === "darwin" && arch === "arm64") return "darwin-arm64";
	if (platform === "win32" && arch === "x64") return "win32-x64";
	throw new Error(`unsupported host platform/arch: ${platform}/${arch}`);
}

function parseArgs(argv: string[]): { target: Target; versionOverride?: string } {
	let target: Target | undefined;
	let versionOverride: string | undefined;

	for (let index = 0; index < argv.length; index += 1) {
		const arg = argv[index];
		if (arg === "--") {
			continue;
		}
		if (arg === "-h" || arg === "--help") {
			usage();
			process.exit(0);
		}
		if (arg === "--version") {
			versionOverride = argv[index + 1];
			index += 1;
			continue;
		}
		if (arg === "--target") {
			target = argv[index + 1] as Target | undefined;
			index += 1;
			continue;
		}
		if (!arg.startsWith("-") && target === undefined) {
			target = arg as Target;
			continue;
		}

		usage();
		throw new Error(`unexpected argument: ${arg}`);
	}

	if (!target) {
		target = hostTarget();
	}

	if (!(target in TARGETS)) {
		usage();
		throw new Error(`unsupported target: ${target}`);
	}

	return { target, versionOverride };
}

function run(command: string, args: string[], cwd: string): void {
	execFileSync(command, args, {
		cwd,
		stdio: "inherit",
		env: process.env,
	});
}

function findFirstFile(rootDir: string, fileName: string): string | undefined {
	const entries = fs.readdirSync(rootDir, { withFileTypes: true });
	for (const entry of entries) {
		const entryPath = path.join(rootDir, entry.name);
		if (entry.isDirectory()) {
			const nested = findFirstFile(entryPath, fileName);
			if (nested) {
				return nested;
			}
			continue;
		}
		if (entry.isFile() && entry.name === fileName) {
			return entryPath;
		}
	}
	return undefined;
}

function restoreDirectory(sourceDir: string, destinationDir: string): void {
	fs.rmSync(destinationDir, { recursive: true, force: true });
	if (fs.existsSync(sourceDir)) {
		fs.cpSync(sourceDir, destinationDir, { recursive: true });
	}
}

function main(): void {
	const { target, versionOverride } = parseArgs(process.argv.slice(2));
	const extensionDir = path.resolve(__dirname, "..");
	const repoRoot = path.resolve(extensionDir, "..");
	const distDir = path.join(repoRoot, "dist");
	const binDir = path.join(extensionDir, "bin");
	const extensionPackageJsonPath = path.join(extensionDir, "package.json");
	const extensionPackageJson = JSON.parse(fs.readFileSync(extensionPackageJsonPath, "utf8")) as {
		version?: string;
	};
	const version = versionOverride ?? process.env.SPECS_VERSION ?? extensionPackageJson.version;

	if (!version) {
		throw new Error("missing version: pass --version or set SPECS_VERSION");
	}

	const targetConfig = TARGETS[target];
	const archivePath = path.join(distDir, targetConfig.archiveFileName(version));
	if (!fs.existsSync(archivePath)) {
		throw new Error(`expected GoReleaser artifact not found: ${archivePath}`);
	}

	const extractDir = fs.mkdtempSync(path.join(os.tmpdir(), "specs-vsix-extract-"));
	const backupDir = fs.mkdtempSync(path.join(os.tmpdir(), "specs-vsix-backup-"));
	const binBackupDir = path.join(backupDir, "bin");
	const bundledBinaryPath = path.join(binDir, targetConfig.bundledBinary);

	try {
		if (archivePath.endsWith(".zip")) {
			run("unzip", ["-q", archivePath, "-d", extractDir], repoRoot);
		} else {
			run("tar", ["-xzf", archivePath, "-C", extractDir], repoRoot);
		}

		const binarySource = findFirstFile(extractDir, targetConfig.archiveBinary);
		if (!binarySource) {
			throw new Error(`could not find ${targetConfig.archiveBinary} inside ${archivePath}`);
		}

		if (fs.existsSync(binDir)) {
			fs.cpSync(binDir, binBackupDir, { recursive: true });
		}
		fs.rmSync(binDir, { recursive: true, force: true });
		fs.mkdirSync(binDir, { recursive: true });

		fs.copyFileSync(binarySource, bundledBinaryPath);
		if (!bundledBinaryPath.endsWith(".exe")) {
			fs.chmodSync(bundledBinaryPath, 0o755);
		}

		run("pnpm", ["run", "compile"], extensionDir);
		run(
			"pnpm",
			[
				"exec",
				"vsce",
				"package",
				version,
				"--no-git-tag-version",
				"--no-update-package-json",
				"--out",
				path.join(distDir, `specs-${target}.vsix`),
				"--target",
				target,
				"--no-dependencies",
			],
			extensionDir,
		);

		console.log(`built: dist/specs-${target}.vsix`);
	} finally {
		restoreDirectory(binBackupDir, binDir);
		fs.rmSync(extractDir, { recursive: true, force: true });
		fs.rmSync(backupDir, { recursive: true, force: true });
	}
}

main();