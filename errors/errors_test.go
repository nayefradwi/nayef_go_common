package errors

import (
	"testing"
)

func TestErrorDetails(t *testing.T) {
	d := ErrorDetails{Message: "too short", Code: "MIN_LENGTH", Field: "name"}
	if d.Message != "too short" || d.Code != "MIN_LENGTH" || d.Field != "name" {
		t.Fatal("ErrorDetails fields not set correctly")
	}
}

func TestField(t *testing.T) {
	d := Field("email", "REQUIRED", "email is required")
	if d.Field != "email" || d.Code != "REQUIRED" || d.Message != "email is required" {
		t.Fatal("Field helper did not set values correctly")
	}
}

func TestNewResultError_NoDetails(t *testing.T) {
	err := NewResultError("something went wrong", "SOME_CODE")
	if err.Message != "something went wrong" {
		t.Fatalf("expected message 'something went wrong', got %q", err.Message)
	}
	if err.Code != "SOME_CODE" {
		t.Fatalf("expected code 'SOME_CODE', got %q", err.Code)
	}
	if len(err.Errors) != 0 {
		t.Fatalf("expected empty errors map, got %d entries", len(err.Errors))
	}
}

func TestNewResultError_WithDetails(t *testing.T) {
	err := NewResultError("validation failed", CodeValidation,
		Field("email", "REQUIRED", "email is required"),
		Field("email", "FORMAT", "must be valid email"),
		Field("name", "MIN_LENGTH", "too short"),
	)
	if len(err.Errors) != 2 {
		t.Fatalf("expected 2 field keys, got %d", len(err.Errors))
	}
	if len(err.Errors["email"]) != 2 {
		t.Fatalf("expected 2 errors for email, got %d", len(err.Errors["email"]))
	}
	if len(err.Errors["name"]) != 1 {
		t.Fatalf("expected 1 error for name, got %d", len(err.Errors["name"]))
	}
}

func TestResultError_Error(t *testing.T) {
	err := NewResultError("not found", CodeNotFound)
	if err.Error() != "not found" {
		t.Fatalf("Error() returned %q, want 'not found'", err.Error())
	}
}

func TestResultError_ImplementsError(t *testing.T) {
	var _ error = NewResultError("test", CodeInternal)
}

func TestWithCode(t *testing.T) {
	original := NewResultError("msg", CodeBadRequest)
	changed := original.WithCode(CodeForbidden)
	if changed.Code != CodeForbidden {
		t.Fatalf("WithCode returned code %q, want %q", changed.Code, CodeForbidden)
	}
	if original.Code != CodeBadRequest {
		t.Fatal("WithCode mutated the original")
	}
}

func TestWithErrors(t *testing.T) {
	original := NewResultError("bad request", CodeBadRequest)
	changed := original.WithErrors(
		Field("age", "MIN", "must be positive"),
		Field("age", "MAX", "must be under 200"),
		Field("name", "REQUIRED", "name is required"),
	)
	if len(changed.Errors) != 2 {
		t.Fatalf("expected 2 field keys, got %d", len(changed.Errors))
	}
	if len(changed.Errors["age"]) != 2 {
		t.Fatalf("expected 2 errors for age, got %d", len(changed.Errors["age"]))
	}
	if len(changed.Errors["name"]) != 1 {
		t.Fatalf("expected 1 error for name, got %d", len(changed.Errors["name"]))
	}
	if len(original.Errors) != 0 {
		t.Fatal("WithErrors mutated the original")
	}
}

func TestWithErrors_Empty(t *testing.T) {
	err := NewResultError("msg", CodeBadRequest).WithErrors()
	if len(err.Errors) != 0 {
		t.Fatalf("expected empty errors map, got %d entries", len(err.Errors))
	}
}

func TestFactoryBadRequest(t *testing.T) {
	err := BadRequestError("bad")
	assertFactory(t, err, "bad", CodeBadRequest)
}

func TestFactoryUnauthorized(t *testing.T) {
	err := UnauthorizedError("no auth")
	assertFactory(t, err, "no auth", CodeUnauthorized)
}

func TestFactoryForbidden(t *testing.T) {
	err := ForbiddenError("denied")
	assertFactory(t, err, "denied", CodeForbidden)
}

func TestFactoryNotFound(t *testing.T) {
	err := NotFoundError("missing")
	assertFactory(t, err, "missing", CodeNotFound)
}

func TestFactoryInternal(t *testing.T) {
	err := InternalError("broke")
	assertFactory(t, err, "broke", CodeInternal)
}

func TestFactoryInvalidInput(t *testing.T) {
	err := InvalidInputError("wrong")
	assertFactory(t, err, "wrong", CodeInvalidInput)
}

func TestFactoryUnknown(t *testing.T) {
	err := UnknownError("???")
	assertFactory(t, err, "???", CodeUnknown)
}

func TestNewValidationError(t *testing.T) {
	err := NewValidationError(
		Field("email", "REQUIRED", "email is required"),
		Field("name", "REQUIRED", "name is required"),
	)
	if err.Message != "Invalid" {
		t.Fatalf("expected message 'Invalid', got %q", err.Message)
	}
	if err.Code != CodeValidation {
		t.Fatalf("expected code %q, got %q", CodeValidation, err.Code)
	}
	if len(err.Errors) != 2 {
		t.Fatalf("expected 2 field keys, got %d", len(err.Errors))
	}
}

func TestNewValidationError_Empty(t *testing.T) {
	err := NewValidationError()
	if err.Code != CodeValidation {
		t.Fatalf("expected code %q, got %q", CodeValidation, err.Code)
	}
	if len(err.Errors) != 0 {
		t.Fatalf("expected empty errors map, got %d entries", len(err.Errors))
	}
}

func assertFactory(t *testing.T, err *ResultError, expectedMsg, expectedCode string) {
	t.Helper()
	if err.Message != expectedMsg {
		t.Fatalf("expected message %q, got %q", expectedMsg, err.Message)
	}
	if err.Code != expectedCode {
		t.Fatalf("expected code %q, got %q", expectedCode, err.Code)
	}
	if len(err.Errors) != 0 {
		t.Fatalf("expected no errors, got %d", len(err.Errors))
	}
}
