package pagination

import (
	"context"
	"encoding/base64"
	"encoding/json"
)

const (
	AfterKey  = "after"
	BeforeKey = "before"
)

func EncodeCursor[C any](v C) string {
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

func DecodeCursor[C any](s string) (v C, has bool, err error) {
	if s == "" {
		return v, false, nil
	}

	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return v, false, err
	}

	if err := json.Unmarshal(b, &v); err != nil {
		return v, false, err
	}

	return v, true, nil
}

type CursorPageQuery struct {
	After    string `json:"after"`
	Before   string `json:"before"`
	PageSize int    `json:"page_size"`
}

type cursorPageQueryKey struct{}

func NewCursorPageQuery(after, before string, pageSize int) CursorPageQuery {
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	return CursorPageQuery{
		After:    after,
		Before:   before,
		PageSize: pageSize,
	}
}

func (q CursorPageQuery) Forward() bool {
	return q.Before == ""
}

func (q CursorPageQuery) Cursor() string {
	if q.Forward() {
		return q.After
	}
	return q.Before
}

func (q CursorPageQuery) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, cursorPageQueryKey{}, q)
}

func CursorPageQueryFromContext(ctx context.Context) CursorPageQuery {
	q, ok := ctx.Value(cursorPageQueryKey{}).(CursorPageQuery)
	if !ok {
		return NewCursorPageQuery("", "", 10)
	}

	return q
}
