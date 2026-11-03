package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

// jevStub serves canned System One answers and records each request body.
func jevStub(t *testing.T, answer string, seen *[]jevRequest, calls *int32) *Client {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v1/systemone", r.URL.Path)
		require.Equal(t, "Bearer key", r.Header.Get("Authorization"))
		if calls != nil {
			atomic.AddInt32(calls, 1)
		}
		if seen != nil {
			var req jevRequest
			require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
			*seen = append(*seen, req)
		}
		_, _ = w.Write([]byte(`{"model":"jev-1.14.0","answers":{"q":` + answer + `},"usage":{"input_tokens":10}}`))
	}))
	t.Cleanup(server.Close)

	client, err := New(Config{Provider: ProviderJev, BaseURL: server.URL + "/v1", APIKey: "key", Model: "jev-latest", Cache: true})
	require.NoError(t, err)
	return client
}

func TestJevYesNoAsksNoulAndReportsChosenSideProbability(t *testing.T) {
	var seen []jevRequest
	client := jevStub(t, `{"type":"noul","noul":0.2}`, &seen, nil)
	require.True(t, client.CanJudge())

	verdict, err := client.Judge(context.Background(), JudgeRequest{Question: "is it a login page?", Inputs: []string{"<form>"}})
	require.NoError(t, err)
	require.Equal(t, "no", verdict.Label)
	require.InDelta(t, 0.8, verdict.Confidence, 1e-9)
	require.Equal(t, "jev-1.14.0", verdict.Model)

	require.Len(t, seen, 1)
	require.Equal(t, "jev-latest", seen[0].Model)
	require.Equal(t, "<form>", seen[0].State, "a single input is sent as plain state")
	require.Equal(t, jevQuestion{Type: "noul", Instructions: "is it a login page?"}, seen[0].Questions[jevQuestionName])
}

func TestJevLabelsAskChoiceAndSendSeveralInputsAsArray(t *testing.T) {
	var seen []jevRequest
	client := jevStub(t, `{"type":"choice","choice":"different","confidence":0.9,"probabilities":{"same":0.1,"different":0.9}}`, &seen, nil)

	verdict, err := client.Judge(context.Background(), JudgeRequest{Question: "same access?", Labels: []string{"same", "different"}, Inputs: []string{"a", "b"}})
	require.NoError(t, err)
	require.Equal(t, Verdict{Label: "different", Confidence: 0.9, Model: "jev-1.14.0"}, verdict)

	require.Equal(t, []any{"a", "b"}, seen[0].State)
	question := seen[0].Questions[jevQuestionName]
	require.Equal(t, "choice", question.Type)
	require.Contains(t, question.Criteria, "same")
	require.Contains(t, question.Criteria, "different")
}

func TestJevJudgeSharesCacheAndBudget(t *testing.T) {
	var calls int32
	client := jevStub(t, `{"type":"noul","noul":0.9}`, nil, &calls)
	client.budget = 1

	req := JudgeRequest{Question: "q", Inputs: []string{"body"}}
	for range 2 {
		_, err := client.Judge(context.Background(), req)
		require.NoError(t, err)
	}
	require.EqualValues(t, 1, calls, "repeat is served from cache")

	_, err := client.Judge(context.Background(), JudgeRequest{Question: "q", Inputs: []string{"other"}})
	require.ErrorContains(t, err, "budget exhausted")
}

func TestJevRejectsCompleteAndMalformedAnswers(t *testing.T) {
	client := jevStub(t, `{"type":"score"}`, nil, nil)

	_, err := client.Complete(context.Background(), Request{Prompt: "hi"})
	require.ErrorContains(t, err, "judge requests only")

	_, err = client.Judge(context.Background(), JudgeRequest{Question: "q", Inputs: []string{"x"}})
	require.ErrorContains(t, err, "unexpected")
}

func TestJevReportsProviderErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"detail":{"message":"Unknown model: jev-nope"}}`))
	}))
	t.Cleanup(server.Close)
	client, err := New(Config{Provider: ProviderJev, BaseURL: server.URL + "/v1", APIKey: "key", Model: "jev-nope"})
	require.NoError(t, err)

	_, err = client.Judge(context.Background(), JudgeRequest{Question: "q", Inputs: []string{"x"}})
	require.ErrorContains(t, err, "Unknown model")
}

func TestJevPresetRequiresKey(t *testing.T) {
	t.Setenv(APIKeyEnv, "")
	_, err := New(Config{Provider: ProviderJev, Model: "jev-latest"})
	require.ErrorContains(t, err, "api key required")
}

func TestChatProviderCannotJudge(t *testing.T) {
	client, err := New(Config{Model: "m", BaseURL: "http://x/v1"})
	require.NoError(t, err)
	require.False(t, client.CanJudge())

	_, err = client.Judge(context.Background(), JudgeRequest{Question: "q", Inputs: []string{"x"}})
	require.ErrorContains(t, err, "does not support judge")
}
