package llm

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/projectdiscovery/utils/errkit"
)

// Format controls the shape of a completion.
type Format struct {
	// JSON asks the provider to return a single JSON value. Callers that need a
	// specific shape state it in the prompt and validate the result; this only
	// requests JSON rather than prose.
	JSON bool
}

// Request is a single completion.
type Request struct {
	// System is an optional system instruction.
	System string
	// Prompt is the user input.
	Prompt string
	// Format optionally constrains the output.
	Format Format
}

// backend is one provider wire protocol. Keeping it an interface is what lets a
// native Anthropic backend be added later without touching callers or the
// cache, budget and concurrency handling that wrap it.
type backend interface {
	complete(ctx context.Context, model string, req Request) (string, error)
}

// Client is a provider-agnostic LLM client with caching, a call budget, and a
// concurrency limit. It is safe for concurrent use.
type Client struct {
	backend backend
	model   string

	cache   *cache
	budget  int32
	calls   int32
	tickets chan struct{}
}

// New builds a client from config. The API key is read from LLM_API_KEY; local
// providers need none.
func New(config Config) (*Client, error) {
	if strings.TrimSpace(config.Model) == "" {
		return nil, errkit.New("no llm model configured")
	}

	baseURL, err := config.resolveBaseURL()
	if err != nil {
		return nil, err
	}

	timeout := config.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}

	concurrency := config.MaxConcurrency
	if concurrency <= 0 {
		concurrency = 1
	}

	apiKey := config.APIKey
	if apiKey == "" {
		apiKey = os.Getenv(APIKeyEnv)
	}

	client := &Client{
		backend: newOpenAIBackend(baseURL, apiKey, timeout),
		model:   config.Model,
		budget:  int32(config.MaxCalls),
		tickets: make(chan struct{}, concurrency),
	}

	if config.Cache {
		client.cache = newCache()
	}

	return client, nil
}

// Complete returns a completion for the request.
//
// A cached response is returned without consuming the budget or a concurrency
// slot, so a warm cache both costs nothing and cannot be rate-limited. The
// budget is checked and consumed only for calls that actually reach the
// provider.
func (client *Client) Complete(ctx context.Context, req Request) (string, error) {
	if strings.TrimSpace(req.Prompt) == "" {
		return "", errkit.New("empty prompt")
	}

	key := client.key(req)

	if client.cache != nil {
		if cached, ok := client.cache.get(key); ok {
			return cached, nil
		}
	}

	if err := client.reserve(); err != nil {
		return "", err
	}

	client.tickets <- struct{}{}
	defer func() { <-client.tickets }()

	response, err := client.backend.complete(ctx, client.model, req)
	if err != nil {
		return "", err
	}

	if client.cache != nil {
		client.cache.set(key, response)
	}

	return response, nil
}

// reserve consumes one unit of the call budget, if a budget is set.
func (client *Client) reserve() error {
	if client.budget <= 0 {
		return nil
	}

	if atomic.AddInt32(&client.calls, 1) > client.budget {
		return errkit.Newf("llm call budget exhausted (%d calls)", client.budget)
	}

	return nil
}

// key is the cache key: every input that can change the answer, hashed so a
// large prompt or response body does not sit in a map key.
func (client *Client) key(req Request) string {
	hasher := sha256.New()
	for _, part := range []string{client.model, req.System, req.Prompt} {
		_, _ = hasher.Write([]byte(part))
		_, _ = hasher.Write([]byte{0})
	}
	if req.Format.JSON {
		_, _ = hasher.Write([]byte("json"))
	}

	return hex.EncodeToString(hasher.Sum(nil))
}

// cache is a trivial concurrency-safe response cache. Entries live for the life
// of the client, which matches a single scan.
type cache struct {
	mu      sync.RWMutex
	entries map[string]string
}

func newCache() *cache {
	return &cache{entries: make(map[string]string)}
}

func (c *cache) get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, ok := c.entries[key]

	return value, ok
}

func (c *cache) set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = value
}
