package config

import (
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Configuration struct {
	Server struct {
		Port string `mapstructure:"port"`
		Host string `mapstructure:"host"`
	} `mapstructure:"server"`

	Database struct {
		Host     string `mapstructure:"host"`
		Port     string `mapstructure:"port"`
		User     string `mapstructure:"user"`
		Password string `mapstructure:"password"`
		Name     string `mapstructure:"name"`
		SSLMode  string `mapstructure:"ssl_mode"`
	} `mapstructure:"database"`

	Webhook struct {
		URL string `mapstructure:"url"`
	} `mapstructure:"webhook"`

	Redis struct {
		Host string `mapstructure:"host"`
		Port string `mapstructure:"port"`
	} `mapstructure:"redis"`

	Environment string
}

var AppSettings Configuration

func LoadSettings() error {
	env := strings.ToLower(os.Getenv("APP_ENV"))
	if env == "" || (env != "prod" && env != "production") {
		env = "test"
	} else {
		env = "prod"
	}

	AppSettings.Environment = env

	viper.SetConfigName("config." + env)
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")

	envConfigErr := viper.ReadInConfig()

	if envConfigErr != nil {
		viper.SetConfigName("config")
		if err := viper.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return err
			}
		}
	}

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", "5432")
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "postgres")
	viper.SetDefault("database.name", "auto_message_sender")
	viper.SetDefault("database.ssl_mode", "disable")
	viper.SetDefault("webhook.url", "https://webhook.site/c3f13233-1ed4-429e-9649-8133b3b9c9cd")
	viper.SetDefault("webhook.auth_key_name", "x-ins-auth-key")
	viper.SetDefault("webhook.auth_key", "INS.me1x9uMcyYGlhKKQVPoc.bO3j9aZwRTOcA2Ywo")
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", "6379")

	return viper.Unmarshal(&AppSettings)
}
