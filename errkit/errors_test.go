package errkit

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/multierr"

	stderrors "errors"
	"log/slog"
)

// what are these tests ?
// Below tests check for interoperability of this package with other error packages
// like pkg/errors and go.uber.org/multierr and std errors as well

func TestErrorAs(t *testing.T) {
	// Create a new error with a specific class and wrap it
	x := New("this is a nuclei error").SetKind(ErrKindNetworkPermanent).Build()
	y := errors.Wrap(x, "this is a wrap error")

	// Attempt to unwrap the error to a specific type
	ne := &ErrorX{}
	if !errors.As(y, &ne) {
		t.Fatal("expected to be able to unwrap")
	}

	// Wrap the specific error type into another error and try unwrapping again
	wrapped := Wrap(ne, "this is a wrapped error")
	if !errors.As(wrapped, &ne) {
		t.Fatal("expected to be able to unwrap")
	}

	// Combine multiple errors into a multierror and attempt to unwrap to the specific type
	errs := []error{
		stderrors.New("this is a std error"),
		x,
		errors.New("this is a pkg error"),
	}
	multi := multierr.Combine(errs...)
	if !errors.As(multi, &ne) {
		t.Fatal("expected to be able to unwrap")
	}
}

func TestErrorIs(t *testing.T) {
	// Create a new error, wrap it, and check if the original error can be found
	x := New("this is a nuclei error").SetKind(ErrKindNetworkPermanent).Build()
	y := errors.Wrap(x, "this is a wrap error")
	if !errors.Is(y, x) {
		t.Fatal("expected to be able to find the original error")
	}

	// Wrap the original error with a custom wrapper and check again
	wrapped := Wrap(x, "this is a wrapped error")
	if !stderrors.Is(wrapped, x) {
		t.Fatal("expected to be able to find the original error")
	}

	// Combine multiple errors into a multierror and check if the original error can be found
	errs := []error{
		stderrors.New("this is a std error"),
		x,
		errors.New("this is a pkg error"),
	}
	multi := multierr.Combine(errs...)
	if !errors.Is(multi, x) {
		t.Fatal("expected to be able to find the original error")
	}
}

// TestErrorUtil commented out to avoid circular dependency with errorutil
// func TestErrorUtil(t *testing.T) {
//	utilErr := errorutil.New("got err while executing http://206.189.19.240:8000/wp-content/plugins/wp-automatic/inc/csv.php <- POST http://206.189.19.240:8000/wp-content/plugins/wp-automatic/inc/csv.php giving up after 2 attempts: Post \"http://206.189.19.240:8000/wp-content/plugins/wp-automatic/inc/csv.php\": [:RUNTIME] ztls fallback failed <- dial tcp 206.189.19.240:8000: connect: connection refused")
//	x := ErrorX{}
//	parseError(&x, utilErr)
//	if len(x.errs) != 3 {
//		t.Fatal("expected 3 errors")
//	}
// }

func TestErrKindCheck(t *testing.T) {
	x := New("port closed or filtered").SetKind(ErrKindNetworkPermanent)
	t.Run("Errkind With Normal Error", func(t *testing.T) {
		wrapped := Wrap(x, "this is a wrapped error")
		if !IsKind(wrapped, ErrKindNetworkPermanent) {
			t.Fatal("expected to be able to find the original error")
		}
	})

	// mix of multiple kinds
	tmp := New("i/o timeout").SetKind(ErrKindNetworkTemporary)
	t.Run("Errkind With Multiple Kinds", func(t *testing.T) {
		wrapped := Append(x, tmp)
		errx := FromError(wrapped)
		val, ok := errx.kind.(*multiKind)
		require.True(t, ok, "expected to be able to find the original error")
		require.Equal(t, 2, len(val.kinds))
	})

	// duplicate kinds
	t.Run("Errkind With Duplicate Kinds", func(t *testing.T) {
		wrapped := Append(x, x)
		errx := FromError(wrapped)
		require.True(t, errx.kind.Is(ErrKindNetworkPermanent), "expected to be able to find the original error")
	})
}

