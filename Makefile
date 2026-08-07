.PHONY: build run test vet docs docs-serve-prod docs-clean

build:
	go build ./...

run:
	go run ./cmd/immichframe-opencloud

test:
	go test ./...

vet:
	go vet ./...

# --- Docs ---
# Service reference site (env vars, example config, deprecations).
# Runs dschmidt/opencloud-service-docs-action at the SHA pinned in
# .github/workflows/docs.yml — identical code path to CI.

docs:
	DOCS_OUTPUT="$(CURDIR)/docs/generated" bash .github/docs/run.sh

docs-serve-prod:
	cd .github/docs/.cache/site && pnpm run serve

docs-clean:
	rm -rf .github/docs/.cache docs/generated
