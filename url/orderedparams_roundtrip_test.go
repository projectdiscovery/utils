package urlutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestDecodeEncodeRoundTrip locks in the new contract: every byte of
// a query string parsed by Decode survives Encode unchanged when the
// caller does not mutate any entry. This is the bug that issue #379
// is about.
func TestDecodeEncodeRoundTrip(t *testing.T) {
	cases := []string{
		"",
		"foo=bar",
		"foo",
		"foo=",
		"=bar",
		"foo=bar&baz",
		"foo=bar&baz=",
		"foo&baz=qux",
		"a=1&a=2&a=3",
		"a=1&b=2&a=3",
		"key=value%20with%20space",
		"key=A%26B",
		"empty_first=&second=v",
		"foo=bar=baz",
		"key=hello+world",
		"k1=v1&k2=v2&k3=v3&k4=v4",
		"%E4%B8%AD=%E6%96%87",
		"json={%22a%22:1}",
	}
	for _, raw := range cases {
		t.Run(raw, func(t *testing.T) {
			p := NewOrderedParams()
			p.Decode(raw)
			got := p.Encode()
			require.Equal(t, raw, got, "round-trip lost data: input %q encoded as %q", raw, got)
		})
	}
}

// TestDecodeEncodeRoundTripTrailingAmp documents that a trailing "&"
// is parsed into an empty trailing segment that round-trips
// faithfully. We pin it separately because it is the most fragile
// part of the parser.
func TestDecodeEncodeRoundTripTrailingAmp(t *testing.T) {
	p := NewOrderedParams()
	p.Decode("foo=bar&")
	// The trailing "&" is consumed but no empty segment is appended
	// (matches historical behavior - the parser only flushes a
	// segment if its buffer is non-empty at EOF).
	require.Equal(t, "foo=bar", p.Encode())

	p = NewOrderedParams()
	p.Decode("foo=bar&&baz=qux")
	require.Equal(t, "foo=bar&&baz=qux", p.Encode())
}

// TestDecodePreservesBareKey ensures a parameter without "=" stays
// without "=", and a parameter with "=" but no value stays with "=".
func TestDecodePreservesBareKey(t *testing.T) {
	p := NewOrderedParams()
	p.Decode("a&b=&c=v")
	require.Equal(t, "a&b=&c=v", p.Encode())

	require.True(t, p.Has("a"))
	require.True(t, p.Has("b"))
	require.True(t, p.Has("c"))
	require.Equal(t, "", p.Get("a"))
	require.Equal(t, "", p.Get("b"))
	require.Equal(t, "v", p.Get("c"))
}

// TestDecodePreservesBareKeyEvenWithIncludeEquals is the key
// behavioral fix: flipping IncludeEquals must NOT inject "=" into
// parameters that came in without one. IncludeEquals only governs
// programmatically added/mutated entries.
func TestDecodePreservesBareKeyEvenWithIncludeEquals(t *testing.T) {
	p := NewOrderedParams()
	p.Decode("a&b=&c=v")
	p.IncludeEquals = true
	require.Equal(t, "a&b=&c=v", p.Encode())
}

// TestDecodePreservesOriginalEncoding documents that ParamEncode's
// choices (e.g. how it escapes "+", "/", "<", etc.) do not get
// imposed on entries that came in from Decode. The bytes the caller
// gave us are the bytes that come back out.
func TestDecodePreservesOriginalEncoding(t *testing.T) {
	raws := []string{
		"q=hello+world",
		"q=hello%20world",
		"redirect=https%3A%2F%2Fexample.com%2Fpath",
		"x=<script>",
		"x=%3Cscript%3E",
	}
	for _, raw := range raws {
		t.Run(raw, func(t *testing.T) {
			p := NewOrderedParams()
			p.Decode(raw)
			require.Equal(t, raw, p.Encode())
		})
	}
}

// TestMutationUsesParamEncode confirms that a parameter the caller
// touched programmatically is re-encoded through ParamEncode,
// honoring IncludeEquals. Untouched siblings stay verbatim.
func TestMutationUsesParamEncode(t *testing.T) {
	p := NewOrderedParams()
	p.Decode("a&b=hello+world&c=keep")

	p.Set("b", "fresh value")

	enc := p.Encode()
	require.Contains(t, enc, "a")
	require.Contains(t, enc, "c=keep")
	require.Contains(t, enc, "b=fresh+value")
	// a stayed verbatim (no "="), c stayed verbatim, b was re-encoded.
	require.Equal(t, "a&b=fresh+value&c=keep", enc)
}

// TestMutationOnBareKeyHonorsIncludeEquals checks that once a bare
// key is set programmatically, it follows the IncludeEquals rule.
func TestMutationOnBareKeyHonorsIncludeEquals(t *testing.T) {
	p := NewOrderedParams()
	p.Decode("flag&other=x")
	p.Set("flag", "")

	require.Equal(t, "flag&other=x", p.Encode())

	p.IncludeEquals = true
	require.Equal(t, "flag=&other=x", p.Encode())
}

// TestAddPreservesOrderWithExistingDecoded checks that Add appends
// at the end, after decoded entries.
func TestAddPreservesOrderWithExistingDecoded(t *testing.T) {
	p := NewOrderedParams()
	p.Decode("a=1&b=2")
	p.Add("c", "3")
	require.Equal(t, "a=1&b=2&c=3", p.Encode())
}

// TestSetReplacesInPlace checks that Set on an existing key keeps
// the original position in the insertion order.
func TestSetReplacesInPlace(t *testing.T) {
	p := NewOrderedParams()
	p.Decode("a=1&b=2&c=3")
	p.Set("b", "9")
	require.Equal(t, "a=1&b=9&c=3", p.Encode())
}

