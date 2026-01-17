package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// Config represents the main vulgar configuration
type Config struct {
	// Communication
	Google   GoogleConfig   `toml:"google"`
	Slack    SlackConfig    `toml:"slack"`
	Discord  DiscordConfig  `toml:"discord"`
	Telegram TelegramConfig `toml:"telegram"`

	// Development / Git Forges
	GitHub   GitHubConfig   `toml:"github"`
	GitLab   GitLabConfig   `toml:"gitlab"`
	Codeberg CodebergConfig `toml:"codeberg"`

	// AI Providers
	OpenAI      OpenAIConfig      `toml:"openai"`
	Anthropic   AnthropicConfig   `toml:"anthropic"`
	HuggingFace HuggingFaceConfig `toml:"huggingface"`

	// Payment & Services
	Stripe StripeConfig `toml:"stripe"`
	Twilio TwilioConfig `toml:"twilio"`

	// Productivity
	Notion   NotionConfig   `toml:"notion"`
	Airtable AirtableConfig `toml:"airtable"`

	// Cloud
	AWS AWSConfig `toml:"aws"`

	// Settings
	Defaults DefaultsConfig `toml:"defaults"`
}

// GoogleConfig holds Google OAuth 2.0 credentials
type GoogleConfig struct {
	ClientID     string `toml:"client_id"`
	ClientSecret string `toml:"client_secret"`
}

// SlackConfig holds Slack API credentials
type SlackConfig struct {
	Token          string `toml:"token"`
	DefaultChannel string `toml:"default_channel"`
}

// DiscordConfig holds Discord credentials
type DiscordConfig struct {
	WebhookURL string `toml:"webhook_url"`
	BotToken   string `toml:"bot_token"`
}

// TelegramConfig holds Telegram credentials
type TelegramConfig struct {
	BotToken      string `toml:"bot_token"`
	DefaultChatID string `toml:"default_chat_id"`
}

// GitHubConfig holds GitHub credentials
type GitHubConfig struct {
	Token        string `toml:"token"`
	DefaultOwner string `toml:"default_owner"`
}

// GitLabConfig holds GitLab credentials
type GitLabConfig struct {
	Token    string   `toml:"token"`
	URL      string   `toml:"url"`      // defaults to https://gitlab.com
	Projects []string `toml:"projects"` // default projects to track (e.g., "group/project")
}

// CodebergConfig holds Codeberg (Gitea) credentials
type CodebergConfig struct {
	Token string `toml:"token"`
	URL   string `toml:"url"` // defaults to https://codeberg.org
}

// OpenAIConfig holds OpenAI credentials
type OpenAIConfig struct {
	APIKey string `toml:"api_key"`
	Model  string `toml:"model"`
}

// AnthropicConfig holds Anthropic (Claude) credentials
type AnthropicConfig struct {
	APIKey string `toml:"api_key"`
	Model  string `toml:"model"`
}

// HuggingFaceConfig holds HuggingFace credentials
type HuggingFaceConfig struct {
	APIKey string `toml:"api_key"`
}

// StripeConfig holds Stripe credentials
type StripeConfig struct {
	APIKey string `toml:"api_key"`
}

// TwilioConfig holds Twilio credentials
type TwilioConfig struct {
	AccountSID string `toml:"account_sid"`
	AuthToken  string `toml:"auth_token"`
	FromNumber string `toml:"from_number"`
}

// NotionConfig holds Notion credentials
type NotionConfig struct {
	APIKey string `toml:"api_key"`
}

// AirtableConfig holds Airtable credentials
type AirtableConfig struct {
	APIKey string `toml:"api_key"`
	BaseID string `toml:"base_id"`
}

// AWSConfig holds AWS credentials (for S3, etc.)
type AWSConfig struct {
	AccessKeyID     string `toml:"access_key_id"`
	SecretAccessKey string `toml:"secret_access_key"`
	Region          string `toml:"region"`
}

// DefaultsConfig holds default settings
type DefaultsConfig struct {
	OutputFormat  string `toml:"output_format"` // table, json, yaml
	Color         bool   `toml:"color"`
	WorkflowsPath string `toml:"workflows_path"` // path to workflows directory
}

