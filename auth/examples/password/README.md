# Email + password

```sh
go run ./password
```

```sh
# signup → tokens
curl -s localhost:8080/signup -d '{"email":"a@example.com","password":"correct horse"}'

# login: 5 tries per email per 15 min, then 429
TOKENS=$(curl -s localhost:8080/login -d '{"email":"a@example.com","password":"correct horse"}')
ACCESS=$(echo $TOKENS | jq -r .access_token); REFRESH=$(echo $TOKENS | jq -r .refresh_token)

curl -s localhost:8080/me -H "Authorization: Bearer $ACCESS"

# refresh rotates: the old refresh token is spent
NEW=$(curl -s localhost:8080/refresh -d "{\"refresh_token\":\"$REFRESH\"}")

# reusing a spent refresh token ends the whole login (both 401)
curl -s localhost:8080/refresh -d "{\"refresh_token\":\"$REFRESH\"}"
curl -s localhost:8080/refresh -d "{\"refresh_token\":\"$(echo $NEW | jq -r .refresh_token)\"}"

# logout ends one login, logout-all ends every login of this user
TOKENS=$(curl -s localhost:8080/login -d '{"email":"a@example.com","password":"correct horse"}')
curl -s localhost:8080/logout -d "{\"refresh_token\":\"$(echo $TOKENS | jq -r .refresh_token)\"}"
TOKENS=$(curl -s localhost:8080/login -d '{"email":"a@example.com","password":"correct horse"}')
curl -s -X POST localhost:8080/logout-all -H "Authorization: Bearer $(echo $TOKENS | jq -r .access_token)"
```
