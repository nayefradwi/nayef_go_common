package auth

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base32"
	"encoding/binary"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	. "github.com/nayefradwi/nayef_go_common/errors"
)

var (
	ErrTotpNotFound = errors.New("totp not found")
	ErrTotpEnabled  = errors.New("totp already enabled")
)

type Totp struct {
	OwnerId       uuid.UUID
	Secret        []byte
	LastStep      int64
	ConfirmedAt   time.Time
	RecoveryCodes [][]byte
}

type TotpConfig struct {
	Key    []byte
	Issuer string
}

type TotpManager struct {
	store       TotpStore
	aead        cipher.AEAD
	recoveryKey []byte
	issuer      string
	now         func() time.Time
}

// 6 digits, 30s, SHA-1: the only parameters authenticator apps reliably support
const (
	totpDigits      = 6
	totpPeriod      = 30
	totpSecretBytes = 20
	totpKeyBytes    = 32

	recoveryCodeCount = 10
	recoveryCodeHalf  = 5
	recoveryCodeBytes = 7 // 7 bytes → 12 base32 chars, trimmed to 10 (50 bits)
)

var totpEncoding = base32.StdEncoding.WithPadding(base32.NoPadding)

func NewTotpManager(store TotpStore, config TotpConfig) (TotpManager, error) {
	if store == nil {
		return TotpManager{}, BadRequestError("totp store must not be nil")
	}

	if len(config.Key) != totpKeyBytes {
		return TotpManager{}, BadRequestError("totp key must be 32 bytes")
	}

	if config.Issuer == "" {
		return TotpManager{}, BadRequestError("totp issuer must not be empty")
	}

	block, err := aes.NewCipher(config.Key)
	if err != nil {
		return TotpManager{}, InternalError("failed to create totp cipher: " + err.Error())
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return TotpManager{}, InternalError("failed to create totp cipher: " + err.Error())
	}

	// a derived key keeps the AES key out of the recovery code HMAC
	recoveryKey := hmacSum(config.Key, "totp-recovery-codes")
	return TotpManager{store: store, aead: aead, recoveryKey: recoveryKey, issuer: config.Issuer, now: time.Now}, nil
}

func (m TotpManager) Enroll(ctx context.Context, ownerId uuid.UUID, account string) (string, string, error) {
	secret := make([]byte, totpSecretBytes)
	if _, err := rand.Read(secret); err != nil {
		return "", "", InternalError("failed to generate totp secret: " + err.Error())
	}

	sealed, err := m.seal(ownerId, secret)
	if err != nil {
		return "", "", err
	}

	if err := m.store.Save(ctx, ownerId, sealed); err != nil {
		return "", "", err
	}

	encoded := totpEncoding.EncodeToString(secret)
	query := url.Values{"secret": {encoded}, "issuer": {m.issuer}}
	uri := fmt.Sprintf("otpauth://totp/%s?%s", url.PathEscape(m.issuer+":"+account), query.Encode())
	return encoded, uri, nil
}

func (m TotpManager) Confirm(ctx context.Context, ownerId uuid.UUID, code string) ([]string, error) {
	totp, err := m.get(ctx, ownerId)
	if err != nil {
		return nil, err
	}

	if !totp.ConfirmedAt.IsZero() {
		return nil, BadRequestError("totp already enabled")
	}

	step, err := m.match(ownerId, totp, code)
	if err != nil {
		return nil, err
	}

	codes, hashes, err := m.newRecoveryCodes()
	if err != nil {
		return nil, err
	}

	if err := invalidIfNotFound(m.store.Confirm(ctx, ownerId, step, hashes)); err != nil {
		return nil, err
	}

	return codes, nil
}

func (m TotpManager) Verify(ctx context.Context, ownerId uuid.UUID, code string) error {
	totp, err := m.get(ctx, ownerId)
	if err != nil {
		return err
	}

	if totp.ConfirmedAt.IsZero() {
		return UnauthorizedError("Invalid code")
	}

	step, err := m.match(ownerId, totp, code)
	if err != nil {
		return err
	}

	return invalidIfNotFound(m.store.UseStep(ctx, ownerId, step))
}

func (m TotpManager) Enabled(ctx context.Context, ownerId uuid.UUID) (bool, error) {
	totp, err := m.store.Get(ctx, ownerId)
	if errors.Is(err, ErrTotpNotFound) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return !totp.ConfirmedAt.IsZero(), nil
}

