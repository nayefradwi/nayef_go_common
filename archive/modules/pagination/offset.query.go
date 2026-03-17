package pagination

import "context"

var DefaultPageQuery = NewOffsetPageQuery(1, 10)

const (
	PageKey     = "page"
	PageSizeKey = "page_size"
)

type OffsetPageQuery struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

type offsetPageQueryKey struct{}

func NewOffsetPageQuery(page, pageSize int) OffsetPageQuery {
	if page < 1 {
		page = 1
	}

	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	return OffsetPageQuery{
		Page:     page,
		PageSize: pageSize,
	}
}

func (q OffsetPageQuery) Offset() int {
	return (q.Page - 1) * q.PageSize
}

func (q OffsetPageQuery) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, offsetPageQueryKey{}, q)
}

func OffsetPageQueryFromContext(ctx context.Context) OffsetPageQuery {
	q, ok := ctx.Value(offsetPageQueryKey{}).(OffsetPageQuery)
	if !ok {
		return DefaultPageQuery
	}

	return q
}
