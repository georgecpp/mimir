// In misc/package.go

package misc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/slack-go/slack"
)

// SpotifyDashboard holds the metadata and timestamp of the Spotify message
type SpotifyDashboard struct {
    Artist    string
    Song      string
    ImageURL  string
    SlackMessageTimestamp string
    SlackChannelId string
    mu        sync.Mutex // Add a sync.Mutex for synchronization
}

var MySpotifyDashboard SpotifyDashboard

// AutoUpdateCurrentSpotifyDashboard updates the SpotifyDashboard with the latest information
func (sd *SpotifyDashboard) AutoUpdateCurrentSpotifyDashboard(client *slack.Client) (slack.Attachment, error) {
    sd.mu.Lock()
    defer sd.mu.Unlock()

    currentPlayingTrack, err := GetCurrentPlayingTrack()
	if err != nil {
		fmt.Println("GetCurrentPlayingTrack failed with error: %w", err)
	}
    sd.Artist = currentPlayingTrack.Artist
    sd.Song = currentPlayingTrack.Song
    sd.ImageURL = currentPlayingTrack.ImageURL

    spotifyAttachment := BuildSpotifyAttachment(currentPlayingTrack)
    _, _, _, err = client.UpdateMessage(
        sd.SlackChannelId,
        sd.SlackMessageTimestamp,
        slack.MsgOptionAttachments(spotifyAttachment),
    )
    if err != nil {
        return slack.Attachment{}, fmt.Errorf("client.UpdateMessage failed to update message: %w", err)
    }
    return spotifyAttachment, nil
}

func (sd *SpotifyDashboard) CreateSpotifyDashboard(artist, song, imageURL, timestamp, channelId string) {
    sd.mu.Lock()
    defer sd.mu.Unlock()

    sd.Artist = artist
    sd.Song = song
    sd.ImageURL = imageURL
    sd.SlackMessageTimestamp = timestamp 
    sd.SlackChannelId = channelId
}

type CurrentPlayingTrackResponse struct {
    Artist  string
    Song    string
    ImageURL string
}

func GetCurrentPlayingTrack() (CurrentPlayingTrackResponse, error) {
	accessToken := Shared.GetSpotifyAccessToken()
	// Make a GET request to Spotify API
	url := "https://api.spotify.com/v1/me/player/currently-playing"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return CurrentPlayingTrackResponse{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return CurrentPlayingTrackResponse{}, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return CurrentPlayingTrackResponse{}, fmt.Errorf("unexpected response: %s", resp.Status)
	}

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return CurrentPlayingTrackResponse{}, fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract the information from 'data'
	artist := data["item"].(map[string]interface{})["artists"].([]interface{})[0].(map[string]interface{})["name"].(string)
	song := data["item"].(map[string]interface{})["name"].(string)
	imageURL := data["item"].(map[string]interface{})["album"].(map[string]interface{})["images"].([]interface{})[1].(map[string]interface{})["url"].(string)

	return CurrentPlayingTrackResponse{
		Artist:   artist,
		Song:     song,
		ImageURL: imageURL,
	}, nil
}

func SkipToNextTrack() error {
	accessToken := Shared.GetSpotifyAccessToken()

	url := "https://api.spotify.com/v1/me/player/next"
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected response: %s", resp.Status)
	}

	return nil
}

func SkipToPreviousTrack() error {
    accessToken := Shared.GetSpotifyAccessToken()

	url := "https://api.spotify.com/v1/me/player/previous"
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected response: %s", resp.Status)
	}

	return nil
}

func BuildSpotifyAttachment(track CurrentPlayingTrackResponse) slack.Attachment {
	// Create the image block
	albumImageBlock := slack.NewImageBlock(track.ImageURL, "Album Cover", "", nil)

	// Create the text block with artist and song details
	textBlock := slack.NewTextBlockObject(slack.MarkdownType, fmt.Sprintf("*Artist:* %s\n*Song:* %s", track.Artist, track.Song), false, false)

	// Create the section block with the text and image blocks
	songMetadataBlock := slack.NewSectionBlock(textBlock, nil, nil)

	// Create buttons for controls
	previousButton := slack.NewButtonBlockElement("skip_previous", "skip_previous", slack.NewTextBlockObject(slack.PlainTextType, "⏪", false, false))
	playPauseButton := slack.NewButtonBlockElement("play_pause", "play_pause", slack.NewTextBlockObject(slack.PlainTextType, "▶️/⏸️", false, false))
	nextButton := slack.NewButtonBlockElement("skip_next", "skip_next", slack.NewTextBlockObject(slack.PlainTextType, "⏩", false, false))

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

	return attachment
}