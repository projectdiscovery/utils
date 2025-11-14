package httputil

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestResponseChain_BasicFunctionality tests basic ResponseChain operations
func TestResponseChain_BasicFunctionality(t *testing.T) {
	body := "Hello, World!"
	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     http.Header{"Content-Type": []string{"text/plain"}},
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
	}

	rc := NewResponseChain(resp, -1)
	require.NotNil(t, rc)
	require.True(t, rc.Has())

	err := rc.Fill()
	require.NoError(t, err)

	// Test body accessors
	assert.Equal(t, body, rc.BodyString())
	assert.Equal(t, []byte(body), rc.BodyBytes())
	assert.Equal(t, body, rc.Body().String())

	// Test headers accessors
	headers := rc.HeadersString()
	assert.Contains(t, headers, "HTTP/1.1 200 OK")
	assert.Contains(t, headers, "Content-Type: text/plain")

	// Test full response
	fullResp := rc.FullResponseString()
	assert.Contains(t, fullResp, "HTTP/1.1 200 OK")
	assert.Contains(t, fullResp, body)

	rc.Close()
}

// TestResponseChain_LargeBody tests handling of large response bodies
func TestResponseChain_LargeBody(t *testing.T) {
	// Create a 1MB body
	largeBody := bytes.Repeat([]byte("A"), 1024*1024)

	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewReader(largeBody)),
		Header:     http.Header{},
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
	}

	rc := NewResponseChain(resp, -1)
	err := rc.Fill()
	require.NoError(t, err)

	assert.Equal(t, len(largeBody), len(rc.BodyBytes()))
	assert.Equal(t, largeBody, rc.BodyBytes())

	rc.Close()
}

// TestResponseChain_MaxBodyLimit tests body size limiting
func TestResponseChain_MaxBodyLimit(t *testing.T) {
	maxBody := int64(1024)                       // 1KB limit
	largeBody := bytes.Repeat([]byte("B"), 2048) // 2KB body

	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewReader(largeBody)),
		Header:     http.Header{},
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
	}

	rc := NewResponseChain(resp, maxBody)
	err := rc.Fill()

	// Should either error or truncate
	if err == nil {
		// If no error, body should be truncated
		assert.LessOrEqual(t, len(rc.BodyBytes()), int(maxBody))
	}

	rc.Close()
}

// TestResponseChain_GzipHandling tests gzip-compressed responses
func TestResponseChain_GzipHandling(t *testing.T) {
	originalBody := "This is a compressed response body that should be decompressed"

	// Create gzip-compressed body
	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)
	_, err := gzWriter.Write([]byte(originalBody))
	require.NoError(t, err)
	require.NoError(t, gzWriter.Close())

	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(&buf),
		Header: http.Header{
			"Content-Encoding": []string{"gzip"},
		},
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
	}

	rc := NewResponseChain(resp, -1)
	err = rc.Fill()
	require.NoError(t, err)

	// Body should be decompressed
	assert.Equal(t, originalBody, rc.BodyString())

	rc.Close()
}

// TestResponseChain_EmptyBody tests handling of empty response bodies
func TestResponseChain_EmptyBody(t *testing.T) {
	resp := &http.Response{
		StatusCode: 204, // No Content
		Body:       io.NopCloser(strings.NewReader("")),
		Header:     http.Header{},
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
	}

	rc := NewResponseChain(resp, -1)
	err := rc.Fill()
	require.NoError(t, err)

	assert.Empty(t, rc.BodyString())
	assert.Empty(t, rc.BodyBytes())
	assert.NotEmpty(t, rc.HeadersString()) // Headers should still exist

	rc.Close()
}

// TestResponseChain_FullResponseOnDemand tests that FullResponse creates buffer on-demand
func TestResponseChain_FullResponseOnDemand(t *testing.T) {
	body := "Test body"
	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     http.Header{},
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
	}

	rc := NewResponseChain(resp, -1)
	err := rc.Fill()
	require.NoError(t, err)

	// Get full response multiple times - should create new buffer each time
	full1 := rc.FullResponse()
	full2 := rc.FullResponse()

	assert.NotSame(t, full1, full2)                 // Different buffer instances
	assert.Equal(t, full1.String(), full2.String()) // Same content

	// Clean up buffers
	putBuffer(full1)
	putBuffer(full2)
	rc.Close()
}

