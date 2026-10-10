# Account activation

```sh
go run ./activation
```

```sh
# signup sends an activation link (server log) and returns no tokens
curl -s localhost:8080/signup -d '{"email":"a@example.com","password":"correct horse"}'

# 403 until activated (401 on a wrong password, so only the owner learns this)
curl -s localhost:8080/login -d '{"email":"a@example.com","password":"correct horse"}'

# from the link: ...?email=a%40example.com&token=TOKEN
curl -s localhost:8080/activate -d '{"email":"a@example.com","token":"TOKEN"}'

curl -s localhost:8080/login -d '{"email":"a@example.com","password":"correct horse"}'   # tokens
```

No resend endpoint: add one that calls `activations.Issue` again for an inactive account.
