# RBAC with claims

Role and scopes are written into the session at login; routes only read claims, no DB. The session keeps those claims through every refresh, so a role change shows up at the next login (call `RevokeOwner` to force one).

```sh
go run ./rbac-claims
```

```sh
TOKENS=$(curl -s localhost:8080/signup -d '{"email":"a@example.com","password":"correct horse"}')
ACCESS=$(echo $TOKENS | jq -r .access_token)

curl -s localhost:8080/me -H "Authorization: Bearer $ACCESS"                    # role member
curl -s localhost:8080/reports -H "Authorization: Bearer $ACCESS"               # 200, scope reports:read
curl -s -X POST localhost:8080/reports -H "Authorization: Bearer $ACCESS"       # 403, no reports:write
curl -s localhost:8080/admin -H "Authorization: Bearer $ACCESS"                 # 403, role is member

psql postgres://postgres:postgres@localhost:5432/auth_examples -c "UPDATE users SET role = 'admin' WHERE email = 'a@example.com'"

# still 403 after a refresh: the claims came from login
ACCESS=$(curl -s localhost:8080/refresh -d "{\"refresh_token\":\"$(echo $TOKENS | jq -r .refresh_token)\"}" | jq -r .access_token)
curl -s localhost:8080/admin -H "Authorization: Bearer $ACCESS"

ACCESS=$(curl -s localhost:8080/login -d '{"email":"a@example.com","password":"correct horse"}' | jq -r .access_token)
curl -s localhost:8080/admin -H "Authorization: Bearer $ACCESS"                 # 200
```