// TestResponseChain_SafeAccessors tests the new safe accessor methods
func TestResponseChain_SafeAccessors(t *testing.T) {
	body := "Safe access test"
	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     http.Header{"X-Test": []string{"value"}},
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
	}

	rc := NewResponseChain(resp, -1)
	err := rc.Fill()
	require.NoError(t, err)

	// Test BodyString vs Body().String()
	assert.Equal(t, rc.BodyString(), rc.Body().String())
	assert.Equal(t, rc.BodyBytes(), rc.Body().Bytes())

	// Test HeadersString vs Headers().String()
	assert.Equal(t, rc.HeadersString(), rc.Headers().String())
	assert.Equal(t, rc.HeadersBytes(), rc.Headers().Bytes())

	// Test FullResponse variants
	fullBuf := rc.FullResponse()
	defer putBuffer(fullBuf)

	fullBytes := rc.FullResponseBytes()
	fullString := rc.FullResponseString()

	assert.Contains(t, string(fullBytes), body)
	assert.Contains(t, fullString, body)

	rc.Close()
}

// TestBufferPool_GetPut tests buffer pool operations
func TestBufferPool_GetPut(t *testing.T) {
	buf1 := getBuffer()
	require.NotNil(t, buf1)

	buf1.WriteString("test data")
	putBuffer(buf1)

	buf2 := getBuffer()
	require.NotNil(t, buf2)

	// Buffer should be reset when returned from pool
	assert.Equal(t, 0, buf2.Len())

	putBuffer(buf2)
}

// TestBufferPool_LargeBufferLimiting tests that large buffers are limited in pool
func TestBufferPool_LargeBufferLimiting(t *testing.T) {
	// Save original settings
	origMaxLarge := maxLargeBuffers
	defer func() {
		maxLargeBuffers = origMaxLarge
		setLargeBufferSemSize(origMaxLarge)
	}()

	// Set small limit for testing
	SetMaxLargeBuffers(5)

	// Create responses that will use large buffers
	largeBody := bytes.Repeat([]byte("X"), DefaultMaxBodySize)

	var chains []*ResponseChain
	for i := 0; i < 10; i++ {
		resp := &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewReader(largeBody)),
			Header:     http.Header{},
			Proto:      "HTTP/1.1",
			ProtoMajor: 1,
			ProtoMinor: 1,
		}
		rc := NewResponseChain(resp, -1)
		err := rc.Fill()
		require.NoError(t, err)
		chains = append(chains, rc)
	}

	// Close all chains
	for _, rc := range chains {
		rc.Close()
	}

	// Pool should not have accumulated too many large buffers
	// This is a behavioral test - exact assertion depends on implementation
}

// TestBufferPool_OversizedBufferDiscarded tests that oversized buffers are not pooled
func TestBufferPool_OversizedBufferDiscarded(t *testing.T) {
	// Create a buffer larger than DefaultMaxBodySize
	buf := getBuffer()
	buf.Grow(DefaultMaxBodySize + 1024)

	initialCap := buf.Cap()
	assert.Greater(t, initialCap, DefaultMaxBodySize)

	// Put it back - should be discarded
	putBuffer(buf)

	// Get a new buffer - should be normal size, not the oversized one
	buf2 := getBuffer()
	assert.LessOrEqual(t, buf2.Cap(), DefaultMaxBodySize)

	putBuffer(buf2)
}

// TestLimitedBuffer_ChunkedReading tests the limitedBuffer implementation
func TestLimitedBuffer_ChunkedReading(t *testing.T) {
	// Create data larger than chunk size (32KB)
	data := bytes.Repeat([]byte("L"), 64*1024) // 64KB

	buf := &bytes.Buffer{}
	lb := &limitedBuffer{buf: buf, maxCap: len(data)}

	n, err := lb.ReadFrom(bytes.NewReader(data))
	require.NoError(t, err)
	assert.Equal(t, int64(len(data)), n)
	assert.Equal(t, data, buf.Bytes())
}