func TestMarshalError(t *testing.T) {
	x := New("port closed or filtered").SetKind(ErrKindNetworkPermanent)
	wrapped := Wrap(x, "this is a wrapped error")
	marshalled, err := json.Marshal(wrapped)
	require.NoError(t, err, "expected to be able to marshal the error")
	require.Equal(t, `{"errors":["port closed or filtered","this is a wrapped error"],"kind":"network-permanent-error"}`, string(marshalled))
}

func TestErrorString(t *testing.T) {
	var x error = New("i/o timeout")
	x = With(x, "ip", "10.0.0.1", "port", 80)
	x = WithMessage(x, "tcp dial error")
	x = Append(x, errors.New("some other error"))

	require.Equal(t,
		`cause="i/o timeout" ip=10.0.0.1 port=80 chain="tcp dial error; some other error"`,
		x.Error(),
	)
}

func TestWithAttr(t *testing.T) {
	// Test WithAttr function
	originalErr := New("connection failed")

	// Add attributes using WithAttr
	err := WithAttr(originalErr, slog.String("resource", "database"), slog.Int("port", 5432))

	// Verify the error can be unwrapped to ErrorX
	var errx *ErrorX
	require.True(t, errors.As(err, &errx), "expected to be able to unwrap to ErrorX")

	// Check that attributes were added
	attrs := errx.Attrs()
	require.Len(t, attrs, 2, "expected 2 attributes")

	// Verify specific attributes
	foundResource := false
	foundPort := false
	for _, attr := range attrs {
		if attr.Key == "resource" && attr.Value.String() == "database" {
			foundResource = true
		}
		if attr.Key == "port" && attr.Value.Int64() == 5432 {
			foundPort = true
		}
	}
	require.True(t, foundResource, "expected to find resource attribute")
	require.True(t, foundPort, "expected to find port attribute")

	// Test that WithAttr works with nil error
	nilErr := WithAttr(nil, slog.String("test", "value"))
	require.Nil(t, nilErr, "expected nil error to return nil")

	// Test that WithAttr works with empty attrs
	emptyErr := WithAttr(originalErr)
	require.Equal(t, originalErr, emptyErr, "expected original error when no attrs provided")
}

// The tests below cover https://github.com/projectdiscovery/utils/issues/823:
// wrapping an *ErrorX must copy its attributes, never write into them, so a
// package-level sentinel stays unchanged no matter how often it is wrapped.

func TestWrappingDoesNotMutateFirstError(t *testing.T) {
	t.Parallel()

	wrappers := map[string]func(first, second error) error{
		"Append":  func(first, second error) error { return Append(first, second) },
		"Join":    func(first, second error) error { return Join(first, second) },
		"Combine": func(first, second error) error { return Combine(first, second) },
		"Wrap":    func(first, second error) error { return Append(Wrap(first, "wrapped"), second) },
	}
	for name, wrap := range wrappers {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			sentinel := New("unexpected error", "component", "git")
			inner := With(New("git show failed"), "stderr", "fatal: path does not exist")

			err := wrap(sentinel, inner)

			require.Equal(t, `cause="unexpected error" component=git`, sentinel.Error())
			require.Equal(t,
				[]string{"component=git", "stderr=fatal: path does not exist"},
				attrStrings(err),
			)
		})
	}
}

func TestAppendSameSentinelRepeatedly(t *testing.T) {
	t.Parallel()

	sentinel := New("unexpected error")
	for i := range 10 {
		err := Append(sentinel, With(New("git show failed"), "stdout", "", "stderr", "fatal"))
		err = Append(sentinel, err)
		require.Lenf(t, GetAttr(err), 2, "attributes after call %d", i+1)
	}

	require.Empty(t, GetAttr(sentinel))
}

// slog.Record stores five attributes inline and the rest in a slice. With()
// leaves that slice with spare capacity here, so a shallow copy of the
// sentinel's record would let both results write to the same array slot.
func TestWrappedErrorsDoNotShareAttributeStorage(t *testing.T) {
	t.Parallel()

	sentinel := With(
		New("unexpected error", "a1", 1, "a2", 2, "a3", 3, "a4", 4, "a5", 5, "a6", 6),
		"a7", 7, "a8", 8,
	)

	first := Append(sentinel, With(New("first"), "k", "one"))
	second := Append(sentinel, With(New("second"), "k", "two"))

	require.Equal(t, "one", GetAttrValue(first, "k").String())
	require.Equal(t, "two", GetAttrValue(second, "k").String())
	require.Len(t, GetAttr(sentinel), 8)
}

