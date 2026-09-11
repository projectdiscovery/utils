// Package llm is a small, shared client for large language model providers.
//
// It exists so the ProjectDiscovery ecosystem has one LLM client rather than
// several: nuclei semantic matchers and the DSL llm_prompt helper both build on
// it. Any OpenAI-compatible endpoint is supported, which covers local runtimes
// (ollama, llama.cpp, vLLM, LM Studio) as well as hosted providers. There is no
// implicit hosted default: a caller must set Provider or BaseURL so a scan
// cannot send data off-box by accident.
package llm

import (
	"strings"
	"time"

	"github.com/projectdiscovery/utils/errkit"
)

// APIKeyEnv is the environment variable consulted when Config.APIKey is empty.
const APIKeyEnv = "LLM_API_KEY"

// ProviderOpenAICompatible talks to any /v1/chat/completions endpoint. BaseURL
// is required; this name never maps to a hosted vendor.
const ProviderOpenAICompatible = "openai-compatible"

const defaultTimeout = 2 * time.Minute

// defaultMaxCacheEntries bounds the in-memory response cache. When MaxCalls is
// set and smaller, that value is used instead so the cache cannot outgrow the
// call budget.
const defaultMaxCacheEntries = 4096

// presets are base URLs for well known OpenAI-compatible endpoints. Anything
// else is reached by setting BaseURL directly.
var presets = map[string]string{
	"openai":     "https://api.openai.com/v1",
	"ollama":     "http://localhost:11434/v1",
	"llamacpp":   "http://localhost:8080/v1",
	"vllm":       "http://localhost:8000/v1",
	"lmstudio":   "http://localhost:1234/v1",
	"groq":       "https://api.groq.com/openai/v1",
	"openrouter": "https://openrouter.ai/api/v1",
	"together":   "https://api.together.xyz/v1",
}

var hostedPresets = map[string]struct{}{
	"openai":     {},
	"groq":       {},
	"openrouter": {},
	"together":   {},
}

// Config describes how to reach a provider and how to bound its use.
type Config struct {
	// Provider selects a preset endpoint. Ignored when BaseURL is set.
	Provider string
	// APIKey overrides LLM_API_KEY. Empty means read the environment. Hosted
	// presets require a key; local presets and a custom BaseURL do not.
	APIKey string
	// BaseURL is any OpenAI-compatible endpoint, including a local one.
	BaseURL string
	// Model is the model identifier passed to the provider.
	Model string
	// Timeout bounds a single completion call. Zero uses a default.
	Timeout time.Duration
	// Cache enables in-memory response caching keyed on the request.
	Cache bool
	// MaxCalls caps completions for the lifetime of the client. Zero is
	// unlimited. It exists so a misbehaving caller cannot run up an unbounded
	// bill against a paid provider.
	MaxCalls int
	// MaxConcurrency caps in-flight completions. Zero or negative means one.
	MaxConcurrency int
}

// resolveBaseURL returns the endpoint to talk to, applying a preset when no
// explicit URL is given. An empty provider does not fall through to OpenAI.
func (config Config) resolveBaseURL() (string, error) {
	if url := strings.TrimSuffix(config.BaseURL, "/"); url != "" {
		return url, nil
	}

	provider := strings.ToLower(strings.TrimSpace(config.Provider))
	if provider == "" {
		return "", errkit.New("no llm provider or base url configured")
	}
	if provider == ProviderOpenAICompatible {
		return "", errkit.New("openai-compatible provider requires a base url")
	}

	preset, ok := presets[provider]
	if !ok {
		return "", errkit.Newf("unknown llm provider %q, set a base url for a custom endpoint", config.Provider)
	}

	return preset, nil
}

func requiresAPIKey(provider, baseURL string) bool {
	name := strings.ToLower(strings.TrimSpace(provider))
	if _, ok := hostedPresets[name]; ok {
		return true
	}
	normalized := strings.TrimSuffix(baseURL, "/")
	for hosted := range hostedPresets {
		if presets[hosted] == normalized {
			return true
		}
	}
	return false
}
