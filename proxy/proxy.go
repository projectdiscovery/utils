// Package proxyutils validates proxy URLs and probes proxies to find one
// that accepts connections.
package proxyutils

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/projectdiscovery/utils/errkit"
	"github.com/remeh/sizedwaitgroup"
)

type proxyResult struct {
	AliveProxy string
	Error      error
}

// ProxyProbeConcurrency (default 8)
var ProxyProbeConcurrency = 8

const (
	SOCKS5 = "socks5"
	HTTP   = "http"
	HTTPS  = "https"
)

// GetAnyAliveProxy takes proxies as input and returns the first alive proxy
// or returns error if all of them not alive
func GetAnyAliveProxy(timeoutInSec int, proxies ...string) (string, error) {
	ctx, cancel := context.WithCancel(context.Background())
	// Stops the producer and aborts probes still dialing once we return.
	defer cancel()

	probes := sizedwaitgroup.New(ProxyProbeConcurrency)

	// Every proxy yields at most one result, so a buffer of len(proxies)
	// means no sender ever blocks, even after the receiver has returned.
	resChan := make(chan proxyResult, len(proxies))

	go func() {
		for _, v := range proxies {
			proxy, err := GetProxyURL(v)
			if err != nil {
				resChan <- proxyResult{AliveProxy: "", Error: err}

				continue
			}

			if probes.AddWithContext(ctx) != nil {
				break
			}

			go func(proxyAddr url.URL) {
				defer probes.Done()

				resChan <- testProxyConn(ctx, proxyAddr, timeoutInSec)
			}(proxy)
		}

		probes.Wait()
		close(resChan)
	}()

	errstack := []string{}

	for result := range resChan {
		if result.AliveProxy != "" {
			// found alive proxy return now
			return result.AliveProxy, nil
		} else if result.Error != nil {
			errstack = append(errstack, result.Error.Error())
		}
	}

	// all proxies are dead
	return "", errkit.Newf("all proxies are dead got : %v", strings.Join(errstack, " : "))
}

// dial and test if proxy is open
func testProxyConn(ctx context.Context, proxyAddr url.URL, timeoutInSec int) proxyResult {
	var dialer net.Dialer

	dialer.Timeout = time.Duration(timeoutInSec) * time.Second
	addr := fmt.Sprintf("%s:%s", proxyAddr.Hostname(), proxyAddr.Port())

	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return proxyResult{AliveProxy: "", Error: err}
	}

	_ = conn.Close()

	return proxyResult{AliveProxy: proxyAddr.String(), Error: nil}
}

// GetProxyURL returns a Proxy URL after validating if given proxy url is valid
func GetProxyURL(proxyAddr string) (url.URL, error) {
	if url, err := url.Parse(proxyAddr); err == nil && isSupportedProtocol(url.Scheme) {
		return *url, nil
	}
	return url.URL{}, errkit.New("invalid proxy format (It should be http[s]/socks5://[username:password@]host:port)")
}

// isSupportedProtocol checks given protocols are supported
func isSupportedProtocol(value string) bool {
	return value == HTTP || value == HTTPS || value == SOCKS5
}
