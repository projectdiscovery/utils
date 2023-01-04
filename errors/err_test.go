package errorutil_test

import (
	"fmt"
	"strings"
	"testing"

	errors "github.com/projectdiscovery/utils/errors"
)

func TestErrorEqual(t *testing.T) {
	err1 := fmt.Errorf("error init x")
	err2 := errors.NewWithErr(err1)
	err3 := errors.NewWithTag("testing", "error init")
	var errnil error

	if !errors.IsAny(err1, err2, errnil) {
		t.Errorf("expected errors to be equal")
	}
	if errors.IsAny(err1, err3, errnil) {
		t.Errorf("expected error to be not equal")
	}
}

func TestWrapWithNil(t *testing.T) {
	err1 := errors.NewWithTag("niltest", "non nil error").WithLevel(errors.Fatal)
	var errx error

	if errors.WrapwithNil(errx, err1) != nil {
		t.Errorf("when base error is nil ")
	}
}

func TestStackTrace(t *testing.T) {
	err := errors.New("base error")
	relay := func(err error) error {
		return err
	}
	errx := relay(err)

	t.Run("teststack", func(t *testing.T) {
		if strings.Contains(errx.Error(), "captureStack") {
			t.Errorf("stacktrace should be disabled by default")
		}
		errors.ShowStackTrace = true
		if !strings.Contains(errx.Error(), "captureStack") {
			t.Errorf("missing stacktrace got %v", errx.Error())
		}
	})
}

func TestErrorCallback(t *testing.T) {
	callbackExecuted := false

	err := errors.NewWithTag("callback", "got error").WithCallback(func(level errors.ErrorLevel, err string, tags ...string) {
		if level != errors.Runtime {
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
