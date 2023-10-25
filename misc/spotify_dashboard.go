// In misc/package.go

package misc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

// SpotifyDashboard holds the metadata and timestamp of the Spotify message
type SpotifyDashboard struct {
    Artist    string
    Song      string
    ImageURL  string
    SlackMessageTimestamp string
    mu        sync.Mutex // Add a sync.Mutex for synchronization
}

var MySpotifyDashboard SpotifyDashboard

// AutoUpdateCurrentSpotifyDashboard updates the SpotifyDashboard with the latest information
func (sd *SpotifyDashboard) AutoUpdateCurrentSpotifyDashboard() {
    sd.mu.Lock()
    defer sd.mu.Unlock()

    currentPlayingTrack, err := GetCurrentPlayingTrack()
	if err != nil {
		fmt.Println("GetCurrentPlayingTrack failed with error: %w", err)
	}
    sd.Artist = currentPlayingTrack.Artist
    sd.Song = currentPlayingTrack.Song
    sd.ImageURL = currentPlayingTrack.ImageURL
}

func (sd *SpotifyDashboard) CreateSpotifyDashboard(artist, song, imageURL, timestamp string) {
    sd.mu.Lock()
    defer sd.mu.Unlock()

    sd.Artist = artist
    sd.Song = song
    sd.ImageURL = imageURL
    sd.SlackMessageTimestamp = timestamp 
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