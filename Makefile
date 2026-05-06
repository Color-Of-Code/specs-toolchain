.PHONY: lint lint-ts format format-docs format-check check test vet build build-webview build-engine build-extension package-extension deploy-dev

lint:
	cd engine && go run ./cmd/specs lint --style

lint-ts:
	cd extension && pnpm run lint
	node extension/node_modules/eslint/bin/eslint.js --max-warnings 0 --config eslint.webview.config.mjs extension/src engine/internal/visualize/web/src

format:
	cd engine && go run ./cmd/specs format

format-docs:
	cd engine && go run ./cmd/specs format --at ../docs

format-check:
	cd engine && go run ./cmd/specs format --check

vet:
	cd engine && go vet ./...

test:
	cd engine && go test ./...

check: format-check lint lint-ts vet test

build: build-engine build-extension

build-webview:
	cd extension && pnpm run build-webview

build-engine: build-webview
	mkdir -p bin
	cd engine/cmd/specs && go build -o ../../../bin/specs

build-extension:
	cd extension && pnpm run compile

package-extension:
	cd extension && pnpm run package:bundled

deploy-dev: build-engine build-extension
	cd extension && pnpm run symlink
	ln -sf "$(CURDIR)/bin/specs" extension/bin/specs
