package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/ambientlabscomputing/ucrs/sdk/types"
	"gopkg.in/yaml.v3"
)

const (
	manifestPath     = ".underleaf/deploy.yaml"
	maxServices      = 20
	sourceTypeGitHub = "github"
)

// SourceService resolves gh: source references to AppManifest + archive URLs.
type SourceService struct {
	github *gitHubClient
}

// NewSourceService creates a SourceService.
func NewSourceService() *SourceService {
	return &SourceService{
		github: newGitHubClient(),
	}
}

// Resolve parses a source string like "gh:owner/repo" or "gh:owner/repo@ref",
// fetches .underleaf/deploy.yaml via the GitHub API, validates it, and returns
// a ResolvedSource ready for server_api to consume.
func (s *SourceService) Resolve(ctx context.Context, source, refOverride, token string) (*types.ResolvedSource, error) {
	owner, repo, ref, err := parseGHSource(source, refOverride)
	if err != nil {
		return nil, err
	}

	slog.InfoContext(ctx, "resolving source", "owner", owner, "repo", repo, "ref", ref)

	// Fetch repo metadata (to get default branch when ref is empty and to validate existence)
	meta, err := s.github.GetRepoMeta(ctx, owner, repo, token)
	if err != nil {
		return nil, fmt.Errorf("cannot reach repo %s/%s: %w", owner, repo, err)
	}
	if ref == "" {
		ref = meta.DefaultBranch
	}

	// Fetch and parse .underleaf/deploy.yaml
	raw, err := s.github.GetFileContent(ctx, owner, repo, manifestPath, ref, token)
	if err != nil {
		return nil, fmt.Errorf("repo %s/%s@%s does not contain %s: %w", owner, repo, ref, manifestPath, err)
	}

	var manifest types.AppManifest
	if err := yaml.Unmarshal(raw, &manifest); err != nil {
		return nil, fmt.Errorf("invalid YAML in %s for %s/%s@%s: %w", manifestPath, owner, repo, ref, err)
	}

	if err := validateManifest(&manifest, owner, repo); err != nil {
		return nil, err
	}

	// Derive slug from name if not set
	if manifest.Slug == "" {
		manifest.Slug = slugify(manifest.Name)
	}

	archiveURL := s.github.ArchiveURL(owner, repo, ref)

	return &types.ResolvedSource{
		Type:       sourceTypeGitHub,
		Owner:      owner,
		Repo:       repo,
		Ref:        ref,
		ArchiveURL: archiveURL,
		Manifest:   manifest,
		RepoMeta: types.RepoMeta{
			DefaultBranch: meta.DefaultBranch,
			Private:       meta.Private,
			Description:   meta.Description,
			Stars:         meta.StargazersCount,
		},
	}, nil
}

// parseGHSource extracts owner, repo, ref from "gh:owner/repo" or "gh:owner/repo@ref".
// refOverride, if non-empty, supersedes any ref encoded in source.
func parseGHSource(source, refOverride string) (owner, repo, ref string, err error) {
	if !strings.HasPrefix(source, "gh:") {
		return "", "", "", fmt.Errorf("unsupported source prefix — expected \"gh:owner/repo\" (got %q)", source)
	}
	spec := strings.TrimPrefix(source, "gh:")

	// Split on "@" to extract optional ref
	atParts := strings.SplitN(spec, "@", 2)
	if len(atParts) == 2 {
		spec = atParts[0]
		ref = atParts[1]
	}

	// Owner/repo
	parts := strings.SplitN(spec, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", "", fmt.Errorf("invalid source format — expected \"gh:owner/repo\" (got %q)", source)
	}
	owner, repo = parts[0], parts[1]

	if refOverride != "" {
		ref = refOverride
	}
	return
}

// validateManifest performs security and sanity checks on a parsed AppManifest.
func validateManifest(m *types.AppManifest, owner, repo string) error {
	if m.Name == "" {
		return fmt.Errorf("manifest in %s/%s is missing required field \"name\"", owner, repo)
	}
	if m.Version == "" {
		return fmt.Errorf("manifest in %s/%s is missing required field \"version\"", owner, repo)
	}
	if len(m.Services) == 0 {
		return fmt.Errorf("manifest in %s/%s must define at least one service", owner, repo)
	}
	if len(m.Services) > maxServices {
		return fmt.Errorf("manifest in %s/%s defines %d services (max %d)", owner, repo, len(m.Services), maxServices)
	}

	for i, svc := range m.Services {
		if svc.Name == "" {
			return fmt.Errorf("service at index %d in %s/%s is missing \"name\"", i, owner, repo)
		}
		hasImage := svc.Image != ""
		hasBuild := svc.Build != nil
		if !hasImage && !hasBuild {
			return fmt.Errorf("service %q in %s/%s: one of \"image\" or \"build\" is required", svc.Name, owner, repo)
		}
		if hasImage && hasBuild {
			return fmt.Errorf("service %q in %s/%s: \"image\" and \"build\" are mutually exclusive", svc.Name, owner, repo)
		}

		// Security: reject localhost/loopback image references
		if hasImage && (strings.HasPrefix(svc.Image, "localhost:") || strings.HasPrefix(svc.Image, "127.0.0.1:")) {
			return fmt.Errorf("service %q in %s/%s: image %q targets a localhost registry — not allowed for remote deployments", svc.Name, owner, repo, svc.Image)
		}

		// Security: build context must not escape the repo root
		if hasBuild {
			ctx := svc.Build.Context
			if ctx == "" {
				ctx = "."
			}
			if strings.Contains(ctx, "..") {
				return fmt.Errorf("service %q in %s/%s: build context %q must not escape the repo root", svc.Name, owner, repo, ctx)
			}
		}
	}

	return nil
}

// slugify converts a name to a URL-safe slug (lowercase, spaces → hyphens, strip non-alnum).
func slugify(name string) string {
	name = strings.ToLower(name)
	var b strings.Builder
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-':
			b.WriteRune(r)
		case r == ' ', r == '_':
			b.WriteRune('-')
		}
	}
	return strings.Trim(b.String(), "-")
}
