package auth

import (
	"bytes"
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

type memTotpStore struct {
	mu   sync.Mutex
	rows map[uuid.UUID]*Totp
}

func (m *memTotpStore) Save(_ context.Context, owner uuid.UUID, secret []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if r, ok := m.rows[owner]; ok && !r.ConfirmedAt.IsZero() {
		return ErrTotpEnabled
	}
	m.rows[owner] = &Totp{OwnerId: owner, Secret: secret}
	return nil
}

func (m *memTotpStore) Get(_ context.Context, owner uuid.UUID) (Totp, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	r, ok := m.rows[owner]
	if !ok {
		return Totp{}, ErrTotpNotFound
	}
	return *r, nil
}

func (m *memTotpStore) Confirm(_ context.Context, owner uuid.UUID, step int64, codes [][]byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	r, ok := m.rows[owner]
	if !ok || !r.ConfirmedAt.IsZero() || r.LastStep >= step {
		return ErrTotpNotFound
	}
	r.ConfirmedAt, r.LastStep, r.RecoveryCodes = time.Now(), step, codes
	return nil
}

func (m *memTotpStore) SetRecoveryCodes(_ context.Context, owner uuid.UUID, codes [][]byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	r, ok := m.rows[owner]
	if !ok || r.ConfirmedAt.IsZero() {
		return ErrTotpNotFound
	}
	r.RecoveryCodes = codes
	return nil
}

func (m *memTotpStore) UseRecoveryCode(_ context.Context, owner uuid.UUID, hash []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	r, ok := m.rows[owner]
	if !ok || r.ConfirmedAt.IsZero() {
		return ErrTotpNotFound
	}
	for i, h := range r.RecoveryCodes {
		if bytes.Equal(h, hash) {
			r.RecoveryCodes = append(r.RecoveryCodes[:i:i], r.RecoveryCodes[i+1:]...)
			return nil
		}
	}
	return ErrTotpNotFound
}

func (m *memTotpStore) UseStep(_ context.Context, owner uuid.UUID, step int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	r, ok := m.rows[owner]
	if !ok || r.ConfirmedAt.IsZero() || r.LastStep >= step {
		return ErrTotpNotFound
	}
	r.LastStep = step
	return nil
}

func (m *memTotpStore) Delete(_ context.Context, owner uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.rows, owner)
	return nil
}

var testTotpStart = time.Unix(1_700_000_010, 0)

func mustTotp(t *testing.T) (TotpManager, *memTotpStore, *time.Time) {
	t.Helper()
	store := &memTotpStore{rows: map[uuid.UUID]*Totp{}}
	m, err := NewTotpManager(store, TotpConfig{Key: []byte(testSecret[:32]), Issuer: "acme"})
	if err != nil {
		t.Fatal(err)
	}
	clock := testTotpStart
	m.now = func() time.Time { return clock }
	return m, store, &clock
}

func codeAt(t *testing.T, secret string, at time.Time) string {
	t.Helper()
	raw, err := totpEncoding.DecodeString(secret)
	if err != nil {
		t.Fatal(err)
	}
	return totpCode(raw, uint64(at.Unix()/totpPeriod), totpDigits)
}

func enrolled(t *testing.T, m TotpManager, clock *time.Time) string {
	secret, _ := enrolledWithCodes(t, m, clock)
	return secret
}

func enrolledWithCodes(t *testing.T, m TotpManager, clock *time.Time) (string, []string) {
	t.Helper()
	ctx := context.Background()
	secret, _, err := m.Enroll(ctx, testOwner, "a@b.c")
	if err != nil {
		t.Fatal(err)
	}
	codes, err := m.Confirm(ctx, testOwner, codeAt(t, secret, *clock))
	if err != nil {
		t.Fatal(err)
	}
	*clock = clock.Add(totpPeriod * time.Second)
	return secret, codes
}

func TestTotpCode_RFC6238Vectors(t *testing.T) {
	secret := []byte("12345678901234567890")
	vectors := map[int64]string{
		59:          "94287082",
		1111111109:  "07081804",
		1111111111:  "14050471",
		1234567890:  "89005924",
		2000000000:  "69279037",
		20000000000: "65353130",
	}
	for unix, want := range vectors {
		if got := totpCode(secret, uint64(unix/30), 8); got != want {
			t.Errorf("T=%d: got %s, want %s", unix, got, want)
		}
	}
}