func (m TotpManager) RegenerateRecoveryCodes(ctx context.Context, ownerId uuid.UUID) ([]string, error) {
	codes, hashes, err := m.newRecoveryCodes()
	if err != nil {
		return nil, err
	}

	err = m.store.SetRecoveryCodes(ctx, ownerId, hashes)
	if errors.Is(err, ErrTotpNotFound) {
		return nil, BadRequestError("totp not enabled")
	}

	if err != nil {
		return nil, err
	}

	return codes, nil
}

func (m TotpManager) UseRecoveryCode(ctx context.Context, ownerId uuid.UUID, code string) error {
	return invalidIfNotFound(m.store.UseRecoveryCode(ctx, ownerId, hmacSum(m.recoveryKey, normalizeRecoveryCode(code))))
}

func (m TotpManager) RecoveryCodesLeft(ctx context.Context, ownerId uuid.UUID) (int, error) {
	totp, err := m.store.Get(ctx, ownerId)
	if errors.Is(err, ErrTotpNotFound) {
		return 0, nil
	}

	if err != nil {
		return 0, err
	}

	return len(totp.RecoveryCodes), nil
}

func (m TotpManager) Disable(ctx context.Context, ownerId uuid.UUID) error {
	return m.store.Delete(ctx, ownerId)
}

func (m TotpManager) get(ctx context.Context, ownerId uuid.UUID) (Totp, error) {
	totp, err := m.store.Get(ctx, ownerId)
	if errors.Is(err, ErrTotpNotFound) {
		return Totp{}, UnauthorizedError("Invalid code")
	}

	return totp, err
}

func (m TotpManager) match(ownerId uuid.UUID, totp Totp, code string) (int64, error) {
	if !isDigits(code, totpDigits) {
		return 0, UnauthorizedError("Invalid code")
	}

	secret, err := m.open(ownerId, totp.Secret)
	if err != nil {
		return 0, err
	}

	now := m.now().Unix() / totpPeriod
	for step := now - 1; step <= now+1; step++ {
		want := totpCode(secret, uint64(step), totpDigits)
		if subtle.ConstantTimeCompare([]byte(want), []byte(code)) == 1 {
			return step, nil
		}
	}

	return 0, UnauthorizedError("Invalid code")
}

// owner id as AAD: a secret copied onto another owner's row fails to open
func (m TotpManager) seal(ownerId uuid.UUID, secret []byte) ([]byte, error) {
	nonce := make([]byte, m.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, InternalError("failed to generate totp nonce: " + err.Error())
	}

	return m.aead.Seal(nonce, nonce, secret, ownerId[:]), nil
}

func (m TotpManager) open(ownerId uuid.UUID, sealed []byte) ([]byte, error) {
	size := m.aead.NonceSize()
	if len(sealed) < size {
		return nil, InternalError("failed to decrypt totp secret")
	}

	secret, err := m.aead.Open(nil, sealed[:size], sealed[size:], ownerId[:])
	if err != nil {
		return nil, InternalError("failed to decrypt totp secret")
	}

	return secret, nil
}

func (m TotpManager) newRecoveryCodes() ([]string, [][]byte, error) {
	codes := make([]string, recoveryCodeCount)
	hashes := make([][]byte, recoveryCodeCount)
	for i := range codes {
		b := make([]byte, recoveryCodeBytes)
		if _, err := rand.Read(b); err != nil {
			return nil, nil, InternalError("failed to generate recovery code: " + err.Error())
		}

		raw := strings.ToLower(totpEncoding.EncodeToString(b))[:2*recoveryCodeHalf]
		codes[i] = raw[:recoveryCodeHalf] + "-" + raw[recoveryCodeHalf:]
		hashes[i] = hmacSum(m.recoveryKey, raw)
	}

	return codes, hashes, nil
}

func normalizeRecoveryCode(code string) string {
	return strings.ToLower(strings.NewReplacer("-", "", " ", "").Replace(code))
}

func totpCode(secret []byte, counter uint64, digits int) string {
	h := hmac.New(sha1.New, secret)
	binary.Write(h, binary.BigEndian, counter)
	sum := h.Sum(nil)

	offset := sum[len(sum)-1] & 0x0f
	value := binary.BigEndian.Uint32(sum[offset:offset+4]) & 0x7fffffff

	mod := uint32(1)
	for range digits {
		mod *= 10
	}

	return fmt.Sprintf("%0*d", digits, value%mod)
}

func isDigits(s string, length int) bool {
	if len(s) != length {
		return false
	}

	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}

	return true
}

func invalidIfNotFound(err error) error {
	if errors.Is(err, ErrTotpNotFound) {
		return UnauthorizedError("Invalid code")
	}

	return err
}
