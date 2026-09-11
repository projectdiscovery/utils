package llm

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

// chatStub serves a canned or counted /v1/chat/completions response.
func chatStub(t *testing.T, handler http.HandlerFunc) Config {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	return Config{BaseURL: server.URL + "/v1", Model: "test-model"}
}

func okBody(content string) string {
	return fmt.Sprintf(`{"choices":[{"index":0,"finish_reason":"stop","message":{"role":"assistant","content":%q}}]}`, content)
}

func TestResolveBaseURLPresetsAndDefault(t *testing.T) {
	cases := map[string]string{
		"":                  "https://api.openai.com/v1",
		"openai":            "https://api.openai.com/v1",
		"openai-compatible": "https://api.openai.com/v1",
		"ollama":            "http://localhost:11434/v1",
	}
	for provider, want := range cases {
		got, err := Config{Provider: provider}.resolveBaseURL()
		require.NoError(t, err, provider)
		require.Equal(t, want, got, provider)
	}
}

func TestResolveBaseURLExplicitWins(t *testing.T) {
	got, err := Config{Provider: "ollama", BaseURL: "http://host:9/v1/"}.resolveBaseURL()
	require.NoError(t, err)
	require.Equal(t, "http://host:9/v1", got, "explicit base url overrides preset and trailing slash is trimmed")
}

func TestResolveBaseURLUnknownProvider(t *testing.T) {
	_, err := Config{Provider: "nope"}.resolveBaseURL()
	require.ErrorContains(t, err, "unknown llm provider")
}

func TestNewRequiresModel(t *testing.T) {
	_, err := New(Config{Provider: "ollama"})
	require.ErrorContains(t, err, "no llm model configured")
}

func TestCompleteReturnsContent(t *testing.T) {
	cfg := chatStub(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(okBody("hello")))
	})
	client, err := New(cfg)
	require.NoError(t, err)

	got, err := client.Complete(context.Background(), Request{Prompt: "hi"})
	require.NoError(t, err)
	require.Equal(t, "hello", got)
}

func TestCompleteRejectsEmptyPrompt(t *testing.T) {
	client, err := New(Config{Model: "m", BaseURL: "http://x/v1"})
	require.NoError(t, err)
	_, err = client.Complete(context.Background(), Request{Prompt: "  "})
	require.ErrorContains(t, err, "empty prompt")
}

func TestTruncationIsReportedNotEmpty(t *testing.T) {
	cfg := chatStub(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"choices":[{"finish_reason":"length","message":{"content":""}}],"usage":{"completion_tokens":4096}}`))
	})
	client, _ := New(cfg)
	_, err := client.Complete(context.Background(), Request{Prompt: "hi"})
	require.ErrorContains(t, err, "truncated after 4096 tokens")
}

func TestCacheServesRepeatWithoutCallingProvider(t *testing.T) {
	var calls int32
	cfg := chatStub(t, func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&calls, 1)
		_, _ = w.Write([]byte(okBody("cached")))
	})
	cfg.Cache = true
	client, _ := New(cfg)

	for range 3 {
		got, err := client.Complete(context.Background(), Request{Prompt: "same"})
		require.NoError(t, err)
		require.Equal(t, "cached", got)
	}
	require.Equal(t, int32(1), atomic.LoadInt32(&calls), "repeat prompts must hit the cache")
}

func TestBudgetStopsCallsButCacheStillServes(t *testing.T) {
	var calls int32
	cfg := chatStub(t, func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&calls, 1)
		_, _ = w.Write([]byte(okBody("x")))
	})
	cfg.Cache = true
	cfg.MaxCalls = 2
	client, _ := New(cfg)

	// two distinct prompts consume the budget
	_, err := client.Complete(context.Background(), Request{Prompt: "a"})
	require.NoError(t, err)
	_, err = client.Complete(context.Background(), Request{Prompt: "b"})
	require.NoError(t, err)

	// a third distinct prompt is over budget
	_, err = client.Complete(context.Background(), Request{Prompt: "c"})
	require.ErrorContains(t, err, "budget exhausted")

	// but a cached prompt still serves, since cache hits bypass the budget
	got, err := client.Complete(context.Background(), Request{Prompt: "a"})
	require.NoError(t, err)
	require.Equal(t, "x", got)
	require.Equal(t, int32(2), atomic.LoadInt32(&calls))
}

func TestConcurrencyLimitIsRespected(t *testing.T) {
	var inFlight, peak int32
	cfg := chatStub(t, func(w http.ResponseWriter, _ *http.Request) {
		n := atomic.AddInt32(&inFlight, 1)
		for {
			old := atomic.LoadInt32(&peak)
			if n <= old || atomic.CompareAndSwapInt32(&peak, old, n) {
				break
			}
		}
		_, _ = w.Write([]byte(okBody("x")))
		atomic.AddInt32(&inFlight, -1)
	})
	cfg.MaxConcurrency = 2
	client, _ := New(cfg)

	var wg sync.WaitGroup
	for i := range 20 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, _ = client.Complete(context.Background(), Request{Prompt: fmt.Sprintf("p%d", i)})
		}(i)
	}
	wg.Wait()
	require.LessOrEqual(t, atomic.LoadInt32(&peak), int32(2), "must not exceed MaxConcurrency in flight")
}

func TestConfigAPIKeyReachesAuthorizationHeader(t *testing.T) {
	var gotAuth string
	cfg := chatStub(t, func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(okBody("ok")))
	})
	cfg.APIKey = "explicit-key" // must win over the LLM_API_KEY env

	client, err := New(cfg)
	require.NoError(t, err)
	_, err = client.Complete(context.Background(), Request{Prompt: "hi"})
	require.NoError(t, err)
	require.Equal(t, "Bearer explicit-key", gotAuth)
}
