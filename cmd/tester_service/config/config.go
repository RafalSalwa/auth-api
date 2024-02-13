package config

import (
	"errors"
	"fmt"
	"github.com/RafalSalwa/interview-app-srv/pkg/http/auth"
	"github.com/RafalSalwa/interview-app-srv/pkg/logger"
	"github.com/spf13/viper"
	"os"
)

var configPath string

type Config struct {
	ServiceName string         `mapstructure:"serviceName"`
	App         App            `mapstructure:"app"`
	Logger      *logger.Config `mapstructure:"logger"`
	HTTP        HTTP           `mapstructure:"http"`
	Auth        auth.Auth      `mapstructure:"auth"`
}

type App struct {
	Env   string `mapstructure:"env"`
	Debug bool   `mapstructure:"debug"`
}

type HTTP struct {
	Addr                string `mapstructure:"addr"`
	Development         bool   `mapstructure:"development"`
	BasePath            string `mapstructure:"basePath"`
	DebugHeaders        bool   `mapstructure:"debugHeaders"`
	HTTPClientDebug     bool   `mapstructure:"httpClientDebug"`
	DebugErrorsResponse bool   `mapstructure:"debugErrorsResponse"`
}

func InitConfig() (*Config, error) {
	cfg := &Config{}
	path, err := getEnvPath()
	if err != nil {
		return nil, err
	}
	viper.SetConfigType("yaml")
	viper.SetConfigFile(path)

	if err := viper.ReadInConfig(); err != nil {
		return nil, errors.New("viper.ReadInConfig")
	}

	if err := viper.Unmarshal(cfg); err != nil {
		return nil, errors.New("viper.Unmarshal")
	}
	return cfg, nil
}

func getEnvPath() (string, error) {
	getwd, err := os.Getwd()
	if err != nil {
		return "", errors.New("os.Getwd")
	}
	configPath = fmt.Sprintf("%s/cmd/tester_service/config/config.yaml", getwd)

	return configPath, nil
}
