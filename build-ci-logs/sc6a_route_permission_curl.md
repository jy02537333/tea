# SC6a Route Permission cURL Examples

These examples verify granular permissions for admin order routes.

## Accept (requires permission: order:accept)

```bash
BASE="http://localhost:9393"
ORDER_ID="123"
ADMIN_TOKEN="<token-with-order:accept>"

curl -i -X POST "$BASE/api/v1/admin/orders/$ORDER_ID/accept" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json"
# expect: 200 with {"ok": true, "status": "accepted"}
```

## Reject (requires permission: order:reject)

```bash
BASE="http://localhost:9393"
ORDER_ID="123"
ADMIN_TOKEN="<token-with-order:reject>"

curl -i -X POST "$BASE/api/v1/admin/orders/$ORDER_ID/reject" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"reason":"not_available"}'
# expect: 200 with {"ok": true, "status": "rejected"}
```

## Missing permission example (should be 403)

```bash
BASE="http://localhost:9393"
ORDER_ID="123"
NO_PERM_TOKEN="<token-without-required-permission>"

curl -i -X POST "$BASE/api/v1/admin/orders/$ORDER_ID/accept" \
  -H "Authorization: Bearer $NO_PERM_TOKEN"
# expect: 403 with error: missing_permission
```