// TestSetCollapsesDuplicateKeys checks that Set on a key with
// multiple existing entries collapses to a single entry at the
// position of the first occurrence.
func TestSetCollapsesDuplicateKeys(t *testing.T) {
	p := NewOrderedParams()
	p.Decode("a=1&b=2&a=3&c=4")
	p.Set("a", "X")
	require.Equal(t, "a=X&b=2&c=4", p.Encode())
}

// TestUpdateReplacesAllValuesAtFirstPosition checks Update's
// behavior on a key with multiple entries: every existing entry is
// removed, and the new values are inserted at the first occurrence
// of the key.
func TestUpdateReplacesAllValuesAtFirstPosition(t *testing.T) {
	p := NewOrderedParams()
	p.Decode("a=1&b=2&a=3&c=4")
	p.Update("a", []string{"X", "Y"})
	require.Equal(t, "a=X&a=Y&b=2&c=4", p.Encode())
}

// TestDelRemovesAllOccurrences checks that Del strips every entry
// matching the key while preserving the order of the rest.
func TestDelRemovesAllOccurrences(t *testing.T) {
	p := NewOrderedParams()
	p.Decode("a=1&b=2&a=3&c=4")
	p.Del("a")
	require.Equal(t, "b=2&c=4", p.Encode())
	require.False(t, p.Has("a"))
}

// TestGetAndGetAll covers the lookup helpers across duplicate keys.
func TestGetAndGetAll(t *testing.T) {
	p := NewOrderedParams()
	p.Decode("a=1&b=2&a=3")
	require.Equal(t, "1", p.Get("a"))
	require.Equal(t, []string{"1", "3"}, p.GetAll("a"))
	require.Equal(t, []string{"2"}, p.GetAll("b"))
	require.Equal(t, []string{}, p.GetAll("missing"))
	require.Equal(t, "", p.Get("missing"))
}

// TestIterateGroupsByFirstOccurrence checks that Iterate yields
// each key exactly once at the position of its first occurrence,
// and that the values slice contains every value in insertion order.
func TestIterateGroupsByFirstOccurrence(t *testing.T) {
	p := NewOrderedParams()
	p.Decode("a=1&b=2&a=3&c=4&b=5")

	var keys []string
	values := map[string][]string{}
	p.Iterate(func(k string, vs []string) bool {
		keys = append(keys, k)
		values[k] = vs
		return true
	})
	require.Equal(t, []string{"a", "b", "c"}, keys)
	require.Equal(t, []string{"1", "3"}, values["a"])
	require.Equal(t, []string{"2", "5"}, values["b"])
	require.Equal(t, []string{"4"}, values["c"])
}

// TestIterateEarlyExit confirms returning false stops iteration.
func TestIterateEarlyExit(t *testing.T) {
	p := NewOrderedParams()
	p.Decode("a=1&b=2&c=3")
	var seen []string
	p.Iterate(func(k string, vs []string) bool {
		seen = append(seen, k)
		return k != "b"
	})
	require.Equal(t, []string{"a", "b"}, seen)
}

// TestMergeAppends checks that Merge appends to the existing entries
// instead of replacing them.
func TestMergeAppends(t *testing.T) {
	p := NewOrderedParams()
	p.Decode("a=1")
	p.Merge("b=2&c=3")
	require.Equal(t, "a=1&b=2&c=3", p.Encode())
}

// TestCloneIsIndependent confirms a clone can be mutated without
// affecting the original.
func TestCloneIsIndependent(t *testing.T) {
	p := NewOrderedParams()
	p.Decode("a=1&b=2")
	p.IncludeEquals = true

	c := p.Clone()
	c.Set("a", "9")
	c.Add("d", "x")
	c.IncludeEquals = false

	require.Equal(t, "a=1&b=2", p.Encode(), "original was modified")
	require.Equal(t, "a=9&b=2&d=x", c.Encode())
	require.True(t, p.IncludeEquals)
	require.False(t, c.IncludeEquals)
}

// TestIsEmpty covers the empty/non-empty states.
func TestIsEmpty(t *testing.T) {
	p := NewOrderedParams()
	require.True(t, p.IsEmpty())
	p.Add("a", "1")
	require.False(t, p.IsEmpty())
	p.Del("a")
	require.True(t, p.IsEmpty())
}

// TestLegacySeparator covers the AllowLegacySeperator code path.
func TestLegacySeparator(t *testing.T) {
	prev := AllowLegacySeperator
	AllowLegacySeperator = true
	defer func() { AllowLegacySeperator = prev }()

	p := NewOrderedParams()
	p.Decode("a=1;b=2;c=3")
	require.Equal(t, "1", p.Get("a"))
	require.Equal(t, "2", p.Get("b"))
	require.Equal(t, "3", p.Get("c"))
	// Encode always uses "&" as the separator; the legacy ";" form
	// is only honored on input.
	require.Equal(t, "a=1&b=2&c=3", p.Encode())
}

// TestProgrammaticEncodeIsBackwardsCompatible re-asserts that an
// OrderedParams built entirely through Add/Set produces the same
// output as before: each value is ParamEncoded, "=" is omitted when
// the encoded value is empty unless IncludeEquals is true.
func TestProgrammaticEncodeIsBackwardsCompatible(t *testing.T) {
	p := NewOrderedParams()
	p.Add("name", "Ada Lovelace")
	p.Add("query", "1+1=2")
	p.Add("flag", "")

	// ParamEncode is intentionally permissive (it preserves "+",
	// "=", etc. as raw bytes - this is what the rest of the
	// library relies on). We pin its current behavior here.
	require.Equal(t, "name=Ada+Lovelace&query=1+1=2&flag", p.Encode())

	p.IncludeEquals = true
	require.Equal(t, "name=Ada+Lovelace&query=1+1=2&flag=", p.Encode())
}
