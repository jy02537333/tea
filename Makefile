SHELL := /bin/bash
MAKEFLAGS += --warn-undefined-variables

# Optional extra args, e.g. make package PACKAGE_ARGS="--os linux --arch amd64"
PACKAGE_ARGS ?=

.PHONY: up package test test-api test-admin-fe test-wx-fe verify-sprint-b verify-sprint-b-strict

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

verify-sprint-b:
	@echo "[make verify-sprint-b] Sprint B membership flow verification (normal mode)"
	@echo "Note: Requires tea-api running and USER_TOKEN/ADMIN_TOKEN exported"
	@bash scripts/run_membership_integration.sh
	@bash scripts/assert_membership_flow.sh
	@echo "[make verify-sprint-b] Sprint B verification completed successfully"

verify-sprint-b-strict:
	@echo "[make verify-sprint-b-strict] Sprint B membership flow verification (strict mode)"
	@echo "Note: Requires tea-api running and USER_TOKEN/ADMIN_TOKEN exported"
	@bash scripts/run_membership_integration.sh
	@bash scripts/assert_membership_flow.sh --strict
	@echo "[make verify-sprint-b-strict] Sprint B strict verification completed successfully"
