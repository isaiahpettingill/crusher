package codex

import (
	"net/http"
	"net/url"

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
		url := *req.URL
		codexURL := mustParseCodexEndpoint()
		url.Scheme = codexURL.Scheme
		url.Host = codexURL.Host
		url.Path = codexURL.Path
		req.URL = &url
	}
	return t.transport.RoundTrip(req)
}

func mustParseCodexEndpoint() *url.URL {
	url, err := url.Parse(CodexEndpoint)
	if err != nil {
		panic(err)
	}
	return url
}
