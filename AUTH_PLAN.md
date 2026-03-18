# Auth Package Migration Plan

## What it provides

### Core JWT layer (~260 lines across `modules/auth`)
- `JwtTokenProvider` — sign and parse JWTs with pluggable algorithms (HMAC, RSA, ECDSA)
- `JwtTokenProviderConfig` — builder-pattern config for signing method, keys, expiry, issuer
- Signer/parser closures for each algorithm family
- `Token` model with context storage/retrieval, expiry check
- `HashingConfig` — bcrypt hash + compare (34 lines)
- `TokenDTO` — access/refresh pair response model
- Interfaces: `ITokenProvider`, `ITokenStore`, `IRefreshTokenProvider`, `IRefreshTokenProviderWithRevoke`, `IReferenceTokenProvider`

### Token provider compositions (~160 lines across `services/auth`)
- `JwtRefreshTokenProvider` — pairs two `JwtTokenProvider`s (short-lived access + long-lived refresh)
- `JwtRefreshTokenWithRevokeProvider` — wraps refresh provider + `ITokenStore` for server-side revocation
- `JwtReferenceTokenProvider` — opaque reference tokens backed by a store (neither access nor refresh JWT leaves the server)

### Auth middleware (~75 lines in `services/rest`)
- `JwtAuthenticationMiddleware` — extracts bearer token, parses JWT, puts `Token` in context
- `JwtReferenceTokenAuthenicationMiddleware` — same but looks up opaque token via store

## Rating: High value

This is the most architecturally substantial package in the repo. It's not wrapping stdlib — it's building a layered token system on top of `golang-jwt` with three concrete strategies (stateless JWT, JWT with revocation, opaque reference tokens) that compose through clean interfaces. The `ITokenStore` interface means the storage backend is pluggable without touching any provider logic.

The bcrypt wrapper is low value on its own (2 functions over stdlib), but it's a minor addition to a package you'd already import.

## Security issues to fix during migration

### 1. Access and refresh tokens are interchangeable (Critical)
`SignClaims` never writes a `type` claim into the JWT, and `GetClaims` never checks one. The `Token.Type` field exists but is never populated. A refresh token can be used as an access token (and vice versa) if both providers share the same signing key — which `NewDefaultJwtRefreshTokenProvider` does. An attacker who captures a long-lived refresh token (30 days) gets full access token privileges.

**Fix:** Write a `type` claim during signing and validate it during parsing. Access and refresh providers must reject tokens of the wrong type.

### 2. Hardcoded default secret key (Critical)
`defaultSecretKey = "SuperSecretKeyShouldBeOverriden"` is a publicly known value. `NewDefaultJwtTokenProvider()` and `NewDefaultJwtRefreshTokenProvider()` are one call away from production use with this key. Anyone can forge valid tokens.

**Fix:** Remove default secret key entirely. Require explicit key configuration — constructor should return an error if the key is empty or missing.

### 3. Claims map is mutated in place (High)
`SignClaims` writes `iss`, `iat`, `exp`, `owner` directly into the caller's map. If a caller reuses the same claims map for access then refresh token generation, the second call overwrites the first's expiry. This can produce refresh tokens with access token lifetimes. Also a race condition if called concurrently with a shared map.

**Fix:** Clone the map before writing into it.

### 4. bcrypt 72-byte silent truncation (Medium)
`HashingConfig.Hash` passes the password straight to `bcrypt.GenerateFromPassword`, which silently truncates at 72 bytes. Two passwords sharing the same first 72 bytes hash identically.

**Fix:** Either reject passwords over 72 bytes, or pre-hash with SHA-256 before bcrypt.

### 5. UUID generation failure silently ignored (Medium)
`id, _ := uuid.NewV7()` in both `JwtRefreshTokenWithRevokeProvider` and `JwtReferenceTokenProvider`. If generation fails, a zero-value UUID is returned. Multiple tokens get the same ID — `StoreToken` could overwrite an existing token or `RevokeToken` could revoke the wrong one.

**Fix:** Return the error instead of discarding it.

### 6. No audience validation (Medium)
`GetClaims` extracts `owner`, `iat`, `exp` but never checks `aud`. Without audience binding, a token issued for Service A is valid at Service B if they share a signing key — common in microservice setups.

**Fix:** Add optional audience validation to `JwtTokenProviderConfig` and enforce it during parsing.

## Other issues to address during migration

- `zap.L().Fatal()` calls in config validation — a library should never `Fatal`. Return errors instead.
- The typos (`UseAuthenitcation`, `hanlder`, `Authenication`) flagged in the plan are still present in the middleware.
- `RevokeToken`/`RevokeOwner` in the revoke provider have redundant `if err != nil { return err }; return nil` — just `return err`.
