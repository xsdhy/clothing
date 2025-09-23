package config

import (
	"github.com/caarlos0/env/v10"
	"github.com/sirupsen/logrus"
)

type Config struct {
	HTTPPort string `env:"HTTP_PORT" envDefault:"8080"`

	OpenRouterAPIKey string `env:"OPENROUTER_API_KEY" envDefault:""`
	GeminiAPIKey     string `env:"GEMINI_API_KEY" envDefault:""`
	AiHubMixAPIKey   string `env:"AIHUBMIX_API_KEY" envDefault:""`
	DashscopeAPIKey  string `env:"DASHSCOPE_API_KEY" envDefault:""`
}

func ParseConfig() (Config, error) {
	var Conf Config
	err := env.Parse(&Conf)
	if err != nil {
		logrus.WithError(err).Error("env.Parse error")
		return Config{}, err
	}
	logrus.Debugf("%#v\n", Conf)
	return Conf, nil
}
