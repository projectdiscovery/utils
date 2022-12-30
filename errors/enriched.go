package errors

import (
	"bytes"
	"fmt"
	"runtime/debug"
	"strings"
)

type ErrorLevel uint

const (
	Panic ErrorLevel = iota
	Fatal
	Runtime // Default
)

func (l ErrorLevel) String() string {
	switch l {
	case Panic:
		return "PANIC"
	case Fatal:
		return "FTL"
	case Runtime:
		return "RUNTIME"
	}
	return ""
}

// *Error is enriched version of normal error
// with tags, stacktrace and other methods
type Error struct {
	errString  string
	StackTrace string
	Tags       []string
	Level      ErrorLevel

	//OnError is called when Error() method is triggered
	OnError func()
}

// withTag assignes tag to Error
func (e *Error) WithTag(tag ...string) *Error {
	if e.Tags == nil {
		e.Tags = tag
	} else {
		e.Tags = append(e.Tags, tag...)
	}
	return e
}

// withLevel assinges level to Error
func (e *Error) WithLevel(level ErrorLevel) *Error {
	e.Level = level
	return e
}

// returns formated *Error string
func (e *Error) Error() string {
	defer func() {
		if e.OnError != nil {
			e.OnError()
		}
	}()
	e.captureStack()
	var buff bytes.Buffer

	if len(e.Tags) > 0 {
		buff.WriteString(fmt.Sprintf("[%v]", strings.Join(e.Tags, " ")))
	}
	buff.WriteString(fmt.Sprintf("[%v] %v\n", e.Level.String(), e.errString))
	buff.WriteString(fmt.Sprintf("Stacktrace:\n%v\n", e.StackTrace))
	return buff.String()
}

// wraps given error
func (e *Error) Wrap(err ...error) *Error {
	// wraps like a stack
	for _, v := range err {
		if v == nil {
			continue
		}
		e = e.Wrapf(v.Error())
	}
	return e
}

// Wrapf wraps given message
func (e *Error) Wrapf(format string, args ...any) *Error {
	// unlike wrapping `right -> left` it wraps like a stack (bottom -> up)
	msg := fmt.Sprintf(format, args...)
	if e.errString == "" {
		e.errString = msg
	} else {
		e.errString = fmt.Sprintf("%v:\n%v", msg, e.errString)
	}
	return e
}

// captureStack
func (e *Error) captureStack() {
	// can be furthur improved to format
	// ref https://github.com/go-errors/errors/blob/33d496f939bc762321a636d4035e15c302eb0b00/stackframe.go
	e.StackTrace = string(debug.Stack())
}

// Equal returns true if error matches anyone of given errors
func (e *Error) Equal(err ...error) bool {
	for _, v := range err {
		if ee, ok := v.(*Error); ok {
			if e.errString == ee.errString {
				return true
			}
		} else {
			// not an enriched error but a simple eror
			if e.errString == v.Error() {
				return true
			}
		}
	}
	return false
}

// New
func New(format string, args ...any) *Error {
	ee := &Error{
		errString: fmt.Sprintf(format, args...),
	}
	return ee
}

func NewWithErr(err error) *Error {
	return New(err.Error())
}

// NewWithTag creates an error with tag
func NewWithTag(tag string, format string, args ...any) *Error {
	ee := New(format, args...)
	ee.Tags = []string{tag}
	return ee
}
