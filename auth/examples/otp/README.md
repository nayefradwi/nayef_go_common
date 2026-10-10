# OTP login

Passwordless: the first verified code creates the account.

```sh
OTP_SECRET=$(openssl rand -hex 32) go run ./otp
```

```sh
# request a code; it shows up in the server log (msg=deliver)
curl -s localhost:8080/otp/request -d '{"email":"a@example.com"}'

# again within a minute → 429 (resend cap); 20 requests per IP per hour → 429
curl -s localhost:8080/otp/request -d '{"email":"a@example.com"}'

# verify → tokens; 5 wrong codes and the code is dead
ACCESS=$(curl -s localhost:8080/otp/verify -d '{"email":"a@example.com","code":"123456"}' | jq -r .access_token)

curl -s localhost:8080/me -H "Authorization: Bearer $ACCESS"
```
