package errkit

import (
	"encoding/json"
	"testing"

	"github.com/pkg/errors"
	errorutil "github.com/projectdiscovery/utils/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/multierr"

	stderrors "errors"
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

func TestErrorUtil(t *testing.T) {
	utilErr := errorutil.New("got err while executing http://206.189.19.240:8000/wp-content/plugins/wp-automatic/inc/csv.php <- POST http://206.189.19.240:8000/wp-content/plugins/wp-automatic/inc/csv.php giving up after 2 attempts: Post \"http://206.189.19.240:8000/wp-content/plugins/wp-automatic/inc/csv.php\": [:RUNTIME] ztls fallback failed <- dial tcp 206.189.19.240:8000: connect: connection refused")
	x := ErrorX{}
	parseError(&x, utilErr)
	if len(x.errs) != 3 {
		t.Fatal("expected 3 errors")
	}
}

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
