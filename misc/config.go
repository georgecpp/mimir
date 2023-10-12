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
	SpotifyRedirectUri	string	`mapstructure:"SPOTIFY_REDIRECT_URI"`
	SpotifyAuthorizeBaseUrl	string	`mapstructure:"SPOTIFY_AUTHORIZE_BASE_URL"`
	SpotifyAuthorizeScopesString	string	`mapstructure:"SPOTIFY_AUTHORIZE_SCOPES_STRING"`
	SpotifyAccessTokenUrl	string	`mapstructure:"SPOTIFY_ACCESS_TOKEN_URL"`
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