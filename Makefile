.PHONY: lint format format-check check build build-engine build-extension package-extension deploy-dev

lint:
	cd engine && go run ./cmd/specs lint --style

format:
	cd engine && go run ./cmd/specs format

format-check:
	cd engine && go run ./cmd/specs format --check

check: format-check lint

build: build-engine build-extension

build-engine:
	cd engine/cmd/specs && go build -o ../../../specs

build-extension:
	cd extension && pnpm run compile

package-extension:
	cd extension && pnpm run package:bundled

deploy-dev: build-engine build-extension
	cd extension && pnpm run symlink
