package config

import (
	"fmt"
	"os"
	"regexp"

	"gh-mirror/pkg/models"
	"gh-mirror/pkg/platform"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Platforms    map[string]PlatformConfig `yaml:"platforms"`
	Source       string                   `yaml:"source"`
	Destinations []string                 `yaml:"destinations"`
	Sync         SyncConfig               `yaml:"sync"`
	Cron         CronConfig               `yaml:"cron"`
}

type PlatformConfig struct {
	Token   string `yaml:"token"`
	BaseURL string `yaml:"base_url"`
}

type SyncConfig struct {
	TimeoutMinutes int `yaml:"timeout_minutes"`
}

type CronConfig struct {
	Enabled       bool `yaml:"enabled"`
	IntervalHours int  `yaml:"interval_hours"`
}

var envVarPattern = regexp.MustCompile(`\$\{([^}]+)\}`)

func expandEnvValue(value string) string {
	return envVarPattern.ReplaceAllStringFunc(value, func(match string) string {
		envName := match[2 : len(match)-1]
		return os.Getenv(envName)
	})
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	expanded := expandEnvValue(string(data))

	var cfg Config
	if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		return nil, err
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	if c.Source == "" {
		return &ConfigError{Field: "source", Message: "required"}
	}

	sourceID := models.PlatformID(c.Source)
	if _, err := platform.Create(sourceID); err != nil {
		return &ConfigError{Field: "source", Message: fmt.Sprintf("unsupported platform: %s", c.Source)}
	}

	if c.Platforms == nil {
		c.Platforms = make(map[string]PlatformConfig)
	}

	if _, hasSourceConfig := c.Platforms[c.Source]; !hasSourceConfig {
		return &ConfigError{Field: fmt.Sprintf("platforms.%s", c.Source), Message: "platform configuration required"}
	}

	for _, dest := range c.Destinations {
		destID := models.PlatformID(dest)
		if _, err := platform.Create(destID); err != nil {
			return &ConfigError{Field: "destinations", Message: fmt.Sprintf("unsupported platform: %s", dest)}
		}
		if _, hasDestConfig := c.Platforms[dest]; !hasDestConfig {
			return &ConfigError{Field: fmt.Sprintf("platforms.%s", dest), Message: "platform configuration required"}
		}
	}

	for _, dest := range c.Destinations {
		if dest == c.Source {
			return &ConfigError{Field: "destinations", Message: fmt.Sprintf("destination cannot be same as source: %s", c.Source)}
		}
	}

	if c.Sync.TimeoutMinutes == 0 {
		c.Sync.TimeoutMinutes = 30
	}

	return nil
}

type ConfigError struct {
	Field   string
	Message string
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf("config: %s %s", e.Field, e.Message)
}
