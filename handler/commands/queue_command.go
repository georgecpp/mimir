package commands

import (
	"fmt"
	"github.com/georgecpp/mimir/misc"
	"github.com/slack-go/slack"
)

func HandleQueueCommand(command slack.SlashCommand, client *slack.Client) (interface{}, error) {
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
	_, err := misc.GetCurrentPlayingTrack()
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

	myQueueData, err := misc.GetUserQueue()

	var queueBlocks []slack.Block

	for i, item := range myQueueData[:5] {
		position := i + 1 // Adjust to start the count from 1 instead of 0
		// Create the section title dynamically with the position in the queue
		sectionTitle := slack.NewTextBlockObject("plain_text", fmt.Sprintf("#%d", position), false, false)
		imageBlock := slack.NewImageBlockElement(item.AlbumLogo, "album logo")
		// Create a text block for the song information
		songText := fmt.Sprintf("*%s*\t\t\t%s\t\t\t%s", item.SongTitle, item.Artist, item.Duration)
		songBlock := slack.NewTextBlockObject("mrkdwn", songText, false, false)

		// Create a section block with the song's text and image blocks as fields
		sectionBlock := slack.NewSectionBlock(sectionTitle, []*slack.TextBlockObject{songBlock}, slack.NewAccessory(imageBlock))

		// Append the section block to the list of queue blocks
		queueBlocks = append(queueBlocks, sectionBlock)
	}

	attachment := slack.Attachment{
		Blocks: slack.Blocks{
			BlockSet: queueBlocks,
		},
	}
	return attachment, nil
}
