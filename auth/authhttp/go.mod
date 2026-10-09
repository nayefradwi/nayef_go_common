module github.com/nayefradwi/nayef_go_common/auth/authhttp

go 1.27.0

replace github.com/nayefradwi/nayef_go_common/auth => ../

require (
	github.com/google/uuid v1.6.0
	github.com/nayefradwi/nayef_go_common/auth v0.0.0-00010101000000-000000000000
	github.com/nayefradwi/nayef_go_common/errors v1.0.8
	github.com/nayefradwi/nayef_go_common/httputil v1.0.7
)

require (
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	golang.org/x/crypto v0.49.0 // indirect
)
