package errors

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestErrorDetails(t *testing.T) {
	d := ErrorDetails{Message: "too short", Code: "MIN_LENGTH", Field: "name"}
	require.Equal(t, "too short", d.Message)
	require.Equal(t, "MIN_LENGTH", d.Code)
	require.Equal(t, "name", d.Field)
}

func TestField(t *testing.T) {
	d := Field("email", "REQUIRED", "email is required")
	require.Equal(t, "email", d.Field)
	require.Equal(t, "REQUIRED", d.Code)
	require.Equal(t, "email is required", d.Message)
}

func TestNewResultError_NoDetails(t *testing.T) {
	err := NewResultError("something went wrong", "SOME_CODE")
	require.Equal(t, "something went wrong", err.Message)
	require.Equal(t, "SOME_CODE", err.Code)
	require.Len(t, err.Errors, 0)
}

func TestNewResultError_WithDetails(t *testing.T) {
	err := NewResultError("validation failed", CodeValidation,
		Field("email", "REQUIRED", "email is required"),
		Field("email", "FORMAT", "must be valid email"),
		Field("name", "MIN_LENGTH", "too short"),
	)
	require.Len(t, err.Errors, 2)
	require.Len(t, err.Errors["email"], 2)
	require.Len(t, err.Errors["name"], 1)
}

func TestResultError_Error(t *testing.T) {
	err := NewResultError("not found", CodeNotFound)
	require.Equal(t, "not found", err.Error())
}

func TestResultError_ImplementsError(t *testing.T) {
	var _ error = NewResultError("test", CodeInternal)
}

func TestWithCode(t *testing.T) {
	original := NewResultError("msg", CodeBadRequest)
	changed := original.WithCode(CodeForbidden)
	require.Equal(t, CodeForbidden, changed.Code)
	require.Equal(t, CodeBadRequest, original.Code)
}

func TestWithErrors(t *testing.T) {
	original := NewResultError("bad request", CodeBadRequest)
	changed := original.WithErrors(
		Field("age", "MIN", "must be positive"),
		Field("age", "MAX", "must be under 200"),
		Field("name", "REQUIRED", "name is required"),
	)
	require.Len(t, changed.Errors, 2)
	require.Len(t, changed.Errors["age"], 2)
	require.Len(t, changed.Errors["name"], 1)
	require.Len(t, original.Errors, 0)
}

func TestWithErrors_Empty(t *testing.T) {
	err := NewResultError("msg", CodeBadRequest).WithErrors()
	require.Len(t, err.Errors, 0)
}

func TestFactoryBadRequest(t *testing.T) {
	err := BadRequestError("bad")
	assertFactory(t, err, "bad", CodeBadRequest, http.StatusBadRequest)
}

func TestFactoryUnauthorized(t *testing.T) {
	err := UnauthorizedError("no auth")
	assertFactory(t, err, "no auth", CodeUnauthorized, http.StatusUnauthorized)
}

func TestFactoryForbidden(t *testing.T) {
	err := ForbiddenError("denied")
	assertFactory(t, err, "denied", CodeForbidden, http.StatusForbidden)
}

func TestFactoryNotFound(t *testing.T) {
	err := NotFoundError("missing")
	assertFactory(t, err, "missing", CodeNotFound, http.StatusNotFound)
}

func TestFactoryInternal(t *testing.T) {
	err := InternalError("broke")
	assertFactory(t, err, "broke", CodeInternal, http.StatusInternalServerError)
}

func TestFactoryInvalidInput(t *testing.T) {
	err := InvalidInputError("wrong")
	assertFactory(t, err, "wrong", CodeInvalidInput, http.StatusUnprocessableEntity)
}

func TestFactoryUnknown(t *testing.T) {
	err := UnknownError("???")
	assertFactory(t, err, "???", CodeUnknown, http.StatusInternalServerError)
}

func TestWithStatus(t *testing.T) {
	original := NewResultError("msg", CodeBadRequest)
	changed := original.WithStatus(http.StatusTeapot)
	require.Equal(t, http.StatusTeapot, changed.Status)
	require.Equal(t, 0, original.Status)
}

func TestNewResultErrorWithStatus(t *testing.T) {
	err := NewResultErrorWithStatus("nope", CodeForbidden, http.StatusForbidden,
		Field("token", "EXPIRED", "token expired"),
	)
	require.Equal(t, "nope", err.Message)
	require.Equal(t, CodeForbidden, err.Code)
	require.Equal(t, http.StatusForbidden, err.Status)
	require.Len(t, err.Errors["token"], 1)
}

func TestNewResultError_NoStatus(t *testing.T) {
	err := NewResultError("no status", CodeInternal)
	require.Equal(t, 0, err.Status)
}

func TestNewValidationError(t *testing.T) {
	err := NewValidationError(
		Field("email", "REQUIRED", "email is required"),
		Field("name", "REQUIRED", "name is required"),
	)
	require.Equal(t, "Invalid", err.Message)
	require.Equal(t, CodeValidation, err.Code)
	require.Len(t, err.Errors, 2)
}

func TestNewValidationError_Empty(t *testing.T) {
	err := NewValidationError()
	require.Equal(t, CodeValidation, err.Code)
	require.Len(t, err.Errors, 0)
}

func assertFactory(t *testing.T, err *ResultError, expectedMsg, expectedCode string, expectedStatus int) {
	t.Helper()
	require.Equal(t, expectedMsg, err.Message)
	require.Equal(t, expectedCode, err.Code)
	require.Equal(t, expectedStatus, err.Status)
	require.Len(t, err.Errors, 0)
}
