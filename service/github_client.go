package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

const githubAPIBase = "https://api.github.com"

// gitHubClient is a minimal GitHub API client used by SourceService.
type gitHubClient struct {
	http *http.Client
}

func newGitHubClient() *gitHubClient {
	return &gitHubClient{
		http: &http.Client{Timeout: 15 * time.Second},
	}
}

// githubRepoMeta is the subset of fields we care about from GET /repos/{owner}/{repo}
type githubRepoMeta struct {
	DefaultBranch   string `json:"default_branch"`
	Private         bool   `json:"private"`
	Description     string `json:"description"`
	StargazersCount int    `json:"stargazers_count"`
}

// githubFileContent is the response from GET /repos/{owner}/{repo}/contents/{path}
type githubFileContent struct {
	Type     string `json:"type"`     // "file"
	Encoding string `json:"encoding"` // "base64"
	Content  string `json:"content"`
}

// GetRepoMeta fetches public metadata for a GitHub repository.
func (g *gitHubClient) GetRepoMeta(ctx context.Context, owner, repo, token string) (*githubRepoMeta, error) {
	url := fmt.Sprintf("%s/repos/%s/%s", githubAPIBase, owner, repo)
	var meta githubRepoMeta
	if err := g.doGet(ctx, url, token, &meta); err != nil {
		return nil, fmt.Errorf("failed to fetch repo metadata for %s/%s: %w", owner, repo, err)
	}
	return &meta, nil
}

// GetFileContent fetches and base64-decodes the content of a file in a GitHub repo at the given ref.
func (g *gitHubClient) GetFileContent(ctx context.Context, owner, repo, path, ref, token string) ([]byte, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/contents/%s?ref=%s", githubAPIBase, owner, repo, path, ref)
	var fc githubFileContent
	if err := g.doGet(ctx, url, token, &fc); err != nil {
		return nil, fmt.Errorf("failed to fetch %s from %s/%s@%s: %w", path, owner, repo, ref, err)
	}
	if fc.Type != "file" {
		return nil, fmt.Errorf("%s is not a file in %s/%s (type=%s)", path, owner, repo, fc.Type)
	}
	if fc.Encoding != "base64" {
		return nil, fmt.Errorf("unexpected encoding %q for %s in %s/%s", fc.Encoding, path, owner, repo)
	}
	decoded, err := base64.StdEncoding.DecodeString(sanitizeBase64(fc.Content))
	if err != nil {
		return nil, fmt.Errorf("failed to base64-decode %s: %w", path, err)
	}
	return decoded, nil
}

// ArchiveURL returns the GitHub tarball download URL for the given ref.
// For private repos the caller must include a valid token in the HTTP request headers when downloading.
func (g *gitHubClient) ArchiveURL(owner, repo, ref string) string {
	return fmt.Sprintf("%s/repos/%s/%s/tarball/%s", githubAPIBase, owner, repo, ref)
}

func (g *gitHubClient) doGet(ctx context.Context, url, token string, dest interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	slog.Debug("GitHub API request", "url", url)
	resp, err := g.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return json.NewDecoder(resp.Body).Decode(dest)
	case http.StatusNotFound:
		return fmt.Errorf("not found (404): %s", url)
	case http.StatusUnauthorized:
		return fmt.Errorf("unauthorized (401) — provide a GitHub PAT for private repos")
	case http.StatusForbidden:
		return fmt.Errorf("forbidden (403) — rate limit exceeded or missing token scope")
	default:
		return fmt.Errorf("GitHub API error %d: %s", resp.StatusCode, url)
	}
}

// sanitizeBase64 strips newlines GitHub inserts into multi-line base64 content.
func sanitizeBase64(s string) string {
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] != '\n' && s[i] != '\r' {
			out = append(out, s[i])
		}
	}
	return string(out)
}
