package auth

import (
	"bytes"
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type memOtp struct {
	hash             []byte
	attempts, sends  int
	sentAt, expireAt time.Time
}

type memOtpStore struct {
	mu   sync.Mutex
	rows map[string]*memOtp
}

func (m *memOtpStore) Save(_ context.Context, key string, hash []byte, l OtpLimits) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	r, ok := m.rows[key]
	if !ok || !now.Before(r.expireAt) {
		m.rows[key] = &memOtp{hash: hash, sends: 1, sentAt: now, expireAt: now.Add(l.TTL)}
		return nil
	}
	if now.Before(r.sentAt.Add(l.ResendAfter)) || r.sends >= l.MaxSends || r.attempts >= l.MaxAttempts {
		return ErrOtpResendBlocked
	}
	r.hash, r.sends, r.sentAt, r.expireAt = hash, r.sends+1, now, now.Add(l.TTL)
	return nil
}

func (m *memOtpStore) Attempt(_ context.Context, key string, max int) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	r, ok := m.rows[key]
	if !ok || r.attempts >= max || !time.Now().Before(r.expireAt) {
		return nil, ErrOtpNotFound
	}
	r.attempts++
	return r.hash, nil
}

func (m *memOtpStore) Consume(_ context.Context, key string, hash []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	r, ok := m.rows[key]
	if !ok || !bytes.Equal(r.hash, hash) {
		return ErrOtpNotFound
	}
	delete(m.rows, key)
	return nil
}

func (m *memOtpStore) DeleteExpired(context.Context) error { return nil }

func testOtpConfig() OtpConfig {
	return OtpConfig{
		Secret:    []byte(testSecret),
		Length:    6,
		OtpLimits: OtpLimits{TTL: time.Minute, MaxSends: 3, MaxAttempts: 3},
	}
}

func mustOtpManager(t *testing.T) OtpManager {
	t.Helper()
	m, err := NewOtpManager(&memOtpStore{rows: map[string]*memOtp{}}, testOtpConfig())
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func wrong(code string) string {
	if code[0] == '0' {
		return "1" + code[1:]
	}
	return "0" + code[1:]
}

func TestOtp_WrongThenRightThenReused(t *testing.T) {
	ctx := context.Background()
	m := mustOtpManager(t)
	code, err := m.Issue(ctx, "login:a")
	if err != nil {
		t.Fatal(err)
	}

	if err := m.Verify(ctx, "login:a", wrong(code)); err == nil {
		t.Fatal("wrong code accepted")
	}
	if err := m.Verify(ctx, "login:b", code); err == nil {
		t.Fatal("code accepted under another key")
	}
	if err := m.Verify(ctx, "login:a", code); err != nil {
		t.Fatal("right code rejected after one miss", err)
	}
	if err := m.Verify(ctx, "login:a", code); err == nil {
		t.Fatal("used code accepted twice")
	}
}

func TestOtp_LockedAfterMaxAttempts(t *testing.T) {
	ctx := context.Background()
	m := mustOtpManager(t)
	code, _ := m.Issue(ctx, "k")
	for range m.config.MaxAttempts {
		_ = m.Verify(ctx, "k", wrong(code))
	}
	if err := m.Verify(ctx, "k", code); err == nil {
		t.Fatal("right code accepted after lockout")
	}
	if _, err := m.Issue(ctx, "k"); !errors.Is(err, ErrOtpResendBlocked) {
		t.Fatal("resend while locked should be blocked", err)
	}
}

func TestOtp_ResendReplacesCode(t *testing.T) {
	ctx := context.Background()
	m := mustOtpManager(t)
	first, _ := m.Issue(ctx, "k")
	second, err := m.Issue(ctx, "k")
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Skip("codes collided, 1 in 10^6")
	}
	if err := m.Verify(ctx, "k", first); err == nil {
		t.Fatal("old code accepted after resend")
	}
	if err := m.Verify(ctx, "k", second); err != nil {
		t.Fatal(err)
	}
}

func TestNewCode(t *testing.T) {
	code, err := newCode(6)
	if err != nil {
		t.Fatal(err)
	}
	if len(code) != 6 {
		t.Fatalf("len %d", len(code))
	}
	for _, c := range code {
		if c < '0' || c > '9' {
			t.Fatalf("non-digit %q", code)
		}
	}
}

func TestNewOtpManager_Rejects(t *testing.T) {
	store := &memOtpStore{}
	bad := map[string]func(*OtpConfig){
		"short secret": func(c *OtpConfig) { c.Secret = c.Secret[:31] },
		"length 5":     func(c *OtpConfig) { c.Length = 5 },
		"zero ttl":     func(c *OtpConfig) { c.TTL = 0 },
		"zero sends":   func(c *OtpConfig) { c.MaxSends = 0 },
		"zero tries":   func(c *OtpConfig) { c.MaxAttempts = 0 },
	}
	for name, mutate := range bad {
		c := testOtpConfig()
		c.Secret = []byte(testSecret[:32])
		mutate(&c)
		if _, err := NewOtpManager(store, c); err == nil {
			t.Error(name + " should be rejected")
		}
	}
	if _, err := NewOtpManager(nil, testOtpConfig()); err == nil {
		t.Error("nil store should be rejected")
	}
	c := testOtpConfig()
	c.Secret = []byte(testSecret[:32])
	if _, err := NewOtpManager(store, c); err != nil {
		t.Error("32-byte secret, length 6 should pass", err)
	}
}
