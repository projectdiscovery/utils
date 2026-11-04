package updateutils

import (
	"crypto/x509"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVersionCheckClientRejectsUntrustedCertificate(t *testing.T) {
	var received atomic.Bool
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		received.Store(true)
		_, _ = w.Write([]byte(`{"version":"v9.9.9"}`))
	}))
	t.Cleanup(srv.Close)

	resp, err := DefaultHttpClient.Get(srv.URL + "/api/v1/tools/test?machine_id=test")
	if resp != nil {
		_ = resp.Body.Close()
	}
	require.Error(t, err)
	var unknownAuthority x509.UnknownAuthorityError
	require.ErrorAs(t, err, &unknownAuthority)
	require.False(t, received.Load(), "version check must not reach a server with an untrusted certificate")
}
