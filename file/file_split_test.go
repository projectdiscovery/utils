package fileutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func collectChan(c chan string) []string {
	var got []string
	for v := range c {
		got = append(got, v)
	}
	return got
}

func TestSplitLineByRunes_NoSeparators(t *testing.T) {
	require.Equal(t, []string{"single"}, splitLineByRunes(" single ", nil))
	require.Nil(t, splitLineByRunes("   ", nil))
}

func TestSplitLineByRunes_Comma(t *testing.T) {
	require.Equal(t, []string{"a", "b", "c"}, splitLineByRunes("a,b,c", []rune{','}))
	require.Equal(t, []string{"a", "b"}, splitLineByRunes("  a , ,b  ", []rune{','}))
	require.Empty(t, splitLineByRunes(",,,", []rune{','}))
}

func TestSplitLineByRunes_MultipleSeparators(t *testing.T) {
	require.Equal(t, []string{"a", "b", "c", "d"}, splitLineByRunes("a,b;c\td", []rune{',', ';', '\t'}))
}

func TestReadFileWithReaderSplit_NoSeparators(t *testing.T) {
	r := strings.NewReader("alpha\nbeta\n  gamma  \n\n")
	ch, err := ReadFileWithReaderSplit(r)
	require.NoError(t, err)
	require.Equal(t, []string{"alpha", "beta", "gamma"}, collectChan(ch))
}

func TestReadFileWithReaderSplit_Comma(t *testing.T) {
	r := strings.NewReader("1.1.1.1, 8.8.8.8\n9.9.9.9\n,,\n# comment\n")
	ch, err := ReadFileWithReaderSplit(r, ',')
	require.NoError(t, err)
	require.Equal(t, []string{"1.1.1.1", "8.8.8.8", "9.9.9.9", "# comment"}, collectChan(ch))
}

func TestReadFileSplit_RoundTrip(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "resolvers.txt")
	body := "1.1.1.1,8.8.8.8\n9.9.9.9\n  ,  ,  \n10.10.10.10 ,11.11.11.11\n"
	require.NoError(t, os.WriteFile(path, []byte(body), 0o600))

	ch, err := ReadFileSplit(path, ',')
	require.NoError(t, err)
	require.Equal(t, []string{"1.1.1.1", "8.8.8.8", "9.9.9.9", "10.10.10.10", "11.11.11.11"}, collectChan(ch))
}

func TestReadFileSplit_MissingFile(t *testing.T) {
	_, err := ReadFileSplit("/no/such/file.txt", ',')
	require.Error(t, err)
}

func TestReadFileSplit_NoSeparatorEqualsReadFile(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "lines.txt")
	body := "first\nsecond\n   third   \n\nfourth\n"
	require.NoError(t, os.WriteFile(path, []byte(body), 0o600))

	ch, err := ReadFileSplit(path)
	require.NoError(t, err)
	require.Equal(t, []string{"first", "second", "third", "fourth"}, collectChan(ch))
}
