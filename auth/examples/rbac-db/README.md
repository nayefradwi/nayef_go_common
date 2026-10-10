# RBAC with a DB

The `can(permission)` check runs one query per request against `role_permissions` (created and seeded at startup). A role change applies on the next request, no new login needed. Routes are layered with nested `r.Group`: authenticated → `reports:read` → `reports:write`.

```sh
go run ./rbac-db
```

```sh
ACCESS=$(curl -s localhost:8080/signup -d '{"email":"a@example.com","password":"correct horse"}' | jq -r .access_token)

curl -s localhost:8080/reports -H "Authorization: Bearer $ACCESS"               # 200, member has reports:read
curl -s -X POST localhost:8080/reports -H "Authorization: Bearer $ACCESS"       # 403

psql postgres://postgres:postgres@localhost:5432/auth_examples -c "UPDATE users SET role = 'admin' WHERE email = 'a@example.com'"
curl -s -X POST localhost:8080/reports -H "Authorization: Bearer $ACCESS"       # 201, same token
```
