package codex

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/charmbracelet/crusher/internal/log"
	"github.com/charmbracelet/crusher/internal/oauth"
)

// NewClient rewrites OpenAI Responses requests to ChatGPT's Codex endpoint.
func NewClient(token *oauth.Token, debug bool) *http.Client {
	transport := http.DefaultTransport
	if debug {
		transport = log.NewHTTPClient().Transport
	}
	return &http.Client{Transport: &transportWithAccount{token: token, transport: transport}}
}

type transportWithAccount struct {
	token     *oauth.Token
	transport http.RoundTripper
}

func (t *transportWithAccount) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	if t.token != nil && t.token.AccountID != "" {
		req.Header.Set("ChatGPT-Account-Id", t.token.AccountID)
	}
	if req.URL.Path == "/v1/responses" || req.URL.Path == "/responses" || req.URL.Path == "/chat/completions" {
		if err := normalizeRequest(req); err != nil {
			return nil, err
		}
		url := *req.URL
		codexURL := mustParseCodexEndpoint()
		url.Scheme = codexURL.Scheme
		url.Host = codexURL.Host
		url.Path = codexURL.Path
		req.URL = &url
	}
	return t.transport.RoundTrip(req)
}

func normalizeRequest(req *http.Request) error {
	if req.Body == nil {
		return nil
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return err
	}
	if err := req.Body.Close(); err != nil {
		return err
	}

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		restoreBody(req, body)
		return nil
	}

	changed := false
	if payload["model"] == "gpt-5.5-fast" {
		payload["model"] = "gpt-5.5"
		payload["service_tier"] = "priority"
		changed = true
	}
	if applyInstructions(payload) {
		changed = true
	}
	if _, ok := payload["max_output_tokens"]; ok {
		delete(payload, "max_output_tokens")
		changed = true
	}
	if !changed {
		restoreBody(req, body)
		return nil
	}

	body, err = json.Marshal(payload)
	if err != nil {
		return err
	}
	restoreBody(req, body)
	return nil
}

func applyInstructions(payload map[string]any) bool {
	if instructions, ok := payload["instructions"].(string); ok && instructions != "" {
		return false
	}

	changed := false
	for _, key := range []string{"input", "messages"} {
		items, ok := payload[key].([]any)
		if !ok {
			continue
		}
		instructions, rest := extractSystemInstructions(items)
		if instructions == "" {
			continue
		}
		payload["instructions"] = instructions
		payload[key] = rest
		changed = true
		break
	}
	if instructions, ok := payload["instructions"].(string); !ok || instructions == "" {
		payload["instructions"] = defaultInstructions
		changed = true
	}
	return changed
}

const defaultInstructions = "You are Crusher, a terminal AI coding assistant. Be direct, accurate, and helpful."

func extractSystemInstructions(items []any) (string, []any) {
	var instructions []string
	rest := make([]any, 0, len(items))
	for _, item := range items {
		obj, ok := item.(map[string]any)
		if !ok || obj["role"] != "system" {
			rest = append(rest, item)
			continue
		}
		content := contentText(obj["content"])
		if content != "" {
			instructions = append(instructions, content)
		}
	}
	return strings.Join(instructions, "\n\n"), rest
}

func contentText(content any) string {
	switch content := content.(type) {
	case string:
		return content
	case []any:
		var parts []string
		for _, part := range content {
			obj, ok := part.(map[string]any)
			if !ok {
				continue
			}
			text, _ := obj["text"].(string)
			if text != "" {
				parts = append(parts, text)
			}
		}
		return strings.Join(parts, "\n")
	default:
		return ""
	}
}

func restoreBody(req *http.Request, body []byte) {
	req.Body = io.NopCloser(bytes.NewReader(body))
	req.ContentLength = int64(len(body))
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(body)), nil
	}
}

func mustParseCodexEndpoint() *url.URL {
	url, err := url.Parse(CodexEndpoint)
	if err != nil {
		panic(err)
	}
	return url
}
