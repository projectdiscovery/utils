package errkit

import (
	"encoding/json"
	stderrors "errors"
	"log/slog"
	"reflect"
	"testing"
)

// The tests below cover the "nothing to report" inputs: all-nil arguments,
// an *ErrorX that holds no errors, and foreign errors whose Cause or Unwrap
// yields nothing.

// emptyCauseError is a foreign error whose Cause reports no underlying error.
type emptyCauseError struct{}

func (emptyCauseError) Error() string { return "no cause available" }
func (emptyCauseError) Cause() error  { return nil }

// nilEntriesError is a foreign error whose Unwrap returns only nil entries.
type nilEntriesError struct{}

func (nilEntriesError) Error() string   { return "joined nothing" }
func (nilEntriesError) Unwrap() []error { return []error{nil, nil} }

// messageError is a static error type, so tests can build distinct errors from
// a string without errors.New.
type messageError string

func (m messageError) Error() string { return string(m) }

// requireNilError compares to the literal nil, so it fails on a typed nil
// pointer wrapped in a non-nil error interface.
func requireNilError(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatalf("expected a nil error, got %#v", err)
	}
}

func requireError(t *testing.T, err error) {
	t.Helper()

	if err == nil {
		t.Fatal("expected an error, got nil")
	}
}

func requireEqual(t *testing.T, want, got any) {
	t.Helper()

	if !reflect.DeepEqual(want, got) {
		t.Fatalf("not equal:\nwant: %#v\ngot:  %#v", want, got)
	}
}

func requireEmpty(t *testing.T, what string, length int) {
	t.Helper()

	if length != 0 {
		t.Fatalf("expected %s to be empty, it has %d entries", what, length)
	}
}

// requireErrorText checks the message of a non-nil error.
func requireErrorText(t *testing.T, err error, want string) {
	t.Helper()

	requireError(t, err)

	if got := err.Error(); got != want {
		t.Fatalf("error text:\nwant: %s\ngot:  %s", want, got)
	}
}

func requireNotPanics(t *testing.T, body func()) {
	t.Helper()

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("unexpected panic: %v", r)
		}
	}()

	body()
}

func aggregators() map[string]func(...error) error {
	return map[string]func(...error) error{
		"Append":  Append,
		"Join":    Join,
		"Combine": Combine,
	}
}

// Aggregating nil errors must give a literal nil, not a non-nil error that
// holds nothing.
func TestAggregatorsAllNilReturnNil(t *testing.T) {
	t.Parallel()

	inputs := map[string][]error{
		"no arguments":        nil,
		"one nil":             {nil},
		"two nils":            {nil, nil},
		"nil *ErrorX":         {(*ErrorX)(nil)},
		"nil and nil *ErrorX": {nil, (*ErrorX)(nil), nil},
	}

	for name, aggregate := range aggregators() {
		for inputName, input := range inputs {
			t.Run(name+"/"+inputName, func(t *testing.T) {
				t.Parallel()

				requireNilError(t, aggregate(input...))
			})
		}
	}
}

func TestAggregatorsSkipNilInputs(t *testing.T) {
	t.Parallel()

	for name, aggregate := range aggregators() {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := aggregate(nil, messageError("first"), (*ErrorX)(nil), messageError("second"), nil)

			requireErrorText(t, err, `cause="first" chain="second"`)
		})
	}
}

// An *ErrorX holding no errors comes from the zero value, from appending one,
// or from a nil pointer (FromError(nil)). No method may panic on it.
func TestErrorXWithoutErrorsDoesNotPanic(t *testing.T) {
	t.Parallel()

	var appended *ErrorX

	if !stderrors.As(Append(new(ErrorX)), &appended) {
		t.Fatal("appending a non-nil error must return an *ErrorX")
	}

	states := map[string]*ErrorX{
		"zero value":     new(ErrorX),
		"appended empty": appended,
		"nil pointer":    nil,
		"FromError(nil)": FromError(nil),
	}

	for name, errx := range states {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			requireNotPanics(t, func() {
				_ = errx.Error()
				_ = errx.Unwrap()
				_ = errx.Errors()
				_ = errx.Attrs()
				_ = errx.Cause()
				_ = errx.Kind()
				_ = errx.Build()
				_ = errx.Is(messageError("x"))
				_ = errx.Is(nil)
				_ = stderrors.Is(errx, messageError("x"))
				_, _ = json.Marshal(errx)
				_ = errx.SetKind(ErrKindDeadline)
				_ = errx.ResetKind()
				_ = errx.SetAttr(slog.String("k", "v"))
				errx.Msg("m")
				errx.Msgf("%s", "m")
			})
		})
	}
}

