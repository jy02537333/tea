# Minimal Integration Checklist

This document tracks minimal, non-blocking integration steps and quick permission checks for admin routes.

## Admin Order Permissions (SC6a)

- Permissions required:
  - `order:accept`
  - `order:reject`
- RBAC seeding:
  - Both permissions are seeded via `tea-api/internal/service/seed.go` and granted to the `admin` role (idempotent `FirstOrCreate`).

### Quick cURL Checks

Refer to archived examples: build-ci-logs/sc6a_route_permission_curl.md

- Accept (requires `order:accept`)
  - Endpoint: `POST /api/v1/admin/orders/:id/accept`
  - Expect: 200 with `{ ok: true, status: "accepted" }`
- Reject (requires `order:reject`)
  - Endpoint: `POST /api/v1/admin/orders/:id/reject`
  - Expect: 200 with `{ ok: true, status: "rejected" }`
- Missing permission
  - Expect: `403` with `missing_permission`

### Evidence Links

- Focused tests output: build-ci-logs/sc6a_tests_focus_output.txt
- Route permission cURL examples: build-ci-logs/sc6a_route_permission_curl.md

### Notes

- `AdminAcceptOrder` only allows transition when `status=2` (paid) → becomes `status=3`.
- `AdminRejectOrder` only allows when `status=3` → delegates to refund path setting `status=5` (cancelled) and `pay_status=4` (refunded).
- Commission rollback not auto-triggered on `AdminReject`; follow-up PR will align with `AdminRefundConfirm`, or finance handles via reconcile tooling.
