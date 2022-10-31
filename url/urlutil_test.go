package urlutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	// full url
	U, err := Parse("http://127.0.0.1/a")
	require.Nil(t, err, "could not parse url")
	require.Equal(t, "http", U.Scheme, "different scheme")
	require.Equal(t, "127.0.0.1", U.Host, "different host")
	require.Equal(t, "/a", U.RequestURI, "different request uri")

	// full url with port
	U, err = Parse("http://127.0.0.1:1000/a")
	require.Nil(t, err, "could not parse url")
	require.Equal(t, "http", U.Scheme, "different scheme")
	require.Equal(t, "127.0.0.1", U.Host, "different host")
	require.Equal(t, "1000", U.Port, "different host")
	require.Equal(t, "/a", U.RequestURI, "different request uri")

	// partial url without port
	U, err = Parse("a.b.c.d")
	require.Nil(t, err, "could not parse url")
	require.Equal(t, "https", U.Scheme, "different scheme")
	require.Equal(t, "a.b.c.d", U.Host, "different host")
	require.Equal(t, "443", U.Port, "different host")
	require.Equal(t, "", U.RequestURI, "different request uri")

	// partial url with protocol and no port
	U, err = Parse("https://a.b.c.d")
	require.Nil(t, err, "could not parse url")
	require.Equal(t, "https", U.Scheme, "different scheme")
	require.Equal(t, "a.b.c.d", U.Host, "different host")
	require.Equal(t, "443", U.Port, "different host")
	require.Equal(t, "", U.RequestURI, "different request uri")

	// replacing port
	U, err = Parse("https://a.b.c.d")
	require.Nil(t, err, "could not parse url")
	U.Port = "15000"
	require.Equal(t, "https://a.b.c.d:15000", U.String(), "port not replaced")

	// replacing port
	U, err = Parse("https://a.b.c.d//d")
	require.Nil(t, err, "could not parse url")
	require.Equal(t, "https://a.b.c.d:443//d", U.String(), "unexpected url")
	
	// fragmented url
	U, err = Parse("http://127.0.0.1/#a")
	require.Nil(t, err, "could not parse url")
	require.Equal(t, "http", U.Scheme, "different scheme")
	require.Equal(t, "127.0.0.1", U.Host, "different host")
	require.Equal(t, "a", U.Fragment, "different fragment")
	require.Equal(t, "http://127.0.0.1:80/#a", U.String(), "different full url")
}
