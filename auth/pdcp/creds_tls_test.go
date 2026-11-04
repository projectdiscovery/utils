package pdcp

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

func newProfileServer(t *testing.T, gotKey *atomic.Value) *httptest.Server {
	t.Helper()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey.Store(r.Header.Get(ApiKeyHeaderName))
		_, _ = w.Write([]byte(`{"name":"test","email":"test@example.com"}`))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestValidateAPIKeyRejectsUntrustedCertificate(t *testing.T) {
	var gotKey atomic.Value
	srv := newProfileServer(t, &gotKey)

	h := &PDCPCredHandler{}
	creds, err := h.ValidateAPIKey("test-key", srv.URL, "test")
	require.Error(t, err)
	require.Nil(t, creds)
	var unknownAuthority x509.UnknownAuthorityError
	require.ErrorAs(t, err, &unknownAuthority)
	require.Nil(t, gotKey.Load(), "API key must not reach a server with an untrusted certificate")
}

func TestValidateAPIKeyTrustedCertificate(t *testing.T) {
	var gotKey atomic.Value
	srv := newProfileServer(t, &gotKey)

	pool := x509.NewCertPool()
	pool.AddCert(srv.Certificate())
	orig := apiClient
	t.Cleanup(func() { apiClient = orig })
	apiClient = newAPIClient()
	transport, ok := apiClient.HTTPClient.Transport.(*http.Transport)
	require.True(t, ok)
	transport.TLSClientConfig = &tls.Config{RootCAs: pool, MinVersion: tls.VersionTLS12}

	h := &PDCPCredHandler{}
	creds, err := h.ValidateAPIKey("test-key", srv.URL, "test")
	require.NoError(t, err)
	require.Equal(t, "test@example.com", creds.Email)
	require.Equal(t, "test-key", gotKey.Load())
}
