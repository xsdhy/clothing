package config

import (
	"github.com/caarlos0/env/v10"
	"github.com/sirupsen/logrus"
)

type Config struct {
	HTTPPort string `env:"HTTP_PORT" envDefault:"8080"`

	OpenRouterAPIKey  string `env:"OPENROUTER_API_KEY" envDefault:"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"`
	OpenRouterReferer string `env:"OPENROUTER_REFERER" envDefault:""`
	OpenRouterTitle   string `env:"OPENROUTER_TITLE" envDefault:""`
	GeminiAPIKey      string `env:"GEMINI_API_KEY" envDefault:""`
}

func ParseConfig() (Config, error) {
	var Conf Config
	err := env.Parse(&Conf)
	if err != nil {
		logrus.WithError(err).Error("env.Parse error")
	}
	logrus.Debugf("%#v\n", Conf)
	return Conf, nil
}
