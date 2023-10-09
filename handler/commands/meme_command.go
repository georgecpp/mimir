package commands

import (
	"fmt"
	"math/rand"

	"github.com/georgecpp/mimir/misc"
	"github.com/slack-go/slack"
)

// HandleMemeCommand will take care of /meme submissions
func HandleMemeCommand(command slack.SlashCommand, client *slack.Client) error {
	var subreddit string

	// Check if a subreddit was provided as an argument
	if command.Text != "" {
		subreddit = command.Text
	} else {
		// Use default subreddits if none provided
		subreddits := []string{"dankmemes", "196", "memes", "ProgrammerHumor"}
		subreddit = subreddits[rand.Intn(len(subreddits))]
	}

	ch := make(chan string)

	// Fetch a random meme
	go misc.FetchMeme(subreddit, ch)

	// Get the meme URL
	memeURL := <-ch
	close(ch)

	if memeURL != "" {
		// Post the meme to the Slack channel
		_, _, err := client.PostMessage(command.ChannelID, slack.MsgOptionText(memeURL, false))
		if err != nil {
			return fmt.Errorf("failed to post message: %w", err)
		}
		return nil
	}

	return fmt.Errorf("could not fetch a meme from subreddit: %s", subreddit)
}
