package auth

import (
	"context"
	"testing"
	"time"
)

func mustOneTimeTokens(t *testing.T, store *memOtpStore) OneTimeTokenManager {
	t.Helper()
	m, err := NewOneTimeTokenManager(store, CodeLimits{TTL: time.Minute, MaxSends: 3, MaxAttempts: 3})
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func TestOneTimeToken_SingleUse(t *testing.T) {
	ctx := context.Background()
	m := mustOneTimeTokens(t, &memOtpStore{rows: map[string]*memOtp{}})
	token, err := m.Issue(ctx, "reset:a@b.c")
	if err != nil {
		t.Fatal(err)
	}

	if err := m.Verify(ctx, "reset:a@b.c", token); err != nil {
		t.Fatalf("first verify: %v", err)
	}
	if err := m.Verify(ctx, "reset:a@b.c", token); err == nil {
		t.Fatal("token verified twice")
	}
}

func TestOneTimeToken_CheckKeepsTokenUntilConsume(t *testing.T) {
	ctx := context.Background()
	m := mustOneTimeTokens(t, &memOtpStore{rows: map[string]*memOtp{}})
	token, _ := m.Issue(ctx, "mfa:1")

	for range 2 {
		if err := m.Check(ctx, "mfa:1", token); err != nil {
			t.Fatalf("check: %v", err)
		}
	}
	if err := m.Consume(ctx, "mfa:1", token); err != nil {
		t.Fatalf("consume: %v", err)
	}
}

func TestOneTimeToken_LockedAfterMaxAttempts(t *testing.T) {
	ctx := context.Background()
	m := mustOneTimeTokens(t, &memOtpStore{rows: map[string]*memOtp{}})
	token, _ := m.Issue(ctx, "mfa:1")

	for range 3 {
		if err := m.Check(ctx, "mfa:1", "guess"); err == nil {
			t.Fatal("wrong token passed")
		}
	}
	if err := m.Check(ctx, "mfa:1", token); err == nil {
		t.Fatal("right token passed after max attempts")
	}
}

func TestOneTimeToken_ConsumeRejectsExpired(t *testing.T) {
	ctx := context.Background()
	store := &memOtpStore{rows: map[string]*memOtp{}}
	m := mustOneTimeTokens(t, store)
	token, _ := m.Issue(ctx, "reset:a@b.c")
	store.rows["reset:a@b.c"].expireAt = time.Now().Add(-time.Second)

	if err := m.Consume(ctx, "reset:a@b.c", token); err == nil {
		t.Fatal("expired token consumed")
	}
}
