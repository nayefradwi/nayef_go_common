package pagination

func MapCursorPage[T any, D any](p CursorPage[T], f func(T) D) CursorPage[D] {
	items := make([]D, len(p.Items))
	for i, v := range p.Items {
		items[i] = f(v)
	}

	return CursorPage[D]{
		HasNext: p.HasNext,
		HasPrev: p.HasPrev,
		Items:   items,
		After:   p.After,
		Before:  p.Before,
	}
}
