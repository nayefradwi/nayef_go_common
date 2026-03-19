package otp

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubCodeGenerator struct {
	code string
}

func (s stubCodeGenerator) GenerateOtp() string {
	return s.code
}

type stubOtpRepository struct {
	otp         *OTP
	getErr      error
	upsertErr   error
	upsertedOtp *OTP
}

func (s *stubOtpRepository) GetOtp(_ context.Context, _ string) (*OTP, error) {
	return s.otp, s.getErr
}

func (s *stubOtpRepository) UpsertOtp(_ context.Context, otp *OTP) error {
	s.upsertedOtp = otp
	return s.upsertErr
}

var defaultConfig = OtpConfig{
	ExpiresIn:   5 * time.Minute,
	MaxTries:    3,
	ResendAfter: 1 * time.Minute,
}

func TestGenerateOtp_HappyPath(t *testing.T) {
	repo := &stubOtpRepository{getErr: redis.Nil}
	gen := stubCodeGenerator{code: "123456"}
	svc := NewOtpService(repo, gen, defaultConfig)

	otp, err := svc.GenerateOtp(context.Background(), "owner-1")
	require.NoError(t, err)
	require.NotNil(t, otp)
	assert.Equal(t, "owner-1", otp.OwnerId)
	assert.Equal(t, "123456", otp.Code)
	assert.Equal(t, HashCode("123456"), otp.HashedCode)
	assert.False(t, otp.IsExpired())
	assert.NotNil(t, repo.upsertedOtp)
}

func TestGenerateOtp_ReturnsExistingOtp(t *testing.T) {
	now := time.Now().UTC()
	existing := NewOtp("owner-1", "", HashCode("999999"), 0, 3, now.Add(5*time.Minute), now)
	repo := &stubOtpRepository{otp: existing}
	gen := stubCodeGenerator{code: "123456"}
	svc := NewOtpService(repo, gen, defaultConfig)

	otp, err := svc.GenerateOtp(context.Background(), "owner-1")
	require.NoError(t, err)
	assert.Equal(t, existing, otp)
	assert.Nil(t, repo.upsertedOtp, "should not upsert when returning existing OTP")
}

func TestGenerateOtp_RegeneratesExpiredOtp(t *testing.T) {
	expired := NewOtp("owner-1", "", HashCode("old"), 0, 3, time.Now().UTC().Add(-1*time.Minute), time.Now().UTC().Add(-10*time.Minute))
	repo := &stubOtpRepository{otp: expired}
	gen := stubCodeGenerator{code: "newcode"}
	svc := NewOtpService(repo, gen, defaultConfig)

	otp, err := svc.GenerateOtp(context.Background(), "owner-1")
	require.NoError(t, err)
	assert.Equal(t, "newcode", otp.Code)
	assert.NotNil(t, repo.upsertedOtp)
}

func TestGenerateOtp_UpsertFails(t *testing.T) {
	repo := &stubOtpRepository{
		getErr:    redis.Nil,
		upsertErr: fmt.Errorf("redis down"),
	}
	gen := stubCodeGenerator{code: "123456"}
	svc := NewOtpService(repo, gen, defaultConfig)

	otp, err := svc.GenerateOtp(context.Background(), "owner-1")
	assert.Nil(t, otp)
	assert.Error(t, err)
}

func TestVerifyOtp_HappyPath(t *testing.T) {
	code := "123456"
	otp := NewOtp("owner-1", code, HashCode(code), 0, 3, time.Now().UTC().Add(5*time.Minute), time.Now().UTC())
	repo := &stubOtpRepository{otp: otp}
	gen := stubCodeGenerator{}
	svc := NewOtpService(repo, gen, defaultConfig)

	err := svc.VerifyOtp(context.Background(), "owner-1", code)
	assert.NoError(t, err)
}

func TestVerifyOtp_ExpiredCode(t *testing.T) {
	code := "123456"
	otp := NewOtp("owner-1", code, HashCode(code), 0, 3, time.Now().UTC().Add(-1*time.Minute), time.Now().UTC())
	repo := &stubOtpRepository{otp: otp}
	svc := NewOtpService(repo, stubCodeGenerator{}, defaultConfig)

	err := svc.VerifyOtp(context.Background(), "owner-1", code)
	assert.Equal(t, ErrExpiredOTP, err)
}

func TestVerifyOtp_WrongOwner(t *testing.T) {
	repo := &stubOtpRepository{getErr: redis.Nil}
	svc := NewOtpService(repo, stubCodeGenerator{}, defaultConfig)

	err := svc.VerifyOtp(context.Background(), "other-owner", "123456")
	assert.Equal(t, ErrOtpNotFound, err)
}

func TestVerifyOtp_IncorrectCode(t *testing.T) {
	code := "123456"
	otp := NewOtp("owner-1", code, HashCode(code), 0, 3, time.Now().UTC().Add(5*time.Minute), time.Now().UTC())
	repo := &stubOtpRepository{otp: otp}
	svc := NewOtpService(repo, stubCodeGenerator{}, defaultConfig)

	err := svc.VerifyOtp(context.Background(), "owner-1", "wrong-code")
	assert.Equal(t, ErrIncorrectOTP, err)
	assert.NotNil(t, repo.upsertedOtp, "should persist incremented retry count")
	assert.Equal(t, 1, repo.upsertedOtp.RetryCount)
}

func TestVerifyOtp_MaxRetriesExceeded(t *testing.T) {
	code := "123456"
	otp := NewOtp("owner-1", code, HashCode(code), 3, 3, time.Now().UTC().Add(5*time.Minute), time.Now().UTC())
	repo := &stubOtpRepository{otp: otp}
	svc := NewOtpService(repo, stubCodeGenerator{}, defaultConfig)

	err := svc.VerifyOtp(context.Background(), "owner-1", code)
	assert.Equal(t, ErrMaxTriesExceeded, err)
}

func TestOtpJsonSerialization(t *testing.T) {
	otp := NewOtp("owner-1", "secret-code", "hashed-secret", 0, 3, time.Now().UTC(), time.Now().UTC())
	data, err := json.Marshal(otp)
	require.NoError(t, err)

	var fields map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &fields))

	assert.Contains(t, fields, "owner_id")
	assert.NotContains(t, fields, "code")
	assert.NotContains(t, fields, "hashed_code")
	assert.NotContains(t, fields, "Code")
	assert.NotContains(t, fields, "HashedCode")
}