func TestTotp_SkewIsOneStep(t *testing.T) {
	step := totpPeriod * time.Second
	for _, tc := range []struct {
		offset time.Duration
		ok     bool
	}{{-step, true}, {0, true}, {step, true}, {-2 * step, false}, {2 * step, false}} {
		m, _, clock := mustTotp(t)
		secret := enrolled(t, m, clock)
		*clock = clock.Add(10 * step)

		err := m.Verify(context.Background(), testOwner, codeAt(t, secret, clock.Add(tc.offset)))
		if (err == nil) != tc.ok {
			t.Errorf("offset %v: err = %v, want ok = %v", tc.offset, err, tc.ok)
		}
	}
}

func TestTotp_CodeCannotBeReplayed(t *testing.T) {
	m, _, clock := mustTotp(t)
	secret := enrolled(t, m, clock)
	code := codeAt(t, secret, *clock)

	if err := m.Verify(context.Background(), testOwner, code); err != nil {
		t.Fatal(err)
	}
	if err := m.Verify(context.Background(), testOwner, code); err == nil {
		t.Fatal("code accepted twice")
	}
}

func TestTotp_VerifyBeforeConfirm(t *testing.T) {
	m, _, clock := mustTotp(t)
	secret, _, err := m.Enroll(context.Background(), testOwner, "a@b.c")
	if err != nil {
		t.Fatal(err)
	}

	if err := m.Verify(context.Background(), testOwner, codeAt(t, secret, *clock)); err == nil {
		t.Fatal("unconfirmed totp verified")
	}
}

func TestTotp_EnrollCannotReplaceConfirmed(t *testing.T) {
	m, _, clock := mustTotp(t)
	secret := enrolled(t, m, clock)

	if _, _, err := m.Enroll(context.Background(), testOwner, "a@b.c"); err == nil {
		t.Fatal("re-enroll over confirmed totp")
	}
	if err := m.Verify(context.Background(), testOwner, codeAt(t, secret, *clock)); err != nil {
		t.Fatalf("old secret stopped working: %v", err)
	}
}

func TestTotp_SecretBoundToOwner(t *testing.T) {
	m, store, clock := mustTotp(t)
	secret := enrolled(t, m, clock)
	other := uuid.New()
	row := *store.rows[testOwner]
	row.OwnerId = other
	store.rows[other] = &row

	if err := m.Verify(context.Background(), other, codeAt(t, secret, *clock)); err == nil {
		t.Fatal("secret opened under another owner")
	}
}

func TestTotp_EnrollUri(t *testing.T) {
	m, _, _ := mustTotp(t)
	secret, uri, err := m.Enroll(context.Background(), testOwner, "a b@c.d")
	if err != nil {
		t.Fatal(err)
	}
	want := "otpauth://totp/acme:a%20b@c.d?issuer=acme&secret=" + secret
	if uri != want {
		t.Fatalf("got %s, want %s", uri, want)
	}
}

func TestTotp_RecoveryCodeWorksOnce(t *testing.T) {
	ctx := context.Background()
	m, _, clock := mustTotp(t)
	_, codes := enrolledWithCodes(t, m, clock)
	if len(codes) != recoveryCodeCount {
		t.Fatalf("got %d codes", len(codes))
	}

	if err := m.UseRecoveryCode(ctx, testOwner, codes[0]); err != nil {
		t.Fatal(err)
	}
	if err := m.UseRecoveryCode(ctx, testOwner, codes[0]); err == nil {
		t.Fatal("recovery code used twice")
	}
}

func TestTotp_RecoveryCodeNormalized(t *testing.T) {
	ctx := context.Background()
	m, _, clock := mustTotp(t)
	_, codes := enrolledWithCodes(t, m, clock)

	if err := m.UseRecoveryCode(ctx, testOwner, " "+strings.ToUpper(codes[0][:3])+codes[0][3:]+" "); err != nil {
		t.Fatalf("%q: %v", codes[0], err)
	}
	if err := m.UseRecoveryCode(ctx, testOwner, strings.ReplaceAll(codes[1], "-", "")); err != nil {
		t.Fatal(err)
	}
}

func TestTotp_RegenerateReplacesRecoveryCodes(t *testing.T) {
	ctx := context.Background()
	m, _, clock := mustTotp(t)
	_, old := enrolledWithCodes(t, m, clock)
	if _, err := m.RegenerateRecoveryCodes(ctx, testOwner); err != nil {
		t.Fatal(err)
	}

	if err := m.UseRecoveryCode(ctx, testOwner, old[0]); err == nil {
		t.Fatal("old code still works")
	}
}

func TestTotp_DisableRemovesRecoveryCodes(t *testing.T) {
	ctx := context.Background()
	m, _, clock := mustTotp(t)
	_, codes := enrolledWithCodes(t, m, clock)
	if err := m.Disable(ctx, testOwner); err != nil {
		t.Fatal(err)
	}

	if err := m.UseRecoveryCode(ctx, testOwner, codes[0]); err == nil {
		t.Fatal("recovery code works after disable")
	}
}
