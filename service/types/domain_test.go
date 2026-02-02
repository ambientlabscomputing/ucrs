package types

import (
	"testing"
	"time"
)

func TestRiskClass_String(t *testing.T) {
	tests := []struct {
		name     string
		rc       RiskClass
		expected string
	}{
		{"Low risk", RiskLow, "low"},
		{"Medium risk", RiskMedium, "medium"},
		{"High risk", RiskHigh, "high"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.rc) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, string(tt.rc))
			}
		})
	}
}

func TestTrustTier_String(t *testing.T) {
	tests := []struct {
		name     string
		tt       TrustTier
		expected string
	}{
		{"Official", TrustOfficial, "official"},
		{"Certified", TrustCertified, "certified"},
		{"Community", TrustCommunity, "community"},
		{"Experimental", TrustExperimental, "experimental"},
		{"Local", TrustLocal, "local"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if string(test.tt) != test.expected {
				t.Errorf("expected %s, got %s", test.expected, string(test.tt))
			}
		})
	}
}

func TestArtifactType_String(t *testing.T) {
	tests := []struct {
		name     string
		at       ArtifactType
		expected string
	}{
		{"OCI", ArtifactOCI, "oci"},
		{"Binary", ArtifactBinary, "binary"},
		{"Git", ArtifactGit, "git"},
		{"NPM", ArtifactNPM, "npm"},
		{"PyPI", ArtifactPyPI, "pypi"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.at) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, string(tt.at))
			}
		})
	}
}

func TestCapability_Creation(t *testing.T) {
	now := time.Now()
	cap := Capability{
		ID:              "iot.light.control",
		Version:         "1.0.0",
		Description:     "Control smart lights",
		RiskClass:       RiskMedium,
		AllowedVerbs:    []string{"on", "off", "dim"},
		PermissionClass: "device.control",
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if cap.ID != "iot.light.control" {
		t.Errorf("expected ID iot.light.control, got %s", cap.ID)
	}
	if cap.Version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", cap.Version)
	}
	if cap.RiskClass != RiskMedium {
		t.Errorf("expected risk class medium, got %s", cap.RiskClass)
	}
	if len(cap.AllowedVerbs) != 3 {
		t.Errorf("expected 3 allowed verbs, got %d", len(cap.AllowedVerbs))
	}
}

func TestProvider_Creation(t *testing.T) {
	now := time.Now()
	provider := Provider{
		ProviderID: "ambient.hue-mcp",
		Version:    "2.1.0",
		Capabilities: []CapabilityRef{
			{ID: "iot.light.control", VersionRange: "^1.0"},
		},
		Artifact: Artifact{
			Type:   ArtifactOCI,
			URI:    "registry.underleaf.io/ambient/hue-mcp:2.1.0",
			Digest: "sha256:abc123",
		},
		TrustTier:      TrustOfficial,
		SandboxProfile: "standard",
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if provider.ProviderID != "ambient.hue-mcp" {
		t.Errorf("expected provider ID ambient.hue-mcp, got %s", provider.ProviderID)
	}
	if provider.TrustTier != TrustOfficial {
		t.Errorf("expected trust tier official, got %s", provider.TrustTier)
	}
	if len(provider.Capabilities) != 1 {
		t.Errorf("expected 1 capability, got %d", len(provider.Capabilities))
	}
}

func TestRegistrySnapshot_Creation(t *testing.T) {
	snapshot := RegistrySnapshot{
		Version:   "1234567890",
		Timestamp: time.Now(),
		Capabilities: []Capability{
			{ID: "test.capability", Version: "1.0.0"},
		},
		Providers: []Provider{
			{ProviderID: "test.provider", Version: "1.0.0"},
		},
		Manifest: SignedManifest{
			KeyID:     "key-001",
			Signature: "base64sig",
			Algorithm: "Ed25519",
			SignedAt:  time.Now(),
		},
		ETag: "\"abc123\"",
	}

	if snapshot.Version != "1234567890" {
		t.Errorf("expected version 1234567890, got %s", snapshot.Version)
	}
	if len(snapshot.Capabilities) != 1 {
		t.Errorf("expected 1 capability, got %d", len(snapshot.Capabilities))
	}
	if len(snapshot.Providers) != 1 {
		t.Errorf("expected 1 provider, got %d", len(snapshot.Providers))
	}
	if snapshot.Manifest.Algorithm != "Ed25519" {
		t.Errorf("expected algorithm Ed25519, got %s", snapshot.Manifest.Algorithm)
	}
}

func TestDeltaUpdate_Creation(t *testing.T) {
	delta := DeltaUpdate{
		SinceVersion: "1000",
		ToVersion:    "2000",
		Timestamp:    time.Now(),
		AddedCapabilities: []Capability{
			{ID: "new.capability", Version: "1.0.0"},
		},
		UpdatedCapabilities: []Capability{
			{ID: "updated.capability", Version: "1.1.0"},
		},
		RemovedCapabilities: []string{"old.capability"},
		AddedProviders: []Provider{
			{ProviderID: "new.provider", Version: "1.0.0"},
		},
		UpdatedProviders: []Provider{
			{ProviderID: "updated.provider", Version: "2.0.0"},
		},
		RemovedProviders: []string{"old.provider:1.0.0"},
	}

	if delta.SinceVersion != "1000" {
		t.Errorf("expected since version 1000, got %s", delta.SinceVersion)
	}
	if delta.ToVersion != "2000" {
		t.Errorf("expected to version 2000, got %s", delta.ToVersion)
	}
	if len(delta.AddedCapabilities) != 1 {
		t.Errorf("expected 1 added capability, got %d", len(delta.AddedCapabilities))
	}
	if len(delta.RemovedCapabilities) != 1 {
		t.Errorf("expected 1 removed capability, got %d", len(delta.RemovedCapabilities))
	}
}
