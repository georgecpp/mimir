package handler

import (
	"errors"
	"github.com/georgecpp/mimir/handler/commands"
	"github.com/georgecpp/mimir/handler/events"
	"github.com/georgecpp/mimir/handler/interactions"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

// HandleEventMessage will take an event and handle it properly based on the type of event
func HandleEventMessage(event slackevents.EventsAPIEvent, client *slack.Client) (interface{}, error) {
	switch event.Type {
	// first we check if this is an CallbackEvent
	case slackevents.CallbackEvent:
		innerEvent := event.InnerEvent
		// Yet another type switch on the actual Data to see if its an AppMentionEvent
		switch ev := innerEvent.Data.(type) {
		case *slackevents.AppMentionEvent:
			// The application has been mentioned since this Event is a Mention event
			return nil, events.HandleAppMentionEvent(ev, client)
		}
	default:
		return nil, errors.New("unsupported event type")
	}
	return nil, nil
}

// HandleSlashCommand will take a slash command and route to the appropriate function
func HandleSlashCommand(command slack.SlashCommand, client *slack.Client) (interface{}, error) {
	// We need to switch depending on the command
	switch command.Command {
	case "/hello":
		// This was a hello command, so pass it along to the proper function
		return nil, commands.HandleHelloCommand(command, client)
	case "/was-this-article-helpful":
		return commands.HandleIsArticleGood(command, client)
	case "/meme":
		return nil, commands.HandleMemeCommand(command, client)
	case "/spotify-auth":
		return nil, commands.HandleSpotifyAuthCommand(command, client)
	case "/spotify":
		return commands.HandleSpotifyCommand(command, client)
	}

	return nil, nil
}

func HandleInteractionEvent(interaction slack.InteractionCallback, client *slack.Client) (interface{}, error) {
	// This is where we would handle the interaction
	// Switch depending on the Type
	switch interaction.ActionCallback.BlockActions[0].ActionID {
	case "skip_next":
		return interactions.HandleSkipNextInteraction(interaction, client)
	case "skip_previous":
		return interactions.HandleSkipPreviousInteraction(interaction, client)
	case "pause":
		return interactions.HandlePlayPauseInteraction(interaction, client)
	case "play":
		return interactions.HandlePlayPauseInteraction(interaction, client)
	}
	return nil, nil
}
