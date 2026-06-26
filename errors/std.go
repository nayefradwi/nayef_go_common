package errors

import "errors"

func New(text string) error         { return errors.New(text) }
func Is(err, target error) bool     { return errors.Is(err, target) }
func As(err error, target any) bool { return errors.As(err, target) }
func Unwrap(err error) error        { return errors.Unwrap(err) }
func Join(errs ...error) error      { return errors.Join(errs...) }
