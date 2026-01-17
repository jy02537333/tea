SHELL := /bin/bash
MAKEFLAGS += --warn-undefined-variables

## monitor-pr52-ci: Monitor latest API Validation run for PR #52 and auto-comment result
.PHONY: monitor-pr52-ci
monitor-pr52-ci:
	@GH_TOKEN=$${GH_TOKEN:-$$(cat .github_token 2>/dev/null || true)} \
	OWNER=jy02537333 REPO=tea PR_NUMBER=52 BRANCH=chore/ci-disable-api-validation-on-master WORKFLOW_FILE=api-validation.yml \
	POLL_INTERVAL=30 \
	bash scripts/ci/monitor_workflow_and_comment.sh

# Optional extra args, e.g. make package PACKAGE_ARGS="--os linux --arch amd64"
PACKAGE_ARGS ?=

.PHONY: up package test test-api test-admin-fe test-wx-fe verify-sprint-a verify-sprint-a-strict verify-sprint-b verify-sprint-b-strict verify-sprint-a-e2e prepare-tokens run-min-integration verify-sprint-c

.PHONY: auto-test
auto-test:
	@echo "[make auto-test] running one-click flow tests (backend + e2e scripts)"
	bash scripts/auto_test.sh

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

prepare-tokens:
	@echo "[make prepare-tokens] preparing ADMIN_TOKEN and USER_TOKEN"
	mkdir -p build-ci-logs
	API_BASE=$${API_BASE:-http://127.0.0.1:9292} bash ./scripts/prepare_tokens.sh

run-min-integration:
	@echo "[make run-min-integration] running minimal integration to generate Sprint C evidence"
	mkdir -p build-ci-logs
	API_BASE=$${API_BASE:-http://127.0.0.1:9292} bash ./scripts/run_min_integration.sh

verify-sprint-c:
	@echo "[make verify-sprint-c] verifying Sprint C evidence under build-ci-logs (non-blocking in CI)"
	bash ./scripts/verify_sprint_c.sh
