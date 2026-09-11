// Package llm is a small, shared client for large language model providers.
//
// It exists so the ProjectDiscovery ecosystem has one LLM client rather than
// several: nuclei semantic matchers and the DSL llm_prompt helper both build on
// it. Any OpenAI-compatible endpoint is supported, which covers local runtimes
// (ollama, llama.cpp, vLLM, LM Studio) as well as hosted providers, and a local
// model is the intended default so a scan need not send data to a third party.
package llm

import (
	"strings"
	"time"

	"github.com/projectdiscovery/utils/errkit"
)

// APIKeyEnv is the single place an API key is read from. A key is never a
// constructor argument or a flag, so it cannot be captured in a config struct
// or leak through process args and shell history.
const APIKeyEnv = "LLM_API_KEY"

// ProviderOpenAICompatible talks to any /v1/chat/completions endpoint.
const ProviderOpenAICompatible = "openai-compatible"

const defaultTimeout = 2 * time.Minute

// presets are base URLs for well known OpenAI-compatible endpoints. Anything
// else is reached by setting BaseURL directly, so this table only needs the
// common local runtimes and a couple of hosted providers.
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

// Config describes how to reach a provider and how to bound its use.
type Config struct {
	// Provider selects a preset endpoint. Ignored when BaseURL is set.
	Provider string
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
// explicit URL is given.
func (config Config) resolveBaseURL() (string, error) {
	if url := strings.TrimSuffix(config.BaseURL, "/"); url != "" {
		return url, nil
	}

	provider := strings.ToLower(strings.TrimSpace(config.Provider))
	if provider == "" || provider == ProviderOpenAICompatible {
		provider = "openai"
	}

	preset, ok := presets[provider]
	if !ok {
		return "", errkit.Newf("unknown llm provider %q, set a base url for a custom endpoint", config.Provider)
	}

	return preset, nil
}
