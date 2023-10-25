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
			Text: errorMessage,
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
		fmt.Println("GetCurrentPlayingTrack failed with error: %w", err)
		return nil, nil
	}

	// Create the image block
	albumImageBlock := slack.NewImageBlock(currentPlayingTrack.ImageURL, "Album Cover", "", nil)

	// Create the text block with artist and song details
	textBlock := slack.NewTextBlockObject(slack.MarkdownType, fmt.Sprintf("*Artist:* %s\n*Song:* %s", currentPlayingTrack.Artist, currentPlayingTrack.Song), false, false)

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
	 _, slackMessageTimestamp, err := client.PostMessage(command.ChannelID, slack.MsgOptionAttachments(attachment))
	 if err != nil {
		 return nil, fmt.Errorf("failed to post message: %w", err)
	 }

	 misc.MySpotifyDashboard.CreateSpotifyDashboard(
		currentPlayingTrack.Artist,
		currentPlayingTrack.Song,
		currentPlayingTrack.ImageURL,
		slackMessageTimestamp,
	 )
	 
	return attachment, nil
}