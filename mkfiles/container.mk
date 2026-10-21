CONTAINER_MK := $(lastword $(MAKEFILE_LIST))
CONTAINER_MK_DIR := $(dir $(abspath $(CONTAINER_MK)))

.PHONY: container-build
container-build:
	@cd $(CONTAINER_MK_DIR)../ \
	&& docker build \
		-f Dockerfile . \
	;
