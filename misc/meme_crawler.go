package misc

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

type RedditResponse struct {
	Data struct {
		Children []struct {
			Data struct {
				URL string `json:"url"`
			} `json:"data"`
		} `json:"children"`
	} `json:"data"`
}

func FetchMeme(subreddit string, ch chan string) {
	url := fmt.Sprintf("https://www.reddit.com/r/%s/top/.json", subreddit)

	for {
		resp, err := http.Get(url)

		if err != nil {
			fmt.Println("Error fetching data:", err)
			ch <- ""
			time.Sleep(1 * time.Second) // Wait for 1 second before retrying
			continue
		}

		if resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()

			var data RedditResponse
			err = json.NewDecoder(resp.Body).Decode(&data)
			if err != nil {
				fmt.Println("Error decoding JSON:", err)
				ch <- ""
				return
			}

			if len(data.Data.Children) == 0 {
				fmt.Println("No memes found in", subreddit)
				ch <- ""
				return
			}

			randGen := rand.New(rand.NewSource(time.Now().UnixNano()))
			randomIndex := randGen.Intn(len(data.Data.Children))
			ch <- data.Data.Children[randomIndex].Data.URL
			return
		} else if resp.StatusCode == 429 {
			fmt.Println("Received 429. Retrying in 1 second...")
			resp.Body.Close()
			time.Sleep(1 * time.Second) // Wait for 1 second before retrying
		} else {
			fmt.Println("Error: Unexpected status code:", resp.StatusCode)
			resp.Body.Close()
			ch <- ""
			return
		}
	}
}