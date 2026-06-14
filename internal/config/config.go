package config

import (
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
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

type Settings struct {
	TokenRefreshEndTTL time.Duration `env:"TOKEN_REFRESH_END_TTL" env-required:"true"`
	TokenAccessEndTTL  time.Duration `env:"TOKEN_ACCESS_END_TTL" env-required:"true"`
	SecretKey          string        `env:"SECRET_KEY"`
	DefaultTimeout     time.Duration `yaml:"default_timeout"`
}

type KeyBuilder struct {
	Prev    string `yaml:"prev"`
	Version string `yaml:"version"`
}

type Redis struct {
	Host           string        `env:"REDIS_HOST"`
	Addr           string        `env:"REDIS_ADDR"`
	Password       string        `env:"REDIS_PASSWORD"`
	DB             int           `env:"REDIS_DB"`
	DefaultTimeout time.Duration `yaml:"default_timeout"`
}

type Hasher struct {
	Time    uint32 `env:"HASH_TIME" env-required:"true"`
	Memory  uint32 `env:"HASH_MEMORY" env-required:"true"`
	Threads uint8  `env:"HASH_THREADS" env-required:"true"`
	KeyLen  uint32 `env:"HASH_KEY_LEN" env-required:"true"`
	SaltLen uint32 `env:"HASH_SALT_LEN" env-required:"true"`
}

type Config struct {
	HTTP       HTTP       `yaml:"http"`
	GRPC       GRPC       `yaml:"grpc"`
	Redis      Redis      `yaml:"db"`
	Settings   Settings   `yaml:"settings"`
	KeyBuilder KeyBuilder `yaml:"key_builder"`
	Hasher     Hasher     `yaml:"hasher"`
}

func New(path string) (*Config, error) {
	cfg := &Config{}

	if err := cleanenv.ReadConfig(path, cfg); err != nil {
		return nil, fmt.Errorf("reading config file %s: %w", path, err)
	}

	return cfg, nil
}

func LoadEnv(filenames ...string) error {
	if len(filenames) == 0 {
		return godotenv.Load()
	}

	for _, filename := range filenames {
		if err := godotenv.Load(filename); err != nil {
			return fmt.Errorf("loading env file %s: %w", filename, err)
		}
	}
	return nil
}