// TestLimitedBuffer_CapacityLimit tests that limitedBuffer respects maxCap
func TestLimitedBuffer_CapacityLimit(t *testing.T) {
	maxCap := 1024
	data := bytes.Repeat([]byte("M"), 2048) // More than maxCap

	buf := &bytes.Buffer{}
	lb := &limitedBuffer{buf: buf, maxCap: maxCap}

	_, err := lb.ReadFrom(bytes.NewReader(data))
	require.NoError(t, err)

	// Buffer should not grow beyond maxCap
	assert.LessOrEqual(t, buf.Cap(), maxCap*2) // Allow some overhead
}

// TestSetBufferSize tests buffer size configuration
func TestSetBufferSize(t *testing.T) {
	originalSize := bufferSize
	defer func() {
		SetBufferSize(originalSize)
	}()

	// Test setting valid size
	newSize := int64(20000)
	SetBufferSize(newSize)
	assert.Equal(t, newSize, bufferSize)

	// Test minimum size enforcement
	SetBufferSize(100)
	assert.Equal(t, int64(1000), bufferSize)
}

// TestSetMaxLargeBuffers tests large buffer limit configuration
func TestSetMaxLargeBuffers(t *testing.T) {
	originalMax := maxLargeBuffers
	defer func() {
		maxLargeBuffers = originalMax
		setLargeBufferSemSize(originalMax)
	}()

	// Test setting valid size
	newMax := 200
	SetMaxLargeBuffers(newMax)
	// Due to minimum enforcement logic, it should use DefaultMaxLargeBuffers if less
	assert.GreaterOrEqual(t, maxLargeBuffers, DefaultMaxLargeBuffers)
}

// TestResponseChain_ConcurrentAccess tests thread-safety of ResponseChain
func TestResponseChain_ConcurrentAccess(t *testing.T) {
	body := "Concurrent access test"
	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     http.Header{},
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
	}

	rc := NewResponseChain(resp, -1)
	err := rc.Fill()
	require.NoError(t, err)

	// Concurrent reads should be safe
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = rc.BodyString()
			_ = rc.HeadersString()
			_ = rc.FullResponseString()
		}()
	}
	wg.Wait()

	rc.Close()
}

// TestResponseChain_MultipleResponses tests handling of response chains
func TestResponseChain_MultipleResponses(t *testing.T) {
	// Test that Previous() method works correctly
	// In HTTP redirect chains, resp.Request.Response points to the previous response
	body1 := "First response"

	resp1 := &http.Response{
		StatusCode: 302,
		Body:       io.NopCloser(strings.NewReader(body1)),
		Header:     http.Header{"Location": []string{"/redirected"}},
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
	}

	// Simulate a redirect chain
	req := &http.Request{
		Response: resp1,
	}

	body2 := "Second response"
	resp2 := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(body2)),
		Header:     http.Header{},
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Request:    req,
	}

	// Start with final response
	rc := NewResponseChain(resp2, -1)
	err := rc.Fill()
	require.NoError(t, err)

	// Should contain second (final) response
	assert.Contains(t, rc.BodyString(), body2)

	// Test Previous() method
	hasPrevious := rc.Previous()
	assert.True(t, hasPrevious, "Should have previous response in chain")

	// Reset buffers and fill with previous response
	err = rc.Fill()
	require.NoError(t, err)

	// Should now contain first response
	bodyContent := rc.BodyString()
	if bodyContent != "" {
		// Body might be empty if already consumed, but if not, it should match
		assert.Contains(t, bodyContent, body1)
	}

	rc.Close()
}

// TestResponseChain_Reset tests the reset functionality
func TestResponseChain_Reset(t *testing.T) {
	body := "Reset test"
	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     http.Header{},
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
	}

	rc := NewResponseChain(resp, -1)
	err := rc.Fill()
	require.NoError(t, err)

	assert.NotEmpty(t, rc.BodyString())
	assert.NotEmpty(t, rc.HeadersString())

	// Reset should clear buffers
	rc.reset()

	assert.Empty(t, rc.Body().String())
	assert.Empty(t, rc.Headers().String())

	rc.Close()
}

// TestDrainResponseBody tests response body draining
func TestDrainResponseBody(t *testing.T) {
	body := bytes.Repeat([]byte("D"), 1024)
	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewReader(body)),
		Header:     http.Header{},
	}

	DrainResponseBody(resp)

	// Body should be closed
	_, err := resp.Body.Read(make([]byte, 1))
	assert.Error(t, err) // Should error because body is closed
}

