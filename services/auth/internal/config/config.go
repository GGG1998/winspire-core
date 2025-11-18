package config

import (
	"github.com/kelseyhightower/envconfig"
)

// Config holds all configuration for the application
type Config struct {
	// Supabase Configuration
	SupabaseURL          string `envconfig:"SUPABASE_URL" required:"true"`
	SupabaseAnonKey      string `envconfig:"SUPABASE_ANON_KEY" required:"true"`
	SupabaseServiceKey   string `envconfig:"SUPABASE_SERVICE_ROLE_KEY" required:"true"`
	SupabaseJWTSecret    string `envconfig:"SUPABASE_JWT_SECRET" required:"true"`

	// Database Configuration
	DatabaseURL string `envconfig:"DATABASE_URL" required:"true"`

	// Server Configuration
	Port string `envconfig:"PORT" default:"8080"`
	Env  string `envconfig:"ENV" default:"development"`

	// OAuth Providers
	DiscordClientID     string `envconfig:"DISCORD_CLIENT_ID"`
	DiscordClientSecret string `envconfig:"DISCORD_CLIENT_SECRET"`
	TwitchClientID      string `envconfig:"TWITCH_CLIENT_ID"`
	TwitchClientSecret  string `envconfig:"TWITCH_CLIENT_SECRET"`
	GoogleClientID      string `envconfig:"GOOGLE_CLIENT_ID"`
	GoogleClientSecret  string `envconfig:"GOOGLE_CLIENT_SECRET"`
	FacebookClientID    string `envconfig:"FACEBOOK_CLIENT_ID"`
	FacebookClientSecret string `envconfig:"FACEBOOK_CLIENT_SECRET"`
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

