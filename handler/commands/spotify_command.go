package commands

import (
	"fmt"
	"encoding/json"
	"net/http"

	"github.com/georgecpp/mimir/misc"
	"github.com/slack-go/slack"
)

func HandleSpotifyCommand(command slack.SlashCommand, client *slack.Client) (interface{}, error) {	
	
	accessToken := misc.Shared.GetSpotifyAccessToken() // Retrieve the Spotify access token
	// Check if the access token is set
	if accessToken == "" {
		// Access token is not set, return an error message
		errorMessage := "Not connected to Spotify. Run /spotify-auth to enable this!"
		
		attachment := slack.Attachment{
			Color: "#FF0000", // Red color
			Text: errorMessage,
		}

		_, _, err := client.PostMessage(command.ChannelID, slack.MsgOptionAttachments(attachment))
		if err != nil {
			return nil, fmt.Errorf("failed to post message: %w", err)
		}

		return nil, nil
	}

	// Make a GET request to Spotify API
	url := "https://api.spotify.com/v1/me/player/currently-playing"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response: %s", resp.Status)
	}

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract the information from 'data'
	artist := data["item"].(map[string]interface{})["artists"].([]interface{})[0].(map[string]interface{})["name"].(string)
	song := data["item"].(map[string]interface{})["name"].(string)
	imageURL := data["item"].(map[string]interface{})["album"].(map[string]interface{})["images"].([]interface{})[1].(map[string]interface{})["url"].(string)

	// Create the image block
	albumImageBlock := slack.NewImageBlock(imageURL, "Album Cover", "", nil)

	// Create the text block with artist and song details
	textBlock := slack.NewTextBlockObject(slack.MarkdownType, fmt.Sprintf("*Artist:* %s\n*Song:* %s", artist, song), false, false)

	// Create the section block with the text and image blocks
	songMetadataBlock := slack.NewSectionBlock(textBlock, nil, nil)

	// Create buttons for controls
	previousButton := slack.NewButtonBlockElement("", "skip_previous", slack.NewTextBlockObject(slack.PlainTextType, "⏪", false, false))
	playPauseButton := slack.NewButtonBlockElement("", "play_pause", slack.NewTextBlockObject(slack.PlainTextType, "▶️/⏸️", false, false))
	nextButton := slack.NewButtonBlockElement("", "skip_next", slack.NewTextBlockObject(slack.PlainTextType, "⏩", false, false))
	
	// Create an action block with buttons
	actionBlock := slack.NewActionBlock(
		"controls",
		previousButton,
		playPauseButton,
		nextButton,
	)

	// Create the attachment
	attachment := slack.Attachment{
		Blocks: slack.Blocks{
			BlockSet: []slack.Block{
				albumImageBlock,
				songMetadataBlock,
				actionBlock,
			},
		},
	}

	 // Post the message to the channel
	 _, _, err = client.PostMessage(command.ChannelID, slack.MsgOptionAttachments(attachment))
	 if err != nil {
		 return nil, fmt.Errorf("failed to post message: %w", err)
	 }

	return attachment, nil
}