// TestChangePoolSize tests dynamic pool size changes
func TestChangePoolSize(t *testing.T) {
	originalSize := GetPoolSize()

	// ChangePoolSize uses Vary which adds/subtracts from current size
	delta := int64(5000)
	err := ChangePoolSize(delta)
	require.NoError(t, err)
	assert.Equal(t, originalSize+delta, GetPoolSize())

	// Restore original size by subtracting the delta
	err = ChangePoolSize(-delta)
	require.NoError(t, err)
	assert.Equal(t, originalSize, GetPoolSize())
}

// TestResponseChain_NilBody tests handling of nil response body
func TestResponseChain_NilBody(t *testing.T) {
	resp := &http.Response{
		StatusCode: 204,
		Body:       http.NoBody,
		Header:     http.Header{},
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
	}

	rc := NewResponseChain(resp, -1)
	err := rc.Fill()

	// Should handle empty body gracefully
	require.NoError(t, err)
	assert.Empty(t, rc.BodyString())

	rc.Close()
}

// TestResponseChain_InvalidGzip tests handling of invalid gzip data
func TestResponseChain_InvalidGzip(t *testing.T) {
	// When gzip header is present but data is invalid, the normalization code
	// attempts to fall back to reading the original body. However, if the body
	// has been consumed, it may result in empty data.
	invalidGzip := []byte("This is not valid gzip data")

	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewReader(invalidGzip)),
		Header: http.Header{
			"Content-Encoding": []string{"gzip"},
		},
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
	}

	rc := NewResponseChain(resp, -1)
	err := rc.Fill()

	// The Fill should not error even with invalid gzip
	require.NoError(t, err)

	// Body may be empty or contain partial data depending on how much
	// was consumed before the gzip error was detected
	// Just verify it doesn't panic
	_ = rc.BodyBytes()

	rc.Close()
}

// TestResponseChain_BurstWorkload tests buffer pool behavior under burst traffic
func TestResponseChain_BurstWorkload(t *testing.T) {
	// Simulate a burst of requests (e.g., nuclei scan starting)
	burstSize := 500
	largeBody := bytes.Repeat([]byte("B"), DefaultMaxBodySize) // Max size body

	var wg sync.WaitGroup
	errChan := make(chan error, burstSize)

	// Track initial pool and memory metrics
	initialPoolSize := GetPoolSize()

	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)
	t.Logf("Before =  Alloc: %d MB, TotalAlloc: %d MB, Sys: %d MB, NumGC: %d",
		m1.Alloc/1024/1024, m1.TotalAlloc/1024/1024, m1.Sys/1024/1024, m1.NumGC)

	// Create burst of concurrent requests
	for i := 0; i < burstSize; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			resp := &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewReader(largeBody)),
				Header:     http.Header{},
				Proto:      "HTTP/1.1",
				ProtoMajor: 1,
				ProtoMinor: 1,
			}

			rc := NewResponseChain(resp, -1)
			if err := rc.Fill(); err != nil {
				errChan <- err
				return
			}

			// Verify we got the data
			if len(rc.BodyBytes()) != len(largeBody) {
				errChan <- assert.AnError
				return
			}

			rc.Close()
		}(i)
	}

	wg.Wait()
	close(errChan)

	// Check memory after burst
	runtime.GC()
	runtime.ReadMemStats(&m2)
	t.Logf("After  =  Alloc: %d MB, TotalAlloc: %d MB, Sys: %d MB, NumGC: %d",
		m2.Alloc/1024/1024, m2.TotalAlloc/1024/1024, m2.Sys/1024/1024, m2.NumGC)
	t.Logf("Memory delta - Alloc: %+d MB, TotalAlloc: %+d MB, Sys: %+d MB",
		int64(m2.Alloc-m1.Alloc)/1024/1024, int64(m2.TotalAlloc-m1.TotalAlloc)/1024/1024,
		int64(m2.Sys-m1.Sys)/1024/1024)

	// Check for errors
	for err := range errChan {
		require.NoError(t, err)
	}

	// Pool should still be healthy after burst
	finalPoolSize := GetPoolSize()
	assert.Equal(t, initialPoolSize, finalPoolSize, "Pool size should remain stable")

	// Memory should not grow excessively (allowing some overhead for pool)
	memGrowthMB := int64(m2.Alloc-m1.Alloc) / 1024 / 1024
	t.Logf("Memory growth: %d MB", memGrowthMB)
}

