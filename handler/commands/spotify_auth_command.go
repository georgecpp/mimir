package commands

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/georgecpp/mimir/misc"
	"github.com/slack-go/slack"
)

type TinyURLResponse struct {
	Data struct {
		TinyURL string `json:"tiny_url"`
	} `json:"data"`
}

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

	// Shorten the URL using TinyURL
	shortenedURL, err := shortenURL(config, authURL)
	if err != nil {
		return fmt.Errorf("[spotify-auth] failed to shorten URL: %w", err)
	}

	// Send the shortened URL to the user
	message := fmt.Sprintf("Let's go!\nClick here: %s to authenticate with Spotify and let's get this party started 🎶", shortenedURL)
	_, _, err = client.PostMessage(command.ChannelID, slack.MsgOptionText(message, false))
	if err != nil {
		return fmt.Errorf("[spotify-auth] failed to post message: %w", err)
	}

	return nil
}

func shortenURL(config misc.Config, longURL string) (string, error) {
	client := &http.Client{}

	payload := strings.NewReader(fmt.Sprintf(`{"url": "%s"}`, longURL)) // Construct the JSON payload

	req, err := http.NewRequest("POST", config.TinyUrlApiCreateUrl, payload)

	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", config.TinyUrlAccessToken))

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	var response TinyURLResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return "", err
	}

	return response.Data.TinyURL, nil
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