package config

import (
	"time"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	Server ServerConfig
	Client ClientConfig
	Logger LoggerConfig
	PoW    PoWConfig
}

type ServerConfig struct {
	Address         string        `env:"SERVER_ADDRESS" envDefault:":8080"`
	ReadTimeout     time.Duration `env:"SERVER_READ_TIMEOUT" envDefault:"5s"`
	WriteTimeout    time.Duration `env:"SERVER_WRITE_TIMEOUT" envDefault:"5s"`
	ShutdownTimeout time.Duration `env:"SERVER_SHUTDOWN_TIMEOUT" envDefault:"10s"`
	MaxConnections  int           `env:"SERVER_MAX_CONNECTIONS" envDefault:"1000"`
}

type ClientConfig struct {
	ServerAddress  string        `env:"CLIENT_SERVER_ADDRESS" envDefault:"localhost:8080"`
	ConnectTimeout time.Duration `env:"CLIENT_CONNECT_TIMEOUT" envDefault:"5s"`
	ReadTimeout    time.Duration `env:"CLIENT_READ_TIMEOUT" envDefault:"5s"`
	WriteTimeout   time.Duration `env:"CLIENT_WRITE_TIMEOUT" envDefault:"5s"`
	MaxAttempts    int           `env:"CLIENT_MAX_ATTEMPTS" envDefault:"3"`
	RetryDelay     time.Duration `env:"CLIENT_RETRY_DELAY" envDefault:"1s"`
	MaxRetryDelay  time.Duration `env:"CLIENT_MAX_RETRY_DELAY" envDefault:"30s"`
}

type LoggerConfig struct {
	Level  string `env:"LOG_LEVEL" envDefault:"info"`
	Pretty bool   `env:"LOG_PRETTY" envDefault:"true"`
	JSON   bool   `env:"LOG_JSON" envDefault:"false"`
}

type PoWConfig struct {
	InitialDifficulty uint8         `env:"POW_INITIAL_DIFFICULTY" envDefault:"4"`
	MaxDifficulty     uint8         `env:"POW_MAX_DIFFICULTY" envDefault:"8"`
	ChallengeTTL      time.Duration `env:"POW_CHALLENGE_TTL" envDefault:"5m"`
	AdjustInterval    time.Duration `env:"POW_ADJUST_INTERVAL" envDefault:"10s"`
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Address:         ":8080",
			ReadTimeout:     5 * time.Second,
			WriteTimeout:    5 * time.Second,
			ShutdownTimeout: 10 * time.Second,
			MaxConnections:  1000,
		},
		Client: ClientConfig{
			ServerAddress:  "localhost:8080",
			ConnectTimeout: 5 * time.Second,
			ReadTimeout:    5 * time.Second,
			WriteTimeout:   5 * time.Second,
			MaxAttempts:    3,
			RetryDelay:     time.Second,
			MaxRetryDelay:  30 * time.Second,
		},
		Logger: LoggerConfig{
			Level:  "info",
			Pretty: true,
			JSON:   false,
		},
		PoW: PoWConfig{
			InitialDifficulty: 4,
			MaxDifficulty:     8,
			ChallengeTTL:      5 * time.Minute,
			AdjustInterval:    10 * time.Second,
		},
	}
}
