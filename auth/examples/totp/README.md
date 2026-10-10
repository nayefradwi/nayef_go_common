# TOTP

Login with TOTP on returns a pending token instead of a session. The challenge runs `pending.Check` (counts an attempt) → `totp.Verify` → `pending.Consume` → session, so 5 wrong codes kill the pending token and the password has to be entered again. Each TOTP code works once.

```sh
TOTP_KEY=$(openssl rand -hex 32) go run ./totp
```

Codes come from an authenticator app (scan `uri`) or `oathtool --totp -b SECRET`.

```sh
ACCESS=$(curl -s localhost:8080/signup -d '{"email":"a@example.com","password":"correct horse"}' | jq -r .access_token)

# enroll → secret + otpauth:// uri; confirm → 10 recovery codes, shown once
curl -s -X POST localhost:8080/mfa/enroll -H "Authorization: Bearer $ACCESS"
curl -s localhost:8080/mfa/confirm -H "Authorization: Bearer $ACCESS" -d "{\"code\":\"$(oathtool --totp -b SECRET)\"}"

# login now asks for a second step
MFA=$(curl -s localhost:8080/login -d '{"email":"a@example.com","password":"correct horse"}' | jq -r .mfa_token)
curl -s localhost:8080/mfa/challenge -d "{\"email\":\"a@example.com\",\"mfa_token\":\"$MFA\",\"code\":\"$(oathtool --totp -b SECRET)\"}"

# or spend a recovery code instead
MFA=$(curl -s localhost:8080/login -d '{"email":"a@example.com","password":"correct horse"}' | jq -r .mfa_token)
curl -s localhost:8080/mfa/recover -d "{\"email\":\"a@example.com\",\"mfa_token\":\"$MFA\",\"recovery_code\":\"abcde-fghij\"}"

# disable needs a fresh code (wait for the next 30s step if you just used one)
curl -s localhost:8080/mfa/disable -H "Authorization: Bearer $ACCESS" -d "{\"code\":\"$(oathtool --totp -b SECRET)\"}"
```
