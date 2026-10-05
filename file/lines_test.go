package fileutil

import (
	"bufio"
	"errors"
	"io"
	"io/fs"
	"iter"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func collectLines(t *testing.T, seq iter.Seq2[string, error]) []string {
	t.Helper()
	var out []string
	for v, err := range seq {
		require.NoError(t, err)
		out = append(out, v)
	}
	return out
}

func writeTempFile(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "lines.txt")
	require.NoError(t, os.WriteFile(path, []byte(body), 0o600))
	return path
}

func TestLines_Default_EmitsRawLines(t *testing.T) {
	path := writeTempFile(t, "alpha\nbeta\n  gamma  \n\nepsilon\n")
	got := collectLines(t, Lines(path))
	require.Equal(t, []string{"alpha", "beta", "  gamma  ", "", "epsilon"}, got)
}

func TestLines_WithTrimSpace(t *testing.T) {
	path := writeTempFile(t, "  alpha  \n\tbeta\t\n")
	got := collectLines(t, Lines(path, WithTrimSpace()))
	require.Equal(t, []string{"alpha", "beta"}, got)
}

func TestLines_WithSkipEmpty(t *testing.T) {
	path := writeTempFile(t, "alpha\n\nbeta\n\n\n")
	got := collectLines(t, Lines(path, WithSkipEmpty()))
	require.Equal(t, []string{"alpha", "beta"}, got)
}

func TestLines_WithTrimSpace_SkipEmpty_DropsBlankLines(t *testing.T) {
	path := writeTempFile(t, "alpha\n   \nbeta\n")
	got := collectLines(t, Lines(path, WithTrimSpace(), WithSkipEmpty()))
	require.Equal(t, []string{"alpha", "beta"}, got)
}

func TestLines_WithSplit_Comma(t *testing.T) {
	path := writeTempFile(t, "1.1.1.1,8.8.8.8\n9.9.9.9\n")
	got := collectLines(t, Lines(path, WithSplit(',')))
	require.Equal(t, []string{"1.1.1.1", "8.8.8.8", "9.9.9.9"}, got)
}

func TestLines_WithSplit_MultipleSeparators(t *testing.T) {
	path := writeTempFile(t, "a,b;c\td\n")
	got := collectLines(t, Lines(path, WithSplit(',', ';', '\t')))
	require.Equal(t, []string{"a", "b", "c", "d"}, got)
}

func TestLines_ResolverFileScenario(t *testing.T) {
	// resolver-file scenario: comma-separated entries with whitespace and
	// blanks; this is what the original PR was trying to add a one-shot helper for.
	path := writeTempFile(t, "1.1.1.1, 8.8.8.8\n9.9.9.9\n  ,  ,  \n10.10.10.10 ,11.11.11.11\n")
	got := collectLines(t, Lines(path,
		WithSplit(','),
		WithTrimSpace(),
		WithSkipEmpty(),
	))
	require.Equal(t, []string{"1.1.1.1", "8.8.8.8", "9.9.9.9", "10.10.10.10", "11.11.11.11"}, got)
}

func TestLines_WithFilter_DropsComments(t *testing.T) {
	path := writeTempFile(t, "alpha\n# comment\nbeta\n# another\n")
	got := collectLines(t, Lines(path,
		WithFilter(func(s string) bool { return !strings.HasPrefix(s, "#") }),
	))
	require.Equal(t, []string{"alpha", "beta"}, got)
}

func TestLines_WithBufferSize(t *testing.T) {
	path := writeTempFile(t, "short\n"+strings.Repeat("x", 1024)+"\n")
	got := collectLines(t, Lines(path, WithBufferSize(2048)))
	require.Len(t, got, 2)
	require.Equal(t, "short", got[0])
	require.Len(t, got[1], 1024)
}

func TestLines_MissingFile_YieldsErrorPair(t *testing.T) {
	var values []string
	var gotErr error
	for v, err := range Lines("/no/such/file.txt") {
		if err != nil {
			gotErr = err
			continue
		}
		values = append(values, v)
	}
	require.Empty(t, values)
	require.Error(t, gotErr)
	require.True(t, errors.Is(gotErr, fs.ErrNotExist), "expected fs.ErrNotExist, got %v", gotErr)
}

