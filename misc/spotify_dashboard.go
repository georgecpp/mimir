// In misc/package.go

package misc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/slack-go/slack"
)

// SpotifyDashboard holds the metadata and timestamp of the Spotify message
type SpotifyDashboard struct {
	Artist                string
	Song                  string
	ImageURL              string
	SlackMessageTimestamp string
	SlackChannelId        string
	IsPlaying             bool
	DeviceId              string
	mu                    sync.Mutex // Add a sync.Mutex for synchronization
}

var MySpotifyDashboard SpotifyDashboard
var stopPolling chan struct{} // Channel to signal stopping the polling

// StartSpotifyPolling starts the polling mechanism
func StartSpotifyPolling(client *slack.Client) {
	//stopPolling = make(chan struct{})
	//
	//go func() {
	//	for {
	//		select {
	//		case <-stopPolling:
	//			return // Stop polling when signal is received
	//		default:
	//			// Polling logic here
	//			// For example, update the dashboard
	//			_, err := MySpotifyDashboard.AutoUpdateCurrentSpotifyDashboard(client)
	//			if err != nil {
	//				fmt.Printf("Error updating Spotify dashboard: %v\n", err)
	//			}
	//
	//			// Sleep for a specified interval (e.g., 1 second)
	//			time.Sleep(1 * time.Second)
	//		}
	//	}
	//}()
}

// StopSpotifyPolling stops the polling mechanism
func StopSpotifyPolling() {
	if stopPolling != nil {
		close(stopPolling)
	}
}

// AutoUpdateCurrentSpotifyDashboard updates the SpotifyDashboard with the latest information
func (sd *SpotifyDashboard) AutoUpdateCurrentSpotifyDashboard(client *slack.Client, lastAction string, userName string) (slack.Attachment, error) {
	sd.mu.Lock()
	defer sd.mu.Unlock()

	currentPlayingTrack, err := GetCurrentPlayingTrack()
	if err != nil {
		// Check if the error is a 429 response
		if isRateLimitError(err) {
			// Retry after the specified time
			retryAfter, err := getRetryAfterValue(err)
			if err != nil {
				return slack.Attachment{}, fmt.Errorf("failed to parse Retry-After header: %w", err)
			}
			time.Sleep(time.Duration(retryAfter) * time.Second)
			return slack.Attachment{}, nil
		}
		return slack.Attachment{}, fmt.Errorf("GetCurrentPlayingTrack failed with error: %w", err)
	}

	// Check if the currently playing track is the same as the one in the dashboard
	// and the state is the same
	if currentPlayingTrack.Song == sd.Song && currentPlayingTrack.IsPlaying == sd.IsPlaying {
		// No need to update, as the track is the same and the state as well.
		return slack.Attachment{}, nil
	}

	sd.Artist = currentPlayingTrack.Artist
	sd.Song = currentPlayingTrack.Song
	sd.ImageURL = currentPlayingTrack.ImageURL
	sd.IsPlaying = currentPlayingTrack.IsPlaying
	sd.DeviceId = currentPlayingTrack.DeviceId

	spotifyAttachment := BuildSpotifyAttachment(currentPlayingTrack, lastAction, userName)
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

func (sd *SpotifyDashboard) CreateSpotifyDashboard(cpt CurrentPlayingTrackResponse, timestamp string, channelId string) {
	sd.mu.Lock()
	defer sd.mu.Unlock()

	sd.Artist = cpt.Artist
	sd.Song = cpt.Song
	sd.ImageURL = cpt.ImageURL
	sd.IsPlaying = cpt.IsPlaying
	sd.DeviceId = cpt.DeviceId
	sd.SlackMessageTimestamp = timestamp
	sd.SlackChannelId = channelId
}

type CurrentPlayingTrackResponse struct {
	Artist    string
	Song      string
	ImageURL  string
	IsPlaying bool
	DeviceId  string
}

// GetActiveDevice retrieves the active device ID
func GetActiveDevice() (string, error) {
	// Make a GET request to Spotify API to get the list of devices
	accessToken := Shared.GetSpotifyAccessToken()
	devicesURL := "https://api.spotify.com/v1/me/player/devices"
	devicesReq, err := http.NewRequest("GET", devicesURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create devices request: %w", err)
	}

	devicesReq.Header.Set("Authorization", "Bearer "+accessToken)

	httpClient := &http.Client{}
	devicesResp, err := httpClient.Do(devicesReq)
	if err != nil {
		return "", fmt.Errorf("failed to make devices request: %w", err)
	}
	defer devicesResp.Body.Close()

	if devicesResp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected devices response: %s", devicesResp.Status)
	}

	var devicesData map[string][]map[string]interface{}
	if err := json.NewDecoder(devicesResp.Body).Decode(&devicesData); err != nil {
		return "", fmt.Errorf("failed to decode devices response: %w", err)
	}

	// Find the active device
	for _, device := range devicesData["devices"] {
		if isActive := device["is_active"].(bool); isActive {
			return device["id"].(string), nil
		}
	}

	return "", fmt.Errorf("no active device found")
}