var (
	// Global config instance
	globalConfig *Config
	configPath   string
)

// ConfigDir returns the vulgar config directory path
func ConfigDir() string {
	// Check XDG_CONFIG_HOME first
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		return filepath.Join(xdgConfig, "vulgar")
	}

	// Fall back to ~/.config/vulgar
	home, err := os.UserHomeDir()
	if err != nil {
		return ".vulgar"
	}
	return filepath.Join(home, ".config", "vulgar")
}

// ConfigPath returns the full path to the config file
func ConfigPath() string {
	return filepath.Join(ConfigDir(), "config.toml")
}

// Load loads the configuration from disk
func Load() (*Config, error) {
	configPath = ConfigPath()

	// Start with defaults
	cfg := &Config{
		Defaults: DefaultsConfig{
			OutputFormat: "table",
			Color:        true,
		},
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// No config file, use defaults + env vars
		cfg.expandEnvVars()
		globalConfig = cfg
		return cfg, nil
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	// Parse TOML
	if err := toml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Expand environment variables
	cfg.expandEnvVars()

	globalConfig = cfg
	return cfg, nil
}

// Get returns the global config (loads if not already loaded)
func Get() *Config {
	if globalConfig == nil {
		cfg, err := Load()
		if err != nil {
			// Return empty config on error
			return &Config{
				Defaults: DefaultsConfig{
					OutputFormat: "table",
					Color:        true,
				},
			}
		}
		return cfg
	}
	return globalConfig
}

// Save saves the configuration to disk
func Save(cfg *Config) error {
	configDir := ConfigDir()

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to TOML
	data, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	configPath := ConfigPath()
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	globalConfig = cfg
	return nil
}

// Exists returns true if a config file exists
func Exists() bool {
	_, err := os.Stat(ConfigPath())
	return err == nil
}

// envVarRegex matches ${VAR_NAME} patterns
var envVarRegex = regexp.MustCompile(`\$\{([^}]+)\}`)

// expandEnvVars expands ${VAR} patterns in config values
// NOTE: This only expands explicit ${VAR} references in the TOML file.
// It does NOT automatically read from environment variables.
func (c *Config) expandEnvVars() {
	// Google OAuth
	c.Google.ClientID = expandEnv(c.Google.ClientID)
	c.Google.ClientSecret = expandEnv(c.Google.ClientSecret)

	// Slack
	c.Slack.Token = expandEnv(c.Slack.Token)

	// Discord
	c.Discord.WebhookURL = expandEnv(c.Discord.WebhookURL)
	c.Discord.BotToken = expandEnv(c.Discord.BotToken)

	// Telegram
	c.Telegram.BotToken = expandEnv(c.Telegram.BotToken)

	// GitHub
	c.GitHub.Token = expandEnv(c.GitHub.Token)

	// OpenAI
	c.OpenAI.APIKey = expandEnv(c.OpenAI.APIKey)

	// Anthropic
	c.Anthropic.APIKey = expandEnv(c.Anthropic.APIKey)

	// Stripe
	c.Stripe.APIKey = expandEnv(c.Stripe.APIKey)

	// Twilio
	c.Twilio.AccountSID = expandEnv(c.Twilio.AccountSID)
	c.Twilio.AuthToken = expandEnv(c.Twilio.AuthToken)

	// Notion
	c.Notion.APIKey = expandEnv(c.Notion.APIKey)

	// Airtable
	c.Airtable.APIKey = expandEnv(c.Airtable.APIKey)

	// AWS/S3
	c.AWS.AccessKeyID = expandEnv(c.AWS.AccessKeyID)
	c.AWS.SecretAccessKey = expandEnv(c.AWS.SecretAccessKey)

	// HuggingFace
	c.HuggingFace.APIKey = expandEnv(c.HuggingFace.APIKey)
}

// expandEnv expands ${VAR} patterns in a string
func expandEnv(s string) string {
	if s == "" {
		return s
	}

	return envVarRegex.ReplaceAllStringFunc(s, func(match string) string {
		// Extract variable name from ${VAR_NAME}
		varName := strings.TrimPrefix(strings.TrimSuffix(match, "}"), "${")
		if val := os.Getenv(varName); val != "" {
			return val
		}
		return match // Keep original if env var not set
	})
}

// DefaultConfigTemplate returns a template for the config file
func DefaultConfigTemplate() string {
	return `# Vulgar Configuration
# Location: ~/.config/vulgar/config.toml
#
# Use ${ENV_VAR} syntax to reference environment variables.
# Example: api_key = "${MY_SECRET_KEY}"

# ============================================================================
# GOOGLE APIS (Drive, Sheets, Calendar)
# ============================================================================
[google]
# OAuth 2.0 credentials (Desktop app type from Google Cloud Console)
# 1. Go to https://console.cloud.google.com/apis/credentials
# 2. Create OAuth 2.0 Client ID (Desktop app)
# 3. Copy client_id and client_secret here
# 4. Run 'vulgar gdrive login' to authenticate
client_id = ""
client_secret = ""

# ============================================================================
# COMMUNICATION
# ============================================================================
[slack]
token = ""
default_channel = ""

[discord]
webhook_url = ""
bot_token = ""

[telegram]
bot_token = ""
default_chat_id = ""

# ============================================================================
# DEVELOPMENT
# ============================================================================
[github]
token = ""
default_owner = ""

[gitlab]
token = ""
url = ""  # defaults to https://gitlab.com, set for self-hosted
projects = []  # default projects to track, e.g., ["group/project1", "group/project2"]

# ============================================================================
# AI PROVIDERS
# ============================================================================
[openai]
api_key = ""
model = "gpt-4"

[anthropic]
api_key = ""
model = "claude-3-opus-20240229"

[huggingface]
api_key = ""

# ============================================================================
# PAYMENT & SERVICES
# ============================================================================
[stripe]
api_key = ""

[twilio]
account_sid = ""
auth_token = ""
from_number = ""

# ============================================================================
# PRODUCTIVITY
# ============================================================================
[notion]
api_key = ""

[airtable]
api_key = ""
base_id = ""

# ============================================================================
# CLOUD (AWS/S3)
# ============================================================================
[aws]
access_key_id = ""
secret_access_key = ""
region = "us-east-1"

# ============================================================================
# DEFAULTS
# ============================================================================
[defaults]
output_format = "table"  # table, json, yaml
color = true
workflows_path = "workflows"  # path to workflows directory (relative to current dir or absolute)
`
}

// GetGoogleOAuthCredentials returns Google OAuth client ID and secret
func GetGoogleOAuthCredentials() (clientID, clientSecret string, ok bool) {
	cfg := Get()
	if cfg.Google.ClientID != "" && cfg.Google.ClientSecret != "" {
		return cfg.Google.ClientID, cfg.Google.ClientSecret, true
	}
	return "", "", false
}

// GoogleTokenPath returns the path to the Google OAuth token file
func GoogleTokenPath() string {
	return filepath.Join(ConfigDir(), "google_token.json")
}

// GetSlackToken returns the Slack token
func GetSlackToken() (string, bool) {
	cfg := Get()
	return cfg.Slack.Token, cfg.Slack.Token != ""
}

// GetDiscordWebhook returns the Discord webhook URL
func GetDiscordWebhook() (string, bool) {
	cfg := Get()
	return cfg.Discord.WebhookURL, cfg.Discord.WebhookURL != ""
}

// GetDiscordBotToken returns the Discord bot token
func GetDiscordBotToken() (string, bool) {
	cfg := Get()
	return cfg.Discord.BotToken, cfg.Discord.BotToken != ""
}

// GetTelegramBotToken returns the Telegram bot token
func GetTelegramBotToken() (string, bool) {
	cfg := Get()
	return cfg.Telegram.BotToken, cfg.Telegram.BotToken != ""
}

// GetGitHubToken returns the GitHub token
func GetGitHubToken() (string, bool) {
	cfg := Get()
	return cfg.GitHub.Token, cfg.GitHub.Token != ""
}

// GetGitLabToken returns the GitLab token and URL
func GetGitLabToken() (string, bool) {
	cfg := Get()
	return cfg.GitLab.Token, cfg.GitLab.Token != ""
}

// GetGitLabURL returns the GitLab URL (defaults to gitlab.com)
func GetGitLabURL() string {
	cfg := Get()
	if cfg.GitLab.URL != "" {
		return cfg.GitLab.URL
	}
	return "https://gitlab.com"
}

// GetGitLabProjects returns the configured GitLab projects
func GetGitLabProjects() []string {
	cfg := Get()
	return cfg.GitLab.Projects
}

// GetCodebergToken returns the Codeberg token
func GetCodebergToken() (string, bool) {
	cfg := Get()
	return cfg.Codeberg.Token, cfg.Codeberg.Token != ""
}

// GetCodebergURL returns the Codeberg URL (defaults to codeberg.org)
func GetCodebergURL() string {
	cfg := Get()
	if cfg.Codeberg.URL != "" {
		return cfg.Codeberg.URL
	}
	return "https://codeberg.org"
}

// GetOpenAIKey returns the OpenAI API key and model
func GetOpenAIKey() (key string, model string, ok bool) {
	cfg := Get()
	model = cfg.OpenAI.Model
	if model == "" {
		model = "gpt-4"
	}
	return cfg.OpenAI.APIKey, model, cfg.OpenAI.APIKey != ""
}

// GetAnthropicKey returns the Anthropic API key and model
func GetAnthropicKey() (key string, model string, ok bool) {
	cfg := Get()
	model = cfg.Anthropic.Model
	if model == "" {
		model = "claude-3-opus-20240229"
	}
	return cfg.Anthropic.APIKey, model, cfg.Anthropic.APIKey != ""
}

// GetHuggingFaceKey returns the HuggingFace API key
func GetHuggingFaceKey() (string, bool) {
	cfg := Get()
	return cfg.HuggingFace.APIKey, cfg.HuggingFace.APIKey != ""
}

// GetStripeKey returns the Stripe API key
func GetStripeKey() (string, bool) {
	cfg := Get()
	return cfg.Stripe.APIKey, cfg.Stripe.APIKey != ""
}

// GetTwilioCredentials returns Twilio credentials
func GetTwilioCredentials() (accountSID, authToken, fromNumber string, ok bool) {
	cfg := Get()
	if cfg.Twilio.AccountSID != "" && cfg.Twilio.AuthToken != "" {
		return cfg.Twilio.AccountSID, cfg.Twilio.AuthToken, cfg.Twilio.FromNumber, true
	}
	return "", "", "", false
}

// GetNotionKey returns the Notion API key
func GetNotionKey() (string, bool) {
	cfg := Get()
	return cfg.Notion.APIKey, cfg.Notion.APIKey != ""
}

// GetAirtableCredentials returns Airtable credentials
func GetAirtableCredentials() (apiKey, baseID string, ok bool) {
	cfg := Get()
	if cfg.Airtable.APIKey != "" {
		return cfg.Airtable.APIKey, cfg.Airtable.BaseID, true
	}
	return "", "", false
}

// GetAWSCredentials returns AWS credentials
func GetAWSCredentials() (accessKeyID, secretAccessKey, region string, ok bool) {
	cfg := Get()
	if cfg.AWS.AccessKeyID != "" && cfg.AWS.SecretAccessKey != "" {
		region := cfg.AWS.Region
		if region == "" {
			region = "us-east-1"
		}
		return cfg.AWS.AccessKeyID, cfg.AWS.SecretAccessKey, region, true
	}
	return "", "", "", false
}

// RequireConfig prints an error message if config is not set up
func RequireConfig(service string) error {
	if !Exists() {
		return fmt.Errorf("%s requires configuration.\n\nRun: vulgar init\nThen edit: %s", service, ConfigPath())
	}
	return nil
}

// GetWorkflowsPath returns the configured workflows path, or a default if not set
func GetWorkflowsPath() string {
	cfg := Get()
	if cfg.Defaults.WorkflowsPath != "" {
		return cfg.Defaults.WorkflowsPath
	}
	// Default to "workflows" directory in current working directory
	return "workflows"
}
