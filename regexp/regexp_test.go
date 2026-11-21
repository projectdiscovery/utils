package regexp

import (
	"errors"
	"os"
	"regexp/syntax"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
	"unicode"
)

func TestCompile(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		engine  EngineType
		wantErr bool
	}{
		{
			name:    "valid standard pattern",
			pattern: "test[0-9]+",
			engine:  EngineStandard,
			wantErr: false,
		},
		{
			name:    "valid regexp2 pattern with lookahead",
			pattern: "test(?=123)",
			engine:  EngineRegexp2,
			wantErr: false,
		},
		{
			name:    "valid RE2 pattern",
			pattern: "test.*",
			engine:  EngineRE2,
			wantErr: false,
		},
		{
			name:    "invalid pattern",
			pattern: "[",
			engine:  EngineStandard,
			wantErr: true,
		},
		{
			name:    "auto engine with standard pattern",
			pattern: "test[0-9]+",
			engine:  EngineAuto,
			wantErr: false,
		},
		{
			name:    "auto engine with regexp2 pattern",
			pattern: "test(?=123)",
			engine:  EngineAuto,
			wantErr: false,
		},
		{
			name:    "auto engine with RE2 pattern",
			pattern: "test.*",
			engine:  EngineAuto,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re, err := Compile(tt.pattern, WithEngine(tt.engine))
			if (err != nil) != tt.wantErr {
				t.Errorf("Compile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && re == nil {
				t.Error("Compile() returned nil Regexp when no error expected")
			}
		})
	}
}

