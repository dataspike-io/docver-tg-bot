package main

import (
	"github.com/spf13/viper"
)

const secret = "********"

type config struct {
	dataspikeUrl    string
	DataspikeToken  SecretString
	httpPort        int
	prompt          string
	TelegramToken   SecretString
	telegramOffset  int
	telegramTimeout int
	webhookPath     string
	webhookUrl      string
}

func newConfig() config {
	viper.SetDefault("HTTP_PORT", 8080)
	viper.SetDefault("TG_OFFSET", 0)
	viper.SetDefault("TG_TIMEOUT", 60)
	viper.SetDefault("DEBUG_MODE", true)

	// read config
	cfg := config{
		dataspikeUrl:    viper.GetString("DS_URL"),
		DataspikeToken:  NewSecretString(viper.GetString("DS_TOKEN")),
		httpPort:        viper.GetInt("HTTP_PORT"),
		prompt:          viper.GetString("PROMPT"),
		TelegramToken:   NewSecretString(viper.GetString("TG_TOKEN")),
		telegramOffset:  viper.GetInt("TG_OFFSET"),
		telegramTimeout: viper.GetInt("TG_TIMEOUT"),
		webhookPath:     viper.GetString("WEBHOOK_PATH"),
		webhookUrl:      viper.GetString("WEBHOOK_URL"),
	}

	return cfg
}

type SecretString struct {
	secret string
}

func NewSecretString(s string) SecretString {
	return SecretString{s}
}

func (s SecretString) RawString() string {
	return s.secret
}

func (s SecretString) String() string {
	return secret
}

func (s SecretString) ToString() string {
	return secret
}

func (s SecretString) MarshalJSON() ([]byte, error) {
	return []byte(secret), nil
}

func (s SecretString) MarshalText() ([]byte, error) {
	return []byte(secret), nil
}
