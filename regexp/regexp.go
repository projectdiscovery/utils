package regexp

import (
	"os"
	"regexp"

	"github.com/dlclark/regexp2"
	stringsutil "github.com/projectdiscovery/utils/strings"
	"github.com/wasilibs/go-re2"
)

// EngineType represents the type of regexp engine to use
type EngineType string

const (
	// EngineStandard uses the standard Go regexp engine
	EngineStandard EngineType = "standard"
	// EngineRegexp2 uses the regexp2 engine for .NET-style regex
	EngineRegexp2 EngineType = "regexp2"
	// EngineRE2 uses the RE2 engine for linear time matching
	EngineRE2 EngineType = "re2"
	// EngineAuto automatically selects the most appropriate engine
	EngineAuto EngineType = "auto"
)

// Option represents a configuration option for the regexp
type Option func(*Regexp)

// Regexp represents an extended regular expression with additional options
type Regexp struct {
	standard *regexp.Regexp
	regexp2  *regexp2.Regexp
	re2      *re2.Regexp
	engine   EngineType
	pattern  string // Store the original pattern
}

// WithEngine sets the regexp engine type
func WithEngine(engine EngineType) Option {
	return func(r *Regexp) {
		r.engine = engine
	}
}

// detectEngine analyzes the pattern and returns the most appropriate engine
func detectEngine(pattern string) EngineType {
	// Check for .NET-style features that regexp2 handles better
	hasNetStyle := stringsutil.ContainsAnyI(pattern,
		"(?<",   // Named capture groups
		"(?=",   // Positive lookahead
		"(?!)",  // Negative lookahead
		"(?<=",  // Positive lookbehind
		"(?<!)", // Negative lookbehind
	)
	if hasNetStyle {
		return EngineRegexp2
	}

	// Check for features that might cause catastrophic backtracking
	hasDangerousBacktracking := stringsutil.ContainsAnyI(pattern,
		".*",  // Greedy wildcard
		".+",  // Greedy one or more
		".*?", // Lazy wildcard
		".+?", // Lazy one or more
	)
	if hasDangerousBacktracking {
		return EngineRE2
	}

	// Default to standard engine for simple patterns
	return EngineStandard
}

// compileWithStandard attempts to compile the pattern with the standard Go regexp engine
func compileWithStandard(pattern string) (*Regexp, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return &Regexp{
		standard: re,
		engine:   EngineStandard,
		pattern:  pattern, // Store the pattern
	}, nil
}

// compileWithRE2 attempts to compile the pattern with the RE2 engine
func compileWithRE2(pattern string) (*Regexp, error) {
	originalStderr := os.Stderr
	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0666)
	if err != nil {
		// If we can't open os.DevNull, we can't suppress stderr.
		// Proceed without suppression, or return an error.
		// For now, let's proceed without suppression and let re2.Compile behave as usual.
		// Alternatively, we could log this issue or return a specific error.
		// log.Printf("Warning: could not open os.DevNull to suppress stderr: %v", err)
	} else {
		os.Stderr = devNull
		defer func() {
			os.Stderr = originalStderr
			devNull.Close()
		}()
	}

	re, err := re2.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return &Regexp{
		re2:     re,
		engine:  EngineRE2,
		pattern: pattern, // Store the pattern
	}, nil
}

// compileWithRegexp2 attempts to compile the pattern with the regexp2 engine
func compileWithRegexp2(pattern string) (*Regexp, error) {
	re, err := regexp2.Compile(pattern, regexp2.RE2)
	if err != nil {
		return nil, err
	}
	return &Regexp{
		regexp2: re,
		engine:  EngineRegexp2,
		pattern: pattern, // Store the pattern
	}, nil
}

