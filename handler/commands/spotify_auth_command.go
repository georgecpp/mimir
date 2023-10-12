package commands

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/georgecpp/mimir/misc"
	"github.com/slack-go/slack"
)

func HandleSpotifyAuthCommand(command slack.SlashCommand, client *slack.Client) error {

	config, err := misc.LoadConfig("../../")
	if err != nil {
		return fmt.Errorf("[spotify-auth] failed to load config: %w", err)
	}
	
	state, err := generateRandomString(16)
    if err != nil {
		return fmt.Errorf("[spotify-auth] failed to generate random state: %w", err)
    }

    scope := config.SpotifyAuthorizeScopesString

    authURL := fmt.Sprintf("%s?%s", config.SpotifyAuthorizeBaseUrl, buildQueryParams(config, state, scope))

    // Send the URL to the user
    message := "Click [here] (" + authURL + ") to authenticate with Spotify."
    _, _, err = client.PostMessage(command.ChannelID, slack.MsgOptionText(message, false))
    if err != nil {
		return fmt.Errorf("[spotify-auth] failed to post message: %w", err)
    }
  
    return nil
}

func generateRandomString(length int) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func buildQueryParams(config misc.Config, state string, scope string) string {
	params := map[string]string{
		"response_type": "code",
		"client_id":     config.SpotifyClientId,
		"scope":         scope,
		"redirect_uri":  config.SpotifyRedirectUri,
		"state":         state,
	}

	var parts []string
	for key, value := range params {
		parts = append(parts, fmt.Sprintf("%s=%s", key, value))
	}

	return strings.Join(parts, "&")
}