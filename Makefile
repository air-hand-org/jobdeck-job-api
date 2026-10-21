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
	@for module in github.com/air-hand-org/api-contracts/packages/jobdeck/auth github.com/air-hand-org/api-contracts/packages/jobdeck/job; do \
		cp -R "$$(go list -m -f '{{.Dir}}' $$module)/third_party" "third_party/licenses/linux-amd64/$$module/third_party" || exit 1; \
		chmod -R u+w "third_party/licenses/linux-amd64/$$module/third_party" || exit 1; \
	done

.PHONY: fmt
fmt: go-fmt

.PHONY: lint
lint: go-lint

.PHONY: check-clean
check-clean:
	@git status --short
	@test -z "$$(git status --porcelain)"

.PHONY: test
test: go-test
