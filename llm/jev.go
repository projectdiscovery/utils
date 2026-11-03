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

// jevBackend speaks TypeSafe's System One API. Jev returns calibrated
// probabilities for typed questions instead of generated text, so the inputs
// travel as state, apart from the question, and cannot steer the answer into
// free-form output.
type jevBackend struct {
	client  *http.Client
	baseURL string
	apiKey  string
}

type jevRequest struct {
	Model     string                 `json:"model"`
	State     any                    `json:"state"`
	Questions map[string]jevQuestion `json:"questions"`
}

type jevQuestion struct {
	Type         string             `json:"type"`
	Instructions string             `json:"instructions"`
	Criteria     map[string]*string `json:"criteria,omitempty"`
}

type jevResponse struct {
	Model   string `json:"model"`
	Answers map[string]struct {
		Type          string             `json:"type"`
		Noul          *float64           `json:"noul"`
		Choice        string             `json:"choice"`
		Probabilities map[string]float64 `json:"probabilities"`
	} `json:"answers"`
}

// jevQuestionName keys the single question each request carries.
const jevQuestionName = "q"

func newJevBackend(baseURL, apiKey string, timeout time.Duration) *jevBackend {
	return &jevBackend{
		client:  &http.Client{Timeout: timeout},
		baseURL: strings.TrimSuffix(baseURL, "/"),
		apiKey:  apiKey,
	}
}

func (b *jevBackend) complete(context.Context, string, Request) (string, error) {
	return "", errkit.New("jev answers judge requests only, not free-text completions")
}

func (b *jevBackend) judge(ctx context.Context, model string, req JudgeRequest) (Verdict, error) {
	question := jevQuestion{Type: "noul", Instructions: req.Question}
	if len(req.Labels) > 0 {
		question.Type = "choice"
		question.Criteria = make(map[string]*string, len(req.Labels))
		for _, label := range req.Labels {
			question.Criteria[label] = nil
		}
	}

	var state any = req.Inputs
	if len(req.Inputs) == 1 {
		state = req.Inputs[0]
	}

	body, err := json.Marshal(jevRequest{
		Model:     model,
		State:     state,
		Questions: map[string]jevQuestion{jevQuestionName: question},
	})
	if err != nil {
		return Verdict{}, errkit.Wrap(err, "could not encode jev request")
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, b.baseURL+"/systemone", bytes.NewReader(body))
	if err != nil {
		return Verdict{}, errkit.Wrap(err, "could not build jev request")
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+b.apiKey)

	resp, err := b.client.Do(httpReq)
	if err != nil {
		return Verdict{}, errkit.Wrap(err, "could not reach jev")
	}
	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return Verdict{}, errkit.Wrap(err, "could not read jev response")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := strings.TrimSpace(string(raw))
		if len(msg) > maxErrorBody {
			msg = msg[:maxErrorBody]
		}
		return Verdict{}, errkit.Newf("jev returned status %d: %s", resp.StatusCode, msg)
	}

	var response jevResponse
	if err := json.Unmarshal(raw, &response); err != nil {
		return Verdict{}, errkit.Wrap(err, "could not decode jev response")
	}
	answer, ok := response.Answers[jevQuestionName]
	if !ok {
		return Verdict{}, errkit.New("jev returned no answer")
	}

	switch {
	case question.Type == "noul" && answer.Noul != nil:
		if *answer.Noul >= 0.5 {
			return Verdict{Label: "yes", Confidence: *answer.Noul, Model: response.Model}, nil
		}
		return Verdict{Label: "no", Confidence: 1 - *answer.Noul, Model: response.Model}, nil
	case question.Type == "choice" && answer.Choice != "":
		probability, ok := answer.Probabilities[answer.Choice]
		if !ok {
			return Verdict{}, errkit.Newf("jev returned no probability for %q", answer.Choice)
		}
		return Verdict{Label: answer.Choice, Confidence: probability, Model: response.Model}, nil
	}
	return Verdict{}, errkit.Newf("jev returned an unexpected %q answer", answer.Type)
}
