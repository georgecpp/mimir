package commands

import (
	"fmt"
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
			Text:  errorMessage,
		}

		_, _, err := client.PostMessage(command.ChannelID, slack.MsgOptionAttachments(attachment))
		if err != nil {
			return nil, fmt.Errorf("failed to post message: %w", err)
		}

		return nil, nil
	}

	// Get the currently playing track nice and tidy.
	currentPlayingTrack, err := misc.GetCurrentPlayingTrack()
	if err != nil {
		return nil, fmt.Errorf("GetCurrentPlayingTrack failed with error: %w", err)
	}

	spotifyAttachment := misc.BuildSpotifyAttachment(currentPlayingTrack)

	// Post the message to the channel
	_, slackMessageTimestamp, err := client.PostMessage(command.ChannelID, slack.MsgOptionAttachments(spotifyAttachment))
	if err != nil {
		return nil, fmt.Errorf("failed to post message: %w", err)
	}

	misc.MySpotifyDashboard.CreateSpotifyDashboard(
		currentPlayingTrack.Artist,
		currentPlayingTrack.Song,
		currentPlayingTrack.ImageURL,
		slackMessageTimestamp,
		command.ChannelID,
	)

	misc.StartSpotifyPolling(client)

	return spotifyAttachment, nil
}
