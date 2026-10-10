# API keys / m2m

Three route groups:
- user only (`authenticate`): create, list, revoke keys. A key can't mint keys.
- key only (`ApiKeyMiddleware`): `/m2m/ping`.
- either (`EitherMiddleware`, split by the `sk_live_` prefix): `/reports`, narrowed per route with `Has(IsKey)` or `Has(ByKind(userCheck, keyCheck))`.

```sh
go run ./apikeys
```

```sh
ACCESS=$(curl -s localhost:8080/signup -d '{"email":"a@example.com","password":"correct horse"}' | jq -r .access_token)

# the raw key is in this response only
KEY=$(curl -s localhost:8080/keys -H "Authorization: Bearer $ACCESS" \
  -d '{"name":"ci","scopes":["reports:read"]}' | jq -r .key)
curl -s localhost:8080/keys -H "Authorization: Bearer $ACCESS"                   # list, no secrets

curl -s localhost:8080/m2m/ping -H "Authorization: Bearer $KEY"                  # 200
curl -s localhost:8080/m2m/ping -H "Authorization: Bearer $ACCESS"               # 401, users not allowed

curl -s localhost:8080/reports -H "Authorization: Bearer $KEY"                   # 200, kind key
curl -s localhost:8080/reports -H "Authorization: Bearer $ACCESS"                # 200, kind user
curl -s -X POST localhost:8080/reports/ingest -H "Authorization: Bearer $KEY"    # 202, keys only
curl -s -X POST localhost:8080/reports/ingest -H "Authorization: Bearer $ACCESS" # 403
curl -s -X POST localhost:8080/reports -H "Authorization: Bearer $KEY"           # 403, key lacks reports:write
curl -s -X POST localhost:8080/reports -H "Authorization: Bearer $ACCESS"        # 201, users always

# revoke is owner-scoped: another user's key id is a 404
ID=$(curl -s localhost:8080/keys -H "Authorization: Bearer $ACCESS" | jq -r '.[0].id')
curl -s -X DELETE localhost:8080/keys/$ID -H "Authorization: Bearer $ACCESS"
curl -s localhost:8080/m2m/ping -H "Authorization: Bearer $KEY"                  # 401
```
