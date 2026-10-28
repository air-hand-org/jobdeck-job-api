all: setup build

include ./mkfiles/*.mk

.PHONY: setup
setup:
	@aqua i && mise i

.PHONY: generate
generate:
	@sqlc generate

.PHONY: build
build: go-build licenses

# Match the platform currently published by _build-images.yaml.
.PHONY: licenses
licenses:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go-licenses check ./cmd/server ./cmd/migrator ./cmd/entrypoint --ignore github.com/air-hand-org/jobdeck-job-api --disallowed_types=forbidden,unknown
	rm -rf third_party/licenses/linux-amd64
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go-licenses save ./cmd/server ./cmd/migrator ./cmd/entrypoint --ignore github.com/air-hand-org/jobdeck-job-api --save_path third_party/licenses/linux-amd64

.PHONY: fmt
fmt: go-fmt

.PHONY: lint
lint: go-lint

.PHONY: check-clean
check-clean:
	@git status --short
	@test -z "$$(git status --porcelain)"

.PHONY: schema-docs
schema-docs:
	@bash scripts/schema-docs.sh

.PHONY: schema-erd
schema-erd:
	npm exec --yes --package=@liam-hq/cli@0.7.24 -- liam erd build --format tbls --input docs/schema/schema.json --output-dir tmp/schema-erd

.PHONY: test
test: go-test
