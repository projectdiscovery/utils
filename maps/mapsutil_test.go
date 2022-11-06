package mapsutil

import (
	"crypto/tls"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMergeMaps(t *testing.T) {
	m1Str := map[string]interface{}{"a": 1, "b": 2}
	m2Str := map[string]interface{}{"b": 2, "c": 3}
	rStr := map[string]interface{}{"a": 1, "b": 2, "c": 3}
	rrStr := Merge(m1Str, m2Str)
	require.EqualValues(t, rStr, rrStr)

	m1Int := map[int]interface{}{1: 1, 2: 2}
	m2Int := map[int]interface{}{1: 1, 2: 2, 3: 3, 4: 4}
	m3Int := map[int]interface{}{1: 1, 5: 5}
	rInt := map[int]interface{}{1: 1, 2: 2, 3: 3, 4: 4, 5: 5}
	rrInt := Merge(m1Int, m2Int, m3Int)
	require.EqualValues(t, rInt, rrInt)
}

var (
	req = &http.Request{
		Method: "POST",
		URL: &url.URL{
			Scheme: "http",
			Host:   "test.ts",
			Path:   "/",
		},
		Host:             "test.ts",
		Proto:            "HTTP/1.1",
		ProtoMajor:       2,
		ProtoMinor:       1,
		Header:           http.Header{},
		Body:             io.NopCloser(strings.NewReader("test")),
		ContentLength:    1000,
		TransferEncoding: []string{""},
		Close:            true,
		Trailer:          http.Header{},
		TLS:              &tls.ConnectionState{},
	}

	resp = &http.Response{
		Status:           "200 OK",
		StatusCode:       200,
		Proto:            "HTTP/1.1",
		ProtoMajor:       2,
		ProtoMinor:       1,
		Header:           http.Header{},
		Body:             io.NopCloser(strings.NewReader("test")),
		ContentLength:    1000,
		TransferEncoding: []string{""},
		Close:            true,
		Uncompressed:     false,
		Trailer:          http.Header{},
		Request:          &http.Request{},
		TLS:              &tls.ConnectionState{},
	}
)

func TestHTTPToMap(t *testing.T) {
	bufBody := new(strings.Builder)
	// nolint:errcheck
	io.Copy(bufBody, resp.Body)

	bufHeaders := new(strings.Builder)
	// nolint:errcheck
	io.Copy(bufHeaders, resp.Body)

	m := HTTPToMap(resp, bufBody.String(), bufHeaders.String(), time.Duration(2), "")

	require.NotNil(t, m)
	require.NotEmpty(t, m)
}

func TestDNSToMap(t *testing.T) {
	// not implemented
}

func TestHTTPRequestToMap(t *testing.T) {
	m, err := HTTPRequestToMap(req)

	require.Nil(t, err)
	require.NotNil(t, m)
	require.NotEmpty(t, m)
}

func TestHTTPResponseToMap(t *testing.T) {
	m, err := HTTPResponseToMap(resp)

	require.Nil(t, err)
	require.NotNil(t, m)
	require.NotEmpty(t, m)
}

func TestGetKeys(t *testing.T) {
	t.Run("GetKeys(empty)", func(t *testing.T) {
		got := GetKeys(map[string]interface{}{})
		require.Empty(t, got)
	})

	t.Run("GetKeys(string)", func(t *testing.T) {
		got := GetKeys(map[string]interface{}{"a": "a", "b": "b"})
		require.ElementsMatch(t, []string{"a", "b"}, got)
	})

	t.Run("GetKeys(int)", func(t *testing.T) {
		got := GetKeys(map[int]interface{}{1: "a", 2: "b"})
		require.ElementsMatch(t, []int{1, 2}, got)
	})

	t.Run("GetKeys(bool)", func(t *testing.T) {
		got := GetKeys(map[bool]interface{}{true: "a", false: "b"})
		require.ElementsMatch(t, []bool{true, false}, got)
	})
}

func TestGetValues(t *testing.T) {
	t.Run("GetValues(empty)", func(t *testing.T) {
		got := GetValues(map[string]interface{}{})
		require.Empty(t, got)
	})

	t.Run("GetValues(string)", func(t *testing.T) {
		got := GetValues(map[string]interface{}{"a": "a", "b": "b"})
		require.ElementsMatch(t, []interface{}{"a", "b"}, got)
	})

	t.Run("GetValues(int)", func(t *testing.T) {
		got := GetValues(map[string]interface{}{"a": 1, "b": 2})
		require.ElementsMatch(t, []interface{}{1, 2}, got)
	})

	t.Run("GetValues(bool)", func(t *testing.T) {
		got := GetValues(map[string]interface{}{"a": true, "b": false})
		require.ElementsMatch(t, []interface{}{true, false}, got)
	})
}

func TestDifference(t *testing.T) {
	t.Run("Difference(empty)", func(t *testing.T) {
		got := Difference(map[string]interface{}{}, []string{}...)
		require.EqualValues(t, map[string]interface{}{}, got)
	})

	t.Run("Difference(string)", func(t *testing.T) {
		got := Difference(map[string]interface{}{"a": 1, "b": 2, "c": 3}, []string{"a"}...)
		require.EqualValues(t, map[string]interface{}{"b": 2, "c": 3}, got)
	})

	t.Run("Difference(int)", func(t *testing.T) {
		got := Difference(map[int]interface{}{1: "a", 2: "b", 3: "c"}, []int{1}...)
		require.EqualValues(t, map[int]interface{}{2: "b", 3: "c"}, got)
	})

	t.Run("Difference(bool)", func(t *testing.T) {
		got := Difference(map[bool]interface{}{true: 1, false: 2}, []bool{true}...)
		require.EqualValues(t, map[bool]interface{}{false: 2}, got)
	})
}

func TestSliceToMap(t *testing.T) {
	t.Run("SliceToMap(empty)", func(t *testing.T) {
		got := SliceToMap([]string{}, "")
		require.EqualValues(t, map[string]string{}, got)
	})

	t.Run("SliceToMap(string)", func(t *testing.T) {
		got := SliceToMap([]string{"a", "b", "c", "d"}, "")
		require.EqualValues(t, map[string]string{"a": "b", "c": "d"}, got)
	})

	t.Run("SliceToMap(string odd)", func(t *testing.T) {
		got := SliceToMap([]string{"a", "b", "c"}, "")
		require.EqualValues(t, map[string]string{"a": "b", "c": ""}, got)
	})

	t.Run("SliceToMap(int odd)", func(t *testing.T) {
		got := SliceToMap([]int{1, 2, 3}, 0)
		require.EqualValues(t, map[int]int{1: 2, 3: 0}, got)
	})
}

func TestIsEmpty(t *testing.T) {
	t.Run("IsEmpty(empty)", func(t *testing.T) {
		got := IsEmpty(map[string]string{})
		require.EqualValues(t, true, got)
	})

	t.Run("IsEmpty(string)", func(t *testing.T) {
		got := IsEmpty(map[string]string{"a": "b"})
		require.EqualValues(t, false, got)
	})

	t.Run("IsEmpty(int)", func(t *testing.T) {
		got := IsEmpty(map[int]int{1: 2})
		require.EqualValues(t, false, got)
	})
}

func TestClear(t *testing.T) {
	t.Run("Clear(nil)", func(t *testing.T) {
		var m map[string]string
		Clear(m)
		require.Empty(t, m)
	})

	t.Run("Clear(m)", func(t *testing.T) {
		m := map[string]string{"a": "a", "b": "b"}
		Clear(m)
		require.Empty(t, m)
	})

	t.Run("Clear(m1,m2)", func(t *testing.T) {
		m1 := map[string]string{"a": "a", "b": "b"}
		m2 := map[string]string{"a": "a", "b": "b"}
		Clear(m1, m2)
		require.Empty(t, m1)
		require.Empty(t, m2)
	})
}
