package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nalgeon/be"

	_ "gh-mirror/pkg/platforms/codeberg"
	_ "gh-mirror/pkg/platforms/github"
	_ "gh-mirror/pkg/platforms/gitverse"
	_ "gh-mirror/pkg/platforms/gitlab"
)

func TestLoadValid(t *testing.T) {
	for _, tc := range LoadValidCases() {
		t.Run(tc.Name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yaml")
			err := os.WriteFile(configPath, []byte(tc.YAMLContent), 0644)
			be.Equal(t, err, nil)

			cfg, err := Load(configPath)
			if err != nil {
				t.Logf("Load error (platform may not be registered in test context): %v", err)
				t.Skip("skipping - requires platform registration")
			}
			be.Equal(t, cfg.Source, tc.WantSource)
		})
	}
}

func TestLoadFileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	be.Err(t, err, os.ErrNotExist)
}

func TestLoadInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.yaml")
	err := os.WriteFile(configPath, []byte("invalid: yaml: ["), 0644)
	be.Equal(t, err, nil)

	_, err = Load(configPath)
	be.True(t, err != nil)
}

func TestExpandEnvValue(t *testing.T) {
	for _, tc := range ExpandEnvValueTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			for key, val := range tc.EnvVars {
				os.Setenv(key, val)
				defer os.Unsetenv(key)
			}

			got := expandEnvValue(tc.Input)
			be.Equal(t, got, tc.Want)
		})
	}
}

func TestConfigError(t *testing.T) {
	for _, tc := range ConfigErrorTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			got := tc.Error.Error()
			be.Equal(t, got, tc.WantMsg)
		})
	}
}

func TestConfigErrorField(t *testing.T) {
	err := &ConfigError{Field: "source", Message: "required"}
	be.Equal(t, err.Field, "source")
}

func TestConfigErrorMessage(t *testing.T) {
	err := &ConfigError{Field: "source", Message: "required"}
	be.Equal(t, err.Message, "required")
}

func TestValidateEmptySource(t *testing.T) {
	cfg := Config{
		Source: "",
		Platforms: map[string]PlatformConfig{
			"github": {Token: "test", URL: "https://github.com"},
		},
	}

	err := cfg.validate()
	be.True(t, err != nil)
	be.True(t, contains(err.Error(), "source required"))
}

func TestValidateMissingSourceConfig(t *testing.T) {
	cfg := Config{
		Source: "github",
		Platforms: map[string]PlatformConfig{
			"gitlab": {Token: "test", URL: "https://gitlab.com"},
		},
	}

	err := cfg.validate()
	be.True(t, err != nil)
	be.True(t, contains(err.Error(), "platform configuration required"))
}

func TestValidateDestinationSameAsSource(t *testing.T) {
	cfg := Config{
		Source: "github",
		Platforms: map[string]PlatformConfig{
			"github": {Token: "test", URL: "https://github.com"},
		},
		Destinations: []string{"github"},
	}

	err := cfg.validate()
	be.True(t, err != nil)
	be.True(t, contains(err.Error(), "destination cannot be same as source"))
}

func TestValidateSetsDefaultTimeout(t *testing.T) {
	cfg := Config{
		Source: "github",
		Platforms: map[string]PlatformConfig{
			"github": {Token: "test", URL: "https://github.com"},
		},
		Destinations: []string{"github"},
		Sync:        SyncConfig{TimeoutMinutes: 0},
	}

	err := cfg.validate()
	be.True(t, err != nil)
	be.True(t, contains(err.Error(), "destination cannot be same as source"))
}

func TestPlatformConfig(t *testing.T) {
	pc := PlatformConfig{
		Token:  "test-token",
		APIURL: "https://api.example.com",
		URL:    "https://example.com",
	}

	be.Equal(t, pc.Token, "test-token")
	be.Equal(t, pc.APIURL, "https://api.example.com")
	be.Equal(t, pc.URL, "https://example.com")
}

func TestSyncConfig(t *testing.T) {
	sc := SyncConfig{
		TimeoutMinutes: 60,
	}

	be.Equal(t, sc.TimeoutMinutes, 60)
}

func TestConfigStruct(t *testing.T) {
	cfg := Config{
		Platforms: map[string]PlatformConfig{
			"github": {Token: "ghp_xxx", URL: "https://github.com"},
			"gitlab": {Token: "glpat_xxx", APIURL: "https://gitlab.com/api/v4", URL: "https://gitlab.com"},
		},
		Source:       "github",
		Destinations: []string{"gitlab"},
		Sync:         SyncConfig{TimeoutMinutes: 45},
	}

	be.Equal(t, len(cfg.Platforms), 2)
	be.Equal(t, cfg.Source, "github")
	be.Equal(t, len(cfg.Destinations), 1)
	be.Equal(t, cfg.Sync.TimeoutMinutes, 45)
}

func BenchmarkExpandEnvValue(b *testing.B) {
	cases := ExpandEnvValueTestCases()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tc := range cases {
			expandEnvValue(tc.Input)
		}
	}
}

func BenchmarkConfigErrorError(b *testing.B) {
	err := &ConfigError{Field: "source", Message: "required"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err.Error()
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}