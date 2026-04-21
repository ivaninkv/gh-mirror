package config

type LoadTestCase struct {
	Name        string
	YAMLContent string
	EnvVars     map[string]string
	WantSource  string
}

type ExpandEnvValueTestCase struct {
	Name     string
	Input    string
	EnvVars  map[string]string
	Want     string
}

type ConfigErrorTestCase struct {
	Name    string
	Error   *ConfigError
	WantMsg string
}

func LoadValidCases() []LoadTestCase {
	return []LoadTestCase{
		{
			Name: "valid config with all fields",
			YAMLContent: `
platforms:
  github:
    token: "ghp_test"
    url: "https://github.com"
  gitlab:
    token: "glpat_test"
    api_url: "https://gitlab.com/api/v4"
    url: "https://gitlab.com"
source: github
destinations:
  - gitlab
sync:
  timeout_minutes: 60
`,
			WantSource: "github",
		},
		{
			Name: "valid config with multiple destinations",
			YAMLContent: `
platforms:
  github:
    token: "ghp_test"
    url: "https://github.com"
  gitlab:
    token: "glpat_test"
    api_url: "https://gitlab.com/api/v4"
    url: "https://gitlab.com"
  gitverse:
    token: "gv_test"
    api_url: "https://api.gitverse.ru"
    url: "https://gitverse.ru"
source: github
destinations:
  - gitlab
  - gitverse
`,
			WantSource: "github",
		},
	}
}

func ExpandEnvValueTestCases() []ExpandEnvValueTestCase {
	return []ExpandEnvValueTestCase{
		{
			Name:    "no env vars",
			Input:   "plain value",
			EnvVars: map[string]string{},
			Want:    "plain value",
		},
		{
			Name:    "single env var",
			Input:   "token: ${MY_TOKEN}",
			EnvVars: map[string]string{"MY_TOKEN": "secret123"},
			Want:    "token: secret123",
		},
		{
			Name:    "multiple env vars",
			Input:   "url: https://${HOST}:${PORT}",
			EnvVars: map[string]string{"HOST": "example.com", "PORT": "443"},
			Want:    "url: https://example.com:443",
		},
		{
			Name:    "env var not set",
			Input:   "token: ${UNDEFINED_VAR}",
			EnvVars: map[string]string{},
			Want:    "token: ",
		},
		{
			Name:    "mixed content",
			Input:   "api_url: https://${API_HOST}/v1?token=${API_TOKEN}",
			EnvVars: map[string]string{"API_HOST": "api.test.com", "API_TOKEN": "abc"},
			Want:    "api_url: https://api.test.com/v1?token=abc",
		},
	}
}

func ConfigErrorTestCases() []ConfigErrorTestCase {
	return []ConfigErrorTestCase{
		{
			Name:    "source required error",
			Error:   &ConfigError{Field: "source", Message: "required"},
			WantMsg: "config: source required",
		},
		{
			Name:    "unsupported platform error",
			Error:   &ConfigError{Field: "source", Message: "unsupported platform: github"},
			WantMsg: "config: source unsupported platform: github",
		},
		{
			Name:    "missing platform config error",
			Error:   &ConfigError{Field: "platforms.github", Message: "platform configuration required"},
			WantMsg: "config: platforms.github platform configuration required",
		},
		{
			Name:    "destination same as source error",
			Error:   &ConfigError{Field: "destinations", Message: "destination cannot be same as source: github"},
			WantMsg: "config: destinations destination cannot be same as source: github",
		},
	}
}