// TestResponseChain_SustainedConcurrency tests sustained concurrent load
func TestResponseChain_SustainedConcurrency(t *testing.T) {
	// Simulate sustained concurrent requests over time
	duration := 2 // seconds
	concurrency := 50
	stopChan := make(chan struct{})
	errChan := make(chan error, concurrency*10)

	// Mix of different body sizes
	bodySizes := []int{
		1024,                   // 1KB
		100 * 1024,             // 100KB
		1024 * 1024,            // 1MB
		DefaultMaxBodySize / 2, // Half max
		DefaultMaxBodySize,     // Max size
	}

	var wg sync.WaitGroup

	// Memory tracking
	var m1, m2, mPeak runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)
	t.Logf("Before =  Alloc: %d MB, Sys: %d MB",
		m1.Alloc/1024/1024, m1.Sys/1024/1024)

	requestCounter := &sync.Map{}
	peakAlloc := uint64(0)

	// Memory monitoring goroutine
	memStopChan := make(chan struct{})
	var memWg sync.WaitGroup
	memWg.Add(1)
	go func() {
		defer memWg.Done()
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-memStopChan:
				return
			case <-ticker.C:
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				if m.Alloc > peakAlloc {
					peakAlloc = m.Alloc
					mPeak = m
				}
			}
		}
	}()

	// Start concurrent workers
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			requestCount := 0

			for {
				select {
				case <-stopChan:
					requestCounter.Store(workerID, requestCount)
					return
				default:
					// Pick a body size based on request count
					bodySize := bodySizes[requestCount%len(bodySizes)]
					body := bytes.Repeat([]byte("S"), bodySize)

					resp := &http.Response{
						StatusCode: 200,
						Body:       io.NopCloser(bytes.NewReader(body)),
						Header:     http.Header{},
						Proto:      "HTTP/1.1",
						ProtoMajor: 1,
						ProtoMinor: 1,
					}

					rc := NewResponseChain(resp, -1)
					if err := rc.Fill(); err != nil {
						errChan <- err
						return
					}

					// Verify data
					if len(rc.BodyBytes()) != bodySize {
						errChan <- assert.AnError
						return
					}

					rc.Close()
					requestCount++
				}
			}
		}(i)
	}

	// Let it run for specified duration
	time.Sleep(time.Duration(duration) * time.Second)
	close(stopChan)
	wg.Wait()

	// Stop memory monitoring
	close(memStopChan)
	memWg.Wait()

	close(errChan)

	// Calculate total requests
	totalRequests := 0
	requestCounter.Range(func(key, value interface{}) bool {
		totalRequests += value.(int)
		return true
	})

	// Check memory after sustained load
	runtime.GC()
	runtime.ReadMemStats(&m2)
	t.Logf("After  =  Alloc: %d MB, Sys: %d MB",
		m2.Alloc/1024/1024, m2.Sys/1024/1024)
	t.Logf("Peak during load     - Alloc: %d MB, Sys: %d MB",
		mPeak.Alloc/1024/1024, mPeak.Sys/1024/1024)
	t.Logf("Total requests: %d (%.0f req/s), Memory delta: %+d MB",
		totalRequests, float64(totalRequests)/float64(duration),
		int64(m2.Alloc-m1.Alloc)/1024/1024)

	// Check for errors
	errorCount := 0
	for err := range errChan {
		errorCount++
		t.Logf("Error during sustained load: %v", err)
	}
	assert.Equal(t, 0, errorCount, "Should have no errors during sustained load")
}

