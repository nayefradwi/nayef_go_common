# Reset password

```sh
go run ./reset
```

```sh
curl -s localhost:8080/signup -d '{"email":"a@example.com","password":"correct horse"}'
REFRESH=$(curl -s localhost:8080/login -d '{"email":"a@example.com","password":"correct horse"}' | jq -r .refresh_token)

# same 202 for unknown emails, so this can't be used to find accounts
curl -s localhost:8080/password/reset -d '{"email":"nobody@example.com"}'
curl -s localhost:8080/password/reset -d '{"email":"a@example.com"}'

# the link is in the server log: ...?email=a%40example.com&token=TOKEN
curl -s localhost:8080/password/reset/confirm -d '{"email":"a@example.com","token":"TOKEN","password":"new password"}'

# every earlier session is gone, the old password too
curl -s localhost:8080/refresh -d "{\"refresh_token\":\"$REFRESH\"}"                     # 401
curl -s localhost:8080/login -d '{"email":"a@example.com","password":"correct horse"}'   # 401
curl -s localhost:8080/login -d '{"email":"a@example.com","password":"new password"}'    # 200
```
