package errors_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/projectdiscovery/utils/errors"
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
		if ee, ok := errx.(errors.Error); ok {
			ee.ShowStackTrace()
		}
		if !strings.Contains(errx.Error(), "captureStack") {
			t.Errorf("missing stacktrace got %v", errx.Error())
		}
	})
}
