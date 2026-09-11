package llm

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/projectdiscovery/utils/errkit"
	openai "github.com/sashabaranov/go-openai"
)

// openAIBackend speaks /v1/chat/completions. The same wire format covers every
// local runtime and most hosted providers, so one backend serves both.
type openAIBackend struct {
	client *openai.Client
}

func newOpenAIBackend(baseURL, apiKey string, timeout time.Duration) *openAIBackend {
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = baseURL
	config.HTTPClient = &http.Client{Timeout: timeout}

	return &openAIBackend{client: openai.NewClientWithConfig(config)}
}

func (b *openAIBackend) complete(ctx context.Context, model string, req Request) (string, error) {
	messages := make([]openai.ChatCompletionMessage, 0, 2)
	if req.System != "" {
		messages = append(messages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleSystem, Content: req.System})
	}
	messages = append(messages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: req.Prompt})

	// no MaxTokens: reasoning models spend an unpredictable amount before they
	// emit content, and a cap truncates them to an empty response. temperature
	// 0 so a cache miss on the same input tends to reproduce the same answer.
	request := openai.ChatCompletionRequest{
		Model:       model,
		Temperature: 0,
		Messages:    messages,
	}
	if req.Format.JSON {
		request.ResponseFormat = &openai.ChatCompletionResponseFormat{Type: openai.ChatCompletionResponseFormatTypeJSONObject}
	}

	response, err := b.client.CreateChatCompletion(ctx, request)
	if err != nil {
		return "", errkit.Wrap(err, "could not reach llm provider")
	}

	if len(response.Choices) == 0 {
		return "", errkit.New("llm provider returned no choices")
	}

	choice := response.Choices[0]
	// a server-side token cap truncates to empty-ish content rather than
	// erroring, so name the cause instead of surfacing a blank answer
	if choice.FinishReason == openai.FinishReasonLength {
		return "", errkit.Newf("llm response truncated after %d tokens, raise the model output limit", response.Usage.CompletionTokens)
	}

	return strings.TrimSpace(choice.Message.Content), nil
}
