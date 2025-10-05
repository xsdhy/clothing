package config

import (
	"github.com/caarlos0/env/v10"
	"github.com/sirupsen/logrus"
)

type Config struct {
	HTTPPort string `env:"HTTP_PORT" envDefault:"8080"`

	DBType     string `env:"DBType" envDefault:"sqlite"`
	DSNURL     string `env:"DSN_URL" envDefault:""`
	DBUser     string `env:"DBUser" envDefault:""`
	DBPassword string `env:"DBPassword" envDefault:""`
	DBAddr     string `env:"DBAddr" envDefault:""`
	DBName     string `env:"DBName" envDefault:"monitor"`
	DBPath     string `env:"DBPath" envDefault:""`
	DBPort     string `env:"DBPort" envDefault:"3306"`

	StorageType          string `env:"STORAGE_TYPE" envDefault:"local"`
	StorageLocalDir      string `env:"STORAGE_LOCAL_DIR" envDefault:"data/images"`
	StoragePublicBaseURL string `env:"STORAGE_PUBLIC_BASE_URL" envDefault:"/files"`

	OpenRouterAPIKey string `env:"OPENROUTER_API_KEY" envDefault:""`
	GeminiAPIKey     string `env:"GEMINI_API_KEY" envDefault:""`
	AiHubMixAPIKey   string `env:"AIHUBMIX_API_KEY" envDefault:""`
	DashscopeAPIKey  string `env:"DASHSCOPE_API_KEY" envDefault:""`
	VolcengineAPIKey string `env:"VOLCENGINE_API_KEY" envDefault:""`
	FalAPIKey        string `env:"FAL_KEY" envDefault:""`
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
