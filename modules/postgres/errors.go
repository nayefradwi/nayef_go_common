package postgres

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nayefradwi/nayef_go_common/result"
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
	return mappedErr
}

func mapPgError(err error, message string) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return result.NotFoundError(message)
	}

	if pgErr, ok := err.(*pgconn.PgError); ok {
		switch pgErr.Code {
		case UniqueViolationErr:
			return result.NewResultError(message, "DUPLICATE")
		case ForeignKeyViolationErr:
			return result.NewResultError(message, "FOREIGN_KEY_VIOLATION")
		case NotNullViolationErr:
			return result.NewResultError(message, "NOT_NULL_VIOLATION")
		case CheckViolationErr:
			return result.NewResultError(message, "CHECK_VIOLATION")
		case ExclusionViolationErr:
			return result.NewResultError(message, "EXCLUSION_VIOLATION")
		case InvalidTextRepresentationErr:
			return result.NewResultError(message, "INVALID_INPUT")
		case NumericValueOutOfRangeErr:
			return result.NewResultError(message, "VALUE_OUT_OF_RANGE")
		case StringDataRightTruncationErr:
			return result.NewResultError(message, "STRING_TRUNCATION")
		case DivisionByZeroErr:
			return result.NewResultError(message, "DIVISION_BY_ZERO")
		case InvalidForeignKeyErr:
			return result.NewResultError(message, "INVALID_FOREIGN_KEY")
		case DeadlockDetectedErr:
			return result.NewResultError(message, "DEADLOCK")
		}
	}

	return result.InternalError(message)
}
