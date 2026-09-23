package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/projectdiscovery/utils/errkit"
)

const maxErrorBody = 2048

// openAIBackend speaks /v1/chat/completions over net/http so llm does not add
// a module-wide third-party client dependency.
type openAIBackend struct {
	client  *http.Client
	baseURL string
	apiKey  string
}

type chatRequest struct {
	Model          string              `json:"model"`
	Temperature    float32             `json:"temperature"`
	Messages       []chatMessage       `json:"messages"`
	ResponseFormat *chatResponseFormat `json:"response_format,omitempty"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponseFormat struct {
	Type string `json:"type"`
}

type chatResponse struct {
	Choices []struct {
		FinishReason string `json:"finish_reason"`
		Message      struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
}

func newOpenAIBackend(baseURL, apiKey string, timeout time.Duration) *openAIBackend {
	return &openAIBackend{
		client:  &http.Client{Timeout: timeout},
		baseURL: strings.TrimSuffix(baseURL, "/"),
		apiKey:  apiKey,
	}
}

func (b *openAIBackend) complete(ctx context.Context, model string, req Request) (string, error) {
	messages := make([]chatMessage, 0, 2)
	if req.System != "" {
		messages = append(messages, chatMessage{Role: "system", Content: req.System})
	}
	messages = append(messages, chatMessage{Role: "user", Content: req.Prompt})

	// no max_tokens: reasoning models spend an unpredictable amount before they
	// emit content, and a cap truncates them to an empty response. temperature
	// 0 so a cache miss on the same input tends to reproduce the same answer.
	payload := chatRequest{
		Model:       model,
		Temperature: 0,
		Messages:    messages,
	}
	if req.Format.JSON {
		payload.ResponseFormat = &chatResponseFormat{Type: "json_object"}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", errkit.Wrap(err, "could not encode llm request")
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, b.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", errkit.Wrap(err, "could not build llm request")
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if b.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+b.apiKey)
	}

	resp, err := b.client.Do(httpReq)
	if err != nil {
		return "", errkit.Wrap(err, "could not reach llm provider")
	}
	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", errkit.Wrap(err, "could not read llm response")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := strings.TrimSpace(string(raw))
		if len(msg) > maxErrorBody {
			msg = msg[:maxErrorBody]
		}
		return "", errkit.Newf("llm provider returned status %d: %s", resp.StatusCode, msg)
	}

	var response chatResponse
	if err := json.Unmarshal(raw, &response); err != nil {
		return "", errkit.Wrap(err, "could not decode llm response")
	}
	if len(response.Choices) == 0 {
		return "", errkit.New("llm provider returned no choices")
	}

	choice := response.Choices[0]
	if choice.FinishReason == "length" {
		return "", errkit.Newf("llm response truncated after %d tokens, raise the model output limit", response.Usage.CompletionTokens)
	}

	return strings.TrimSpace(choice.Message.Content), nil
}