func TestLines_BreakStopsIterationEarly(t *testing.T) {
	path := writeTempFile(t, "a\nb\nc\nd\n")
	var seen []string
	for v, err := range Lines(path) {
		require.NoError(t, err)
		seen = append(seen, v)
		if len(seen) == 2 {
			break
		}
	}
	require.Equal(t, []string{"a", "b"}, seen)
}

func TestLinesReader_Default(t *testing.T) {
	r := strings.NewReader("alpha\nbeta\n  gamma  \n\n")
	got := collectLines(t, LinesReader(r))
	require.Equal(t, []string{"alpha", "beta", "  gamma  ", ""}, got)
}

func TestLinesReader_AllOptionsCombined(t *testing.T) {
	r := strings.NewReader("# header\n1.1.1.1, 8.8.8.8\n\n# tail\n")
	got := collectLines(t, LinesReader(r,
		WithSplit(','),
		WithTrimSpace(),
		WithSkipEmpty(),
		WithFilter(func(s string) bool { return !strings.HasPrefix(s, "#") }),
	))
	require.Equal(t, []string{"1.1.1.1", "8.8.8.8"}, got)
}

func TestLinesReader_PropagatesScannerError(t *testing.T) {
	// A reader that always errors should surface the error as a final pair.
	r := errReader{err: io.ErrUnexpectedEOF}
	var seen []string
	var gotErr error
	for v, err := range LinesReader(r) {
		if err != nil {
			gotErr = err
			continue
		}
		seen = append(seen, v)
	}
	require.Empty(t, seen)
	require.ErrorIs(t, gotErr, io.ErrUnexpectedEOF)
}

func TestReadLinesStream_Options(t *testing.T) {
	longLine := strings.Repeat("x", 70*1024)
	tests := []struct {
		name    string
		body    string
		opts    []LineOption
		want    []string
		wantErr error
	}{
		{
			name: "default preserves whitespace and empty lines",
			body: "  alpha  \n\n # comment\nbeta\n",
			want: []string{"  alpha  ", "", " # comment", "beta"},
		},
		{
			name: "shared trim and skip options",
			body: "  alpha  \n \t \n\nbeta\n",
			opts: []LineOption{WithTrimSpace(), WithSkipEmpty()},
			want: []string{"alpha", "beta"},
		},
		{
			name: "comments are checked without trimming output",
			body: "  # comment\n  alpha  \n\nalpha # inline\n",
			opts: []LineOption{WithComment("#")},
			want: []string{"  alpha  ", "", "alpha # inline"},
		},
		{
			name: "comment trim and skip options combined",
			body: "  # comment\n  alpha  \n \t \n\nbeta\n",
			opts: []LineOption{WithComment("#"), WithTrimSpace(), WithSkipEmpty()},
			want: []string{"alpha", "beta"},
		},
		{
			name: "empty comment prefix keeps all lines",
			body: "# comment\n\n",
			opts: []LineOption{WithComment("")},
			want: []string{"# comment", ""},
		},
		{
			name: "shared buffer option accepts large lines",
			body: longLine + "\n",
			opts: []LineOption{WithBufferSize(128 * 1024)},
			want: []string{longLine},
		},
		{
			name:    "default buffer reports scanner error",
			body:    "alpha\n" + longLine + "\n",
			want:    []string{"alpha"},
			wantErr: bufio.ErrTooLong,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeTempFile(t, tt.body)
			lines, errs := ReadLinesStream(path, tt.opts...)
			var got []string
			for line := range lines {
				got = append(got, line)
			}
			var gotErrors []error
			for err := range errs {
				gotErrors = append(gotErrors, err)
			}
			require.Equal(t, tt.want, got)
			if tt.wantErr == nil {
				require.Empty(t, gotErrors)
			} else {
				require.Len(t, gotErrors, 1)
				require.ErrorIs(t, gotErrors[0], tt.wantErr)
			}
		})
	}
}

type errReader struct{ err error }

func (e errReader) Read(p []byte) (int, error) { return 0, e.err }
