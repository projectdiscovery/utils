package httputil

import (
	"net/http"
	"net/url"
	"sync"
)

// Option represents a configuration option for the cookie jar
type Option func(*CookieJar)

// WithCookieJar sets an existing cookie jar to wrap
func WithCookieJar(jar http.CookieJar) Option {
	return func(cj *CookieJar) {
		cj.jar = jar
	}
}

// CookieJar is a thread-safe wrapper around http.CookieJar
type CookieJar struct {
	jar http.CookieJar
	mu  sync.RWMutex
}

// New creates a new thread-safe cookie jar with the given options
// If no jar is provided, creates a simple in-memory cookie jar
func New(opts ...Option) *CookieJar {
	cj := &CookieJar{}

	// Apply options
	for _, opt := range opts {
		opt(cj)
	}

	// If no jar was provided, create a simple in-memory one
	if cj.jar == nil {
		cj.jar = &memoryCookieJar{
			cookies: make(map[string][]*http.Cookie),
		}
	}

	return cj
}

// SetCookies implements http.CookieJar.SetCookies
func (cj *CookieJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	if cj.jar == nil {
		return
	}

	cj.mu.Lock()
	defer cj.mu.Unlock()
	cj.jar.SetCookies(u, cookies)
}

// Cookies implements http.CookieJar.Cookies
func (cj *CookieJar) Cookies(u *url.URL) []*http.Cookie {
	if cj.jar == nil {
		return nil
	}

	cj.mu.RLock()
	defer cj.mu.RUnlock()
	return cj.jar.Cookies(u)
}
