package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	. "github.com/nayefradwi/nayef_go_common/errors"
)

var ErrSessionNotFound = errors.New("session not found")

type Session struct {
	FamilyId        uuid.UUID
	OwnerId         uuid.UUID
	RefreshHash     []byte
	AccessHash      []byte
	Claims          map[string]any
	ExpiresAt       time.Time
	AccessExpiresAt time.Time
	RotatedAt       time.Time
}

type SessionManager struct {
	store      SessionStore
	jwt        *JwtTokenProvider
	accessTTL  time.Duration
	refreshTTL time.Duration
}

type issuedSession struct {
	session Session
	refresh string
	access  string
}

func (i *issuedSession) bind(familyId, ownerId uuid.UUID, claims map[string]any) {
	i.session.FamilyId, i.session.OwnerId, i.session.Claims = familyId, ownerId, claims
}

func (i *issuedSession) setAccess(token string, hash []byte, expiresAt time.Time) {
	i.access = token
	i.session.AccessHash, i.session.AccessExpiresAt = hash, expiresAt
}

func NewJwtSessionManager(store SessionStore, access JwtTokenProvider, refreshTTL time.Duration) (SessionManager, error) {
	if access.Config.TokenType != AccessTokenType {
		return SessionManager{}, BadRequestError("session jwt provider must issue access tokens")
	}
	return newSessionManager(store, &access, access.Config.ExpiresIn, refreshTTL)
}

func NewOpaqueSessionManager(store SessionStore, accessTTL, refreshTTL time.Duration) (SessionManager, error) {
	return newSessionManager(store, nil, accessTTL, refreshTTL)
}

func newSessionManager(store SessionStore, jwt *JwtTokenProvider, accessTTL, refreshTTL time.Duration) (SessionManager, error) {
	if store == nil {
		return SessionManager{}, BadRequestError("session store must not be nil")
	}

	if accessTTL <= 0 || refreshTTL <= 0 {
		return SessionManager{}, BadRequestError("session ttls must be positive")
	}

	return SessionManager{store: store, jwt: jwt, accessTTL: accessTTL, refreshTTL: refreshTTL}, nil
}

func (s SessionManager) Issue(ctx context.Context, ownerId uuid.UUID, claims map[string]any) (TokenDTO, error) {
	familyId, err := uuid.NewV7()
	if err != nil {
		return EmptyTokenDTO(), InternalError("failed to generate session id: " + err.Error())
	}

	next, err := s.newSession()
	if err != nil {
		return EmptyTokenDTO(), err
	}

	next.bind(familyId, ownerId, claims)
	if err := s.store.Create(ctx, next.session); err != nil {
		return EmptyTokenDTO(), err
	}

	return s.tokens(next)
}

func (s SessionManager) Refresh(ctx context.Context, refreshToken string) (TokenDTO, error) {
	hash := HashOpaqueToken(refreshToken)
	next, err := s.newSession()
	if err != nil {
		return EmptyTokenDTO(), err
	}

	rotated, err := s.store.Rotate(ctx, hash, next.session)
	if err == nil {
		next.session = rotated
		return s.tokens(next)
	}

	if !errors.Is(err, ErrSessionNotFound) {
		return EmptyTokenDTO(), err
	}

	old, err := s.store.GetByRefresh(ctx, hash)
	if errors.Is(err, ErrSessionNotFound) {
		return EmptyTokenDTO(), UnauthorizedError("Invalid token")
	}

	if err != nil {
		return EmptyTokenDTO(), err
	}

	if !old.RotatedAt.IsZero() {
		// a rotated refresh token came back: assume it leaked and end the whole login
		if err := s.store.DeleteFamily(ctx, old.FamilyId); err != nil {
			return EmptyTokenDTO(), err
		}
	}

	return EmptyTokenDTO(), UnauthorizedError("Invalid token")
}

func (s SessionManager) Revoke(ctx context.Context, refreshToken string) error {
	old, err := s.store.GetByRefresh(ctx, HashOpaqueToken(refreshToken))
	if errors.Is(err, ErrSessionNotFound) {
		return nil
	}

	if err != nil {
		return err
	}

	return s.store.DeleteFamily(ctx, old.FamilyId)
}

func (s SessionManager) RevokeOwner(ctx context.Context, ownerId uuid.UUID) error {
	return s.store.DeleteOwner(ctx, ownerId)
}

func (s SessionManager) VerifyAccess(ctx context.Context, accessToken string) (Token, error) {
	if s.jwt != nil {
		return Token{}, InternalError("jwt sessions are verified with the jwt provider")
	}

	session, err := s.store.GetByAccess(ctx, HashOpaqueToken(accessToken))
	if errors.Is(err, ErrSessionNotFound) {
		return Token{}, UnauthorizedError("Invalid token")
	}

	if err != nil {
		return Token{}, err
	}

	if time.Now().After(session.AccessExpiresAt) {
		return Token{}, UnauthorizedError("Invalid token")
	}

	return Token{
		OwnerId:   session.OwnerId,
		ExpiresAt: session.AccessExpiresAt,
		Claims:    session.Claims,
		Type:      AccessTokenType,
	}, nil
}

func (s SessionManager) DeleteExpired(ctx context.Context) error {
	return s.store.DeleteExpired(ctx)
}

func (s SessionManager) newSession() (issuedSession, error) {
	now := time.Now().UTC()
	refresh, refreshHash, err := NewOpaqueToken()
	if err != nil {
		return issuedSession{}, err
	}

	next := issuedSession{
		session: Session{RefreshHash: refreshHash, ExpiresAt: now.Add(s.refreshTTL)},
		refresh: refresh,
	}

	if s.jwt != nil {
		return next, nil
	}

	access, accessHash, err := NewOpaqueToken()
	if err != nil {
		return issuedSession{}, err
	}

	next.setAccess(access, accessHash, now.Add(s.accessTTL))
	return next, nil
}

func (s SessionManager) tokens(issued issuedSession) (TokenDTO, error) {
	if s.jwt == nil {
		return NewTokenDTOWithRefresh(issued.access, issued.refresh), nil
	}

	signed, err := s.jwt.SignClaims(issued.session.OwnerId, issued.session.Claims)
	if err != nil {
		return EmptyTokenDTO(), err
	}

	return NewTokenDTOWithRefresh(signed, issued.refresh), nil
}

var _ AccessVerifier = SessionManager{}