func TestErrorXWithoutErrorsAccessors(t *testing.T) {
	t.Parallel()

	states := map[string]*ErrorX{"zero value": new(ErrorX), "nil pointer": nil}

	for name, errx := range states {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			requireEmpty(t, "Errors", len(errx.Errors()))
			requireEmpty(t, "Unwrap", len(errx.Unwrap()))
			requireEmpty(t, "Attrs", len(errx.Attrs()))
			requireNilError(t, errx.Cause())

			if errx.Is(messageError("x")) {
				t.Error("an error holding nothing must not match a target")
			}

			requireEqual(t, ErrKindUnknown, errx.Kind())
		})
	}

	t.Run("Error", func(t *testing.T) {
		t.Parallel()

		requireEqual(t, `cause=""`, new(ErrorX).Error())
		requireEqual(t, "<nil>", (*ErrorX)(nil).Error())
	})

	t.Run("json.Marshal", func(t *testing.T) {
		t.Parallel()

		// json.Marshal sorts the keys of the map MarshalJSON builds.
		out, err := json.Marshal(new(ErrorX))
		if err != nil {
			t.Fatal(err)
		}

		requireEqual(t, `{"errors":[],"kind":"unknown-error"}`, string(out))

		out, err = json.Marshal((*ErrorX)(nil))
		if err != nil {
			t.Fatal(err)
		}

		requireEqual(t, "null", string(out))
	})
}

// Build returns the error interface, so a nil *ErrorX must come back as a
// nil interface rather than a non-nil interface wrapping a nil pointer.
func TestBuildNilReturnsNilInterface(t *testing.T) {
	t.Parallel()

	requireNilError(t, (*ErrorX)(nil).Build())
	requireNilError(t, FromError(nil).Build())
}

// SetKind, ResetKind and SetAttr on a nil receiver are no-ops, so chained
// calls on FromError(nil) stay safe.
func TestNilErrorXMutatorsAreNoOps(t *testing.T) {
	t.Parallel()

	var errx *ErrorX

	for name, got := range map[string]*ErrorX{
		"SetKind":   errx.SetKind(ErrKindDeadline),
		"ResetKind": errx.ResetKind(),
		"SetAttr":   errx.SetAttr(slog.String("k", "v")),
	} {
		if got != nil {
			t.Errorf("%s on a nil receiver returned %#v, want nil", name, got)
		}
	}
}

// The classification and slog helpers take any error, nil included, and must
// not panic. An error with nothing to classify has the unknown kind.
func TestHelpersWithEmptyInput(t *testing.T) {
	t.Parallel()

	empty := new(ErrorX)
	inputs := map[string]error{
		"nil":            nil,
		"empty *ErrorX":  empty,
		"empty appended": Append(empty),
	}

	for name, input := range inputs {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			requireNotPanics(t, func() {
				requireEqual(t, ErrKindUnknown, GetErrorKind(input))
				requireEqual(t, []ErrKind{ErrKindUnknown}, GetAllErrorKinds(input, DefaultErrorKinds...))
				requireEqual(t, false, IsKind(input, DefaultErrorKinds...))
				requireEqual(t, false, IsNetworkTemporaryErr(input))
				requireEqual(t, false, IsNetworkPermanentErr(input))
				requireEqual(t, false, IsDeadlineErr(input))
				requireEmpty(t, "GetAttr", len(GetAttr(input)))
				requireEqual(t, slog.Value{}, GetAttrValue(input, "k"))
				requireEmpty(t, "Errors", len(Errors(input)))
				requireNilError(t, Cause(input))
				_ = ToSlogAttrs(input)
				_ = ToSlogAttrGroup(input)
			})
		})
	}
}

// nilInputs are the two spellings of "no error": the nil interface and the
// nil *ErrorX that FromError(nil) returns once it is stored in an error.
func nilInputs() map[string]error {
	return map[string]error{
		"nil":         nil,
		"nil *ErrorX": FromError(nil),
	}
}

// Wrapping helpers return nil for a nil error and a non-nil error otherwise;
// the aggregators follow the same rule. A nil *ErrorX counts as a nil error.
func TestWrappersNilInNilOut(t *testing.T) {
	t.Parallel()

	wrappers := map[string]func(error) error{
		"Wrap":         func(err error) error { return Wrap(err, "m") },
		"Wrapf":        func(err error) error { return Wrapf(err, "%s", "m") },
		"WithMessage":  func(err error) error { return WithMessage(err, "m") },
		"WithMessagef": func(err error) error { return WithMessagef(err, "%s", "m") },
		"With":         func(err error) error { return With(err, "k", "v") },
		"WithAttr":     func(err error) error { return WithAttr(err, slog.String("k", "v")) },
	}

	for name, wrap := range wrappers {
		for inputName, input := range nilInputs() {
			t.Run(name+"/"+inputName, func(t *testing.T) {
				t.Parallel()

				requireNilError(t, wrap(input))
			})
		}

		t.Run(name+"/non-nil", func(t *testing.T) {
			t.Parallel()

			requireError(t, wrap(messageError("boom")))
		})
	}
}