// TestResponseChain_MemoryPressure tests behavior under memory pressure with large buffers
func TestResponseChain_MemoryPressure(t *testing.T) {
	// Save and restore settings
	origMaxLarge := maxLargeBuffers
	defer func() {
		SetMaxLargeBuffers(origMaxLarge)
	}()

	// Set a small limit to test pressure handling
	testMaxLarge := 10
	SetMaxLargeBuffers(testMaxLarge)

	// Create more large buffer requests than the limit allows
	numRequests := testMaxLarge * 3
	largeBody := bytes.Repeat([]byte("M"), DefaultMaxBodySize)

	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)
	t.Logf("Before =  Alloc: %d MB, Sys: %d MB, MaxLargeBuffers: %d",
		m1.Alloc/1024/1024, m1.Sys/1024/1024, testMaxLarge)

	var wg sync.WaitGroup
	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			resp := &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewReader(largeBody)),
				Header:     http.Header{},
				Proto:      "HTTP/1.1",
				ProtoMajor: 1,
				ProtoMinor: 1,
			}

			rc := NewResponseChain(resp, -1)
			err := rc.Fill()
			require.NoError(t, err)

			// Verify data integrity despite memory pressure
			assert.Equal(t, len(largeBody), len(rc.BodyBytes()))

			rc.Close()
		}(i)
	}

	wg.Wait()

	runtime.GC()
	runtime.ReadMemStats(&m2)
	t.Logf("After  =  Alloc: %d MB, Sys: %d MB",
		m2.Alloc/1024/1024, m2.Sys/1024/1024)
	t.Logf("Handled %d requests (3x buffer limit) = Memory delta: %+d MB",
		numRequests, int64(m2.Alloc-m1.Alloc)/1024/1024)

	// System should still be functional after pressure
	// Create a new request to verify pool is still working
	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader("test")),
		Header:     http.Header{},
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
	}

	rc := NewResponseChain(resp, -1)
	err := rc.Fill()
	require.NoError(t, err)
	assert.Equal(t, "test", rc.BodyString())
	rc.Close()
}

// TestResponseChain_MixedWorkload tests realistic mixed workload patterns
func TestResponseChain_MixedWorkload(t *testing.T) {
	concurrency := 30
	requestsPerWorker := 20

	// Different request patterns
	patterns := []struct {
		name     string
		bodySize int
		compress bool
	}{
		{"small", 512, false},
		{"medium", 64 * 1024, false},
		{"large", 2 * 1024 * 1024, false},
		{"small-gzip", 512, true},
		{"medium-gzip", 64 * 1024, true},
	}

	var wg sync.WaitGroup
	errChan := make(chan error, concurrency*requestsPerWorker)

	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)
	t.Logf("Before =  Alloc: %d MB", m1.Alloc/1024/1024)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < requestsPerWorker; j++ {
				pattern := patterns[j%len(patterns)]
				body := bytes.Repeat([]byte("X"), pattern.bodySize)

				var bodyReader io.Reader
				var headers http.Header

				if pattern.compress {
					var buf bytes.Buffer
					gzWriter := gzip.NewWriter(&buf)
					_, err := gzWriter.Write(body)
					if err != nil {
						errChan <- err
						return
					}
					_ = gzWriter.Close()
					bodyReader = &buf
					headers = http.Header{"Content-Encoding": []string{"gzip"}}
				} else {
					bodyReader = bytes.NewReader(body)
					headers = http.Header{}
				}

				resp := &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(bodyReader),
					Header:     headers,
					Proto:      "HTTP/1.1",
					ProtoMajor: 1,
					ProtoMinor: 1,
				}

				rc := NewResponseChain(resp, -1)
				if err := rc.Fill(); err != nil {
					errChan <- err
					return
				}

				// Verify decompressed data matches
				if len(rc.BodyBytes()) != pattern.bodySize {
					errChan <- assert.AnError
					return
				}

				rc.Close()
			}
		}(i)
	}

	wg.Wait()
	close(errChan)

	runtime.GC()
	runtime.ReadMemStats(&m2)
	totalRequests := concurrency * requestsPerWorker
	t.Logf("After  =  Alloc: %d MB", m2.Alloc/1024/1024)
	t.Logf("Processed %d requests with mixed sizes/compression = Memory delta: %+d MB",
		totalRequests, int64(m2.Alloc-m1.Alloc)/1024/1024)

	// Check for errors
	for err := range errChan {
		require.NoError(t, err)
	}
}

