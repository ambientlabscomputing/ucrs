package utils

import (
	"fmt"
	"log/slog"
	"os"
	"reflect"

	"gopkg.in/yaml.v3"
)

var Defaults = map[string]interface{}{
	"Port":        "8082",
	"Address":     "0.0.0.0",
	"BasePath":    "/api/v1/registry",
	"LogLevel":    "debug",
	"LogFormat":   "text",
	"LogToStderr": true,
	"LogToFile":   false,
	"LogFilePath": "./logs/capability_registry.log",
}

func LoadSettings() *Settings {
	slog.Info("Loading settings")

	settings := Settings{}

	// initialize settings with defaults
	val := reflect.ValueOf(&settings).Elem()
	for key, value := range Defaults {
		field := val.FieldByName(key)
		if field.IsValid() && field.CanSet() {
			field.Set(reflect.ValueOf(value))
		}
	}

	// read config.yaml, config file path set by CONFIG_PATH env var, default to "./config.yaml"
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "./config.yaml"
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		directory, dirErr := os.Getwd()
		if dirErr != nil {
			directory = "unknown"
		}
		dirContents, _ := os.ReadDir(directory)

		slog.Error(
			"failed to read config file",
			"error", err,
			"path", configPath,
			"directory", directory,
			"contents", dirContents,
		)
		panic(err)
	}

	// load yaml
	var fileSettings Settings
	err = yaml.Unmarshal(data, &fileSettings)
	if err != nil {
		panic(err)
	}

	// override defaults with file settings
	fileVal := reflect.ValueOf(&fileSettings).Elem()
	for i := 0; i < fileVal.NumField(); i++ {
		field := fileVal.Type().Field(i)
		fileFieldValue := fileVal.Field(i)

		if fileFieldValue.IsValid() && !fileFieldValue.IsZero() {
			settingsField := val.FieldByName(field.Name)
			if settingsField.IsValid() && settingsField.CanSet() {
				settingsField.Set(fileFieldValue)
			}
		}
	}

	// validate settings
	if err := settings.Validate(); err != nil {
		panic(err)
	}

	return &settings
}

func (s *Settings) Validate() error {
	if !s.LogToStderr && !s.LogToFile {
		return fmt.Errorf("at least one of LogToStderr or LogToFile must be true")
	}

	return nil
}

type Settings struct {
	Port        string `yaml:"port"`
	Address     string `yaml:"address"`
	BasePath    string `yaml:"base_path"`
	LogLevel    string `yaml:"log_level"`
	LogFormat   string `yaml:"log_format"`
	LogToStderr bool   `yaml:"log_to_stderr"`
	LogToFile   bool   `yaml:"log_to_file"`
	LogFilePath string `yaml:"log_file_path"`

	Auth     AuthConfig     `yaml:"auth"`
	Mongo    MongoConfig    `yaml:"mongo"`
	EventBus EventBusConfig `yaml:"event_bus"`
	Signing  SigningConfig  `yaml:"signing"`
	Snapshot SnapshotConfig `yaml:"snapshot"`
}

type AuthConfig struct {
	AuthDomain   string       `yaml:"auth_domain"`
	AuthAudience string       `yaml:"auth_audience"`
	AuthClientID string       `yaml:"auth_client_id"`
	Username     string       `yaml:"username"`
	Password     SecretString `yaml:"password"`
	Scopes       string       `yaml:"scopes"`
}

type MongoConfig struct {
	URI      string       `yaml:"uri"`
	Username string       `yaml:"username"`
	Password SecretString `yaml:"password"`
	Database string       `yaml:"database"`
}

type EventBusConfig struct {
	Enabled bool   `yaml:"enabled"`
	URL     string `yaml:"url"`
}

type SigningConfig struct {
	PrivateKeyPath string `yaml:"private_key_path"`
	PublicKeyPath  string `yaml:"public_key_path"`
	KeyID          string `yaml:"key_id"`
}

type SnapshotConfig struct {
	RegenerateOnStart bool `yaml:"regenerate_on_start"`
}

type SecretString string

func (s SecretString) String() string {
	if s == "" {
		return ""
	}
	return "***REDACTED***"
}

func (s SecretString) MarshalYAML() (interface{}, error) {
	return s.String(), nil
}

func (s SecretString) Value() string {
	return string(s)
}

func (s SecretString) Empty() bool {
	return string(s) == ""
}
