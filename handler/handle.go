package handler

import (
	"errors"
	"log"

	"github.com/georgecpp/mimir/handler/commands"
	"github.com/georgecpp/mimir/handler/events"
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
	}

	return nil, nil
}

func HandleInteractionEvent(interaction slack.InteractionCallback, client *slack.Client) (interface{}, error) {
	// This is where we would handle the interaction
	// Switch depending on the Type
	log.Printf("The action called is: %s\n", interaction.ActionID)
	log.Printf("The response was of type: %s\n", interaction.Type)
	switch interaction.Type {
	case slack.InteractionTypeBlockActions:
		// This is a block action, so we need to handle it
		for _, action := range interaction.ActionCallback.BlockActions {
			log.Printf("%+v", action)
			log.Println("Selected option: ", action.SelectedOptions)
		}
	}
	return nil, nil
}