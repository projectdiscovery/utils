package httputil

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"

	mapsutil "github.com/projectdiscovery/utils/maps"
	"github.com/projectdiscovery/utils/sync/sizedpool"
)

var (
	// reasonably high default allowed allocs
	DefaultBytesBufferAlloc = int64(10000)
)

func ChangePoolSize(x int64) error {
	return bufPool.Vary(context.Background(), x)
}

func GetPoolSize() int64 {
	return bufPool.Size()
}

// use buffer pool for storing response body
// and reuse it for each request
var bufPool *sizedpool.SizedPool[*bytes.Buffer]

// CachedResponse stores cached response data with reference counting
type CachedResponse struct {
	Body         []byte // Copy of body bytes
	FullResponse []byte // Copy of full response bytes
	RefCount     int32  // Atomic reference counter
}

// ResponseCache provides thread-safe response caching with reference counting
type ResponseCache struct {
	cache *mapsutil.SyncLockMap[string, *CachedResponse]
}

var globalResponseCache = &ResponseCache{
	cache: mapsutil.NewSyncLockMap[string, *CachedResponse](),
}

// hashBytes computes SHA256 hash of data
func hashBytes(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// GetOrStore retrieves cached response or stores a new one
// Returns: (cached response, wasCached)
func (rc *ResponseCache) GetOrStore(bodyHash, fullHash string, bodyData, fullData []byte) (*CachedResponse, bool) {
	// Check if full response is cached (most common case)
	if cached, ok := rc.cache.Get(fullHash); ok {
		atomic.AddInt32(&cached.RefCount, 1)
		return cached, true
	}

	// Store new response
	// Make copies to avoid retaining reference to original buffer
	bodyCopy := make([]byte, len(bodyData))
	fullCopy := make([]byte, len(fullData))
	copy(bodyCopy, bodyData)
	copy(fullCopy, fullData)

	cr := &CachedResponse{
		Body:         bodyCopy,
		FullResponse: fullCopy,
		RefCount:     1,
	}

	// Store by full hash (primary) and body hash (secondary) for lookup
	// Both point to the same CachedResponse object
	_ = rc.cache.Set(fullHash, cr)
	if bodyHash != fullHash {
		_ = rc.cache.Set(bodyHash, cr)
	}
	return cr, false
}

// Release releases a reference to a cached response
// Note: We track bodyHash for cleanup, but primarily use fullHash for lookups
func (rc *ResponseCache) Release(fullHash, bodyHash string) {
	if cached, ok := rc.cache.Get(fullHash); ok {
		if atomic.AddInt32(&cached.RefCount, -1) <= 0 {
			// Clean up both hash entries to avoid memory leaks
			rc.cache.Delete(fullHash)
			if bodyHash != fullHash {
				rc.cache.Delete(bodyHash)
			}
		}
	}
}

// GetCachedBody retrieves cached body by hash (without incrementing ref count)
// Used for runtime resolution of hashes in DSL evaluation
func GetCachedBody(hash string) ([]byte, bool) {
	if cached, ok := globalResponseCache.cache.Get(hash); ok {
		return cached.Body, true
	}
	return nil, false
}

// GetCachedFullResponse retrieves cached full response by hash (without incrementing ref count)
// Used for runtime resolution of hashes in DSL evaluation
func GetCachedFullResponse(hash string) ([]byte, bool) {
	if cached, ok := globalResponseCache.cache.Get(hash); ok {
		return cached.FullResponse, true
	}
	return nil, false
}

func init() {
	var p = &sync.Pool{
		New: func() any {
			// The Pool's New function should generally only return pointer
			// types, since a pointer can be put into the return interface
			// value without an allocation:
			return new(bytes.Buffer)
		},
	}
	var err error
	bufPool, err = sizedpool.New[*bytes.Buffer](
		sizedpool.WithPool[*bytes.Buffer](p),
		sizedpool.WithSize[*bytes.Buffer](int64(DefaultBytesBufferAlloc)),
	)
	if err != nil {
		panic(err)
	}
}

// getBuffer returns a buffer from the pool
func getBuffer() *bytes.Buffer {
	buff, _ := bufPool.Get(context.Background())
	return buff
}

// putBuffer returns a buffer to the pool
func putBuffer(buf *bytes.Buffer) {
	buf.Reset()
	bufPool.Put(buf)
}

// Performance Notes:
// do not use http.Response once we create ResponseChain from it
// as this reuses buffers and saves allocations and also drains response
// body automatically.
// In required cases it can be used but should never be used for anything
// related to response body.
// Bytes.Buffer returned by getters should not be used and are only meant for convinience
// purposes like .String() or .Bytes() calls.
// Remember to call Close() on ResponseChain once you are done with it.

// ResponseChain is a response chain for a http request
// on every call to previous it returns the previous response
// if it was redirected.
type ResponseChain struct {
	headers      *bytes.Buffer
	body         *bytes.Buffer
	fullResponse *bytes.Buffer
	cachedResp   *CachedResponse // nil if not using cache
	bodyHash     string          // SHA256 hash of body
	fullHash     string          // SHA256 hash of full response
	resp         *http.Response
	reloaded     bool // if response was reloaded to its previous redirect
}

// NewResponseChain creates a new response chain for a http request
// with a maximum body size. (if -1 stick to default 4MB)
func NewResponseChain(resp *http.Response, maxBody int64) *ResponseChain {
	if maxBody > 0 && resp.Body != nil {
		resp.Body = http.MaxBytesReader(nil, resp.Body, maxBody)
	}
	return &ResponseChain{
		headers:      getBuffer(),
		body:         getBuffer(),
		fullResponse: getBuffer(),
		resp:         resp,
	}
}

// Response returns the current response in the chain
func (r *ResponseChain) Headers() *bytes.Buffer {
	// Headers are part of fullResponse, but if using cache we need to extract them
	// For now, if cached, return nil since headers are typically not accessed separately
	// and are included in FullResponse(). This maintains backward compatibility.
	if r.cachedResp != nil {
		// Headers are in fullResponse, but we don't have easy way to extract just headers
		// In practice, code should use FullResponse() instead of Headers() when cache is used
		// Return empty buffer to avoid nil pointer issues
		return bytes.NewBuffer(nil)
	}
	return r.headers
}

// Body returns the current response body in the chain
func (r *ResponseChain) Body() *bytes.Buffer {
	if r.cachedResp != nil {
		// Return a buffer wrapper that reads from cache
		return bytes.NewBuffer(r.cachedResp.Body)
	}
	return r.body
}

// FullResponse returns the current response in the chain
func (r *ResponseChain) FullResponse() *bytes.Buffer {
	if r.cachedResp != nil {
		// Return a buffer wrapper that reads from cache
		return bytes.NewBuffer(r.cachedResp.FullResponse)
	}
	return r.fullResponse
}

// BodyHash returns the hash of the response body
func (r *ResponseChain) BodyHash() string {
	return r.bodyHash
}

// FullHash returns the hash of the full response
func (r *ResponseChain) FullHash() string {
	return r.fullHash
}

// IsCached returns true if this response is using cached data
func (r *ResponseChain) IsCached() bool {
	return r.cachedResp != nil
}

// previous updates response pointer to previous response
// if it was redirected and returns true else false
func (r *ResponseChain) Previous() bool {
	if r.resp != nil && r.resp.Request != nil && r.resp.Request.Response != nil {
		r.resp = r.resp.Request.Response
		r.reloaded = true
		return true
	}
	return false
}

// Fill buffers
func (r *ResponseChain) Fill() error {
	r.reset()
	if r.resp == nil {
		return fmt.Errorf("response is nil")
	}

	// load headers
	err := DumpResponseIntoBuffer(r.resp, false, r.headers)
	if err != nil {
		return fmt.Errorf("error dumping response headers: %s", err)
	}

	if r.resp.StatusCode != http.StatusSwitchingProtocols && !r.reloaded {
		// Note about reloaded:
		// this is a known behaviour existing from earlier version
		// when redirect is followed and operators are executed on all redirect chain
		// body of those requests is not available since its already been redirected
		// This is not a issue since redirect happens with empty body according to RFC
		// but this may be required sometimes
		// Solution: Manual redirect using dynamic matchers or hijack redirected responses
		// at transport level at replace with bytes buffer and then use it

		// load body
		err = readNNormalizeRespBody(r, r.body)
		if err != nil {
			return fmt.Errorf("error reading response body: %s", err)
		}

		// response body should not be used anymore
		// drain and close
		DrainResponseBody(r.resp)
	}

	// join headers and body
	r.fullResponse.Write(r.headers.Bytes())
	r.fullResponse.Write(r.body.Bytes())

	// Compute hashes after normalization and full response construction
	bodyBytes := r.body.Bytes()
	fullBytes := r.fullResponse.Bytes()
	r.bodyHash = hashBytes(bodyBytes)
	r.fullHash = hashBytes(fullBytes)

	// Check cache and use cached version if available
	cached, wasCached := globalResponseCache.GetOrStore(r.bodyHash, r.fullHash, bodyBytes, fullBytes)
	if wasCached {
		// Release buffers immediately since we're using cached version
		putBuffer(r.headers)
		putBuffer(r.body)
		putBuffer(r.fullResponse)
		r.headers = nil
		r.body = nil
		r.fullResponse = nil
		r.cachedResp = cached
	} else {
		// Keep buffers, store reference for future cleanup
		r.cachedResp = cached
	}

	return nil
}

// Close the response chain and releases the buffers.
func (r *ResponseChain) Close() {
	// Release cache reference if using cached response
	if r.cachedResp != nil && r.fullHash != "" {
		globalResponseCache.Release(r.fullHash, r.bodyHash)
		r.cachedResp = nil
	}

	// Release buffers (will be no-op if already released due to cache hit)
	if r.headers != nil {
		putBuffer(r.headers)
		r.headers = nil
	}
	if r.body != nil {
		putBuffer(r.body)
		r.body = nil
	}
	if r.fullResponse != nil {
		putBuffer(r.fullResponse)
		r.fullResponse = nil
	}

	// Clear hash references
	r.bodyHash = ""
	r.fullHash = ""
}

// Has returns true if the response chain has a response
func (r *ResponseChain) Has() bool {
	return r.resp != nil
}

// Request is request of current response
func (r *ResponseChain) Request() *http.Request {
	if r.resp == nil {
		return nil
	}
	return r.resp.Request
}

// Response is response of current response
func (r *ResponseChain) Response() *http.Response {
	return r.resp
}

// reset without releasing the buffers
// useful for redirect chain
func (r *ResponseChain) reset() {
	// Clear cached response reference if set (for redirect chains)
	if r.cachedResp != nil && r.fullHash != "" {
		globalResponseCache.Release(r.fullHash, r.bodyHash)
		r.cachedResp = nil
		r.bodyHash = ""
		r.fullHash = ""
	}

	// Reset buffers only if they exist (not released due to cache hit)
	if r.headers != nil {
		r.headers.Reset()
	}
	if r.body != nil {
		r.body.Reset()
	}
	if r.fullResponse != nil {
		r.fullResponse.Reset()
	}
}
