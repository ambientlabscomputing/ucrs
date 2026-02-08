package utils

import (
	"os"
	"testing"
)

func TestSecretString_String(t *testing.T) {
	secret := SecretString("my-secret-password")
	if secret.String() != "***REDACTED***" {
		t.Errorf("expected ***REDACTED***, got %s", secret.String())
	}
}

func TestSecretString_Value(t *testing.T) {
	secret := SecretString("my-secret-password")
	if secret.Value() != "my-secret-password" {
		t.Errorf("expected my-secret-password, got %s", secret.Value())
	}
}

func TestSecretString_Empty(t *testing.T) {
	secret := SecretString("")
	if !secret.Empty() {
		t.Error("expected empty secret to be true")
	}

	secret = SecretString("not-empty")
	if secret.Empty() {
		t.Error("expected non-empty secret to be false")
	}
}

func TestSettings_Validate(t *testing.T) {
	tests := []struct {
		name      string
		settings  Settings
		expectErr bool
	}{
		{
			name: "valid settings with file",
			settings: Settings{
				LogToStderr: false,
				LogToFile:   true,
			},
			expectErr: false,
		},
		{
			name: "valid settings with both",
			settings: Settings{
				LogToStderr: true,
				LogToFile:   true,
			},
			expectErr: false,
		},
		{
			name: "invalid settings with neither",
			settings: Settings{
				LogToStderr: false,
				LogToFile:   false,
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.settings.Validate()
			if (err != nil) != tt.expectErr {
				t.Errorf("expected error: %v, got: %v", tt.expectErr, err)
			}
		})
	}
}

func TestLoadSettings_FromFile(t *testing.T) {
	// Create a temporary config file
	configContent := `port: "8083"
address: "127.0.0.1"
base_path: "/api/v1/test"
log_level: "info"
log_format: "json"
log_to_stderr: true
log_to_file: false
auth:
  auth_domain: "test.auth0.com"
  auth_audience: "test-api"
  auth_client_id: "test-client"
  username: "test-user"
  password: "test-pass"
  scopes: "read:test"
mongo:
  uri: "mongodb://localhost:27017"
  username: "test"
  password: "test"
  database: "test_db"
event_bus:
  enabled: false
  url: ""
signing:
  private_key_path: "/tmp/private.key"
  public_key_path: "/tmp/public.key"
  key_id: "test-key-001"
`
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(configContent)); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}
	tmpFile.Close()

	// Set CONFIG_PATH environment variable
	os.Setenv("CONFIG_PATH", tmpFile.Name())
	defer os.Unsetenv("CONFIG_PATH")

	// Verify settings
	settings := LoadSettings()
	if settings.Port != "8083" {
		t.Errorf("expected port 8083, got %s", settings.Port)
	}
	if settings.Address != "127.0.0.1" {
		t.Errorf("expected address 127.0.0.1, got %s", settings.Address)
	}
	if settings.BasePath != "/api/v1/test" {
		t.Errorf("expected base path /api/v1/test, got %s", settings.BasePath)
	}
	if settings.LogLevel != "info" {
		t.Errorf("expected log level info, got %s", settings.LogLevel)
	}
	if settings.LogFormat != "json" {
		t.Errorf("expected log format json, got %s", settings.LogFormat)
	}
	if !settings.LogToStderr {
		t.Error("expected log to stderr true")
	}
	if settings.LogToFile {
		t.Error("expected log to file false")
	}
	if settings.Auth.AuthDomain != "test.auth0.com" {
		t.Errorf("expected auth domain test.auth0.com, got %s", settings.Auth.AuthDomain)
	}
	if settings.Mongo.Database != "test_db" {
		t.Errorf("expected database test_db, got %s", settings.Mongo.Database)
	}
	if settings.Signing.KeyID != "test-key-001" {
		t.Errorf("expected key ID test-key-001, got %s", settings.Signing.KeyID)
	}
}

func TestDefaults(t *testing.T) {
	if Defaults["Port"] != "8083" {
		t.Errorf("expected default port 8083, got %v", Defaults["Port"])
	}
	if Defaults["Address"] != "0.0.0.0" {
		t.Errorf("expected default address 0.0.0.0, got %v", Defaults["Address"])
	}
	if Defaults["BasePath"] != "/api/v1/registry" {
		t.Errorf("expected default base path /api/v1/registry, got %v", Defaults["BasePath"])
	}
	if Defaults["LogLevel"] != "debug" {
		t.Errorf("expected default log level debug, got %v", Defaults["LogLevel"])
	}
	if Defaults["LogFormat"] != "text" {
		t.Errorf("expected default log format text, got %v", Defaults["LogFormat"])
	}
	if Defaults["LogToStderr"] != true {
		t.Errorf("expected default log to stderr true, got %v", Defaults["LogToStderr"])
	}
	if Defaults["LogToFile"] != false {
		t.Errorf("expected default log to file false, got %v", Defaults["LogToFile"])
	}
}
