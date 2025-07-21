package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Server   configServer   `mapstructure:"server" json:"server"`
	Database configDatabase `mapstructure:"database" json:"database"`
	Logger   configLogger   `mapstructure:"logger" json:"logger"`
}

type configServer struct {
	Port string `mapstructure:"port" json:"port"`
	Host string `mapstructure:"host" json:"host"`
}

type configDatabase struct {
	Driver   string `mapstructure:"driver" json:"driver"`
	Host     string `mapstructure:"host" json:"host"`
	Port     string `mapstructure:"port" json:"port"`
	Username string `mapstructure:"username" json:"username"`
	Password string `mapstructure:"password" json:"password"`
	Name     string `mapstructure:"name" json:"name"`
}

type configLogger struct {
	Level string `mapstructure:"level" json:"level"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("Config file not found, using defaults or environment variables.")
		} else {
			return nil, fmt.Errorf("error reading config: %w", err)
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode config: %w", err)
	}

	return &cfg, nil
}
