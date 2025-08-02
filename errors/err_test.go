package errorutil_test

import (
	"errors"
	"strings"
	"testing"

	errorutil "github.com/projectdiscovery/utils/errors"
)

type customError struct {
	msg string
}

func (c *customError) Error() string {
	return c.msg
}

func TestErrorEqual(t *testing.T) {
	err1 := errors.New("error init x")
	err2 := errorutil.NewWithErr(err1)
	err3 := errorutil.NewWithTag("testing", "error init")
	var errnil error

	if !errorutil.IsAny(err1, err2, errnil) {
		t.Errorf("expected errors to be equal")
	}
	if errorutil.IsAny(err1, err3, errnil) {
		t.Errorf("expected error to be not equal")
	}
}

func TestWrapWithNil(t *testing.T) {
	err1 := errorutil.NewWithTag("niltest", "non nil error").WithLevel(errorutil.Fatal)
	var errx error

	if errorutil.WrapwithNil(errx, err1) != nil {
		t.Errorf("when base error is nil ")
	}
}


func TestErrorCallback(t *testing.T) {
	callbackExecuted := false

	err := errorutil.NewWithTag("callback", "got error").WithCallback(func(level errorutil.ErrorLevel, err string, tags ...string) {
		if level != errorutil.Runtime {
			t.Errorf("Default error level should be Runtime")
		}
		if tags[0] != "callback" {
			t.Errorf("missing callback")
		}
		callbackExecuted = true
	})

	errval := err.Error()

	if !strings.Contains(errval, "callback") || !strings.Contains(errval, "got error") || !strings.Contains(errval, "RUNTIME") {
		t.Errorf("error content missing expected values `callback,got error and Runtime` in error value but got %v", errval)
	}

	if !callbackExecuted {
		t.Errorf("error callback failed to execute")
	}
}

func TestErrorIs(t *testing.T) {
	var ErrTest = errors.New("test error")

	err := errorutil.NewWithErr(ErrTest).Msgf("message %s", "test")

	if !errors.Is(err, ErrTest) {
		t.Errorf("expected error to match ErrTest")
	}
}

func TestUnwrap(t *testing.T) {
	// Test basic unwrapping
	baseErr := errors.New("base error")
	wrappedErr := errorutil.NewWithErr(baseErr)

	if !errors.Is(wrappedErr, baseErr) {
		t.Errorf("expected wrapped error to match base error")
	}

	// Test unwrapping thru error chain
	middleErr := errorutil.NewWithErr(baseErr).WithTag("middle")
	topErr := errorutil.NewWithErr(middleErr).WithTag("top")

	if !errors.Is(topErr, baseErr) {
		t.Errorf("expected topErr to match baseErr through chain")
	}

	if !errors.Is(topErr, middleErr) {
		t.Errorf("expected topErr to match middleErr")
	}

	// Test direct unwrap method
	if unwrapped := errors.Unwrap(wrappedErr); unwrapped != baseErr {
		t.Errorf("expected direct unwrap to return baseErr, got %v", unwrapped)
	}

	// Test unwrapping with Wrap method
	err1 := errors.New("first error")
	err2 := errors.New("second error")
	combined := errorutil.New("combined error").Wrap(err1, err2)

	if !errors.Is(combined, err1) {
		t.Errorf("expected combined error to match err1")
	}

	// Test errors.As functionality
	customErr := &customError{msg: "custom error"}
	wrappedCustom := errorutil.NewWithErr(customErr).WithTag("wrapped")

	var targetCustom *customError
	if !errors.As(wrappedCustom, &targetCustom) {
		t.Errorf("expected errors.As to find custom error type")
	}

	if targetCustom.msg != "custom error" {
		t.Errorf("expected custom error message 'custom error', got %s", targetCustom.msg)
	}
}
