package pagination

type OffsetPage[T any] struct {
	Page          int  `json:"page"`
	PageSize      int  `json:"page_size"`
	HasNext       bool `json:"has_next"`
	HasPrev       bool `json:"has_prev"`
	TotalItems    int  `json:"total_items"`
	NumberOfPages int  `json:"number_of_pages"`
	Items         []T  `json:"items"`
}

func NewOffsetPage[T any](page, pageSize, totalItems int, items []T) OffsetPage[T] {
	numberOfPages := totalItems / pageSize
	if totalItems%pageSize != 0 {
		numberOfPages++
	}

	hasNext := page < numberOfPages
	hasPrev := page > 1

	return OffsetPage[T]{
		Page:          page,
		PageSize:      pageSize,
		HasNext:       hasNext,
		HasPrev:       hasPrev,
		TotalItems:    totalItems,
		NumberOfPages: numberOfPages,
		Items:         items,
	}
}
