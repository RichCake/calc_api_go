package config

import (
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type TimeConfig struct {
	TimeAdd time.Duration `env:"TIME_ADDITION_MS" env-default:"1s"`
	TimeSub time.Duration `env:"TIME_SUBTRACTION_MS" env-default:"1s"`
	TimeMul time.Duration `env:"TIME_MULTIPLICATIONS_MS" env-default:"1s"`
	TimeDiv time.Duration `env:"TIME_DIVISIONS_MS" env-default:"1s"`
}

type AuthConfig struct {
	TokenTTL time.Duration `env:"AUTH_TOKEN_TTL" env-default:"1h"`
}

type Config struct {
	Addr      string `env:"ORCHESTRATOR_PORT" env-default:"8080"`
	GRPCPort  string `env:"TASKS_PORT" env-default:"50051"`
	SecretKey string `env:"SECRET_KEY"`
	TimeConf  TimeConfig
	AuthCon   AuthConfig
}

func ConfigFromEnv() (*Config, error) {
	var cfg Config
	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