// Compile creates a new Regexp with the given pattern and options
func Compile(pattern string, opts ...Option) (*Regexp, error) {
	r := &Regexp{
		engine:  EngineStandard, // default engine
		pattern: pattern,        // Store the pattern
	}

	// Apply options
	for _, opt := range opts {
		opt(r)
	}

	// If auto engine is selected, try different engines in sequence
	if r.engine == EngineAuto {
		// First try with the detected engine
		detectedEngine := detectEngine(pattern)
		var err error

		switch detectedEngine {
		case EngineRE2:
			r, err = compileWithRE2(pattern)
			if err == nil {
				return r, nil
			}
		case EngineRegexp2:
			r, err = compileWithRegexp2(pattern)
			if err == nil {
				return r, nil
			}
		}

		// If the detected engine failed or was standard, try RE2 first
		r, err = compileWithRE2(pattern)
		if err == nil {
			return r, nil
		}

		// Then try regexp2
		r, err = compileWithRegexp2(pattern)
		if err == nil {
			return r, nil
		}

		// Finally fall back to standard
		return compileWithStandard(pattern)
	}

	// For non-auto mode, compile directly with the specified engine
	switch r.engine {
	case EngineRE2:
		return compileWithRE2(pattern)
	case EngineRegexp2:
		return compileWithRegexp2(pattern)
	default:
		return compileWithStandard(pattern)
	}
}

// Match reports whether the byte slice b contains any match of the regular expression.
func (r *Regexp) Match(b []byte) bool {
	switch r.engine {
	case EngineRegexp2:
		match, _ := r.regexp2.MatchString(string(b))
		return match
	case EngineRE2:
		return r.re2.Match(b)
	default:
		return r.standard.Match(b)
	}
}

// MatchString reports whether the string s contains any match of the regular expression.
func (r *Regexp) MatchString(s string) bool {
	switch r.engine {
	case EngineRegexp2:
		match, _ := r.regexp2.MatchString(s)
		return match
	case EngineRE2:
		return r.re2.MatchString(s)
	default:
		return r.standard.MatchString(s)
	}
}

// Find returns a slice holding the text of the leftmost match in b of the regular expression.
func (r *Regexp) Find(b []byte) []byte {
	switch r.engine {
	case EngineRegexp2:
		match, _ := r.regexp2.FindStringMatch(string(b))
		if match == nil {
			return nil
		}
		return []byte(match.String())
	case EngineRE2:
		return r.re2.Find(b)
	default:
		return r.standard.Find(b)
	}
}

// FindString returns a string holding the text of the leftmost match in s of the regular expression.
func (r *Regexp) FindString(s string) string {
	switch r.engine {
	case EngineRegexp2:
		match, _ := r.regexp2.FindStringMatch(s)
		if match == nil {
			return ""
		}
		return match.String()
	case EngineRE2:
		return string(r.re2.FindString(s))
	default:
		return r.standard.FindString(s)
	}
}

// FindStringSubmatch returns a slice of strings holding the text of the leftmost match
// of the regular expression in s and the matches of its subexpressions.
func (r *Regexp) FindStringSubmatch(s string) []string {
	switch r.engine {
	case EngineRegexp2:
		match, _ := r.regexp2.FindStringMatch(s)
		if match == nil {
			return nil
		}
		matchGroups := match.Groups()
		if len(matchGroups) == 0 {
			return nil
		}
		groups := make([]string, len(matchGroups))
		for i, group := range matchGroups {
			groups[i] = group.String()
		}
		return groups
	case EngineRE2:
		return r.re2.FindStringSubmatch(s)
	default:
		return r.standard.FindStringSubmatch(s)
	}
}

// FindStringSubmatchIndex returns a slice holding the index pairs identifying the leftmost match
// of the regular expression in s and the matches of its subexpressions.
func (r *Regexp) FindStringSubmatchIndex(s string) []int {
	switch r.engine {
	case EngineRegexp2:
		match, _ := r.regexp2.FindStringMatch(s)
		if match == nil {
			return nil
		}
		// Convert regexp2 groups to index pairs
		indices := make([]int, 0, (match.GroupCount()+1)*2)
		for i := 0; i <= match.GroupCount(); i++ {
			group := match.GroupByNumber(i)
			indices = append(indices, group.Index, group.Index+group.Length)
		}
		return indices
	case EngineRE2:
		return r.re2.FindStringSubmatchIndex(s)
	default:
		return r.standard.FindStringSubmatchIndex(s)
	}
}

// String returns the original regular expression pattern.
// For standard and RE2 engines, it calls their String() method.
// For Regexp2, it returns the stored pattern as Regexp2.String() returns the FSM.
func (r *Regexp) String() string {
	switch r.engine {
	case EngineRegexp2:
		return r.pattern // Return stored pattern for regexp2
	case EngineRE2:
		return r.re2.String()
	default: // EngineStandard
		return r.standard.String()
	}
}