// The inspecting helpers give the same zero result for a nil *ErrorX as for a
// nil error.
func TestInspectorsTreatNilErrorXAsNil(t *testing.T) {
	t.Parallel()

	target := messageError("target")
	inspectors := map[string]func(error) any{
		"Is":                    func(err error) any { return Is(err, target) },
		"IsKind":                func(err error) any { return IsKind(err, ErrKindNetworkPermanent) },
		"IsNetworkTemporaryErr": func(err error) any { return IsNetworkTemporaryErr(err) },
		"IsNetworkPermanentErr": func(err error) any { return IsNetworkPermanentErr(err) },
		"IsDeadlineErr":         func(err error) any { return IsDeadlineErr(err) },
		"Errors":                func(err error) any { return Errors(err) },
		"Cause":                 func(err error) any { return Cause(err) },
		"GetAttr":               func(err error) any { return GetAttr(err) },
		"GetAttrValue":          func(err error) any { return GetAttrValue(err, "k") },
		"GetErrorKind":          func(err error) any { return GetErrorKind(err) },
		"GetAllErrorKinds":      func(err error) any { return GetAllErrorKinds(err) },
		"ToSlogAttrs":           func(err error) any { return ToSlogAttrs(err) },
	}

	for name, inspect := range inspectors {
		want := inspect(nil)

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			requireEqual(t, want, inspect(FromError(nil)))
		})
	}
}

// nilXCauseError is a foreign error whose Cause is a nil *ErrorX.
type nilXCauseError struct{}

func (nilXCauseError) Error() string { return "cause wrapper" }
func (nilXCauseError) Cause() error  { return FromError(nil) }

// nilXWrapError is a foreign error whose Unwrap is a nil *ErrorX.
type nilXWrapError struct{}

func (nilXWrapError) Error() string { return "wrap wrapper" }
func (nilXWrapError) Unwrap() error { return FromError(nil) }

// timeoutError reports a timeout the way net errors do.
type timeoutError struct{}

func (timeoutError) Error() string { return "i/o timeout" }
func (timeoutError) Timeout() bool { return true }

// A nil *ErrorX nested in another error must be dropped like a nil error, so
// the other errors keep their place and Cause is never a typed nil.
func TestNestedNilErrorXIsDropped(t *testing.T) {
	t.Parallel()

	nilX := FromError(nil)
	first := messageError("x")

	tests := map[string]struct {
		input     error
		wantError string
		wantCause string
	}{
		"joined with an error": {
			input:     stderrors.Join(nilX, first),
			wantError: `cause="x"`,
			wantCause: "x",
		},
		"joined after an error": {
			input:     stderrors.Join(first, nilX),
			wantError: `cause="x"`,
			wantCause: "x",
		},
		"nil Cause of a CauseError": {
			input:     nilXCauseError{},
			wantError: `cause="cause wrapper"`,
			wantCause: "cause wrapper",
		},
		"nil Unwrap of a wrapper": {
			input:     nilXWrapError{},
			wantError: `cause="wrap wrapper"`,
			wantCause: "wrap wrapper",
		},
	}

	for name, testCase := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := Append(testCase.input)

			requireErrorText(t, err, testCase.wantError)

			if got := len(Errors(err)); got != 1 {
				t.Fatalf("Errors has %d entries, want 1", got)
			}

			cause := Cause(err)

			requireNotPanics(t, func() { _ = cause.Error() })
			requireEqual(t, testCase.wantCause, cause.Error())
		})
	}
}

// The nil *ErrorX is dropped without disturbing the errors around it.
func TestNestedNilErrorXKeepsOtherErrors(t *testing.T) {
	t.Parallel()

	nilX := FromError(nil)
	first := messageError("x")

	t.Run("Cause keeps the identity of the other error", func(t *testing.T) {
		t.Parallel()

		requireEqual(t, error(first), Cause(Append(stderrors.Join(nilX, first))))
	})

	t.Run("joined nil *ErrorX alone", func(t *testing.T) {
		t.Parallel()

		// errors.Join keeps a nil *ErrorX, so the joined error is itself
		// non-nil and is kept under its own message.
		err := Append(stderrors.Join(nilX))

		requireError(t, err)

		if _, isErrorX := stderrors.AsType[*ErrorX](Cause(err)); isErrorX {
			t.Fatal("Cause must not be a *ErrorX")
		}
	})

	t.Run("classifiers see the other errors", func(t *testing.T) {
		t.Parallel()

		for name, err := range map[string]error{
			"Join of nil and timeout": Join(nilX, timeoutError{}),
			"Append of errors.Join":   Append(stderrors.Join(nilX, timeoutError{})),
		} {
			requireEqual(t, error(timeoutError{}), Cause(err))

			if !IsNetworkTemporaryErr(err) {
				t.Errorf("%s: want a temporary network error", name)
			}
		}
	})
}

// Foreign errors that report no cause or only nil entries must keep their own
// message instead of becoming an empty ErrorX or an error holding a nil.
func TestParseErrorWithNothingToUnwrap(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input error
		want  string
	}{
		"nil Cause":        {input: emptyCauseError{}, want: `cause="no cause available"`},
		"nil Unwrap items": {input: nilEntriesError{}, want: `cause="joined nothing"`},
	}

	for name, testCase := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var got string

			requireNotPanics(t, func() { got = Append(testCase.input).Error() })
			requireEqual(t, testCase.want, got)

			if n := len(Errors(testCase.input)); n != 1 {
				t.Fatalf("Errors has %d entries, want 1", n)
			}
		})
	}
}
