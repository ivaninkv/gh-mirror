package config

import (
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

type Config struct {
	GitHub   GitHubConfig   `yaml:"github"`
	GitVerse GitVerseConfig `yaml:"gitverse"`
	Sync     SyncConfig     `yaml:"sync"`
	Cron     CronConfig     `yaml:"cron"`
}

type GitHubConfig struct {
	Token string `yaml:"token"`
}

type GitVerseConfig struct {
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
	if c.GitHub.Token == "" {
		return &ConfigError{Field: "github.token", Message: "required"}
	}
	if c.GitVerse.Token == "" {
		return &ConfigError{Field: "gitverse.token", Message: "required"}
	}
	if c.GitVerse.BaseURL == "" {
		c.GitVerse.BaseURL = "https://gitverse.ru/api/v1"
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
	return "config: " + e.Field + " " + e.Message
}
