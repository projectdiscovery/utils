package errkit

import (
	"strings"

	"github.com/projectdiscovery/utils/env"
	"golang.org/x/exp/maps"
)

var (
	// MaxChainedKinds is the maximum number of error kinds that can be chained in a error
	MaxChainedKinds = env.GetEnvOrDefault("MAX_CHAINED_ERR_KINDS", 3)
)

var (
	// ErrClassNetwork indicates an error related to network operations
	// these may be resolved by retrying the operation with exponential backoff
	// ex: Timeout awaiting headers, i/o timeout etc
	ErrClassNetworkTemporary ErrKind = NewPrimitiveErrKind("network-temporary-error", "temporary network error", isNetworkTemporaryErr)
	// ErrClassNetworkPermanent indicates a permanent error related to network operations
	// these may not be resolved by retrying and need manual intervention
	// ex: no address found for host
	ErrClassNetworkPermanent = NewPrimitiveErrKind("network-permanent-error", "permanent network error", isNetworkPermanentErr)
	// ErrClassDeadline indicates a timeout error in logical operations
	// these are custom deadlines set by nuclei itself to prevent infinite hangs
	// and in most cases are server side issues (ex: server connects but does not respond at all)
	// a manual intervention is required
	ErrClassDeadline = NewPrimitiveErrKind("deadline-error", "deadline error", isDeadlineErr)
	// ErrClassUnknown indicates an unknown error class
	// that has not been implemented yet this is used as fallback when converting a slog Item
	ErrClassUnknown = NewPrimitiveErrKind("unknown-error", "unknown error", nil)
)

// ErrKind is an interface that represents a kind of error
type ErrKind interface {
	// Is checks if current error kind is same as given error kind
	Is(ErrKind) bool
	// IsParent checks if current error kind is parent of given error kind
	// this allows heirarchical classification of errors and app specific handling
	IsParent(ErrKind) bool
	// RepresentsError checks if given error is of this kind
	Represents(*ErrorX) bool
	// Description returns predefined description of the error kind
	// this can be used to show user friendly error messages in case of error
	Description() string
	// String returns the string representation of the error kind
	String() string
}

var _ ErrKind = &primitiveErrKind{}

// primitiveErrKind is kind of error used in classification
type primitiveErrKind struct {
	id         string
	info       string
	represents func(*ErrorX) bool
}

func (e *primitiveErrKind) Is(kind ErrKind) bool {
	return e.id == kind.String()
}

func (e *primitiveErrKind) IsParent(kind ErrKind) bool {
	return false
}

func (e *primitiveErrKind) Represents(err *ErrorX) bool {
	if e.represents != nil {
		return e.represents(err)
	}
	return false
}

func (e *primitiveErrKind) String() string {
	return e.id
}

func (e *primitiveErrKind) Description() string {
	return e.info
}

// NewPrimitiveErrKind creates a new primitive error kind
func NewPrimitiveErrKind(id string, info string, represents func(*ErrorX) bool) ErrKind {
	p := &primitiveErrKind{id: id, info: info, represents: represents}
	return p
}

func isNetworkTemporaryErr(err *ErrorX) bool {
	return false
}

// isNetworkPermanentErr checks if given error is a permanent network error
func isNetworkPermanentErr(err *ErrorX) bool {
	// to implement
	return false
}

// isDeadlineErr checks if given error is a deadline error
func isDeadlineErr(err *ErrorX) bool {
	// to implement
	return false
}

type multiKind struct {
	kinds []ErrKind
}

func (e *multiKind) Is(kind ErrKind) bool {
	for _, k := range e.kinds {
		if k.Is(kind) {
			return true
		}
	}
	return false
}

func (e *multiKind) IsParent(kind ErrKind) bool {
	for _, k := range e.kinds {
		if k.IsParent(kind) {
			return true
		}
	}
	return false
}

func (e *multiKind) Represents(err *ErrorX) bool {
	for _, k := range e.kinds {
		if k.Represents(err) {
			return true
		}
	}
	return false
}

func (e *multiKind) String() string {
	var str string
	for _, k := range e.kinds {
		str += k.String() + ","
	}
	return strings.TrimSuffix(str, ",")
}

func (e *multiKind) Description() string {
	var str string
	for _, k := range e.kinds {
		str += k.Description() + "\n"
	}
	return strings.TrimSpace(str)
}

// CombineErrKinds combines multiple error kinds into a single error kind
// this is not recommended but available if needed
// It is currently used in ErrorX while printing the error
// It is recommended to implement a hierarchical error kind
// instead of using this outside of errkit
func CombineErrKinds(kind ...ErrKind) ErrKind {
	// while combining it also consolidates child error kinds into parent
	// but note it currently does not support deeply nested childs
	// and can only consolidate immediate childs
	f := &multiKind{}
	uniq := map[ErrKind]struct{}{}
	for _, k := range kind {
		if k == nil {
			continue
		}
		if val, ok := k.(*multiKind); ok {
			for _, v := range val.kinds {
				uniq[v] = struct{}{}
			}
		} else {
			uniq[k] = struct{}{}
		}
	}
	all := maps.Keys(uniq)
	for _, k := range all {
		for u := range uniq {
			if k.IsParent(u) {
				delete(uniq, k)
			}
		}
	}
	f.kinds = maps.Keys(uniq)
	if len(f.kinds) > MaxChainedKinds {
		f.kinds = f.kinds[:MaxChainedKinds]
	}
	return f
}
