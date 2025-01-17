package auth

type JwtReferenceTokenProvider[T string | int] struct {
	tokenProvider IRefreshTokenProvider[T]
}

var DefaultJwtReferenceTokenProviderConfig = NewJwtReferenceTokenProvider[string](
	NewDefaultJwtRefreshTokenProvider[string](),
)

func NewJwtReferenceTokenProvider[T string | int](tokenProvider IRefreshTokenProvider[T]) JwtReferenceTokenProvider[T] {
	return JwtReferenceTokenProvider[T]{
		tokenProvider: tokenProvider,
	}
}
