package auth

type Claim[T string | int | float32 | float64] struct {
	Key   string
	Value T
}

func NewClaim[T string | int | float32 | float64](key string, value T) Claim[T] {
	return Claim[T]{Key: key, Value: value}
}

func ClaimsFromMap[T string | int | float32 | float64](m map[string]T) []Claim[T] {
	claims := make([]Claim[T], 0, len(m))

	for k, v := range m {
		claims = append(claims, NewClaim(k, v))
	}

	return claims
}

func (c Claim[T]) GetKey() string {
	return c.Key
}

func (c Claim[T]) GetValue() T {
	return c.Value
}
