package config

import (
	"fmt"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	OSS      OSSConfig      `mapstructure:"oss"`
	Auth     AuthConfig     `mapstructure:"auth"`
}

type ServerConfig struct {
	GRPCAddr string `mapstructure:"grpc_addr"`
	HTTPAddr string `mapstructure:"http_addr"`
}

type DatabaseConfig struct {
	DSN              string `mapstructure:"dsn"`
	MigrationsSource string `mapstructure:"migrations_source"`
}

type OSSConfig struct {
	Endpoint        string `mapstructure:"endpoint"`
	AccessKey       string `mapstructure:"access_key"`
	SecretKey       string `mapstructure:"secret_key"`
	Bucket          string `mapstructure:"bucket"`
	UseSSL          bool   `mapstructure:"use_ssl"`
	MaxUploadSizeKB int    `mapstructure:"max_upload_size_kb"`
}

type AuthConfig struct {
	JWTSecret  string        `mapstructure:"jwt_secret"`
	JWTIssuer  string        `mapstructure:"jwt_issuer"`
	JWTTTL     time.Duration `mapstructure:"jwt_ttl"`
	RefreshTTL time.Duration `mapstructure:"refresh_ttl"`
}

func Load(path string) (*Config, error) {
	if path == "" {
		return nil, fmt.Errorf("config path is required")
	}

	v := viper.New()
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg, viper.DecodeHook(mapstructure.StringToTimeDurationHookFunc())); err != nil {
		return nil, fmt.Errorf("decode config: %w", err)
	}
	return &cfg, nil
}
