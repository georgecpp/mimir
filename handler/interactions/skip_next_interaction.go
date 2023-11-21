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
	lastAction := interaction.ActionCallback.BlockActions[0].ActionID
	userName := interaction.User.Name
	spotifyAttachment, err := misc.MySpotifyDashboard.AutoUpdateCurrentSpotifyDashboard(client, lastAction, userName)
	if err != nil {
		return nil, fmt.Errorf("AutoUpdateCurrentSpotifyDashboard failed with error: %w", err)
	}
	return spotifyAttachment, nil
}
