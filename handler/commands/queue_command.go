package commands

import (
	"fmt"
	"github.com/slack-go/slack"
)

func HandleQueueCommand(command slack.SlashCommand, client *slack.Client) (interface{}, error) {
	sectionTitle := slack.NewTextBlockObject("plain_text", "Spotify Queue", false, false)

	// Dummy data for the Spotify queue
	queueData := []struct {
		AlbumLogo string
		SongTitle string
		Artist    string
		Duration  string
	}{
		{AlbumLogo: "https://i.scdn.co/image/ab67616d00004851522088789d49e216d9818292", SongTitle: "Song Title 1", Artist: "Artist 1", Duration: "2:35"},
		{AlbumLogo: "https://i.scdn.co/image/ab67616d00004851522088789d49e216d9818292", SongTitle: "Song Title 2", Artist: "Artist 2", Duration: "3:10"},
		// Add more dummy data as needed
	}

	var queueBlocks []slack.Block

	for _, song := range queueData {
		// Create an image block for the album logo
		imageBlock := slack.NewImageBlockElement(song.AlbumLogo, "album logo")

		// Create a text block for the song information
		songText := fmt.Sprintf("*%s* by %s - %s", song.SongTitle, song.Artist, song.Duration)
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
