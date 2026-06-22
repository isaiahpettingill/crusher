package codex

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestApplyFastMode(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequest(http.MethodPost, "https://api.openai.com/v1/responses", strings.NewReader(`{"model":"gpt-5.5-fast"}`))
	require.NoError(t, err)

	require.NoError(t, normalizeRequest(req))

	body, err := io.ReadAll(req.Body)
	require.NoError(t, err)
	var payload map[string]string
	require.NoError(t, json.Unmarshal(body, &payload))
	require.Equal(t, "gpt-5.5", payload["model"])
	require.Equal(t, "priority", payload["service_tier"])
	require.Equal(t, int64(len(body)), req.ContentLength)
}

func TestNormalizeRequestLeavesOtherModelsAlone(t *testing.T) {
	t.Parallel()

	const body = `{"model":"gpt-5.5"}`
	req, err := http.NewRequest(http.MethodPost, "https://api.openai.com/v1/responses", strings.NewReader(body))
	require.NoError(t, err)

	require.NoError(t, normalizeRequest(req))

	got, err := io.ReadAll(req.Body)
	require.NoError(t, err)
	var payload map[string]string
	require.NoError(t, json.Unmarshal(got, &payload))
	require.Equal(t, "gpt-5.5", payload["model"])
	require.Equal(t, defaultInstructions, payload["instructions"])
}

func TestNormalizeRequestMovesSystemInputToInstructions(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequest(http.MethodPost, "https://api.openai.com/v1/responses", strings.NewReader(`{
		"model":"gpt-5.5",
		"input":[
			{"role":"system","content":"You are direct."},
			{"role":"user","content":"Explain this repo."}
		]
	}`))
	require.NoError(t, err)

	require.NoError(t, normalizeRequest(req))

	body, err := io.ReadAll(req.Body)
	require.NoError(t, err)
	var payload struct {
		Instructions string `json:"instructions"`
		Input        []struct {
			Role string `json:"role"`
		} `json:"input"`
	}
	require.NoError(t, json.Unmarshal(body, &payload))
	require.Equal(t, "You are direct.", payload.Instructions)
	require.Len(t, payload.Input, 1)
	require.Equal(t, "user", payload.Input[0].Role)
}

func TestNormalizeRequestAddsFallbackInstructions(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequest(http.MethodPost, "https://api.openai.com/v1/responses", strings.NewReader(`{
		"model":"gpt-5.5",
		"input":[{"role":"user","content":"Explain this repo."}]
	}`))
	require.NoError(t, err)

	require.NoError(t, normalizeRequest(req))

	body, err := io.ReadAll(req.Body)
	require.NoError(t, err)
	var payload struct {
		Instructions string `json:"instructions"`
	}
	require.NoError(t, json.Unmarshal(body, &payload))
	require.Equal(t, defaultInstructions, payload.Instructions)
}

func TestNormalizeRequestRemovesMaxOutputTokens(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequest(http.MethodPost, "https://api.openai.com/v1/responses", strings.NewReader(`{
		"model":"gpt-5.5",
		"input":[{"role":"user","content":"Explain this repo."}],
		"max_output_tokens":1024
	}`))
	require.NoError(t, err)

	require.NoError(t, normalizeRequest(req))

	body, err := io.ReadAll(req.Body)
	require.NoError(t, err)
	var payload map[string]any
	require.NoError(t, json.Unmarshal(body, &payload))
	require.NotContains(t, payload, "max_output_tokens")
	require.Equal(t, "gpt-5.5", payload["model"])
}