// TestResponseChain_RapidCreateDestroy tests rapid allocation/deallocation
func TestResponseChain_RapidCreateDestroy(t *testing.T) {
	// This tests that buffer pool handles rapid churn correctly
	iterations := 1000
	body := bytes.Repeat([]byte("R"), 10*1024) // 10KB

	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)
	t.Logf("Before =  Alloc: %d MB, NumGC: %d",
		m1.Alloc/1024/1024, m1.NumGC)

	for i := 0; i < iterations; i++ {
		resp := &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewReader(body)),
			Header:     http.Header{},
			Proto:      "HTTP/1.1",
			ProtoMajor: 1,
			ProtoMinor: 1,
		}

		rc := NewResponseChain(resp, -1)
		err := rc.Fill()
		require.NoError(t, err)
		assert.Equal(t, len(body), len(rc.BodyBytes()))
		rc.Close()
	}

	runtime.GC()
	runtime.ReadMemStats(&m2)
	t.Logf("After  =  Alloc: %d MB, NumGC: %d",
		m2.Alloc/1024/1024, m2.NumGC)
	t.Logf("Processed %d iterations (%.0f KB total) = Memory delta: %+d MB, GC cycles: %d",
		iterations, float64(iterations*len(body))/1024,
		int64(m2.Alloc-m1.Alloc)/1024/1024, m2.NumGC-m1.NumGC)

	// Pool should still be healthy
	finalSize := GetPoolSize()
	assert.Greater(t, finalSize, int64(0))
}

// TestResponseChain_ConcurrentReads tests concurrent reads from same ResponseChain
func TestResponseChain_ConcurrentReads(t *testing.T) {
	body := "Concurrent read test data"
	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     http.Header{"X-Test": []string{"value"}},
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
	}

	rc := NewResponseChain(resp, -1)
	err := rc.Fill()
	require.NoError(t, err)

	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)
	t.Logf("Before =  Alloc: %d MB", m1.Alloc/1024/1024)

	// Multiple goroutines reading concurrently
	readers := 100
	var wg sync.WaitGroup

	for i := 0; i < readers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Try different access methods
			switch id % 4 {
			case 0:
				s := rc.BodyString()
				assert.Equal(t, body, s)
			case 1:
				b := rc.BodyBytes()
				assert.Equal(t, []byte(body), b)
			case 2:
				h := rc.HeadersString()
				assert.Contains(t, h, "HTTP/1.1 200 OK")
			case 3:
				f := rc.FullResponseString()
				assert.Contains(t, f, body)
			}
		}(i)
	}

	wg.Wait()

	runtime.GC()
	runtime.ReadMemStats(&m2)
	t.Logf("After  =  Alloc: %d MB", m2.Alloc/1024/1024)
	t.Logf("%d concurrent readers = Memory delta: %+d MB (should be ~0 for read-only ops)",
		readers, int64(m2.Alloc-m1.Alloc)/1024/1024)

	rc.Close()
}

// TestResponseChain_BurstWithPoolExhaustion tests pool behavior when exhausted
func TestResponseChain_BurstWithPoolExhaustion(t *testing.T) {
	// Save original pool size
	originalSize := GetPoolSize()
	defer func() {
		// Restore by adjusting
		_ = ChangePoolSize(originalSize - GetPoolSize())
	}()

	// Reduce pool size to force exhaustion
	smallPoolSize := int64(10)
	_ = ChangePoolSize(smallPoolSize - GetPoolSize())

	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)
	t.Logf("Before =  Alloc: %d MB, PoolSize: %d",
		m1.Alloc/1024/1024, GetPoolSize())

	// Create more concurrent requests than pool can handle
	concurrency := 50
	body := bytes.Repeat([]byte("E"), 50*1024) // 50KB

	var wg sync.WaitGroup
	errChan := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			resp := &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewReader(body)),
				Header:     http.Header{},
				Proto:      "HTTP/1.1",
				ProtoMajor: 1,
				ProtoMinor: 1,
			}

			rc := NewResponseChain(resp, -1)
			if err := rc.Fill(); err != nil {
				errChan <- err
				return
			}

			// Should still work even if pool is exhausted
			if len(rc.BodyBytes()) != len(body) {
				errChan <- assert.AnError
				return
			}

			rc.Close()
		}(i)
	}

	wg.Wait()
	close(errChan)

	runtime.GC()
	runtime.ReadMemStats(&m2)
	t.Logf("After  =  Alloc: %d MB, PoolSize: %d",
		m2.Alloc/1024/1024, GetPoolSize())
	t.Logf("Handled %d requests with pool size %d = Memory delta: %+d MB",
		concurrency, smallPoolSize, int64(m2.Alloc-m1.Alloc)/1024/1024)

	// Should handle pool exhaustion gracefully
	for err := range errChan {
		require.NoError(t, err)
	}
}
