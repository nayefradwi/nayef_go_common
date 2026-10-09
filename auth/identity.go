package auth

import (
	"context"
	"slices"

	"github.com/google/uuid"
)

type IdentityKind string

const (
	KindUser IdentityKind = "user"
	KindKey  IdentityKind = "key"
)

type Identity struct {
	OwnerId uuid.UUID
	Kind    IdentityKind
	Claims  map[string]any
}

type identityKey struct{}

func IdentityFromToken(t Token) Identity {
	return Identity{OwnerId: t.OwnerId, Kind: KindUser, Claims: t.Claims}
}

func WithIdentity(ctx context.Context, id Identity) context.Context {
	return context.WithValue(ctx, identityKey{}, id)
}

func GetIdentity(ctx context.Context) (Identity, bool) {
	id, ok := ctx.Value(identityKey{}).(Identity)
	return id, ok
}

func (i Identity) HasClaim(name string, values ...string) bool {
	claim, ok := i.Claims[name]
	if !ok || claim == nil || claim == false {
		return false
	}

	if len(values) == 0 {
		return true
	}

	switch c := claim.(type) {
	case string:
		return slices.Contains(values, c)
	case []string:
		return slices.ContainsFunc(c, func(s string) bool { return slices.Contains(values, s) })
	case []any:
		return slices.ContainsFunc(c, func(v any) bool {
			s, ok := v.(string)
			return ok && slices.Contains(values, s)
		})
	}

	return false
}
