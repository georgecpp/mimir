package interactions

import (
	"fmt"
	"github.com/georgecpp/mimir/misc"
	"github.com/slack-go/slack"
)

func HandlePlayPauseInteraction(interaction slack.InteractionCallback, client *slack.Client) (interface{}, error) {
	var err error
	cpt, err := misc.GetCurrentPlayingTrack()
	if err != nil {
		return nil, fmt.Errorf("[HandlePlayPauseInteraction]: GetCurrentPlayingTrack failed with error: %w", err)
	}
	playing := cpt.IsPlaying
	if playing {
		err = misc.PauseTrack()
		if err != nil {
			return nil, fmt.Errorf("PauseTrack failed with error: %w", err)
		}
	} else {
		err = misc.StartResumeTrack()
		if err != nil {
			return nil, fmt.Errorf("StartResumeTrack failed with error: %w", err)
		}
	}
	spotifyAttachment, err := misc.MySpotifyDashboard.AutoUpdateCurrentSpotifyDashboard(client)
	if err != nil {
		return nil, fmt.Errorf("AutoUpdateCurrentSpotifyDashboard failed with error: %w", err)
	}
	return spotifyAttachment, nil
}
