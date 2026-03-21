package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResultRunner_DoExecutesFn(t *testing.T) {
	var r ResultRunner
	called := false
	r.Do(func() error {
		called = true
		return errors.New("boom")
	})
	require.True(t, called)
	require.EqualError(t, r.Error, "boom")
}

func TestResultRunner_DoSkipsOnExistingError(t *testing.T) {
	existing := errors.New("prior error")
	r := ResultRunner{Error: existing}
	counter := 0
	r.Do(func() error {
		counter++
		return nil
	})
	require.Equal(t, 0, counter)
	require.Equal(t, existing, r.Error)
}

func TestResultRunner_DoChain(t *testing.T) {
	var r ResultRunner
	sentinel := errors.New("second failed")

	r.Do(func() error { return nil })
	r.Do(func() error { return sentinel })
	r.Do(func() error {
		t.Fatal("third Do should have been skipped")
		return nil
	})

	require.Equal(t, sentinel, r.Error)
}

func TestResultRunner_DoNilError(t *testing.T) {
	var r ResultRunner
	r.Do(func() error { return nil })
	require.NoError(t, r.Error)
}

func TestResultRunnerWithParam_DoPassesParam(t *testing.T) {
	var r ResultRunnerWithParam[string]
	received := ""
	r.Do("hello", func(p string) error {
		received = p
		return errors.New("oops")
	})
	require.Equal(t, "hello", received)
	require.EqualError(t, r.Error, "oops")
}

func TestResultRunnerWithParam_DoSkipsOnExistingError(t *testing.T) {
	existing := errors.New("already failed")
	r := ResultRunnerWithParam[int]{Error: existing}
	counter := 0
	r.Do(42, func(p int) error {
		counter++
		return nil
	})
	require.Equal(t, 0, counter)
	require.Equal(t, existing, r.Error)
}

func TestResultRunnerWithParam_DoNilError(t *testing.T) {
	var r ResultRunnerWithParam[int]
	r.Do(7, func(p int) error { return nil })
	require.NoError(t, r.Error)
}
