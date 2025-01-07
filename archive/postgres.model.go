package common

// type DbOperator interface {
// 	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
// 	QueryRow(ctx context.Context, sql string, arguments ...interface{}) pgx.Row
// 	Query(ctx context.Context, sql string, arguments ...interface{}) (pgx.Rows, error)
// 	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
// }

// type paginationParamsKey struct{}

// type DbOperatorKey struct{}

// type TransactionFunc func(ctx context.Context, tx pgx.Tx) error

// type PaginatedResponse[T any] struct {
// 	PageSize       int    `json:"pageSize"`
// 	EndCursor      string `json:"endCursor"`
// 	PreviousCursor string `json:"previousCursor"`
// 	HasNext        bool   `json:"hasNext"`
// 	HasPrevious    bool   `json:"hasPrevious"`
// 	ItemsLength    int    `json:"itemsLength"`
// 	Items          []T    `json:"items"`
// }

// type Cursorable interface {
// 	GetCursorValue() []string
// }

// func CreatePaginatedResponse[T any](
// 	pageSize int,
// 	endCursor,
// 	previousCursor Cursorable,
// 	items []T,
// ) PaginatedResponse[T] {
// 	endCursorValue := createCursorValue(endCursor)
// 	previousCursorValue := createCursorValue(previousCursor)
// 	return PaginatedResponse[T]{
// 		PageSize:       pageSize,
// 		EndCursor:      endCursorValue,
// 		PreviousCursor: previousCursorValue,
// 		HasNext:        len(items) >= pageSize,
// 		HasPrevious:    len(items) >= pageSize && previousCursorValue != "",
// 		ItemsLength:    len(items),
// 		Items:          items,
// 	}
// }

// func CreateEmptyPaginatedResponse[T any](pageSize int) PaginatedResponse[T] {
// 	return PaginatedResponse[T]{
// 		PageSize:    pageSize,
// 		HasNext:     false,
// 		ItemsLength: 0,
// 		Items:       []T{},
// 	}
// }

// type paginationQueryBuilder struct {
// 	query *PaginationQuery
// }

// type PaginationQuery struct {
// 	BaseSql               string
// 	Conditions            []string
// 	Direction             int
// 	CursorValue           string
// 	CursorKeys            []string
// 	OrderByQuery          []string
// 	PageSize              int
// 	RefreshCompareSymbol  string
// 	ForwardCompareSymbol  string
// 	BackwardCompareSymbol string
// 	Op                    DbOperator
// }

// type PaginationParams struct {
// 	PageSize  int
// 	Cursor    string
// 	Direction int
// }

// func NewPaginationQueryBuilder(baseSql string, orderByQuery []string) *paginationQueryBuilder {
// 	return &paginationQueryBuilder{
// 		query: &PaginationQuery{
// 			BaseSql:               baseSql,
// 			OrderByQuery:          orderByQuery,
// 			PageSize:              10,
// 			Direction:             1,
// 			ForwardCompareSymbol:  ">",
// 			BackwardCompareSymbol: "<",
// 			RefreshCompareSymbol:  ">=",
// 		},
// 	}
// }

// func (b *paginationQueryBuilder) WithOperator(op DbOperator) *paginationQueryBuilder {
// 	b.query.Op = op
// 	return b
// }

// func (b *paginationQueryBuilder) WithParams(params PaginationParams) *paginationQueryBuilder {
// 	return b.WithCursor(params.Cursor).WithDirection(params.Direction).WithPageSize(params.PageSize)
// }

// func (b *paginationQueryBuilder) WithDirection(direction int) *paginationQueryBuilder {
// 	b.query.Direction = direction
// 	return b
// }

// func (b *paginationQueryBuilder) WithCursor(cursor string) *paginationQueryBuilder {
// 	b.query.CursorValue = cursor
// 	return b
// }

// func (b *paginationQueryBuilder) WithCursorKeys(cursorKeys []string) *paginationQueryBuilder {
// 	b.query.CursorKeys = cursorKeys
// 	return b
// }

// func (b *paginationQueryBuilder) WithConditions(conditions []string) *paginationQueryBuilder {
// 	b.query.Conditions = conditions
// 	return b
// }

// func (b *paginationQueryBuilder) WithPageSize(pageSize int) *paginationQueryBuilder {
// 	b.query.PageSize = pageSize
// 	return b
// }

// func (b *paginationQueryBuilder) WithCompareSymbols(forward, refresh, backward string) *paginationQueryBuilder {
// 	b.query.ForwardCompareSymbol = forward
// 	b.query.RefreshCompareSymbol = refresh
// 	b.query.BackwardCompareSymbol = backward
// 	return b
// }

