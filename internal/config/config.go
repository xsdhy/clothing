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
	DBPath     string `env:"DBPath" envDefault:"datas/clothing.db"`
	DBPort     string `env:"DBPort" envDefault:"3306"`

	StorageType          string `env:"STORAGE_TYPE" envDefault:"local"`
	StorageLocalDir      string `env:"STORAGE_LOCAL_DIR" envDefault:"datas/images"`
	StoragePublicBaseURL string `env:"STORAGE_PUBLIC_BASE_URL" envDefault:"/files"`

	// S3-compatible storage configuration
	StorageS3Region          string `env:"STORAGE_S3_REGION"`
	StorageS3Bucket          string `env:"STORAGE_S3_BUCKET"`
	StorageS3Prefix          string `env:"STORAGE_S3_PREFIX"`
	StorageS3Endpoint        string `env:"STORAGE_S3_ENDPOINT"`
	StorageS3AccessKeyID     string `env:"STORAGE_S3_ACCESS_KEY_ID"`
	StorageS3SecretAccessKey string `env:"STORAGE_S3_SECRET_ACCESS_KEY"`
	StorageS3SessionToken    string `env:"STORAGE_S3_SESSION_TOKEN"`
	StorageS3ForcePathStyle  bool   `env:"STORAGE_S3_FORCE_PATH_STYLE" envDefault:"false"`

	// Alibaba Cloud OSS configuration
	StorageOSSEndpoint        string `env:"STORAGE_OSS_ENDPOINT"`
	StorageOSSBucket          string `env:"STORAGE_OSS_BUCKET"`
	StorageOSSPrefix          string `env:"STORAGE_OSS_PREFIX"`
	StorageOSSAccessKeyID     string `env:"STORAGE_OSS_ACCESS_KEY_ID"`
	StorageOSSAccessKeySecret string `env:"STORAGE_OSS_ACCESS_KEY_SECRET"`

	// Tencent Cloud COS configuration
	StorageCOSBucketURL string `env:"STORAGE_COS_BUCKET_URL"`
	StorageCOSPrefix    string `env:"STORAGE_COS_PREFIX"`
	StorageCOSSecretID  string `env:"STORAGE_COS_SECRET_ID"`
	StorageCOSSecretKey string `env:"STORAGE_COS_SECRET_KEY"`

	// Cloudflare R2 configuration
	StorageR2AccountID       string `env:"STORAGE_R2_ACCOUNT_ID"`
	StorageR2Endpoint        string `env:"STORAGE_R2_ENDPOINT"`
	StorageR2Region          string `env:"STORAGE_R2_REGION" envDefault:"auto"`
	StorageR2Bucket          string `env:"STORAGE_R2_BUCKET"`
	StorageR2Prefix          string `env:"STORAGE_R2_PREFIX"`
	StorageR2AccessKeyID     string `env:"STORAGE_R2_ACCESS_KEY_ID"`
	StorageR2SecretAccessKey string `env:"STORAGE_R2_SECRET_ACCESS_KEY"`

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
