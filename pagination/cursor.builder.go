package pagination

import (
	"context"
	"errors"
	"slices"
)

var ErrBackwardUnsupported = errors.New("pagination: backward paging not supported")

type CursorBuilder[T any, C any] struct {
	query    CursorPageQuery
	forward  func(ctx context.Context, cur C, first bool, limit int32) ([]T, error)
	backward func(ctx context.Context, cur C, limit int32) ([]T, error)
	key      func(T) C
	reverse  bool
}

func NewCursor[T any, C any](q CursorPageQuery) *CursorBuilder[T, C] {
	return &CursorBuilder[T, C]{query: q, reverse: true}
}

func (b *CursorBuilder[T, C]) OnForward(f func(ctx context.Context, cur C, first bool, limit int32) ([]T, error)) *CursorBuilder[T, C] {
	b.forward = f
	return b
}

func (b *CursorBuilder[T, C]) OnBackward(f func(ctx context.Context, cur C, limit int32) ([]T, error)) *CursorBuilder[T, C] {
	b.backward = f
	return b
}

func (b *CursorBuilder[T, C]) Key(f func(T) C) *CursorBuilder[T, C] {
	b.key = f
	return b
}

// NoReverse: use when the backward query already returns rows in natural order.
func (b *CursorBuilder[T, C]) NoReverse() *CursorBuilder[T, C] {
	b.reverse = false
	return b
}

func (b *CursorBuilder[T, C]) Execute(ctx context.Context) (CursorPage[T], error) {
	limit := int32(b.query.PageSize + 1)
	forward := b.query.Forward()

	var rows []T
	var hadCursor bool
	var err error

	if forward {
		cur, has, derr := DecodeCursor[C](b.query.After)
		if derr != nil {
			return CursorPage[T]{}, derr
		}
		hadCursor = has
		rows, err = b.forward(ctx, cur, !has, limit)
	} else {
		if b.backward == nil {
			return CursorPage[T]{}, ErrBackwardUnsupported
		}
		cur, _, derr := DecodeCursor[C](b.query.Before)
		if derr != nil {
			return CursorPage[T]{}, derr
		}
		hadCursor = true
		rows, err = b.backward(ctx, cur, limit)
	}
	if err != nil {
		return CursorPage[T]{}, err
	}

	hasExtra := len(rows) > b.query.PageSize
	if hasExtra {
		rows = rows[:b.query.PageSize]
	}
	if !forward && b.reverse {
		slices.Reverse(rows)
	}

	hasNext, hasPrev := hasExtra, hadCursor
	if !forward {
		hasNext, hasPrev = true, hasExtra
	}

	var after, before string
	if len(rows) > 0 {
		after = EncodeCursor(b.key(rows[len(rows)-1]))
		before = EncodeCursor(b.key(rows[0]))
	}

	return NewCursorPage(hasNext, hasPrev, rows, after, before), nil
}
