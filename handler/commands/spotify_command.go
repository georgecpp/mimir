package commands

import (
	"fmt"

	"github.com/slack-go/slack"
)

func HandleSpotifyCommand(command slack.SlashCommand, client *slack.Client) (interface{}, error) {
	// Define your dummy data
	artist := "Dummy Artist"
	song := "Dummy Song"
	imageURL := "https://i.scdn.co/image/ab67616d0000b273a63fc9073db1233ea6c7ae74"

	// Create the image block
	imageBlock := slack.NewImageBlockElement(imageURL, "Album Cover")

	// Create the text block with artist and song details
	textBlock := slack.NewTextBlockObject(slack.MarkdownType, fmt.Sprintf("*Artist:* %s\n*Song:* %s", artist, song), false, false)

	// Create the section block with the text and image blocks
	sectionBlock := slack.NewSectionBlock(textBlock, nil, slack.NewAccessory(imageBlock))

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
				sectionBlock,
				actionBlock,
			},
		},
	}

	 // Post the message to the channel
	 _, _, err := client.PostMessage(command.ChannelID, slack.MsgOptionAttachments(attachment))
	 if err != nil {
		 return nil, fmt.Errorf("failed to post message: %w", err)
	 }

	return attachment, nil
}