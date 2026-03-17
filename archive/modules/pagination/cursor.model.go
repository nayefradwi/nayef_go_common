package pagination

type CursorPage[T any] struct {
	HasNext bool   `json:"has_next"`
	HasPrev bool   `json:"has_prev"`
	Items   []T    `json:"items"`
	After   string `json:"after"`
	Before  string `json:"before"`
}

func NewCursorPage[T any](hasNext, hasPrev bool, items []T, after, before string) CursorPage[T] {
	return CursorPage[T]{
		HasNext: hasNext,
		HasPrev: hasPrev,
		Items:   items,
		After:   after,
		Before:  before,
	}
}
