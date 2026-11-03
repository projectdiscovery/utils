package llm

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
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

// judger is implemented by backends that classify natively and return a
// probability, instead of asking a chat model to report its own confidence.
type judger interface {
	judge(ctx context.Context, model string, req JudgeRequest) (Verdict, error)
}

// JudgeRequest asks which label fits the inputs.
type JudgeRequest struct {
	// Question is the instruction. It never carries the inputs, which is what
	// keeps text inside them from being read as instructions.
	Question string
	// Labels are the allowed answers. Empty means a yes or no question.
	Labels []string
	// Inputs are the data being judged, such as one or more responses.
	Inputs []string
}

// Verdict is the label a judge chose and its probability.
type Verdict struct {
	Label      string  `json:"label"`
	Confidence float64 `json:"confidence"`
	// Model is the exact model version that answered, when the provider
	// reports one, so a verdict can be traced after an alias moves.
	Model string `json:"model,omitempty"`
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

// New builds a client from config. Hosted presets require Config.APIKey or
// LLM_API_KEY; local presets and a custom BaseURL do not.
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
	if requiresAPIKey(config.Provider, baseURL) && apiKey == "" {
		return nil, errkit.New("llm api key required (set LLM_API_KEY or Config.APIKey)")
	}

	var backend backend = newOpenAIBackend(baseURL, apiKey, timeout)
	if strings.EqualFold(strings.TrimSpace(config.Provider), ProviderJev) || baseURL == presets[ProviderJev] {
		backend = newJevBackend(baseURL, apiKey, timeout)
	}

	client := &Client{
		backend: backend,
		model:   config.Model,
		budget:  int32(config.MaxCalls),
		tickets: make(chan struct{}, concurrency),
	}

	if config.Cache {
		max := defaultMaxCacheEntries
		if config.MaxCalls > 0 && config.MaxCalls < max {
			max = config.MaxCalls
		}
		client.cache = newCache(max)
	}

	return client, nil
}

// Complete returns a completion for the request.
func (client *Client) Complete(ctx context.Context, req Request) (string, error) {
	if strings.TrimSpace(req.Prompt) == "" {
		return "", errkit.New("empty prompt")
	}

	return client.call(ctx, client.key(req), func() (string, error) {
		return client.backend.complete(ctx, client.model, req)
	})
}

// CanJudge reports whether the provider answers Judge.
func (client *Client) CanJudge() bool {
	_, ok := client.backend.(judger)
	return ok
}

// Judge asks the provider which label fits the inputs. It shares the cache,
// budget and concurrency limit with Complete.
func (client *Client) Judge(ctx context.Context, req JudgeRequest) (Verdict, error) {
	judge, ok := client.backend.(judger)
	if !ok {
		return Verdict{}, errkit.New("llm provider does not support judge requests")
	}
	if strings.TrimSpace(req.Question) == "" {
		return Verdict{}, errkit.New("empty question")
	}
	if len(req.Inputs) == 0 {
		return Verdict{}, errkit.New("no inputs to judge")
	}

	parts := append([]string{"judge", req.Question, strings.Join(req.Labels, "\x00")}, req.Inputs...)
	raw, err := client.call(ctx, client.hash(parts...), func() (string, error) {
		verdict, err := judge.judge(ctx, client.model, req)
		if err != nil {
			return "", err
		}
		encoded, err := json.Marshal(verdict)
		return string(encoded), err
	})
	if err != nil {
		return Verdict{}, err
	}

	var verdict Verdict
	if err := json.Unmarshal([]byte(raw), &verdict); err != nil {
		return Verdict{}, errkit.Wrap(err, "could not decode cached verdict")
	}
	return verdict, nil
}

// call runs one provider request under the cache, budget and concurrency limit.
//
// A cached response is returned without consuming the budget or a concurrency
// slot, so a warm cache both costs nothing and cannot be rate-limited. The
// budget is checked and consumed only for calls that actually wait to reach
// the provider; a cancelled wait refunds the reservation.
func (client *Client) call(ctx context.Context, key string, request func() (string, error)) (string, error) {
	if client.cache != nil {
		if cached, ok := client.cache.get(key); ok {
			return cached, nil
		}
	}

	if err := client.reserve(); err != nil {
		return "", err
	}

	select {
	case <-ctx.Done():
		client.release()
		return "", ctx.Err()
	case client.tickets <- struct{}{}:
	}
	defer func() { <-client.tickets }()

	response, err := request()
	if err != nil {
		return "", err
	}

	if client.cache != nil {
		client.cache.set(key, response)
	}

	return response, nil
}

func (client *Client) reserve() error {
	if client.budget <= 0 {
		return nil
	}

	if atomic.AddInt32(&client.calls, 1) > client.budget {
		atomic.AddInt32(&client.calls, -1)
		return errkit.Newf("llm call budget exhausted (%d calls)", client.budget)
	}

	return nil
}

func (client *Client) release() {
	if client.budget > 0 {
		atomic.AddInt32(&client.calls, -1)
	}
}

// key is the cache key: every input that can change the answer, hashed so a
// large prompt or response body does not sit in a map key.
func (client *Client) key(req Request) string {
	format := ""
	if req.Format.JSON {
		format = "json"
	}
	return client.hash(req.System, req.Prompt, format)
}

func (client *Client) hash(parts ...string) string {
	hasher := sha256.New()
	for _, part := range append([]string{client.model}, parts...) {
		_, _ = hasher.Write([]byte(part))
		_, _ = hasher.Write([]byte{0})
	}

	return hex.EncodeToString(hasher.Sum(nil))
}

type cache struct {
	mu      sync.RWMutex
	max     int
	entries map[string]string
}

func newCache(max int) *cache {
	if max <= 0 {
		max = defaultMaxCacheEntries
	}
	return &cache{max: max, entries: make(map[string]string)}
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
	if _, exists := c.entries[key]; !exists && len(c.entries) >= c.max {
		for existing := range c.entries {
			delete(c.entries, existing)
			break
		}
	}
	c.entries[key] = value
}
