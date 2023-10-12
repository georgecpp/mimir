package misc

import (
	"github.com/spf13/viper"
)

// Config stores all configurations of the app
// the values are read by viper from a config file or env variables
type Config struct {
	SlackAuthToken     string  `mapstructure:"SLACK_AUTH_TOKEN"`
	SlackChannelId		string	`mapstructure:"SLACK_CHANNEL_ID"`
	SlackAppToken		string	`mapstructure:"SLACK_APP_TOKEN"`
	SpotifyClientId		string	`mapstructure:"SPOTIFY_CLIENT_ID"`
	SpotifyClientSecret	string	`mapstructure:"SPOTIFY_CLIENT_SECRET"`
}

// LoadConfig reads config from file or env variables
func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}