package errkit

import (
	"errors"
	"log/slog"
)

// Proxy to StdLib errors.Is
func Is(err error, target ...error) bool {
	if err == nil {
		return false
	}
	for _, t := range target {
		if t == nil {
			continue
		}
		if errors.Is(err, t) {
			return true
		}
	}
	return false
}

// Proxy to StdLib errors.As
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// Combine combines multiple errors into a single error
func Combine(errs ...error) error {
	if len(errs) == 0 {
		return nil
	}
	x := &ErrorX{}
	for _, err := range errs {
		if err == nil {
			continue
		}
		parseError(x, err)
	}
	return x
}

// Wrap wraps the given error with the message
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	x := &ErrorX{}
	parseError(x, err)
	x.Msgf(message)
	return x
}

// Wrapf wraps the given error with the message
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	x := &ErrorX{}
	parseError(x, err)
	x.Msgf(format, args...)
	return x
}

// Errors returns all underlying errors there were appended or joined
func Errors(err error) []error {
	if err == nil {
		return nil
	}
	x := &ErrorX{}
	parseError(x, err)
	return x.errs
}

// Append appends given errors and returns a new error
// it ignores all nil errors
func Append(errs ...error) error {
	if len(errs) == 0 {
		return nil
	}
	x := &ErrorX{}
	for _, err := range errs {
		if err == nil {
			continue
		}
		parseError(x, err)
	}
	return x
}

// Cause returns the original error that caused this error
func Cause(err error) error {
	if err == nil {
		return nil
	}
	x := &ErrorX{}
	parseError(x, err)
	return x.Cause()
}

// WithMessage
func WithMessage(err error, message string) error {
	if err == nil {
		return nil
	}
	x := &ErrorX{}
	parseError(x, err)
	x.Msgf(message)
	return x
}

// WithMessagef
func WithMessagef(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	x := &ErrorX{}
	parseError(x, err)
	x.Msgf(format, args...)
	return x
}

// IsNetworkTemporaryErr checks if given error is a temporary network error
func IsNetworkTemporaryErr(err error) bool {
	if err == nil {
		return false
	}
	x := &ErrorX{}
	parseError(x, err)
	return isNetworkTemporaryErr(x)
}

// WithAttr wraps error with given attributes
//
// err = errkit.WithAttr(err,slog.Any("resource",domain))
func WithAttr(err error, attrs ...slog.Attr) error {
	if err == nil {
		return nil
	}
	x := &ErrorX{}
	parseError(x, err)
	x.attrs = append(x.attrs, attrs...)
	if len(x.attrs) > MaxErrorDepth {
		x.attrs = x.attrs[:MaxErrorDepth]
	}
	return x
}

// SlogAttrGroup returns a slog attribute group for the given error
// it is in format of:
//
//	{
//		"data": {
//			"kind": "<error-kind>",
//			"cause": "<cause>",
//			"errors": [
//				<errs>...
//			]
//		}
//	}
func SlogAttrGroup(err error) slog.Attr {
	attrs := SlogAttrs(err)
	g := slog.GroupValue(
		attrs..., // append all attrs
	)
	return slog.Any("data", g)
}

// SlogAttrs returns slog attributes for the given error
// it is in format of:
//
//	{
//		"kind": "<error-kind>",
//		"cause": "<cause>",
//		"errors": [
//			<errs>...
//		]
//	}
func SlogAttrs(err error) []slog.Attr {
	x := &ErrorX{}
	parseError(x, err)
	attrs := []slog.Attr{}
	if x.kind != nil {
		attrs = append(attrs, slog.Any("kind", x.kind.String()))
	}
	if cause := x.Cause(); cause != nil {
		attrs = append(attrs, slog.Any("cause", cause))
	}
	if len(x.errs) > 0 {
		attrs = append(attrs, slog.Any("errors", x.errs))
	}
	return attrs
}
