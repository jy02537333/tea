SHELL := /bin/bash
MAKEFLAGS += --warn-undefined-variables

# Optional extra args, e.g. make package PACKAGE_ARGS="--os linux --arch amd64"
PACKAGE_ARGS ?=

.PHONY: up package test test-api test-admin-fe test-wx-fe

up:
	@echo "[make up] starting tea-api via run-tea-api.sh"
	bash run-tea-api.sh

package:
	@echo "[make package] bundling tea-api with args: $(PACKAGE_ARGS)"
	bash tea-api/scripts/package-tea-api.sh $(PACKAGE_ARGS)

test-api:
	@echo "[make test-api] running Go tests for tea-api (may fail on legacy scripts package)"
	cd tea-api && go test ./...

test-admin-fe:
	@echo "[make test-admin-fe] running admin-fe typecheck tests"
	cd admin-fe && pnpm test

test-wx-fe:
	@echo "[make test-wx-fe] running wx-fe typecheck tests"
	cd wx-fe && pnpm test

test: test-api test-admin-fe test-wx-fe
	@echo "[make test] all tests finished"
