package interactions

import (
	"fmt"

	"github.com/georgecpp/mimir/misc"
	"github.com/slack-go/slack"
)

func HandleSkipNextInteraction(interaction slack.InteractionCallback, client *slack.Client) (interface{}, error) {
	err := misc.SkipToNextTrack()
	if err != nil {
		return nil, fmt.Errorf("SkipToNextTrack failed with error: %w", err)
	}
	spotifyAttachment, err := misc.MySpotifyDashboard.AutoUpdateCurrentSpotifyDashboard(client)
	if err != nil {
		return nil, fmt.Errorf("AutoUpdateCurrentSpotifyDashboard failed with error: %w", err)
	}
	return spotifyAttachment, nil
}