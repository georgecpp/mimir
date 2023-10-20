package misc

import (
	"sync"
)

// SharedData holds data shared across command functions
type SharedData struct {
	spotifyAccessToken string
	mutex             sync.Mutex
}

var Shared SharedData

// SetSpotifyAccessToken sets the Spotify access token with concurrency safety
func (s *SharedData) SetSpotifyAccessToken(token string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.spotifyAccessToken = token
}

// GetSpotifyAccessToken retrieves the Spotify access token with concurrency safety
func (s *SharedData) GetSpotifyAccessToken() string {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.spotifyAccessToken
}