// Run with -race: Append must only read the sentinel, so wrapping it from
// several goroutines needs no locking.
func TestAppendSentinelConcurrently(t *testing.T) {
	t.Parallel()

	sentinel := New("unexpected error")

	var workers sync.WaitGroup

	for range 8 {
		workers.Go(func() {
			_ = Append(sentinel, With(New("x"), "k", "v"))
		})
	}

	workers.Wait()
	require.Empty(t, GetAttr(sentinel))
}

// The tests below cover https://github.com/projectdiscovery/utils/issues/826:
// only the setters may modify an *ErrorX, so reading a shared sentinel needs
// no locking and leaves it unchanged.

// Run with -race.
func TestReadMethodsConcurrently(t *testing.T) {
	t.Parallel()

	sentinel := New("unexpected error", "component", "git")

	var readers sync.WaitGroup

	for range 8 {
		readers.Go(func() {
			_ = sentinel.Kind()
			_ = sentinel.Error()
			_ = sentinel.Attrs()
			_ = sentinel.Is(sentinel)
			_, _ = json.Marshal(sentinel)
		})
	}

	readers.Wait()
	require.Equal(t, ErrKindUnknown, sentinel.Kind())
}

// Kind must not store the ErrKindUnknown it returns for an error without a
// kind, or every error that wraps the sentinel afterwards inherits it.
func TestKindDoesNotSetKind(t *testing.T) {
	t.Parallel()

	sentinel := New("unexpected error")
	before := GetErrorKind(Wrap(sentinel, "wrapped")).String()

	require.Equal(t, ErrKindUnknown, sentinel.Kind())
	require.Equal(t, before, GetErrorKind(Wrap(sentinel, "wrapped")).String())
}

// The tests below cover https://github.com/projectdiscovery/utils/issues/827:
// an errkit error without a kind is classified from its message, the same way
// a plain error is.

const noSuchHostMsg = "dial tcp: lookup example.invalid: no such host"

var errNoSuchHost = stderrors.New(noSuchHostMsg)

func TestKindDetectionOnErrkitErrors(t *testing.T) {
	t.Parallel()

	errs := map[string]error{
		"errors.New":            errNoSuchHost,
		"New":                   New(noSuchHostMsg),
		"Wrap of a plain error": Wrap(errNoSuchHost, "request failed"),
		"Wrap of New":           Wrap(New(noSuchHostMsg), "request failed"),
		"fmt.Errorf of New":     fmt.Errorf("request failed: %w", New(noSuchHostMsg)),
	}
	for name, err := range errs {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			require.True(t, IsKind(err, ErrKindNetworkPermanent))
			require.Equal(t, ErrKindNetworkPermanent.String(), GetErrorKind(err).String())
		})
	}
}

func TestErrorWithoutKindReportsUnknown(t *testing.T) {
	t.Parallel()

	err := Wrap(New("unexpected error"), "wrapped")

	require.Equal(t, ErrKindUnknown.String(), GetErrorKind(err).String())

	marshalled, jsonErr := json.Marshal(err)
	require.NoError(t, jsonErr)
	require.JSONEq(t, `{"errors":["unexpected error","wrapped"],"kind":"unknown-error"}`, string(marshalled))
}

// Detection only fills in a missing kind; it never replaces one set with
// SetKind.
func TestSetKindOverridesDetection(t *testing.T) {
	t.Parallel()

	err := Wrap(New(noSuchHostMsg).SetKind(ErrKindNetworkTemporary), "request failed")

	require.False(t, IsKind(err, ErrKindNetworkPermanent))
	require.Equal(t, ErrKindNetworkTemporary.String(), GetErrorKind(err).String())
}

func attrStrings(err error) []string {
	attrs := GetAttr(err)

	out := make([]string, 0, len(attrs))
	for _, attr := range attrs {
		out = append(out, attr.String())
	}

	return out
}
