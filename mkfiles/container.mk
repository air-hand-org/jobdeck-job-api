CONTAINER_MK := $(lastword $(MAKEFILE_LIST))
CONTAINER_MK_DIR := $(dir $(abspath $(CONTAINER_MK)))

.PHONY: container-build
container-build:
	@cd $(CONTAINER_MK_DIR)../ \
	&& GITHUB_CONTENTS_TOKEN="$$(gh auth token)" docker build \
		--secret id=github_contents_token,env=GITHUB_CONTENTS_TOKEN \
		-f Dockerfile . \
	;