func TestMatch(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		input   []byte
		engine  EngineType
		want    bool
	}{
		{
			name:    "standard engine match",
			pattern: "test[0-9]+",
			input:   []byte("test123"),
			engine:  EngineStandard,
			want:    true,
		},
		{
			name:    "standard engine no match",
			pattern: "test[0-9]+",
			input:   []byte("testabc"),
			engine:  EngineStandard,
			want:    false,
		},
		{
			name:    "regexp2 engine with lookahead",
			pattern: "test(?=123)",
			input:   []byte("test123"),
			engine:  EngineRegexp2,
			want:    true,
		},
		{
			name:    "RE2 engine match",
			pattern: "test.*",
			input:   []byte("test123"),
			engine:  EngineRE2,
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re, err := Compile(tt.pattern, WithEngine(tt.engine))
			if err != nil {
				t.Fatalf("Compile() error = %v", err)
			}
			if got := re.Match(tt.input); got != tt.want {
				t.Errorf("Match() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatchString(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		input   string
		engine  EngineType
		want    bool
	}{
		{
			name:    "standard engine match",
			pattern: "test[0-9]+",
			input:   "test123",
			engine:  EngineStandard,
			want:    true,
		},
		{
			name:    "standard engine no match",
			pattern: "test[0-9]+",
			input:   "testabc",
			engine:  EngineStandard,
			want:    false,
		},
		{
			name:    "regexp2 engine with lookahead",
			pattern: "test(?=123)",
			input:   "test123",
			engine:  EngineRegexp2,
			want:    true,
		},
		{
			name:    "RE2 engine match",
			pattern: "test.*",
			input:   "test123",
			engine:  EngineRE2,
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re, err := Compile(tt.pattern, WithEngine(tt.engine))
			if err != nil {
				t.Fatalf("Compile() error = %v", err)
			}
			if got := re.MatchString(tt.input); got != tt.want {
				t.Errorf("MatchString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFind(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		input   []byte
		engine  EngineType
		want    []byte
	}{
		{
			name:    "standard engine find",
			pattern: "test[0-9]+",
			input:   []byte("abc test123 def"),
			engine:  EngineStandard,
			want:    []byte("test123"),
		},
		{
			name:    "standard engine no find",
			pattern: "test[0-9]+",
			input:   []byte("abc testabc def"),
			engine:  EngineStandard,
			want:    nil,
		},
		{
			name:    "regexp2 engine with lookahead",
			pattern: "test(?=123)",
			input:   []byte("abc test123 def"),
			engine:  EngineRegexp2,
			want:    []byte("test"),
		},
		{
			name:    "RE2 engine find",
			pattern: "test.*",
			input:   []byte("abc test123 def"),
			engine:  EngineRE2,
			want:    []byte("test123 def"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re, err := Compile(tt.pattern, WithEngine(tt.engine))
			if err != nil {
				t.Fatalf("Compile() error = %v", err)
			}
			got := re.Find(tt.input)
			if string(got) != string(tt.want) {
				t.Errorf("Find() = %v, want %v", string(got), string(tt.want))
			}
		})
	}
}

func TestFindString(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		input   string
		engine  EngineType
		want    string
	}{
		{
			name:    "standard engine find",
			pattern: "test[0-9]+",
			input:   "abc test123 def",
			engine:  EngineStandard,
			want:    "test123",
		},
		{
			name:    "standard engine no find",
			pattern: "test[0-9]+",
			input:   "abc testabc def",
			engine:  EngineStandard,
			want:    "",
		},
		{
			name:    "regexp2 engine with lookahead",
			pattern: "test(?=123)",
			input:   "abc test123 def",
			engine:  EngineRegexp2,
			want:    "test",
		},
		{
			name:    "RE2 engine find",
			pattern: "test.*",
			input:   "abc test123 def",
			engine:  EngineRE2,
			want:    "test123 def",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re, err := Compile(tt.pattern, WithEngine(tt.engine))
			if err != nil {
				t.Fatalf("Compile() error = %v", err)
			}
			if got := re.FindString(tt.input); got != tt.want {
				t.Errorf("FindString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindStringSubmatch(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		input   string
		engine  EngineType
		want    []string
	}{
		{
			name:    "standard engine with capture group",
			pattern: "test([0-9]+)",
			input:   "abc test123 def",
			engine:  EngineStandard,
			want:    []string{"test123", "123"},
		},
		{
			name:    "standard engine no match",
			pattern: "test([0-9]+)",
			input:   "abc testabc def",
			engine:  EngineStandard,
			want:    nil,
		},
		{
			name:    "regexp2 engine with named capture",
			pattern: "test(?<num>[0-9]+)",
			input:   "abc test123 def",
			engine:  EngineRegexp2,
			want:    []string{"test123", "123"},
		},
		{
			name:    "RE2 engine with capture group",
			pattern: "test([0-9]+)",
			input:   "abc test123 def",
			engine:  EngineRE2,
			want:    []string{"test123", "123"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re, err := Compile(tt.pattern, WithEngine(tt.engine))
			if err != nil {
				t.Fatalf("Compile() error = %v", err)
			}
			got := re.FindStringSubmatch(tt.input)
			if len(got) != len(tt.want) {
				t.Errorf("FindStringSubmatch() length = %v, want %v", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("FindStringSubmatch()[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestFindStringSubmatchIndex(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		input   string
		engine  EngineType
		want    []int
	}{
		{
			name:    "standard engine with capture group",
			pattern: "test([0-9]+)",
			input:   "abc test123 def",
			engine:  EngineStandard,
			want:    []int{4, 11, 8, 11}, // [start, end, group1_start, group1_end]
		},
		{
			name:    "standard engine no match",
			pattern: "test([0-9]+)",
			input:   "abc testabc def",
			engine:  EngineStandard,
			want:    nil,
		},
		{
			name:    "regexp2 engine with named capture",
			pattern: "test(?<num>[0-9]+)",
			input:   "abc test123 def",
			engine:  EngineRegexp2,
			want:    []int{4, 11, 8, 11}, // [start, end, group1_start, group1_end]
		},
		{
			name:    "RE2 engine with capture group",
			pattern: "test([0-9]+)",
			input:   "abc test123 def",
			engine:  EngineRE2,
			want:    []int{4, 11, 8, 11}, // [start, end, group1_start, group1_end]
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re, err := Compile(tt.pattern, WithEngine(tt.engine))
			if err != nil {
				t.Fatalf("Compile() error = %v", err)
			}
			got := re.FindStringSubmatchIndex(tt.input)
			if len(got) != len(tt.want) {
				t.Errorf("FindStringSubmatchIndex() length = %v, want %v", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("FindStringSubmatchIndex()[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		engine  EngineType
	}{
		{
			name:    "standard engine",
			pattern: "test[0-9]+",
			engine:  EngineStandard,
		},
		{
			name:    "regexp2 engine",
			pattern: "test(?=123)",
			engine:  EngineRegexp2,
		},
		{
			name:    "RE2 engine",
			pattern: "test.*",
			engine:  EngineRE2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re, err := Compile(tt.pattern, WithEngine(tt.engine))
			if err != nil {
				t.Fatalf("Compile() error = %v", err)
			}
			if got := re.String(); got != tt.pattern {
				t.Errorf("String() = %v, want %v", got, tt.pattern)
			}
		})
	}
}

// TestCompileRE2Syntax checks that Compile with the RE2 engine accepts and
// rejects the same patterns as RE2 itself, including \C which only RE2
// supports, and that auto selection still falls back to regexp2.
func TestCompileRE2Syntax(t *testing.T) {
	t.Parallel()

	tests := []struct {
		pattern string
		wantErr bool
	}{
		{pattern: `a+b`, wantErr: false},
		{pattern: `(?P<name>x)\d{1,3}`, wantErr: false},
		{pattern: `\pN+[[:alpha:]]`, wantErr: false},
		{pattern: `\C`, wantErr: false},
		{pattern: nonASCIIGroup, wantErr: false},
		{pattern: overflowRepeat, wantErr: false},
		{pattern: `a(`, wantErr: true},
		{pattern: `a**`, wantErr: true},
		{pattern: `(?=x)`, wantErr: true},
		{pattern: `(?<=x)`, wantErr: true},
		{pattern: backReference, wantErr: true},
		{pattern: oversizedRepeat, wantErr: true},
		{pattern: `\Z`, wantErr: true},
	}

	for _, testCase := range tests {
		t.Run(testCase.pattern, func(t *testing.T) {
			t.Parallel()

			_, err := Compile(testCase.pattern, WithEngine(EngineRE2))
			if (err != nil) != testCase.wantErr {
				t.Errorf("Compile() error = %v, wantErr %v", err, testCase.wantErr)
			}
		})
	}

	auto, err := Compile(`a(?=b)`, WithEngine(EngineAuto))
	if err != nil {
		t.Fatalf("Compile() with lookahead and EngineAuto error = %v", err)
	}

	if auto.engine != EngineRegexp2 {
		t.Errorf("engine = %v, want %v", auto.engine, EngineRegexp2)
	}
}

// TestCompileRE2SyntaxErrorUnwrap checks that the error for a pattern the gate rejects keeps the
// *syntax.Error reachable through errors.As, with the context in front of its text.
func TestCompileRE2SyntaxErrorUnwrap(t *testing.T) {
	t.Parallel()

	_, err := Compile(`a(`, WithEngine(EngineRE2))
	if err == nil {
		t.Fatal("Compile() succeeded, want error")
	}

	var synErr *syntax.Error
	if !errors.As(err, &synErr) {
		t.Fatalf("Compile() error = %v (%T), want one wrapping *syntax.Error", err, err)
	}

	if synErr.Code != syntax.ErrMissingParen {
		t.Errorf("syntax error code = %q, want %q", synErr.Code, syntax.ErrMissingParen)
	}

	const want = "unsupported RE2 syntax: error parsing regexp: missing closing ): `a(`"
	if err.Error() != want {
		t.Errorf("Compile() error = %q, want %q", err.Error(), want)
	}
}

// checkCompile compiles pattern with engine and reports through t, which is
// safe to call from several goroutines, if it is accepted or rejected
// contrary to valid, or if an accepted pattern fails to match "xaab".
func checkCompile(t *testing.T, engine EngineType, pattern string, valid bool) {
	t.Helper()

	got, err := Compile(pattern, WithEngine(engine))
	if (err == nil) != valid {
		t.Errorf("Compile(%q, %s) error = %v, valid %v", pattern, engine, err, valid)

		return
	}

	if valid && !got.MatchString("xaab") {
		t.Errorf("Compile(%q, %s) does not match %q", pattern, engine, "xaab")
	}
}

// TestCompileConcurrentRE2 compiles with the RE2 engine from many goroutines
// at once. Compile must not panic and must leave the process-wide os.Stdout
// and os.Stderr untouched, since other goroutines are free to write to them.
func TestCompileConcurrentRE2(t *testing.T) {
	t.Parallel()

	const (
		goroutines = 200
		rounds     = 5
	)

	stdout, stderr := os.Stdout, os.Stderr

	start := make(chan struct{})

	var group sync.WaitGroup

	for range goroutines {
		group.Go(func() {
			<-start

			for range rounds {
				for _, engine := range []EngineType{EngineRE2, EngineAuto} {
					checkCompile(t, engine, "a+b", true)
					checkCompile(t, engine, "a(", false)
				}
			}
		})
	}

	done := make(chan struct{})

	go func() {
		group.Wait()
		close(done)
	}()

	close(start)

	select {
	case <-done:
	case <-time.After(time.Minute):
		t.Fatal("concurrent Compile did not finish within a minute")
	}

	if os.Stdout != stdout {
		t.Errorf("os.Stdout was replaced: got %v, want %v", os.Stdout, stdout)
	}

	if os.Stderr != stderr {
		t.Errorf("os.Stderr was replaced: got %v, want %v", os.Stderr, stderr)
	}
}

const (
	// gateFullCorpusEnv names the environment variable that enables TestCheckRE2SyntaxFullCorpus.
	gateFullCorpusEnv = "REGEXP_RE2_GATE_FULL"

	nonASCIIGroup       = `(?P<é>x)`
	overflowRepeat      = `a{0,99999999999999999999}`
	oversizedRepeat     = `a{1001}`
	reversedRange       = `[z-a]`
	backReference       = `(a)\1`
	pythonBackReference = `(?P=n)`
	oneByteInClass      = `[\C]`
)

// assertGateAcceptsWhatRE2Accepts fails the test for every pattern that checkRE2Syntax rejects and RE2
// accepts. A pattern the gate lets through but RE2 rejects is a known limitation and is not checked.
// It returns the number of patterns the gate rejected.
func assertGateAcceptsWhatRE2Accepts(t *testing.T, patterns []string) int {
	t.Helper()

	rejected := 0

	for _, pattern := range patterns {
		gateErr := checkRE2Syntax(pattern)
		if gateErr == nil {
			continue
		}

		rejected++

		_, re2Err := compileRE2(pattern)
		if re2Err == nil {
			t.Errorf("checkRE2Syntax(%.80q) = %v, but RE2 accepts it", pattern, gateErr)
		}
	}

	return rejected
}

// TestCheckRE2SyntaxAcceptsWhatRE2Accepts is the invariant behind checkRE2Syntax: the gate must never
// reject a pattern that RE2 accepts. It runs a fixed set of about 200 patterns that covers every
// trusted error code, every exception and every counterexample found so far. Each pattern the gate
// rejects is compiled by RE2, which logs it to stderr, so the set is kept small; the large generated
// corpus is in TestCheckRE2SyntaxFullCorpus.
func TestCheckRE2SyntaxAcceptsWhatRE2Accepts(t *testing.T) {
	t.Parallel()

	patterns := re2GateFixedPatterns()
	rejected := assertGateAcceptsWhatRE2Accepts(t, patterns)
	t.Logf("%d patterns, %d rejected by the gate", len(patterns), rejected)
}

// TestCheckRE2SyntaxFullCorpus runs the invariant of TestCheckRE2SyntaxAcceptsWhatRE2Accepts over a
// corpus of about 90,000 patterns. Run it after upgrading Go or go-re2 by setting REGEXP_RE2_GATE_FULL=1;
// RE2 logs every pattern the gate rejects, so with -v the output is several megabytes.
func TestCheckRE2SyntaxFullCorpus(t *testing.T) {
	t.Parallel()

	if os.Getenv(gateFullCorpusEnv) != "1" {
		t.Skipf("set %s=1 to run the full corpus", gateFullCorpusEnv)
	}

	for name, patterns := range re2GatePatterns() {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			rejected := assertGateAcceptsWhatRE2Accepts(t, patterns)
			t.Logf("%d patterns, %d rejected by the gate", len(patterns), rejected)
		})
	}
}

// TestCheckRE2SyntaxVerdict pins which constructs the gate rejects, so RE2 is never asked about them
// and logs nothing, and which it leaves to RE2.
func TestCheckRE2SyntaxVerdict(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		pattern  string
		rejected bool
	}{
		{name: "lookahead", pattern: `a(?=b)`, rejected: true},
		{name: "negative lookahead", pattern: `a(?!b)`, rejected: true},
		{name: "atomic group", pattern: `(?>a+)`, rejected: true},
		{name: "possessive plus", pattern: `a++`, rejected: true},
		{name: "possessive star", pattern: `a*+`, rejected: true},
		{name: "backreference", pattern: backReference, rejected: true},
		{name: "named backreference", pattern: `(?<n>a)\k<n>`, rejected: true},
		{name: "python backreference", pattern: `(?P<n>a)(?P=n)`, rejected: true},
		{name: "conditional", pattern: `(a)(?(1)b|c)`, rejected: true},
		{name: "start of match anchor", pattern: `\Gabc`, rejected: true},
		{name: "keep out", pattern: `a\Kb`, rejected: true},
		{name: "grapheme cluster", pattern: `\X`, rejected: true},
		{name: "line break", pattern: `\R`, rejected: true},
		{name: "horizontal space", pattern: `\h`, rejected: true},
		{name: "property as range end", pattern: `[a-\pL]`, rejected: true},
		{name: "braced property as range end", pattern: `[a-\p{L}]`, rejected: true},
		{name: "unclosed group", pattern: `(a`, rejected: true},
		{name: "unopened group", pattern: `a)`, rejected: true},
		{name: "unclosed class", pattern: `[a`, rejected: true},
		{name: "repeat without operand", pattern: `*a`, rejected: true},
		{name: "trailing backslash", pattern: `a\`, rejected: true},
		{name: "plain pattern", pattern: `a+b`, rejected: false},
		{name: "one-byte any", pattern: `\C`, rejected: false},
		{name: "underscore script", pattern: `\p{Old_Italic}`, rejected: false},
		{name: "non-ASCII group name", pattern: nonASCIIGroup, rejected: false},
		{name: "overflowing repeat", pattern: overflowRepeat, rejected: false},
		{name: "encoded surrogate", pattern: "\xed\xa0\x80", rejected: false},
		{name: "invalid UTF-8", pattern: "\xff", rejected: false},
		{name: "unknown property", pattern: `\p{Foo}`, rejected: false},
		{name: "one-byte any in class", pattern: oneByteInClass, rejected: false},
		{name: "reversed range", pattern: reversedRange, rejected: false},
		{name: "lookbehind", pattern: `(?<=a)b`, rejected: false},
		{name: "oversized repeat", pattern: oversizedRepeat, rejected: false},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if got := checkRE2Syntax(testCase.pattern) != nil; got != testCase.rejected {
				t.Errorf("checkRE2Syntax(%q) rejected = %v, want %v", testCase.pattern, got, testCase.rejected)
			}
		})
	}
}

// re2GateFixedPatterns returns the patterns of TestCheckRE2SyntaxAcceptsWhatRE2Accepts.
func re2GateFixedPatterns() []string {
	patterns := []string{
		// Trusted codes: unsupported group, repeat operator, brackets, parentheses, escape.
		`(?=a)`, `(?!a)`, `(?>a)`, pythonBackReference, `(?(1)a|b)`, `(?#c)`, `(?z)`, `(?i`, `(?R)`, `(?-x)`, `(?|a|b)`,
		`a**`, `a++`, `a*+`, `a?+`, `a{2}+`, `a{2}{3}`,
		`[a`, `[`, `[[:alnum:]`, `[^`, `[a-`,
		`(`, `(a`, `(?:a`, `(?P<n>a`, `a(b`,
		`*a`, `+`, `?`, `|*`, `(*)`, `a|*`,
		`\`, `a\`, `)`, `a)`, `(a))`,
		`\1`, backReference, `\8`, `[\1]`, `\Z`, `\G`, `\K`, `\X`, `\R`, `\h`, `\H`, `\e`, `\cA`, `\k<n>`, `\g1`, `\x`,
		`\x{110000}`, `\x{D800}`, `\é`, `[\é]`, `\` + `u0041`, `\N`, `\o{101}`,
		// \p and \P as the end of a range are an invalid escape, which RE2 rejects too.
		`[a-\pL]`, `[a-\p{L}]`, `[a-\P{L}]`, `[a-\p{Greek}]`, `[a-\pX]`, `[+--\pL]`,
		// Exceptions: \C is RE2-only, and \p names depend on Unicode tables.
		`\C`, `\C+`, oneByteInClass, `a\Cb`, `\C\C`, `\p`, `\P`, `\p{`, `\p{}`, `[\p]`, `\pX`, `\p{Foo}`,
		// Counterexamples: Go rejects, RE2 accepts. Encoded surrogates, overflowing counts, deep nesting,
		// non-ASCII group names.
		"\xed\xa0\x80", "a\xed\xa0\x80b", "[\xed\xa0\x80]", "\xed\xbf\xbf", "\xed\xa0\x80+", "(\xed\xb0\x80)",
		overflowRepeat, `(a{1}){99999999999999999999}`, `a{99999999999999999999,}`,
		nonASCIIGroup, `(?<é>x)`, `(?P<имя>x)`, `(?<имя>x)`,
		strings.Repeat("(a*", 1001) + strings.Repeat(")", 1001), strings.Repeat("(", 2000) + strings.Repeat(")", 2000),
		// Rejected by RE2 but left to it by the gate.
		reversedRange, `[a-\d]`, "\xff", "a\xffb", oversizedRepeat, `a{2,1}`, `(?<=a)b`, `(?<!a)b`, `[[:foo:]]`,
	}

	names := []string{
		"Old_Italic", "Canadian_Aboriginal", "SignWriting", "Tai_Le", "Nag_Mundari", "Old_North_Arabian", "Meetei_Mayek",
		"Phags_Pa", "Pau_Cin_Hau", "Hanifi_Rohingya", "Anatolian_Hieroglyphs", "Greek", "Latin", "Han", "L", "Lu", "Nd",
	}
	classes := make([]string, 0, len(names)*6)

	for _, name := range names {
		classes = append(classes, `\p{`+name+`}`, `\P{`+name+`}`, `[\p{`+name+`}]+`, `x.*\P{`+name+`}`,
			`\p{^`+name+`}`, `[^\p{`+name+`}]`)
	}

	return slices.Concat(patterns, classes)
}

// re2GatePatterns returns the pattern corpora for TestCheckRE2SyntaxFullCorpus, keyed by name.
func re2GatePatterns() map[string][]string {
	const (
		printableASCII = 0x7f - 0x20
		posixNames     = 16
	)

	escapes := make([]string, 0, printableASCII*4)
	posixClasses := make([]string, 0, posixNames*7)

	for char := rune(0x20); char < 0x7f; char++ {
		escaped := `\` + string(char)
		escapes = append(escapes, escaped, `[`+escaped+`]`, `[^`+escaped+`]`, `a`+escaped+`+`)
	}

	for _, name := range []string{
		"alnum", "alpha", "ascii", "blank", "cntrl", "digit", "graph", "lower", "print", "punct", "space", "upper",
		"word", "xdigit", "foo", "",
	} {
		posixClasses = append(posixClasses, "[[:"+name+":]]", "[[:^"+name+":]]", "[^[:"+name+":]]", "[a[:"+name+":]b]",
			"[:"+name+":]", "[[:"+name+":]", "[[:"+name+"]]")
	}

	names := []string{"Any", "Foo", "greek", "L&", "Lu ", "", "^", "^L"}
	for name := range unicode.Categories {
		names = append(names, name)
	}

	for name := range unicode.Scripts {
		names = append(names, name)
	}

	for name := range unicode.Properties {
		names = append(names, name)
	}

	for letter := 'A'; letter <= 'Z'; letter++ {
		names = append(names, string(letter))
	}

	unicodeClasses := make([]string, 0, len(names)*9)

	for _, name := range names {
		unicodeClasses = append(unicodeClasses, `\p{`+name+`}`, `\P{`+name+`}`, `\p{^`+name+`}`, `\P{^`+name+`}`,
			`[\p{`+name+`}]`, `[^\P{`+name+`}]`, `\p`+name, `\P`+name, `[\p`+name+`]`)
	}

	return map[string][]string{
		"escapes":             escapes,
		"posix classes":       posixClasses,
		"unicode classes":     unicodeClasses,
		"perl classes":        perlClassPatterns(),
		"groups and flags":    groupAndFlagPatterns(),
		"repeats":             repeatPatterns(),
		"numeric escapes":     numericEscapePatterns(),
		"quoting and anchors": miscPatterns(),
		"nesting and size":    nestingAndSizePatterns(),
		"invalid UTF-8":       invalidUTF8Patterns(),
		"generated":           generatedPatterns(),
	}
}

// invalidUTF8Patterns returns every lone high byte, every two-byte sequence with a continuation byte, and
// every encoded surrogate (ED A0..BF 80..BF), bare, inside a class and inside text.
func invalidUTF8Patterns() []string {
	var sequences []string

	for lead := 0x80; lead <= 0xff; lead++ {
		sequences = append(sequences, string([]byte{byte(lead)}))
	}

	for lead := 0xc0; lead <= 0xff; lead++ {
		for next := 0x80; next <= 0xbf; next++ {
			sequences = append(sequences, string([]byte{byte(lead), byte(next)}))
		}
	}

	for second := 0xa0; second <= 0xbf; second++ {
		for third := 0x80; third <= 0xbf; third++ {
			sequences = append(sequences, string([]byte{0xed, byte(second), byte(third)}))
		}
	}

	patterns := make([]string, 0, len(sequences)*3)
	for _, sequence := range sequences {
		patterns = append(patterns, sequence, "["+sequence+"]", "a"+sequence+"b")
	}

	return patterns
}

func perlClassPatterns() []string {
	classes := []string{`\d`, `\D`, `\s`, `\S`, `\w`, `\W`, `\h`, `\H`, `\R`, `\X`, `\N`, `\K`}
	patterns := make([]string, 0, len(classes)*9)

	for _, class := range classes {
		patterns = append(patterns, class, class+`+`, `[`+class+`]`, `[^`+class+`]`, `[a`+class+`]`, `[`+class+`-z]`,
			`[a-`+class+`]`, `(?i)`+class, `(?s)`+class+`.`)
	}

	return patterns
}

func groupAndFlagPatterns() []string {
	patterns := []string{
		`(?P<n>x)`, `(?<n>x)`, `(?P<n>x)(?P<n>y)`, `(?<n>x)(?<n>y)`, `(?P<n1>x)`, `(?P<1n>x)`, `(?P<n-x>x)`,
		`(?P<_>x)`, `(?P<>x)`, `(?<>x)`, `(?P<é>x)`, `(?<é>x)`, `(?P<n`, `(?<n`, `(?P=n)`, `(?P>n)`, `(?'n'x)`,
		`(?P<n>x)\k<n>`, `\g1`, `(?<=x)`, `(?<!x)`, `(?=x)`, `(?!x)`, `(?>x)`, `(?#c)`, `(?|a|b)`, `(?R)`, `(?1)`,
		`(?(1)a|b)`, `(?`, `(?)`, `(?:)`, `(?i)`, `(?^)`, `(?^i)`, `(?-)`, `(?--i)`, `(?i-)`, `(?i-i)`,
	}

	var addFlagGroups func(flags string, depth int)

	addFlagGroups = func(flags string, depth int) {
		patterns = append(patterns, `(?`+flags+`)`, `(?`+flags+`:a+)`, `x(?`+flags+`)y`)

		if depth == 0 {
			return
		}

		for _, flag := range "imsUxualn-:" {
			addFlagGroups(flags+string(flag), depth-1)
		}
	}

	addFlagGroups("", 3)

	return patterns
}

func repeatPatterns() []string {
	counts := []string{
		"", "0", "1", "2", "255", "256", "999", "1000", "1001", "1 ", " 1", "-1", "01", "00", "a",
		"99999999999999999999", "1e3",
	}
	operators := []string{`*`, `+`, `?`, `**`, `*+`, `*?`, `+?`, `??`, `*??`, `+*`, `?*`, `?+`, `{2}{3}`, `{2}*`, `*{2}`}

	patterns := make([]string, 0, len(operators)*9+len(counts)*(7+6*len(counts))+13)

	for _, operator := range operators {
		patterns = append(patterns, `a`+operator, `(a)`+operator, `[a]`+operator, `^`+operator, `\b`+operator,
			`(?:)`+operator, `|`+operator, `(`+operator+`)`, `a|`+operator)
	}

	for _, first := range counts {
		patterns = append(patterns, `a{`+first+`}`, `a{`+first+`,}`, `a{,`+first+`}`, `a{`+first+`}?`,
			`(a{`+first+`}){`+first+`}`, `(?:a{`+first+`})*`, `(?:a*){`+first+`}`)

		for _, second := range counts {
			patterns = append(patterns, `a{`+first+`,`+second+`}`, `a{`+first+`,`+second+`}?`,
				`(a{`+first+`}){`+second+`}`, `(?:a{`+first+`,`+second+`}){`+second+`}`,
				`(a{`+first+`,`+second+`}){2}`, `(?:(?:a{`+first+`}){`+second+`}){2}`)
		}
	}

	return append(patterns, `(a{500}){2}`, `(a{500}){3}`, `((a{10}){10}){10}`, `((a{10}){10}){11}`,
		`((a{10}){10}){10}b`, `(((a{10}){10}){10}){2}`, `.{1000}`, `.{1000}.{1000}`, `[a-z]{1000}`,
		`\pL{1000}`, `(\pL{1000}){1000}`, `(?:.{500}){2}`, `(.{1000}){1000}`)
}

func numericEscapePatterns() []string {
	const numbered = 1000

	patterns := make([]string, 0, numbered*3+32)
	patterns = append(patterns,
		`\x`, `\x4`, `\x41`, `\x414`, `\xg1`, `\x{`, `\x{}`, `\x{41`, `\x{41}`, `\x{0041}`, `\x{00000041}`,
		`\x{10FFFF}`, `\x{110000}`, `\x{D800}`, `\x{DFFF}`, `\x{-1}`, `\x{ 41}`, `\x{4 1}`, `\x{g}`,
		`[\x{41}-\x{5a}]`, `[\x{5a}-\x{41}]`, `\`+`u0041`, `\U00000041`, `\o{101}`, `\N{U+41}`, `\cA`, `\c`, `\e`, `\ca`,
	)

	for number := range numbered {
		digits := strconv.Itoa(number)
		patterns = append(patterns, `\`+digits, `(a)\`+digits, `[\`+digits+`]`)
	}

	return patterns
}

func miscPatterns() []string {
	return []string{
		``, `|`, `a|`, `|a`, `||`, `()`, `(|)`, `(?:)`, `(?:|)`, `^*`, `$+`, `\b*`, `\A+`, `\z?`, `^^`, `$$`,
		`\Q`, `\Qabc`, `\E`, `\Q\E`, `\Q.\E+`, `\Qa\\E`, `[\Q]\E]`, `\Qa\Eb\Qc`, `\Q\Q\E\E`, `\A`, `\z`, `\Z`,
		`\b`, `\B`, `\<`, `\>`, `\G`, `\a`, `\f`, `\n`, `\r`, `\t`, `\v`, `\C`, `\C*`, `[\C]`, `\C\C`, `\Ca`,
		"a\x00b", "\x00", "\xff", "a\xffb", "\xc3", "é", "[é-ü]", "[\xff]", "\U0010ffff", string(rune(0xfffd)),
		`.`, `.*`, `.+?`, `(.*)(.*)`, `[]a]`, `[]`, `[^]a]`, `[a-]`, `[-a]`, `[a-a]`, `[z-a]`, `[a\-z]`, `[\d-z]`,
		`[a-\d]`, `[[]`, `[[.a.]]`, `[[=a=]]`, `[`, `]`, `}`, `{`, `a{`, `a}`, `)`, `(`, `(?`, `(?:`, `*`, `+`, `?`,
	}
}

func nestingAndSizePatterns() []string {
	depths := []int{1, 2, 10, 100, 500, 999, 1000, 1001, 1002, 2000, 10000}
	sizes := []int{1, 100, 1000, 10000, 100000}
	patterns := make([]string, 0, len(depths)*8+len(sizes)*8)

	for _, depth := range depths {
		patterns = append(patterns,
			strings.Repeat("(", depth)+"a"+strings.Repeat(")", depth),
			strings.Repeat("(?:", depth)+"a"+strings.Repeat(")", depth),
			strings.Repeat("(a", depth)+strings.Repeat(")", depth),
			strings.Repeat("(a|", depth)+"b"+strings.Repeat(")", depth),
			strings.Repeat("[", depth)+strings.Repeat("]", depth),
			strings.Repeat("(a*", depth)+strings.Repeat(")*", depth),
			strings.Repeat("a*", depth),
			strings.Repeat("(", depth)+strings.Repeat(")", depth),
		)
	}

	for _, size := range sizes {
		words := make([]string, size/5+1)
		for index := range words {
			words[index] = "w" + strconv.Itoa(index)
		}

		patterns = append(patterns,
			strings.Repeat("a", size),
			strings.Repeat("é", size),
			strings.Join(words, "|"),
			"(?:"+strings.Join(words, ")|(?:")+")",
			strings.Repeat("[a-z]", size/10+1),
			strings.Repeat(`\pL`, size/10+1),
			strings.Repeat("a?", size/10+1)+strings.Repeat("a", size/10+1),
			"("+strings.Repeat("a", size)+")",
		)
	}

	return patterns
}

// generatedPatterns builds random token sequences from a fixed seed, so every run covers the same patterns.
func generatedPatterns() []string {
	tokens := []string{
		"a", "b", ".", "^", "$", "|", "(", ")", "(?:", "(?i)", "(?s:", "(?P<n>", "(?<m>", "[", "]", "[^", "-", "\\",
		"*", "+", "?", "{", "}", "{2}", "{1,3}", "{1000}", "{1001}", ",", `\d`, `\pL`, `\p{Greek}`, `\P{^N}`,
		`[:alpha:]`, `\b`, `\B`, `\A`, `\z`, `\Z`, `\C`, `\Q`, `\E`, `\x41`, `\x{41}`, `\0`, `\1`, `\12`, `\n`, `\_`,
		"é", "\xff", "(?=", "(?!", "(?<=", "(?P=n)", "*?", "+?", "??", "{2,}?", "#", " ", "(?x)", "(?U)", "(?m)",
		"(?-i)",
	}
	state := uint64(1)
	next := func(limit int) int {
		// A linear congruential generator (Knuth's MMIX constants) keeps the corpus identical on every run.
		state = state*6364136223846793005 + 1442695040888963407

		return int(state>>33) % limit
	}
	patterns := make([]string, 0, 60000)

	for range cap(patterns) {
		var pattern strings.Builder

		for range 1 + next(10) {
			pattern.WriteString(tokens[next(len(tokens))])
		}

		patterns = append(patterns, pattern.String())
	}

	return patterns
}
