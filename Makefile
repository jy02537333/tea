SHELL := /bin/bash
MAKEFLAGS += --warn-undefined-variables

# Optional extra args, e.g. make package PACKAGE_ARGS="--os linux --arch amd64"
PACKAGE_ARGS ?=

.PHONY: up package

up:
	@echo "[make up] starting tea-api via run-tea-api.sh"
	bash run-tea-api.sh

package:
	@echo "[make package] bundling tea-api with args: $(PACKAGE_ARGS)"
	bash tea-api/scripts/package-tea-api.sh $(PACKAGE_ARGS)
