package postgres

import (
	"errors"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nayefradwi/nayef_go_common/core"
	"go.uber.org/zap"
)

const (
	UniqueViolationErr           = "23505"
	ForeignKeyViolationErr       = "23503"
	NotNullViolationErr          = "23502"
	CheckViolationErr            = "23514"
	ExclusionViolationErr        = "23P01"
	InvalidTextRepresentationErr = "22P02"
	NumericValueOutOfRangeErr    = "22003"
	StringDataRightTruncationErr = "22001"
	DivisionByZeroErr            = "22012"
	InvalidForeignKeyErr         = "42830"
	DeadlockDetectedErr          = "40P01"
	SerializationFailureErr      = "40001"
	LockNotAvailableErr          = "55P03"
	InsufficientPrivilegeErr     = "42501"
	InvalidColumnReferenceErr    = "42P10"
	UndefinedColumnErr           = "42703"
	UndefinedTableErr            = "42P01"
	SyntaxErrorErr               = "42601"
	TransactionRollbackErr       = "40000"
	InvalidDatetimeFormatErr     = "22007"
	InvalidBooleanValueErr       = "22023"
)

func MapPgError(err error, message string) error {
	if err == nil {
		return nil
	}

	mappedErr := mapPgError(err, message)
	zap.L().Error("Postgres error:", zap.Error(err))
	return mappedErr
}

func mapPgError(err error, message string) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return core.NotFoundError(message)
	}

	if pgErr, ok := err.(*pgconn.PgError); ok {
		switch pgErr.Code {
		case UniqueViolationErr:
			return core.NewResultError(message, "DUPLICATE")
		case ForeignKeyViolationErr:
			return core.NewResultError(message, "FOREIGN_KEY_VIOLATION")
		case NotNullViolationErr:
			return core.NewResultError(message, "NOT_NULL_VIOLATION")
		case CheckViolationErr:
			return core.NewResultError(message, "CHECK_VIOLATION")
		case ExclusionViolationErr:
			return core.NewResultError(message, "EXCLUSION_VIOLATION")
		case InvalidTextRepresentationErr:
			return core.NewResultError(message, "INVALID_INPUT")
		case NumericValueOutOfRangeErr:
			return core.NewResultError(message, "VALUE_OUT_OF_RANGE")
		case StringDataRightTruncationErr:
			return core.NewResultError(message, "STRING_TRUNCATION")
		case DivisionByZeroErr:
			return core.NewResultError(message, "DIVISION_BY_ZERO")
		case InvalidForeignKeyErr:
			return core.NewResultError(message, "INVALID_FOREIGN_KEY")
		case DeadlockDetectedErr:
			return core.NewResultError(message, "DEADLOCK")
		}
	}

	return core.InternalError(message)
}
