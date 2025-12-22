SHELL := /bin/bash
MAKEFLAGS += --warn-undefined-variables

# Optional extra args, e.g. make package PACKAGE_ARGS="--os linux --arch amd64"
PACKAGE_ARGS ?=

.PHONY: up package test test-api test-admin-fe test-wx-fe verify-sprint-a verify-sprint-a-strict verify-sprint-b verify-sprint-b-strict verify-sprint-a-e2e

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

verify-sprint-a:
	@echo "[make verify-sprint-a] asserting admin endpoints and order amount deduction"
	bash scripts/assert_api_validation.sh

verify-sprint-a-strict:
	@echo "[make verify-sprint-a-strict] asserting with REQUIRE_ORDER_CHECK=1"
	REQUIRE_ORDER_CHECK=1 bash scripts/assert_api_validation.sh

verify-sprint-b:
	@echo "[make verify-sprint-b] asserting Sprint B membership flow"
	bash scripts/assert_membership_flow.sh

verify-sprint-b-strict:
	@echo "[make verify-sprint-b-strict] asserting Sprint B membership flow with REQUIRE_MEMBERSHIP_CHECK=1"
	REQUIRE_MEMBERSHIP_CHECK=1 bash scripts/assert_membership_flow.sh

# One-click: run stateful check to generate evidence, then strict assertions
verify-sprint-a-e2e:
	@echo "[make verify-sprint-a-e2e] running stateful checks + strict Sprint A assertions"
	bash scripts/local_api_check.sh
	REQUIRE_ORDER_CHECK=1 bash scripts/assert_api_validation.sh
