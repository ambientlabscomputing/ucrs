package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type TokenManager interface {
	GetToken(ctx context.Context) (string, error)
}

type AppTokenManager struct {
	settings  *Settings
	client    *http.Client
	token     string
	expiresAt int64
}

func NewAppTokenManager(settings *Settings, client *http.Client) *AppTokenManager {
	return &AppTokenManager{
		settings: settings,
		client:   client,
	}
}

func (tm *AppTokenManager) GetToken(ctx context.Context) (string, error) {
	logger := GetLogger(ctx)

	// Check if we have a cached token that's still valid
	if tm.token != "" && time.Now().Unix() < tm.expiresAt-60 {
		logger.Debug("using cached token", "len", len(tm.token), "expires_at", time.Unix(int64(tm.expiresAt), 0).Format(time.RFC3339))
		return tm.token, nil
	}

	// Fetch a new token
	if err := tm.fetchNewToken(ctx); err != nil {
		logger.Error("failed to fetch new token", "error", err)
		return "", err
	}

	return tm.token, nil
}

func (tm *AppTokenManager) fetchNewToken(ctx context.Context) error {
	logger := GetLogger(ctx)

	tokenURL := tm.settings.Auth.AuthDomain
	if !strings.HasPrefix(tokenURL, "http://") && !strings.HasPrefix(tokenURL, "https://") {
		tokenURL = "https://" + tokenURL
	}
	tokenURL = tokenURL + "/oauth/token"

	payload := map[string]interface{}{
		"grant_type": "password",
		"username":   tm.settings.Auth.Username,
		"password":   tm.settings.Auth.Password,
		"audience":   tm.settings.Auth.AuthAudience,
		"client_id":  tm.settings.Auth.AuthClientID,
		"scope":      tm.settings.Auth.Scopes,
	}

	logger.Debug("fetching new token", "url", tokenURL, "payload", payload)

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		logger.Error("failed to marshal payload", "error", err)
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		logger.Error("failed to create new request", "error", err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	if traceID := GetTraceID(ctx); traceID != "" {
		req.Header.Set("X-Trace-ID", traceID)
	}

	resp, err := tm.client.Do(req)
	if err != nil {
		logger.Error("failed to do request", "error", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		logger.Error("failed to fetch token", "status_code", resp.StatusCode, "body", string(bodyBytes))
		return fmt.Errorf("failed to fetch token: %s", string(bodyBytes))
	}

	var tokenResponse struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
	}

	err = json.NewDecoder(resp.Body).Decode(&tokenResponse)
	if err != nil {
		logger.Error("failed to decode token response", "error", err)
		return err
	}

	tm.token = tokenResponse.AccessToken
	tm.expiresAt = time.Now().Unix() + tokenResponse.ExpiresIn

	logger.Info("fetched new token", "len", len(tm.token), "expires_at", time.Unix(int64(tm.expiresAt), 0).Format(time.RFC3339))

	return nil
}

// ExtractJWTFromToken extracts the JWT portion from an API token or returns the token as-is
// API tokens have format: uf_<random>:<JWT>
// This function returns just the JWT portion for validation purposes
func ExtractJWTFromToken(token string) string {
	// If this is an API token (uf_ prefix), extract just the JWT portion
	if strings.HasPrefix(token, "uf_") {
		parts := strings.SplitN(token, ":", 2)
		if len(parts) == 2 {
			return parts[1] // Return just the JWT
		}
	}
	// Otherwise return the token as-is (regular JWT or Auth0 token)
	return token
}