func GetCurrentPlayingTrack() (CurrentPlayingTrackResponse, error) {
	accessToken := Shared.GetSpotifyAccessToken()
	deviceId, err := GetActiveDevice()
	if err != nil {
		return CurrentPlayingTrackResponse{}, fmt.Errorf("failed to retrieve active device")
	}
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

	if resp.StatusCode == http.StatusNoContent {
		// No currently playing track
		return CurrentPlayingTrackResponse{}, fmt.Errorf("no currently playing track")
	}

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
	isPlaying := data["is_playing"].(bool)

	return CurrentPlayingTrackResponse{
		Artist:    artist,
		Song:      song,
		ImageURL:  imageURL,
		IsPlaying: isPlaying,
		DeviceId:  deviceId,
	}, nil
}

type UserQueueItem struct {
	AlbumLogo string
	SongTitle string
	Artist    string
	Duration  string
}

func GetUserQueue() ([]UserQueueItem, error) {
	accessToken := Shared.GetSpotifyAccessToken()

	// Make a GET request to Spotify API for the user's queue
	url := "https://api.spotify.com/v1/me/player/queue"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response: %s", resp.Status)
	}

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check if the 'queue' key exists in the response and is an array
	queue, exists := data["queue"].([]interface{})
	if !exists {
		return nil, fmt.Errorf("queue not found in response or not an array")
	}

	var userQueue []UserQueueItem

	// Iterate through the queue items and extract relevant information
	for _, item := range queue {
		itemData, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		artist := itemData["artists"].([]interface{})[0].(map[string]interface{})["name"].(string)
		song := itemData["name"].(string)
		album := itemData["album"].(map[string]interface{})
		imageURL := album["images"].([]interface{})[1].(map[string]interface{})["url"].(string)
		durationMs := int(itemData["duration_ms"].(float64))
		duration := fmt.Sprintf("%d:%02d", durationMs/60000, (durationMs/1000)%60)

		queueItem := UserQueueItem{
			AlbumLogo: imageURL,
			SongTitle: song,
			Artist:    artist,
			Duration:  duration,
		}

		userQueue = append(userQueue, queueItem)
	}

	return userQueue, nil
}

func PauseTrack() error {
	accessToken := Shared.GetSpotifyAccessToken()

	url := "https://api.spotify.com/v1/me/player/pause"
	req, err := http.NewRequest("PUT", url, nil)
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

func StartResumeTrack() error {
	accessToken := Shared.GetSpotifyAccessToken()

	url := "https://api.spotify.com/v1/me/player/play"
	req, err := http.NewRequest("PUT", url, nil)
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

func BuildSpotifyAttachment(track CurrentPlayingTrackResponse, lastAction string, userName string) slack.Attachment {

	// Create a section block for displaying last action and user
	lastActionBlock := slack.NewSectionBlock(
		slack.NewTextBlockObject(slack.MarkdownType, fmt.Sprintf("*Last Action:* %s\n*DJ:* %s", lastAction, userName), false, false),
		nil,
		nil,
	)

	// Create a divider block
	dividerBlock := slack.NewDividerBlock()

	// Create the image block
	albumImageBlock := slack.NewImageBlock(track.ImageURL, "Album Cover", "", nil)

	// Create the text block with artist and song details
	textBlock := slack.NewTextBlockObject(slack.MarkdownType, fmt.Sprintf("*Artist:* %s\n*Song:* %s", track.Artist, track.Song), false, false)

	// Create the section block with the text and image blocks
	songMetadataBlock := slack.NewSectionBlock(textBlock, nil, nil)

	// Create buttons for controls
	previousButton := slack.NewButtonBlockElement("skip_previous", "skip_previous", slack.NewTextBlockObject(slack.PlainTextType, "⏪", false, false))

	// Create play/pause button based on track state
	var playPauseButton *slack.ButtonBlockElement
	if track.IsPlaying {
		playPauseButton = slack.NewButtonBlockElement("pause", "pause", slack.NewTextBlockObject(slack.PlainTextType, "⏸️", false, false))
	} else {
		playPauseButton = slack.NewButtonBlockElement("play", "play", slack.NewTextBlockObject(slack.PlainTextType, "▶️", false, false))
	}

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
				dividerBlock,
				lastActionBlock,
			},
		},
	}

	return attachment
}
