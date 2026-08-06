package config

import (
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type GRPC struct {
	Port        string `yaml:"port"`
	UserApiHost string `yaml:"user_api_host"`
	UserApiPort string `yaml:"user_api_port"`
}

type HTTP struct {
	Port              string        `yaml:"port" env-default:"9080"`
	MaxHeaderBytes    int           `yaml:"max_header_bytes" env-default:"4096"`
	ReadHeaderTimeout time.Duration `yaml:"read_header_timeout" env-default:"10s"`
	WriteTimeout      time.Duration `yaml:"write_timeout" env-default:"10s"`
	BaseHTTPPath      string        `yaml:"base_http_path" env-default:"/api"`
}

type Middleware struct {
	AllowedOrigins   []string `yaml:"allowed_origins"`
	AllowedMethods   []string `yaml:"allowed_methods"`
	AllowedHeaders   []string `yaml:"allowed_headers"`
	ExposedHeaders   []string `yaml:"exposed_headers"`
	AllowCredentials bool     `yaml:"allow_credentials"`
	MaxAge           int      `yaml:"max_age"`
}

type Settings struct {
	DefaultTimeout time.Duration `yaml:"default_timeout"`
}

type KeyBuilder struct {
	Prev    string `yaml:"prev"`
	Version string `yaml:"version"`
}

type Redis struct {
	DefaultTimeout time.Duration `yaml:"default_timeout"`
}

type Vault struct {
	SecretToken string `env:"VAULT_SECRET_TOKEN" env-required:"true"`
	Address     string `env:"VAULT_ADDRESS" env-required:"true"`
	Mount       string `env:"VAULT_MOUNT" env-required:"true"`
}

type VaultPaths struct {
	Token  string `env:"TOKEN_PATH" env-required:"true"`
	Redis  string `env:"REDIS_PATH" env-required:"true"`
	Hasher string `env:"HASHER_PATH" env-required:"true"`
}
type Config struct {
	HTTP       HTTP       `yaml:"http"`
	Middleware Middleware `yaml:"middleware"`
	GRPC       GRPC       `yaml:"grpc"`
	Redis      Redis      `yaml:"db"`
	Settings   Settings   `yaml:"settings"`
	KeyBuilder KeyBuilder `yaml:"key_builder"`
	Vault      Vault
	VaultPaths VaultPaths
}

func New(path string) (*Config, error) {
	cfg := &Config{}

	if err := cleanenv.ReadConfig(path, cfg); err != nil {
		return nil, fmt.Errorf("reading config file %s: %w", path, err)
	}

	if err := cleanenv.ReadEnv(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
