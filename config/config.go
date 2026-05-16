package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Upload   UploadConfig   `mapstructure:"upload"`
}

type ServerConfig struct {
	Port int `mapstructure:"port"`
}

type DatabaseConfig struct {
	DSN string `mapstructure:"dsn"`
}

type JWTConfig struct {
	Secret string `mapstructure:"secret"`
}

type UploadConfig struct {
	Path string `mapstructure:"path"`
}

var AppConfig *Config

func Load() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	viper.SetDefault("server.port", 8080)
	viper.SetDefault("upload.path", "./uploads")
	viper.SetDefault("jwt.secret", "change-this-secret-in-production")

	viper.AutomaticEnv()
	viper.BindEnv("database.dsn", "IM_DB_DSN")
	viper.BindEnv("jwt.secret", "IM_JWT_SECRET")
	viper.BindEnv("server.port", "IM_PORT")
	viper.BindEnv("upload.path", "IM_UPLOAD_PATH")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("读取配置文件失败: %w", err)
		}
	}

	AppConfig = &Config{}
	if err := viper.Unmarshal(AppConfig); err != nil {
		return fmt.Errorf("解析配置失败: %w", err)
	}

	if err := os.MkdirAll(AppConfig.Upload.Path, 0755); err != nil {
		return fmt.Errorf("创建上传目录失败: %w", err)
	}

	return nil
}
