package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"

	"github.com/RocketChat/Rocket.Chat.Go.SDK/models"
	"gopkg.in/yaml.v2"
)

type AppConfig struct {
	ListenAddr string
	RocketChat RCConfig
}

// RCChannelInfo - Channel configuration
type RCChannelInfo struct {
	DefaultChannelName string `yaml:"default_channel_name"`
}

// RCConfig - Rocket.Chat webhook configuration
type RCConfig struct {
	Endpoint       url.URL                `yaml:"endpoint"`
	Credentials    models.UserCredentials `yaml:"credentials"`
	SeverityColors map[string]string      `yaml:"severity_colors"`
	Channel        RCChannelInfo          `yaml:"channel"`
}

func checkConfig(config RCConfig) error {
	if config.Credentials.Name == "" {
		return errors.New("rocket.chat name not provided")
	}
	if config.Credentials.Email == "" {
		return errors.New("rocket.chat email not provided")
	}
	if config.Credentials.Password == "" {
		return errors.New("rocket.chat password not provided")
	}
	if config.Endpoint.Host == "" {
		return errors.New("rocket.chat host not provided")
	}
	if config.Endpoint.Scheme == "" {
		return errors.New("rocket.chat scheme not provided")
	}
	return nil
}

func LoadConfig(configFile, listenAddress string) (AppConfig, error) {
	rcCfg := RCConfig{}

	// Load the config from the file
	f, err := os.Open(configFile)
	if err != nil {
		return AppConfig{}, fmt.Errorf("cannot open config file for reading: %w", err)
	}

	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	if err := yaml.NewDecoder(f).Decode(&rcCfg); err != nil {
		return AppConfig{}, fmt.Errorf("cannot unmarshal config: %w", err)
	}

	if err := checkConfig(rcCfg); err != nil {
		return AppConfig{}, fmt.Errorf("missing Rocket.Chat config parameters: %w", err)
	}

	return AppConfig{
		ListenAddr: listenAddress,
		RocketChat: rcCfg,
	}, nil
}
