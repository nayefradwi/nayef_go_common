# nayef_go_common

This is a commons monorepo that consists of multiple modules that I typically use when starting a new project. To reduce
the amount of boilerplate I have to re-write with every project, I have extracted packages from other side projects,
iterated over them multiple times, and tested them in other side projects. This is a project that I typically update
from time to time based on things I realize as I use these packages.

It mostly targets my preferred stack in go which is:
- routing: chi
- database: sqlc + pgx

DISCLAIMER: Claude was used in parts of the code as well as the test cases while majority of the planning,
structuring, and code was implemented by me.

## Objectives

- Reduce boilerplate and have consistent code across different fun projects I build
- Reduce the time it takes to have something up and running when bootstrapping new projects
- Having something central that keeps getting updated and tested instead of scattering ideas across different projects

## Modules

Each module targets a specific area of boilerplate that I am trying to avoid re-writing. Overview of the modules:
- errors: A custom error that includes status, code, message and field validations for better errors
- errorspb: A mapping layer for the core errors to protobuf for microservices / grpc based APIs
- httputil: utility methods for parsing json objects and writing responses
- grpcutil: similar to httputil but focuses on grpc and protobufs
- pgutil: utility methods for postgres (right now only has connection)
- redisutil: similar to pgutil but for redis
- pagination: helper methods and models for generic pagination, currently only supports limit and offset, cursor based is pending
- validation: no reflection based validators following a style similar to fluentvalidation in C#
- auth: Implementation of jwt, hashing (wrapper on bcrypt), and introducing providers to implement jwt, refresh, and opaque tokens
- otp: OTP helper methods along with a redis based implementation for generating codes
- locking: A locking interface along with an abstraction on redsync to reduce boilerplate code

### errors & errorspb

errors and errorspb modules allow me to not create a custom error for every project I build as well as allows consistency
across different modules to have the same error structure. This structure gives the frontend control on how to deal
with errors coming from the backend as well as a clear message for tracking purposes.

```json
{
    "message": "something went wrong",
    "code": "FAILED_TO_CONNECT"
}
```
The module also has a factory that makes creating domain errors easy, an example from the otp module:

```go
var (
	ErrIncorrectOTP     = errors.NewResultError("incorrect otp", IncorrectOtpErrorCode)
	ErrMaxTriesExceeded = errors.NewResultError("max tries exceeded", MaxTriesExceededCode)
	ErrExpiredOTP       = errors.NewResultError("expired otp", ExpiredOtpCode)
	ErrOtpNotFound      = errors.NewResultError("otp not found, request a new one", OtpNotFoundCode)
)
```

or using pre-existing codes like unauthorized:

```go
err := errors.UnauthorizedError("Invalid jwt token")
```

### httputil & grpcutil

This module is used to avoid handling JSON responses always by either introducing a middleware like this:

```go
func JSONMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
```

or avoiding having to write a wrapper for this logic always:

```go

    w.Header().Set("Content-Type", "application/json")
    result, err := somethingThatCouldFail()
    if err {
        w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": err.Error()})
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(result)

```

instead:

```go
    jw := httputil.NewJsonResponseWriter(w)
    result, err := somethingThatCouldFail()
    jw.WriteJsonResponse(result, err) // defaults to 200
```

or with different custom codes:

```go
    jw := httputil.NewJsonResponseWriter(w).WithErrorStatus(500).WithSuccessStatus(201)
    result, err := somethingThatCouldFail()
    jw.WriteJsonResponse(result, err)
```

this is an example using grpc:
```go
	gWriter := grpc.NewGrpcResponseWriter[protov1.ExampleDTO]()
    result, err := somethingThatCouldFail()
    return gWriter.WriteResponse(result, err) // this will do mapping to errorspb in case there is an error
```

**IMPORTANT: In error flows / cases the writer will always write the same structure as the core errors package**

### pgutil and redisutil

These only provide simpler connection logic to postgres and redis by:
1. setting up a client connection (pool in case of postgres)
2. test the connection (ping in the case of redis)
3. panics on error given that these connections are typically a must for API operations

it basically helps avoiding to rewrite these couple of lines into something like:

```go
func bootstrap(ctx context.Context) {
    pool, redis := pgutil.ConnectToPostgres(ctx, url), redisutil.ConnectToRedis(ctx, url)
    // continue to use them either for dependency set up or something else
}
```

### pagination

This module avoids having to re-implement pagination parsing, validation, and response building for every endpoint that needs it. It provides an HTTP middleware that extracts pagination parameters from the query string and stores them in context, as well as generic response models with pre-calculated metadata:

