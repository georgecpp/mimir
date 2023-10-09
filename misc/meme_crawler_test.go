package misc

import (
	"fmt"
	"testing"
)

func TestFetchMeme(t *testing.T) {
	subreddits := []string{"dankmemes", "196", "memes", "ProgrammerHumor"}

	ch := make(chan string)

	for _, subreddit := range subreddits {
		go FetchMeme(subreddit, ch)
	}

	memeURL := <-ch
	if memeURL != "" {
		fmt.Println("Random Meme URL:", memeURL)
		return
	}

	fmt.Println("Could not fetch a meme from any subreddit.")
}