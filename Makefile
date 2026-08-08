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
docs:
	bash dev/build-docs.sh

docs-serve-prod:
	cd .cache/service-docs/site && pnpm run serve

docs-clean:
	rm -rf .cache/service-docs