```go
// attach the middleware to your router
r.Use(pagination.SetOffsetPaginationMiddleware)

// in your handler, retrieve the query from context
func listHandler(w http.ResponseWriter, r *http.Request) {
    query := pagination.OffsetPageQueryFromContext(r.Context()) // defaults to page 1, size 10
    items, total, err := repo.List(ctx, query.Offset(), query.PageSize)
    jw := httputil.NewJsonResponseWriter(w)
    jw.WriteJsonResponse(pagination.NewOffsetPage(query.Page, query.PageSize, total, items), err)
}
```

**IMPORTANT: page size is capped at 100 and page defaults to 1 if not provided or invalid**

### validation

This module avoids re-writing validation logic from scratch for every project by providing a fluent, no-reflection rule builder similar to FluentValidation in C#. Rules are collected and executed together so all failures are returned at once:

```go
v := validation.NewValidator()
stringFactory := validation.StringValidationRuleFactory{}
numFactory := validation.NumValidationRuleFactory[int]{}

validation.AddRule(v, stringFactory.IsRequired(req.Email, "email"))
validation.AddRule(v, stringFactory.IsEmail(req.Email, "email"))
validation.AddRule(v, numFactory.MinValue(req.Age, "age", 18))

return validation.Validate() // which returns an error
```

or using a custom rule when the built-in ones are not enough:

```go
validation.AddRule(v, stringFactory.Must(req.Username, "username", "username is already taken", func() bool {
    return !repo.UsernameExists(req.Username)
}))
```

### auth

This module avoids re-implementing JWT signing, password hashing, and authentication middleware for every project. It provides a layered set of token providers depending on how much control over token revocation is needed:

```go
// stateless jwt (no revocation)
accessConfig := auth.NewJwtTokenProviderConfig(secret, 15*time.Minute)
provider := auth.NewJwtTokenProvider(accessConfig)
token, err := provider.SignClaims(userId, map[string]any{"role": "admin"})

// access + refresh token pair
refreshConfig := auth.NewJwtTokenProviderConfig(secret, 7*24*time.Hour)
refreshProvider := auth.NewJwtRefreshTokenProvider(
    auth.NewJwtTokenProvider(refreshConfig),
    auth.NewJwtTokenProvider(accessConfig),
)
dto, err := refreshProvider.GenerateToken(userId, claims) // dto.AccessToken, dto.RefreshToken
```

or using reference tokens where all tokens are stored in a database for full revocation support:

```go
refProvider := auth.NewJwtReferenceTokenProvider(refreshProvider, tokenStore)
dto, err := refProvider.GenerateToken(userId, claims) // both tokens are IDs, not raw JWTs
refProvider.RevokeOwner(userId) // invalidate all sessions
```

password hashing is also provided as a thin wrapper over bcrypt:

```go
hc := auth.NewHashingConfig(10)
hash, err := hc.Hash(password)
ok := auth.CompareHash(password, hash)
```

protecting routes is done through the provided middleware:

```go
r.Use(auth.NewJwtAuthenticationMiddleware(provider).UseAuthentication)

// retrieve the token in a handler
token := auth.GetToken(r.Context())
```

### otp

This module avoids re-implementing OTP generation, hashing, retry limiting, and expiry logic for every project. It provides a service backed by Redis that handles the full OTP lifecycle:

```go
config := otp.OtpConfig{
    ExpiresIn:   5 * time.Minute,
    MaxTries:    3,
    ResendAfter: 1 * time.Minute,
}
generator := otp.NewCodeGenerator(6, false) // 6-digit numeric code
repo := otp.NewRedisOtpRepository(redisClient)
service := otp.NewOtpService(repo, generator, config)

// generate and send to user
generatedOtp, err := service.GenerateOtp(ctx, userId) // returns existing OTP if resend period has not passed
sendSms(generatedOtp.Code)

// verify when the user submits
err = service.VerifyOtp(ctx, userId, submittedCode)
```

**IMPORTANT: codes are stored hashed and verification uses constant-time comparison to prevent timing attacks**

### locking

This module avoids re-implementing distributed locking boilerplate on top of Redsync for every project. It provides a simple interface for acquiring and releasing locks with configurable retry and TTL behaviour:

```go
locker := locking.NewDistributedLockerFromClient(redisClient)
params := locking.DefaultLockParams // TTL: 2s, wait: 100ms, retries: 10

// acquire, run, auto-release
err := locker.RunWithLock(ctx, "resource:"+resourceId, params, func() error {
    // critical section
    return doSomething()
})
```

or when multiple resources need to be locked together:

```go
err := locker.RunWithLocks(ctx, []string{"account:"+a, "account:"+b}, params, func() error {
    return transfer(a, b, amount)
})
```

**IMPORTANT: if acquiring any lock in a multi-lock call fails, all already-acquired locks are automatically released**