// func (b *paginationQueryBuilder) Build() PaginationQuery {
// 	if b.query == nil {
// 		panic("invalid pagination query")
// 	}
// 	if b.query.BaseSql == "" || b.query.PageSize <= 0 ||
// 		b.query.CursorKeys == nil || b.query.OrderByQuery == nil ||
// 		len(b.query.CursorKeys) != len(b.query.OrderByQuery) ||
// 		len(b.query.OrderByQuery) == 0 || len(b.query.CursorKeys) == 0 {
// 		panic("invalid pagination query")
// 	}
// 	return *b.query
// }

// func (q PaginationQuery) GetCurrentCursor() []string {
// 	decoded, _ := base64.StdEncoding.DecodeString(q.CursorValue)
// 	return strings.Split(string(decoded), ",")
// }

// func (q *PaginationQuery) getFinalArgAndJoinedConditions() (int, string) {
// 	finalArgIndex := 1
// 	for _, condition := range q.Conditions {
// 		if condition != "" &&
// 			condition != " " &&
// 			condition != "AND" &&
// 			condition != "OR" &&
// 			condition != "and" &&
// 			condition != "or" {
// 			finalArgIndex++
// 		}
// 	}
// 	var joinedConditions string
// 	if finalArgIndex != 1 {
// 		joinedConditions = strings.Join(q.Conditions, " ")
// 	}
// 	return finalArgIndex, joinedConditions
// }

// func (q *PaginationQuery) getFormatedPaginationConditionQuery(finalArgIndex int) string {
// 	unformattedPaginationConditonWithCursor := "AND %s %s %s"
// 	if q.CursorValue == "" {
// 		return ""
// 	}
// 	cursorKeys := q.getCursorKeysJoined()
// 	argsForCursors := q.getArgsForCursorsJoined(finalArgIndex)
// 	return fmt.Sprintf(
// 		unformattedPaginationConditonWithCursor,
// 		cursorKeys,
// 		q.getDirectionString(),
// 		argsForCursors,
// 	)
// }

// func (q *PaginationQuery) getDirectionString() string {
// 	if q.Direction == 0 {
// 		return q.RefreshCompareSymbol
// 	} else if q.Direction > 0 {
// 		return q.ForwardCompareSymbol
// 	}
// 	return q.BackwardCompareSymbol
// }

// func (q *PaginationQuery) getFormattedOrderByQuery() string {
// 	return strings.Join(q.OrderByQuery, ", ")
// }

// func (q *PaginationQuery) getCursorKeysJoined() string {
// 	return "(" + strings.Join(q.CursorKeys, ", ") + " )"
// }

// func (q *PaginationQuery) getArgsForCursorsJoined(finalArgIndex int) string {
// 	return "(" + strings.Join(q.getArgsForCursors(finalArgIndex), ", ") + " )"
// }

// func (q *PaginationQuery) getArgsForCursors(finalArgIndex int) []string {
// 	args := make([]string, 0)
// 	for i := 0; i < len(q.CursorKeys); i++ {
// 		args = append(args, fmt.Sprintf("$%d", finalArgIndex+i))
// 	}
// 	return args
// }

// func (sql PaginationQuery) CreatePaginationQuery() string {
// 	unformattedSql := sql.BaseSql + " " + "WHERE %s %s ORDER BY %s LIMIT %s;"
// 	finalArgIndex, joinedConditions := sql.getFinalArgAndJoinedConditions()
// 	formattedPaginationCondition := sql.getFormatedPaginationConditionQuery(finalArgIndex)
// 	formattedOrderByQuery := sql.getFormattedOrderByQuery()
// 	formattedSql := fmt.Sprintf(
// 		unformattedSql,
// 		joinedConditions,
// 		formattedPaginationCondition,
// 		formattedOrderByQuery,
// 		strconv.Itoa(sql.PageSize),
// 	)
// 	trimmedSql := strings.ReplaceAll(formattedSql, "\n", " ")
// 	trimmedSql = strings.ReplaceAll(trimmedSql, "\t", "")
// 	trimmedSql = strings.ReplaceAll(trimmedSql, "  ", " ")
// 	return trimmedSql
// }

// func (q PaginationQuery) Query(ctx context.Context, arguments ...interface{}) (pgx.Rows, error) {
// 	sql := q.CreatePaginationQuery()
// 	GetLogger().Debug("pagination query", zap.String("sql", sql))
// 	if q.CursorValue == "" {
// 		return q.Op.Query(ctx, sql, arguments...)
// 	}
// 	cursors := q.GetCurrentCursor()
// 	args := make([]interface{}, 0)
// 	args = append(args, arguments...)
// 	for _, cursor := range cursors {
// 		args = append(args, cursor)
// 	}
// 	return q.Op.Query(ctx, sql, args...)
// }
