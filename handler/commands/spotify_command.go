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
		if err.Error() == "no currently playing track" {
			// No currently playing track, return a message
			noTrackMessage := "No track is currently playing."
			attachment := slack.Attachment{
				Color: "#36a64f", // Green color
				Text:  noTrackMessage,
			}

			_, _, err := client.PostMessage(command.ChannelID, slack.MsgOptionAttachments(attachment))
			if err != nil {
				return nil, fmt.Errorf("failed to post message: %w", err)
			}

			return nil, nil
		}
		if err.Error() == "failed to retrieve active device" {
			// No currently playing track, return a message
			noActiveDeviceMessage := "No active device is currently playing anything!"
			attachment := slack.Attachment{
				Color: "#36a64f", // Green color
				Text:  noActiveDeviceMessage,
			}

			_, _, err := client.PostMessage(command.ChannelID, slack.MsgOptionAttachments(attachment))
			if err != nil {
				return nil, fmt.Errorf("failed to post message: %w", err)
			}

			return nil, nil
		}
		return nil, fmt.Errorf("GetCurrentPlayingTrack failed with error: %w", err)
	}
	spotifyAttachment := misc.BuildSpotifyAttachment(currentPlayingTrack, "/spotify", command.UserName)

	// Post the message to the channel
	_, slackMessageTimestamp, err := client.PostMessage(command.ChannelID, slack.MsgOptionAttachments(spotifyAttachment))
	if err != nil {
		return nil, fmt.Errorf("failed to post message: %w", err)
	}

	misc.MySpotifyDashboard.CreateSpotifyDashboard(
		currentPlayingTrack,
		slackMessageTimestamp,
		command.ChannelID,
	)

	misc.StartSpotifyPolling(client)

	return spotifyAttachment, nil
}
