package regexp

import (
	"testing"
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
