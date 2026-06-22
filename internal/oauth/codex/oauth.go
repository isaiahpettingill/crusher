// Package codex provides OpenAI Codex subscription authentication.
package codex

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/charmbracelet/crusher/internal/oauth"
)

const (
	ProviderID = "codex"

	clientID        = "app_EMoamEEZ73f0CkXaXp7hrann"
	issuer          = "https://auth.openai.com"
	CodexEndpoint   = "https://chatgpt.com/backend-api/codex/responses"
	DeviceURL       = issuer + "/codex/device"
	pollSafetyDelay = 3 * time.Second
	userAgent       = "crusher"
)

type DeviceCode struct {
	DeviceAuthID string `json:"device_auth_id"`
	UserCode     string `json:"user_code"`
	Interval     string `json:"interval"`
}

type deviceToken struct {
	AuthorizationCode string `json:"authorization_code"`
	CodeVerifier      string `json:"code_verifier"`
}

type tokenResponse struct {
	IDToken      string `json:"id_token"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

// RequestDeviceCode starts the ChatGPT Codex device authorization flow.
func RequestDeviceCode(ctx context.Context) (*DeviceCode, error) {
	body := strings.NewReader(fmt.Sprintf(`{"client_id":%q}`, clientID))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, issuer+"/api/accounts/deviceauth/usercode", body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)

	var dc DeviceCode
	if err := doJSON(req, &dc); err != nil {
		return nil, err
	}
	return &dc, nil
}

// PollForToken waits for the user to authorize the device flow.
func PollForToken(ctx context.Context, dc *DeviceCode) (*oauth.Token, error) {
	interval := 5 * time.Second
	if parsed, err := time.ParseDuration(dc.Interval + "s"); err == nil && parsed > 0 {
		interval = parsed
	}
	ticker := time.NewTicker(interval + pollSafetyDelay)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
		}

		deviceToken, err := tryDeviceToken(ctx, dc)
		if err == errPending {
			continue
		}
		if err != nil {
			return nil, err
		}
		return ExchangeCode(ctx, deviceToken.AuthorizationCode, issuer+"/deviceauth/callback", deviceToken.CodeVerifier)
	}
}

// RefreshToken refreshes a Codex OAuth token.
func RefreshToken(ctx context.Context, refreshToken string) (*oauth.Token, error) {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)
	data.Set("client_id", clientID)
	return tokenRequest(ctx, data)
}

// ExchangeCode exchanges an OpenAI authorization code for tokens.
func ExchangeCode(ctx context.Context, code, redirectURI, verifier string) (*oauth.Token, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)
	data.Set("client_id", clientID)
	data.Set("code_verifier", verifier)
	return tokenRequest(ctx, data)
}

func tokenRequest(ctx context.Context, data url.Values) (*oauth.Token, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, issuer+"/oauth/token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", userAgent)

	var response tokenResponse
	if err := doJSON(req, &response); err != nil {
		return nil, err
	}
	token := &oauth.Token{
		AccessToken:  response.AccessToken,
		RefreshToken: response.RefreshToken,
		ExpiresIn:    response.ExpiresIn,
		AccountID:    extractAccountID(response),
	}
	token.SetExpiresAt()
	return token, nil
}

var errPending = fmt.Errorf("authorization pending")

func tryDeviceToken(ctx context.Context, dc *DeviceCode) (*deviceToken, error) {
	body := strings.NewReader(fmt.Sprintf(`{"device_auth_id":%q,"user_code":%q}`, dc.DeviceAuthID, dc.UserCode))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, issuer+"/api/accounts/deviceauth/token", body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusNotFound {
		return nil, errPending
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("device authorization failed: %s - %s", resp.Status, strings.TrimSpace(string(body)))
	}
	var token deviceToken
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, err
	}
	return &token, nil
}

func doJSON(req *http.Request, out any) error {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed: %s - %s", resp.Status, strings.TrimSpace(string(body)))
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func extractAccountID(tokens tokenResponse) string {
	for _, token := range []string{tokens.IDToken, tokens.AccessToken} {
		if accountID := accountIDFromJWT(token); accountID != "" {
			return accountID
		}
	}
	return ""
}

func accountIDFromJWT(token string) string {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return ""
	}
	data, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return ""
	}
	var claims struct {
		ChatGPTAccountID string `json:"chatgpt_account_id"`
		OpenAIAuth       struct {
			ChatGPTAccountID string `json:"chatgpt_account_id"`
		} `json:"https://api.openai.com/auth"`
		Organizations []struct {
			ID string `json:"id"`
		} `json:"organizations"`
	}
	if err := json.Unmarshal(data, &claims); err != nil {
		return ""
	}
	if claims.ChatGPTAccountID != "" {
		return claims.ChatGPTAccountID
	}
	if claims.OpenAIAuth.ChatGPTAccountID != "" {
		return claims.OpenAIAuth.ChatGPTAccountID
	}
	if len(claims.Organizations) > 0 {
		return claims.Organizations[0].ID
	}
	return ""
}

// AuthorizeURL builds the browser OAuth URL. It is kept for future UI flows.
func AuthorizeURL(redirectURI, verifier, state string) string {
	challengeBytes := sha256.Sum256([]byte(verifier))
	params := url.Values{}
	params.Set("response_type", "code")
	params.Set("client_id", clientID)
	params.Set("redirect_uri", redirectURI)
	params.Set("scope", "openid profile email offline_access")
	params.Set("code_challenge", base64.RawURLEncoding.EncodeToString(challengeBytes[:]))
	params.Set("code_challenge_method", "S256")
	params.Set("id_token_add_organizations", "true")
	params.Set("codex_cli_simplified_flow", "true")
	params.Set("state", state)
	params.Set("originator", "crusher")
	return issuer + "/oauth/authorize?" + params.Encode()
}

func Verifier() string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~"
	b := make([]byte, 43)
	for i := range b {
		b[i] = chars[rand.IntN(len(chars))]
	}
	return string(b)
}
