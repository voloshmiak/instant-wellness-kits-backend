package config

import (
	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
)

type Config struct {
	Port            string `env:"PORT" envDefault:"8080"`
	GeocodingAPIKey string `env:"GEOCODING_API_KEY,required"`
	Database        struct {
		Name           string `env:"DB_NAME,required"`
		Password       string `env:"DB_PASSWORD"`
		User           string `env:"DB_USER,required"`
		UserPassword   string `env:"DB_USER_PASSWORD,required"`
		ConnectionName string `env:"INSTANCE_CONNECTION_NAME"`
		Host           string `env:"DB_HOST"`
		Port           string `env:"DB_PORT"`
	}
	Env string `env:"ENV" envDefault:"DEV"`
}

func New() (*Config, error) {
	_ = godotenv.Load(".env")

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
