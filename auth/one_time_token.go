package auth

import (
	"context"

	. "github.com/nayefradwi/nayef_go_common/errors"
)

type OneTimeTokenManager struct {
	store  CodeStore
	limits CodeLimits
}

func NewOneTimeTokenManager(store CodeStore, limits CodeLimits) (OneTimeTokenManager, error) {
	if store == nil {
		return OneTimeTokenManager{}, BadRequestError("code store must not be nil")
	}

	if err := validLimits(limits); err != nil {
		return OneTimeTokenManager{}, err
	}

	return OneTimeTokenManager{store: store, limits: limits}, nil
}

func (m OneTimeTokenManager) Issue(ctx context.Context, key string) (string, error) {
	token, hash, err := NewOpaqueToken()
	if err != nil {
		return "", err
	}

	if err := m.store.Save(ctx, key, hash, m.limits); err != nil {
		return "", err
	}

	return token, nil
}

func (m OneTimeTokenManager) Check(ctx context.Context, key, token string) error {
	_, err := checkCode(ctx, m.store, key, m.limits.MaxAttempts, HashOpaqueToken(token))
	return err
}

func (m OneTimeTokenManager) Consume(ctx context.Context, key, token string) error {
	return consumeCode(ctx, m.store, key, HashOpaqueToken(token))
}

func (m OneTimeTokenManager) Verify(ctx context.Context, key, token string) error {
	if err := m.Check(ctx, key, token); err != nil {
		return err
	}

	return m.Consume(ctx, key, token)
}

func (m OneTimeTokenManager) DeleteExpired(ctx context.Context) error {
	return m.store.DeleteExpired(ctx)
}
