package secrets

type Token struct {
	SecretKey          string `mapstructure:"secret_key"`
	TokenRefreshEndTTL int64  `mapstructure:"token_refresh_end_ttl"`
	TokenAccessEndTTL  int64  `mapstructure:"token_access_end_ttl"`
}

type Redis struct {
	Host     string `mapstructure:"host"`
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type Hasher struct {
	Time    uint32 `mapstructure:"time"`
	Memory  uint32 `mapstructure:"memory"`
	Threads uint8  `mapstructure:"threads"`
	KeyLen  uint32 `mapstructure:"key_len"`
	SaltLen uint32 `mapstructure:"salt_len"`
}

type Secrets struct {
	Token  Token
	Redis  Redis
	Hasher Hasher
}
